// Copyright 2026 The Mangle Authors
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
	"sort"
	"strings"
	"testing"

	"codeberg.org/TauCeti/mangle-go/analysis"
	"codeberg.org/TauCeti/mangle-go/ast"
	"codeberg.org/TauCeti/mangle-go/factstore"
	"codeberg.org/TauCeti/mangle-go/parse"
)

func analyzeSource(t *testing.T, source string) *analysis.ProgramInfo {
	t.Helper()
	unit, err := parse.Unit(strings.NewReader(source))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	pi, err := analysis.AnalyzeOneUnit(unit, nil)
	if err != nil {
		t.Fatalf("analyze: %v", err)
	}
	return pi
}

func premisesString(c ast.Clause) string {
	var parts []string
	for _, p := range c.Premises {
		parts = append(parts, p.String())
	}
	return strings.Join(parts, ", ")
}

func TestReorderPremises(t *testing.T) {
	const base = `Decl a(X). Decl b(Y). Decl big(X). Decl link(X, Y). Decl bad(X).
		Decl c(X, Y). Decl lists(L). Decl q(X). Decl numval(X, T).
		Decl tiny(X, Y). Decl huge(K, Y).
		Decl moded(X, Y) descr [mode('+', '-')].
		Decl activity(X) temporal.
		Decl frozen(X, Y) descr [no_reorder()].
	`
	tests := []struct {
		name string
		rule string
		// want is the premises in the order chosen, or "" if the written
		// order is kept.
		want string
	}{{
		name: "a cross product is avoided",
		rule: "r(X, Y) :- a(X), b(Y), link(X, Y).",
		want: "a(X), link(X,Y), b(Y)",
	}, {
		name: "a rule without a cross product keeps its order",
		rule: "r(X, Z) :- a(X), link(X, Y), c(Y, Z).",
	}, {
		// A negation looks up facts for every row: moved ahead of link, it
		// would run on every a(X) instead of the few rows link keeps.
		name: "a negation never moves ahead of a premise written before it",
		rule: "r(X, Y) :- a(X), link(X, Y), c(Y, Z), !bad(X).",
	}, {
		name: "an inequality filters as soon as its variables are bound",
		rule: "r(X, Y) :- a(X), link(X, Y), c(Y, Z), X != 3.",
		want: "a(X), X != 3, link(X,Y), c(Y,Z)",
	}, {
		name: "an atom whose variables are all bound never moves ahead of a premise written before it",
		rule: "r(X, Y) :- a(X), link(X, Y), big(X).",
	}, {
		name: "an atom with a wildcard is not a filter: it gives a row per match",
		rule: "r(X, Y) :- a(X), link(X, Y), c(X, _).",
	}, {
		// An equality between two variables is compiled away by the analysis
		// (#104), so what reaches the planner binds a variable to a constant.
		name: "an equality that binds a variable to a value goes early",
		rule: "r(X, Y, Z) :- a(X), link(X, Y), Z = 3.",
		want: "Z = 3, a(X), link(X,Y)",
	}, {
		name: "an equality with a constant goes first, so that lookups use the value",
		rule: "r(X, Y) :- a(X), link(X, Y), X = 1.",
		want: "X = 1, a(X), link(X,Y)",
	}, {
		// How many facts huge(/k, Y) matches depends on the data: written
		// after the selective tiny, it stays there.
		name: "a constant does not make an atom go first",
		rule: "r(Y) :- a(X), tiny(X, Y), huge(/k, Y).",
	}, {
		name: "a constant does not make an atom join: the cross product is still avoided",
		rule: "r(Y) :- big(X), c(/k, Y), link(X, Y).",
		want: "big(X), link(X,Y), c(/k,Y)",
	}, {
		name: "a comparison never moves ahead of a premise written before it",
		rule: "r(X, Y) :- a(X), link(X, Y), X < 10.",
	}, {
		name: "a comparison waits for the premises before it, while others move",
		rule: "r(X, Y) :- a(X), b(Y), X < 10, link(X, Y).",
		want: "a(X), link(X,Y), b(Y), :lt(X,10)",
	}, {
		name: "a function application never moves ahead of a premise written before it",
		rule: "r(X, Z) :- a(X), numval(X, T), Z = fn:plus(X, 1).",
	}, {
		name: "a built-in that binds an output keeps its place",
		rule: "r(H) :- lists(L), :match_cons(L, H, T), q(H).",
	}, {
		name: "a rule with a moded predicate is evaluated as written",
		rule: "r(X, Y) :- a(X), b(Y), moded(X, Y).",
	}, {
		name: "a rule with a premise over a no_reorder predicate is evaluated as written",
		rule: "r(X, Y) :- a(X), b(Y), frozen(X, Y).",
	}, {
		name: "a premise over a no_reorder predicate keeps a recursive rule as written too",
		rule: "r(X, Y) :- b(Y), r(X, Z), frozen(Z, Y).",
	}, {
		name: "a rule with a temporal predicate is evaluated as written",
		rule: "r(X, Y) :- a(X), b(Y), link(X, Y), activity(X)@[_, _].",
	}, {
		name: "a let-transform is kept, and the body reordered",
		rule: "r(X, N) :- a(X), b(Y), link(X, Y) |> let N = fn:plus(Y, 1).",
		want: "a(X), link(X,Y), b(Y)",
	}}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			pi := analyzeSource(t, base+test.rule)
			if len(pi.Rules) != 1 {
				t.Fatalf("got %d rules, want 1", len(pi.Rules))
			}
			rule := pi.Rules[0]
			got, changed := ReorderPremises(rule, pi.Decls)
			if test.want == "" {
				if changed {
					t.Errorf("ReorderPremises(%v) = %v, want the written order", rule, got)
				}
				return
			}
			if !changed || premisesString(got) != test.want {
				t.Errorf("ReorderPremises(%v) = %v (changed %v), want premises %s", rule, got, changed, test.want)
			}
			if got.Head.String() != rule.Head.String() || (got.Transform == nil) != (rule.Transform == nil) {
				t.Errorf("ReorderPremises(%v) changed more than the premises: %v", rule, got)
			}
		})
	}
}

