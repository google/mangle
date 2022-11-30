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

package packages

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/mangle/ast"
	"github.com/google/mangle/parse"
)

func makeDecl(t *testing.T, atom ast.Atom, descrAtoms []ast.Atom, bounds []ast.BoundDecl, constraints *ast.InclusionConstraint) ast.Decl {
	t.Helper()
	decl, err := ast.NewDecl(atom, descrAtoms, bounds, constraints)
	if err != nil {
		t.Fatal(err)
	}
	return decl
}

func TestMerge(t *testing.T) {
	tests := []struct {
		desc  string
		input Package
		other Package
		want  Package
	}{
		{
			desc: "units are merged",
			input: Package{
				Name: "",
				units: []parse.SourceUnit{
					{
						Decls: []ast.Decl{
							makeDecl(t, ast.NewAtom("some_decl"), nil, nil, nil),
						},
						Clauses: []ast.Clause{
							ast.NewClause(ast.NewAtom("some_clause"), nil),
						},
					},
					{
						Decls: []ast.Decl{
							makeDecl(t, ast.NewAtom("some_decl_in_another_unit"), nil, nil, nil),
						},
					},
				},
			},
			other: Package{
				Name: "",
				units: []parse.SourceUnit{
					{
						Decls: []ast.Decl{
							makeDecl(t, ast.NewAtom("some_decl"), nil, nil, nil),
						},
						Clauses: []ast.Clause{
							ast.NewClause(ast.NewAtom("some_clause"), nil),
						},
					},
					{
						Decls: []ast.Decl{
							makeDecl(t, ast.NewAtom("some_decl_in_another_unit"), nil, nil, nil),
						},
					},
				},
			},
			want: Package{
				Name: "",
				units: []parse.SourceUnit{
					{
						Decls: []ast.Decl{
							makeDecl(t, ast.NewAtom("some_decl"), nil, nil, nil),
						},
						Clauses: []ast.Clause{
							ast.NewClause(ast.NewAtom("some_clause"), nil),
						},
					},
					{
						Decls: []ast.Decl{
							makeDecl(t, ast.NewAtom("some_decl_in_another_unit"), nil, nil, nil),
						},
					},
					{
						Decls: []ast.Decl{
							makeDecl(t, ast.NewAtom("some_decl"), nil, nil, nil),
						},
						Clauses: []ast.Clause{
							ast.NewClause(ast.NewAtom("some_clause"), nil),
						},
					},
					{
						Decls: []ast.Decl{
							makeDecl(t, ast.NewAtom("some_decl_in_another_unit"), nil, nil, nil),
						},
					},
				},
			},
		},
		{
			desc: "atoms are merged",
			input: Package{
				Name: "",
				Atoms: []ast.Atom{
					ast.NewAtom("atom1"),
				},
			},
			other: Package{
				Name: "",
				Atoms: []ast.Atom{
					ast.NewAtom("atom2"),
				},
			},
			want: Package{
				Name: "",
				Atoms: []ast.Atom{
					ast.NewAtom("atom1"),
					ast.NewAtom("atom2"),
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			err := test.input.Merge(test.other)
			if err != nil {
				t.Fatalf("unexpected error: %v, err", err)
			}
			if diff := cmp.Diff(test.want, test.input, cmp.AllowUnexported(Package{}, ast.Constant{})); diff != "" {
				t.Errorf("Merge() diff (-want +got):\n%s", diff)
			}
		})
	}

}

func TestDecls(t *testing.T) {
	tests := []struct {
		desc  string
		input Package
		want  []ast.Decl
	}{
		{
			desc: "no package name, declarations are returned as is",
			input: Package{
				Name: "",
				units: []parse.SourceUnit{
					{
						Decls: []ast.Decl{
							makeDecl(t, ast.NewAtom("some_decl"), nil, nil, nil),
						},
					},
					{
						Decls: []ast.Decl{
							makeDecl(t, ast.NewAtom("some_decl_in_another_unit"), nil, nil, nil),
						},
					},
				},
			},
			want: []ast.Decl{
				makeDecl(t, ast.NewAtom("some_decl"), nil, nil, nil),
				makeDecl(t, ast.NewAtom("some_decl_in_another_unit"), nil, nil, nil),
			},
		},
		{
			desc: "has package name, declarations are prefixed",
			input: Package{
				Name: "foo.bar",
				units: []parse.SourceUnit{
					{
						Decls: []ast.Decl{
							makeDecl(t, ast.NewAtom("some_decl"), nil, nil, nil),
						},
					},
					{
						Decls: []ast.Decl{
							makeDecl(t, ast.NewAtom("some_decl_in_another_unit"), nil, nil, nil),
						},
					},
				},
			},
			want: []ast.Decl{
				makeDecl(t, ast.NewAtom("foo.bar.some_decl"), nil, nil, nil),
				makeDecl(t, ast.NewAtom("foo.bar.some_decl_in_another_unit"), nil, nil, nil),
			},
		},
		{
			desc: "bound uses declared decl",
			input: Package{
				Name: "foo.bar",
				units: []parse.SourceUnit{
					{
						Decls: []ast.Decl{
							makeDecl(t, ast.NewAtom("some_decl", ast.Variable{"X"}), nil, nil, nil),
							makeDecl(t, ast.NewAtom("some_decl_with_bound"), nil,
								[]ast.BoundDecl{{Bounds: []ast.BaseTerm{ast.String("some_decl")}}}, nil),
						},
					},
				},
			},
			want: []ast.Decl{
				makeDecl(t, ast.NewAtom("foo.bar.some_decl", ast.Variable{"X"}), nil, nil, nil),
				makeDecl(t, ast.NewAtom("foo.bar.some_decl_with_bound"), nil,
					[]ast.BoundDecl{{Bounds: []ast.BaseTerm{ast.String("foo.bar.some_decl")}}}, nil),
			},
		},
		{
			desc: "package decl is not included",
			input: Package{
				Name: "foo.bar",
				units: []parse.SourceUnit{
					{
						Decls: []ast.Decl{
							makeDecl(t, ast.NewAtom("Package"), []ast.Atom{ast.NewAtom("name", ast.String("does.not.matter1"))}, nil, nil),
						},
					},
					{
						Decls: []ast.Decl{
							makeDecl(t, ast.NewAtom("Package"), []ast.Atom{ast.NewAtom("name", ast.String("does.not.matter2"))}, nil, nil),
						},
					},
				},
			},
			want: []ast.Decl{},
		},
	}
	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			got, err := test.input.Decls()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if diff := cmp.Diff(test.want, got, cmp.AllowUnexported(ast.Constant{})); diff != "" {
				t.Errorf("Decls() diff (-want +got):\n%s", diff)
			}
		})
	}

}

