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
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/mangle/ast"
	"github.com/google/mangle/parse"
	"github.com/google/mangle/symbols"
)

func name(str string) ast.Constant {
	n, err := ast.Name(str)
	if err != nil {
		panic(fmt.Errorf("bad name in test case: %s got %w", str, err))
	}
	return n
}

func fml(str string) ast.Term {
	t, err := parse.LiteralOrFormula(str)
	if err != nil {
		panic(fmt.Errorf("bad fml syntax in test case: %s got %w", str, err))
	}
	return t
}

func atom(str string) ast.Atom {
	term, err := parse.Term(str)
	if err != nil {
		panic(fmt.Errorf("bad syntax in test case: %s got %w", str, err))
	}
	atom, ok := term.(ast.Atom)
	if !ok {
		panic(fmt.Errorf("not an atom test case: %v", term))
	}
	return atom
}

func makeDecl(t *testing.T, a ast.Atom, descr []ast.Atom, boundDecls []ast.BoundDecl, incl *ast.InclusionConstraint) ast.Decl {
	t.Helper()
	decl, err := ast.NewDecl(a, descr, boundDecls, incl)
	if err != nil {
		t.Fatal(err)
	}
	return decl
}

func makeSyntheticDecl(t *testing.T, a ast.Atom) ast.Decl {
	t.Helper()
	decl, err := ast.NewSyntheticDecl(a)
	if err != nil {
		t.Fatal(err)
	}
	return decl
}

func TestCheckRulePositive(t *testing.T) {
	tests := []ast.Clause{
		clause("foo(/bar)."),
		clause("foo(X) :- bar(X)."),
		clause("foo(X) :- bar(X), bar(_)."),
		clause("foo(X) :- X = /bar ."),
		clause("foo(X) :- X = Y, bar(Y)."),
		clause("foo(X) :- X = Y, Y = /bar ."),
		clause("foo(X) :- Z = 2, bar(X), X < Z."),
		clause("foo(X) :- X = fn:list:cons(fn:list:get([1], 1), [])."),
		clause("foo([])."),
		clause("foo([23])."),
		clause("foo(fn:cons(1, [23]))."),
		clause("foo({/foo: 37})."),
		clause("foo([37: /foo])."),
		clause("foo(X) :- X = [37]."),
		clause("foo(X) :- Y = 2, X = [Y]."),
		clause("foo(X) :- Y = 2 |> let X = [Y]."),
		clause("foo(X) :- Y = 2, X = fn:list(Y)."),
		clause("foo(X) :- X = {/foo: 37}."),
		clause("foo(X) :- X = [37: /foo]."),
		clause("foo(X) :- bar(X), 0 < X."),
		clause("foo(Y) :- bar(X) |> let Y = fn:plus(X, X)."),
		clause("foo(Y) :- bar(X) |> let Y = fn:list:get(X, 1)."),
		clause("foo(Y) :- bar(X) |> do fn:group_by(X), let Y = fn:sum(X)."),
		clause("c(R,S,T) :- bar(R), bar(S), bar(T), fn:plus(fn:mult(R, R), fn:mult(S,S)) = fn:mult(T,T)."),
		clause("foo(X) :- bar(Y), bar(Z) |> do fn:group_by(Y), let X = fn:collect(Z)."),
	}
	for _, clause := range tests {
		analyzer, _ := New(map[ast.PredicateSym]ast.Decl{
			ast.PredicateSym{"bar", 1}: makeSyntheticDecl(t, atom("bar(X)")),
		}, nil, ErrorForBoundsMismatch)
		if err := analyzer.CheckRule(clause); err != nil {
			t.Errorf("Expected rule %v to be valid, got %v", clause, err)
		}
	}
}

func TestCheckRuleBindsByDecl(t *testing.T) {
	d, err := ast.NewDecl(
		atom("bar(Struct, FieldValue)"),
		[]ast.Atom{
			atom("mode('+', '+')"),
			atom("mode('+', '-')"),
		},
		[]ast.BoundDecl{ast.NewBoundDecl(ast.AnyBound), ast.NewBoundDecl(ast.AnyBound)},
		nil)
	if err != nil {
		t.Fatal(err)
	}

	analyzer, err := New(nil, []ast.Decl{d}, ErrorForBoundsMismatch)
	if err != nil {
		t.Fatalf("New(nil, %v, ErrorForBoundsMismatch) returned unexpected error: %v", d, err)
	}

	// :match_field expects first argument to be bound
	clause := clause("bar(S, F) :- :match_field(S, /random, F).")
	if err := analyzer.CheckRule(clause); err != nil {
		t.Errorf("CheckRule(%v) returned unexpected error: %v", clause, err)
	}
}

func TestCheckRulePositiveExtraFun(t *testing.T) {
	tests := []ast.Clause{
		clause("foo(Y) :- bar(X) |> let Y = fn:plus(X, X), let _ = fn:ring_the_alarm()."),
	}
	for _, clause := range tests {
		analyzer, _ := New(map[ast.PredicateSym]ast.Decl{
			ast.PredicateSym{"bar", 1}: makeSyntheticDecl(t, atom("bar(X)")),
		}, nil, ErrorForBoundsMismatch)
		analyzer.extraFunctions = map[ast.FunctionSym]ast.BaseTerm{
			ast.FunctionSym{"fn:ring_the_alarm", 0}: ast.AnyBound,
		}
		err := analyzer.CheckRule(clause)
		if err != nil {
			t.Errorf("Expected rule %v to be valid, got %v", clause, err)
		}
	}
}

