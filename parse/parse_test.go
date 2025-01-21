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

package parse

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/mangle/ast"
	"github.com/google/mangle/symbols"
)

var emptyDoc = []ast.Atom{ast.NewAtom("doc", ast.String(""))}

func name(str string) ast.Constant {
	res, _ := ast.Name(str)
	return res
}

func equals(left, right ast.Constant) bool {
	return left.Equals(right)
}

func makeDecl(t *testing.T, atom ast.Atom, descrAtoms []ast.Atom, bounds []ast.BoundDecl, constraints *ast.InclusionConstraint) ast.Decl {
	t.Helper()
	decl, err := ast.NewDecl(atom, descrAtoms, bounds, constraints)
	if err != nil {
		t.Fatal(err)
	}
	return decl
}

func TestParseDecl(t *testing.T) {
	inclConstraint := ast.NewInclusionConstraint([]ast.Atom{
		ast.NewAtom("bar", ast.Variable{"X"}),
		ast.NewAtom("baz", ast.Variable{"X"}),
		ast.NewAtom("bak", ast.Variable{"X"}, ast.Variable{"Y"}),
	})
	tests := []struct {
		name string
		str  string
		want []ast.Decl
	}{
		{
			name: "one decl",
			str:  "Decl foo(X,Y).",
			want: []ast.Decl{makeDecl(t, ast.NewAtom("foo", ast.Variable{"X"}, ast.Variable{"Y"}), emptyDoc, nil, nil)},
		},
		{
			name: "one decl one bound",
			str:  "Decl foo(X,Y) bound [/string, fn:List(/string)].",
			want: []ast.Decl{makeDecl(t, ast.NewAtom("foo", ast.Variable{"X"}, ast.Variable{"Y"}),
				nil,
				[]ast.BoundDecl{
					ast.NewBoundDecl(ast.StringBound, ast.ApplyFn{symbols.ListType, []ast.BaseTerm{ast.StringBound}}),
				},
				nil)},
		},
		{
			name: "one decl one bound fancy syntax",
			str:  "Decl foo(X,Y) bound [/string, .List</string>].",
			want: []ast.Decl{makeDecl(t, ast.NewAtom("foo", ast.Variable{"X"}, ast.Variable{"Y"}),
				nil,
				[]ast.BoundDecl{
					ast.NewBoundDecl(ast.StringBound, ast.ApplyFn{Function: ast.FunctionSym{"fn:List", -1}, Args: []ast.BaseTerm{ast.StringBound}}),
				},
				nil)},
		},
		{
			name: "one decl two bounds w/ whitespace",
			str:  "Decl foo(X,Y) bound[   /string, /string ]bound[/number, /number].",
			want: []ast.Decl{makeDecl(t, ast.NewAtom("foo", ast.Variable{"X"}, ast.Variable{"Y"}),
				emptyDoc,
				[]ast.BoundDecl{
					ast.NewBoundDecl(ast.StringBound, ast.StringBound),
					ast.NewBoundDecl(ast.NumberBound, ast.NumberBound),
				},
				nil)},
		},
		{
			name: "one decl one inclusion",
			str:  "Decl foo(X,Y) inclusion[bar(X), baz(X), bak(X, Y)].",
			want: []ast.Decl{makeDecl(t, ast.NewAtom("foo", ast.Variable{"X"}, ast.Variable{"Y"}), emptyDoc, nil, &inclConstraint)},
		},
		{
			name: "one decl one bound one inclusion",
			str:  "Decl foo(X,Y) bound[/string, .Pair</string, /number>] inclusion[bar(X), baz(X), bak(X, Y)].",
			want: []ast.Decl{
				makeDecl(t, ast.NewAtom("foo", ast.Variable{"X"}, ast.Variable{"Y"}),
					emptyDoc,
					[]ast.BoundDecl{
						ast.NewBoundDecl(ast.StringBound, ast.ApplyFn{
							ast.FunctionSym{"fn:Pair", -1}, []ast.BaseTerm{ast.StringBound, ast.NumberBound}}),
					},
					&inclConstraint)},
		},
		{
			name: "one decl descr",
			str: `Decl foo(X,Y)
			       descr[doc("a foo predicate")].`,
			want: []ast.Decl{makeDecl(t, ast.NewAtom("foo", ast.Variable{"X"}, ast.Variable{"Y"}),
				[]ast.Atom{
					ast.NewAtom("doc", ast.String("a foo predicate")),
				},
				nil, nil)},
		},
		{
			name: "one decl descr",
			str: `Decl foo(X,Y)
			       descr[
						   doc("a foo predicate",
							     "goes well with bar"),
							 arg(X, "ID of the foo"),
							 arg(Y, "parameter of the foo")
						].`,
			want: []ast.Decl{makeDecl(t,
				ast.NewAtom("foo", ast.Variable{"X"}, ast.Variable{"Y"}),
				[]ast.Atom{
					ast.NewAtom("doc", ast.String("a foo predicate"), ast.String("goes well with bar")),
					ast.NewAtom("arg", ast.Variable{"X"}, ast.String("ID of the foo")),
					ast.NewAtom("arg", ast.Variable{"Y"}, ast.String("parameter of the foo")),
				},
				nil, nil)},
		},
		{
			name: "one decl descr bound",
			str: `Decl foo(X,Y)
			        descr[doc("a foo predicate"), fundep([X], [Y])]
						  bound[/string, /string].`,
			want: []ast.Decl{makeDecl(t,
				ast.NewAtom("foo", ast.Variable{"X"}, ast.Variable{"Y"}),
				[]ast.Atom{
					ast.NewAtom("doc", ast.String("a foo predicate")),
					ast.NewAtom("fundep",
						ast.ApplyFn{Function: symbols.List, Args: []ast.BaseTerm{ast.Variable{"X"}}},
						ast.ApplyFn{Function: symbols.List, Args: []ast.BaseTerm{ast.Variable{"Y"}}}),
				},
				[]ast.BoundDecl{
					ast.NewBoundDecl(ast.StringBound, ast.StringBound),
				},
				nil)},
		},
		{
			name: "pair tuple struct map union singleton",
			str:  "Decl foo(X,Y) bound [.Pair</string, .Struct<opt /a : /string, /b : .List<.Option</string>>>>, .Map</string, .Union</a, .Singleton</b>>> ].",
			want: []ast.Decl{makeDecl(t, ast.NewAtom("foo", ast.Variable{"X"}, ast.Variable{"Y"}),
				nil,
				[]ast.BoundDecl{
					ast.NewBoundDecl(
						ast.ApplyFn{
							Function: ast.FunctionSym{"fn:Pair", -1},
							Args: []ast.BaseTerm{
								ast.StringBound,
								ast.ApplyFn{
									Function: ast.FunctionSym{"fn:Struct", -1},
									Args: []ast.BaseTerm{
										ast.ApplyFn{
											Function: ast.FunctionSym{"fn:opt", -1},
											Args: []ast.BaseTerm{
												name("/a"),
												ast.StringBound,
											},
										},
										name("/b"),
										ast.ApplyFn{
											Function: ast.FunctionSym{"fn:List", -1},
											Args: []ast.BaseTerm{
												ast.ApplyFn{
													Function: ast.FunctionSym{"fn:Option", -1},
													Args:     []ast.BaseTerm{ast.StringBound},
												},
											},
										},
									},
								},
							},
						},
						ast.ApplyFn{
							Function: ast.FunctionSym{"fn:Map", -1},
							Args: []ast.BaseTerm{
								ast.StringBound,
								ast.ApplyFn{
									Function: ast.FunctionSym{"fn:Union", -1},
									Args: []ast.BaseTerm{
										name("/a"),
										ast.ApplyFn{
											Function: ast.FunctionSym{"fn:Singleton", -1},
											Args:     []ast.BaseTerm{name("/b")},
										},
									},
								},
							},
						},
					),
				},
				nil)},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := Unit(strings.NewReader(test.str))
			if err != nil {
				t.Fatalf("Unit(%v) failed with %v", test.str, err)
			}
			if diff := cmp.Diff(test.want, got.Decls[1:] /* ignore the Package Decl */, cmp.Comparer(equals)); diff != "" {
				t.Errorf("Unit(%v) = %v want %v diff %v", test.str, got, test.want, diff)
			}
		})
	}

}

