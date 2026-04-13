// Copyright 2026 Google LLC
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
	"bytes"
	"strings"
	"testing"

	"codeberg.org/TauCeti/mangle-go/analysis"
	"codeberg.org/TauCeti/mangle-go/ast"
	"codeberg.org/TauCeti/mangle-go/engine"
	"codeberg.org/TauCeti/mangle-go/factstore"
	"codeberg.org/TauCeti/mangle-go/parse"
)

// buildEvalRec parses, analyzes, evaluates the program with a
// MemoryRecorder attached, and returns (programInfo, store, recorder).
func buildEvalRec(t *testing.T, source string) (*analysis.ProgramInfo, factstore.SimpleInMemoryStore, *MemoryRecorder) {
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
	for _, f := range pi.InitialFacts {
		store.Add(f)
	}
	strata, predToStratum, err := analysis.Stratify(analysis.Program{
		EdbPredicates: pi.EdbPredicates,
		IdbPredicates: pi.IdbPredicates,
		Rules:         pi.Rules,
	})
	if err != nil {
		t.Fatalf("stratify: %v", err)
	}
	rec := NewMemoryRecorder()
	if _, err := engine.EvalStratifiedProgramWithStats(pi, strata, predToStratum, store,
		engine.WithDerivationRecorder(rec)); err != nil {
		t.Fatalf("eval: %v", err)
	}
	return pi, store, rec
}

// leafFacts walks a proof tree, collecting the set of EDB-leaf atom strings.
func leafFacts(n *ProofNode, seen map[string]bool, out map[string]bool) {
	if n == nil || seen[n.ID] {
		return
	}
	seen[n.ID] = true
	switch n.Kind {
	case KindEDB:
		out[n.Fact.String()] = true
	default:
		for _, sub := range n.Premises {
			leafFacts(sub, seen, out)
		}
	}
}

func TestFullModeMatchesSimpleOnDatalog(t *testing.T) {
	source := `
		edge(1, 2).
		edge(2, 3).
		edge(3, 4).
		path(X, Y) :- edge(X, Y).
		path(X, Z) :- edge(X, Y), path(Y, Z).
	`
	pi, store, rec := buildEvalRec(t, source)
	goal := mustAtom(t, "path(1, 4)")

	simple, err := Explain(pi, store, goal, Options{})
	if err != nil {
		t.Fatalf("Explain: %v", err)
	}
	full, err := BuildFromRecording(rec, store, goal, Options{})
	if err != nil {
		t.Fatalf("BuildFromRecording: %v", err)
	}

	// The two modes may pick different alternative proofs, but the set of
	// EDB facts they bottom out in must be identical for a given chosen
	// proof. We assert: both produce non-empty results AND the leaf sets
	// for each first proof are identical (they should share the same
	// source edges).
	simpleLeaves := make(map[string]bool)
	leafFacts(simple[0], make(map[string]bool), simpleLeaves)
	fullLeaves := make(map[string]bool)
	leafFacts(full[0], make(map[string]bool), fullLeaves)
	if len(simpleLeaves) == 0 {
		t.Errorf("simple proof has no EDB leaves")
	}
	if len(fullLeaves) == 0 {
		t.Errorf("full proof has no EDB leaves")
	}
	for k := range simpleLeaves {
		if !fullLeaves[k] {
			t.Errorf("full mode missing EDB leaf %q", k)
		}
	}
	for k := range fullLeaves {
		if !simpleLeaves[k] {
			t.Errorf("simple mode missing EDB leaf %q", k)
		}
	}
}

func TestFullExplainLetTransform(t *testing.T) {
	source := `
		measurement(1, 10).
		measurement(2, 20).
		scaled(ID, V) :- measurement(ID, Raw) |> let V = fn:mult(Raw, 2).
	`
	_, store, rec := buildEvalRec(t, source)

	// Sanity: the evaluator produced the expected scaled facts.
	if !store.Contains(mustAtom(t, "scaled(1, 20)")) {
		t.Fatalf("scaled(1, 20) not derived")
	}

	goal := mustAtom(t, "scaled(1, 20)")
	proofs, err := BuildFromRecording(rec, store, goal, Options{})
	if err != nil {
		t.Fatalf("BuildFromRecording: %v", err)
	}
	if len(proofs) != 1 {
		t.Fatalf("got %d proofs, want 1", len(proofs))
	}
	p := proofs[0]
	if p.Kind != KindLetRow {
		t.Errorf("expected KindLetRow, got %v", p.Kind)
	}
	if p.TransformText == "" {
		t.Errorf("TransformText empty")
	}
	if len(p.Premises) != 1 {
		t.Fatalf("expected 1 premise, got %d", len(p.Premises))
	}
	if p.Premises[0].Fact.String() != "measurement(1,10)" {
		t.Errorf("premise fact = %q, want %q", p.Premises[0].Fact.String(), "measurement(1,10)")
	}
	if p.Premises[0].Kind != KindEDB {
		t.Errorf("premise kind = %v, want KindEDB", p.Premises[0].Kind)
	}
}