func TestCheckRuleNegative(t *testing.T) {
	tests := []ast.Clause{
		clause("foo(X)."),
		clause("foo(_)."),
		clause("foo(_) :- bar(_)."),
		clause("foo(X, _) :- foo(2, 1)."),
		clause("foo(X) :- bar(Y)."),
		clause("foo(X) :- X != Y, bar(Y)."),
		clause("foo(X) :- X = Y."),
		clause("foo(X) :- X < Y, foo(Y)."),
		clause("foo(X) :- X = [Y]."),
		clause("foo(X) :- X = fn:list(Y)."),
		clause("foo(X) :- X = fn:list:get(1)."),
		clause("foo(X) :- X = fn:list(fn:list:get(1))."),
		clause("foo(X) :- X = fn:list(_)."),
		// Head variable neither defined by body nor transform.
		clause("foo(A) :- bar(X) |> let Y = fn:plus(X, X)."),
		// Wildcard is like a variable neither defined by body nor transform.
		clause("foo(Y) :- bar(X) |> let Y = fn:plus(X, _)."),
		// Transform variable not defined by body.
		clause("foo(Y) :- bar(X) |> do fn:group_by(A)."),
		// Wildcard is like a variable not defined by body.
		clause("foo(Y) :- bar(X) |> do fn:group_by(_)."),
		// Wildcard references do not make sense in a transform.
		clause("foo(Y) :- bar(X) |> let Y = fn:plus(X, X), let Z = fn:ring_the_alarm(_)."),
		// A transform may not redefine a variable from body.
		clause("foo(X) :- bar(X) |> let X = fn:plus(X, X)."),
		// Wrong arity
		clause("foo(Y) :- bar(X) |> let Y = fn:list:get(1)."),
		// fn:collect() needs at least one argument.
		clause("foo(X) :- bar(Y) |> do fn:group_by(), let X = fn:collect()."),
		// fn:collect_distinct() needs at least one argument.
		clause("foo(X) :- bar(Y) |> do fn:group_by(), let X = fn:collect_distinct()."),
		// In a reducer, need to refer to variables from body.
		clause("foo(X) :- bar(Y), baz(Z) |> do fn:group_by(Y), let X = fn:collect(X)."),
	}
	for _, clause := range tests {
		analyzer, _ := New(map[ast.PredicateSym]ast.Decl{
			ast.PredicateSym{"bar", 1}: makeSyntheticDecl(t, atom("bar(X)")),
		}, nil, ErrorForBoundsMismatch)
		if err := analyzer.CheckRule(clause); err == nil {
			t.Errorf("Expected error for rule %v", clause)
		}
	}
}

func mustDesugar(t *testing.T, decls map[ast.PredicateSym]ast.Decl) map[ast.PredicateSym]*ast.Decl {
	t.Helper()
	desugaredDecls, err := symbols.CheckAndDesugar(decls)
	if err != nil {
		t.Fatal(err)
	}
	return desugaredDecls
}

