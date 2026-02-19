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
	"codeberg.org/TauCeti/mangle-go/ast"
	"codeberg.org/TauCeti/mangle-go/builtin"
	"codeberg.org/TauCeti/mangle-go/factstore"
	"codeberg.org/TauCeti/mangle-go/functional"
	"codeberg.org/TauCeti/mangle-go/unionfind"
)

func premiseAtom(a ast.Atom, lookupFn func(p ast.Atom, cb func(ast.Atom) error) error, subst unionfind.UnionFind) ([]unionfind.UnionFind, error) {
	p, err := functional.EvalAtom(a, subst)
	if err != nil {
		return nil, err
	}
	var solutions []unionfind.UnionFind
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
	// Not a built-in predicate. Call lookupFn.
	lookupFn(p, func(fact ast.Atom) error {
		// TODO: This could be made a lot more efficient by using a persistent
		// data structure for composing the unionfind substitutions.
		if newSubst, err := unionfind.UnifyTermsExtend(p.Args, fact.Args, subst); err == nil {
			solutions = append(solutions, newSubst)
		}
		return nil
	})
	return solutions, nil

}

// premiseNegAtom semi-positive evaluation, i.e. looks up a and fails if it is present.
func premiseNegAtom(a ast.Atom, store factstore.ReadOnlyFactStore, subst unionfind.UnionFind) ([]unionfind.UnionFind, error) {
	n, err := functional.EvalAtom(a, subst)
	if err != nil {
		return nil, err
	}

	if n.Predicate.IsBuiltin() {
		ok, _, err := builtin.Decide(n, &subst)
		if err != nil {
			return nil, err
		}
		if ok {
			return nil, nil // negated: no solution
		}
		return []unionfind.UnionFind{subst}, nil
	}
	solutions := []unionfind.UnionFind{subst}
	// If we find a single fact that unifies, then subst is not a solution.
	err = store.GetFacts(n, func(fact ast.Atom) error {
		if _, err := unionfind.UnifyTermsExtend(n.Args, fact.Args, subst); err == nil {
			solutions = nil
			return errBreak
		}
		return nil
	})
	if err != nil && err != errBreak {
		return nil, err
	}
	return solutions, nil
}

func premiseEq(left, right ast.BaseTerm, subst unionfind.UnionFind) ([]unionfind.UnionFind, error) {
	left, right, err := functional.EvalBaseTermPair(left, right, subst)
	if err != nil {
		return nil, err
	}
	nsubst, err := unionfind.UnifyTermsExtend([]ast.BaseTerm{left}, []ast.BaseTerm{right}, subst)
	if err != nil {
		return nil, nil // Ignore error
	}
	return []unionfind.UnionFind{nsubst}, nil
}

func premiseIneq(left, right ast.BaseTerm, subst unionfind.UnionFind) ([]unionfind.UnionFind, error) {
	left, right, err := functional.EvalBaseTermPair(left, right, subst)
	if err != nil {
		return nil, err
	}
	if _, err := unionfind.UnifyTermsExtend([]ast.BaseTerm{left}, []ast.BaseTerm{right}, subst); err != nil {
		// TODO: Check that error is indeed "cannot unify."
		return []unionfind.UnionFind{subst}, nil
	}
	return nil, nil
}
