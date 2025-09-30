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

// Package engine contains an implementation of the semi-naive evaluation strategy
// for datalog programs. It computes the fixpoint of the consequence operator incrementally
// by applying rules to known facts, taking care of consequences of already seen facts only once,
// until no new facts have been discovered.
package engine

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"go.uber.org/multierr"
	"github.com/google/mangle/analysis"
	"github.com/google/mangle/ast"
	"github.com/google/mangle/factstore"
	"github.com/google/mangle/functional"
	"github.com/google/mangle/rewrite"
	"github.com/google/mangle/unionfind"
)

const deltaStringPrefix = "Δ"

var errBreak = errors.New("break")

// Stats represents strata and their running times.
type Stats struct {
	Strata        [][]ast.PredicateSym
	Duration      []time.Duration
	PredToStratum map[ast.PredicateSym]int
}

type engine struct {
	store         factstore.FactStore
	deltaStore    factstore.FactStore
	programInfo   *analysis.ProgramInfo
	strata        []analysis.Nodeset
	predToStratum map[ast.PredicateSym]int
	predToRules   map[ast.PredicateSym][]ast.Clause
	predToDecl    map[ast.PredicateSym]*ast.Decl
	stats         Stats
	options       EvalOptions
}

// ExternalPredicateCallback is used to query external data sources.
//
// An atom `mydb(input1, ..., inputN, OutputVar1, ..., OutputVarN)` is evaluated
// as follows:
//   - the engine checks whether the fact store contains any facts
//     of the shape `mydb(input1, ..., inputN, _, ..., _)`. If so, we
//     use those for evaluation.
//   - if no facts were found, the engine calls `ShouldQuery(input1, ..., inputN)`
//     if false, evaluation continues.
//   - if true, the engine calls `Query(input1, ..., inputN, filter1, ..., filterM)`
//     and expects outputs callback. Every output tuple gets added
//     as `mydb(input1, ..., inputN, output1, ..., outputN)`
//     fact to the store, and continues evaluation.
//     if ExecuteQuery returns an error, evaluation fails with that error.
//
// If tuples (input1, ..., inputN) are known to yield empty results, the
// implementation can keep track of that and prevent an unnecessary
// call to `ExecuteQuery`.
// filters contains the arguments that are output positions which are either
// variables or constants that are used to match the position (filters proper).
// The implementation may use the constant filter arguments for filter-pushdown,
// but is also free to ignore them. In any case, when constant filter arguments
// are present, only matching facts will be added to the store.
type ExternalPredicateCallback interface {
	// If true, the engine will pass any subgoals of the clause that mention output variables
	// to ExecuteQuery. Otherwise, the pushdown argument will be empty.
	ShouldPushdown() bool
	ShouldQuery(inputs []ast.Constant, filters []ast.BaseTerm, pushdown []ast.Term) bool
	ExecuteQuery(inputs []ast.Constant, filters []ast.BaseTerm, pushdown []ast.Term,
		cb func([]ast.BaseTerm)) error
}

// EvalOptions are used to configure the evaluation.
type EvalOptions struct {
	createdFactLimit int
	totalFactLimit   int
	// if non-nil, only predicates in this allowlist get evaluated.
	predicateAllowList *func(ast.PredicateSym) bool
	externalPredicates map[ast.PredicateSym]ExternalPredicateCallback
}

// EvalOption affects the way the evaluation is performed.
type EvalOption func(*EvalOptions)

// WithCreatedFactLimit is an evaluation option that limits the maximum number of facts created during evaluation.
func WithCreatedFactLimit(limit int) EvalOption {
	return func(o *EvalOptions) { o.createdFactLimit = limit }
}

// WithExternalPredicates allows the user to provide callbacks for external predicates.
func WithExternalPredicates(
	callbacks map[ast.PredicateSym]ExternalPredicateCallback) EvalOption {
	return func(o *EvalOptions) { o.externalPredicates = callbacks }
}