func TestAnalyzePositive(t *testing.T) {
	privateDecl := makeDecl(t, atom("foo.baz(X)"), []ast.Atom{atom("private()")}, nil, nil)
	privateDeclEmptyPackage := makeDecl(t, atom("baz(X)"), []ast.Atom{atom("private()")}, nil, nil)
	tests := []struct {
		descr           string
		knownPredicates map[ast.PredicateSym]ast.Decl
		decls           []ast.Decl
		program         []ast.Clause
		want            ProgramInfo
	}{
		{
			descr:           "self-contained program with two clauses, no decls.",
			knownPredicates: nil,
			decls:           nil,
			program: []ast.Clause{
				clause("foo(/bar)."),
				clause("sna(X) :- foo(X)."),
			},
			want: ProgramInfo{
				IdbPredicates: map[ast.PredicateSym]struct{}{
					ast.PredicateSym{"sna", 1}: {},
				},
				EdbPredicates: map[ast.PredicateSym]struct{}{
					ast.PredicateSym{"foo", 1}: {},
				},
				InitialFacts: []ast.Atom{atom("foo(/bar)")},
				Rules:        []ast.Clause{clause("sna(X) :- foo(X).")},
				Decls: mustDesugar(t, map[ast.PredicateSym]ast.Decl{
					ast.PredicateSym{"foo", 1}: makeSyntheticDecl(t, atom("foo(X0)")),
					ast.PredicateSym{"sna", 1}: makeSyntheticDecl(t, atom("sna(X)")),
				}),
			},
		},
		{
			descr: "overriding a synthetic decl is permitted",
			decls: []ast.Decl{makeDecl(t, atom("foo(X,Y)"), nil, nil, nil)},
			knownPredicates: map[ast.PredicateSym]ast.Decl{
				ast.PredicateSym{"foo", 2}: makeSyntheticDecl(t, atom("foo(X,Y)")),
			},
			program: nil,
			want: ProgramInfo{
				IdbPredicates: map[ast.PredicateSym]struct{}{},
				EdbPredicates: map[ast.PredicateSym]struct{}{},
				Decls: mustDesugar(t, map[ast.PredicateSym]ast.Decl{
					ast.PredicateSym{"foo", 2}: makeDecl(t, atom("foo(X,Y)"), nil, nil, nil),
				}),
			},
		},
		{
			descr:           "foo.baz is private, but referencing predicate has same prefix",
			knownPredicates: nil,
			decls:           []ast.Decl{privateDecl},
			program: []ast.Clause{
				clause("foo.baz(1)."),
				clause("foo.bar(X) :- foo.baz(X)."),
			},
			want: ProgramInfo{
				IdbPredicates: map[ast.PredicateSym]struct{}{
					ast.PredicateSym{"foo.bar", 1}: {},
				},
				EdbPredicates: map[ast.PredicateSym]struct{}{
					ast.PredicateSym{"foo.baz", 1}: {},
				},
				InitialFacts: []ast.Atom{atom("foo.baz(1)")},
				Rules:        []ast.Clause{clause("foo.bar(X) :- foo.baz(X).")},
				Decls: mustDesugar(t, map[ast.PredicateSym]ast.Decl{
					ast.PredicateSym{"foo.baz", 1}: privateDecl,
					ast.PredicateSym{"foo.bar", 1}: makeSyntheticDecl(t, atom("foo.bar(X)")),
				}),
			},
		},
		{
			descr:           "baz is private, but is referenced from the empty package",
			knownPredicates: nil,
			decls:           []ast.Decl{privateDeclEmptyPackage},
			program: []ast.Clause{
				clause("baz(1)."),
				clause("bar(X) :- baz(X)."),
			},
			want: ProgramInfo{
				IdbPredicates: map[ast.PredicateSym]struct{}{
					ast.PredicateSym{"bar", 1}: {},
				},
				EdbPredicates: map[ast.PredicateSym]struct{}{
					ast.PredicateSym{"baz", 1}: {},
				},
				InitialFacts: []ast.Atom{atom("baz(1)")},
				Rules:        []ast.Clause{clause("bar(X) :- baz(X).")},
				Decls: mustDesugar(t, map[ast.PredicateSym]ast.Decl{
					ast.PredicateSym{"baz", 1}: privateDeclEmptyPackage,
					ast.PredicateSym{"bar", 1}: makeSyntheticDecl(t, atom("bar(X)")),
				}),
			},
		},
		{
			descr: "match",
			decls: []ast.Decl{makeDecl(t, atom("input(X)"), nil, []ast.BoundDecl{
				ast.NewBoundDecl(ast.ApplyFn{symbols.ListType, []ast.BaseTerm{ast.NameBound}}),
			}, nil)},
			program: []ast.Clause{
				clause("starts_with_a(X) :- input(X), :match_cons(X, Y, Z), Y = /a ."),
			},
			want: ProgramInfo{
				IdbPredicates: map[ast.PredicateSym]struct{}{
					ast.PredicateSym{"starts_with_a", 1}: struct{}{},
				},
				EdbPredicates: map[ast.PredicateSym]struct{}{
					ast.PredicateSym{"input", 1}: struct{}{},
				},
				InitialFacts: nil,
				Rules: []ast.Clause{
					clause("starts_with_a(X) :- input(X), :match_cons(X, Y, Z), Y = /a ."),
				},
				Decls: mustDesugar(t, map[ast.PredicateSym]ast.Decl{
					ast.PredicateSym{"starts_with_a", 1}: makeSyntheticDecl(t, atom("starts_with_a(X)")),
					ast.PredicateSym{"input", 1}: makeDecl(t, atom("input(X)"), nil, []ast.BoundDecl{
						ast.NewBoundDecl(ast.ApplyFn{symbols.ListType, []ast.BaseTerm{ast.NameBound}}),
					}, nil),
				}),
			},
		},
		{
			descr: "empty array is evaluated",
			program: []ast.Clause{
				clause("a_list([])."),
			},
			want: ProgramInfo{
				IdbPredicates: map[ast.PredicateSym]struct{}{},
				EdbPredicates: map[ast.PredicateSym]struct{}{
					ast.PredicateSym{"a_list", 1}: struct{}{},
				},
				Decls: mustDesugar(t, map[ast.PredicateSym]ast.Decl{
					ast.PredicateSym{"a_list", 1}: makeSyntheticDecl(t, atom("a_list(X0)")),
				}),
				InitialFacts: []ast.Atom{{Predicate: ast.PredicateSym{"a_list", 1}, Args: []ast.BaseTerm{ast.ListNil}}},
			},
		},
	}
	for _, test := range tests {
		got, err := AnalyzeOneUnit(parse.SourceUnit{Clauses: test.program, Decls: test.decls}, test.knownPredicates)
		if err != nil {
			t.Errorf("expected program to be valid, got error %v", err)
		}
		if diff := cmp.Diff(test.want, *got, cmp.AllowUnexported(ast.Constant{})); diff != "" {
			t.Errorf("Analyze(%v,%v,%v) returned diff:\n%s", test.program, test.knownPredicates, test.decls, diff)
		}
	}
}

