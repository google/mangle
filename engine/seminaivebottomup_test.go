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
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/mangle/analysis"
	"github.com/google/mangle/ast"
	"github.com/google/mangle/factstore"
	"github.com/google/mangle/functional"
	"github.com/google/mangle/parse"
	"github.com/google/mangle/unionfind"
)

func atom(s string) ast.Atom {
	term, err := parse.Term(s)
	if err != nil {
		panic(err)
	}
	return term.(ast.Atom)
}

func evalAtom(t *testing.T, str string) ast.Atom {
	term, err := parse.Term(str)
	if err != nil {
		t.Fatalf("error parsing atom from %q: %v", str, err)
	}
	eval, err := functional.EvalAtom(term.(ast.Atom), nil)
	if err != nil {
		t.Fatalf("error eval atom from %v: %v", term, err)
	}
	return eval
}

func clause(str string) ast.Clause {
	clause, err := parse.Clause(str)
	if err != nil {
		panic(fmt.Errorf("bad syntax in test case: %s got %w", str, err))
	}
	return clause
}

func unit(str string) parse.SourceUnit {
	unit, err := parse.Unit(strings.NewReader(str))
	if err != nil {
		panic(fmt.Errorf("bad syntax in test case: %s got %w", str, err))
	}
	return unit
}

var program []ast.Clause

func init() {
	program = []ast.Clause{
		clause("path(X,Y) :- edge(X,Y)."),
		clause("path(X,Z) :- edge(X,Y), path(Y,Z)."),
		clause("not_reachable(X, Y) :- node(X), node(Y), !path(X, Y)."),
		clause("in_cycle_eq(X) :- node(X), path(X, Y), X = Y."),
		clause("in_between(X, Z) :- node(X), node(Y), node(Z), path(X, Y), path(Y, Z), X != Y, Y != Z, X != Z."),
		clause("neighbor_label(X, Y, Num) :- edge(X, Y), label(Y, Num)."),
		clause("has_neighbor(X) :- edge(X, _)."),
		clause("decompose_pair(Y,Z) :- :match_pair(fn:pair(1,2),Y,Z)."),
		clause("decompose_cons(Y,Z) :- :match_cons(fn:list:cons(1,[]),Y,Z)."),
		clause("decompose_nil() :- :match_nil([])."),
		clause("expanded_stuff(X) :- :list:member(X, [1, 2, 3])."),
	}
}

func asMap(preds []ast.PredicateSym) map[ast.PredicateSym]ast.Decl {
	m := make(map[ast.PredicateSym]ast.Decl, len(preds))
	for _, sym := range preds {
		m[sym] = ast.NewSyntheticDeclFromSym(sym)
	}
	return m
}

// analyzeAndEvalProgram analyzes and evaluates a given program on the given facts, modifying the
// fact store in the process.
// Analysis considers only predicates that are found in the store.
func analyzeAndEvalProgram(t *testing.T, clauses []ast.Clause, store factstore.SimpleInMemoryStore, options ...EvalOption) error {
	t.Helper()
	programInfo, err := analysis.AnalyzeOneUnit(parse.SourceUnit{Clauses: clauses}, asMap(store.ListPredicates()))
	if err != nil {
		return fmt.Errorf("analysis: %w", err)
	}
	for _, f := range programInfo.InitialFacts {
		store.Add(f)
	}
	return EvalProgram(programInfo, store, options...)
}

func TestSingleDeltaRule(t *testing.T) {
	rule := clause("path(X,Z,U) :- edge(X,Y), path(Y,W,Z) |> let U = fn:plus(Z, 1).")
	res := makeSingleDeltaRule(rule, 1)
	expectedNewPremises := []ast.Term{atom("edge(X, Y)"), ast.NewAtom("Δpath", ast.Variable{"Y"}, ast.Variable{"W"}, ast.Variable{"Z"})}
	if cmp.Diff(expectedNewPremises, res.Premises) != "" {
		t.Errorf("makeSingleDeltaRule want %v got %v", expectedNewPremises, res)
	}
	if rule.Transform != res.Transform {
		t.Errorf("makeSingleDeltaRule missing transform %v %v", rule.Transform, res.Transform)
	}
}

func TestRewriteIdempotency(t *testing.T) {
	store := factstore.NewSimpleInMemoryStore()
	store.Add(atom(`attr("node", "name", 1, "description")`))
	program := []ast.Clause{
		clause(`
	attrs(Node, Attrs) :- attr(Node, Name, Value, Description),
		Attr = fn:tuple(Name, Value, Description)
		|> do fn:group_by(Node), let Attrs=fn:collect(Attr).
	`),
	}

	programInfo, err := analysis.AnalyzeOneUnit(parse.SourceUnit{Clauses: program}, asMap(store.ListPredicates()))
	if err != nil {
		t.Fatalf("analysis: %v", err)
	}
	for _, f := range programInfo.InitialFacts {
		store.Add(f)
	}

	if err := EvalProgram(programInfo, store); err != nil {
		t.Fatalf("Program evaluation failed %v program %v", err, program)
	}
	facts := store.EstimateFactCount()

	for i := 0; i < 10; i++ {
		if err := EvalProgram(programInfo, store); err != nil {
			t.Fatalf("Program evaluation failed %v program %v", err, program)
		}
	}

	if store.EstimateFactCount() != facts {
		t.Fatalf("Unexpected number of facts %v, expected: %v.", store.EstimateFactCount(), facts+1)
	}
}

