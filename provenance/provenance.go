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

// Package provenance computes proof trees for derived (IDB) facts.
//
// Two modes are supported:
//
//   - "Simple provenance" ([Explain]) works post-hoc: given a program and
//     the final fact store, it backward-chains to assemble proofs. It
//     handles positive Datalog, equality/inequality builtins, stratified
//     negation via closed-world absence, and recursion with cycle
//     detection. Rules that use aggregation or other transforms are
//     skipped in this mode because post-hoc reconstruction of grouping is
//     not reliable.
//
//   - "Full provenance" ([BuildFromRecording]) uses a [MemoryRecorder]
//     installed via [engine.WithDerivationRecorder] to capture derivation
//     events *during* evaluation. It handles the same features as simple
//     provenance, plus let-transforms (row-wise) and do-transforms
//     (aggregation). When a recorder is attached, the engine fires
//     callbacks at every derivation point so the full DAG, including
//     group-feeding facts, is available to [BuildFromRecording] without
//     re-querying the store.
//
// Both modes produce the same [ProofNode] shape and are emitted through
// the same [EmitFacts] / [Print] helpers.
//
// Proof nodes are identified by a 128-bit content hash computed from the
// rule, the derived fact, and the (sorted) identifiers of sub-proofs. That
// makes the proof graph a DAG where identical sub-derivations are shared.
package provenance

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"

	"codeberg.org/TauCeti/mangle-go/analysis"
	"codeberg.org/TauCeti/mangle-go/ast"
	"codeberg.org/TauCeti/mangle-go/factstore"
	"codeberg.org/TauCeti/mangle-go/functional"
	"codeberg.org/TauCeti/mangle-go/unionfind"
)

// Kind distinguishes EDB leaves from derivations.
type Kind int

const (
	// KindEDB is a leaf proof: the fact is stored, not derived.
	KindEDB Kind = iota
	// KindDerived is an interior proof: the fact was derived by a rule.
	KindDerived
	// KindAbsence is a leaf proof of negation: the fact is NOT in the store.
	// Sound under stratified evaluation, which Mangle uses: by the time a
	// negated premise is checked, the stratum defining its predicate has
	// already run to fixpoint.
	KindAbsence
	// KindLetRow is an output of a row-wise let-transform. Its single
	// premise is the conjunctive body row (recorded as a KindDerived
	// node with the body atoms as sub-proofs).
	KindLetRow
	// KindDoAggregate is an output of a do-transform. Its premises are
	// the sub-proofs of the facts that fed this group. The group's
	// key values are stored in GroupKey.
	KindDoAggregate
)

// Binding records one entry of a rule's substitution σ.
type Binding struct {
	Var   ast.Variable
	Value ast.Constant
}

// ProofNode is a node in the proof DAG. Shared sub-proofs point to the same node.
type ProofNode struct {
	// ID is a stable, content-addressed identifier of the form "/proof/<hex>".
	ID string
	// Fact is the ground atom this node proves.
	Fact ast.Atom
	// Kind is KindEDB for leaves or KindDerived for rule applications.
	Kind Kind
	// Rule is the clause that fired; nil when Kind == KindEDB.
	Rule *ast.Clause
	// RuleID is the content-addressed rule identifier "/rule/<hex>"; empty when Kind == KindEDB.
	RuleID string
	// Bindings are the rule's substitution entries; empty when Kind == KindEDB.
	Bindings []Binding
	// Premises are the sub-proofs matching the rule's body atoms, in body order.
	Premises []*ProofNode
	// GroupKey is set on KindDoAggregate nodes; it holds the group-by
	// values that define the group this aggregate was computed over.
	GroupKey []ast.Constant
	// TransformText is set on KindLetRow and KindDoAggregate nodes to the
	// rule's transform clause pretty-printed (e.g. "do fn:group_by(X), let S = fn:sum(Y)").
	TransformText string
	// Partial is true if the proof could not be fully expanded (e.g. rule used
	// a negated premise, a transform, or exceeded MaxDepth).
	Partial bool
}

