// Copyright 2026 The Mangle Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package engine

import (
	"codeberg.org/TauCeti/mangle-go/ast"
	"codeberg.org/TauCeti/mangle-go/builtin"
)

// PremiseOrder describes a rule whose premises the engine evaluated in a
// different order than written. See [WithPremiseReordering].
type PremiseOrder struct {
	// Rule is the rule as written, after the analysis has rewritten it
	// (for example, to wait with a negation until its variables are bound).
	Rule ast.Clause
	// Evaluated is the same rule with its premises in the order evaluated.
	Evaluated ast.Clause
	// NewFacts is -1 for the first round of a stratum. In the later rounds,
	// it is the position in Rule.Premises of the premise that ranges over
	// the facts derived in the previous round only.
	NewFacts int
}

// WithPremiseReordering lets the engine evaluate the premises of a rule in
// a different order than written, so as not to build cross products. The
// written order is kept, except that:
//   - a positive atom that shares a variable with what is bound already
//     goes before one that does not, so that no cross product is built when
//     it can be avoided;
//   - an inequality is evaluated as soon as its variables have values, and
//     an equality without a function application as soon as one side has a
//     value.
//
// The engine does not know how many facts a relation has, so the order it
// chooses is not always faster. If a premise it moves later would have
// found no facts, the written order would have stopped there, and the
// reordered one does the work before it for nothing.
//
// When the written order derives facts without an error, the reordered one
// derives the same facts. A premise that can fail at run time (a built-in
// predicate, or one that applies a function) never moves ahead of a premise
// written before it, so reordering never makes an evaluation fail that
// succeeds as written. It can make an evaluation succeed that fails as
// written: a premise moved ahead can remove the row that the failing
// premise fails on. A negation, or an atom whose variables are all bound,
// never moves ahead of a premise written before it either: each looks up
// facts for every row, and in a store without indexes, which scans the
// relation on every lookup, moving it ahead can cost far more than it
// saves. The limit of [WithCreatedFactLimit] applies to the intermediate
// results of the order evaluated.
//
// Rules with temporal, external, deferred or moded predicates are evaluated
// as written, and so is a rule with a premise over a predicate declared
// with descr [no_reorder()]. The rules evaluated in a different order are
// listed in [Stats].PremiseOrders.
func WithPremiseReordering() EvalOption {
	return func(o *EvalOptions) { o.reorderPremises = true }
}

// ReorderPremises returns clause with its premises in the order the engine
// evaluates them under [WithPremiseReordering], and whether that order
// differs from the written one. decls are the program's declarations; it
// may be nil when the program has none.
func ReorderPremises(clause ast.Clause, decls map[ast.PredicateSym]*ast.Decl) (ast.Clause, bool) {
	order, ok := premiseOrder(clause, decls)
	if !ok {
		return clause, false
	}
	premises := make([]ast.Term, len(order))
	for i, j := range order {
		premises[i] = clause.Premises[j]
	}
	reordered := clause
	reordered.Premises = premises
	return reordered, true
}

