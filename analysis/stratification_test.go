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
	"sort"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"codeberg.org/TauCeti/mangle-go/ast"
	"codeberg.org/TauCeti/mangle-go/parse"
)

func clause(str string) ast.Clause {
	clause, err := parse.Clause(str)
	if err != nil {
		panic(fmt.Errorf("bad syntax in test case: %s got %w", str, err))
	}
	return clause
}

func toOrderMap(predToStratum map[ast.PredicateSym]int) map[int][]ast.PredicateSym {
	unsorted := make(map[int][]ast.PredicateSym)
	for sym, order := range predToStratum {
		unsorted[order] = append(unsorted[order], sym)
	}
	for _, slice := range unsorted {
		sort.Slice(slice, func(i, j int) bool { return slice[i].Symbol < slice[j].Symbol })
	}
	return unsorted
}

func analyze(clauses []ast.Clause) (*ProgramInfo, error) {
	return AnalyzeOneUnit(parse.SourceUnit{Clauses: clauses}, nil)
}

func TestStratificationPositive(t *testing.T) {
	tests := []struct {
		name            string
		program         func() (*ProgramInfo, error)
		wantStrataOrder map[int][]ast.PredicateSym
	}{
		{
			name: "Ignore built-in predicates",
			program: func() (*ProgramInfo, error) {
				return analyze([]ast.Clause{
					clause("foo(X) :- :list:member(X, [/a])."),
				})
			},
			wantStrataOrder: map[int][]ast.PredicateSym{
				0: {{"foo", 1}},
			},
		},
		{
			name: "Cycles are ok as long as they are positive",
			program: func() (*ProgramInfo, error) {
				return analyze([]ast.Clause{
					clause("num(/one)."),
					clause("num(/two)."),
					clause("num(/three)."),
					clause("succ(/one, /two)."),
					clause("succ(/two, /three)."),
					clause("odd(/one)."),
					clause("odd(X) :- num(X), succ(Y,X), even(Y)."),
					clause("even(X) :- num(X), succ(X,Y), odd(X)."),
					clause("count(Z) :- odd(X) |> do fn:group_by(), let Z = fn:count()."),
				})
			},
			wantStrataOrder: map[int][]ast.PredicateSym{
				0: {{"even", 1}, {"odd", 1}},
				1: {{"count", 1}},
			},
		},
		{
			name: "The result is ordered by dependencies",
			program: func() (*ProgramInfo, error) {
				return analyze([]ast.Clause{
					clause("num(/one)."),
					clause("a(X) :- num(X), b(X)."),
					clause("b(X) :- num(X), c(X)."),
					clause("c(X) :- num(X), d(X), b(X)."),
					clause("d(X) :- num(X)."),
				})
			},
			wantStrataOrder: map[int][]ast.PredicateSym{
				0: {{"d", 1}},
				1: {{"b", 1}, {"c", 1}},
				2: {{"a", 1}},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			program, err := test.program()
			if err != nil {
				t.Fatalf("test case unexpectedly did not pass  %v", err)
			}
			strata, predToStratum, err := Stratify(Program{
				program.EdbPredicates, program.IdbPredicates, program.Rules})
			if err != nil {
				t.Fatalf("expected stratification to succeed, got %v", err)
			}
			got := toOrderMap(predToStratum)
			if len(strata) != len(got) {
				t.Errorf("unexpected number of strata wanted %v, got: %v", len(strata), len(got))
			}
			if diff := cmp.Diff(test.wantStrataOrder, got, cmpopts.SortMaps(func(x, y int) bool { return x < y })); diff != "" {
				t.Errorf("want %v, got: %v", test.wantStrataOrder, got)
			}
		})
	}
}

// Separate test since multiple legitimate possible due to map iteration order.
func TestStratificationMultipleStrata(t *testing.T) {
	program, err := analyze([]ast.Clause{
		clause("node(/foo)."),
		clause("node(/bar)."),
		clause("edge(/foo, /bar)."),
		clause("path(X,Y) :- edge(X,Y)."),
		clause("path(X,Z) :- edge(X,Y), path(Y,Z)."),
		clause("not_reachable(X, Y) :- node(X), node(Y), !path(X, Y)."),
		clause("in_cycle_eq(X) :- node(X), path(X, Y), X = Y."),
		clause("in_between(X, Y) :- node(X), node(Y), node(Z), path(X, Y), path(Y, Z), X != Y, Y != Z, X != Z."),
	})

	if err != nil {
		t.Fatalf("test case unexpectedly did not pass  %v", err)
	}

	strata, predToStratum, err := Stratify(Program{
		program.EdbPredicates, program.IdbPredicates, program.Rules})
	if err != nil {
		t.Fatalf("expected stratification to succeed, got %v", err)
	}

	if len(strata) != 4 {
		t.Fatalf("expected 4 strata, got %v", len(strata))
	}

	path, ok := predToStratum[ast.PredicateSym{"path", 2}]
	if !ok {
		t.Fatal("couldn't find 'path'")
	}
	inBetween, ok := predToStratum[ast.PredicateSym{"in_between", 2}]
	if !ok {
		t.Fatal("couldn't find 'in_between'")
	}
	inCycleEq, ok := predToStratum[ast.PredicateSym{"in_cycle_eq", 1}]
	if !ok {
		t.Fatal("couldn't find 'in_cycle_eq'")
	}
	notReachable, ok := predToStratum[ast.PredicateSym{"not_reachable", 2}]
	if !ok {
		t.Fatal("couldn't find 'not_reachable'")
	}

	if path >= inBetween {
		t.Error("expected 'path' < 'in_between'")
	}
	if path >= inCycleEq {
		t.Error("expected 'path' < 'in_cycle_eq'")
	}
	if path >= notReachable {
		t.Error("expected 'path' < 'not_reachable'")
	}
}

func TestStratificationNegative(t *testing.T) {
	tests := []func() (*ProgramInfo, error){
		func() (*ProgramInfo, error) {
			return analyze([]ast.Clause{
				clause("bar(/baz)."),
				clause("foo(X) :- !sna(X), bar(X)."),
				clause("sna(X) :- !foo(X), bar(X)."),
			})
		},
		func() (*ProgramInfo, error) {
			return analyze([]ast.Clause{
				clause("yes(/yes)."),
				clause("no(/no)."),
				clause("yesorno(X) :- !yesorno(X), yes(X)."),
				clause("yesorno(X) :- yesorno(X), no(X)."),
			})
		},
		func() (*ProgramInfo, error) {
			return analyze([]ast.Clause{
				clause("rec(/yes)."),
				clause("rec(Z) :- rec(X) |> do fn:group_by(), let Z = fn:count()."),
			})
		},
	}

	for _, testprogram := range tests {
		program, err := testprogram()
		if err != nil {
			panic(fmt.Errorf("test case did not pass  %v", err))
		}
		nodes, predtostratum, err := Stratify(Program{
			program.EdbPredicates, program.IdbPredicates, program.Rules})
		if err == nil {
			t.Errorf("expected stratification to fail, but succeeded %v %v", nodes, predtostratum)
		}
	}
}
