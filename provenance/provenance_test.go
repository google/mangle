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
	"fmt"
	"strings"
	"testing"

	"codeberg.org/TauCeti/mangle-go/analysis"
	"codeberg.org/TauCeti/mangle-go/ast"
	"codeberg.org/TauCeti/mangle-go/engine"
	"codeberg.org/TauCeti/mangle-go/factstore"
	"codeberg.org/TauCeti/mangle-go/parse"
)

// buildAndEval parses a Mangle source, analyzes + evaluates it, and returns
// the resulting ProgramInfo along with the populated fact store.
func buildAndEval(t *testing.T, source string) (*analysis.ProgramInfo, factstore.SimpleInMemoryStore) {
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
	if _, err := engine.EvalStratifiedProgramWithStats(pi, strata, predToStratum, store); err != nil {
		t.Fatalf("eval: %v", err)
	}
	return pi, store
}

func mustAtom(t *testing.T, s string) ast.Atom {
	t.Helper()
	a, err := parse.Atom(s)
	if err != nil {
		t.Fatalf("parse atom %q: %v", s, err)
	}
	return a
}

func TestExplainSimpleDerivation(t *testing.T) {
	pi, store := buildAndEval(t, `
		bar(1).
		bar(2).
		baz(1).
		foo(X) :- bar(X), baz(X).
	`)
	goal := mustAtom(t, "foo(1)")
	proofs, err := Explain(pi, store, goal, Options{})
	if err != nil {
		t.Fatalf("Explain: %v", err)
	}
	if len(proofs) != 1 {
		t.Fatalf("got %d proofs, want 1", len(proofs))
	}
	p := proofs[0]
	if p.Kind != KindDerived {
		t.Errorf("root kind = %v, want KindDerived", p.Kind)
	}
	if len(p.Premises) != 2 {
		t.Fatalf("got %d premises, want 2", len(p.Premises))
	}
	for i, sub := range p.Premises {
		if sub.Kind != KindEDB {
			t.Errorf("premise[%d] kind = %v, want KindEDB", i, sub.Kind)
		}
	}
	if !strings.HasPrefix(p.ID, "/proof/") || len(p.ID) != len("/proof/")+idHexLen {
		t.Errorf("unexpected proof id %q", p.ID)
	}
}

func TestExplainNoProof(t *testing.T) {
	pi, store := buildAndEval(t, `
		bar(1).
		baz(1).
		foo(X) :- bar(X), baz(X).
	`)
	goal := mustAtom(t, "foo(2)")
	if _, err := Explain(pi, store, goal, Options{}); err != ErrNoProof {
		t.Errorf("want ErrNoProof, got %v", err)
	}
}

func TestExplainNonGroundGoal(t *testing.T) {
	pi, store := buildAndEval(t, `bar(1).`)
	goal := ast.NewAtom("bar", ast.Variable{"X"})
	if _, err := Explain(pi, store, goal, Options{}); err != ErrGoalNotGround {
		t.Errorf("want ErrGoalNotGround, got %v", err)
	}
}

func TestExplainMultipleDerivations(t *testing.T) {
	// foo(1) is derivable by two different rules.
	pi, store := buildAndEval(t, `
		a(1).
		b(1).
		foo(X) :- a(X).
		foo(X) :- b(X).
	`)
	goal := mustAtom(t, "foo(1)")
	proofs, err := Explain(pi, store, goal, Options{MaxProofs: 10})
	if err != nil {
		t.Fatalf("Explain: %v", err)
	}
	if len(proofs) != 2 {
		t.Fatalf("got %d proofs, want 2", len(proofs))
	}
	ids := map[string]bool{proofs[0].ID: true, proofs[1].ID: true}
	if len(ids) != 2 {
		t.Errorf("proofs should have distinct IDs, got %v", ids)
	}
	// Rule IDs should differ.
	if proofs[0].RuleID == proofs[1].RuleID {
		t.Errorf("proofs used the same rule id %q", proofs[0].RuleID)
	}
}