func TestNegation(t *testing.T) {
	store := factstore.NewSimpleInMemoryStore()
	negationProgram := []ast.Clause{
		clause("foo(/a)."),
		clause("foo(/b)."),
		clause("foo(/c)."),
		clause("bar(/a)."),
		clause("bar(/b)."),
		clause("notbar(X) :- foo(X), !bar(X)."),
	}
	if err := analyzeAndEvalProgram(t, negationProgram, store); err != nil {
		t.Errorf("Program evaluation failed %v program %v", err, program)
		return
	}
	expected := []ast.Atom{
		atom("notbar(/c)"),
	}
	for _, fact := range expected {
		if !store.Contains(fact) {
			t.Errorf("expected fact %v in store %v", fact, store)
		}
	}
	notExpected := []ast.Atom{
		atom("notbar(/a)"),
		atom("notbar(/b)"),
	}
	for _, fact := range notExpected {
		if store.Contains(fact) {
			t.Errorf("did not expect fact %v in store %v", fact, store)
		}
	}
}

func rulesByPredicate(rules []ast.Clause) map[ast.PredicateSym][]ast.Clause {
	predToRules := make(map[ast.PredicateSym][]ast.Clause)
	for _, rule := range rules {
		pred := rule.Head.Predicate
		predToRules[pred] = append(predToRules[pred], rule)
	}
	return predToRules
}

func TestMakeDeltaRules(t *testing.T) {
	pathProgram := []ast.Clause{
		clause("path(X,Y) :- edge(X,Y)."),
		clause("path(X,Z) :- edge(X,Y), path(Y,Z)."),
	}
	pathSym := ast.PredicateSym{"path", 2}
	decl := ast.NewSyntheticDeclFromSym(pathSym)
	decls := map[ast.PredicateSym]*ast.Decl{
		pathSym: &decl,
	}
	predToRules := rulesByPredicate(pathProgram)
	predToDeltaRules := makeDeltaRules(decls, predToRules)
	if len(predToDeltaRules) != 1 {
		t.Fatalf("expected one entry, got %v", predToDeltaRules)
	}
	deltaRules := predToDeltaRules[ast.PredicateSym{"path", 2}]
	want := []ast.Clause{
		ast.NewClause(atom("path(X,Z)"), []ast.Term{atom("edge(X, Y)"), ast.NewAtom("Δpath", ast.Variable{"Y"}, ast.Variable{"Z"})}),
	}
	if diff := cmp.Diff(want, deltaRules); diff != "" {
		t.Fatalf("wanted %v diff %v", want, diff)
	}
}

func TestSimpleEval(t *testing.T) {
	store := factstore.NewSimpleInMemoryStore()
	store.Add(atom("node(/a)"))
	store.Add(atom("node(/b)"))
	store.Add(atom("label(/b, 100)"))
	store.Add(atom("node(/c)"))
	store.Add(atom("node(/d)"))
	store.Add(atom("edge(/a,/b)"))
	store.Add(atom("edge(/b,/c)"))
	store.Add(atom("edge(/c,/d)"))

	if err := analyzeAndEvalProgram(t, program, store); err != nil {
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
		atom("neighbor_label(/a, /b, 100)"),
		atom("has_neighbor(/c)"),
		atom("expanded_stuff(1)"),
		atom("expanded_stuff(2)"),
		atom("expanded_stuff(3)"),
	}

	for _, fact := range expected {
		if !store.Contains(fact) {
			t.Errorf("expected fact %v in store %v", fact, store)
		}
	}
}

func TestSimpleEvalDefer(t *testing.T) {
	store := factstore.NewSimpleInMemoryStore()
	store.Add(atom("node(/a)"))
	store.Add(atom("node(/b)"))
	store.Add(atom("node(/c)"))
	store.Add(atom("edge(/a,/b)"))
	store.Add(atom("edge(/b,/c)"))
	store.Add(atom("label(/a, 100)"))

	if err := analyzeAndEvalProgram(t, program, store, func(opt *EvalOptions) {
		// Evaluate only this particular predicate and nothing else.
		predicateAllowList := func(sym ast.PredicateSym) bool {
			return sym == ast.PredicateSym{"path", 2}
		}
		opt.predicateAllowList = &predicateAllowList
	}); err != nil {
		t.Errorf("Program evaluation failed %v program %v", err, program)
		return
	}

	expected := []ast.Atom{
		atom("node(/a)"),
		atom("node(/b)"),
		atom("node(/c)"),
		atom("edge(/a,/b)"),
		atom("edge(/b,/c)"),
		atom("label(/a, 100)"),
		atom("path(/a,/b)"),
		atom("path(/a,/c)"),
		atom("path(/b,/c)"),
	}

	for _, fact := range expected {
		if !store.Contains(fact) {
			t.Errorf("expected fact %v in store %v", fact, store)
		}
	}
	if store.EstimateFactCount() > len(expected) {
		t.Errorf("extra facts: %v", store)
	}
}

