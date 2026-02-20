// Copyright 2022 Google LLC
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

package analysis

import (
	"fmt"
	"strings"

	"codeberg.org/TauCeti/mangle-go/ast"
	"codeberg.org/TauCeti/mangle-go/builtin"
	"codeberg.org/TauCeti/mangle-go/symbols"
	"codeberg.org/TauCeti/mangle-go/unionfind"
)

// variablesForArgMode returns variable arguments that match an argmode mask.
func variablesForArgMode(atom ast.Atom, mode ast.Mode, mask ast.ArgMode) []ast.Variable {
	var boundVars []ast.Variable

	for i, argMode := range mode {
		if argMode&mask == 0 {
			continue
		}

		if v, ok := atom.Args[i].(ast.Variable); ok {
			boundVars = append(boundVars, v)
		}
	}

	return boundVars
}

// CheckRule checks arity and that every variable appearing is bound.
// A variable is bound when:
// - it appears in a positive atom, or
// - is unified (via an equality) with a constant or bound variable.
// These form the basis for datalog safety (ensuring termination).
// We permit mode declaration that modify these conditions.
// A variable is therefore bound when:
// - a mode declaration forces it as input
// - it appears as a column in a positive atom that is not declared as input, or
// - is unified (via an equality) with a constant or bound variable.
// Also checks that every function application expression has the right number of arguments.
func (a *Analyzer) CheckRule(clause ast.Clause) error {
	clause = clause.ReplaceWildcards()
	var (
		boundVars = make(map[ast.Variable]bool)
		headVars  = make(map[ast.Variable]bool)
		seenVars  = make(map[ast.Variable]bool)
	)
	ast.AddVars(clause.Head, headVars)
	ast.AddVars(clause.Head, seenVars)
	// Check that if the head predicate is temporal, the clause has a temporal annotation.
	if decl, ok := a.decl[clause.Head.Predicate]; ok {
		if (decl.IsTemporal() || decl.IsMaybeTemporal()) && clause.HeadTime == nil {
			return fmt.Errorf("temporal predicate %v defined without temporal annotation", clause.Head.Predicate)
		}
	}

	if clause.HeadTime != nil {
		if clause.HeadTime.Start.Type == ast.VariableBound {
			headVars[clause.HeadTime.Start.Variable] = true
			seenVars[clause.HeadTime.Start.Variable] = true
		}
		if clause.HeadTime.End.Type == ast.VariableBound {
			headVars[clause.HeadTime.End.Variable] = true
			seenVars[clause.HeadTime.End.Variable] = true
		}
	}
	uf := unionfind.New()

	if decl, ok := a.decl[clause.Head.Predicate]; ok {

		mode := unifyModes(decl.Modes())
		for _, v := range variablesForArgMode(clause.Head, mode, ast.ArgModeInput) {
			boundVars[v] = true
		}
	}

	if clause.Premises != nil {
		for _, premise := range clause.Premises {
			ast.AddVars(premise, seenVars)
			switch p := premise.(type) {
			case ast.Atom:
				if err := checkAtomArity(p); err != nil {
					return err
				}
				// This is after rewriting, so we can assume that evaluation proceeds left-to-right,
				// we only need to check that variables have been bound by some earlier subgoal.
				if !p.Predicate.IsBuiltin() {
					if decl, ok := a.decl[p.Predicate]; ok && len(decl.Modes()) > 0 {
						for _, v := range variablesForArgMode(p, unifyModes(decl.Modes()), ast.ArgModeOutput|ast.ArgModeInputOutput) {
							boundVars[v] = true
						}
					} else {
						ast.AddVars(p, boundVars)
					}

					// Validate that if the predicate is temporal, it is not used as a bare atom here.
					if decl, ok := a.decl[p.Predicate]; ok {
						if decl.IsTemporal() || decl.IsMaybeTemporal() {
							return fmt.Errorf("temporal predicate %v used without temporal annotation", p.Predicate)
						}
					}

					continue
				}
				// For builtin predicates, there is exactly one mode.
				var builtinVars = make(map[ast.Variable]bool)
				ast.AddVars(p, builtinVars)
				mode := builtin.Predicates[p.Predicate]
				if err := mode.Check(p, boundVars); err != nil {
					return err
				}
				for _, v := range variablesForArgMode(p, mode, ast.ArgModeOutput|ast.ArgModeInputOutput) {
					boundVars[v] = true
				}
				for v := range builtinVars {
					if !boundVars[v] {
						return fmt.Errorf("variable %v in %v will not have a value yet; move the subgoal to the right", v, p)
					}
				}
			case ast.TemporalLiteral:
				// Variables in the underlying literal are bound if it's an Atom
				if atom, ok := p.Literal.(ast.Atom); ok {
					ast.AddVars(atom, boundVars)

					// Validate that the predicate is temporal
					if !atom.Predicate.IsBuiltin() {
						if decl, ok := a.decl[atom.Predicate]; ok {
							if !decl.IsTemporal() && !decl.IsMaybeTemporal() {
								return fmt.Errorf("predicate %v is not declared temporal but used with temporal annotation in %v", atom.Predicate, p)
							}
						}
					}
				}
				// Variables in the interval annotation are also bound (output/binding)
				if p.Interval != nil {
					if p.Interval.Start.Type == ast.VariableBound {
						boundVars[p.Interval.Start.Variable] = true
					}
					if p.Interval.End.Type == ast.VariableBound {
						boundVars[p.Interval.End.Variable] = true
					}
				}
			case ast.Eq:
				if _, isconst := p.Left.(ast.Constant); isconst {
					if v, isvar := p.Right.(ast.Variable); isvar {
						boundVars[v] = true
						break
					}
				}
				if _, isconst := p.Right.(ast.Constant); isconst {
					if v, isvar := p.Left.(ast.Variable); isvar {
						boundVars[v] = true
						break
					}
				}
				if _, isapply := p.Left.(ast.ApplyFn); isapply {
					vars := make(map[ast.Variable]bool)
					ast.AddVars(p.Left, vars)
					for v := range vars {
						if !boundVars[v] {
							return fmt.Errorf("variable %v in apply expression %v not bound", v, p)
						}
					}
					if v, isvar := p.Right.(ast.Variable); isvar {
						boundVars[v] = true
						break
					}
				}
				if _, isapply := p.Right.(ast.ApplyFn); isapply {
					vars := make(map[ast.Variable]bool)
					ast.AddVars(p.Right, vars)
					for v := range vars {
						if !boundVars[v] {
							return fmt.Errorf("variable %v in apply expression %v not bound", v, p)
						}
					}
					if v, isvar := p.Left.(ast.Variable); isvar {
						boundVars[v] = true
						break
					}
				}

				if _, l := p.Left.(ast.Variable); l {
					if _, r := p.Right.(ast.Variable); r {
						var err error
						uf, err = unionfind.UnifyTermsExtend([]ast.BaseTerm{p.Left}, []ast.BaseTerm{p.Right}, uf)
						if err != nil {
							return err
						}
					}
				}
			}
		}
	}

	var (
		transformVarDefs = make(map[ast.Variable]bool)
		transformVarUses = make(map[ast.Variable]bool)
	)
	// If there is no transform, these maps remain empty.
	addTransformVars(clause.Transform, transformVarDefs, transformVarUses)

	// It is ok to write "let _ = fn:foo(...)" but not "fn:foo(_)".
	delete(transformVarDefs, ast.Variable{"_"})
	for v := range transformVarDefs {
		if boundVars[v] {
			return fmt.Errorf("the transform of clause %v redefines variable %v from rule body", clause, v)
		}
	}

	// Every variable encountered in the head or body has to be bound somewhere.
	for v := range seenVars {
		if headVars[v] && transformVarDefs[v] {
			// Head variable is defined in transform.
			continue
		}
		if boundVars[v] {
			// Head variable is range-restricted by positive atom or constant.
			continue
		}
		if x := uf.Get(v); x != nil {
			// Head variable is unified.
			if _, isconst := x.(ast.Constant); isconst {
				continue
			}
			if u, isvar := x.(ast.Variable); isvar {
				if _, ok := boundVars[u]; ok {
					continue
				}
			}
		}
		return fmt.Errorf("variable %v is not bound in %v", v, clause)
	}

	// Every variable use that we saw in transform has to be bound somewhere.
	for v := range transformVarUses {
		// In a transform, we can refer to any variable that appears in the rule.
		if seenVars[v] || transformVarDefs[v] {
			continue
		}
		return fmt.Errorf("variable %v used in transform %v does not appear in clause %v", v, *clause.Transform, clause)
	}

	if hasMultipleTransforms(clause) {
		return fmt.Errorf("composing multiple transforms not implemented yet %v", clause)
	}
	if groupByVars, ok := hasGroupByTransform(clause); ok {
		groupByStmt := clause.Transform.Statements[0]
		if groupByStmt.Var != nil {
			return fmt.Errorf("do-transforms cannot have a variable %v", clause)
		}
		// All variables defined within the transform so far.
		transformDefs := make(map[ast.Variable]bool)
		for _, stmt := range clause.Transform.Statements[1:] {
			if stmt.Var == nil {
				return fmt.Errorf("all statements following group by have to be let-statements %v", clause)
			}
			if !builtin.IsReducerFunction(stmt.Fn.Function) {
				uses := make(map[ast.Variable]bool)
				ast.AddVars(stmt.Fn, uses)
				for v := range uses {
					if !groupByVars[v] && !transformDefs[v] {
						return fmt.Errorf("in %v, variable %v in function %v must be either part of group_by or defined in the transform", clause, v, stmt.Fn)
					}
					// TODO: We could expose the rest of the bound variables as list values.
					// This is left for later, after type-checking transforms.
				}
			}
			transformDefs[*stmt.Var] = true
		}
		if len(groupByVars) != len(groupByStmt.Fn.Args) {
			return fmt.Errorf("each argument of group_by must be a distinct variable, got: %v", groupByStmt)
		}
		// All head variables have to either be part of group_by key or appear in a reducer application.
		for v := range headVars {
			if groupByVars[v] {
				continue
			}
			if transformVarDefs[v] {
				continue
			}
			return fmt.Errorf("head variable %v is neither part of group_by nor aggregated: %v", v, clause.Transform)
		}
	}
	if hasLetTransform(clause) {
		for _, stmt := range clause.Transform.Statements[1:] {
			if stmt.Var == nil {
				return fmt.Errorf("all statements in a let transform have to be let-statements %v", clause)
			}
			if _, ok := builtin.ReducerFunctions[stmt.Fn.Function]; ok {
				return fmt.Errorf("reducer applications %v not allowed in a let-transform %v", stmt.Fn, clause)
			}
		}
	}

	// Check that rules only reference predicates that are defined.
	if err := a.checkPredicates(clause); err != nil {
		return err
	}
	// Check that the RHS predicates are visible from the package of this clause.
	if err := a.checkVisibility(clause); err != nil {
		return err
	}
	// Check that transforms are defined over a relation.
	if clause.Transform != nil && clause.Premises == nil {
		return fmt.Errorf("cannot have a transform without a body %v", clause)
	}
	// Check that the arity (TODO: types) of function application works out.
	if err := a.checkFunctions(clause); err != nil {
		return err
	}

	return nil
}