func TestFullExplainAggregation(t *testing.T) {
	source := `
		sale("apple", 3).
		sale("apple", 5).
		sale("pear", 2).
		totals(Product, Sum) :- sale(Product, N)
			|> do fn:group_by(Product), let Sum = fn:sum(N).
	`
	_, store, rec := buildEvalRec(t, source)

	// Sanity: evaluator got the right totals.
	if !store.Contains(mustAtom(t, `totals("apple", 8)`)) {
		t.Fatalf(`totals("apple", 8) not derived`)
	}

	goal := mustAtom(t, `totals("apple", 8)`)
	proofs, err := BuildFromRecording(rec, store, goal, Options{})
	if err != nil {
		t.Fatalf("BuildFromRecording: %v", err)
	}
	if len(proofs) != 1 {
		t.Fatalf("got %d proofs, want 1", len(proofs))
	}
	p := proofs[0]
	if p.Kind != KindDoAggregate {
		t.Fatalf("expected KindDoAggregate, got %v", p.Kind)
	}
	if len(p.GroupKey) != 1 || p.GroupKey[0].String() != `"apple"` {
		t.Errorf("group key = %v, want [\"apple\"]", p.GroupKey)
	}
	if p.TransformText == "" {
		t.Errorf("TransformText empty")
	}
	// The two apple sales should be listed as input facts.
	if len(p.Premises) != 2 {
		t.Fatalf("expected 2 input-fact premises, got %d", len(p.Premises))
	}
	seenFacts := map[string]bool{}
	for _, sub := range p.Premises {
		// Internal facts were emitted by the rewrite. Walk down to find
		// the original "sale(…)" atom in the proof.
		var findSale func(n *ProofNode) string
		findSale = func(n *ProofNode) string {
			if n == nil {
				return ""
			}
			if n.Fact.Predicate.Symbol == "sale" {
				return n.Fact.String()
			}
			for _, c := range n.Premises {
				if s := findSale(c); s != "" {
					return s
				}
			}
			return ""
		}
		if s := findSale(sub); s != "" {
			seenFacts[s] = true
		}
	}
	if !seenFacts[`sale("apple",3)`] || !seenFacts[`sale("apple",5)`] {
		t.Errorf("expected both sale facts in input-fact chain; got %v", seenFacts)
	}
}

func TestFullEmitFactsSchema(t *testing.T) {
	source := `
		sale("apple", 3).
		sale("apple", 5).
		totals(Product, Sum) :- sale(Product, N)
			|> do fn:group_by(Product), let Sum = fn:sum(N).
	`
	_, store, rec := buildEvalRec(t, source)
	proofs, err := BuildFromRecording(rec, store, mustAtom(t, `totals("apple", 8)`), Options{})
	if err != nil {
		t.Fatalf("BuildFromRecording: %v", err)
	}
	out := factstore.NewSimpleInMemoryStore()
	if err := EmitFacts(proofs, &out); err != nil {
		t.Fatalf("EmitFacts: %v", err)
	}
	counts := map[string]int{}
	for _, p := range out.ListPredicates() {
		out.GetFacts(ast.NewQuery(p), func(a ast.Atom) error {
			counts[a.Predicate.Symbol]++
			return nil
		})
	}
	if counts["uses_transform"] == 0 {
		t.Errorf("expected at least one uses_transform fact")
	}
	if counts["group_key"] == 0 {
		t.Errorf("expected at least one group_key fact")
	}
}

func TestFullPrintAggregation(t *testing.T) {
	source := `
		sale("apple", 3).
		sale("apple", 5).
		totals(Product, Sum) :- sale(Product, N)
			|> do fn:group_by(Product), let Sum = fn:sum(N).
	`
	_, store, rec := buildEvalRec(t, source)
	proofs, err := BuildFromRecording(rec, store, mustAtom(t, `totals("apple", 8)`), Options{})
	if err != nil {
		t.Fatalf("BuildFromRecording: %v", err)
	}
	var buf bytes.Buffer
	if err := Print(&buf, proofs); err != nil {
		t.Fatalf("Print: %v", err)
	}
	got := buf.String()
	for _, want := range []string{"input facts:", "group key:", `"apple"`, "transform:"} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in:\n%s", want, got)
		}
	}
}
