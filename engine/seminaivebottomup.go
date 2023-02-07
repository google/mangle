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
	"github.com/google/mangle/builtin"
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

// EvalOptions are used to configure the evaluation.
type EvalOptions struct {
	createdFactLimit int
	totalFactLimit   int
}

// EvalOption affects the way the evaluation is performed.
type EvalOption func(*EvalOptions)

// WithCreatedFactLimit is an evalution option that limits the maximum number of facts created during evaluation.
func WithCreatedFactLimit(limit int) EvalOption {
	return func(o *EvalOptions) { o.createdFactLimit = limit }
}

// EvalProgram evaluates a given program on the given facts, modifying the fact store in the process.
func EvalProgram(programInfo *analysis.ProgramInfo, store factstore.FactStore, options ...EvalOption) error {
	_, err := EvalProgramWithStats(programInfo, store, options...)
	return err
}

func newEvalOptions(options ...EvalOption) EvalOptions {
	ops := EvalOptions{}
	for _, o := range options {
		o(&ops)
	}
	return ops
}

// EvalProgramWithStats evaluates a given program on the given facts, modifying the fact store in the process.
func EvalProgramWithStats(programInfo *analysis.ProgramInfo, store factstore.FactStore, options ...EvalOption) (Stats, error) {
	strata, predToStratum, err := analysis.Stratify(analysis.Program{
		EdbPredicates: programInfo.EdbPredicates,
		IdbPredicates: programInfo.IdbPredicates,
		Rules:         programInfo.Rules,
	})
	if err != nil {
		return Stats{}, fmt.Errorf("stratification: %w", err)
	}
	predToRules := make(map[ast.PredicateSym][]ast.Clause)
	predToDecl := make(map[ast.PredicateSym]*ast.Decl)
	for _, clause := range programInfo.Rules {
		sym := clause.Head.Predicate
		predToRules[sym] = append(predToRules[sym], clause)
		predToDecl[sym] = programInfo.Decls[sym]
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
	if opts.createdFactLimit > 0 {
		opts.totalFactLimit = store.EstimateFactCount() + opts.createdFactLimit
	}
	e := &engine{store, factstore.NewMultiIndexedInMemoryStore(), programInfo, strata,
		predToStratum, predToRules, predToDecl, stats, opts}
	if err := e.evalStrata(); err != nil {
		return Stats{}, err
	}
	return e.stats, nil
}

func (e *engine) evalStrata() error {
	for _, fact := range e.programInfo.InitialFacts {
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
			deltaStore:    factstore.NewMultiIndexedInMemoryStore(),
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

func (e *engine) eval() error {
	// First round.
	for _, clause := range e.programInfo.Rules {
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
		for {
			newDeltaStore := factstore.NewMultiIndexedInMemoryStore()
			var incrementalFactAdded bool
			for _, predDeltaRule := range deltaRules {
				for _, deltaRule := range predDeltaRule {
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
			e.store.Merge(e.deltaStore)
			if e.options.totalFactLimit > 0 && e.store.EstimateFactCount() > e.options.totalFactLimit {
				return fmt.Errorf("fact size limit reached %d > %d", e.store.EstimateFactCount(), e.options.totalFactLimit)
			}
			e.deltaStore = newDeltaStore
			if !incrementalFactAdded {
				e.store.Merge(e.deltaStore)
				break
			}
		}
	}
	// We reached the fixed point can now apply "do-transforms".
	for _, clause := range e.programInfo.Rules {
		if clause.Transform == nil {
			continue
		}
		internalPremise := clause.Premises[0].(ast.Atom)
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
		EvalTransform(clause.Head, *clause.Transform, substs, func(a ast.Atom) bool {
			a, err := functional.EvalAtom(a, ast.ConstSubstList{})
			if err != nil {
				merr = multierr.Append(merr, err)
				return false
			}
			return e.store.Add(a)
		})
		if merr != nil {
			return merr
		}
	}
	return nil
}

// Evaluates clause, by scanning known facts for each premise and producing
// a solution (conjunctive query, similar to a join).
func (e *engine) oneStepEvalClause(clause ast.Clause) ([]ast.Atom, error) {
	var solutions = []unionfind.UnionFind{unionfind.New()}
	for _, term := range clause.Premises {
		var newsolutions []unionfind.UnionFind
		for _, s := range solutions {
			stepsolutions, err := e.oneStepEvalPremise(term, s)
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

		EvalTransform(head, *clause.Transform, []ast.ConstSubstList{sol.AsConstSubstList()}, func(a ast.Atom) bool {
			facts = append(facts, a)
			return true
		})
		if e.options.totalFactLimit > 0 && e.store.EstimateFactCount() > e.options.totalFactLimit {
			return nil, fmt.Errorf("fact size limit reached evaluting %q %d > %d", clause.Head.String(), e.store.EstimateFactCount(), e.options.totalFactLimit)
		}
	}
	return facts, nil
}

// Evaluates a single premise atom by scanning facts.
func (e *engine) oneStepEvalPremise(premise ast.Term, subst unionfind.UnionFind) ([]unionfind.UnionFind, error) {
	var solutions []unionfind.UnionFind
	switch p := premise.(type) {
	case ast.Atom:
		p, err := functional.EvalAtom(p, subst)
		if err != nil {
			return nil, err
		}
		if p.Predicate.IsBuiltin() {
			ok, nsubsts, err := builtin.Decide(p, &subst)
			if err != nil {
				return nil, err
			}
			if !ok {
				return nil, nil // no solution
			}
			for _, nsubst := range nsubsts {
				solutions = append(solutions, *nsubst)
			}
			return solutions, nil

		}
		// Not a built-in predicate.
		cb := func(fact ast.Atom) error {
			// TODO: This could be made a lot more efficient by using a persistent
			// data structure for composing the unionfind substitutions.
			if newsubst, err := unionfind.UnifyTermsExtend(p.Args, fact.Args, subst); err == nil {
				solutions = append(solutions, newsubst)
			}
			return nil
		}
		if isDeltaPredicate(p.Predicate) {
			e.deltaStore.GetFacts(makeNormalAtom(p), cb)
		} else {
			e.store.GetFacts(p, cb)
		}
		return solutions, nil

	case ast.NegAtom:
		n, err := functional.EvalAtom(p.Atom, subst)
		if err != nil {
			return nil, err
		}

		// If we find a single fact that unifies, then subst is not a solution.
		err = e.store.GetFacts(n, func(fact ast.Atom) error {
			if _, err := unionfind.UnifyTermsExtend(n.Args, fact.Args, subst); err == nil {
				solutions = nil
				return errBreak
			}
			return nil
		})
		if err == nil {
			return []unionfind.UnionFind{subst}, nil
		}
	case ast.Eq:
		left, right, err := functional.EvalBaseTermPair(p.Left, p.Right, subst)
		if err != nil {
			return nil, err
		}
		p = ast.Eq{left, right}
		nsubst, err := unionfind.UnifyTermsExtend([]ast.BaseTerm{p.Left}, []ast.BaseTerm{p.Right}, subst)
		if err != nil {
			return nil, nil // Ignore error
		}
		return []unionfind.UnionFind{nsubst}, nil

	case ast.Ineq:
		left, right, err := functional.EvalBaseTermPair(p.Left, p.Right, subst)
		if err != nil {
			return nil, err
		}
		p = ast.Ineq{left, right}
		if _, err := unionfind.UnifyTermsExtend([]ast.BaseTerm{p.Left}, []ast.BaseTerm{p.Right}, subst); err != nil {
			// TODO: Check that error is indeed "cannot unify."
			return []unionfind.UnionFind{subst}, nil
		}
	}
	return nil, nil
}
