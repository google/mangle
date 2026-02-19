// Copyright 2023 Google LLC
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

package analysis

import (
	"codeberg.org/TauCeti/mangle-go/ast"
)

// VarList is an ordered list of variables.
type VarList struct {
	Vars []ast.Variable
}

// NewVarList converts a map representation to a varlist deterministically.
func NewVarList(m map[ast.Variable]bool) VarList {
	var vars []ast.Variable
	for v := range m {
		vars = append(vars, v)
	}
	return VarList{vars}
}

// AsMap converts VarList to a map representation.
func (vs VarList) AsMap() map[ast.Variable]bool {
	used := make(map[ast.Variable]bool)
	for _, v := range vs.Vars {
		used[v] = true
	}
	return used
}

// Extend returns a new VarList with appended list of variables.
func (vs VarList) Extend(vars []ast.Variable) VarList {
	return VarList{append(vs.Vars, vars...)}
}

// Contains returns true if this VarList contains the given variable.
func (vs VarList) Contains(v ast.Variable) bool {
	return vs.Find(v) != -1
}

// Find returns the index of the given variable, or -1 if not found.
func (vs VarList) Find(v ast.Variable) int {
	for i, u := range vs.Vars {
		if u.Symbol == v.Symbol {
			return i
		}
	}
	return -1
}
