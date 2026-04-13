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

package engine

import (
	"fmt"
	"strings"

	"codeberg.org/TauCeti/mangle-go/ast"
	"codeberg.org/TauCeti/mangle-go/builtin"
	"codeberg.org/TauCeti/mangle-go/functional"
	"codeberg.org/TauCeti/mangle-go/symbols"
)

// TransformKind identifies which kind of transform produced an output fact
// in a [TransformEmit] callback.
type TransformKind int

const (
	// TransformKindLet indicates the emission came from a let-transform.
	TransformKindLet TransformKind = iota
	// TransformKindDo indicates the emission came from a do-transform.
	TransformKindDo
)

// TransformEmit is the richer emission callback. For let-transforms, kind
// is TransformKindLet and groupKey and inputFacts are nil. For do-transforms,
// kind is TransformKindDo, groupKey holds the group-by values, and
// inputFacts (if non-nil) lists the facts that fed this group in input order.
type TransformEmit func(out ast.Atom, kind TransformKind, groupKey []ast.Constant, inputFacts []ast.Atom) bool

// EvalTransform evaluates a transform, emitting one output fact per
// derivation. This is the narrow public entry point; see also
// [EvalTransformWithInputFacts] for callers that need per-emission context
// (group keys, contributing input facts).
func EvalTransform(
	head ast.Atom,
	transform ast.Transform,
	input []ast.ConstSubstList,
	emit func(atom ast.Atom) bool) error {

	return EvalTransformWithInputFacts(head, transform, input, nil,
		func(out ast.Atom, _ TransformKind, _ []ast.Constant, _ []ast.Atom) bool {
			return emit(out)
		})
}

// EvalTransformWithInputFacts evaluates a transform with a richer emit
// callback. If inputFacts is non-nil it must have len(input) entries:
// inputFacts[i] is the fact that produced row input[i]. Used by the
// provenance recorder to reconstruct which facts fed each aggregate group.
func EvalTransformWithInputFacts(
	head ast.Atom,
	transform ast.Transform,
	input []ast.ConstSubstList,
	inputFacts []ast.Atom,
	emit TransformEmit) error {
	if transform.IsLetTransform() {
		return evalLet(head, transform, input, emit)
	}
	return evalDo(head, transform, input, inputFacts, emit)
}

// evalLet evaluates a let transform. This consists of a number of let statements,
// each acting on a single row.
func evalLet(
	head ast.Atom,
	transform ast.Transform,
	rows []ast.ConstSubstList,
	emit TransformEmit) error {
	for _, init := range rows {
		subst := init
		for _, stmt := range transform.Statements {
			con, err := functional.EvalApplyFn(stmt.Fn, subst)
			if err != nil {
				return err
			}
			subst = subst.Extend(*stmt.Var, con)
		}
		emit(head.ApplySubst(subst).(ast.Atom), TransformKindLet, nil, nil)
	}
	return nil
}

type grouped struct {
	key     []ast.Constant
	values  []ast.ConstSubstList
	indices []int // positions in the original input (for input-fact lookup)
}

// groupKeyString builds a collision-free string key for a group-by key.
// Each component's String() form is length-prefixed and terminated so the
// encoding is injective for any sequence of constants.
func groupKeyString(keys []ast.Constant) string {
	var sb strings.Builder
	for _, k := range keys {
		s := k.String()
		fmt.Fprintf(&sb, "%d:%s|", len(s), s)
	}
	return sb.String()
}

// evalDo evaluates a do statement (currently: only "|> do fn::group_by")
// A do-transform acts on a whole result relation.
func evalDo(
	head ast.Atom,
	transform ast.Transform,
	input []ast.ConstSubstList,
	inputFacts []ast.Atom,
	emit TransformEmit) error {

	doStmt := transform.Statements[0]
	switch doStmt.Fn.Function.Symbol {
	case symbols.GroupBy.Symbol:
		keyToGroup := make(map[string]grouped)

		for idx, subst := range input {
			keyLen := len(doStmt.Fn.Args)
			key := make([]ast.Constant, keyLen)
			for i, v := range doStmt.Fn.Args {
				key[i] = subst.Get(v.(ast.Variable)).(ast.Constant)
			}
			h := groupKeyString(key)
			group, ok := keyToGroup[h]
			if !ok {
				group = grouped{key, nil, nil}
			}
			group.values = append(group.values, subst)
			group.indices = append(group.indices, idx)
			keyToGroup[h] = group
		}
		// Now apply reductions.
		for _, group := range keyToGroup {
			var subst ast.ConstSubstList
			for i, v := range group.key {
				subst = subst.Extend(doStmt.Fn.Args[i].(ast.Variable), v)
			}
			for _, stmt := range transform.Statements[1:] {
				if builtin.IsReducerFunction(stmt.Fn.Function) {
					con, err := functional.EvalReduceFn(stmt.Fn, group.values)
					if err != nil {
						return err
					}
					subst = subst.Extend(*stmt.Var, con)
				} else {
					// Do a "map" of rows (substitutions set).
					// First, copy over bindings from subst
					con, err := functional.EvalApplyFn(stmt.Fn, subst)
					if err != nil {
						return err
					}
					subst = subst.Extend(*stmt.Var, con)
				}
			}
			var groupFacts []ast.Atom
			if inputFacts != nil {
				groupFacts = make([]ast.Atom, len(group.indices))
				for i, idx := range group.indices {
					groupFacts[i] = inputFacts[idx]
				}
			}
			emit(head.ApplySubst(subst).(ast.Atom), TransformKindDo, group.key, groupFacts)
		}
	default:
	}
	return nil
}
