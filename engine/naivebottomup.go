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

package engine

import (
	"fmt"

	"codeberg.org/TauCeti/mangle-go/analysis"
	"codeberg.org/TauCeti/mangle-go/ast"
	"codeberg.org/TauCeti/mangle-go/builtin"
	"codeberg.org/TauCeti/mangle-go/factstore"
	"codeberg.org/TauCeti/mangle-go/functional"
	"codeberg.org/TauCeti/mangle-go/parse"
	"codeberg.org/TauCeti/mangle-go/unionfind"
)

type naiveEngine struct {
	store         factstore.FactStore
	programInfo   analysis.ProgramInfo
	strata        []analysis.Nodeset
	predToStratum map[ast.PredicateSym]int
}

// EvalProgramNaive evaluates a given program on the given facts, modifying the fact store in the process.
func EvalProgramNaive(program []ast.Clause, store factstore.SimpleInMemoryStore) error {
	preds := store.ListPredicates()
	knownPredicates := make(map[ast.PredicateSym]ast.Decl, len(preds))
	for _, sym := range preds {
		knownPredicates[sym] = ast.NewSyntheticDeclFromSym(sym)
	}
	programInfo, err := analysis.AnalyzeOneUnit(parse.SourceUnit{Clauses: program}, knownPredicates)
	if err != nil {
		return fmt.Errorf("analysis: %w", err)
	}
	strata, predToStratum, err := analysis.Stratify(analysis.Program{
		EdbPredicates: programInfo.EdbPredicates,
		IdbPredicates: programInfo.IdbPredicates,
		Rules:         programInfo.Rules,
	})
	if err != nil {
		return fmt.Errorf("stratification: %w", err)
	}
	naiveEngine{store, *programInfo, strata, predToStratum}.evalStrata()
	return nil
}

func (e naiveEngine) evalStrata() {
	for _, fact := range e.programInfo.InitialFacts {
		f, _ := functional.EvalAtom(fact, nil)
		e.store.Add(f)
	}
	for i := 0; i < len(e.strata); i++ {
		stratumEdbPredicates := make(map[ast.PredicateSym]struct{})
		stratumIdbPredicates := make(map[ast.PredicateSym]struct{})
		for sym, stratum := range e.predToStratum {
			if stratum < i {
				stratumEdbPredicates[sym] = struct{}{}
			} else if stratum == i {
				stratumIdbPredicates[sym] = struct{}{}
			}
		}
		var stratumRules []ast.Clause
		stratumDecls := make(map[ast.PredicateSym]*ast.Decl)
		for _, clause := range e.programInfo.Rules {
			sym := clause.Head.Predicate
			if _, ok := stratumIdbPredicates[sym]; ok {
				stratumRules = append(stratumRules, clause)
				stratumDecls[sym] = e.programInfo.Decls[sym]
			}
		}
		naiveEngine{
			store:       e.store,
			programInfo: analysis.ProgramInfo{stratumEdbPredicates, stratumIdbPredicates, nil, nil, stratumRules, stratumDecls, nil},
		}.eval()
	}
}

// This is very inefficient because it generates the same facts over and over again.
func (e naiveEngine) eval() {
	for {
		factadded := false
		for _, clause := range e.programInfo.Rules {
			if clause.Transform != nil {
				continue
			}
			for _, fact := range e.oneStepEvalClause(clause) {
				if e.store.Add(fact) {
					factadded = true
				}
			}
		}
		if !factadded {
			break
		}
	}
	for _, clause := range e.programInfo.Rules {
		if clause.Transform == nil {
			continue
		}
		internalPremise := clause.Premises[0].(ast.Atom)
		var substs []ast.ConstSubstList
		e.store.GetFacts(internalPremise, func(fact ast.Atom) error {
			var subst ast.ConstSubstList
			for i, baseTerm := range internalPremise.Args {
				v, _ := baseTerm.(ast.Variable)
				subst = subst.Extend(v, fact.Args[i].(ast.Constant))
			}
			substs = append(substs, subst)
			return nil
		})
		EvalTransform(clause.Head, *clause.Transform, substs, e.store.Add)
	}
}

// Evaluates clause, by scanning known facts for each premise and producing
// a solution (conjunctive query, similar to a join).
func (e naiveEngine) oneStepEvalClause(clause ast.Clause) []ast.Atom {
	var solutions = []unionfind.UnionFind{unionfind.New()}
	for _, term := range clause.Premises {
		var newsolutions []unionfind.UnionFind
		for _, s := range solutions {
			newsolutions = append(newsolutions, e.oneStepEvalPremise(term, s)...)
		}

		solutions = newsolutions
	}

	var facts []ast.Atom
	for _, sol := range solutions {
		facts = append(facts, clause.Head.ApplySubst(sol).(ast.Atom))
	}
	return facts
}

// Evaluates a single premise atom by scanning facts.
func (e naiveEngine) oneStepEvalPremise(premise ast.Term, subst unionfind.UnionFind) []unionfind.UnionFind {
	var solutions []unionfind.UnionFind
	switch p := premise.(type) {
	case ast.Atom:
		p, err := functional.EvalAtom(p, subst)
		if err != nil {
			return nil
		}
		if p.Predicate.IsBuiltin() {
			res, nsubsts, err := builtin.Decide(p, &subst)
			if err != nil {
				// Treat errors in built-in predicate evaluation as false.
				return nil
			}
			if !res {
				return nil
			}
			for _, nsubst := range nsubsts {
				solutions = append(solutions, *nsubst)
			}
			return solutions
		}
		// Not a built-in predicate.
		e.store.GetFacts(p, func(fact ast.Atom) error {
			// TODO: This could be made a lot more efficient by using a persistent
			// data structure for composing the unionfind substitutions.
			if newsubst, err := unionfind.UnifyTermsExtend(p.Args, fact.Args, subst); err == nil {
				solutions = append(solutions, newsubst)
			}
			return nil
		})
	case ast.NegAtom:
		a, err := functional.EvalAtom(p.Atom, subst)
		if err != nil {
			return nil
		}
		e.store.GetFacts(a, func(fact ast.Atom) error {
			if _, err := unionfind.UnifyTermsExtend(p.Atom.Args, fact.Args, subst); err != nil {
				solutions = append(solutions, subst)
			}
			return nil
		})
	case ast.Eq:
		if newsubst, err := unionfind.UnifyTermsExtend([]ast.BaseTerm{p.Left}, []ast.BaseTerm{p.Right}, subst); err == nil {
			solutions = append(solutions, newsubst)
		}
	case ast.Ineq:
		if _, err := unionfind.UnifyTermsExtend([]ast.BaseTerm{p.Left}, []ast.BaseTerm{p.Right}, subst); err != nil {
			solutions = append(solutions, subst)
		}
	}
	return solutions
}