func makeNonSynthetic(t *testing.T, decl ast.Decl) ast.Decl {
	return makeDecl(t, decl.DeclaredAtom, nil, decl.Bounds, decl.Constraints)
}

func TestAnalyzeNegative(t *testing.T) {
	fooDecl := makeSyntheticDecl(t, atom("foo(X,Y)"))
	fooThreeArgs := makeSyntheticDecl(t, atom("foo(X, Y, Z)"))
	privateDecl := makeDecl(t, atom("bar.pred(X)"), []ast.Atom{atom("private()")}, nil, nil)
	tests := []struct {
		descr           string
		knownPredicates map[ast.PredicateSym]ast.Decl
		decls           []ast.Decl
		program         []ast.Clause
	}{
		{
			descr:           "foo not defined, referenced in premise",
			knownPredicates: nil,
			decls:           nil,
			program: []ast.Clause{
				clause("sna(X) :- foo(X)."),
			},
		},
		{
			descr:           "bar.pred is private",
			knownPredicates: nil,
			decls:           []ast.Decl{privateDecl},
			program: []ast.Clause{
				clause("bar.pred(1)."),
				clause("foo.pred(X) :- bar.pred(X)."),
			},
		},
		{
			descr:           "foo defined with different arity",
			knownPredicates: nil,
			decls:           nil,
			program: []ast.Clause{
				clause("foo(1)."),
				clause("foo(1, 'a')."),
			}},
		{
			descr:           "foo not defined, referenced in transform",
			knownPredicates: nil,
			decls:           nil,
			program: []ast.Clause{
				// Function "foo" is not defined.
				clause("sna(X) :- sna(X) |> do fn:foo(X)."),
			}},
		{
			descr:           ":match variables have to be distinct",
			knownPredicates: nil,
			decls:           nil,
			program: []ast.Clause{
				clause("input([/foo])."),
				clause("do_the_match(X, Y) :- input(X), :match(X, Y, Y)."),
			}},
		{
			descr:           ":match matched variable has to be distinct",
			knownPredicates: nil,
			decls:           nil,
			program: []ast.Clause{
				clause("input([/foo])."),
				clause("do_the_match(X, Y) :- input(X), :match(X, Y, X)."),
			}},
		{
			descr:           "variable Y does not appear anywhere",
			knownPredicates: nil,
			decls:           nil,
			program: []ast.Clause{
				clause("ok(/foo)."),
				clause("sna(Y) :- ok(X) |> do fn:group_by(X), let X = fn:count()."),
			}},
		{
			descr:           "cannot reference Y, not grouped",
			knownPredicates: nil,
			decls:           nil,
			program: []ast.Clause{
				clause("ok(/foo, /bar)."),
				clause("sna(X, Y, Z) :- ok(X, Y) |> do fn:group_by(X), let Z = fn:count()."),
			}},
		{
			descr:           "cannot call reducer in let-transform",
			knownPredicates: nil,
			decls:           nil,
			program: []ast.Clause{
				clause("ok(/foo, /bar)."),
				clause("sna(X, Y, Z) :- ok(X, Y) |> let X = fn:sum(Y), let Z = fn:count()."),
			},
		},
		{
			descr:           "arity mismatch with decl",
			knownPredicates: nil,
			decls:           []ast.Decl{makeSyntheticDecl(t, atom("foo(X,Y)"))},
			program: []ast.Clause{
				clause("foo(1)."),
			},
		},
		{
			descr:           "arity mismatch with known predicate",
			knownPredicates: map[ast.PredicateSym]ast.Decl{ast.PredicateSym{"foo", 2}: fooDecl},
			decls:           nil,
			program: []ast.Clause{
				clause("foo(1)."),
			},
		},
		{
			descr: "synthetic cannot override non-synthetic",
			knownPredicates: map[ast.PredicateSym]ast.Decl{
				ast.PredicateSym{"foo", 2}: makeNonSynthetic(t, fooDecl),
			},
			decls:   []ast.Decl{fooDecl},
			program: nil,
		},
		{
			descr: "arity mismatch between decl and known predicate",
			knownPredicates: map[ast.PredicateSym]ast.Decl{
				ast.PredicateSym{"foo", 2}: makeNonSynthetic(t, fooDecl),
			},
			decls:   []ast.Decl{fooThreeArgs},
			program: nil,
		},
		{
			descr: "arity mismatch between decl and known predicate",
			knownPredicates: map[ast.PredicateSym]ast.Decl{
				ast.PredicateSym{"foo", 3}: fooThreeArgs,
			},
			decls:   []ast.Decl{fooDecl},
			program: nil,
		},
	}
	for _, test := range tests {
		if _, err := AnalyzeOneUnit(parse.SourceUnit{Clauses: test.program, Decls: test.decls}, test.knownPredicates); err == nil { // if no error
			t.Errorf("%q: expected error for invalid program %v", test.descr, test.program)
		}
	}
}