func TestExplainRecursiveClosure(t *testing.T) {
	pi, store := buildAndEval(t, `
		edge(1, 2).
		edge(2, 3).
		edge(3, 4).
		path(X, Y) :- edge(X, Y).
		path(X, Z) :- edge(X, Y), path(Y, Z).
	`)
	goal := mustAtom(t, "path(1, 4)")
	proofs, err := Explain(pi, store, goal, Options{MaxDepth: 20})
	if err != nil {
		t.Fatalf("Explain: %v", err)
	}
	if len(proofs) == 0 {
		t.Fatalf("no proofs for transitive path")
	}
	// Walk down premises: each path(·, ·) proof should terminate in a chain
	// of edge/path that eventually grounds out in edge EDB facts.
	seen := make(map[string]bool)
	var walk func(n *ProofNode) int
	walk = func(n *ProofNode) int {
		if n == nil || seen[n.ID] {
			return 0
		}
		seen[n.ID] = true
		maxChild := 0
		for _, sub := range n.Premises {
			if c := walk(sub); c > maxChild {
				maxChild = c
			}
		}
		return 1 + maxChild
	}
	if got := walk(proofs[0]); got < 3 {
		t.Errorf("recursive proof depth = %d, want >= 3 (edge + path + edge)", got)
	}
}

func TestExplainCycleDetection(t *testing.T) {
	// Mutually recursive rules with a cycle that should be cut.
	// The base case makes p(1) derivable; the cycle must not hang.
	pi, store := buildAndEval(t, `
		base(1).
		p(X) :- base(X).
		p(X) :- q(X).
		q(X) :- p(X).
	`)
	goal := mustAtom(t, "p(1)")
	proofs, err := Explain(pi, store, goal, Options{MaxProofs: 5})
	if err != nil {
		t.Fatalf("Explain: %v", err)
	}
	if len(proofs) == 0 {
		t.Fatalf("expected at least the base-case proof")
	}
}

func TestContentHashDeterminism(t *testing.T) {
	src := `
		bar(1).
		baz(1).
		foo(X) :- bar(X), baz(X).
	`
	goal := mustAtom(t, "foo(1)")

	run := func() string {
		pi, store := buildAndEval(t, src)
		proofs, err := Explain(pi, store, goal, Options{})
		if err != nil {
			t.Fatalf("Explain: %v", err)
		}
		return proofs[0].ID
	}
	id1, id2 := run(), run()
	if id1 != id2 {
		t.Errorf("content-addressed IDs should be stable across runs: %q vs %q", id1, id2)
	}
}

func TestSharedSubProofID(t *testing.T) {
	// bar(1) appears as a premise in two different derivations.
	pi, store := buildAndEval(t, `
		bar(1).
		baz(1).
		qux(1).
		foo1(X) :- bar(X), baz(X).
		foo2(X) :- bar(X), qux(X).
	`)
	p1, err := Explain(pi, store, mustAtom(t, "foo1(1)"), Options{})
	if err != nil {
		t.Fatalf("Explain foo1: %v", err)
	}
	p2, err := Explain(pi, store, mustAtom(t, "foo2(1)"), Options{})
	if err != nil {
		t.Fatalf("Explain foo2: %v", err)
	}
	// Both proofs should reference the same bar(1) EDB leaf ID.
	bar1ID := p1[0].Premises[0].ID
	if bar1ID != p2[0].Premises[0].ID {
		t.Errorf("bar(1) leaf should share ID across proofs, got %q vs %q",
			bar1ID, p2[0].Premises[0].ID)
	}
}

