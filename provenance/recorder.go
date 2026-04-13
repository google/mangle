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
	"sort"

	"codeberg.org/TauCeti/mangle-go/ast"
	"codeberg.org/TauCeti/mangle-go/engine"
	"codeberg.org/TauCeti/mangle-go/factstore"
	"codeberg.org/TauCeti/mangle-go/unionfind"
)

// Event is one recorded derivation. A given output fact may have multiple
// events (alternative derivations); the builder keeps the first per
// (kind, fact-hash) and discards duplicates so proof IDs stay stable.
type Event struct {
	Kind          EventKind
	Rule          ast.Clause
	Head          ast.Atom
	Subst         unionfind.UnionFind // set for EventRule only
	Row           ast.ConstSubstList  // set for EventLet only
	PremiseFacts  []ast.Atom          // set for EventRule only; one per body atom
	GroupKey      []ast.Constant      // set for EventDo only
	InputFacts    []ast.Atom          // set for EventDo only; one per row in the group
	Output        ast.Atom            // always set
	TransformText string              // set for EventLet and EventDo
}

// EventKind tags the type of a recorded derivation.
type EventKind int

const (
	// EventRule is a plain Datalog rule firing (no transform).
	EventRule EventKind = iota
	// EventLet is a let-transform output (one per input row).
	EventLet
	// EventDo is a do-transform output (one per group).
	EventDo
)

// MemoryRecorder implements [engine.DerivationRecorder] by buffering every
// event in memory. After evaluation, pass the recorder to
// [BuildFromRecording] to build proof trees.
//
// MemoryRecorder is not safe for concurrent use; the engine does not run
// eval in parallel, so this is fine in practice.
type MemoryRecorder struct {
	// byOutputHash maps an output atom's hash to all events concluding it.
	// Multiple entries mean the same fact was derived in more than one
	// way during evaluation.
	byOutputHash map[uint64][]*Event
	all          []*Event
}

// Ensure MemoryRecorder satisfies the engine interface.
var _ engine.DerivationRecorder = (*MemoryRecorder)(nil)

// NewMemoryRecorder returns a fresh, empty recorder.
func NewMemoryRecorder() *MemoryRecorder {
	return &MemoryRecorder{byOutputHash: make(map[uint64][]*Event)}
}

// RuleFired implements engine.DerivationRecorder.
func (r *MemoryRecorder) RuleFired(rule ast.Clause, head ast.Atom, subst unionfind.UnionFind, premiseFacts []ast.Atom) {
	ev := &Event{Kind: EventRule, Rule: rule, Head: head, Subst: subst, PremiseFacts: premiseFacts, Output: head}
	r.add(ev)
}

// LetEmit implements engine.DerivationRecorder.
func (r *MemoryRecorder) LetEmit(rule ast.Clause, head ast.Atom, row ast.ConstSubstList, output ast.Atom) {
	ev := &Event{Kind: EventLet, Rule: rule, Head: head, Row: row, Output: output}
	if rule.Transform != nil {
		ev.TransformText = rule.Transform.String()
	}
	r.add(ev)
}

// DoEmit implements engine.DerivationRecorder.
func (r *MemoryRecorder) DoEmit(rule ast.Clause, head ast.Atom, groupKey []ast.Constant, inputFacts []ast.Atom, output ast.Atom) {
	ev := &Event{
		Kind:       EventDo,
		Rule:       rule,
		Head:       head,
		GroupKey:   append([]ast.Constant(nil), groupKey...),
		InputFacts: append([]ast.Atom(nil), inputFacts...),
		Output:     output,
	}
	if rule.Transform != nil {
		ev.TransformText = rule.Transform.String()
	}
	r.add(ev)
}

func (r *MemoryRecorder) add(ev *Event) {
	r.all = append(r.all, ev)
	h := ev.Output.Hash()
	r.byOutputHash[h] = append(r.byOutputHash[h], ev)
}

// Events returns all recorded events in order of arrival.
func (r *MemoryRecorder) Events() []*Event { return r.all }

