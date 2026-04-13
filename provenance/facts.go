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
	"fmt"
	"io"
	"strings"

	"codeberg.org/TauCeti/mangle-go/ast"
	"codeberg.org/TauCeti/mangle-go/factstore"
)

// The schema written by [EmitFacts]:
//
//	proves(ProofID, Fact)              — this node concludes Fact
//	uses_rule(ProofID, RuleID)         — which rule fired (omitted for leaves)
//	premise(ProofID, Index, SubID)     — the Index-th premise was proved by SubID
//	edb_leaf(ProofID, Fact)            — terminal: Fact comes from the store
//	absence_leaf(ProofID, Fact)        — terminal: Fact is NOT in the store
//	                                     (closed-world proof of a negated premise)
//	binding(ProofID, VarName, Value)   — substitution entries
//	rule_source(RuleID, ClauseText)    — human-readable form of a rule
//	uses_transform(ProofID, TransformText) — set for let/do-transform nodes
//	group_key(ProofID, Index, Value)   — per-position group-by values for do-aggregates
//
// Note: absence leaves do not emit a proves(...) fact — semantically they
// "prove" the negation of Fact, not Fact itself. Walk them via premise(...).
//
// Facts are encoded uniformly as list constants: foo(a, b) becomes
// [/foo, /a, /b]. Variable names and rule-source text are String constants.

// EmitFacts walks the proof DAG (visiting each unique ProofNode once) and
// writes the schema predicates into out.
func EmitFacts(proofs []*ProofNode, out factstore.FactStore) error {
	seen := make(map[string]bool)
	var walk func(n *ProofNode) error
	walk = func(n *ProofNode) error {
		if n == nil || seen[n.ID] {
			return nil
		}
		seen[n.ID] = true

		pid, err := ast.Name(n.ID)
		if err != nil {
			return err
		}
		factList := atomAsList(n.Fact)

		switch n.Kind {
		case KindEDB:
			out.Add(ast.NewAtom("edb_leaf", pid, factList))
			out.Add(ast.NewAtom("proves", pid, factList))
		case KindAbsence:
			out.Add(ast.NewAtom("absence_leaf", pid, factList))
		case KindDerived, KindLetRow, KindDoAggregate:
			out.Add(ast.NewAtom("proves", pid, factList))
			rid, err := ast.Name(n.RuleID)
			if err != nil {
				return err
			}
			out.Add(ast.NewAtom("uses_rule", pid, rid))
			if n.Rule != nil {
				out.Add(ast.NewAtom("rule_source", rid, ast.String(n.Rule.String())))
			}
			for _, b := range n.Bindings {
				out.Add(ast.NewAtom("binding", pid, ast.String(b.Var.Symbol), b.Value))
			}
			if n.TransformText != "" {
				out.Add(ast.NewAtom("uses_transform", pid, ast.String(n.TransformText)))
			}
			for i, v := range n.GroupKey {
				out.Add(ast.NewAtom("group_key", pid, ast.Number(int64(i)), v))
			}
			for i, sub := range n.Premises {
				if sub == nil {
					continue
				}
				subID, err := ast.Name(sub.ID)
				if err != nil {
					return err
				}
				out.Add(ast.NewAtom("premise", pid, ast.Number(int64(i)), subID))
				if err := walk(sub); err != nil {
					return err
				}
			}
		}
		return nil
	}
	for _, p := range proofs {
		if err := walk(p); err != nil {
			return err
		}
	}
	return nil
}

// atomAsList encodes foo(a, b) as [/foo, a, b]. The predicate symbol becomes
// a name constant by prepending "/". Arguments must already be constants
// (the atom is expected to be ground).
func atomAsList(a ast.Atom) ast.Constant {
	predSym := a.Predicate.Symbol
	if !strings.HasPrefix(predSym, "/") {
		predSym = "/" + predSym
	}
	predName, err := ast.Name(predSym)
	if err != nil {
		predName = ast.String(a.Predicate.Symbol)
	}
	items := make([]ast.Constant, 0, len(a.Args)+1)
	items = append(items, predName)
	for _, arg := range a.Args {
		if c, ok := arg.(ast.Constant); ok {
			items = append(items, c)
		} else {
			items = append(items, ast.String(arg.String()))
		}
	}
	return ast.List(items)
}

// Print writes a human-readable proof tree to w. Shared sub-proofs are
// printed once and then referenced by ID ("… see /proof/abc").
func Print(w io.Writer, proofs []*ProofNode) error {
	printed := make(map[string]bool)
	for i, p := range proofs {
		if i > 0 {
			if _, err := fmt.Fprintln(w); err != nil {
				return err
			}
		}
		if _, err := fmt.Fprintf(w, "proof %d of %d:\n", i+1, len(proofs)); err != nil {
			return err
		}
		if err := printNode(w, p, "  ", printed); err != nil {
			return err
		}
	}
	return nil
}

func printNode(w io.Writer, n *ProofNode, indent string, printed map[string]bool) error {
	if n == nil {
		return nil
	}
	tag := ""
	if n.Partial {
		tag = " [partial]"
	}
	if _, err := fmt.Fprintf(w, "%s%s  (%s)%s\n", indent, n.Fact.String(), n.ID, tag); err != nil {
		return err
	}
	if printed[n.ID] {
		if _, err := fmt.Fprintf(w, "%s  ... see %s\n", indent, n.ID); err != nil {
			return err
		}
		return nil
	}
	printed[n.ID] = true

	switch n.Kind {
	case KindEDB:
		if _, err := fmt.Fprintf(w, "%s  [EDB]\n", indent); err != nil {
			return err
		}
	case KindAbsence:
		if _, err := fmt.Fprintf(w, "%s  [absent: !%s]\n", indent, n.Fact.String()); err != nil {
			return err
		}
	case KindLetRow, KindDoAggregate, KindDerived:
		ruleStr := ""
		if n.Rule != nil {
			ruleStr = n.Rule.String()
		}
		if _, err := fmt.Fprintf(w, "%s  by rule %s:  %s\n", indent, n.RuleID, ruleStr); err != nil {
			return err
		}
		if len(n.Bindings) > 0 {
			var parts []string
			for _, b := range n.Bindings {
				parts = append(parts, fmt.Sprintf("%s=%s", b.Var.Symbol, b.Value.String()))
			}
			if _, err := fmt.Fprintf(w, "%s  with %s\n", indent, strings.Join(parts, ", ")); err != nil {
				return err
			}
		}
		if n.TransformText != "" {
			if _, err := fmt.Fprintf(w, "%s  transform: %s\n", indent, n.TransformText); err != nil {
				return err
			}
		}
		if len(n.GroupKey) > 0 {
			var parts []string
			for _, v := range n.GroupKey {
				parts = append(parts, v.String())
			}
			if _, err := fmt.Fprintf(w, "%s  group key: (%s)\n", indent, strings.Join(parts, ", ")); err != nil {
				return err
			}
		}
		if len(n.Premises) > 0 {
			label := "premises"
			if n.Kind == KindDoAggregate {
				label = "input facts"
			}
			if _, err := fmt.Fprintf(w, "%s  %s:\n", indent, label); err != nil {
				return err
			}
			for _, sub := range n.Premises {
				if err := printNode(w, sub, indent+"    ", printed); err != nil {
					return err
				}
			}
		}
	}
	return nil
}