func TestEmitFactsSchema(t *testing.T) {
	pi, store := buildAndEval(t, `
		bar(1).
		baz(1).
		foo(X) :- bar(X), baz(X).
	`)
	goal := mustAtom(t, "foo(1)")
	proofs, err := Explain(pi, store, goal, Options{})
	if err != nil {
		t.Fatalf("Explain: %v", err)
	}
	out := factstore.NewSimpleInMemoryStore()
	if err := EmitFacts(proofs, &out); err != nil {
		t.Fatalf("EmitFacts: %v", err)
	}
	counts := map[string]int{}
	for _, p := range out.ListPredicates() {
		counts[p.Symbol] = 0
	}
	for _, p := range out.ListPredicates() {
		out.GetFacts(ast.NewQuery(p), func(a ast.Atom) error {
			counts[a.Predicate.Symbol]++
			return nil
		})
	}
	// Expect: 1 proves for root + 2 proves for leaves = 3 proves,
	// 1 uses_rule, 1 rule_source, 2 premise, 1 binding (X=1), 2 edb_leaf.
	want := map[string]int{
		"proves":      3,
		"uses_rule":   1,
		"rule_source": 1,
		"premise":     2,
		"binding":     1,
		"edb_leaf":    2,
	}
	for sym, c := range want {
		if counts[sym] != c {
			t.Errorf("predicate %s: got %d facts, want %d", sym, counts[sym], c)
		}
	}
}

func TestEmitFactsAtomEncoding(t *testing.T) {
	pi, store := buildAndEval(t, `bar(1).`)
	goal := mustAtom(t, "bar(1)")
	proofs, err := Explain(pi, store, goal, Options{})
	if err != nil {
		t.Fatalf("Explain: %v", err)
	}
	out := factstore.NewSimpleInMemoryStore()
	if err := EmitFacts(proofs, &out); err != nil {
		t.Fatalf("EmitFacts: %v", err)
	}
	// Find the proves fact and check its second arg is a list [/bar, 1].
	var found bool
	for _, p := range out.ListPredicates() {
		if p.Symbol != "proves" {
			continue
		}
		out.GetFacts(ast.NewQuery(p), func(a ast.Atom) error {
			factList, ok := a.Args[1].(ast.Constant)
			if !ok {
				t.Fatalf("proves arg 1 is not a constant: %v", a.Args[1])
			}
			// String form should be [/bar, 1].
			if got := factList.String(); got != "[/bar, 1]" {
				t.Errorf("list encoding: got %q, want %q", got, "[/bar, 1]")
			}
			found = true
			return nil
		})
	}
	if !found {
		t.Errorf("no proves fact emitted")
	}
}

func TestPrintOutput(t *testing.T) {
	pi, store := buildAndEval(t, `
		bar(1).
		baz(1).
		foo(X) :- bar(X), baz(X).
	`)
	goal := mustAtom(t, "foo(1)")
	proofs, err := Explain(pi, store, goal, Options{})
	if err != nil {
		t.Fatalf("Explain: %v", err)
	}
	var buf bytes.Buffer
	if err := Print(&buf, proofs); err != nil {
		t.Fatalf("Print: %v", err)
	}
	got := buf.String()
	wantFragments := []string{
		"foo(1)",
		"by rule /rule/",
		"with X=1",
		"premises:",
		"bar(1)",
		"baz(1)",
		"[EDB]",
	}
	for _, frag := range wantFragments {
		if !strings.Contains(got, frag) {
			t.Errorf("Print output missing %q in:\n%s", frag, got)
		}
	}
}

// Small fuzz-ish check: MaxProofs caps the result size.
func TestMaxProofs(t *testing.T) {
	pi, store := buildAndEval(t, `
		a(1).
		b(1).
		c(1).
		foo(X) :- a(X).
		foo(X) :- b(X).
		foo(X) :- c(X).
	`)
	goal := mustAtom(t, "foo(1)")
	for _, max := range []int{1, 2, 3} {
		proofs, err := Explain(pi, store, goal, Options{MaxProofs: max})
		if err != nil {
			t.Fatalf("MaxProofs=%d Explain: %v", max, err)
		}
		if len(proofs) != max {
			t.Errorf("MaxProofs=%d: got %d proofs, want %d", max, len(proofs), max)
		}
	}
}