func makeRulesMap(clauses []ast.Clause) map[ast.PredicateSym][]ast.Clause {
	res := make(map[ast.PredicateSym][]ast.Clause)
	for _, clause := range clauses {
		res[clause.Head.Predicate] = append(res[clause.Head.Predicate], clause)
	}
	return res
}

type boundsTestCase struct {
	programInfo ProgramInfo
	rulesMap    map[ast.PredicateSym][]ast.Clause
	nameTrie    symbols.NameTrie
}

func makeDesugaredDecls(t *testing.T, decls ...ast.Decl) map[ast.PredicateSym]*ast.Decl {
	t.Helper()
	declMap := make(map[ast.PredicateSym]ast.Decl)
	for _, decl := range decls {
		declMap[decl.DeclaredAtom.Predicate] = decl
	}
	desugaredDecls, err := symbols.CheckAndDesugar(declMap)
	if err != nil {
		t.Fatal(err)
	}
	return desugaredDecls
}

func newBoundsTestCase(t *testing.T, clauses []ast.Clause, decls []ast.Decl) boundsTestCase {
	return newBoundsTestCaseWithNameTrie(t, clauses, decls, nil)
}

func newBoundsTestCaseWithNameTrie(t *testing.T, clauses []ast.Clause, decls []ast.Decl, nameTrie symbols.NameTrie) boundsTestCase {
	t.Helper()
	idbSymbols := make(map[ast.PredicateSym]struct{})
	for _, clause := range clauses {
		idbSymbols[clause.Head.Predicate] = struct{}{}
	}

	return boundsTestCase{
		programInfo: ProgramInfo{nil, idbSymbols, nil, clauses, makeDesugaredDecls(t, decls...)},
		rulesMap:    makeRulesMap(clauses),
		nameTrie:    nameTrie,
	}
}

func makeSimpleDecl(a ast.Atom, bound ...ast.BaseTerm) ast.Decl {
	return ast.Decl{
		DeclaredAtom: a,
		Bounds:       []ast.BoundDecl{{bound}},
	}
}