// Options tunes the explainer.
type Options struct {
	// MaxProofs caps the number of alternative proofs returned for a goal.
	// Defaults to 1 (shortest proof only) when zero.
	MaxProofs int
	// MaxDepth caps proof-tree depth. Defaults to 64 when zero. Exceeding the
	// depth marks sub-proofs as Partial rather than failing.
	MaxDepth int
}

const (
	defaultMaxProofs = 1
	defaultMaxDepth  = 64
)

// ErrNoProof indicates that no proof was found for the goal. The goal may
// not be in the store, or its derivation uses features outside simple
// provenance (aggregation, transforms, temporal annotations).
var ErrNoProof = errors.New("provenance: no proof found")

// ErrGoalNotGround indicates the caller supplied a goal with free variables.
// Callers should materialize goals by querying the store first.
var ErrGoalNotGround = errors.New("provenance: goal must be ground")

// Explain returns proof trees for the given ground goal. Pass a non-nil
// [Options] to cap alternative proofs or proof depth.
//
// If the goal predicate is EDB and the fact is stored, a single KindEDB
// leaf is returned. If the predicate is IDB, each returned tree corresponds
// to one successful rule application.
func Explain(program *analysis.ProgramInfo, store factstore.ReadOnlyFactStore, goal ast.Atom, opts Options) ([]*ProofNode, error) {
	if opts.MaxProofs == 0 {
		opts.MaxProofs = defaultMaxProofs
	}
	if opts.MaxDepth == 0 {
		opts.MaxDepth = defaultMaxDepth
	}
	if !isGround(goal) {
		return nil, ErrGoalNotGround
	}
	e := &explainer{
		program: program,
		store:   store,
		opts:    opts,
		cache:   make(map[uint64][]*ProofNode),
		onStack: make(map[uint64]bool),
		ruleIDs: make(map[int]string),
	}
	proofs := e.explain(goal, 0)
	if len(proofs) == 0 {
		return nil, ErrNoProof
	}
	return proofs, nil
}

type explainer struct {
	program *analysis.ProgramInfo
	store   factstore.ReadOnlyFactStore
	opts    Options
	// cache memoizes proofs per ground goal hash. Avoids recomputing proofs
	// of facts that appear as premises in multiple parent proofs.
	cache map[uint64][]*ProofNode
	// onStack tracks goals currently being proved to break cycles.
	onStack map[uint64]bool
	// ruleIDs memoizes content-addressed rule IDs keyed by index in program.Rules.
	ruleIDs map[int]string
}

func (e *explainer) explain(goal ast.Atom, depth int) []*ProofNode {
	if depth > e.opts.MaxDepth {
		return []*ProofNode{{Fact: goal, Partial: true, ID: partialID(goal)}}
	}
	h := goal.Hash()
	if cached, ok := e.cache[h]; ok {
		return cached
	}
	if e.onStack[h] {
		return nil
	}
	e.onStack[h] = true
	defer delete(e.onStack, h)

	var proofs []*ProofNode

	if e.isEDB(goal.Predicate) && e.store.Contains(goal) {
		proofs = append(proofs, &ProofNode{
			ID:   edbProofID(goal),
			Fact: goal,
			Kind: KindEDB,
		})
	}

	for ruleIdx, rule := range e.program.Rules {
		if len(proofs) >= e.opts.MaxProofs {
			break
		}
		if rule.Head.Predicate != goal.Predicate {
			continue
		}
		if rule.Transform != nil {
			continue
		}
		rulePremises := rule.Premises
		uf := unionfind.New()
		headUF, err := unionfind.UnifyTermsExtend(rule.Head.Args, baseTermsFrom(goal), uf)
		if err != nil {
			continue
		}
		remaining := e.opts.MaxProofs - len(proofs)
		for _, sol := range e.solveBody(rulePremises, headUF, depth, remaining) {
			if len(proofs) >= e.opts.MaxProofs {
				break
			}
			proof, ok := e.buildProof(&e.program.Rules[ruleIdx], ruleIdx, rule, goal, sol, depth)
			if !ok {
				continue
			}
			proofs = append(proofs, proof)
		}
	}

	e.cache[h] = proofs
	return proofs
}