func TestParsePackage(t *testing.T) {
	tests := []struct {
		name string
		str  string
		want []ast.Decl
	}{
		{
			name: "Regular package",
			str:  "Package foo.bar!",
			want: []ast.Decl{makeDecl(t, ast.NewAtom("Package"), []ast.Atom{ast.NewAtom("name", ast.String("foo.bar"))}, nil, nil)},
		},
		{
			name: "White space allowed after identifier",
			str:  "Package foo.bar !",
			want: []ast.Decl{makeDecl(t, ast.NewAtom("Package"), []ast.Atom{ast.NewAtom("name", ast.String("foo.bar"))}, nil, nil)},
		},
		{
			name: "Package with atoms",
			str:  `Package foo [atom1("hello"), atom2()]!`,
			want: []ast.Decl{makeDecl(t, ast.NewAtom("Package"), []ast.Atom{ast.NewAtom("name", ast.String("foo")), ast.NewAtom("atom1", ast.String("hello")), ast.NewAtom("atom2")}, nil, nil)},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := Unit(strings.NewReader(test.str))
			if err != nil {
				t.Fatalf("Unit(%v) failed with %v", test.str, err)
			}
			if diff := cmp.Diff(test.want, got.Decls, cmp.Comparer(equals)); diff != "" {
				t.Errorf("Unit(%v) = %v want %v diff %v", test.str, got, test.want, diff)
			}
		})
	}

}