// premiseOrder returns the positions of clause's premises in evaluation
// order, or false when the written order is kept.
func premiseOrder(clause ast.Clause, decls map[ast.PredicateSym]*ast.Decl) ([]int, bool) {
	n := len(clause.Premises)
	if n < 2 || clause.HeadTime != nil {
		return nil, false
	}
	if clause.Transform != nil && !clause.Transform.IsLetTransform() {
		// Rules with do-transforms are evaluated over a single premise.
		return nil, false
	}
	if decl, ok := decls[clause.Head.Predicate]; ok && !plainDecl(decl) {
		return nil, false
	}
	premises := make([]premise, n)
	for i, term := range clause.Premises {
		p, ok := classifyPremise(term, decls)
		if !ok {
			return nil, false
		}
		premises[i] = p
	}

	bound := make(map[ast.Variable]bool)
	placed := make([]bool, n)
	order := make([]int, 0, n)
	// Some premises wait for every premise written before them.
	placeable := func(i int) bool {
		if placed[i] || !premises[i].ready(bound) {
			return false
		}
		if premises[i].waits(bound) {
			for j := 0; j < i; j++ {
				if !placed[j] {
					return false
				}
			}
		}
		return true
	}
	// first returns the first placeable premise, in written order, that
	// satisfies accept, or -1.
	first := func(accept func(p premise) bool) int {
		for i := range premises {
			if placeable(i) && accept(premises[i]) {
				return i
			}
		}
		return -1
	}
	for len(order) < n {
		i := first(func(p premise) bool { return p.filters(bound) })
		if i < 0 {
			i = first(func(p premise) bool { return p.kind == eqPremise })
		}
		if i < 0 {
			i = first(func(p premise) bool { return p.connected(bound) })
		}
		if i < 0 {
			i = first(func(p premise) bool { return true })
		}
		if i < 0 {
			// No premise can go next: keep the written order.
			return nil, false
		}
		placed[i] = true
		order = append(order, i)
		premises[i].bind(bound)
	}
	for i, j := range order {
		if i != j {
			return order, true
		}
	}
	return nil, false
}

// plainDecl reports whether a predicate is evaluated bottom-up from the
// fact store, the only kind of predicate whose rules can be reordered.
func plainDecl(decl *ast.Decl) bool {
	return !decl.IsExternal() && !decl.DeferredPredicate() && !decl.IsTemporal() &&
		!decl.IsMaybeTemporal() && len(decl.Modes()) == 0
}

type premiseKind int

const (
	atomPremise    premiseKind = iota // positive atom over the fact store
	builtinPremise                    // positive atom of a built-in predicate
	negPremise                        // negated atom
	ineqPremise                       // X != Y
	eqPremise                         // X = Y
)

// premise is what the planner knows about one premise of a rule.
type premise struct {
	kind premiseKind
	term ast.Term
	// vars are the variables of the premise, without the wildcard "_".
	vars []ast.Variable
	// mayFail is set when evaluating the premise can return an error.
	mayFail bool
	// wildcard is set when the premise has a "_" argument.
	wildcard bool
	// mode and out are set for built-in predicates: out are the variables
	// in an output position.
	mode ast.Mode
	out  map[ast.Variable]bool
}

var wildcard = ast.Variable{Symbol: "_"}

// classifyPremise returns what the planner needs to know about term, or
// false when the planner does not reorder rules with such a premise.
func classifyPremise(term ast.Term, decls map[ast.PredicateSym]*ast.Decl) (premise, bool) {
	p := premise{term: term}
	vars := make(map[ast.Variable]bool)
	ast.AddVars(term, vars)
	_, p.wildcard = vars[wildcard]
	delete(vars, wildcard)
	for v := range vars {
		p.vars = append(p.vars, v)
	}
	switch t := term.(type) {
	case ast.Atom:
		if t.Predicate.IsBuiltin() {
			mode, ok := builtin.Predicates[t.Predicate]
			if !ok || len(mode) != len(t.Args) {
				return p, false
			}
			p.kind, p.mayFail, p.mode = builtinPremise, true, mode
			p.out = make(map[ast.Variable]bool)
			for i, m := range mode {
				if v, ok := t.Args[i].(ast.Variable); ok && m&(ast.ArgModeOutput|ast.ArgModeInputOutput) != 0 {
					p.out[v] = true
				}
			}
			return p, true
		}
		if !plainAtom(t, decls) {
			return p, false
		}
		p.kind = atomPremise
		return p, true
	case ast.NegAtom:
		if !t.Atom.Predicate.IsBuiltin() && !plainAtom(t.Atom, decls) {
			return p, false
		}
		p.kind, p.mayFail = negPremise, t.Atom.Predicate.IsBuiltin()
		return p, true
	case ast.Ineq:
		if p.wildcard {
			return p, false
		}
		p.kind, p.mayFail = ineqPremise, hasApply(t.Left) || hasApply(t.Right)
		return p, true
	case ast.Eq:
		if p.wildcard {
			return p, false
		}
		p.kind, p.mayFail = eqPremise, hasApply(t.Left) || hasApply(t.Right)
		return p, true
	}
	return p, false
}