// bodySolution carries a successful body unifier plus the ground premise
// atoms and their sub-proofs.
type bodySolution struct {
	subst        unionfind.UnionFind
	premiseAtoms []ast.Atom
	subProofs    []*ProofNode
	partial      bool
}

func (e *explainer) solveBody(premises []ast.Term, uf unionfind.UnionFind, depth, need int) []bodySolution {
	return e.solveBodyRec(premises, uf, depth, need, nil, nil, false)
}

func (e *explainer) solveBodyRec(premises []ast.Term, uf unionfind.UnionFind, depth, need int, accAtoms []ast.Atom, accProofs []*ProofNode, partial bool) []bodySolution {
	if len(premises) == 0 {
		return []bodySolution{{subst: uf, premiseAtoms: accAtoms, subProofs: accProofs, partial: partial}}
	}
	if need <= 0 {
		return nil
	}
	first, rest := premises[0], premises[1:]
	switch p := first.(type) {
	case ast.Atom:
		return e.solveAtomPremise(p, rest, uf, depth, need, accAtoms, accProofs, partial)
	case ast.Eq:
		ok, err := evalEq(p.Left, p.Right, uf, true)
		if err != nil || !ok {
			return nil
		}
		return e.solveBodyRec(rest, uf, depth, need, accAtoms, accProofs, partial)
	case ast.Ineq:
		ok, err := evalEq(p.Left, p.Right, uf, false)
		if err != nil || !ok {
			return nil
		}
		return e.solveBodyRec(rest, uf, depth, need, accAtoms, accProofs, partial)
	case ast.NegAtom:
		ground, err := functional.EvalAtom(p.Atom, uf)
		if err != nil || !isGround(ground) {
			// Non-ground negation slipped past safety checks; mark partial.
			return e.solveBodyRec(rest, uf, depth, need, accAtoms, accProofs, true)
		}
		if e.store.Contains(ground) {
			// Negated premise fails: the atom IS in the store.
			return nil
		}
		leaf := &ProofNode{
			ID:   absenceProofID(ground),
			Fact: ground,
			Kind: KindAbsence,
		}
		newAtoms := append(append([]ast.Atom(nil), accAtoms...), ground)
		newProofs := append(append([]*ProofNode(nil), accProofs...), leaf)
		return e.solveBodyRec(rest, uf, depth, need, newAtoms, newProofs, partial)
	default:
		// Temporal premises and any other exotic term — mark partial but
		// continue so the user sees an annotated result rather than nothing.
		return e.solveBodyRec(rest, uf, depth, need, accAtoms, accProofs, true)
	}
}

func (e *explainer) solveAtomPremise(pAtom ast.Atom, rest []ast.Term, uf unionfind.UnionFind, depth, need int, accAtoms []ast.Atom, accProofs []*ProofNode, partial bool) []bodySolution {
	pattern, err := functional.EvalAtom(pAtom, uf)
	if err != nil {
		return nil
	}
	var results []bodySolution
	_ = e.store.GetFacts(pattern, func(fact ast.Atom) error {
		if len(results) >= need {
			return nil
		}
		extended, err := unionfind.UnifyTermsExtend(pattern.Args, fact.Args, uf)
		if err != nil {
			return nil
		}
		subProofs := e.explain(fact, depth+1)
		if len(subProofs) == 0 {
			return nil
		}
		sub := subProofs[0]
		newAtoms := append(append([]ast.Atom(nil), accAtoms...), fact)
		newProofs := append(append([]*ProofNode(nil), accProofs...), sub)
		newPartial := partial || sub.Partial
		tail := e.solveBodyRec(rest, extended, depth, need-len(results), newAtoms, newProofs, newPartial)
		results = append(results, tail...)
		return nil
	})
	return results
}