func TestParsePackageError(t *testing.T) {
	tests := []struct {
		name string
		str  string
	}{
		{
			name: "Identifier with two dots next to each other",
			str:  "Package foo..bar!",
		},
		{
			name: "Package without ! ending",
			str:  "Package foo#",
		},
		{
			name: "Package ends with dot and space",
			str:  "Package foo. ",
		},
		{
			name: "Identifier with capital letter",
			str:  "Package fOo!",
		},
		{
			name: "Identifier ends with .",
			str:  "Package foo.!",
		},
		{
			name: "Identifier starts with .",
			str:  "Package .foo!",
		},
		{
			name: "Identifier is empty",
			str:  "Package!",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if _, err := Unit(strings.NewReader(test.str)); err == nil {
				t.Fatalf("Unit(%v) did not fail", test.str)
			}
		})
	}

}

func TestParseUnitDecls(t *testing.T) {
	tests := []struct {
		name string
		str  string
		want []ast.Decl
	}{
		{
			name: "Regular use",
			str:  "Use foo.bar!",
			want: []ast.Decl{
				makeDecl(t, ast.NewAtom("Package"), []ast.Atom{ast.NewAtom("name", ast.String(""))}, nil, nil),
				makeDecl(t, ast.NewAtom("Use"), []ast.Atom{ast.NewAtom("name", ast.String("foo.bar"))}, nil, nil),
			},
		},
		{
			name: "Regular use with package",
			str:  "Package foo! Use bar!",
			want: []ast.Decl{
				makeDecl(t, ast.NewAtom("Package"), []ast.Atom{ast.NewAtom("name", ast.String("foo"))}, nil, nil),
				makeDecl(t, ast.NewAtom("Use"), []ast.Atom{ast.NewAtom("name", ast.String("bar"))}, nil, nil),
			},
		},
		{
			name: "Use with atoms (trailing comma)",
			str:  `Use foo.bar [atom1("hello"), atom2(),]!`,
			want: []ast.Decl{
				makeDecl(t, ast.NewAtom("Package"), []ast.Atom{ast.NewAtom("name", ast.String(""))}, nil, nil),
				makeDecl(t, ast.NewAtom("Use"), []ast.Atom{ast.NewAtom("name", ast.String("foo.bar")), ast.NewAtom("atom1", ast.String("hello")), ast.NewAtom("atom2")}, nil, nil),
			},
		},
		{
			name: "One decl",
			str:  `Decl pred().`,
			want: []ast.Decl{
				makeDecl(t, ast.NewAtom("Package"), []ast.Atom{ast.NewAtom("name", ast.String(""))}, nil, nil),
				makeDecl(t, ast.NewAtom("pred"), nil, nil, nil),
			},
		},
		{
			name: "Multiple uses",
			str:  `Use foo! Use bar!`,
			want: []ast.Decl{
				makeDecl(t, ast.NewAtom("Package"), []ast.Atom{ast.NewAtom("name", ast.String(""))}, nil, nil),
				makeDecl(t, ast.NewAtom("Use"), []ast.Atom{ast.NewAtom("name", ast.String("foo"))}, nil, nil),
				makeDecl(t, ast.NewAtom("Use"), []ast.Atom{ast.NewAtom("name", ast.String("bar"))}, nil, nil),
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := Unit(strings.NewReader(test.str))
			if err != nil {
				t.Fatalf("Unit(%v) failed with %v", test.str, err)
			}
			if diff := cmp.Diff(test.want, got.Decls, cmp.Comparer(equals)); diff != "" {
				t.Errorf("Unit(%v) = %v want %v diff %v", test.str, got, test.want, diff)
			}
		})
	}
}