// EvalProgram evaluates a given program on the given facts, modifying the fact store in the process.
// Deprecated: use EvalStratifiedProgramWithStats instead.
func EvalProgram(programInfo *analysis.ProgramInfo, store factstore.FactStore, options ...EvalOption) error {
	_, err := EvalProgramWithStats(programInfo, store, options...)
	return err
}

func newEvalOptions(options ...EvalOption) EvalOptions {
	ops := EvalOptions{}
	allPredicates := func(ast.PredicateSym) bool {
		return true
	}
	ops.predicateAllowList = &allPredicates
	ops.externalPredicates = make(map[ast.PredicateSym]ExternalPredicateCallback)
	for _, o := range options {
		o(&ops)
	}
	return ops
}

// EvalProgramWithStats evaluates a given program on the given facts, modifying the fact store in the process.
// Deprecated: use EvalStratifiedProgramWithStats instead.
func EvalProgramWithStats(programInfo *analysis.ProgramInfo, store factstore.FactStore, options ...EvalOption) (Stats, error) {
	strata, predToStratum, err := analysis.Stratify(analysis.Program{
		EdbPredicates: programInfo.EdbPredicates,
		IdbPredicates: programInfo.IdbPredicates,
		Rules:         programInfo.Rules,
	})
	if err != nil {
		return Stats{}, fmt.Errorf("stratification: %w", err)
	}
	return EvalStratifiedProgramWithStats(programInfo, strata, predToStratum, store, options...)
}

// EvalStratifiedProgramWithStats evaluates a given stratified program on the given facts,
// modifying the fact store in the process.
func EvalStratifiedProgramWithStats(programInfo *analysis.ProgramInfo,
	strata []analysis.Nodeset, predToStratum map[ast.PredicateSym]int,
	store factstore.FactStore, options ...EvalOption) (Stats, error) {

	predToRules := make(map[ast.PredicateSym][]ast.Clause)
	predToDecl := make(map[ast.PredicateSym]*ast.Decl)
	for sym := range programInfo.Decls {
		predToDecl[sym] = programInfo.Decls[sym]
	}
	for _, clause := range programInfo.Rules {
		sym := clause.Head.Predicate
		predToRules[sym] = append(predToRules[sym], clause)
	}
	stats := Stats{
		Strata:        make([][]ast.PredicateSym, len(strata), len(strata)),
		Duration:      make([]time.Duration, len(strata), len(strata)),
		PredToStratum: predToStratum,
	}
	for sym, stratum := range predToStratum {
		stats.Strata[stratum] = append(stats.Strata[stratum], sym)
	}
	opts := newEvalOptions(options...)
	for sym := range opts.externalPredicates {
		decl := predToDecl[sym]
		if decl == nil {
			return Stats{}, fmt.Errorf("ext callback for a predicate %v without decl", sym)
		}
		if !decl.IsExternal() {
			return Stats{}, fmt.Errorf("ext callback for predicate %v that is not marked as external()", sym)
		}
	}
	if opts.createdFactLimit > 0 {
		opts.totalFactLimit = store.EstimateFactCount() + opts.createdFactLimit
	}
	e := &engine{store, factstore.NewMultiIndexedArrayInMemoryStore(), programInfo, strata,
		predToStratum, predToRules, predToDecl, stats, opts}
	if err := e.evalStrata(); err != nil {
		return Stats{}, err
	}
	return e.stats, nil
}