// EventsFor returns all events whose output equals the given atom.
// Multiple entries mean alternative derivations of the same fact.
func (r *MemoryRecorder) EventsFor(a ast.Atom) []*Event {
	var out []*Event
	for _, ev := range r.byOutputHash[a.Hash()] {
		if ev.Output.Equals(a) {
			out = append(out, ev)
		}
	}
	return out
}

// BuildFromRecording assembles proof trees for the given goal from a
// recording. It uses the store to distinguish EDB leaves from derived
// facts, and to satisfy negated premises via closed-world absence.
// Options are the same as for [Explain].
func BuildFromRecording(rec *MemoryRecorder, store factstore.ReadOnlyFactStore, goal ast.Atom, opts Options) ([]*ProofNode, error) {
	if opts.MaxProofs == 0 {
		opts.MaxProofs = defaultMaxProofs
	}
	if opts.MaxDepth == 0 {
		opts.MaxDepth = defaultMaxDepth
	}
	if !isGround(goal) {
		return nil, ErrGoalNotGround
	}
	b := &builder{
		rec:     rec,
		store:   store,
		opts:    opts,
		cache:   make(map[uint64][]*ProofNode),
		onStack: make(map[uint64]bool),
		ruleIDs: make(map[string]string),
	}
	proofs := b.build(goal, 0)
	if len(proofs) == 0 {
		return nil, ErrNoProof
	}
	return proofs, nil
}

type builder struct {
	rec     *MemoryRecorder
	store   factstore.ReadOnlyFactStore
	opts    Options
	cache   map[uint64][]*ProofNode
	onStack map[uint64]bool
	ruleIDs map[string]string // rule.String() -> rule content ID
}

func (b *builder) build(goal ast.Atom, depth int) []*ProofNode {
	if depth > b.opts.MaxDepth {
		return []*ProofNode{{Fact: goal, Partial: true, ID: partialID(goal)}}
	}
	h := goal.Hash()
	if cached, ok := b.cache[h]; ok {
		return cached
	}
	if b.onStack[h] {
		return nil
	}
	b.onStack[h] = true
	defer delete(b.onStack, h)

	var proofs []*ProofNode
	events := b.rec.EventsFor(goal)
	if len(events) == 0 {
		// Either a leaf EDB fact or a fact the recorder missed (e.g., an
		// internal predicate that was never re-queried from the store).
		if b.store.Contains(goal) {
			proofs = append(proofs, &ProofNode{ID: edbProofID(goal), Fact: goal, Kind: KindEDB})
		}
		b.cache[h] = proofs
		return proofs
	}
	for _, ev := range events {
		if len(proofs) >= b.opts.MaxProofs {
			break
		}
		p := b.buildFromEvent(ev, depth)
		if p == nil {
			continue
		}
		proofs = append(proofs, p)
	}
	b.cache[h] = proofs
	return proofs
}

func (b *builder) ruleID(r ast.Clause) string {
	key := r.String()
	if id, ok := b.ruleIDs[key]; ok {
		return id
	}
	id := ruleContentID(r)
	b.ruleIDs[key] = id
	return id
}

func (b *builder) buildFromEvent(ev *Event, depth int) *ProofNode {
	ruleID := b.ruleID(ev.Rule)
	switch ev.Kind {
	case EventRule:
		return b.buildRule(ev, ruleID, depth)
	case EventLet:
		return b.buildLet(ev, ruleID, depth)
	case EventDo:
		return b.buildDo(ev, ruleID, depth)
	}
	return nil
}