func TestParseUnitPositive(t *testing.T) {
	tests := []struct {
		name string
		str  string
		want []ast.Clause
	}{
		{
			name: "empty program 1",
			str:  "",
			want: nil,
		},
		{
			name: "empty program 2",
			str:  "  \n\n  ",
			want: nil,
		},
		{
			name: "one clause, no body.",
			str:  "foo(X).",
			want: []ast.Clause{ast.NewClause(ast.NewAtom("foo", ast.Variable{"X"}), nil)},
		},
		{
			name: "one clause, no body, wildcard.",
			str:  "foo(_).",
			want: []ast.Clause{ast.NewClause(ast.NewAtom("foo", ast.Variable{"_"}), nil)},
		},
		{
			name: "one clause, with body.",
			str:  "foo(X) \u21d0 bar(X).",
			want: []ast.Clause{ast.NewClause(ast.NewAtom("foo", ast.Variable{"X"}), []ast.Term{ast.NewAtom("bar", ast.Variable{"X"})})},
		},
		{
			name: "one clause, with body, both contain '.', trailing comma",
			str:  "foo.bar(X) :- bar.foo(X),.",
			want: []ast.Clause{ast.NewClause(ast.NewAtom("foo.bar", ast.Variable{"X"}), []ast.Term{ast.NewAtom("bar.foo", ast.Variable{"X"})})},
		},
		{
			name: "one clause, with body and one do-transform.",
			str:  "foo(X) :- bar(X) |> do fn:party(), let Z = fn:foo(X).",
			want: []ast.Clause{{
				ast.NewAtom("foo", ast.Variable{"X"}),
				[]ast.Term{ast.NewAtom("bar", ast.Variable{"X"})},
				&ast.Transform{
					[]ast.TransformStmt{
						{nil, ast.ApplyFn{ast.FunctionSym{"fn:party", 0}, nil}},
						{&ast.Variable{"Z"}, ast.ApplyFn{ast.FunctionSym{"fn:foo", 1}, []ast.BaseTerm{ast.Variable{"X"}}}},
					},
					nil,
				},
			},
			},
		},
		{
			name: "one clause, with body and one do-transform and let-transform.",
			str:  "foo(X, ZZ) :- bar(X) |> do fn:party(), let Z = fn:foo(X) |> let ZZ = fn:mul(Z, 2).",
			want: []ast.Clause{{
				ast.NewAtom("foo", ast.Variable{"X"}, ast.Variable{"ZZ"}),
				[]ast.Term{ast.NewAtom("bar", ast.Variable{"X"})},
				&ast.Transform{
					[]ast.TransformStmt{
						{nil, ast.ApplyFn{ast.FunctionSym{"fn:party", 0}, nil}},
						{&ast.Variable{"Z"}, ast.ApplyFn{ast.FunctionSym{"fn:foo", 1}, []ast.BaseTerm{ast.Variable{"X"}}}},
					},
					&ast.Transform{
						[]ast.TransformStmt{
							{
								&ast.Variable{"ZZ"}, ast.ApplyFn{
									ast.FunctionSym{"fn:mul", 2},
									[]ast.BaseTerm{ast.Variable{"Z"}, ast.Number(2)}},
							},
						},
						nil,
					},
				},
			},
			},
		},
		{
			name: "one clause, with body and one let-transform.",
			str:  "foo(X) :- bar(X) |> let Y = fn:plus(X, 1).",
			want: []ast.Clause{{
				ast.NewAtom("foo", ast.Variable{"X"}),
				[]ast.Term{ast.NewAtom("bar", ast.Variable{"X"})},
				&ast.Transform{
					[]ast.TransformStmt{
						{&ast.Variable{"Y"}, ast.ApplyFn{ast.FunctionSym{"fn:plus", 2}, []ast.BaseTerm{ast.Variable{"X"}, ast.Number(1)}}},
					},
					nil,
				},
			}},
		},
		{
			name: "two clauses",
			str:  "foo(X). foo(X). foo(X) :- bar(X).\nfoo(X) :- bar(X).\n",
			want: []ast.Clause{
				ast.NewClause(ast.NewAtom("foo", ast.Variable{"X"}), nil),
				ast.NewClause(ast.NewAtom("foo", ast.Variable{"X"}), nil),
				ast.NewClause(ast.NewAtom("foo", ast.Variable{"X"}), []ast.Term{ast.NewAtom("bar", ast.Variable{"X"})}),
				ast.NewClause(ast.NewAtom("foo", ast.Variable{"X"}), []ast.Term{ast.NewAtom("bar", ast.Variable{"X"})}),
			},
		},
		{
			name: "two clauses with comments",
			str:  "foo(X). foo(X). foo(X) :- # Some comment about this\nbar(X).\n# Some comment about stuff\nfoo(X) :- bar(X).\n",
			want: []ast.Clause{
				ast.NewClause(ast.NewAtom("foo", ast.Variable{"X"}), nil),
				ast.NewClause(ast.NewAtom("foo", ast.Variable{"X"}), nil),
				ast.NewClause(ast.NewAtom("foo", ast.Variable{"X"}), []ast.Term{ast.NewAtom("bar", ast.Variable{"X"})}),
				ast.NewClause(ast.NewAtom("foo", ast.Variable{"X"}), []ast.Term{ast.NewAtom("bar", ast.Variable{"X"})}),
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := Unit(strings.NewReader(test.str))
			if err != nil {
				t.Fatalf("Program(%v) failed with %v", test.str, err)
			}
			if diff := cmp.Diff(test.want, got.Clauses, cmp.Comparer(equals)); diff != "" {
				t.Errorf("Program(%v) = %v want %v diff %v", test.str, got, test.want, diff)
			}
		})
	}
}