// When no premise can go next, the written order is kept.
func TestReorderPremisesKeepsAnOrderItCannotImprove(t *testing.T) {
	rule := clause("r(X) :- a(X), !b(X, Y).")
	if got, changed := ReorderPremises(rule, nil); changed {
		t.Errorf("ReorderPremises(%v) = %v, want the written order", rule, got)
	}
}

func allFacts(store factstore.FactStore) []string {
	var facts []string
	for _, p := range store.ListPredicates() {
		store.GetFacts(ast.NewQuery(p), func(a ast.Atom) error {
			facts = append(facts, a.String())
			return nil
		})
	}
	sort.Strings(facts)
	return facts
}

func TestPremiseReorderingReportsWhatItChanged(t *testing.T) {
	pi := analyzeSource(t, `
		a(1). a(2). b(1). b(2). link(1, 2). edge(1, 2). edge(2, 3). start(1).
		r(X, Y) :- a(X), b(Y), link(X, Y).
		t(X, Y) :- a(X), link(X, Y).
		reach(X) :- start(X).
		reach(Z) :- reach(X), edge(Y, Z), edge(X, Y).
	`)
	stats, err := EvalProgramWithStats(pi, factstore.NewSimpleInMemoryStore(), WithPremiseReordering())
	if err != nil {
		t.Fatal(err)
	}
	var got []string
	for _, o := range stats.PremiseOrders {
		newFacts := "first round"
		if o.NewFacts >= 0 {
			newFacts = "new facts: " + o.Rule.Premises[o.NewFacts].String()
		}
		got = append(got, o.Rule.String()+" => "+premisesString(o.Evaluated)+" ("+newFacts+")")
	}
	sort.Strings(got)
	want := []string{
		"r(X,Y) :- a(X), b(Y), link(X,Y). => a(X), link(X,Y), b(Y) (first round)",
		"reach(Z) :- reach(X), edge(Y,Z), edge(X,Y). => reach(X), edge(X,Y), edge(Y,Z) (first round)",
		"reach(Z) :- reach(X), edge(Y,Z), edge(X,Y). => reach(X), edge(X,Y), edge(Y,Z) (new facts: reach(X))",
	}
	if strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Errorf("PremiseOrders:\n%s\nwant:\n%s", strings.Join(got, "\n"), strings.Join(want, "\n"))
	}

	stats, err = EvalProgramWithStats(pi, factstore.NewSimpleInMemoryStore())
	if err != nil {
		t.Fatal(err)
	}
	if len(stats.PremiseOrders) != 0 {
		t.Errorf("without WithPremiseReordering, PremiseOrders = %v, want none", stats.PremiseOrders)
	}
}

