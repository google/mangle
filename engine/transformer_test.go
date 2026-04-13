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
	"testing"

	"codeberg.org/TauCeti/mangle-go/ast"
	"codeberg.org/TauCeti/mangle-go/factstore"
	"codeberg.org/TauCeti/mangle-go/parse"
)

type testCase struct {
	initialFacts  []ast.Atom
	clause        ast.Clause
	expectedFacts []ast.Atom
}

func runEval(test testCase, t *testing.T) {
	store := factstore.NewSimpleInMemoryStore()
	for _, fact := range test.initialFacts {
		store.Add(fact)
	}

	var input []ast.ConstSubstList
	premise := test.clause.Premises[0].(ast.Atom)
	store.GetFacts(premise, func(atom ast.Atom) error {
		var subst ast.ConstSubstList
		for i, arg := range premise.Args {
			subst = subst.Extend(arg.(ast.Variable), atom.Args[i].(ast.Constant))
		}
		input = append(input, subst)
		return nil
	})
	EvalTransform(test.clause.Head, *test.clause.Transform, input, store.Add)
	for _, fact := range test.expectedFacts {
		if !store.Contains(fact) {
			t.Errorf("for clause %v did not find %v in store %v", test.clause, fact, store)
		}
	}
}

func TestMap(t *testing.T) {
	tests := []testCase{
		{
			initialFacts:  []ast.Atom{atom("bar(0)"), atom("bar(1)")},
			clause:        clause("foo(X) :- bar(Y) |> let X = fn:plus(Y, 1)."),
			expectedFacts: []ast.Atom{atom("foo(1)"), atom("foo(2)")},
		},
		{
			initialFacts:  []ast.Atom{atom("bar(0)"), atom("bar(1)")},
			clause:        clause("foo(Z) :- bar(Y) |> let X = fn:plus(Y, 1), let Z = fn:plus(X, 1)."),
			expectedFacts: []ast.Atom{atom("foo(2)"), atom("foo(3)")},
		},
	}

	for _, test := range tests {
		runEval(test, t)
	}
}

func TestReduce(t *testing.T) {
	tests := []testCase{
		{
			initialFacts:  []ast.Atom{atom("bar(0)"), atom("bar(1)")},
			clause:        clause("foo(X) :- bar(Y) |> do fn:group_by(), let X = fn:sum(Y)."),
			expectedFacts: []ast.Atom{atom("foo(1)")},
		},
		{
			initialFacts:  []ast.Atom{atom("bar(0)"), atom("bar(1)")},
			clause:        clause("foo(X) :- bar(Y) |> do fn:group_by(), let X = fn:count()."),
			expectedFacts: []ast.Atom{atom("foo(2)")},
		},
		{
			initialFacts:  []ast.Atom{atom("bar(0)"), atom("bar(1)")},
			clause:        clause("foo(X) :- bar(Y) |> do fn:group_by(), let X = fn:min(Y)."),
			expectedFacts: []ast.Atom{atom("foo(0)")},
		},
		{
			initialFacts:  []ast.Atom{atom("bar(0)"), atom("bar(1)")},
			clause:        clause("foo(Min, Max) :- bar(Y) |> do fn:group_by(), let Min = fn:min(Y), let Max = fn:max(Y)."),
			expectedFacts: []ast.Atom{atom("foo(0, 1)")},
		},
		// An alternative way to fn:count()
		{
			initialFacts:  []ast.Atom{atom("bar(0)"), atom("bar(1)")},
			clause:        clause("foo(Num) :- bar(Y) |> do fn:group_by(), let L = fn:collect(Y), let Num = fn:list:len(L)."),
			expectedFacts: []ast.Atom{atom("foo(2)")},
		},
	}

	for _, test := range tests {
		runEval(test, t)
	}
}

func TestGroupBy(t *testing.T) {
	tests := []testCase{
		{
			initialFacts:  []ast.Atom{atom("bar(0, 11)"), atom("bar(0, 12)")},
			clause:        clause("foo(Y, X) :- bar(Y, Z) |> do fn:group_by(Y), let X = fn:sum(Z)."),
			expectedFacts: []ast.Atom{atom("foo(0, 23)")},
		},
		{
			initialFacts: []ast.Atom{
				atom(`bar("a", 11, 100)`),
				atom(`bar("a", 11, 150)`),
				atom(`bar("b", 3, 200))`),
			},
			clause: clause(`foo(Y, Count, Sum, Max) :- bar(Y, Z, A)
			|> do fn:group_by(Y), let Count = fn:count(), let Sum = fn:sum(Z), let Max = fn:max(A).`),
			expectedFacts: []ast.Atom{
				atom(`foo("a", 2, 22, 150)`),
				atom(`foo("b", 1, 3, 200)`),
			},
		},
	}

	for _, test := range tests {
		runEval(test, t)
	}
}

// TestGroupByManyKeys exercises a large number of distinct group keys to
// guard against the old FNV-32 hash-collision bug in which two distinct
// keys could silently merge into a single group.
func TestGroupByManyKeys(t *testing.T) {
	const numKeys = 5000
	var facts []ast.Atom
	for i := 0; i < numKeys; i++ {
		a, err := parse.Atom(fmt.Sprintf("bar(%d, 10)", i))
		if err != nil {
			t.Fatalf("parse: %v", err)
		}
		b, err := parse.Atom(fmt.Sprintf("bar(%d, 32)", i))
		if err != nil {
			t.Fatalf("parse: %v", err)
		}
		facts = append(facts, a, b)
	}
	c := clause("foo(Y, S) :- bar(Y, Z) |> do fn:group_by(Y), let S = fn:sum(Z).")

	store := factstore.NewSimpleInMemoryStore()
	for _, f := range facts {
		store.Add(f)
	}
	var input []ast.ConstSubstList
	premise := c.Premises[0].(ast.Atom)
	store.GetFacts(premise, func(a ast.Atom) error {
		var subst ast.ConstSubstList
		for i, arg := range premise.Args {
			subst = subst.Extend(arg.(ast.Variable), a.Args[i].(ast.Constant))
		}
		input = append(input, subst)
		return nil
	})
	if err := EvalTransform(c.Head, *c.Transform, input, store.Add); err != nil {
		t.Fatalf("EvalTransform: %v", err)
	}
	for i := 0; i < numKeys; i++ {
		want, err := parse.Atom(fmt.Sprintf("foo(%d, 42)", i))
		if err != nil {
			t.Fatalf("parse: %v", err)
		}
		if !store.Contains(want) {
			t.Fatalf("missing aggregate %v — suggests distinct group keys collided", want)
		}
	}
}