func TestParseUnitNegative(t *testing.T) {
	tests := []struct {
		name string
		str  string
	}{
		{
			name: "missing body",
			str:  "a(B) :-",
		},
		{
			name: "invalid function",
			str:  "fn:(n()).",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := Unit(strings.NewReader(test.str))
			if err == nil { // if no error
				t.Errorf("Unit(%v)=%v want error", test.str, got)
			}
		})
	}
}

func TestParseClausePositive(t *testing.T) {
	tests := []struct {
		name string
		str  string
		want ast.Clause
	}{
		{
			name: "bodiless clause 1",
			str:  "foo(/bar.baz.Bak).",
			want: ast.NewClause(ast.NewAtom("foo", name("/bar.baz.Bak")), nil),
		},
		{
			name: "bodiless clause 2; invalid but should parse ok.",
			str:  "foo(X).",
			want: ast.NewClause(ast.NewAtom("foo", ast.Variable{"X"}), nil),
		},
		{
			name: "weird recursive clause",
			str:  "foo(/bar-1logjam) :- foo(/bar-1logjam).",
			want: ast.NewClause(
				ast.NewAtom("foo", name("/bar-1logjam")),
				[]ast.Term{
					ast.NewAtom("foo", name("/bar-1logjam")),
				},
			),
		},
		{
			name: "weird clause with body",
			str:  "foo(Xyz) :- Xyz = /foo, /foo != Xyz.",
			want: ast.NewClause(ast.NewAtom("foo", ast.Variable{"Xyz"}), []ast.Term{
				ast.Eq{ast.Variable{"Xyz"}, name("/foo")},
				ast.Ineq{name("/foo"), ast.Variable{"Xyz"}},
			}),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := Clause(test.str)
			if err != nil {
				t.Fatalf("Clause(%v) failed with %v", test.str, err)
			}
			if !got.Head.Equals(test.want.Head) {
				t.Fatalf("Clause(%v) = %v want %v", test.str, got, test.want)
			}
			if diff := cmp.Diff(test.want.Premises, got.Premises, cmp.Comparer(equals)); diff != "" {
				t.Errorf("Clause(%v) = %v want %v diff %v", test.str, got, test.want.Premises, diff)
			}
		})
	}
}

func TestParseClauseNegative(t *testing.T) {
	tests := []struct {
		name string
		str  string
	}{
		{
			name: "missing end",
			str:  "foo(/bar)",
		},
		{
			name: "bad predicate name",
			str:  "_(/bar).",
		},
		{
			name: "variable as head?!",
			str:  "X :- X = /foo, /foo != X.",
		},
		{
			name: "constant as head?!",
			str:  "/foo :- foo(/bar).",
		},
		{
			name: "missing body",
			str:  "a(B) :-",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := Clause(test.str)
			if err == nil { // if no error
				t.Errorf("Clause(%v)=%v want error", test.str, got)
			}
		})
	}
}