func TestExplainStratifiedNegation(t *testing.T) {
	// unreachable(X) :- node(X), !reachable(X).
	// The proof should include an absence leaf for reachable(3).
	pi, store := buildAndEval(t, `
		node(1).
		node(2).
		node(3).
		reachable(1).
		reachable(2).
		unreachable(X) :- node(X), !reachable(X).
	`)
	proofs, err := Explain(pi, store, mustAtom(t, "unreachable(3)"), Options{})
	if err != nil {
		t.Fatalf("Explain: %v", err)
	}
	if len(proofs) != 1 {
		t.Fatalf("got %d proofs, want 1", len(proofs))
	}
	p := proofs[0]
	if p.Partial {
		t.Errorf("proof marked partial; negation should be fully handled")
	}
	if len(p.Premises) != 2 {
		t.Fatalf("got %d premises, want 2 (node + !reachable)", len(p.Premises))
	}
	// First premise: node(3) — EDB.
	if p.Premises[0].Kind != KindEDB {
		t.Errorf("premise[0] kind = %v, want KindEDB", p.Premises[0].Kind)
	}
	// Second premise: !reachable(3) — absence.
	absNode := p.Premises[1]
	if absNode.Kind != KindAbsence {
		t.Fatalf("premise[1] kind = %v, want KindAbsence", absNode.Kind)
	}
	if absNode.Fact.Predicate.Symbol != "reachable" {
		t.Errorf("absence fact pred = %q, want %q", absNode.Fact.Predicate.Symbol, "reachable")
	}

	// unreachable(1) should NOT be derivable: reachable(1) IS in the store.
	if _, err := Explain(pi, store, mustAtom(t, "unreachable(1)"), Options{}); err != ErrNoProof {
		t.Errorf("unreachable(1) should have no proof (reachable(1) exists), got err=%v", err)
	}
}

func TestEmitFactsAbsenceLeaf(t *testing.T) {
	pi, store := buildAndEval(t, `
		node(1).
		reachable(1).
		node(2).
		unreachable(X) :- node(X), !reachable(X).
	`)
	proofs, err := Explain(pi, store, mustAtom(t, "unreachable(2)"), Options{})
	if err != nil {
		t.Fatalf("Explain: %v", err)
	}
	out := factstore.NewSimpleInMemoryStore()
	if err := EmitFacts(proofs, &out); err != nil {
		t.Fatalf("EmitFacts: %v", err)
	}
	// There should be one absence_leaf fact.
	var absCount, provesForAbsence int
	for _, p := range out.ListPredicates() {
		if p.Symbol == "absence_leaf" {
			out.GetFacts(ast.NewQuery(p), func(a ast.Atom) error {
				absCount++
				return nil
			})
		}
	}
	if absCount != 1 {
		t.Errorf("absence_leaf count = %d, want 1", absCount)
	}
	// The absence leaf should NOT have a proves() fact (it proves ¬Fact, not Fact).
	var absID ast.Constant
	out.GetFacts(ast.NewQuery(ast.PredicateSym{"absence_leaf", 2}), func(a ast.Atom) error {
		absID = a.Args[0].(ast.Constant)
		return nil
	})
	for _, p := range out.ListPredicates() {
		if p.Symbol != "proves" {
			continue
		}
		out.GetFacts(ast.NewQuery(p), func(a ast.Atom) error {
			if a.Args[0].Equals(absID) {
				provesForAbsence++
			}
			return nil
		})
	}
	if provesForAbsence != 0 {
		t.Errorf("absence leaf should not emit proves(); got %d", provesForAbsence)
	}
}

// Sanity: a print dump for manual inspection in -v mode.
func TestExplainShowcase(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	pi, store := buildAndEval(t, `
		edge(1, 2).
		edge(2, 3).
		path(X, Y) :- edge(X, Y).
		path(X, Z) :- edge(X, Y), path(Y, Z).
	`)
	proofs, err := Explain(pi, store, mustAtom(t, "path(1, 3)"), Options{MaxDepth: 10})
	if err != nil {
		t.Fatalf("Explain: %v", err)
	}
	var buf bytes.Buffer
	if err := Print(&buf, proofs); err != nil {
		t.Fatalf("Print: %v", err)
	}
	t.Logf("provenance:\n%s", buf.String())
	_ = fmt.Sprintf // keep fmt import if logging is disabled
}
