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

package symbols

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/google/mangle/ast"
)

func name(n string) ast.Constant {
	c, err := ast.Name(n)
	if err != nil {
		panic(err)
	}
	return c
}

func TestDesugarNoBound(t *testing.T) {
	decls := map[ast.PredicateSym]ast.Decl{
		ast.PredicateSym{"foo", 1}: {
			ast.NewAtom("foo", ast.Variable{"X"}),
			nil,
			nil, // no bounds
			nil,
		},
	}
	desugared, err := CheckAndDesugar(decls)
	if err != nil {
		t.Fatal(err)
	}
	got, ok := desugared[ast.PredicateSym{"foo", 1}]
	if !ok {
		t.Fatalf("no decl found: %v", desugared)
	}

	want := []ast.BoundDecl{ast.NewBoundDecl(ast.AnyBound)}
	if diff := cmp.Diff(want, got.Bounds, cmp.AllowUnexported(ast.Constant{})); diff != "" {
		t.Errorf("CheckAndDesugar(%v)[foo] bounds diff -want +got %s %v", decls, diff, got)
	}
}

func TestDesugarPropagate(t *testing.T) {
	fooPrefixType, _ := ast.Name("/foo")
	decls := map[ast.PredicateSym]ast.Decl{
		ast.PredicateSym{"foo", 1}: {
			ast.NewAtom("foo", ast.Variable{"X"}),
			nil,
			[]ast.BoundDecl{
				ast.NewBoundDecl(ast.StringBound),
				ast.NewBoundDecl(fooPrefixType),
			},
			nil,
		},
		ast.PredicateSym{"bar", 1}: {
			ast.NewAtom("bar", ast.Variable{"X"}),
			nil,
			[]ast.BoundDecl{ast.NewBoundDecl(ast.String("foo"))},
			&ast.InclusionConstraint{[]ast.Atom{
				ast.NewAtom("foo", ast.Variable{"X"}),
			}, [][]ast.Atom{
				nil,
			}},
		},
		ast.PredicateSym{"baz", 2}: {
			ast.NewAtom("baz", ast.Variable{"X"}, ast.Variable{"Y"}),
			nil,
			[]ast.BoundDecl{ast.NewBoundDecl(ast.String("foo"), ast.String("bar"))},
			&ast.InclusionConstraint{[]ast.Atom{
				ast.NewAtom("foo", ast.Variable{"X"}),
			}, [][]ast.Atom{
				nil,
			}},
		},
	}
	wants := map[ast.PredicateSym]struct {
		bounds []ast.BoundDecl
		incl   ast.InclusionConstraint
	}{
		ast.PredicateSym{"foo", 1}: {
			bounds: []ast.BoundDecl{
				ast.NewBoundDecl(ast.StringBound),
				ast.NewBoundDecl(fooPrefixType),
			},
			incl: ast.InclusionConstraint{nil,
				[][]ast.Atom{
					nil,
					nil,
				},
			},
		},
		ast.PredicateSym{"bar", 1}: {
			bounds: []ast.BoundDecl{
				ast.NewBoundDecl(ast.ApplyFn{UnionType, []ast.BaseTerm{
					ast.StringBound,
					fooPrefixType}})},
			incl: ast.InclusionConstraint{
				nil,
				[][]ast.Atom{
					[]ast.Atom{ast.NewAtom("foo", ast.Variable{"X"})},
				}},
		},
		ast.PredicateSym{"baz", 2}: {
			bounds: []ast.BoundDecl{
				ast.NewBoundDecl(
					ast.ApplyFn{UnionType, []ast.BaseTerm{
						ast.StringBound,
						fooPrefixType}},
					ast.ApplyFn{UnionType, []ast.BaseTerm{
						ast.StringBound,
						fooPrefixType}}),
			},
			incl: ast.InclusionConstraint{
				nil,
				[][]ast.Atom{
					[]ast.Atom{
						ast.NewAtom("foo", ast.Variable{"X"}),
						ast.NewAtom("bar", ast.Variable{"Y"}),
					},
				}},
		},
	}
	desugared, err := CheckAndDesugar(decls)
	if err != nil {
		t.Fatal(err)
	}
	for sym, want := range wants {
		decl, ok := desugared[sym]
		if !ok {
			t.Errorf("CheckAndDesugar(%v) does not contain %v.", decls, sym)
			continue
		}

		if diff := cmp.Diff(want.bounds, decl.Bounds, cmp.AllowUnexported(ast.Constant{})); diff != "" {
			t.Errorf("CheckAndDesugar(%v)[%v] bounds diff -want +got %s", decls, sym, diff)
		}
		if decl.Constraints != nil {
			sortAtoms := cmpopts.SortSlices(func(a, b ast.Atom) bool { return a.Hash() < b.Hash() })
			if diff := cmp.Diff(want.incl, *decl.Constraints, sortAtoms, cmp.AllowUnexported(ast.Constant{})); diff != "" {
				t.Errorf("CheckAndDesugar(%v)[%v] constraints diff -want +got %s", decls, sym, diff)
			}
		}
	}
}

func TestDesugarCyclic(t *testing.T) {
	decls := map[ast.PredicateSym]ast.Decl{
		ast.PredicateSym{"foo", 1}: {
			ast.NewAtom("foo", ast.Variable{"X"}),
			nil,
			[]ast.BoundDecl{ast.NewBoundDecl(ast.String("bar"))},
			nil,
		},
		ast.PredicateSym{"bar", 1}: {
			ast.NewAtom("bar", ast.Variable{"X"}),
			nil,
			[]ast.BoundDecl{ast.NewBoundDecl(ast.String("foo"))},
			nil,
		},
		ast.PredicateSym{"baz", 2}: {
			ast.NewAtom("baz", ast.Variable{"X"}, ast.Variable{"Y"}),
			nil,
			[]ast.BoundDecl{ast.NewBoundDecl(ast.String("foo"), ast.String("bar"))},
			nil,
		},
	}
	got, err := CheckAndDesugar(decls)
	if err == nil { // if NO error
		t.Errorf("CheckAndDesugar(%v)=%v expected cyclic dependency error", decls, got)
	}
	if !strings.Contains(err.Error(), "foo->bar") {
		t.Errorf("CheckAndDesugar(%v) failed with %q, expected to see foo->bar dependency cycle fragment", decls, err.Error())
	}
}