func TestParseLiteralOrFormulaPositive(t *testing.T) {
	tests := []struct {
		name string
		str  string
		want ast.Term
	}{
		{
			name: "equality 1",
			str:  "X=Y",
			want: ast.Eq{ast.Variable{"X"}, ast.Variable{"Y"}},
		},
		{
			name: "equality 2",
			str:  "X=/foo",
			want: ast.Eq{ast.Variable{"X"}, name("/foo")},
		},
		{
			name: "equality 3",
			str:  "/foo =Y",
			want: ast.Eq{name("/foo"), ast.Variable{"Y"}},
		},
		{
			name: "equality 4",
			str:  "[/foo] =Y",
			want: ast.Eq{ast.ApplyFn{symbols.List, []ast.BaseTerm{name("/foo")}}, ast.Variable{"Y"}},
		},
		{
			name: "equality 4 fn",
			str:  "fn:list(/foo) =Y",
			want: ast.Eq{ast.ApplyFn{ast.FunctionSym{"fn:list", 1}, []ast.BaseTerm{name("/foo")}}, ast.Variable{"Y"}},
		},
		{
			name: "inequality 1",
			str:  "X!=Y",
			want: ast.Ineq{ast.Variable{"X"}, ast.Variable{"Y"}},
		},
		{
			name: "inequality 2",
			str:  "X!=/foo",
			want: ast.Ineq{ast.Variable{"X"}, name("/foo")},
		},
		{
			name: "inequality 3",
			str:  "/foo!= Y",
			want: ast.Ineq{name("/foo"), ast.Variable{"Y"}},
		},
		{
			name: "builtin :lt",
			str:  "0 < 1",
			want: ast.NewAtom(":lt", ast.Number(0), ast.Number(1)),
		},
		{
			name: "builtin :le",
			str:  "0 <= 1",
			want: ast.NewAtom(":le", ast.Number(0), ast.Number(1)),
		},
		{
			name: "builtin :gt",
			str:  "1 > 0",
			want: ast.NewAtom(":gt", ast.Number(1), ast.Number(0)),
		},
		{
			name: "builtin :ge",
			str:  "1 >= 0",
			want: ast.NewAtom(":ge", ast.Number(1), ast.Number(0)),
		},
		{
			name: "atom (trailing comma)",
			str:  "foo(/bar,)",
			want: ast.NewAtom("foo", name("/bar")),
		},
		{
			name: "negated atom",
			str:  "!foo(/bar)",
			want: ast.NewNegAtom("foo", name("/bar")),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			term, err := LiteralOrFormula(test.str)
			if err != nil {
				t.Errorf("LiteralOrFormula(%v) failed with %v", test.str, err)
			} else if term == nil {
				t.Errorf("LiteralOrFormula(%v) = nil", test.str)
			} else if !term.Equals(test.want) {
				t.Errorf("LiteralOrFormula(%v) = %v wanted %v", test.str, term, test.want)
			}
		})
	}

}

func TestParseLiteralOrFormulaNegative(t *testing.T) {
	tests := []struct {
		name string
		str  string
	}{
		{
			name: "constant",
			str:  "/bar",
		},
		{
			name: "variable",
			str:  "X",
		},
		{
			name: "negated var",
			str:  "!X",
		},
		{
			name: "list is not a literal",
			str:  "[]",
		},
		{
			name: "nuclear equation",
			str:  "foo(X) = Y",
		},
		{
			name: "nuclear equation",
			str:  "X = foo(Y)",
		},
		{
			name: "missing body",
			str:  "X =",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if term, err := LiteralOrFormula(test.str); err == nil { // if no error
				t.Errorf("LiteralOrFormula(%v) = %v want error", test.str, term)
			}
		})
	}
}