func TestDeclsErrors(t *testing.T) {
	tests := []struct {
		desc  string
		input Package
		want  []ast.Decl
	}{
		{
			desc: "Use has name descr atom with wrong type",
			input: Package{
				Name: "foo.bar",
				units: []parse.SourceUnit{
					{
						Decls: []ast.Decl{
							makeDecl(t, ast.NewAtom("Use"), []ast.Atom{ast.NewAtom("name", ast.Number(1))}, nil, nil),
						},
					},
				},
			},
		},
		{
			desc: "Use has atom with unexpected args length",
			input: Package{
				Name: "foo.bar",
				units: []parse.SourceUnit{
					{
						Decls: []ast.Decl{
							makeDecl(t, ast.NewAtom("Use"), []ast.Atom{{ast.PredicateSym{Symbol: "name", Arity: 1}, []ast.BaseTerm{}}}, nil, nil),
						},
					},
				},
			},
		},
		{
			desc: "Bound uses package but there is no Use Decl",
			input: Package{
				Name: "foo.bar",
				units: []parse.SourceUnit{
					{
						Decls: []ast.Decl{
							makeDecl(t, ast.NewAtom("some_decl_with_bound"), nil,
								[]ast.BoundDecl{{Bounds: []ast.BaseTerm{ast.String("package.id")}}}, nil),
						},
					},
				},
			},
		},
		{
			desc: "Used package is the same current package",
			input: Package{
				Name: "foo.bar",
				units: []parse.SourceUnit{
					{
						Decls: []ast.Decl{
							makeDecl(t, ast.NewAtom("Use"), []ast.Atom{ast.NewAtom("name", ast.String("foo.bar"))}, nil, nil),
						},
					},
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			_, err := test.input.Decls()
			if err == nil {
				t.Fatal("Decls() returned no error")
			}
		})
	}

}

func TestClauses(t *testing.T) {
	tests := []struct {
		desc  string
		input Package
		want  []ast.Clause
	}{
		{
			desc: "no package name, clauses are not rewritten",
			input: Package{
				Name: "",
				units: []parse.SourceUnit{
					{
						Clauses: []ast.Clause{
							ast.NewClause(ast.NewAtom("clause_defined_here"), []ast.Term{ast.NewAtom("other_clause")}),
						},
					},
					{
						Clauses: []ast.Clause{
							ast.NewClause(ast.NewAtom("clause_also_defined_here"), []ast.Term{ast.NewAtom("other_clause")}),
						},
					},
				},
			},
			want: []ast.Clause{
				ast.NewClause(ast.NewAtom("clause_defined_here"), []ast.Term{ast.NewAtom("other_clause")}),
				ast.NewClause(ast.NewAtom("clause_also_defined_here"), []ast.Term{ast.NewAtom("other_clause")}),
			},
		},
		{
			desc: "references to predicates outside the package are left as-is",
			input: Package{
				Name: "foo.bar",
				units: []parse.SourceUnit{
					{
						Clauses: []ast.Clause{
							ast.NewClause(ast.NewAtom("clause_defined_here"), []ast.Term{ast.NewAtom("other_clause")}),
						},
					},
					{
						Clauses: []ast.Clause{
							ast.NewClause(ast.NewAtom("clause_also_defined_here"), []ast.Term{ast.NewAtom("other_clause")}),
						},
					},
				},
			},
			want: []ast.Clause{
				ast.NewClause(ast.NewAtom("foo.bar.clause_defined_here"), []ast.Term{ast.NewAtom("other_clause")}),
				ast.NewClause(ast.NewAtom("foo.bar.clause_also_defined_here"), []ast.Term{ast.NewAtom("other_clause")}),
			},
		},
		{
			desc: "clauses defined in this package are rewritten",
			input: Package{
				Name: "foo.bar",
				units: []parse.SourceUnit{
					{
						Clauses: []ast.Clause{
							ast.NewClause(ast.NewAtom("other_clause"), []ast.Term{ast.String("here")}),
							ast.NewClause(ast.NewAtom("clause_defined_here"), []ast.Term{ast.NewAtom("other_clause")}),
						},
					},
					{
						Clauses: []ast.Clause{
							ast.NewClause(ast.NewAtom("clause_also_defined_here"), []ast.Term{ast.NewAtom("other_clause")}),
						},
					},
				},
			},
			want: []ast.Clause{
				ast.NewClause(ast.NewAtom("foo.bar.other_clause"), []ast.Term{ast.String("here")}),
				ast.NewClause(ast.NewAtom("foo.bar.clause_defined_here"), []ast.Term{ast.NewAtom("foo.bar.other_clause")}),
				ast.NewClause(ast.NewAtom("foo.bar.clause_also_defined_here"), []ast.Term{ast.NewAtom("foo.bar.other_clause")}),
			},
		},
		{
			desc: "clause with a negation is rewritten",
			input: Package{
				Name: "foo.bar",
				units: []parse.SourceUnit{
					{
						Clauses: []ast.Clause{
							ast.NewClause(ast.NewAtom("clause_defined_here"), []ast.Term{ast.NewNegAtom("other_clause")}),
							ast.NewClause(ast.NewAtom("other_clause"), []ast.Term{ast.String("here")}),
						},
					},
				},
			},
			want: []ast.Clause{
				ast.NewClause(ast.NewAtom("foo.bar.clause_defined_here"), []ast.Term{ast.NewNegAtom("foo.bar.other_clause")}),
				ast.NewClause(ast.NewAtom("foo.bar.other_clause"), []ast.Term{ast.String("here")}),
			},
		},
		{
			desc: "clause with corresponding use in RHS is allowed",
			input: Package{
				Name: "foo.bar",
				units: []parse.SourceUnit{
					{
						Decls: []ast.Decl{
							makeDecl(t, ast.NewAtom("Use"), []ast.Atom{ast.NewAtom("name", ast.String("package"))}, nil, nil),
						},
						Clauses: []ast.Clause{
							ast.NewClause(ast.NewAtom("clause_defined_here"), []ast.Term{ast.NewAtom("package.other_clause")}),
						},
					},
				},
			},
			want: []ast.Clause{
				ast.NewClause(ast.NewAtom("foo.bar.clause_defined_here"), []ast.Term{ast.NewAtom("package.other_clause")}),
			},
		},
		{
			desc: "clauses are also rewritten if the decl was declared in this package",
			input: Package{
				Name: "foo.bar",
				units: []parse.SourceUnit{
					{
						Clauses: []ast.Clause{
							ast.NewClause(ast.NewAtom("clause"), []ast.Term{ast.NewAtom("from_decl")}),
						},
						Decls: []ast.Decl{
							makeDecl(t, ast.NewAtom("from_decl"), nil, nil, nil),
						},
					},
				},
			},
			want: []ast.Clause{
				ast.NewClause(ast.NewAtom("foo.bar.clause"), []ast.Term{ast.NewAtom("foo.bar.from_decl")}),
			},
		},
		{
			desc: "decl has different arity",
			input: Package{
				Name: "foo.bar",
				units: []parse.SourceUnit{
					{
						Clauses: []ast.Clause{
							ast.NewClause(ast.NewAtom("clause"), []ast.Term{ast.NewAtom("same_name_different_arity")}),
						},
						Decls: []ast.Decl{
							makeDecl(t, ast.NewAtom("same_name_different_arity", ast.Variable{"X"}, ast.Variable{"Y"}), nil, nil, nil),
						},
					},
				},
			},
			want: []ast.Clause{
				ast.NewClause(ast.NewAtom("foo.bar.clause"), []ast.Term{ast.NewAtom("same_name_different_arity")}),
			},
		},
	}
	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			got, err := test.input.Clauses()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if diff := cmp.Diff(test.want, got, cmp.AllowUnexported(ast.Constant{})); diff != "" {
				t.Fatalf("Clauses() diff (-want +got):\n%s", diff)
			}
		})
	}

}

