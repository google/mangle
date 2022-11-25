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

package unionfind

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/mangle/ast"
	"github.com/google/mangle/parse"
)

func term(s string) ast.BaseTerm {
	term, err := parse.BaseTerm(s)
	if err != nil {
		panic(err)
	}
	return term
}

func equals(left, right ast.Constant) bool {
	return left.Equals(right)
}

func TestUnifyPositive(t *testing.T) {
	tests := []struct {
		name  string
		left  []ast.BaseTerm
		right []ast.BaseTerm
		want  map[ast.Variable]ast.BaseTerm
	}{
		{name: "empty"},
		{
			name:  "basic",
			left:  []ast.BaseTerm{term("X")},
			right: []ast.BaseTerm{term("/bar")},
			want:  map[ast.Variable]ast.BaseTerm{ast.Variable{"X"}: term("/bar")},
		},
		{
			name:  "basic wildcard",
			left:  []ast.BaseTerm{term("_")},
			right: []ast.BaseTerm{term("/bar")},
		},
		{
			name:  "two vars",
			left:  []ast.BaseTerm{term("X"), term("Y")},
			right: []ast.BaseTerm{term("Y"), term("/bar")},
			want: map[ast.Variable]ast.BaseTerm{
				ast.Variable{"X"}: term("/bar"),
				ast.Variable{"Y"}: term("/bar"),
			},
		},
		{
			name:  "two vars wildcard",
			left:  []ast.BaseTerm{term("_"), term("_")},
			right: []ast.BaseTerm{term("Y"), term("/bar")},
		},
	}
	for _, test := range tests {
		uf, err := UnifyTerms(test.left, test.right)
		if err != nil {
			t.Errorf("%s: UnifyTerms(%v, %v) failed %v", test.name, test.left, test.right, err)
			continue
		}
		for key, val := range test.want {
			res := uf.Get(key)
			if res == nil {
				t.Errorf("%s: UnifyTerms(%v, %v)=%v missing %s", test.name, test.left, test.right, uf, key)
			} else if !res.Equals(val) {
				t.Errorf("%s: UnifyTerms(%v, %v)=%v want %v -> %v", test.name, test.left, test.right, uf, key, val)
			}
		}
	}
}

func TestUnifyNegative(t *testing.T) {
	tests := []struct {
		name  string
		left  []ast.BaseTerm
		right []ast.BaseTerm
	}{
		{
			name:  "inconsistent 1",
			left:  []ast.BaseTerm{term("X"), term("X")},
			right: []ast.BaseTerm{term("/bar"), term("/baz")},
		},
		{
			name:  "inconsistent 2",
			left:  []ast.BaseTerm{term("X"), term("X"), term("Y")},
			right: []ast.BaseTerm{term("Y"), term("/baz"), term("/bar")},
		},
	}
	for _, test := range tests {
		if uf, err := UnifyTerms(test.left, test.right); err == nil { // if NO error
			t.Errorf("%s: UnifyTerms(%v,%v)=%v expected error", test.name, test.left, test.right, uf)
		}
	}
}

// unionFindFromMap turns the given map into a UnionFind instance.
func unionFindFromMap(base map[ast.Variable]ast.BaseTerm) (UnionFind, error) {
	var vars []ast.BaseTerm
	var terms []ast.BaseTerm
	for v, c := range base {
		vars = append(vars, v)
		terms = append(terms, c)
	}
	return UnifyTerms(vars, terms)
}