func TestBoundsAnalyzer(t *testing.T) {
	tests := []boundsTestCase{
		newBoundsTestCase(t, []ast.Clause{
			clause("foo(X) :- bar(X), X = 3."),
		}, []ast.Decl{
			makeSimpleDecl(atom("foo(X)"), ast.NumberBound),
			makeSimpleDecl(atom("bar(X)"), ast.NumberBound),
		}),
		newBoundsTestCase(t, []ast.Clause{
			clause("foo(X, Y, Z) :- bar(X, Y), baz(Y, Z)."),
		}, []ast.Decl{
			makeSimpleDecl(atom("foo(X, Y, Z)"), ast.StringBound, ast.NumberBound, ast.NumberBound),
			makeSimpleDecl(atom("bar(A, B)"), ast.StringBound, ast.NumberBound),
			makeSimpleDecl(atom("baz(E, F)"), ast.NumberBound, ast.NumberBound),
		}),
		newBoundsTestCase(t, []ast.Clause{
			clause("foo(X, X) :- bar(X)."),
		}, []ast.Decl{
			makeSimpleDecl(atom("foo(X, Y)"), ast.NumberBound, ast.NumberBound),
			makeSimpleDecl(atom("bar(A)"), ast.NumberBound),
		}),
		newBoundsTestCase(t, []ast.Clause{
			clause("foo(X) :- bar(X)."),
			clause("foo(X) :- baz(X)."),
		}, []ast.Decl{
			makeSimpleDecl(atom("foo(X)"), ast.ApplyFn{symbols.UnionType, []ast.BaseTerm{ast.StringBound, ast.NumberBound}}),
			makeSimpleDecl(atom("bar(A)"), ast.StringBound),
			makeSimpleDecl(atom("baz(A)"), ast.NumberBound),
		}),
		newBoundsTestCase(t, []ast.Clause{
			clause("foo(X) :- bar(X)."),
			clause("foo(X) :- baz(X)."),
		}, []ast.Decl{
			makeSimpleDecl(atom("foo(X)"), ast.NumberBound),
			makeSimpleDecl(atom("bar(A)"), ast.NumberBound),
			makeSimpleDecl(atom("baz(A)"), ast.NumberBound),
		}),
		newBoundsTestCase(t, []ast.Clause{
			clause("foo(1)."),
			clause("foo(X) :- bar(X)."),
			clause("bar(X) :- foo(X)."),
		}, []ast.Decl{
			makeSimpleDecl(atom("foo(X)"), ast.NumberBound),
			makeSimpleDecl(atom("bar(A)"), ast.NumberBound),
		}),
		newBoundsTestCase(t, []ast.Clause{
			clause("foo(1)."),
			clause("foo(['a'])."),
			clause("foo(X) :- bar(X)."),
			clause("bar(X) :- foo(X)."),
		}, []ast.Decl{
			ast.Decl{
				DeclaredAtom: atom("foo(X)"),
				Bounds: []ast.BoundDecl{
					ast.NewBoundDecl(ast.NumberBound),
					ast.NewBoundDecl(symbols.NewListType(ast.StringBound)),
				},
			},
			ast.Decl{
				DeclaredAtom: atom("bar(X)"),
				Bounds: []ast.BoundDecl{
					ast.NewBoundDecl(ast.NumberBound),
					ast.NewBoundDecl(symbols.NewListType(ast.StringBound)),
				},
			},
		}),
		newBoundsTestCase(t, []ast.Clause{
			clause("foo(X) :- bar(X), :match_prefix(X, /foo)."),
		}, []ast.Decl{
			makeSimpleDecl(atom("foo(X)"), ast.NameBound),
			makeSimpleDecl(atom("bar(X)"), name("/foo")),
		}),
		newBoundsTestCase(t, []ast.Clause{
			clause("foo(X) :- bar(Y), :match_entry(Y, 'a', X)."),
		}, []ast.Decl{
			makeSimpleDecl(atom("foo(X)"), ast.NumberBound),
			makeSimpleDecl(atom("bar(Y)"), symbols.NewMapType(ast.StringBound, ast.NumberBound)),
		}),
		newBoundsTestCase(t, []ast.Clause{
			clause("foo(X) :- bar(Y), :match_field(Y, /value, X)."),
		}, []ast.Decl{
			makeSimpleDecl(atom("foo(X)"), ast.NumberBound),
			makeSimpleDecl(atom("bar(Y)"), symbols.NewStructType(name("/value"), ast.NumberBound)),
		}),
		newBoundsTestCase(t, []ast.Clause{
			clause("foo(X) :- bar(Y), :list:member(X, Y)."),
		}, []ast.Decl{
			makeSimpleDecl(atom("foo(X)"), ast.NumberBound),
			makeSimpleDecl(atom("bar(Y)"), symbols.NewListType(ast.NumberBound)),
		}),
		newBoundsTestCase(t, []ast.Clause{
			clause("bar(T) :-	T = fn:string:concat(\"A\", 123)."),
		}, []ast.Decl{
			makeSimpleDecl(atom("bar(T)"), ast.StringBound),
		}),
		newBoundsTestCase(t, []ast.Clause{
			clause("bar(T) :-	T = fn:list:get(['a'], 0)."),
		}, []ast.Decl{
			makeSimpleDecl(atom("bar(T)"), symbols.NewOptionType(ast.StringBound)),
		}),
		newBoundsTestCase(t, []ast.Clause{
			clause("bar(fn:pair('a', 0))."),
		}, []ast.Decl{
			makeSimpleDecl(atom("bar(T)"), symbols.NewPairType(ast.StringBound, ast.NumberBound)),
		}),
		newBoundsTestCase(t, []ast.Clause{
			clause("bar(fn:tuple('a', 0, /foo))."),
		}, []ast.Decl{
			makeSimpleDecl(atom("bar(T)"), symbols.NewTupleType(ast.StringBound, ast.NumberBound, ast.NameBound)),
		}),
		newBoundsTestCase(t, []ast.Clause{
			clause("bar(X) :- X = fn:tuple('a', 0, /foo)."),
		}, []ast.Decl{
			makeSimpleDecl(atom("bar(T)"), symbols.NewTupleType(ast.StringBound, ast.NumberBound, ast.NameBound)),
		}),
		newBoundsTestCase(t, []ast.Clause{
			clause("bar(X) :- :match_field({ /foo: 'a' }, /foo, X)."),
		}, []ast.Decl{
			makeSimpleDecl(atom("bar(S)"), ast.StringBound),
		}),
		newBoundsTestCase(t, []ast.Clause{
			clause("bar(X) :- X = fn:struct:get({ /foo: 'a' }, /foo)."),
		}, []ast.Decl{
			makeSimpleDecl(atom("bar(S)"), ast.StringBound),
		}),
		newBoundsTestCase(t, []ast.Clause{
			clause("foo(X) :- bar(X), :match_prefix(X, /baz)."),
		}, []ast.Decl{
			makeSimpleDecl(atom("foo(X)"), name("/baz")),
			makeSimpleDecl(atom("bar(X)"), symbols.NewUnionType(name("/bar"), name("/baz"))),
		}),
		newBoundsTestCase(t, []ast.Clause{
			clause("foo(X) :- boo(X), !:match_prefix(X, /bazzz)."),
		}, []ast.Decl{
			makeSimpleDecl(atom("foo(X)"), name("/bar")),
			makeSimpleDecl(atom("boo(X)"), symbols.NewUnionType(name("/bar"), name("/bazzz"))),
		}),
		newBoundsTestCase(t, []ast.Clause{
			clause(`foo(X) :- bar(X), :string:starts_with(X, "baz").`),
		}, []ast.Decl{
			makeSimpleDecl(atom("foo(X)"), ast.StringBound),
			makeSimpleDecl(atom("bar(X)"), ast.StringBound),
		}),
		newBoundsTestCase(t, []ast.Clause{
			clause(`foo(X) :- bar(X), :string:ends_with(X, "baz").`),
		}, []ast.Decl{
			makeSimpleDecl(atom("foo(X)"), ast.StringBound),
			makeSimpleDecl(atom("bar(X)"), ast.StringBound),
		}),
		newBoundsTestCase(t, []ast.Clause{
			clause(`foo(X) :- bar(X), :string:contains(X, "baz").`),
		}, []ast.Decl{
			makeSimpleDecl(atom("foo(X)"), ast.StringBound),
			makeSimpleDecl(atom("bar(X)"), ast.StringBound),
		}),
	}
	for _, test := range tests {
		bc, err := newBoundsAnalyzer(&test.programInfo, symbols.NewNameTrie(), nil, test.rulesMap)
		if err != nil {
			t.Fatal(err)
		}
		if err := bc.BoundsCheck(); err != nil {
			t.Errorf("BoundsCheck(%v, %v) returns error when it shouldn't: %v", test.programInfo, test.rulesMap, err)
		}
	}
}