// Moving a premise that can fail ahead of one written before it could make
// it see a row that the earlier premise removes. Here a("x") is removed by
// numval before the comparison and the addition would fail on it.
func TestPremiseReorderingNeverAddsAnError(t *testing.T) {
	pi := analyzeSource(t, `
		a(1). a(2). a("x"). numval(1, /n). numval(2, /n).
		small(X) :- a(X), numval(X, T), X < 2.
		next(X, Z) :- a(X), numval(X, T), Z = fn:plus(X, 1).
	`)
	written := factstore.NewSimpleInMemoryStore()
	if err := EvalProgram(pi, written); err != nil {
		t.Fatalf("as written: %v", err)
	}
	reordered := factstore.NewSimpleInMemoryStore()
	if err := EvalProgram(pi, reordered, WithPremiseReordering()); err != nil {
		t.Fatalf("reordered: %v", err)
	}
	if got, want := allFacts(reordered), allFacts(written); strings.Join(got, " ") != strings.Join(want, " ") {
		t.Errorf("reordered: %v\nwritten: %v", got, want)
	}
}

// A merge predicate combines the facts of a round as they come out of the
// store, in no particular order already. Reordering premises changes when a
// rule derives its facts, not which, so the merged facts are the same. Every
// pair of nodes here has a single shortest path, so that the result does not
// depend on the order of merging either.
func TestPremiseReorderingWithAMergePredicate(t *testing.T) {
	pi := analyzeSource(t, `
		Decl shortest_path(X, Y, P) descr [fundep([X, Y], [P]), merge([P], 'shorter')].
		Decl shorter(P1, P2, P) descr [mode('+', '+', '-'), deferred()].
		node(/a). node(/b). node(/c). node(/d). node(/e).
		edge(/a, /b). edge(/b, /c). edge(/c, /d). edge(/a, /d). edge(/d, /e).
		shortest_path(X, Y, [Y, X]) :- edge(X, Y).
		shortest_path(X, Z, NewPath) :- edge(Y, Z), node(X), shortest_path(X, Y, Path)
			|> let NewPath = fn:list:cons(Z, Path).
		shorter(P1, P2, P) :- fn:list:len(P1) < fn:list:len(P2), P = P1.
		shorter(P1, P2, P) :- fn:list:len(P2) <= fn:list:len(P1), P = P2.
	`)
	for run := 0; run < 20; run++ {
		written := factstore.NewSimpleInMemoryStore()
		if err := EvalProgram(pi, written); err != nil {
			t.Fatalf("as written: %v", err)
		}
		reordered := factstore.NewSimpleInMemoryStore()
		stats, err := EvalProgramWithStats(pi, reordered, WithPremiseReordering())
		if err != nil {
			t.Fatalf("reordered: %v", err)
		}
		if len(stats.PremiseOrders) == 0 {
			t.Fatal("no rule was reordered: the test needs a rule written in a poor order")
		}
		got, want := strings.Join(allFacts(reordered), " "), strings.Join(allFacts(written), " ")
		if got != want {
			t.Fatalf("run %d:\nreordered: %s\nwritten:   %s", run, got, want)
		}
		if !strings.Contains(got, "shortest_path(/a,/e,[/e, /d, /a])") ||
			strings.Contains(got, "shortest_path(/a,/e,[/e, /d, /c, /b, /a])") {
			t.Fatalf("run %d: want only the shortest path from /a to /e: %s", run, got)
		}
	}
}

// fn:collect lists the rows of a group in the order the store returns them,
// which is not specified: the list differs from run to run, with or without
// reordering. The elements, that is the rows, are the same.
func TestPremiseReorderingKeepsCollectedRows(t *testing.T) {
	pi := analyzeSource(t, `
		a(1). a(2). a(3). b(1). b(2). b(3). link(1, 2). link(2, 1). link(3, 3). link(1, 3).
		col(L) :- a(X), b(Y), link(X, Y) |> do fn:group_by(), let L = fn:collect(X).
	`)
	collected := func(opts ...EvalOption) []string {
		store := factstore.NewSimpleInMemoryStore()
		if _, err := EvalProgramWithStats(pi, store, opts...); err != nil {
			t.Fatal(err)
		}
		var elems []string
		store.GetFacts(ast.NewQuery(ast.PredicateSym{Symbol: "col", Arity: 1}), func(fact ast.Atom) error {
			list := fact.Args[0].(ast.Constant)
			list.ListValues(func(c ast.Constant) error {
				elems = append(elems, c.String())
				return nil
			}, func() error { return nil })
			return nil
		})
		sort.Strings(elems)
		return elems
	}
	want := strings.Join(collected(), " ")
	if want != "1 1 2 3" {
		t.Fatalf("as written, collected %s, want 1 1 2 3", want)
	}
	for run := 0; run < 20; run++ {
		if got := strings.Join(collected(WithPremiseReordering()), " "); got != want {
			t.Fatalf("run %d: reordered collected %s, written %s", run, got, want)
		}
	}
}