// evalStrata runs the evaluation for the layers.
func (e *engine) evalStrata() error {
	predicateAllowList := *e.options.predicateAllowList
	for _, fact := range e.programInfo.InitialFacts {
		if !predicateAllowList(fact.Predicate) {
			continue
		}
		f, err := functional.EvalAtom(fact, nil)
		if err != nil {
			return err
		}
		e.store.Add(f)
	}
	for i := 0; i < len(e.strata); i++ {
		stratumEdbPredicates := make(map[ast.PredicateSym]struct{})
		for j := 0; j < i; j++ {
			for _, sym := range e.stats.Strata[j] {
				stratumEdbPredicates[sym] = struct{}{}
			}
		}
		stratumIdbPredicates := make(map[ast.PredicateSym]struct{})
		var stratumRules []ast.Clause
		stratumDecls := make(map[ast.PredicateSym]*ast.Decl)
		for _, sym := range e.stats.Strata[i] {
			stratumIdbPredicates[sym] = struct{}{}
			stratumRules = append(stratumRules, e.predToRules[sym]...)
			stratumDecls[sym] = e.predToDecl[sym]
		}
		stratifiedProgram := rewrite.Rewrite(analysis.Program{stratumEdbPredicates, stratumIdbPredicates, stratumRules})
		start := time.Now()
		e := engine{
			store:         e.store,
			deltaStore:    factstore.NewMultiIndexedArrayInMemoryStore(),
			programInfo:   &analysis.ProgramInfo{stratifiedProgram.EdbPredicates, stratifiedProgram.IdbPredicates, nil, stratifiedProgram.Rules, stratumDecls},
			predToStratum: e.predToStratum,
			predToRules:   e.predToRules,
			predToDecl:    e.predToDecl,
			stats:         e.stats,
			options:       e.options,
		}
		if err := e.eval(); err != nil {
			return err
		}
		e.stats.Duration[i] = time.Since(start)
	}
	return nil
}

func makeDeltaPredicate(pred ast.PredicateSym) ast.PredicateSym {
	return ast.PredicateSym{deltaStringPrefix + pred.Symbol, pred.Arity}
}

func makeDeltaAtom(atom ast.Atom) ast.Atom {
	return ast.Atom{makeDeltaPredicate(atom.Predicate), atom.Args}
}

func isDeltaPredicate(pred ast.PredicateSym) bool {
	return strings.HasPrefix(pred.Symbol, deltaStringPrefix)
}

func makeNormalPredicate(pred ast.PredicateSym) ast.PredicateSym {
	return ast.PredicateSym{strings.TrimPrefix(pred.Symbol, deltaStringPrefix), pred.Arity}
}

func makeNormalAtom(atom ast.Atom) ast.Atom {
	return ast.Atom{makeNormalPredicate(atom.Predicate), atom.Args}
}

// makeSingleDeltaRule turns rule into a delta rule, with i-th subgoal used as "delta subgoal."
// The i-th subgoal must be a positive atom.
func makeSingleDeltaRule(rule ast.Clause, i int) ast.Clause {
	var newpremises []ast.Term

	for j, subgoal := range rule.Premises {
		if i == j {
			atom, _ := subgoal.(ast.Atom)
			newpremises = append(newpremises, makeDeltaAtom(atom))
		} else {
			newpremises = append(newpremises, subgoal)
		}
	}
	clause := ast.NewClause(rule.Head, newpremises)
	clause.Transform = rule.Transform
	return clause
}

// makeDeltaRules takes all rules of all predicates and creates delta rules for each of them.
// A delta rule for R checks whether a newly added fact led to derivation of a new fact via R.
func makeDeltaRules(decls map[ast.PredicateSym]*ast.Decl, predToRules map[ast.PredicateSym][]ast.Clause) map[ast.PredicateSym][]ast.Clause {
	predToDeltaRules := make(map[ast.PredicateSym][]ast.Clause)
	for _, decl := range decls {
		pred := decl.DeclaredAtom.Predicate
		rules := predToRules[pred]
		for _, clause := range rules {
			if clause.Transform != nil && !clause.Transform.IsLetTransform() {
				// Rules with do-transforms are only applied at the very end.
				continue
			}
			var deltaRules []ast.Clause
			for i, subgoal := range clause.Premises {
				// We want one delta rule for each subgoal that can match a positive atoms
				// produced exactly in the last round.
				p, ok := subgoal.(ast.Atom)
				if !ok || p.Predicate.IsBuiltin() {
					continue
				}
				subgoalPred := p.Predicate
				if _, ok := decls[subgoalPred]; ok {
					deltaRule := makeSingleDeltaRule(clause, i)
					deltaRules = append(deltaRules, deltaRule)
				}
			}
			predToDeltaRules[pred] = append(predToDeltaRules[pred], deltaRules...)
		}
	}
	return predToDeltaRules
}