func hasMultipleTransforms(clause ast.Clause) bool {
	return clause.Transform != nil && clause.Transform.Next != nil
}

func hasLetTransform(clause ast.Clause) bool {
	return clause.Transform != nil && clause.Transform.Statements[0].Var != nil
}

func hasGroupByTransform(clause ast.Clause) (map[ast.Variable]bool, bool) {
	if clause.Transform == nil {
		return nil, false
	}
	stmt := clause.Transform.Statements[0]
	if stmt.Fn.Function.Symbol != symbols.GroupBy.Symbol {
		return nil, false
	}
	vars := make(map[ast.Variable]bool)
	for _, arg := range stmt.Fn.Args {
		if v, ok := arg.(ast.Variable); ok {
			vars[v] = true
		}
	}
	return vars, true
}

func addTransformVars(transform *ast.Transform, vardefs map[ast.Variable]bool, varuse map[ast.Variable]bool) {
	if transform == nil {
		return
	}
	for _, transformStmt := range transform.Statements {
		if transformStmt.Var != nil {
			vardefs[*transformStmt.Var] = true
		}
		for _, baseTerm := range transformStmt.Fn.Args {
			ast.AddVars(baseTerm, varuse)
		}
	}
}

func checkAtomArity(atom ast.Atom) error {
	if atom.Predicate.Arity != len(atom.Args) {
		return fmt.Errorf("Arity mismatch: %s expects %d arguments but has %d in %v", atom.Predicate.Symbol, atom.Predicate.Arity, len(atom.Args), atom)
	}
	return nil
}