func (e *explainer) buildProof(ruleRef *ast.Clause, ruleIdx int, rule ast.Clause, goal ast.Atom, sol bodySolution, depth int) (*ProofNode, bool) {
	ruleID, ok := e.ruleIDs[ruleIdx]
	if !ok {
		ruleID = ruleContentID(rule)
		e.ruleIDs[ruleIdx] = ruleID
	}
	bindings := extractBindings(rule, sol.subst)
	node := &ProofNode{
		Fact:     goal,
		Kind:     KindDerived,
		Rule:     ruleRef,
		RuleID:   ruleID,
		Bindings: bindings,
		Premises: sol.subProofs,
		Partial:  sol.partial,
	}
	node.ID = derivedProofID(ruleID, goal, sol.subProofs)
	return node, true
}

func (e *explainer) isEDB(p ast.PredicateSym) bool {
	if e.program == nil {
		return false
	}
	_, ok := e.program.EdbPredicates[p]
	return ok
}

// --- helpers ---

func isGround(a ast.Atom) bool {
	for _, arg := range a.Args {
		if _, ok := arg.(ast.Constant); !ok {
			return false
		}
	}
	return true
}

func baseTermsFrom(a ast.Atom) []ast.BaseTerm {
	out := make([]ast.BaseTerm, len(a.Args))
	copy(out, a.Args)
	return out
}

// extractBindings returns the rule's variables bound to concrete constants
// under the final substitution. Variables that remain unbound are omitted.
func extractBindings(rule ast.Clause, uf unionfind.UnionFind) []Binding {
	vars := collectVars(rule)
	out := make([]Binding, 0, len(vars))
	for _, v := range vars {
		bound := uf.Get(v)
		c, ok := bound.(ast.Constant)
		if !ok {
			continue
		}
		out = append(out, Binding{Var: v, Value: c})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Var.Symbol < out[j].Var.Symbol })
	return out
}

func collectVars(rule ast.Clause) []ast.Variable {
	seen := make(map[string]ast.Variable)
	addFromAtom := func(a ast.Atom) {
		for _, arg := range a.Args {
			if v, ok := arg.(ast.Variable); ok && v.Symbol != "_" {
				seen[v.Symbol] = v
			}
		}
	}
	addFromAtom(rule.Head)
	for _, p := range rule.Premises {
		if a, ok := p.(ast.Atom); ok {
			addFromAtom(a)
		}
	}
	out := make([]ast.Variable, 0, len(seen))
	for _, v := range seen {
		out = append(out, v)
	}
	return out
}

func evalEq(l, r ast.BaseTerm, uf unionfind.UnionFind, wantEqual bool) (bool, error) {
	lv, err := functional.EvalExpr(l, uf)
	if err != nil {
		return false, err
	}
	rv, err := functional.EvalExpr(r, uf)
	if err != nil {
		return false, err
	}
	eq := lv.Equals(rv)
	if wantEqual {
		return eq, nil
	}
	return !eq, nil
}

// --- content-addressed IDs (128 bits, hex-encoded) ---

const idHexLen = 32 // 128 bits

func contentHashHex(parts ...string) string {
	h := sha256.New()
	for _, p := range parts {
		fmt.Fprintf(h, "%d:%s\n", len(p), p)
	}
	sum := h.Sum(nil)
	return hex.EncodeToString(sum[:16])
}

func edbProofID(a ast.Atom) string {
	return "/proof/" + contentHashHex("edb", a.String())
}

func absenceProofID(a ast.Atom) string {
	return "/proof/" + contentHashHex("absence", a.String())
}

func derivedProofID(ruleID string, goal ast.Atom, sub []*ProofNode) string {
	ids := make([]string, len(sub))
	for i, s := range sub {
		ids[i] = s.ID
	}
	parts := append([]string{"derived", ruleID, goal.String()}, ids...)
	return "/proof/" + contentHashHex(parts...)
}

func ruleContentID(r ast.Clause) string {
	return "/rule/" + contentHashHex("rule", r.String())
}

func partialID(a ast.Atom) string {
	return "/proof/partial/" + contentHashHex("partial", a.String())
}
