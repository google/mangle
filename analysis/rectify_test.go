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
	"testing"

	"github.com/google/go-cmp/cmp"
	"codeberg.org/TauCeti/mangle-go/ast"
	"codeberg.org/TauCeti/mangle-go/unionfind"
)

func TestRectifyAtom(t *testing.T) {
	tests := []struct {
		atom      ast.Atom
		used      []ast.Variable
		wantAtom  ast.Atom
		wantFml   []ast.Term
		wantBound []ast.Variable
		fresh     []ast.Variable
	}{
		{
			atom:     atom("foo(_)"),
			wantAtom: atom("foo(_)"),
		},
		{
			atom:      atom("foo(X)"),
			wantAtom:  atom("foo(X)"),
			wantBound: []ast.Variable{ast.Variable{"X"}},
		},
		{
			atom:     atom("foo(X)"),
			used:     []ast.Variable{ast.Variable{"X"}},
			wantAtom: atom("foo(Y)"),
			wantFml:  []ast.Term{ast.Eq{ast.Variable{"Y"}, ast.Variable{"X"}}},
			fresh:    []ast.Variable{ast.Variable{"Y"}},
		},
		{
			atom:      atom("foo(X, Y)"),
			used:      []ast.Variable{ast.Variable{"X"}},
			wantAtom:  atom("foo(Z, Y)"),
			wantFml:   []ast.Term{ast.Eq{ast.Variable{"Z"}, ast.Variable{"X"}}},
			wantBound: []ast.Variable{ast.Variable{"Y"}},
			fresh:     []ast.Variable{ast.Variable{"Z"}},
		},
		{
			atom:     atom("foo(X, 23001)"),
			used:     []ast.Variable{ast.Variable{"X"}},
			wantAtom: atom("foo(Z, Y)"),
			wantFml: []ast.Term{
				ast.Eq{ast.Variable{"Z"}, ast.Variable{"X"}},
				ast.Eq{ast.Variable{"Y"}, ast.Number(23001)}},
			fresh: []ast.Variable{ast.Variable{"Z"}, ast.Variable{"Y"}},
		},
		{
			atom:      atom("foo(X, X)"),
			wantAtom:  atom("foo(X, Y)"),
			wantFml:   []ast.Term{ast.Eq{ast.Variable{"Y"}, ast.Variable{"X"}}},
			wantBound: []ast.Variable{ast.Variable{"X"}},
			fresh:     []ast.Variable{ast.Variable{"Y"}},
		},
		{
			atom:     atom("foo(fn:plus(X, 1))"),
			wantAtom: atom("foo(Y)"),
			wantFml:  []ast.Term{fml("Y = fn:plus(X, 1)")},
			fresh:    []ast.Variable{ast.Variable{"Y"}},
		},
		{
			atom:     atom("foo(fn:plus(A, fn:plus(B, 1)))"),
			wantAtom: atom("foo(Y)"),
			wantFml:  []ast.Term{fml("Y = fn:plus(A, fn:plus(B, 1))")},
			fresh:    []ast.Variable{ast.Variable{"Y"}},
		},
	}
	for _, test := range tests {
		gotAtom, gotFml, freshVars, boundVars := RectifyAtom(test.atom, VarList{test.used})
		if len(freshVars) != len(test.fresh) {
			t.Errorf("RectifyAtom(%v, %v)=%v,%v,%v,%v want %d fresh vars",
				test.atom, test.used, gotAtom, gotFml, freshVars, boundVars, len(test.fresh))
			continue
		}
		if !cmp.Equal(boundVars, test.wantBound, cmp.AllowUnexported(ast.Eq{}, ast.Constant{})) {
			t.Errorf("RectifyAtom(%v, %v)=%v,%v,%v,%v (?!) want bound vars %v",
				test.atom, test.used, gotAtom, gotFml, freshVars, boundVars, test.wantBound)
			continue
		}
		// We need to replace the variables in our "want" terms with the ones
		// that RectifyAtoms produced.
		freshTerm := make([]ast.BaseTerm, len(freshVars))
		for i, v := range freshVars {
			freshTerm[i] = v
		}
		rename, err := unionfind.InitVars(test.fresh, freshTerm)
		if err != nil {
			t.Fatalf("Could not make a substitution %v %v", test.fresh, freshTerm)
		}

		wantAtom := test.wantAtom.ApplySubst(rename)
		if !cmp.Equal(gotAtom, wantAtom, cmp.AllowUnexported(ast.Eq{}, ast.ApplyFn{}, ast.Constant{})) {
			t.Errorf("RectifyAtom(%v, %v)=%v (?!),%v,%v,%v want atom %v",
				test.atom, test.used, gotAtom, gotFml, freshVars, boundVars, wantAtom)
		}

		var wantFml []ast.Term
		for _, fml := range test.wantFml {
			wantFml = append(wantFml, fml.ApplySubst(rename))
		}
		if !cmp.Equal(gotFml, wantFml, cmp.AllowUnexported(ast.Eq{}, ast.ApplyFn{}, ast.Constant{})) {
			t.Errorf("RectifyAtom(%v, %v)=%v,%v (?!),%v,%v want fml %v",
				test.atom, test.used, gotAtom, gotFml, freshVars, boundVars, wantFml)
		}
	}
}