func (a *Analyzer) check(c func(sym ast.PredicateSym) error, clause ast.Clause) error {
	for _, p := range clause.Premises {
		switch p := p.(type) {
		case ast.Atom:
			if err := c(p.Predicate); err != nil {
				return err
			}
		case ast.NegAtom:
			if err := c(p.Atom.Predicate); err != nil {
				return err
			}
		default:
			continue
		}
	}
	return nil
}

func (a *Analyzer) checkPredicates(clause ast.Clause) error {
	return a.check(func(sym ast.PredicateSym) error {
		if _, ok := a.decl[sym]; ok {
			return nil
		}
		if _, ok := builtin.Predicates[sym]; ok {
			return nil
		}
		if len(a.extraPredicates) > 0 {
			if _, ok := a.extraPredicates[sym]; ok {
				return nil
			}
		}

		var arities []int
		for declared := range a.decl {
			if declared.Symbol == sym.Symbol {
				arities = append(arities, declared.Arity)
			}
		}
		if len(arities) > 0 {
			return fmt.Errorf(
				"in clause %q: predicate %s called with %d arguments, but available arities are: %v",
				clause, sym.Symbol, sym.Arity, arities)
		}
		return fmt.Errorf("in clause %q could not find predicate %v", clause, sym)
	}, clause)
}