func (e *engine) hasMergePredicate(pred ast.PredicateSym) (ast.FunDep, ast.PredicateSym, bool) {
	decl := e.predToDecl[pred]
	if decl == nil {
		return ast.FunDep{}, ast.PredicateSym{}, false
	}
	fundeps := decl.FunDeps()
	if len(fundeps) != 1 {
		return ast.FunDep{}, ast.PredicateSym{}, false
	}
	_, mergePred := decl.MergePredicate()
	return fundeps[0], mergePred, true
}

var mergePredMode = []ast.ArgMode{ast.ArgModeInput, ast.ArgModeInput, ast.ArgModeOutput}

// mergeDelta updates e.store with facts from e.deltaStore.
// For facts with custom lattice join operations, replaces facts instead of adding.
func (e *engine) mergeDelta() error {
	err := factstore.GetAllFacts(e.deltaStore, func(fact ast.Atom) error {
		pred := fact.Predicate
		fundep, mergePred, ok := e.hasMergePredicate(pred)
		if !ok {
			// Default case: just add the new fact.
			e.store.Add(fact)
			return nil
		}

		// Merge-predicate case: add fact or replace existing fact.
		// TODO: Generalize to merge predicate with n * 3 columns.
		if len(fundep.Target) != 1 {
			return fmt.Errorf("merging with |target vars| != 1 not implemented: %v", fundep.Target)
		}
		targetColumn := fundep.Target[0]

		// Query existing facts whose columns agree on fundep.Source values.
		queryArgs := make([]ast.BaseTerm, pred.Arity, pred.Arity)
		for i := 0; i < pred.Arity; i++ {
			queryArgs[i] = ast.Variable{"_"}
		}
		for i := range fundep.Source {
			queryArgs[i] = fact.Args[i]
		}
		queryExisting := ast.Atom{pred, queryArgs}
		existing := false
		e.store.GetFacts(queryExisting, func(existingFact ast.Atom) error {
			existing = true
			if fact.Equals(existingFact) {
				return nil // nothing to do.
			}

			// Evaluate merge predicate (top-down) to construct replacement fact.
			merged := false

			// Prepare top-down query with merge predicate.
			mergeQuery := ast.Atom{Predicate: mergePred, Args: []ast.BaseTerm{
				existingFact.Args[targetColumn],

				fact.Args[targetColumn],
				ast.Variable{"_"},
			}}
			err := e.newContext().EvalQuery(mergeQuery, mergePredMode, unionfind.New(), func(mergeFact ast.Atom) error {
				merged = true
				value := mergeFact.Args[2]
				fact.Args[fundep.Target[0]] = value
				if !existingFact.Equals(fact) {
					if storeWithRemove, ok := e.store.(factstore.FactStoreWithRemove); ok {
						storeWithRemove.Remove(existingFact)
					}
					e.store.Add(fact)
					return errBreak
				}
				return nil
			})
			if err == errBreak {
				return nil // Already added.
			}
			if err != nil && err != errBreak {
				return err
			}
			if !merged {
				e.store.Add(fact) // fact and existingFact are incomparable.
			}
			return nil
		})
		if !existing {
			e.store.Add(fact)
		}
		return nil
	})
	return err
}

