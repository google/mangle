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

package engine

import (
	"codeberg.org/TauCeti/mangle-go/ast"
	"codeberg.org/TauCeti/mangle-go/unionfind"
)

// DerivationRecorder receives a callback every time a rule produces an
// output fact. Implementations are responsible for buffering/persisting
// events; the engine does not retain them after the callback returns.
//
// The recorder is an optional evaluation hook used by the provenance
// package to capture "full" provenance during computation. When no
// recorder is configured, none of the callbacks fire — there is no
// overhead in the hot path beyond a nil check per derivation.
type DerivationRecorder interface {
	// RuleFired is invoked for each solution of a plain Datalog rule
	// (including rules whose only transform is a let-transform; see
	// LetEmit for the corresponding post-transform output).
	//
	// rule is the originating clause, head is the ground derived atom,
	// subst is the substitution that satisfied the body, and
	// premiseFacts gives the concrete atoms (one per atom-shaped premise)
	// that the substitution selected. Premises that are not atoms (Eq,
	// Ineq, NegAtom, temporal) appear in premiseFacts as zero-valued
	// atoms at the corresponding position.
	RuleFired(rule ast.Clause, head ast.Atom, subst unionfind.UnionFind, premiseFacts []ast.Atom)

	// LetEmit is invoked once per output of a let-transform, i.e. once
	// per input solution row. output is the fact emitted after the
	// let-statements apply.
	LetEmit(rule ast.Clause, head ast.Atom, row ast.ConstSubstList, output ast.Atom)

	// DoEmit is invoked once per output of a do-transform (one per
	// group). groupKey carries the values of the group-by variables,
	// inputFacts gives the underlying facts that fed this group (one per
	// input row), and output is the aggregated fact.
	DoEmit(rule ast.Clause, head ast.Atom, groupKey []ast.Constant, inputFacts []ast.Atom, output ast.Atom)
}

// WithDerivationRecorder installs a DerivationRecorder. Pass nil to
// disable (the default). When set, the engine calls the recorder's
// methods at each derivation point; when unset, the callbacks are
// skipped with a single nil check.
func WithDerivationRecorder(r DerivationRecorder) EvalOption {
	return func(o *EvalOptions) { o.recorder = r }
}
