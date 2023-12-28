// Copyright 2023 Google LLC
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
	"github.com/google/mangle/ast"
	"github.com/google/mangle/factstore"
	"github.com/google/mangle/functional"
	"github.com/google/mangle/unionfind"
)

// QueryContext groups data needed for evaluating a query top-down (backward chaining).
type QueryContext struct {
	PredToRules map[ast.PredicateSym][]ast.Clause
	PredToDecl  map[ast.PredicateSym]*ast.Decl
	Store       factstore.ReadOnlyFactStore
}

// EvalQuery evaluates a query top-down, according to mode and union-find-subst.
// The mode must consist only of ArgModeInput (+) and ArgModeOutput (-).
// For every input, query.Args[i] is either a constant or a variable that
// is in the domain of subst.
func (q QueryContext) EvalQuery(query ast.Atom, mode []ast.ArgMode, uf unionfind.UnionFind, cb func(fact ast.Atom) error) error {
	for _, clause := range q.PredToRules[query.Predicate] {
		var vars []ast.BaseTerm
		var values []ast.BaseTerm
		for j, arg := range clause.Head.Args {
			v, ok := arg.(ast.Variable)
			if ok && mode[j] == ast.ArgModeInput {
				vars = append(vars, v)
				values = append(values, query.Args[j])
			}
		}

		subst, err := unionfind.UnifyTermsExtend(vars, values, uf)
		if err != nil {
			continue
		}
		sols := []unionfind.UnionFind{subst}
		for _, premise := range clause.Premises {
			var nsolsWorkList []unionfind.UnionFind
			for _, s := range sols {
				nsols, err := q.EvalPremise(premise, s)
				if err != nil {
					return err
				}
				nsolsWorkList = append(nsolsWorkList, nsols...)
			}
			sols = nsolsWorkList
		}
		if sols != nil {
			for _, s := range sols {
				result := clause.Head.ApplySubst(s).(ast.Atom)
				if err := cb(result); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// EvalPremise evaluates a single premise top-down.
// This is similar to PROLOG style SLD resolution: even though we
// have negated atoms, they are treated them as lookups (stratified semantics).
func (q QueryContext) EvalPremise(premise ast.Term, subst unionfind.UnionFind) ([]unionfind.UnionFind, error) {
	var solutions []unionfind.UnionFind
	switch p := premise.(type) {
	case ast.Atom:
		p, err := functional.EvalAtom(p, subst)
		if err != nil {
			return nil, err
		}
		decl := q.PredToDecl[p.Predicate]
		if decl != nil && decl.DeferredPredicate() {
			err := q.EvalQuery(p, decl.Modes()[0], subst, func(fact ast.Atom) error {
				newsubst, err := unionfind.UnifyTermsExtend(p.Args, fact.Args, subst)
				if err != nil {
					return err
				}
				solutions = append(solutions, newsubst)
				return nil
			})
			if err != nil {
				return nil, err
			}
			return solutions, nil
		}
		return premiseAtom(p, q.Store.GetFacts, subst)

	case ast.NegAtom:
		return premiseNegAtom(p.Atom, q.Store, subst)

	case ast.Eq:
		return premiseEq(p.Left, p.Right, subst)

	case ast.Ineq:
		return premiseIneq(p.Left, p.Right, subst)
	}
	return nil, nil
}
