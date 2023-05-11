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
	"encoding/binary"

	"github.com/google/mangle/ast"
	"github.com/google/mangle/functional"
	"github.com/google/mangle/symbols"

	"hash/fnv"
)

// EvalTransform evaluates a transform.
func EvalTransform(
	head ast.Atom,
	transform ast.Transform,
	input []ast.ConstSubstList,
	emit func(atom ast.Atom) bool) error {

	if transform.IsLetTransform() {
		return evalLet(head, transform, input, emit)
	}
	return evalDo(head, transform, input, emit)
}

// evalLet evaluates a let transform. This consists of a number of let statements,
// each acting on a single row.
func evalLet(
	head ast.Atom,
	transform ast.Transform,
	rows []ast.ConstSubstList,
	emit func(atom ast.Atom) bool) error {
	for _, init := range rows {
		subst := init
		for _, stmt := range transform.Statements {
			con, err := functional.EvalApplyFn(stmt.Fn, subst)
			if err != nil {
				return err
			}
			subst = subst.Extend(*stmt.Var, con)
		}
		emit(head.ApplySubst(subst).(ast.Atom))
	}
	return nil
}

type grouped struct {
	key    []ast.Constant
	values []ast.ConstSubstList
}

// evalDo evaluates a do statement (currently: only "|> do fn::group_by")
// A do-transform acts on a whole result relation.
func evalDo(
	head ast.Atom,
	transform ast.Transform,
	input []ast.ConstSubstList,
	emit func(atom ast.Atom) bool) error {

	doStmt := transform.Statements[0]
	switch doStmt.Fn.Function.Symbol {
	case symbols.GroupBy.Symbol:
		keyToGroup := make(map[uint32]grouped)

		for _, subst := range input {
			keyLen := len(doStmt.Fn.Args)
			key := make([]ast.Constant, keyLen)
			b := make([]byte, 8*keyLen)
			for i, v := range doStmt.Fn.Args {
				value := subst.Get(v.(ast.Variable)).(ast.Constant)
				binary.LittleEndian.PutUint64(b[i*8:], value.Hash())
				key[i] = value
			}
			hasher := fnv.New32()
			hasher.Write(b)
			h := hasher.Sum32()
			group, ok := keyToGroup[h]
			if !ok {
				group = grouped{key, nil}
			}
			group.values = append(group.values, subst)
			keyToGroup[h] = group
		}
		// Now apply reductions.
		for _, group := range keyToGroup {
			var subst ast.ConstSubstList
			for i, v := range group.key {
				subst = subst.Extend(doStmt.Fn.Args[i].(ast.Variable), v)
			}
			for _, stmt := range transform.Statements[1:] {
				con, err := functional.EvalReduceFn(stmt.Fn, group.values)
				if err != nil {
					return err
				}
				subst = subst.Extend(*stmt.Var, con)
			}
			emit(head.ApplySubst(subst).(ast.Atom))
		}
	default:
	}
	return nil
}