// The other direction of TestPremiseReorderingNeverAddsAnError: reordering can
// make an evaluation succeed that fails as written. Written, the comparison
// sees the string from lists2 and fails. Reordered as a, gen, lists2, :lt,
// gen finds no fact for X = 1, so no row reaches the comparison.
func TestPremiseReorderingCanAvoidAnError(t *testing.T) {
	pi := analyzeSource(t, `
		a(1). lists2("not a number"). gen(2, /k).
		p(X, K) :- a(X), lists2(L), L < 3, gen(X, K).
	`)
	if err := EvalProgram(pi, factstore.NewSimpleInMemoryStore()); err == nil || !strings.Contains(err.Error(), "not a number") {
		t.Fatalf("as written: got error %v, want one about the string", err)
	}
	store := factstore.NewSimpleInMemoryStore()
	stats, err := EvalProgramWithStats(pi, store, WithPremiseReordering())
	if err != nil {
		t.Fatalf("reordered: %v", err)
	}
	if len(stats.PremiseOrders) != 1 || premisesString(stats.PremiseOrders[0].Evaluated) != "a(X), gen(X,K), lists2(L), :lt(L,3)" {
		t.Fatalf("PremiseOrders = %v, want the rule reordered as a, gen, lists2, :lt", stats.PremiseOrders)
	}
	if facts := allFacts(store); strings.Contains(strings.Join(facts, " "), "p(") {
		t.Errorf("reordered derived %v, want no p facts", facts)
	}
}

// Stats.PremiseOrders has one entry for a reordered rule, and one for its
// delta variant, however many rounds the recursion takes: here 50, two edges
// per round along a chain of 100.
func TestPremiseOrdersHaveOneEntryPerRule(t *testing.T) {
	var sb strings.Builder
	sb.WriteString("start(0).\n")
	for i := 0; i < 100; i++ {
		fmt.Fprintf(&sb, "edge(%d, %d).\n", i, i+1)
	}
	sb.WriteString("reach(X) :- start(X).\nreach(Z) :- reach(X), edge(Y, Z), edge(X, Y).\n")
	pi := analyzeSource(t, sb.String())
	store := factstore.NewSimpleInMemoryStore()
	stats, err := EvalProgramWithStats(pi, store, WithPremiseReordering())
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(strings.Join(allFacts(store), " "), "reach(100)") {
		t.Fatal("the recursion did not reach the end of the chain")
	}
	var newFacts []int
	for _, o := range stats.PremiseOrders {
		newFacts = append(newFacts, o.NewFacts)
	}
	sort.Ints(newFacts)
	if fmt.Sprint(newFacts) != "[-1 0]" {
		t.Errorf("PremiseOrders has entries with NewFacts %v, want [-1 0]: one for the rule, one for its delta variant", newFacts)
	}
}

// A no_reorder predicate that is itself recursive keeps its rules as written
// in every round: in the delta variant, the premise over it ranges over the
// previous round's new facts, and is still recognized. Without no_reorder,
// node(Z) would be followed by edge(Y, Z), ahead of path(X, Y).
func TestNoReorderKeepsDeltaVariantsAsWritten(t *testing.T) {
	pi := analyzeSource(t, `
		Decl path(X, Y) descr [no_reorder()].
		node(1). node(2). node(3). node(4). edge(1, 2). edge(2, 3). edge(3, 4).
		path(X, Y) :- edge(X, Y).
		path(X, Z) :- node(Z), path(X, Y), edge(Y, Z).
	`)
	written := factstore.NewSimpleInMemoryStore()
	if err := EvalProgram(pi, written); err != nil {
		t.Fatal(err)
	}
	store := factstore.NewSimpleInMemoryStore()
	stats, err := EvalProgramWithStats(pi, store, WithPremiseReordering())
	if err != nil {
		t.Fatal(err)
	}
	if len(stats.PremiseOrders) != 0 {
		t.Errorf("PremiseOrders = %v, want none: the rule has a premise over path", stats.PremiseOrders)
	}
	if got, want := strings.Join(allFacts(store), " "), strings.Join(allFacts(written), " "); got != want || !strings.Contains(got, "path(1,4)") {
		t.Errorf("reordered: %s\nwritten:   %s", got, want)
	}
}