func TestBoundsAnalyzerNegative(t *testing.T) {
	tests := []boundsTestCase{
		newBoundsTestCase(t, []ast.Clause{
			clause("foo(X, Y, Z) :- bar(X, Y), baz(Y, Z)."),
		}, []ast.Decl{
			makeSimpleDecl(atom("foo(X, Y, Z)"), ast.StringBound, ast.StringBound, ast.NumberBound),
			makeSimpleDecl(atom("bar(A, B)"), ast.StringBound, ast.NumberBound),
			makeSimpleDecl(atom("baz(E, F)"), ast.NumberBound, ast.NumberBound),
		}),
		newBoundsTestCase(t, []ast.Clause{
			clause("foo('hello')."),
		}, []ast.Decl{
			makeSimpleDecl(atom("foo(Num)"), ast.NumberBound),
		}),
		newBoundsTestCase(t, []ast.Clause{
			clause("foo(X) :- bar(X), X = 'hello'."),
		}, []ast.Decl{
			makeSimpleDecl(atom("foo(X)"), ast.NumberBound),
			makeSimpleDecl(atom("bar(X)"), ast.NumberBound),
		}),
		newBoundsTestCase(t, []ast.Clause{
			clause("foo(X) :- bar(X), :lt(X, 10)."),
		}, []ast.Decl{
			makeSimpleDecl(atom("foo(X)"), ast.StringBound),
			makeSimpleDecl(atom("bar(X)"), ast.StringBound),
		}),
		newBoundsTestCase(t, []ast.Clause{
			clause("foo(X, Y) :- bar(X, Y)."),
		}, []ast.Decl{
			makeSimpleDecl(atom("foo(X, Y)"), ast.StringBound, ast.NumberBound),
			makeSimpleDecl(atom("bar(A, B)"), ast.AnyBound, ast.NumberBound),
		}),
		newBoundsTestCase(t, []ast.Clause{
			clause("foo(X, Y) :- bar(X, Y), bar(X, Z), bar(Z, Y)."),
		}, []ast.Decl{
			makeSimpleDecl(atom("foo(X, Y)"), ast.StringBound, ast.NumberBound),
			makeSimpleDecl(atom("bar(A, B)"), ast.StringBound, ast.NumberBound),
		}),
		newBoundsTestCase(t, []ast.Clause{
			clause("foo(X) :- bar(X)."),
			clause("foo(X) :- baz(X)."),
		}, []ast.Decl{
			makeSimpleDecl(atom("foo(X)"), ast.NumberBound),
			makeSimpleDecl(atom("bar(A)"), ast.StringBound),
			makeSimpleDecl(atom("baz(A)"), ast.NumberBound),
		}),
		newBoundsTestCase(t, []ast.Clause{
			clause("foo(X) :- bar(X, Y), baz(['hello'])."),
		}, []ast.Decl{
			makeSimpleDecl(atom("foo(X)"), ast.NumberBound),
			makeSimpleDecl(atom("bar(A, B)"), ast.StringBound, ast.StringBound),
			makeSimpleDecl(atom("baz(A)"), symbols.NewListType(ast.NumberBound)),
		}),
		newBoundsTestCase(t, []ast.Clause{
			clause("bar(X) :- :match_field({ /foo: 'a' }, /baz, X)."),
		}, []ast.Decl{
			makeSimpleDecl(atom("bar(S)"), ast.StringBound),
		}),
		newBoundsTestCase(t, []ast.Clause{
			clause("bar(X) :- X = fn:struct:get({ /foo: 2 }, /foo)."),
		}, []ast.Decl{
			makeSimpleDecl(atom("bar(S)"), ast.StringBound),
		}),
		newBoundsTestCase(t, []ast.Clause{
			clause("bar(X) :- X = fn:struct:get({ /foo: 2 }, /baz)."),
		}, []ast.Decl{
			makeSimpleDecl(atom("bar(S)"), ast.StringBound),
		}),
		newBoundsTestCase(t, []ast.Clause{
			clause("bar(X) :- :match_field({ /foo: 2 }, /foo, X)."),
		}, []ast.Decl{
			makeSimpleDecl(atom("bar(S)"), ast.StringBound),
		}),
		newBoundsTestCase(t, []ast.Clause{
			clause("foo(X) :- bar(X), :match_prefix(X, /foo)."),
		}, []ast.Decl{
			makeSimpleDecl(atom("foo(X)"), ast.NameBound),
			makeSimpleDecl(atom("bar(X)"), name("/bar")),
		}),
		newBoundsTestCase(t, []ast.Clause{
			clause("foo(X) :- bar(X), :match_prefix(X, /baz)."),
		}, []ast.Decl{
			makeSimpleDecl(atom("foo(X)"), name("/bar")),
			makeSimpleDecl(atom("bar(X)"), symbols.NewUnionType(name("/bar"), name("/baz"))),
		}),
		newBoundsTestCase(t, []ast.Clause{
			clause("bar(T) :-	T = fn:list:get(['a'], 0)."),
		}, []ast.Decl{
			makeSimpleDecl(atom("bar(T)"), symbols.NewOptionType(ast.NumberBound)),
		}),
	}
	for _, test := range tests {
		bc, err := newBoundsAnalyzer(&test.programInfo, symbols.NewNameTrie(), nil, test.rulesMap)
		if err != nil {
			t.Fatal(err)
		}
		if err := bc.BoundsCheck(); err == nil { // if NO error
			t.Errorf("BoundsCheck should have returned error but did not: %v", test.rulesMap)
		}
	}
}

