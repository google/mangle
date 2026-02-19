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

package analysis

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"codeberg.org/TauCeti/mangle-go/ast"
)

func TestRewriteClauseNop(t *testing.T) {
	clause := ast.Clause{Head: atom("foo(X)"), Premises: []ast.Term{atom("bar(X)")}}
	got := RewriteClause(nil, clause)
	if len(got.Premises) != 1 || !got.Premises[0].Equals(clause.Premises[0]) {
		t.Errorf("RewriteClause(nil, %v)=%v want %v", clause, got, clause)
	}
}

func TestRewriteClauseNegative(t *testing.T) {
	d, err := ast.NewDecl(
		atom("bar(Z)"),
		[]ast.Atom{atom("reflects(/bar)")},
		[]ast.BoundDecl{ast.NewBoundDecl(name("/bar"))},
		nil)
	if err != nil {
		t.Fatal(err)
	}
	decls := map[ast.PredicateSym]*ast.Decl{ast.PredicateSym{"bar", 1}: &d}
	cl := clause("foo(X) :- bar(X).")
	got := RewriteClause(decls, cl)
	if len(got.Premises) != 1 || !got.Premises[0].Equals(cl.Premises[0]) {
		t.Errorf("RewriteClause(%v, %v)=%v want %v", decls, cl, got, cl)
	}
}

func TestRewriteClauseInputMode(t *testing.T) {
	d, err := ast.NewDecl(
		atom("bar(Z)"),
		[]ast.Atom{
			atom("reflects(/bar)"),
			atom("mode('+')"),
		},
		[]ast.BoundDecl{ast.NewBoundDecl(name("/bar"))},
		nil)
	if err != nil {
		t.Fatal(err)
	}
	decls := map[ast.PredicateSym]*ast.Decl{ast.PredicateSym{"bar", 1}: &d}
	clause := clause("foo(X) :- bar(X).")
	got := RewriteClause(decls, clause)
	want := atom(":match_prefix(X, /bar)")
	if len(got.Premises) != 1 || !got.Premises[0].Equals(want) {
		t.Errorf("RewriteClause(%v, %v)=%v want %v", decls, clause, got, want)
	}
}

func TestRewriteClauseDefinedPreviously(t *testing.T) {
	scanDecl, err := ast.NewSyntheticDecl(atom("scan(X)"))
	if err != nil {
		t.Fatal(err)
	}
	d, err := ast.NewDecl(
		atom("bar(Z)"),
		[]ast.Atom{
			atom("reflects(/bar)"),
		},
		[]ast.BoundDecl{ast.NewBoundDecl(name("/bar"))},
		nil)
	if err != nil {
		t.Fatal(err)
	}
	decls := map[ast.PredicateSym]*ast.Decl{
		ast.PredicateSym{"scan", 1}: &scanDecl,
		ast.PredicateSym{"bar", 1}:  &d,
	}
	clause := clause("foo(X) :- scan(X), bar(X).")
	got := RewriteClause(decls, clause)
	want := atom(":match_prefix(X, /bar)")
	if len(got.Premises) != 2 || !got.Premises[1].Equals(want) {
		t.Errorf("RewriteClause(%v, %v)=%v want %v", decls, clause, got, want)
	}
}

func TestRewriteClauseDelayNegAtoms(t *testing.T) {
	clause := ast.Clause{
		Head:     atom("foo(X)"),
		Premises: []ast.Term{ast.NegAtom{Atom: atom("bar(X)")}, atom("baz(Y, X)")}}
	got := RewriteClause(nil, clause)
	wantPremises := []ast.Term{atom("baz(Y, X)"), ast.NegAtom{Atom: atom("bar(X)")}}
	if diff := cmp.Diff(wantPremises, got.Premises); diff != "" {
		t.Errorf("RewriteClause(nil, %v)=%v want %v", clause, got, clause)
	}
}

func TestRewriteClauseDelayNegAtomsNone(t *testing.T) {
	clause := ast.Clause{
		Head:     atom("foo(X)"),
		Premises: []ast.Term{ast.NegAtom{Atom: atom("bar(X)")}, ast.NegAtom{Atom: atom("baz(X)")}, atom("bak(Y, X)")}}
	got := RewriteClause(nil, clause)
	wantPremises := []ast.Term{atom("bak(Y, X)"), ast.NegAtom{Atom: atom("bar(X)")}, ast.NegAtom{Atom: atom("baz(X)")}}
	if diff := cmp.Diff(wantPremises, got.Premises); diff != "" {
		t.Errorf("RewriteClause(nil, %v)=%v want %v", clause, got, wantPremises)
	}
}

func TestRewriteClauseDelayNegAtomsUnchanged(t *testing.T) {
	clause := ast.Clause{
		Head:     atom("foo(X)"),
		Premises: []ast.Term{atom("faz(X)"), ast.NegAtom{Atom: atom("bar(X)")}, atom("baz(Y, X)")}}
	got := RewriteClause(nil, clause)
	wantPremises := clause.Premises
	if diff := cmp.Diff(wantPremises, got.Premises); diff != "" {
		t.Errorf("RewriteClause(nil, %v)=%v want %v", clause, got, clause)
	}
}