func TestManyPaths(t *testing.T) {
	store := factstore.NewSimpleInMemoryStore()
	store.Add(atom("node(/a)"))
	store.Add(atom("label(/a, 100)"))

	for i := 1; i <= 10; i++ {
		store.Add(atom(fmt.Sprintf("node(/b%d)", i)))
		store.Add(atom(fmt.Sprintf("node(/c%d)", i)))
		store.Add(atom(fmt.Sprintf("edge(/a,/b%d)", i)))
		store.Add(atom(fmt.Sprintf("edge(/b%d, /c%d)", i, i)))
	}
	store.Add(atom("edge(/c9,/b9)"))
	if err := analyzeAndEvalProgram(t, program, store); err != nil {
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

func TestNonLinear(t *testing.T) {
	store := factstore.NewSimpleInMemoryStore()
	store.Add(atom("par(1, 2)"))
	store.Add(atom("par(2, 3)"))

	// Non-linear rule.
	program := []ast.Clause{
		clause("anc(X, Y) :- par(X, Y)."),
		clause("anc(X, Y) :- anc(X, Z), anc(Z, Y)."),
	}

	if err := analyzeAndEvalProgram(t, program, store); err != nil {
		t.Errorf("Program evaluation failed %v program %v", err, program)
		return
	}
	expected := []ast.Atom{
		atom("anc(1, 2)"),
		atom("anc(2, 3)"),
		atom("anc(1, 3)"),
	}
	for _, fact := range expected {
		if !store.Contains(fact) {
			t.Errorf("expected fact %v in store %v", fact, store)
		}
	}
}

func TestBuiltin(t *testing.T) {
	store := factstore.NewSimpleInMemoryStore()
	store.Add(ast.NewAtom("foo", ast.Number(1)))
	store.Add(ast.NewAtom("foo", ast.Number(2)))
	store.Add(ast.NewAtom("foo", ast.Number(11)))
	program := []ast.Clause{
		clause("lt_two(X) :- foo(X), X < 2."),
		clause("le_two(X) :- foo(X), X <= 2."),
		clause("two_le(X) :- foo(X), 2 <= X."),
		clause("gt_two(X) :- foo(X), X > 2."),
		clause("ge_two(X) :- foo(X), X >= 2."),
		clause("two_ge(X) :- foo(X), 2 >= X."),
		clause("within_ten(X) :- foo(X), :within_distance(10, X, 2)."),
	}
	if err := analyzeAndEvalProgram(t, program, store); err != nil {
		t.Errorf("Program evaluation failed %v program %v", err, program)
		return
	}
	expected := []ast.Atom{
		atom("lt_two(1)"),
		atom("le_two(1)"),
		atom("le_two(2)"),
		atom("two_le(2)"),
		atom("gt_two(11)"),
		atom("ge_two(11)"),
		atom("ge_two(2)"),
		atom("two_ge(2)"),
		atom("within_ten(11)"),
	}

	for _, fact := range expected {
		if !store.Contains(fact) {
			t.Errorf("expected fact %v in store %v", fact, store)
		}
	}
}

func TestListMapStruct(t *testing.T) {
	store := factstore.NewSimpleInMemoryStore()
	program := []ast.Clause{
		// List
		clause(`bar(1).`),
		clause(`foo(X) :- bar(Y), X = [Y].`),
		clause(`baz(Y) :- X = [0,1,2], Y = fn:list:get(X, 1).`),
		// Map
		clause(`bar_map([ "foo" : /foo, "bar": /bar ]).`),
		clause(`bar_map([ "foo" : /hoo, "bar": /zar ]).`),
		clause(`bar_entry(Y, Z) :- bar_map(X), :match_entry(X, "foo", Y), :match_entry(X, "bar", Z).`),
		// Struct
		clause(`bar_struct({ /foo : 1, /bar : "barbar"}).`),
		clause(`bar_extracted(Y, Z) :- bar_struct(X), :match_field(X, /foo, Y), :match_field(X, /bar, Z).`),
	}

	if err := analyzeAndEvalProgram(t, program, store); err != nil {
		t.Errorf("Program evaluation failed %v program %v", err, program)
		return
	}

	expected := []ast.Atom{
		ast.NewAtom("foo", ast.List([]ast.Constant{ast.Number(1)})),
		atom("baz(1)"),
		atom("bar_entry(/foo, /bar)"),
		atom("bar_entry(/hoo, /zar)"),
		ast.NewAtom("bar_extracted", ast.Number(1), ast.String("barbar")),
	}

	for _, fact := range expected {
		if !store.Contains(fact) {
			t.Errorf("expected fact %v in store %v", fact, store)
		}
	}
}

func TestTransform(t *testing.T) {
	store := factstore.NewSimpleInMemoryStore()
	facts := []ast.Atom{
		atom("node(/a)"),
		atom("node(/b)"),
		atom("node(/c)"),
		atom("node(/d)"),
		atom("edge(/a,/b)"),
		atom("edge(/b,/c)"),
		atom("edge(/c,/d)"),
		atom("label(/a, 100)"),
		atom("label(/b, 20)"),
		atom("label(/c, 50)"),
		atom("label(/d, 500)"),
		atom("decompose_pair(1, 2)"),
		ast.NewAtom("decompose_cons", ast.Number(1), ast.ListNil),
		atom("decompose_nil()"),
	}
	for _, fact := range facts {
		store.Add(fact)
	}
	program := []ast.Clause{
		clause(`max_inner(Max, fn:pair(1,2)) :-
				      node(Y), edge(X, Y), edge(Y, _), label(Y, N)
              |> do fn:group_by(), let Max = fn:max(N).`),
	}
	if err := analyzeAndEvalProgram(t, program, store); err != nil {
		t.Errorf("Program evaluation failed %v program %v", err, program)
		return
	}

	wantAtom, err := functional.EvalAtom(atom("max_inner(50, fn:pair(1,2))"), ast.ConstSubstList{})
	if err != nil {
		t.Fatal(err)
	}
	expected := []ast.Atom{
		wantAtom,
	}

	for _, fact := range expected {
		if !store.Contains(fact) {
			t.Errorf("expected fact %v in store %v", fact, store)
		}
	}
	for _, p := range store.ListPredicates() {
		store.GetFacts(ast.NewQuery(p), func(a ast.Atom) error {
			for _, arg := range a.Args {
				_, ok := arg.(ast.Constant)
				if !ok {
					t.Fatalf("found non constant %v %T in atom %v", arg, arg, a)
				}
			}
			return nil
		})
	}
}

func TestCreatedFactLimit(t *testing.T) {
	store := factstore.NewSimpleInMemoryStore()
	for i := 0; i < 1000; i++ {
		store.Add(atom(fmt.Sprintf("other_node(%d)", i)))
	}
	for i := 0; i < 10; i++ {
		store.Add(atom(fmt.Sprintf("node(%d)", i)))
	}

	program := []ast.Clause{
		clause(`node_square(X,Y) :- node(X), node(Y).`),
	}
	if err := analyzeAndEvalProgram(t, program, store, WithCreatedFactLimit(200)); err != nil {
		t.Errorf("Program evaluation failed %v program %v", err, program)
	}

	program = []ast.Clause{
		clause(`node_cube(X,Y,Z) :- node(X), node(Y), node(Z).`),
	}
	startCount := store.EstimateFactCount()
	if err := analyzeAndEvalProgram(t, program, store, WithCreatedFactLimit(500)); err == nil {
		t.Errorf("Program evaluation should have failed, but got %d facts, program %v", store.EstimateFactCount(), program)
	}
	if maxSize := startCount + 600; store.EstimateFactCount() > maxSize {
		t.Errorf("fact size limit is not effective enough got %d facts > %d", store.EstimateFactCount(), maxSize)
	}

	var multipleClauses []ast.Clause
	for i := 0; i < 100; i++ {
		multipleClauses = append(multipleClauses, clause(fmt.Sprintf(`node_square_%d(X,Y) :- node(X), node(Y).`, i)))
	}

	startCount = store.EstimateFactCount()
	if err := analyzeAndEvalProgram(t, multipleClauses, store, WithCreatedFactLimit(500)); err == nil {
		t.Errorf("Program evaluation should have failed, but got %d facts, program %v", store.EstimateFactCount(), multipleClauses)
	}
	if maxSize := startCount + 600; store.EstimateFactCount() > maxSize {
		t.Errorf("fact size limit is not effective enough got %d facts > %d", store.EstimateFactCount(), maxSize)
	}
}

func TestTransformPartsExplosion(t *testing.T) {
	store := factstore.NewSimpleInMemoryStore()
	program := []ast.Clause{
		// The 'component(Subpart, Part, Quantity)' relation expresses that
		//    <Quantity> units of <SubPart> go directly into one unit of <Part>.
		clause(`component(1, 5, 9).`),
		clause(`component(2, 5, 7).`),
		clause(`component(3, 5, 2).`),
		clause(`component(2, 6, 12).`),
		clause(`component(3, 6, 3).`),
		clause(`component(4, 7, 1).`),
		clause(`component(6, 7, 1).`),
		// The `transitive(DescPart, Part, Quantity, Path)` relation
		// expresses that <Quantity> units of <DescPart> go overall into one
		// unit of <Part> along a dependency path <Path>
		clause(`transitive(DescPart, Part, Quantity, []) :- component(DescPart, Part, Quantity).`),
		clause(`transitive(DescPart, Part, Quantity, Path) :-
		  component(SubPart, Part, DirectQuantity),
		  transitive(DescPart, SubPart, DescQuantity, SubPath)
      |> let Quantity = fn:mult(DirectQuantity, DescQuantity),
	       let Path = fn:list:cons(SubPart, SubPath).`),
		// The `full(DescPart, Part, Quantity) relation expresses that <Quantity>
		// units of <DescPart> go overall into one unit of <Part>.
		clause(`full(DescPart, Part, Sum) :-
		  transitive(DescPart, Part, Quantity, Path)
      |> do fn:group_by(DescPart, Part), let Sum = fn:sum(Quantity).`),
	}
	if err := analyzeAndEvalProgram(t, program, store); err != nil {
		t.Errorf("Program evaluation failed %v program %v", err, program)
		return
	}
	transitiveAtom, err := functional.EvalAtom(atom("transitive(2,7,12,[6])"), ast.ConstSubstList{})
	if err != nil {
		t.Fatal(err)
	}
	expected := []ast.Atom{
		transitiveAtom,
		atom("full(2,7,12)"),
	}

	for _, fact := range expected {
		if !store.Contains(fact) {
			t.Errorf("expected fact %v in store %v", fact, store)
		}
	}
}

func TestTransformFib(t *testing.T) {
	store := factstore.NewSimpleInMemoryStore()
	program := []ast.Clause{
		clause(`fib(0, 1).`),
		clause(`fib(1, 1).`),
		clause(`num(1).`),
		clause(`num(2).`),
		clause(`num(3).`),
		clause(`num(4).`),
		clause(`num(5).`),
		clause(`fib(I, O) :- num(I), fib(fn:minus(I, 1), M), fib(fn:minus(I, 2), N) |> let O = fn:plus(M, N).`),
	}
	if err := analyzeAndEvalProgram(t, program, store); err != nil {
		t.Errorf("Program evaluation failed %v program %v", err, program)
		return
	}
	expected := []ast.Atom{
		atom("fib(2,2)"),
		atom("fib(3,3)"),
		atom("fib(4,5)"),
		atom("fib(5,8)"),
	}

	for _, fact := range expected {
		if !store.Contains(fact) {
			t.Errorf("expected fact %v in store %v", fact, store)
		}
	}
}

func TestEmptyArrayProgram(t *testing.T) {
	store := factstore.NewSimpleInMemoryStore()
	program := []ast.Clause{
		clause(`a_list([]).`),
	}
	if err := analyzeAndEvalProgram(t, program, store); err != nil {
		t.Errorf("Program evaluation failed %v program %v", err, program)
		return
	}

	goal := atom("a_list(A)")
	facts := make(map[uint64]ast.Atom)
	if err := store.GetFacts(goal, func(storefact ast.Atom) error {
		subst, err := unionfind.UnifyTerms(goal.Args, storefact.Args)
		if err != nil {
			return nil
		}
		fact := goal.ApplySubst(subst).(ast.Atom)
		facts[fact.Hash()] = fact
		return nil
	}); err != nil {
		t.Errorf("GetFacts(%v): %v", goal, err)
	}
	if got, want := len(facts), 1; got != want {
		t.Errorf("GetFacts: %d!=%d got: %v", got, want, facts)
	}
}

func TestTransformMax(t *testing.T) {
	for _, tt := range []struct {
		program string
		want    ast.Atom
	}{
		{
			program: `bar([0,1,2]).
			  baz(Y) :- bar(X), |> let Y = fn:max(X).`,
			want: atom("baz(2)"),
		},
		{
			program: ` baz(Y) :- X = [0,1,2], Y = fn:max(X).`,
			want:    atom("baz(2)"),
		},
		{
			program: "baz(Y) :- X = [0,1,2], |> let Y = fn:max(X).",
			want:    atom("baz(2)"),
		},
	} {
		store := factstore.NewSimpleInMemoryStore()
		u := unit(tt.program)
		if err := analyzeAndEvalProgram(t, u.Clauses, store); err != nil {
			t.Errorf("analyzeAndEvalProgram(%s) = %v unexpected error", tt.program, err)
		}
		if !store.Contains(tt.want) {
			t.Errorf("expected fact %v in store %v", tt.want, store)
		}
	}
}

func TestTransformErrors(t *testing.T) {
	for _, tt := range []struct {
		program string
		wantErr string
	}{
		{
			program: "baz(Y) :- X = [0,1,2], Y = fn:list:get(1).",
			wantErr: "wrong arity",
		},
		{
			program: "baz(Y) :- X = [0,1,2], Y = fn:collect(1).",
			wantErr: "unknown function fn:collect",
		},
		{
			program: `num(1).
				num(2).
				fib(I, O) :- fib(fn:minus(I, 1), M), fib(fn:minus(I, 2), N), num(I) |> let O = fn:plus(M, N).`,
			wantErr: "evaluation produced something that is not a value",
		},
	} {
		u := unit(tt.program)
		store := factstore.NewSimpleInMemoryStore()
		err := analyzeAndEvalProgram(t, u.Clauses, store)
		if err == nil {
			t.Fatalf("analyzeAndEvalProgram(%s) expected to fail, but it didn't", tt.program)
		} else if !strings.Contains(err.Error(), tt.wantErr) {
			t.Fatalf("analyzeAndEvalProgram(%s) = %v want err contain %q", tt.program, err, tt.wantErr)
		}
	}
}

func TestFunctionEval(t *testing.T) {
	numbers := []ast.Clause{
		clause(`one(1).`),
		clause(`f_one(1.0).`),
	}

	for _, tt := range []struct {
		program string
		want    ast.Atom
		wantErr bool
	}{
		{
			program: `fun(O) :- one(I) |> let O = fn:div().`,
			wantErr: true,
		},
		{
			program: `fun(O) :- one(I) |> let O = fn:div(I).`,
			want:    atom("fun(1)"),
		},
		{
			program: `fun(O) :- one(I) |> let O = fn:div(0).`,
			wantErr: true,
		},
		{
			program: `fun(O) :- one(I) |> let O = fn:div(2).`,
			want:    atom("fun(0)"),
		},
		{
			program: `fun(O) :- one(I) |> let O = fn:div(1, 2).`,
			want:    atom("fun(0)"),
		},
		{
			program: `fun(O) :- one(I) |> let O = fn:div(10, 2, 2).`,
			want:    atom("fun(2)"),
		},
		{
			program: `fun(O) :- one(I) |> let O = fn:div(1.0, 2).`,
			wantErr: true,
		},
		{
			program: `fun(O) :- one(I) |> let O = fn:mult().`,
			want:    atom("fun(1)"),
		},
		{
			program: `fun(O) :- one(I) |> let O = fn:mult(2).`,
			want:    atom("fun(2)"),
		},
		{
			program: `fun(O) :- one(I) |> let O = fn:mult(1, 2).`,
			want:    atom("fun(2)"),
		},
		{
			program: `fun(O) :- one(I) |> let O = fn:mult(1.0, 2).`,
			wantErr: true,
		},
		{
			program: `fun(O) :- one(I) |> let O = fn:plus().`,
			want:    atom("fun(0)"),
		},
		{
			program: `fun(O) :- one(I) |> let O = fn:plus(2).`,
			want:    atom("fun(2)"),
		},
		{
			program: `fun(O) :- one(I) |> let O = fn:plus(1, 2).`,
			want:    atom("fun(3)"),
		},
		{
			program: `fun(O) :- one(I) |> let O = fn:plus(1.0, 2).`,
			wantErr: true,
		},
		{
			program: `fun(O) :- one(I) |> let O = fn:minus().`,
			wantErr: true,
		},
		{
			program: `fun(O) :- one(I) |> let O = fn:minus(2).`,
			want:    atom("fun(-2)"),
		},
		{
			program: `fun(O) :- one(I) |> let O = fn:minus(1, 2).`,
			want:    atom("fun(-1)"),
		},
		{
			program: `fun(O) :- one(I) |> let O = fn:minus(1.0, 2).`,
			wantErr: true,
		},
		{
			program: `fun(O) :- one(I) |> let O = fn:float:div().`,
			wantErr: true,
		},
		{
			program: `fun(O) :- f_one(I) |> let O = fn:float:div(I).`,
			want:    atom("fun(1.0)"),
		},
		{
			program: `fun(O) :- one(I) |> let O = fn:float:div(0.0).`,
			wantErr: true,
		},
		{
			program: `fun(O) :- one(I) |> let O = fn:float:div(2).`,
			want:    atom("fun(0.5)"),
		},
		{
			program: `fun(O) :- one(I) |> let O = fn:float:div(1, 2).`,
			want:    atom("fun(0.5)"),
		},
		{
			program: `fun(O) :- one(I) |> let O = fn:float:div(10, 2, 2.0).`,
			want:    atom("fun(2.5)"),
		},
		{
			program: `e(fn:min()).`,
			wantErr: true,
		},
		{
			program: `e(fn:min(1)).`,
			wantErr: true,
		},
		{
			program: `e(fn:min(1, 2)).`,
			wantErr: true,
		},
		{
			program: `e(fn:min([1, 2])).`,
			want:    atom("e(1)"),
		},
		{
			program: `e(fn:max()).`,
			wantErr: true,
		},
		{
			program: `e(fn:max(1)).`,
			wantErr: true,
		},
		{
			program: `e(fn:max(1, 2)).`,
			wantErr: true,
		},
		{
			program: `e(fn:max([1, 2])).`,
			want:    atom("e(2)"),
		},
		{
			program: `e(fn:float:min()).`,
			wantErr: true,
		},
		{
			program: `e(fn:float:min(1.0)).`,
			wantErr: true,
		},
		{
			program: `e(fn:float:min(1.1, 2.2)).`,
			wantErr: true,
		},
		{
			program: `e(fn:float:min([1.1, 2])).`,
			wantErr: true,
		},
		{
			program: `e(fn:float:min([1.1, 2.2])).`,
			want:    atom("e(1.1)"),
		},
		{
			program: `e(fn:float:max()).`,
			wantErr: true,
		},
		{
			program: `e(fn:float:max(1.0)).`,
			wantErr: true,
		},
		{
			program: `e(fn:float:max(1.1, 2.2)).`,
			wantErr: true,
		},
		{
			program: `e(fn:float:max([1.1, 2])).`,
			wantErr: true,
		},
		{
			program: `e(fn:float:max([1.1, 2.2])).`,
			want:    atom("e(2.2)"),
		},
		{
			program: `e(fn:sum()).`,
			wantErr: true,
		},
		{
			program: `e(fn:sum(1)).`,
			wantErr: true,
		},
		{
			program: `e(fn:sum(1, 2)).`,
			wantErr: true,
		},
		{
			program: `e(fn:sum([1, 2.2])).`,
			wantErr: true,
		},
		{
			program: `e(fn:sum([1, 2])).`,
			want:    atom("e(3)"),
		},
		{
			program: `e(fn:float:sum()).`,
			wantErr: true,
		},
		{
			program: `e(fn:float:sum(1.1)).`,
			wantErr: true,
		},
		{
			program: `e(fn:float:sum(1.1, 3.3)).`,
			wantErr: true,
		},
		{
			program: `e(fn:float:sum([1, 3.3])).`,
			wantErr: true,
		},
		{
			program: `e(fn:float:sum([1.1, 3.3])).`,
			want:    atom("e(4.4)"),
		},
		{
			program: `e(fn:list:append()).`,
			wantErr: true,
		},
		{
			program: `e(fn:list:append(1)).`,
			wantErr: true,
		},
		{
			program: `e(fn:list:append(1, 2)).`,
			wantErr: true,
		},
		{
			program: `e(fn:list:append(1, 2, 3)).`,
			wantErr: true,
		},
		{
			program: `e(fn:list:append([1, 2], 3)).`,
			want:    evalAtom(t, "e([1, 2, 3])"),
		},
		{
			program: `e(fn:list:contains()).`,
			wantErr: true,
		},
		{
			program: `e(fn:list:contains(1)).`,
			wantErr: true,
		},
		{
			program: `e(fn:list:contains(1, 2)).`,
			wantErr: true,
		},
		{
			program: `e(fn:list:contains(1, 2, 3)).`,
			wantErr: true,
		},
		{
			program: `e(fn:list:contains([1, 2], 2)).`,
			want:    atom("e(/true)"),
		},
		{
			program: `e(fn:list:cons()).`,
			wantErr: true,
		},
		{
			program: `e(fn:list:cons(1)).`,
			wantErr: true,
		},
		{
			program: `e(fn:list:cons(1, [2])).`,
			want:    evalAtom(t, "e([1,2])"),
		},
		{
			program: `e(fn:list:cons(1, 2)).`,
			wantErr: true,
		},
		{
			program: `e(fn:list:cons(1, 2, 3)).`,
			wantErr: true,
		},
		{
			program: `e(fn:pair()).`,
			wantErr: true,
		},
		{
			program: `e(fn:pair(1)).`,
			wantErr: true,
		},
		{
			program: `e(fn:pair(1, [2])).`,
			want:    evalAtom(t, "e(fn:pair(1,[2]))"),
		},
		{
			program: `e(fn:pair(1, 2)).`,
			want:    evalAtom(t, "e(fn:pair(1,2))"),
		},
		{
			program: `e(fn:pair(1, 2, 3)).`,
			wantErr: true,
		},
		{
			program: `e(fn:list:len()).`,
			wantErr: true,
		},
		{
			program: `e(fn:list:len(1)).`,
			wantErr: true,
		},
		{
			program: `e(fn:list:len(1, 2)).`,
			wantErr: true,
		},
		{
			program: `e(fn:list:len(1, 2, 3)).`,
			wantErr: true,
		},
		{
			program: `e(fn:list:len([1, 2])).`,
			want:    atom("e(2)"),
		},
		{
			program: `e(fn:list()).`,
			want:    evalAtom(t, "e([])"),
		},
		{
			program: `e(fn:list(1)).`,
			want:    evalAtom(t, "e([1])"),
		},
		{
			program: `e(fn:list(1, "x")).`,
			want:    evalAtom(t, "e([1, 'x'])"),
		},
		{
			program: `e(fn:tuple()).`,
			wantErr: true,
		},
		{
			program: `e(fn:tuple(1)).`,
			want:    atom("e(1)"),
		},
		{
			program: `e(fn:tuple(1, 2)).`,
			want:    evalAtom(t, "e(fn:tuple(1, 2))"),
		},
		{
			program: `e(fn:tuple(1, 2, 3)).`,
			want:    evalAtom(t, "e(fn:tuple(1, fn:tuple(2, 3)))"),
		},
		{
			program: `e(fn:tuple(1, 2, 3,4)).`,
			want:    evalAtom(t, "e(fn:tuple(1, fn:tuple(2, fn:tuple(3,4))))"),
		},
		{
			program: `e(fn:number:to_string()).`,
			wantErr: true,
		},
		{
			program: `e(fn:number:to_string(1, 2)).`,
			wantErr: true,
		},
		{
			program: `e(fn:number:to_string(1.1)).`,
			wantErr: true,
		},
		{
			program: `e(fn:number:to_string(1)).`,
			want:    atom("e('1')"),
		},
		{
			program: `e(fn:name:to_string()).`,
			wantErr: true,
		},
		{
			program: `e(fn:name:to_string(/a, /b)).`,
			wantErr: true,
		},
		{
			program: `e(fn:name:to_string(1)).`,
			wantErr: true,
		},
		{
			program: `e(fn:name:to_string(/abc)).`,
			want:    atom("e('/abc')"),
		},
		{
			program: `e(fn:string:concat()).`,
			want:    atom("e('')"),
		},
		{
			program: `e(fn:string:concat('a', 'b')).`,
			want:    atom("e('ab')"),
		},
		{
			program: `e(fn:string:concat([1, 2])).`,
			wantErr: true,
		},
		{
			program: `e(fn:list:get()).`,
			wantErr: true,
		},
		{
			program: `e(fn:list:get([])).`,
			wantErr: true,
		},
		{
			program: `e(fn:list:get([1, 2], 'a')).`,
			wantErr: true,
		},
		{
			program: `e(fn:list:get(1, 2)).`,
			wantErr: true,
		},
		{
			program: `e(fn:list:get([1, 2], 1)).`,
			want:    atom("e(2)"),
		},
		{
			program: `e(fn:list:get([1, 2], 5)).`,
			wantErr: true,
		},
		{
			program: `e(fn:map(0, 1)).`,
			want:    evalAtom(t, `e(fn:map(0, 1)).`),
		},
		{
			program: `e(fn:map(0)).`,
			wantErr: true,
		},
		{
			program: `e(fn:map:get()).`,
			wantErr: true,
		},
		{
			program: `e(fn:map:get(1)).`,
			wantErr: true,
		},
		{
			program: `e(fn:map:get(fn:map(1, 2), 1)).`,
			want:    atom("e(2)"),
		},
		{
			program: `e(fn:struct(0, 1)).`,
			want:    evalAtom(t, `e(fn:struct(0, 1)).`),
		},
		{
			program: `e(fn:struct(0)).`,
			wantErr: true,
		},
		{
			program: `e(fn:struct:get()).`,
			wantErr: true,
		},
		{
			program: `e(fn:struct:get(1)).`,
			wantErr: true,
		},
		{
			program: `e(fn:struct:get(fn:struct(1, 2), 1)).`,
			want:    atom("e(2)"),
		},
	} {
		t.Run(tt.program, func(t *testing.T) {
			store := factstore.NewSimpleInMemoryStore()
			u := unit(tt.program)
			err := analyzeAndEvalProgram(t, append(u.Clauses, numbers...), store)
			if tt.wantErr {
				if err == nil {
					t.Errorf("eval(%v) expected to fail, but it didn't, store: %v", tt.program, store)
				}
				return
			}
			if err != nil {
				t.Errorf("eval(%v) unexpected error: %v", tt.program, err)
				return
			}
			if !store.Contains(tt.want) {
				t.Errorf("expected fact %v in store %v", tt.want, store)
			}
		})
	}
}

func TestFunctions(t *testing.T) {
	numbers := []ast.Clause{
		clause(`zero(0).`),
		clause(`one(1).`),
		clause(`two(2).`),
	}

	for _, tt := range []struct {
		program string
		want    ast.Atom
		wantErr bool
	}{
		{
			program: `fun(O) :- one(I)|> let O = fn:div(I).`,
			want:    atom("fun(1)"),
		},
		{
			program: `fun(O) :- zero(Z)|> let O = fn:div(Z).`,
			wantErr: true,
		},
		{
			program: `fun(O) :- one(I) |> let O = fn:div(I, 2).`,
			want:    atom("fun(0)"),
		},
		{
			program: `fun(O) :- one(I) |> let O = fn:div(2, I).`,
			want:    atom("fun(2)"),
		},
	} {
		t.Run(tt.program, func(t *testing.T) {
			store := factstore.NewSimpleInMemoryStore()
			u := unit(tt.program)
			err := analyzeAndEvalProgram(t, append(u.Clauses, numbers...), store)
			if tt.wantErr {
				if err == nil {
					t.Errorf("eval(%v) did not fail, store: %v", tt.program, store)
				}
				return
			}
			if err != nil {
				t.Errorf("eval(%v) unexpected error: %v", tt.program, err)
				return
			}
			if !store.Contains(tt.want) {
				t.Errorf("expected fact %v in store %v", tt.want, store)
			}
		})
	}
}

func BenchmarkJoin(b *testing.B) {
	b.ReportAllocs()
	for j := 0; j < b.N; j++ {
		// Given two relations has_num(Thing, Num) has_prop(Thing, Property),
		// we select from one and join with the other.
		program := []ast.Clause{
			clause(`foo(X, Prop) :- has_num(X, Num), Num <= 1, has_prop(X, Prop).`),
		}
		store := factstore.NewIndexedInMemoryStore()
		for i := 0; i < 1_000_000; i++ {
			id, _ := ast.Name(fmt.Sprintf("/thing/%d", i))
			store.Add(ast.NewAtom("has_num", id, ast.Number(int64(i%3))))
			store.Add(ast.NewAtom("has_prop", id, ast.String(fmt.Sprintf("/property/%d", i))))
		}
		programInfo, err := analysis.AnalyzeOneUnit(parse.SourceUnit{Clauses: program}, asMap(store.ListPredicates()))
		if err != nil {
			b.Fatal(fmt.Errorf("analysis: %w", err))
		}
		if err := EvalProgram(programInfo, store); err != nil {
			b.Fatal(fmt.Errorf("evaluation: %w", err))
		}
	}
}