// plainAtom reports whether atom is looked up in the fact store with
// arguments that are variables and constants only, and its predicate is not
// declared no_reorder.
func plainAtom(atom ast.Atom, decls map[ast.PredicateSym]*ast.Decl) bool {
	for _, arg := range atom.Args {
		if hasApply(arg) {
			return false
		}
	}
	pred := atom.Predicate
	if isDeltaPredicate(pred) {
		pred = makeNormalPredicate(pred)
	}
	if decl, ok := decls[pred]; ok && (!plainDecl(decl) || decl.NoReorder()) {
		return false
	}
	return true
}

func hasApply(term ast.BaseTerm) bool {
	_, ok := term.(ast.ApplyFn)
	return ok
}

func allBound(vars []ast.Variable, bound map[ast.Variable]bool) bool {
	for _, v := range vars {
		if !bound[v] {
			return false
		}
	}
	return true
}

// ready reports whether the premise can be evaluated once the variables in
// bound have values.
func (p premise) ready(bound map[ast.Variable]bool) bool {
	switch p.kind {
	case atomPremise:
		return true
	case builtinPremise:
		if p.mode.Check(p.term.(ast.Atom), bound) != nil {
			return false
		}
		for _, v := range p.vars {
			if !bound[v] && !p.out[v] {
				return false
			}
		}
		return true
	case negPremise, ineqPremise:
		return allBound(p.vars, bound)
	case eqPremise:
		eq := p.term.(ast.Eq)
		return eqSideReady(eq.Left, eq.Right, bound) || eqSideReady(eq.Right, eq.Left, bound)
	}
	return false
}

// eqSideReady reports whether target = expr can be evaluated: expr has all
// its variables bound, and target is too or is a variable it can bind. (An
// equality of two variables is compiled away by the analysis, so it only
// reaches here through [ReorderPremises] on a clause that was not analyzed.)
func eqSideReady(target, expr ast.BaseTerm, bound map[ast.Variable]bool) bool {
	vars := make(map[ast.Variable]bool)
	ast.AddVars(expr, vars)
	for v := range vars {
		if !bound[v] {
			return false
		}
	}
	if _, ok := target.(ast.Variable); ok {
		return true
	}
	vars = make(map[ast.Variable]bool)
	ast.AddVars(target, vars)
	for v := range vars {
		if !bound[v] {
			return false
		}
	}
	return true
}

// waits reports whether the premise must not move ahead of a premise written
// before it: it can fail, or it looks up facts only to filter rows. A lookup
// costs a scan of the relation in a store without indexes, so evaluating it
// on more rows than as written can cost more than the rows it removes save.
func (p premise) waits(bound map[ast.Variable]bool) bool {
	return p.mayFail || p.kind == negPremise || (p.kind == atomPremise && p.filters(bound))
}

// filters reports whether the premise binds no new variable, so that it
// can only remove rows. An atom with a "_" argument is not a filter: it
// gives one row for each fact that matches.
func (p premise) filters(bound map[ast.Variable]bool) bool {
	if p.kind == atomPremise && p.wildcard {
		return false
	}
	return allBound(p.vars, bound)
}

// connected reports whether a premise joins with what is bound already
// instead of building a cross product: it has a variable that is bound. A
// built-in predicate is connected through its inputs. A constant argument
// does not count: how many facts it selects depends on the data, which the
// planner does not know.
func (p premise) connected(bound map[ast.Variable]bool) bool {
	switch p.kind {
	case builtinPremise:
		return true
	case atomPremise:
		for _, arg := range p.term.(ast.Atom).Args {
			if v, ok := arg.(ast.Variable); ok && bound[v] {
				return true
			}
		}
	}
	return false
}

// bind adds the variables that the premise binds to bound.
func (p premise) bind(bound map[ast.Variable]bool) {
	switch p.kind {
	case atomPremise, builtinPremise, eqPremise:
		for _, v := range p.vars {
			bound[v] = true
		}
	}
}