func (e *engine) eval() error {
	predicateAllowList := *e.options.predicateAllowList
	// First round.
	for _, clause := range e.programInfo.Rules {
		if !predicateAllowList(clause.Head.Predicate) {
			continue
		}
		if clause.Transform != nil && !clause.Transform.IsLetTransform() {
			// clauses with do-transforms assume a single subgoal as body.
			continue
		}
		facts, err := e.oneStepEvalClause(clause)
		if err != nil {
			return err
		}
		for _, fact := range facts {
			e.deltaStore.Add(fact)
		}
	}
	if e.deltaStore.EstimateFactCount() > 0 {
		// Incremental rounds.
		deltaRules := makeDeltaRules(e.programInfo.Decls, e.predToRules)
		if err := e.mergeDelta(); err != nil {
			return err
		}
		for {
			newDeltaStore := factstore.NewMultiIndexedArrayInMemoryStore()
			var incrementalFactAdded bool
			for _, predDeltaRule := range deltaRules {
				for _, deltaRule := range predDeltaRule {
					if !predicateAllowList(deltaRule.Head.Predicate) {
						continue
					}
					facts, err := e.oneStepEvalClause(deltaRule)
					if err != nil {
						return err
					}
					for _, fact := range facts {
						if !e.store.Contains(fact) && !e.deltaStore.Contains(fact) {
							incrementalFactAdded = newDeltaStore.Add(fact) || incrementalFactAdded
						}
						if e.options.createdFactLimit > 0 && newDeltaStore.EstimateFactCount() > e.options.createdFactLimit {
							return fmt.Errorf("fact size limit reached evaluating %q %d > %d", deltaRule.String(), newDeltaStore.EstimateFactCount(), e.options.createdFactLimit)
						}
					}
				}
			}
			if err := e.mergeDelta(); err != nil {
				return err
			}
			if e.options.totalFactLimit > 0 && e.store.EstimateFactCount() > e.options.totalFactLimit {
				return fmt.Errorf("fact size limit reached %d > %d", e.store.EstimateFactCount(), e.options.totalFactLimit)
			}
			e.deltaStore = newDeltaStore
			if !incrementalFactAdded {
				break
			}
		}
	}
	// We reached the fixed point and can now apply "do-transforms".
	for _, clause := range e.programInfo.Rules {
		if clause.Transform == nil || clause.Transform.IsLetTransform() {
			continue
		}
		internalPremise, ok := clause.Premises[0].(ast.Atom)
		if !ok {
			return fmt.Errorf("expected first premise of clause: %v to be an atom %v", clause, clause.Premises[0])
		}
		var substs []ast.ConstSubstList
		e.store.GetFacts(internalPremise, func(fact ast.Atom) error {
			var subst ast.ConstSubstList
			for i, baseTerm := range internalPremise.Args {
				if v, ok := baseTerm.(ast.Variable); ok {
					if c, ok := fact.Args[i].(ast.Constant); ok {
						subst = subst.Extend(v, c)
					}
				}
			}
			substs = append(substs, subst)
			return nil
		})
		var merr error
		if err := EvalTransform(clause.Head, *clause.Transform, substs, func(a ast.Atom) bool {
			a, err := functional.EvalAtom(a, ast.ConstSubstList{})
			if err != nil {
				merr = multierr.Append(merr, err)
				return false
			}
			return e.store.Add(a)
		}); err != nil {
			return err
		}
		if merr != nil {
			return merr
		}
	}
	return nil
}

// Evaluates clause (a rule), by scanning known facts for each premise and producing
// a solution (conjunctive query, similar to a join).
func (e *engine) oneStepEvalClause(clause ast.Clause) ([]ast.Atom, error) {
	pred := clause.Head.Predicate
	decl := e.predToDecl[pred]
	if decl != nil && decl.DeferredPredicate() {
		return nil, nil
	}

	var solutions = []unionfind.UnionFind{unionfind.New()}
	for _, term := range clause.Premises {
		var newsolutions []unionfind.UnionFind
		for _, s := range solutions {
			stepsolutions, err := e.oneStepEvalPremise(term, s, clause)
			if err != nil {
				return nil, err
			}
			newsolutions = append(newsolutions, stepsolutions...)
			if e.options.createdFactLimit > 0 && len(newsolutions) > e.options.createdFactLimit {
				return nil, fmt.Errorf("fact size limit reached %q %d > %d", clause.Head.String(), len(newsolutions), e.options.createdFactLimit)
			}
		}

		solutions = newsolutions
	}

	var facts []ast.Atom
	for _, sol := range solutions {
		head, err := functional.EvalAtom(clause.Head, sol)
		if err != nil {
			return nil, err
		}
		if clause.Transform == nil {
			facts = append(facts, head)
			continue
		}

		if err := EvalTransform(head, *clause.Transform, []ast.ConstSubstList{sol.AsConstSubstList()}, func(a ast.Atom) bool {
			facts = append(facts, a)
			return true
		}); err != nil {
			return nil, err
		}
		if e.options.totalFactLimit > 0 && e.store.EstimateFactCount() > e.options.totalFactLimit {
			return nil, fmt.Errorf("fact size limit reached evaluting %q %d > %d", clause.Head.String(), e.store.EstimateFactCount(), e.options.totalFactLimit)
		}
	}
	return facts, nil
}