func (a *Analyzer) checkVisibility(clause ast.Clause) error {
	var pkg string
	symbol := clause.Head.Predicate.Symbol
	if lastDot := strings.LastIndex(symbol, "."); lastDot != -1 {
		pkg = symbol[:lastDot]
	} else {
		pkg = ""
	}
	return a.check(func(sym ast.PredicateSym) error {
		d, ok := a.decl[sym]
		if !ok {
			// TODO: Invert default visibility.
			// No decl found. Assume public visibility.
			return nil
		}
		// Predicates defined in the same package are visible.
		if pkg == d.PackageID() {
			return nil
		}
		if !d.Visible() {
			return fmt.Errorf("predicate %q is not public", sym)
		}
		return nil
	}, clause)
}

func (a *Analyzer) checkExprArity(arg ast.BaseTerm) error {
	switch x := arg.(type) {
	case ast.Constant:
		return nil
	case ast.Variable:
		return nil
	case ast.ApplyFn:
		for _, arg := range x.Args {
			if err := a.checkExprArity(arg); err != nil {
				return err
			}
		}
		sym := x.Function
		lookup := func(sym ast.FunctionSym) (bool, bool) {
			_, builtinFun := builtin.Functions[sym]
			if builtinFun {
				return true, false
			}
			var extra bool
			if a.extraFunctions != nil {
				_, extra = a.extraFunctions[sym]
			}
			return false, extra
		}
		// Variable number of arguments.
		isBuiltinVar, isExtraVar := lookup(ast.FunctionSym{sym.Symbol, -1})
		if isExtraVar {
			return nil
		}
		if isBuiltinVar {
			// For var-arity reducer functions (e.g. fn:collect), check we have at least one argument.
			if _, ok := builtin.ReducerFunctions[ast.FunctionSym{sym.Symbol, -1}]; ok && len(x.Args) == 0 {
				return fmt.Errorf("reducer function %v expects at least one argument", sym.Symbol)
			}
			if sym.Symbol == symbols.Struct.Symbol && len(x.Args)%2 != 0 {
				// What if we are in a type expression? remove optional and repeated.
				return fmt.Errorf("expect even number of arguments for %s - use { /key: value, ... } syntax for structs", x)
			}
			if sym.Symbol == symbols.Map.Symbol && len(x.Args)%2 != 0 {
				return fmt.Errorf("expect even number of arguments for %s - use [ key: value, ... ] syntax for maps", x)
			}
			return nil
		}
		if isBuiltin, isExtra := lookup(sym); isBuiltin || isExtra {
			if sym.Arity == len(x.Args) {
				return nil
			}
			return fmt.Errorf("function %v expects %d arguments, provided: %v", sym.Symbol, sym.Arity, x.Args)
		}
		// Arity mismatch. Look for a symbol with the same name.
		for fn := range builtin.Functions {
			if fn.Symbol == sym.Symbol {
				return fmt.Errorf("wrong arity for function %v got %v (%d args) want %d args", fn, x.Args, len(x.Args), fn.Arity)
			}
		}
		if len(a.extraFunctions) > 0 {
			for fn := range a.extraFunctions {
				if fn.Symbol == sym.Symbol {
					return fmt.Errorf("wrong arity for function %v got %v (%d args) want %d args", fn, x.Args, len(x.Args), fn.Arity)
				}
			}
		}

		return fmt.Errorf("unknown function %v", sym)
	default:
		return fmt.Errorf("unexpected: %v", arg)
	}
}

