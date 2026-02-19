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
	"testing"

	"codeberg.org/TauCeti/mangle-go/ast"
	"codeberg.org/TauCeti/mangle-go/factstore"
	"codeberg.org/TauCeti/mangle-go/functional"
)

func TestShortestPaths(t *testing.T) {
	store := factstore.NewSimpleInMemoryStore()
	clauses := []ast.Clause{
		clause("edge(/a, /b)."),
		clause("edge(/b, /c)."),
		clause("edge(/c, /d)."),
		clause("edge(/a, /d)."),

		// The shortest_path relation.
		clause("shortest_path(X, Y, [Y, X]) :- edge(X, Y)."),
		clause("shortest_path(X, Z, NewPath) :- " +
			"shortest_path(X, Y, Path), edge(Y, Z) |> let NewPath = fn:list:cons(Z, Path)."),

		// The merge predicate. This returns the join (least-upper bound)
		// of P1 and P2 in P. It need not be datalog (P is not bound), since
		// it is evaluated top-down.
		clause("shorter(P1, P2, P) :- fn:list:len(P1) < fn:list:len(P2), P = P1."),
		clause("shorter(P1, P2, P) :- fn:list:len(P2) <= fn:list:len(P1), P = P2."),
	}

	// We need to declare two things to enable custom joins:
	// - a functional dependency: X, Y determine P.
	// - a merge predicate name that provides the lattice-join operation.
	shortestPathDecl, err := ast.NewDecl(atom("shortest_path(X, Y, P)"), []ast.Atom{
		atom("fundep([X, Y], [P])"),
		atom("merge([P], 'shorter')"),
	}, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	// The merge predicate is marked deferred(). Such predicates have to come
	// with a mode declaration that specifies what the inputs and outputs are.
	// Merge predicates in particular must have mode('+', '+', '-').
	shorterDecl, err := ast.NewDecl(atom("shorter(P1, P2, P)"), []ast.Atom{
		atom("mode('+', '+', '-')"),
		atom("deferred()"),
	}, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	decls := []ast.Decl{shortestPathDecl, shorterDecl}
	if err := analyzeAndEvalProgramWithDecls(t, clauses, decls, store); err != nil {
		t.Errorf("Program evaluation failed %v program %v", err, program)
		return
	}
	wantShortPath, err := functional.EvalAtom(atom("shortest_path(/a, /d, [/d, /a])"), nil)
	if err != nil {
		t.Fatal(err)
	}
	if !store.Contains(wantShortPath) {
		t.Errorf("store does not contain %v that is shortest: %v", wantShortPath, store)
	}
	notWantLongPath, err := functional.EvalAtom(atom("shortest_path(/a, /d, [/d, /c, /b, /a])"), nil)
	if err != nil {
		t.Fatal(err)
	}
	if store.Contains(notWantLongPath) {
		t.Errorf("store contains path %v that is not shortest: %v", notWantLongPath, store)
	}
}