func (b *builder) buildRule(ev *Event, ruleID string, depth int) *ProofNode {
	// For each positive atom premise, recurse; for non-atom premises
	// (Eq, Ineq, NegAtom), handle as in the simple explainer.
	var premiseProofs []*ProofNode
	partial := false
	for i, p := range ev.Rule.Premises {
		switch term := p.(type) {
		case ast.Atom:
			fact := ev.PremiseFacts[i]
			if fact.Predicate.Symbol == "" {
				// Missing in store — something derived via an unsupported
				// path. Mark partial and continue.
				partial = true
				continue
			}
			sub := b.build(fact, depth+1)
			if len(sub) == 0 {
				partial = true
				continue
			}
			premiseProofs = append(premiseProofs, sub[0])
		case ast.NegAtom:
			// Closed-world absence check.
			// Apply the substitution we have (from the event) and check the store.
			ground, err := applyToNeg(term, ev.Subst)
			if err != nil || !isGround(ground) {
				partial = true
				continue
			}
			if b.store.Contains(ground) {
				// Inconsistency: negation shouldn't have held if the fact is
				// in the store. Treat as partial.
				partial = true
				continue
			}
			premiseProofs = append(premiseProofs, &ProofNode{
				ID:   absenceProofID(ground),
				Fact: ground,
				Kind: KindAbsence,
			})
		case ast.Eq, ast.Ineq:
			// Satisfied by construction (the rule fired). No sub-proof.
		default:
			partial = true
		}
	}
	node := &ProofNode{
		Fact:     ev.Output,
		Kind:     KindDerived,
		Rule:     &ev.Rule,
		RuleID:   ruleID,
		Bindings: extractBindings(ev.Rule, ev.Subst),
		Premises: premiseProofs,
		Partial:  partial,
	}
	node.ID = derivedProofID(ruleID, ev.Output, premiseProofs)
	return node
}

func (b *builder) buildLet(ev *Event, ruleID string, depth int) *ProofNode {
	// A let-transform's only "premise" is the body row that fed it. The
	// recorder captured the row but not its atom-level premises; to show
	// contributing facts we'd need the plain-rule event that produced
	// the row. Resolve each atom in the original body under the row's
	// substitution and look it up in the store.
	var premiseProofs []*ProofNode
	partial := false
	for _, p := range ev.Rule.Premises {
		atom, ok := p.(ast.Atom)
		if !ok {
			continue
		}
		ground := atom.ApplySubst(ev.Row).(ast.Atom)
		if !isGround(ground) {
			partial = true
			continue
		}
		sub := b.build(ground, depth+1)
		if len(sub) == 0 {
			if b.store.Contains(ground) {
				sub = []*ProofNode{{ID: edbProofID(ground), Fact: ground, Kind: KindEDB}}
			} else {
				partial = true
				continue
			}
		}
		premiseProofs = append(premiseProofs, sub[0])
	}
	node := &ProofNode{
		Fact:          ev.Output,
		Kind:          KindLetRow,
		Rule:          &ev.Rule,
		RuleID:        ruleID,
		Premises:      premiseProofs,
		TransformText: ev.TransformText,
		Partial:       partial,
	}
	node.ID = derivedProofID(ruleID, ev.Output, premiseProofs)
	return node
}

func (b *builder) buildDo(ev *Event, ruleID string, depth int) *ProofNode {
	var premiseProofs []*ProofNode
	partial := false
	for _, f := range ev.InputFacts {
		sub := b.build(f, depth+1)
		if len(sub) == 0 {
			if b.store.Contains(f) {
				sub = []*ProofNode{{ID: edbProofID(f), Fact: f, Kind: KindEDB}}
			} else {
				partial = true
				continue
			}
		}
		premiseProofs = append(premiseProofs, sub[0])
	}
	// Deterministic premise order (groups arrive in nondeterministic map order
	// inside evalDo; the *set* of input facts is the same, but the order
	// varies run to run). Sort by proof ID for stable output.
	sort.Slice(premiseProofs, func(i, j int) bool {
		return premiseProofs[i].ID < premiseProofs[j].ID
	})
	node := &ProofNode{
		Fact:          ev.Output,
		Kind:          KindDoAggregate,
		Rule:          &ev.Rule,
		RuleID:        ruleID,
		GroupKey:      ev.GroupKey,
		Premises:      premiseProofs,
		TransformText: ev.TransformText,
		Partial:       partial,
	}
	node.ID = derivedProofID(ruleID, ev.Output, premiseProofs)
	return node
}

// applyToNeg applies a substitution (the rule's solution) to a negated atom.
func applyToNeg(n ast.NegAtom, subst ast.Subst) (ast.Atom, error) {
	if subst == nil {
		return n.Atom, nil
	}
	applied := n.Atom.ApplySubst(subst).(ast.Atom)
	return applied, nil
}
