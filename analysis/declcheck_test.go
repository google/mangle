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
	"fmt"
	"strings"
	"testing"

	"github.com/google/mangle/ast"
	"github.com/google/mangle/parse"
	"github.com/google/mangle/symbols"
)

var (
	declAtom = ast.Atom{ast.PredicateSym{"foobar", 3}, []ast.BaseTerm{
		ast.Variable{"Bar"}, ast.Variable{"Baz"}}}

	docAtomOne = ast.NewAtom("doc", ast.String("a foo predicate"))
	argAtomBar = ast.NewAtom("arg", ast.Variable{"Bar"}, ast.String("an arg docu for Bar"))
	argAtomBaz = ast.NewAtom("arg", ast.Variable{"Baz"}, ast.String("an arg docu for Baz"))
)

func mustDecl(descrAtoms []ast.Atom, boundDecls []ast.BoundDecl, constraint *ast.InclusionConstraint) ast.Decl {
	decl, err := ast.NewDecl(declAtom, descrAtoms, boundDecls, constraint)
	if err != nil {
		panic(fmt.Errorf("bad test data: %v %v %v %v", declAtom, descrAtoms, boundDecls, constraint))
	}
	return decl
}

func checkBoundDecl(boundDecl ast.BoundDecl) []error {
	return newDeclChecker(mustDecl([]ast.Atom{docAtomOne, argAtomBar, argAtomBaz}, []ast.BoundDecl{boundDecl}, nil)).check()
}

func TestBoundsCheckingPositive(t *testing.T) {
	tests := []struct {
		name      string
		boundDecl ast.BoundDecl
	}{
		{
			name:      "simple arguments",
			boundDecl: ast.NewBoundDecl(ast.NameBound, ast.NumberBound),
		},
		{
			name: "proper type expressions",
			boundDecl: ast.NewBoundDecl(
				ast.ApplyFn{symbols.ListType, []ast.BaseTerm{ast.NumberBound}},
				ast.ApplyFn{symbols.PairType, []ast.BaseTerm{ast.NumberBound, ast.StringBound}}),
		},
		{
			name: "union type expressions",
			boundDecl: ast.NewBoundDecl(
				ast.ApplyFn{symbols.UnionType, []ast.BaseTerm{ast.NameBound, ast.NumberBound}},
				ast.ApplyFn{symbols.UnionType, []ast.BaseTerm{ast.NumberBound, ast.StringBound}}),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if errs := checkBoundDecl(test.boundDecl); errs != nil {
				t.Error(errs[0])
			}
		})
	}
}

func TestCheckDeclExternal(t *testing.T) {
	testCases := []struct {
		desc    string
		source  string
		wantErr bool
	}{
		{
			desc: "valid external decl",
			source: `
                Decl testExt(X, Y)
                  descr [
                      external(),
                      mode('+', '-'),
                    ]
                    bound [ /number, /string ]
                .`,
			wantErr: false,
		},
		{
			desc: "external decl with no modes",
			source: `
                Decl testExt(X, Y)
                  descr [
                      external()
                    ]
                    bound [ /number, /string ]
                .`,
			wantErr: true,
		},
		{
			desc: "external decl with two modes",
			source: `
                Decl testExt(X, Y)
                  descr [
                      external(),
                      mode('+', '-'),
                      mode('-', '+')
                    ]
                    bound [ /number, /string ]
                .`,
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			unit, err := parse.Unit(strings.NewReader(tc.source))
			if err != nil {
				t.Fatalf("parse.Unit(%q) failed: %v", tc.source, err)
			}
			if len(unit.Decls) != 2 { // package decl + test decl
				t.Fatalf("expected 1 decl, got %d", len(unit.Decls))
			}
			decl := unit.Decls[1]
			errs := CheckDecl(decl)
			if (len(errs) > 0) != tc.wantErr {
				t.Errorf("CheckDecl() returned errors %v, wantErr=%v", errs, tc.wantErr)
			}
		})
	}
}

func TestBoundsCheckingNegative(t *testing.T) {
	tests := []struct {
		name      string
		boundDecl ast.BoundDecl
	}{
		{
			name:      "not enough bounds",
			boundDecl: ast.NewBoundDecl(ast.StringBound),
		},
		{
			name: "bad type expressions - too many args",
			boundDecl: ast.NewBoundDecl(
				ast.ApplyFn{symbols.ListType, []ast.BaseTerm{ast.NumberBound, ast.AnyBound}},
				ast.AnyBound),
		},
		{
			name: "bad type expression - not enough args",
			boundDecl: ast.NewBoundDecl(
				ast.AnyBound,
				ast.ApplyFn{symbols.PairType, []ast.BaseTerm{ast.NumberBound}}),
		},
		{
			name: "bad type expression - empty union",
			boundDecl: ast.NewBoundDecl(
				ast.AnyBound,
				ast.ApplyFn{symbols.UnionType, nil}),
		},
		{
			name: "bad type expression - tuple with less than 3 args",
			boundDecl: ast.NewBoundDecl(
				ast.AnyBound,
				ast.ApplyFn{symbols.TupleType, []ast.BaseTerm{ast.NumberBound}}),
		},
		{
			name: "predicate bound not allowed in structured type expression",
			boundDecl: ast.NewBoundDecl(
				ast.ApplyFn{symbols.ListType, []ast.BaseTerm{ast.String("foo")}},
				ast.ApplyFn{symbols.PairType, []ast.BaseTerm{ast.NumberBound, ast.AnyBound}}),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if errs := checkBoundDecl(test.boundDecl); errs == nil {
				t.Errorf("%s: expected error %+v", test.name, test.boundDecl)
			}
		})
	}
}

func TestDescrCheckingPositive(t *testing.T) {
	tests := []struct {
		name       string
		descrAtoms []ast.Atom
	}{
		{
			name:       "one-arg doc atom",
			descrAtoms: []ast.Atom{docAtomOne, argAtomBar, argAtomBaz},
		},
		{
			name:       "two-args doc atoms",
			descrAtoms: []ast.Atom{ast.NewAtom("doc", ast.String("a foo predicate"), ast.String("does a lot of work")), argAtomBar, argAtomBaz},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := newDeclChecker(mustDecl(test.descrAtoms, nil, nil))
			if errs := c.check(); errs != nil {
				t.Error(errs[0])
			}
		})
	}
}

func TestDescrCheckingNegative(t *testing.T) {
	tests := []struct {
		name       string
		descrAtoms []ast.Atom
	}{
		{
			name:       "empty doc atom",
			descrAtoms: []ast.Atom{ast.NewAtom("doc"), argAtomBar, argAtomBaz},
		},
		{
			name:       "bad doc atom",
			descrAtoms: []ast.Atom{ast.NewAtom("doc", ast.Number(42)), argAtomBar, argAtomBaz},
		},
		{
			name:       "missing argdoc",
			descrAtoms: []ast.Atom{docAtomOne, argAtomBar},
		},
		{
			name:       "too many argdoc",
			descrAtoms: []ast.Atom{docAtomOne, argAtomBar, argAtomBaz, argAtomBaz},
		},
		{
			name: "wrong argdoc",
			descrAtoms: []ast.Atom{
				docAtomOne,
				argAtomBar,
				ast.NewAtom("arg", ast.Variable{"Wrong"}, ast.String("an arg docu for Wrong")),
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := newDeclChecker(mustDecl(test.descrAtoms, nil, nil))
			if errs := c.check(); errs == nil {
				t.Errorf("%s: expected error for %+v", test.name, c.decl)
			}
		})
	}
}