func TestParseTermPositive(t *testing.T) {
	tests := []struct {
		name string
		str  string
		want ast.Term
	}{
		{
			name: "constant",
			str:  "/bar",
			want: name("/bar"),
		},
		{
			name: "constant 2",
			str:  "/bar/baz",
			want: name("/bar/baz"),
		},
		{
			name: "number",
			str:  "42",
			want: ast.Number(42),
		},
		{
			name: "negative number",
			str:  "-42",
			want: ast.Number(-42),
		},
		{
			name: "number constant",
			str:  "0000",
			want: ast.Number(0),
		},
		{
			name: "float64 constant",
			str:  "-100.123",
			want: ast.Float64(-100.123),
		},

		{
			name: "variable",
			str:  "X",
			want: ast.Variable{"X"},
		},
		{
			name: "atom",
			str:  "foo(/bar)",
			want: ast.NewAtom("foo", name("/bar")),
		},
		{
			name: "atom whitespace",
			str:  " foo( /bar  ,/baz  ,  /bak  )",
			want: ast.NewAtom("foo", name("/bar"), name("/baz"), name("/bak")),
		},
		{
			name: "atom number",
			str:  "foo(100)",
			want: ast.NewAtom("foo", ast.Number(100)),
		},
		{
			name: "atom float number",
			str:  "foo(.123)",
			want: ast.NewAtom("foo", ast.Float64(.123)),
		},
		{
			name: "constant skip whitespace",
			str:  "  /bar",
			want: name("/bar"),
		},
		{
			name: "empty list",
			str:  "[]",
			want: ast.ApplyFn{symbols.List, nil},
		},
		{
			name: "empty list arg",
			str:  "foo([])",
			want: ast.NewAtom("foo",
				ast.ApplyFn{symbols.List, nil}),
		},
		{
			name: "empty list arg fn",
			str:  "foo(fn:list())",
			want: ast.NewAtom("foo",
				ast.ApplyFn{ast.FunctionSym{"fn:list", 0}, nil}),
		},
		{
			name: "singleton list",
			str:  "[1]",
			want: ast.ApplyFn{symbols.List, []ast.BaseTerm{ast.Number(1)}},
		},
		{
			name: "singleton list (trailing comma)",
			str:  "[1,]",
			want: ast.ApplyFn{symbols.List, []ast.BaseTerm{ast.Number(1)}},
		},
		{
			name: "singleton list arg fn'",
			str:  "foo(fn:list(1))",
			want: ast.NewAtom("foo",
				ast.ApplyFn{ast.FunctionSym{"fn:list", 1}, []ast.BaseTerm{ast.Number(1)}}),
		},
		{
			name: "list of length 2",
			str:  "[1, /foo]",
			want: ast.ApplyFn{symbols.List, []ast.BaseTerm{ast.Number(1), name("/foo")}},
		},
		{
			name: "list of length 2 arg fn'",
			str:  "foo(fn:list(1,/bar))",
			want: ast.NewAtom("foo",
				ast.ApplyFn{ast.FunctionSym{"fn:list", 2}, []ast.BaseTerm{ast.Number(1), name("/bar")}}),
		},
		{
			name: "list 2",
			str:  "[[], [1, /foo]]",
			want: ast.ApplyFn{symbols.List, []ast.BaseTerm{
				ast.ApplyFn{symbols.List, nil},
				ast.ApplyFn{symbols.List, []ast.BaseTerm{ast.Number(1), name("/foo")}},
			}},
		},
		{
			name: "empty map",
			str:  "fn:map()",
			want: ast.ApplyFn{symbols.Map, nil},
		},
		{
			name: "map",
			str:  "[ 1 : 'one', 2 : 'two']",
			want: ast.ApplyFn{symbols.Map, []ast.BaseTerm{
				ast.Number(1), ast.String("one"), ast.Number(2), ast.String("two")}},
		},
		{
			name: "map (trailing comma)",
			str:  "[ 1 : 'one', 2 : 'two', ]",
			want: ast.ApplyFn{symbols.Map, []ast.BaseTerm{
				ast.Number(1), ast.String("one"), ast.Number(2), ast.String("two")}},
		},
		{
			name: "bad map - caught in validation",
			str:  "fn:map('foo')",
			want: ast.ApplyFn{symbols.Map, []ast.BaseTerm{ast.String("foo")}},
		},
		{
			name: "bad struct - caught in validation",
			str:  "fn:struct('foo')",
			want: ast.ApplyFn{symbols.Struct, []ast.BaseTerm{ast.String("foo")}},
		},
		{
			name: "struct",
			str:  "{}",
			want: ast.ApplyFn{symbols.Struct, nil},
		},
		{
			name: "struct2",
			str:  "{ /foo : 'bar', /bar : /baz, }",
			want: ast.ApplyFn{symbols.Struct, []ast.BaseTerm{name("/foo"), ast.String("bar"), name("/bar"), name("/baz")}},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			term, err := Term(test.str)
			if err != nil {
				t.Errorf("Term(%v) failed with %v", test.str, err)
			} else if term == nil {
				t.Errorf("Term(%v) = nil", test.str)
			} else if !term.Equals(test.want) {
				t.Errorf("Term(%q) = %v (%T) want %v (%T) ", test.str, term, term, test.want, test.want)
			}
		})
	}
}

