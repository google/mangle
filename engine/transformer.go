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
	"io/ioutil"
	"os"
	"os/exec"
	"plugin"

	"github.com/google/mangle/ast"
	"github.com/google/mangle/functional"
	"github.com/google/mangle/symbols"

	"hash/fnv"
)

// evalScript evaluates a script transform.
func evalScript(
	head ast.Atom,
	transform ast.Transform,
	input []ast.ConstSubstList,
	emit func(atom ast.Atom) bool) error {

	// Create a temporary file for the script.
	tmpfile, err := ioutil.TempFile("", "mangle-script-*.go")
	if err != nil {
		return err
	}
	defer os.Remove(tmpfile.Name())

	// Write the script to the temporary file.
	if _, err := tmpfile.WriteString(transform.Script); err != nil {
		return err
	}
	if err := tmpfile.Close(); err != nil {
		return err
	}

	// Compile the script as a plugin.
	pluginPath := tmpfile.Name() + ".so"
	cmd := exec.Command("go", "build", "-buildmode=plugin", "-o", pluginPath, tmpfile.Name())
	if err := cmd.Run(); err != nil {
		return err
	}
	defer os.Remove(pluginPath)

	// Open the plugin.
	p, err := plugin.Open(pluginPath)
	if err != nil {
		return err
	}

	// Look up the MangleTransform symbol.
	mangleTransform, err := p.Lookup("MangleTransform")
	if err != nil {
		return err
	}

	// Execute the MangleTransform function.
	return mangleTransform.(func(ast.Atom, []ast.ConstSubstList, func(ast.Atom) bool) error)(head, input, emit)
}

// EvalTransform evaluates a transform.
func EvalTransform(
	head ast.Atom,
	transform ast.Transform,
	input []ast.ConstSubstList,
	emit func(atom ast.Atom) bool) error {

	if transform.IsLetTransform() {
		return evalLet(head, transform, input, emit)
	}
	if transform.IsScriptTransform() {
		return evalScript(head, transform, input, emit)
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
	case symbols.AggregateBy.Symbol:
		// Use the same logic as GroupBy - the difference is mainly semantic and in the available functions
		keyToGroup := make(map[uint32]grouped)

		for _, subst := range input {
			keyLen := len(doStmt.Fn.Args)
			key := make([]ast.Constant, keyLen)
			b := make([]byte, 8*keyLen)
			
			if keyLen == 0 {
				// No grouping variables - treat all as one group
				key = []ast.Constant{}
				b = []byte{}
			} else {
				for i, v := range doStmt.Fn.Args {
					value := subst.Get(v.(ast.Variable)).(ast.Constant)
					binary.LittleEndian.PutUint64(b[i*8:], value.Hash())
					key[i] = value
				}
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
		
		// Apply aggregation functions (same as GroupBy)
		for _, group := range keyToGroup {
			var subst ast.ConstSubstList
			for i, v := range group.key {
				if i < len(doStmt.Fn.Args) {
					subst = subst.Extend(doStmt.Fn.Args[i].(ast.Variable), v)
				}
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
