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
	"testing"

	"codeberg.org/TauCeti/mangle-go/ast"
)

func TestVarList(t *testing.T) {
	vs := VarList{[]ast.Variable{ast.Variable{"X"}, ast.Variable{"Y"}}}

	if i := vs.Find(ast.Variable{"X"}); i != 0 {
		t.Errorf("Find(X)=%d want 0", i)
	}
	if !vs.Contains(ast.Variable{"Y"}) {
		t.Errorf("Contains(Y)=false want true")
	}
	if i := vs.Find(ast.Variable{"Z"}); i != -1 {
		t.Errorf("Find(X)=%d want -1", i)
	}
	if vs.Contains(ast.Variable{"Z"}) {
		t.Errorf("Contains(Z)=true want false")
	}

	vs = vs.Extend([]ast.Variable{ast.Variable{"Z"}})
	if !vs.Contains(ast.Variable{"Z"}) {
		t.Errorf("Contains(Z)=false want true")
	}
	if i := vs.Find(ast.Variable{"Z"}); i != 2 {
		t.Errorf("Find(Z)=%d want 2", i)
	}
}
