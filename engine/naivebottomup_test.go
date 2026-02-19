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
)

var naiveProgram []ast.Clause

func init() {
	naiveProgram = []ast.Clause{
		clause("path(X,Y) :- edge(X,Y)."),
		clause("path(X,Z) :- edge(X,Y), path(Y,Z)."),
		clause("not_reachable(X, Y) :- node(X), node(Y), !path(X, Y)."),
		clause("in_cycle_eq(X) :- node(X), path(X, Y), X = Y."),
		clause("in_between(X, Z) :- node(X), node(Y), node(Z), path(X, Y), path(Y, Z), X != Y, Y != Z, X != Z."),
		clause("decompose_pair(Y,Z) :- :match_pair(fn:pair(1,2),Y,Z)."),
		clause("decompose_cons(Y,Z) :- :match_cons(fn:list:cons(1,[]),Y,Z)."),
		clause("decompose_nil() :- :match_nil([])."),
	}
}

func TestSimpleEvalNaive(t *testing.T) {
	store := factstore.NewSimpleInMemoryStore()
	store.Add(atom("node(/a)"))
	store.Add(atom("node(/b)"))
	store.Add(atom("node(/c)"))
	store.Add(atom("node(/d)"))
	store.Add(atom("edge(/a,/b)"))
	store.Add(atom("edge(/b,/c)"))
	store.Add(atom("edge(/c,/d)"))

	if err := EvalProgramNaive(naiveProgram, store); err != nil {
		t.Errorf("Program evaluation failed %v program %v", err, program)
		return
	}

	expected := []ast.Atom{
		atom("path(/a,/b)"),
		atom("path(/a,/c)"),
		atom("path(/a,/d)"),
		atom("path(/b,/c)"),
		atom("path(/b,/d)"),
		atom("path(/c,/d)"),
		atom("decompose_pair(1, 2)"),
		ast.NewAtom("decompose_cons", ast.Number(1), ast.ListNil),
		atom("decompose_nil()"),
	}

	for _, fact := range expected {
		if !store.Contains(fact) {
			t.Errorf("expected fact %v in store %v", fact, store)
		}
	}
}

func TestManyPathsNaive(t *testing.T) {
	store := factstore.NewSimpleInMemoryStore()
	store.Add(atom("node(/a)"))
	for i := 1; i <= 10; i++ {
		store.Add(atom(fmt.Sprintf("node(/b%d)", i)))
		store.Add(atom(fmt.Sprintf("node(/c%d)", i)))
		store.Add(atom(fmt.Sprintf("edge(/a,/b%d)", i)))
		store.Add(atom(fmt.Sprintf("edge(/b%d, /c%d)", i, i)))
	}
	store.Add(atom("edge(/c9,/b9)"))
	if err := EvalProgramNaive(naiveProgram, store); err != nil {
		t.Errorf("Program evaluation failed %v program %v", err, program)
		return
	}

	expected := []ast.Atom{
		atom("path(/a,/c2)"),
		atom("path(/c9,/c9)"),
		atom("in_cycle_eq(/c9)"),
		atom("in_between(/a,/c9)"),
	}

	for _, fact := range expected {
		if !store.Contains(fact) {
			t.Errorf("expected fact %v in store %v", fact, store)
		}
	}
}

func TestBuiltinNaive(t *testing.T) {
	store := factstore.NewSimpleInMemoryStore()
	store.Add(ast.NewAtom("foo", ast.Number(1)))
	store.Add(ast.NewAtom("foo", ast.Number(2)))
	store.Add(ast.NewAtom("foo", ast.Number(11)))
	program := []ast.Clause{
		clause("lt_two(X) :- foo(X), X < 2."),
		clause("le_two(X) :- foo(X), X <= 2."),
		clause("two_le(X) :- foo(X), 2 <= X."),
		clause("within_ten(X) :- foo(X), :within_distance(10, X, 2)."),
	}
	if err := EvalProgramNaive(program, store); err != nil {
		t.Errorf("Program evaluation failed %v program %v", err, program)
		return
	}
	expected := []ast.Atom{
		atom("lt_two(1)"),
		atom("le_two(1)"),
		atom("le_two(2)"),
		atom("two_le(2)"),
		atom("within_ten(11)"),
	}

	for _, fact := range expected {
		if !store.Contains(fact) {
			t.Errorf("expected fact %v in store %v", fact, store)
		}
	}
}