func TestParseTermNegative(t *testing.T) {
	tests := []struct {
		name string
		str  string
	}{
		{
			name: "empty",
			str:  "",
		},
		{
			name: "bad constant",
			str:  "/",
		},
		{
			name: "bad constant 2",
			str:  "/bar/",
		},
		{
			name: "bad constant 3",
			str:  "//",
		},
		{
			name: "bad constant unterminated",
			str:  "/\"",
		},
		{
			name: "bad constant string part",
			str:  "/prefix/\"string\"",
		},
		{
			name: "negated constant?!",
			str:  "!/bar",
		},
		{
			name: "negated variable?!",
			str:  "!X",
		},
		{
			name: "not variable, not atom",
			str:  "x",
		},
		{
			name: "when does it end?",
			str:  "foo(/bar",
		},
		{
			name: "number too big",
			str:  "287326487236487264378264",
		},
		{
			name: "bad float ",
			str:  ".e",
		},
		{
			name: "number part",
			str:  "/catch/[22]",
		},
		{
			name: "bad map ",
			str:  "[ /foo : ]",
		},
		{
			name: "list/map with just trailing comma",
			str:  "[,]",
		},
		{
			name: "struct with just trailing comma",
			str:  "{,}",
		},
		{
			name: "bad struct",
			str:  "{ /foo : }",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got, err := Term(test.str); err == nil { // if no error
				t.Errorf("Term(%v) = %v want error", test.str, got)
			}
		})
	}
}

func TestRoundtrip(t *testing.T) {
	tests := []ast.Constant{
		ast.String("foo"),
		ast.String(`fo"\o`),
		ast.String(`"`),
	}
	for _, c := range tests {
		parsed, err := BaseTerm(c.String())
		if err != nil {
			t.Fatalf("BaseTerm(%v) failed with %v", c, err)
		}
		if !c.Equals(parsed) {
			t.Errorf("BaseTerm(%v) = %v want %v", c, parsed, c)
		}
	}
}

func TestPredicateNamePositive(t *testing.T) {
	tests := []struct {
		name string
		str  string
		want string
	}{
		{
			name: "empty",
			str:  "",
			want: "",
		},
		{
			name: "single letter name",
			str:  "x",
			want: "x",
		},
		{
			name: "simple name",
			str:  "xYz",
			want: "xYz",
		},
		{
			name: "with parameters",
			str:  "xYz(A,B,C)",
			want: "xYz",
		},
		{
			name: "when does it end?",
			str:  "foo(/bar",
			want: "foo",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := PredicateName(test.str)
			if err != nil {
				t.Errorf("PredicateName(%v) failed with %v", test.str, err)
			} else if got != test.want {
				t.Errorf("PredicateName(%v) = %q want %v", test.str, got, test.want)
			}
		})
	}
}

func TestPredicateNameNegative(t *testing.T) {
	tests := []struct {
		name string
		str  string
	}{
		{
			name: "bad constant",
			str:  "/",
		},
		{
			name: "bad constant 2",
			str:  "/bar/",
		},
		{
			name: "bad constant 3",
			str:  "//",
		},
		{
			name: "bad constant unterminated",
			str:  "/\"",
		},
		{
			name: "bad constant string part",
			str:  "/prefix/\"string\"",
		},
		{
			name: "negated constant?!",
			str:  "!/bar",
		},
		{
			name: "negated variable?!",
			str:  "!X",
		},

		{
			name: "number too big",
			str:  "287326487236487264378264",
		},
		{
			name: "number part",
			str:  "/catch/[22]",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got, err := PredicateName(test.str); err == nil && got != "" { // if no error
				t.Errorf("PredicateName(%v) = %q want error or empty string", test.str, got)
			}
		})
	}
}
