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
	"sort"
	"strings"
	"testing"

	"codeberg.org/TauCeti/mangle-go/analysis"
	"codeberg.org/TauCeti/mangle-go/ast"
	"codeberg.org/TauCeti/mangle-go/engine"
	"codeberg.org/TauCeti/mangle-go/factstore"
	"codeberg.org/TauCeti/mangle-go/parse"
)

// proofsOf evaluates source with a recorder and returns, for every derived
// fact, the IDs of its proofs and the rules they show.
func proofsOf(t *testing.T, source string, opts ...engine.EvalOption) map[string]string {
	t.Helper()
	unit, err := parse.Unit(strings.NewReader(source))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	pi, err := analysis.AnalyzeOneUnit(unit, nil)
	if err != nil {
		t.Fatalf("analyze: %v", err)
	}
	store := factstore.NewSimpleInMemoryStore()
	rec := NewMemoryRecorder()
	stats, err := engine.EvalProgramWithStats(pi, store, append(opts, engine.WithDerivationRecorder(rec))...)
	if err != nil {
		t.Fatalf("eval: %v", err)
	}
	if len(opts) > 0 && len(stats.PremiseOrders) == 0 {
		t.Fatalf("no rule was reordered: the test needs a rule written in a poor order")
	}
	proofs := make(map[string]string)
	for pred := range pi.IdbPredicates {
		store.GetFacts(ast.NewQuery(pred), func(fact ast.Atom) error {
			nodes, err := BuildFromRecording(rec, store, fact, Options{})
			if err != nil {
				t.Fatalf("BuildFromRecording(%v): %v", fact, err)
			}
			var parts []string
			for _, n := range nodes {
				parts = append(parts, n.ID+" "+n.Rule.String())
			}
			sort.Strings(parts)
			proofs[fact.String()] = strings.Join(parts, "; ")
			return nil
		})
	}
	return proofs
}

// Reordering premises changes how a rule is evaluated, not what it means:
// the proofs show each rule as written, and are the same.
func TestPremiseReorderingKeepsProofs(t *testing.T) {
	source := `
		a(1). a(2). a(3). b(1). b(2). b(3). link(1, 2). link(2, 3). bad(3).
		start(1). edge(1, 2). edge(2, 3). edge(3, 4). edge(1, 3).
		r(X, Y) :- a(X), b(Y), link(X, Y).
		s(X, Y) :- a(X), b(Y), link(X, Y), !bad(X).
		reach(X) :- start(X).
		reach(Z) :- reach(X), edge(Y, Z), edge(X, Y).
	`
	written := proofsOf(t, source)
	reordered := proofsOf(t, source, engine.WithPremiseReordering())
	if len(written) == 0 {
		t.Fatal("no derived facts")
	}
	for fact, want := range written {
		if got := reordered[fact]; got != want {
			t.Errorf("proofs of %s:\nreordered: %s\nwritten:   %s", fact, got, want)
		}
	}
	if len(reordered) != len(written) {
		t.Errorf("reordered derived %d facts, written %d", len(reordered), len(written))
	}
}
