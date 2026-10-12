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

package provenance

import (
	"strings"
	"testing"

	"codeberg.org/TauCeti/mangle-go/ast"
)

// The store hands facts to the engine in map order, so p(0)'s five derivations
// are recorded in a different order from run to run. EventsFor must return
// them in the same order every time: by the facts they rest on.
func TestAlternativeDerivationsComeInAStableOrder(t *testing.T) {
	source := `
		a(1). a(2). a(3). a(4). a(5).
		p(0) :- a(X).
	`
	want := "a(1) a(2) a(3) a(4) a(5)"
	for run := 0; run < 50; run++ {
		_, _, rec := buildEvalRec(t, source)
		var got []string
		for _, ev := range rec.EventsFor(mustAtom(t, "p(0)")) {
			got = append(got, ev.PremiseFacts[0].String())
		}
		if g := strings.Join(got, " "); g != want {
			t.Fatalf("run %d: derivations of p(0) rest on %s, want %s", run, g, want)
		}
	}
}

// Derivations by different rules come in rule order. p(1) is derived first by
// the rule from start (in the first round) and later by the rule from q, but
// the rule from q sorts first.
func TestAlternativeDerivationsAreOrderedByRule(t *testing.T) {
	source := `
		start(1).
		q(X) :- start(X).
		p(X) :- start(X).
		p(X) :- q(X).
	`
	_, _, rec := buildEvalRec(t, source)
	var got []string
	for _, ev := range rec.EventsFor(mustAtom(t, "p(1)")) {
		got = append(got, ev.Rule.String())
	}
	want := []string{"p(X) :- q(X).", "p(X) :- start(X)."}
	if strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Errorf("derivations of p(1) by %q, want %q", got, want)
	}
}

// c(2) is derived from two groups, and each group's input facts also arrive in
// map order. The groups must still come in the same order every time.
func TestAlternativeAggregatesComeInAStableOrder(t *testing.T) {
	source := `
		a(1, 7). a(3, 7). a(2, 8). a(4, 8).
		c(N) :- a(X, Y) |> do fn:group_by(Y), let N = fn:count().
	`
	want := "7 8"
	for run := 0; run < 100; run++ {
		_, _, rec := buildEvalRec(t, source)
		var got []string
		for _, ev := range rec.EventsFor(mustAtom(t, "c(2)")) {
			got = append(got, ev.GroupKey[0].String())
		}
		if g := strings.Join(got, " "); g != want {
			t.Fatalf("run %d: derivations of c(2) from groups %s, want %s", run, g, want)
		}
	}
}

// p(10) is derived from five rows of a let-transform, which come from the store
// in map order. They must come in the same order every time: by their bindings.
func TestAlternativeLetRowsComeInAStableOrder(t *testing.T) {
	source := `
		a(1, 10). a(2, 10). a(3, 10). a(4, 10). a(5, 10).
		p(N) :- a(X, Y) |> let N = fn:mult(Y, 1).
	`
	want := "1 2 3 4 5"
	for run := 0; run < 100; run++ {
		_, _, rec := buildEvalRec(t, source)
		var got []string
		for _, ev := range rec.EventsFor(mustAtom(t, "p(10)")) {
			got = append(got, ev.Row.Get(ast.Variable{Symbol: "X"}).String())
		}
		if g := strings.Join(got, " "); g != want {
			t.Fatalf("run %d: derivations of p(10) from rows with X = %s, want %s", run, g, want)
		}
	}
}