func TestClausesErrors(t *testing.T) {
	tests := []struct {
		desc  string
		input Package
	}{
		{
			desc: "Use has name descr atom with wrong type",
			input: Package{
				Name: "foo.bar",
				units: []parse.SourceUnit{
					{
						Decls: []ast.Decl{
							makeDecl(t, ast.NewAtom("Use"), []ast.Atom{ast.NewAtom("name", ast.Number(1))}, nil, nil),
						},
					},
				},
			},
		},
		{
			desc: "Use has atom with unexpected args length",
			input: Package{
				Name: "foo.bar",
				units: []parse.SourceUnit{
					{
						Decls: []ast.Decl{
							makeDecl(t, ast.NewAtom("Use"), []ast.Atom{{ast.PredicateSym{Symbol: "name", Arity: 1}, []ast.BaseTerm{}}}, nil, nil),
						},
					},
				},
			},
		},
		{
			desc: "Clause has reference to used package that has not been declared",
			input: Package{
				Name: "foo.bar",
				units: []parse.SourceUnit{
					{
						Clauses: []ast.Clause{
							ast.NewClause(ast.NewAtom("clause_defined_here"), []ast.Term{ast.NewAtom("package.other_clause")}),
						},
					},
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			if _, err := test.input.Clauses(); err == nil {
				t.Fatal("expected error but Clauses() returned none")
			}
		})
	}

}

func TestExtract(t *testing.T) {
	tests := []struct {
		desc  string
		input parse.SourceUnit
		want  Package
	}{
		{
			desc: "package decl contains extra atoms",
			input: parse.SourceUnit{
				Clauses: []ast.Clause{},
				Decls: []ast.Decl{
					makeDecl(t, ast.NewAtom("Package"), []ast.Atom{ast.NewAtom("name", ast.String("foo.bar")), ast.NewAtom("extra", ast.String("string"))}, nil, nil),
				},
			},
			want: Package{
				Name:  "foo.bar",
				Atoms: []ast.Atom{ast.NewAtom("extra", ast.String("string"))},
				units: []parse.SourceUnit{
					{
						Clauses: []ast.Clause{},
						Decls: []ast.Decl{
							makeDecl(t, ast.NewAtom("Package"), []ast.Atom{ast.NewAtom("name", ast.String("foo.bar")), ast.NewAtom("extra", ast.String("string"))}, nil, nil),
						},
					},
				},
			},
		},
		{
			desc: "contains package decl",
			input: parse.SourceUnit{
				Clauses: []ast.Clause{
					ast.NewClause(ast.NewAtom("head", ast.Variable{"X"}), []ast.Term{ast.NewAtom("expanded", ast.Variable{"X"})}),
					ast.NewClause(ast.NewAtom("expanded", ast.Variable{"X"}), []ast.Term{ast.String("constant")}),
				},
				Decls: []ast.Decl{
					makeDecl(t, ast.NewAtom("head"), nil, nil, nil),
					makeDecl(t, ast.NewAtom("Package"), []ast.Atom{ast.NewAtom("name", ast.String("foo.bar"))}, nil, nil),
				},
			},
			want: Package{
				Name:  "foo.bar",
				Atoms: []ast.Atom{},
				units: []parse.SourceUnit{
					{
						Clauses: []ast.Clause{
							ast.NewClause(ast.NewAtom("head", ast.Variable{"X"}), []ast.Term{ast.NewAtom("expanded", ast.Variable{"X"})}),
							ast.NewClause(ast.NewAtom("expanded", ast.Variable{"X"}), []ast.Term{ast.String("constant")}),
						},
						Decls: []ast.Decl{
							makeDecl(t, ast.NewAtom("head"), nil, nil, nil),
							makeDecl(t, ast.NewAtom("Package"), []ast.Atom{ast.NewAtom("name", ast.String("foo.bar"))}, nil, nil),
						},
					},
				},
			},
		},
		{
			desc: "does not contain package decl",
			input: parse.SourceUnit{
				Clauses: []ast.Clause{
					ast.NewClause(ast.NewAtom("head", ast.Variable{"X"}), []ast.Term{ast.NewAtom("rhs", ast.Variable{"X"})}),
				},
				Decls: []ast.Decl{
					makeDecl(t, ast.NewAtom("head"), nil, nil, nil),
				},
			},
			want: Package{
				Name:  "",
				Atoms: []ast.Atom{},
				units: []parse.SourceUnit{
					{
						Clauses: []ast.Clause{
							ast.NewClause(ast.NewAtom("head", ast.Variable{"X"}), []ast.Term{ast.NewAtom("rhs", ast.Variable{"X"})}),
						},
						Decls: []ast.Decl{
							makeDecl(t, ast.NewAtom("head"), nil, nil, nil),
						},
					},
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			got, err := Extract(test.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if diff := cmp.Diff(test.want, got, cmp.AllowUnexported(Package{}, ast.Constant{})); diff != "" {
				t.Errorf("Extract() diff (-want +got):\n%s", diff)
			}
		})
	}
}