func (a *Analyzer) checkFunctions(clause ast.Clause) error {
	// Just check arity. Types left for later.
	for _, p := range clause.Premises {
		switch x := p.(type) {
		case ast.Atom:
			for _, arg := range x.Args {
				if err := a.checkExprArity(arg); err != nil {
					return err
				}
			}
		case ast.NegAtom:
			for _, arg := range x.Atom.Args {
				if err := a.checkExprArity(arg); err != nil {
					return err
				}
			}
		case ast.Eq:
			if err := a.checkExprArity(x.Left); err != nil {
				return err
			}
			if err := a.checkExprArity(x.Right); err != nil {
				return err
			}
		case ast.Ineq:
			if err := a.checkExprArity(x.Left); err != nil {
				return err
			}
			if err := a.checkExprArity(x.Right); err != nil {
				return err
			}
		}
	}

	if clause.Transform == nil {
		return nil
	}
	for _, stmt := range clause.Transform.Statements {
		if err := a.checkExprArity(stmt.Fn); err != nil {
			return err
		}
	}
	return nil
}

// unifyModes takes multiple modes definition for the statement and merges them together into a single mode definition per argument.
// Example: single argument modes [+,+,?] result in ?, but [+,+] results in +.
func unifyModes(modes []ast.Mode) ast.Mode {
	if len(modes) == 0 {
		return ast.Mode{}
	}

	var unifiedMode []ast.ArgMode
	for i := 0; i < len(modes[0]); i++ {
		argMode := modes[0][i]

		for _, m := range modes[1:] {
			if argMode != m[i] {
				argMode = ast.ArgModeInputOutput
				break
			}
		}
		unifiedMode = append(unifiedMode, argMode)
	}

	return unifiedMode
}
