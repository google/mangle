// Copyright 2026 The Mangle Authors
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
	"strings"
	"testing"

	"codeberg.org/TauCeti/mangle-go/analysis"
	"codeberg.org/TauCeti/mangle-go/ast"
	"codeberg.org/TauCeti/mangle-go/factstore"
	"codeberg.org/TauCeti/mangle-go/parse"
)

// An aggregation's premise with a repeated variable only matches facts whose
// arguments at those positions are equal.
func TestAggregationPremiseWithRepeatedVariable(t *testing.T) {
	unit, err := parse.Unit(strings.NewReader(`
		e(1, 1). e(2, 2). e(1, 2).
		n(N) :- e(X, X) |> do fn:group_by(), let N = fn:count().
	`))
	if err != nil {
		t.Fatal(err)
	}
	info, err := analysis.AnalyzeOneUnit(unit, nil)
	if err != nil {
		t.Fatal(err)
	}
	store := factstore.NewSimpleInMemoryStore()
	if err := EvalProgram(info, store); err != nil {
		t.Fatal(err)
	}
	if want := ast.NewAtom("n", ast.Number(2)); !store.Contains(want) {
		t.Errorf("missing %v", want)
	}
}