func TestUnifyExtendPositive(t *testing.T) {
	tests := []struct {
		name string
		// The substitution that will be extended.
		base  map[ast.Variable]ast.BaseTerm
		left  []ast.BaseTerm
		right []ast.BaseTerm
		// Bindings that we want to see in the result.
		want map[ast.Variable]ast.BaseTerm
	}{
		{
			name:  "unify X = Y extending empty subst",
			base:  map[ast.Variable]ast.BaseTerm{},
			left:  []ast.BaseTerm{term("X")},
			right: []ast.BaseTerm{term("Y")},
			want: map[ast.Variable]ast.BaseTerm{
				ast.Variable{"X"}: term("Y"),
				ast.Variable{"Y"}: term("Y"),
			},
		},
		{
			name: "unify X = Y where X is already bound",
			base: map[ast.Variable]ast.BaseTerm{
				ast.Variable{"X"}: term("/c"),
			},
			left:  []ast.BaseTerm{term("X")},
			right: []ast.BaseTerm{term("Y")},
			want: map[ast.Variable]ast.BaseTerm{
				ast.Variable{"X"}: term("/c"),
				ast.Variable{"Y"}: term("/c"),
			},
		},
		{
			name: "unify X = Y where both X and Y were bound",
			base: map[ast.Variable]ast.BaseTerm{
				ast.Variable{"X"}: term("/c"),
				ast.Variable{"Y"}: term("X"),
			},
			left:  []ast.BaseTerm{term("X")},
			right: []ast.BaseTerm{term("Y")},
			want: map[ast.Variable]ast.BaseTerm{
				ast.Variable{"X"}: term("/c"),
				ast.Variable{"Y"}: term("/c"),
			},
		},
		{
			name: "unify X = Y where both X and Y were unified before",
			base: map[ast.Variable]ast.BaseTerm{
				ast.Variable{"X"}: term("Z"),
				ast.Variable{"Y"}: term("Z"),
			},
			left:  []ast.BaseTerm{term("X")},
			right: []ast.BaseTerm{term("/c")},
			want: map[ast.Variable]ast.BaseTerm{
				ast.Variable{"X"}: term("/c"),
				ast.Variable{"Y"}: term("/c"),
				ast.Variable{"Z"}: term("/c"),
			},
		},
		{
			name: "unify X = Y where both X and Y were bound before",
			base: map[ast.Variable]ast.BaseTerm{
				ast.Variable{"X"}: term("/c"),
				ast.Variable{"Y"}: term("Z"),
			},
			left:  []ast.BaseTerm{term("X")},
			right: []ast.BaseTerm{term("Y")},
			want: map[ast.Variable]ast.BaseTerm{
				ast.Variable{"Z"}: term("/c"),
			},
		},
	}
	for _, test := range tests {
		// Prepare the substitution to be extended.
		base, err := unionFindFromMap(test.base)
		if err != nil {
			t.Errorf("broken test case: %s", test.name)
			return
		}
		// Extend the substitution.
		uf, err := UnifyTermsExtend(test.left, test.right, base)
		if err != nil {
			t.Errorf("%s: UnifyTermsExtend(%v,%v,%v) failed %v", test.name, test.left, test.right, base, err)
			return
		}
		for v, c := range test.want {
			if !cmp.Equal(uf.Get(v), c, cmp.Comparer(equals)) {
				t.Errorf("%s: UnifyTermsExtend(%v,%v,%v)=%v want %v -> %v", test.name, test.left, test.right, base, uf, v, c)
			}
		}
	}
}

func TestUnifyExtendNegative(t *testing.T) {
	tests := []struct {
		name string
		// The substitution that will be extended.
		base  map[ast.Variable]ast.BaseTerm
		left  []ast.BaseTerm
		right []ast.BaseTerm
		want  map[ast.Variable]ast.BaseTerm
	}{
		{
			name: "unify fails, one exists",
			base: map[ast.Variable]ast.BaseTerm{
				ast.Variable{"X"}: term("/c1"),
			},
			left:  []ast.BaseTerm{term("X")},
			right: []ast.BaseTerm{term("/c2")},
		},
		{
			name: "unify fails two exist",
			base: map[ast.Variable]ast.BaseTerm{
				ast.Variable{"X"}: term("/c1"),
				ast.Variable{"Y"}: term("/c2"),
			},
			left:  []ast.BaseTerm{term("X")},
			right: []ast.BaseTerm{term("Y")},
		},
	}
	for _, test := range tests {
		// Prepare the substitution to be extended.
		base, err := unionFindFromMap(test.base)
		if err != nil {
			t.Fatalf("error in test case: %s", test.name)
		}
		// Extend the substitution.
		uf, err := UnifyTermsExtend(test.left, test.right, base)
		if err == nil { // if NO error
			t.Errorf("%s: UnifyTermsExtend(%v,%v,%v)=%v want error", test.name, test.left, test.right, base, uf)
		}
	}
}