func TestCollectNames(t *testing.T) {
	extraPredicates := map[ast.PredicateSym]ast.Decl{
		ast.PredicateSym{"z", 1}: makeSimpleDecl(atom("z(X)"), name("/foo/bar")),
	}
	tests := []struct {
		decls map[ast.PredicateSym]*ast.Decl
		want  symbols.NameTrie
	}{
		{
			decls: makeDesugaredDecls(t,
				makeSimpleDecl(atom("a(X)"), name("/foo")),
				makeSimpleDecl(atom("b(X)"), name("/foo")),
				makeSimpleDecl(atom("c(X)"), name("/foo/bar")),
			),
			want: symbols.NewNameTrie().Add([]string{"foo"}).Add([]string{"foo", "bar"}),
		},
		{
			decls: makeDesugaredDecls(t,
				makeSimpleDecl(atom("b(X)"), name("/foo/baz")),
			),
			want: symbols.NewNameTrie().Add([]string{"foo", "bar"}).Add([]string{"foo", "baz"}),
		},
	}
	for _, test := range tests {
		got := collectNames(extraPredicates, test.decls)
		if !cmp.Equal(got, test.want, cmp.AllowUnexported(symbols.NameTrieNode{})) {
			t.Errorf("collectNames(%v) = %v, want %v", test.decls, got, test.want)
		}
	}
}

func TestBoundsAnalyzerWithNames(t *testing.T) {
	test := newBoundsTestCaseWithNameTrie(t, []ast.Clause{
		clause("a(X) :- b(X)."),
	}, []ast.Decl{
		makeSimpleDecl(atom("a(X)"), name("/foo")),
		makeSimpleDecl(atom("b(X)"), name("/foo")),
	},
		symbols.NewNameTrie().Add([]string{"foo"}),
	)
	bc, err := newBoundsAnalyzer(&test.programInfo, test.nameTrie, []ast.Atom{atom("b(/foo/bar)")}, test.rulesMap)
	if err != nil {
		t.Fatal(err)
	}
	if err := bc.BoundsCheck(); err != nil {
		t.Errorf("BoundsCheck(%v, %v) returns error when it shouldn't: %v", test.programInfo, test.rulesMap, err)
	}
}

func TestBoundsAnalyzerWithManyInitialFacts(t *testing.T) {
	test := newBoundsTestCaseWithNameTrie(t, []ast.Clause{
		clause("a(X) :- b(X)."),
	}, []ast.Decl{
		makeSimpleDecl(atom("a(X)"), name("/foo")),
		makeSimpleDecl(atom("b(X)"), name("/foo")),
	},
		symbols.NewNameTrie().Add([]string{"foo"}),
	)
	var facts []ast.Atom
	for i := 0; i < 100000; i++ {
		facts = append(facts, atom(fmt.Sprintf("b(/foo/bar%d)", i)))
	}
	bc, err := newBoundsAnalyzer(&test.programInfo, test.nameTrie, facts, test.rulesMap)
	if err != nil {
		t.Fatal(err)
	}
	if err := bc.BoundsCheck(); err != nil {
		t.Errorf("BoundsCheck(%v, %v) returns error when it shouldn't: %v", test.programInfo, test.rulesMap, err)
	}
}

func TestBoundsAnalyzerWithNamesNegative(t *testing.T) {
	test := newBoundsTestCaseWithNameTrie(t, []ast.Clause{
		clause("a(X) :- b(X)."),
	}, []ast.Decl{
		makeSimpleDecl(atom("a(X)"), name("/foo")),
	},
		symbols.NewNameTrie().Add([]string{"foo"}),
	)
	bc, err := newBoundsAnalyzer(&test.programInfo, test.nameTrie, nil, test.rulesMap)
	if err != nil {
		t.Fatal(err)
	}
	if err := bc.BoundsCheck(); err == nil {
		t.Errorf("BoundsCheck(%v, %v) should have returned error but did not", test.programInfo, test.rulesMap)
	}
}