// Evaluates a single premise atom by scanning facts.
func (e *engine) oneStepEvalPremise(premise ast.Term, subst unionfind.UnionFind, clause ast.Clause) ([]unionfind.UnionFind, error) {
	switch p := premise.(type) {
	case ast.Atom:
		if ext, ok := e.options.externalPredicates[p.Predicate]; ok {
			// We may make an external call and add a whole bunch of facts as side-effect.
			// This will be transparent to the rest of evaluation.
			decl := e.predToDecl[p.Predicate]
			if decl == nil {
				return nil, fmt.Errorf("no decl for predicate %v", p.Predicate)
			}
			mode := decl.Modes()[0]
			var pushdown []ast.Term
			if ext.ShouldPushdown() {
				pushdown = getPushdown(p, mode, clause, subst)
			}
			err := e.newContext().EvalExternalQuery(
				p, mode, ext, pushdown, func(fact ast.Atom) error {
					e.store.Add(fact)
					return nil
				})
			if err != nil {
				return nil, err
			}
		}

		var lookupFn func(p ast.Atom, cb func(ast.Atom) error) error
		if isDeltaPredicate(p.Predicate) {
			lookupFn = func(p ast.Atom, cb func(ast.Atom) error) error {
				return e.deltaStore.GetFacts(makeNormalAtom(p), cb)
			}
		} else {
			lookupFn = e.store.GetFacts
		}
		decl := e.predToDecl[p.Predicate]
		if decl != nil && decl.DeferredPredicate() {
			return e.newContext().EvalPremise(p, subst)
		}
		return premiseAtom(p, lookupFn, subst)

	case ast.NegAtom:
		return premiseNegAtom(p.Atom, e.store, subst)

	case ast.Eq:
		return premiseEq(p.Left, p.Right, subst)

	case ast.Ineq:
		return premiseIneq(p.Left, p.Right, subst)

	}
	return nil, nil
}

func (e *engine) newContext() QueryContext {
	return QueryContext{PredToRules: e.predToRules, PredToDecl: e.predToDecl, Store: e.store,
		ExternalPredicates: e.options.externalPredicates}
}

func getPushdown(premise ast.Atom, mode []ast.ArgMode, clause ast.Clause, subst unionfind.UnionFind) []ast.Term {
	// Find all output variables of the external predicate.
	var outputVars []ast.Variable
	for i, m := range mode {
		if m == ast.ArgModeOutput {
			if v, ok := premise.Args[i].(ast.Variable); ok && v.Symbol != "_" {
				outputVars = append(outputVars, v)
			}
		}
	}
	// When a substituted subgoal mentions this variable, add it to pushdown.
	var pushdown []ast.Term
	for _, subgoal := range clause.Premises {
		if subgoal.Equals(premise) {
			continue
		}
		if _, ok := subgoal.(ast.Atom); ok {
			subgoal = subgoal.ApplySubst(subst)
			freeVars := make(map[ast.Variable]bool)
			ast.AddVars(premise, freeVars)
			for _, v := range outputVars {
				if freeVars[v] {
					pushdown = append(pushdown, subgoal)
				}
			}
		}
	}
	return pushdown
}
