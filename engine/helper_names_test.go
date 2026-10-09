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

// Two aggregations over several premises with the same head predicate each
// get their own helper predicate, so each counts only its own rows.
func TestAggregationsWithOneHeadCountSeparately(t *testing.T) {
	unit, err := parse.Unit(strings.NewReader(`
		o(1, 5). o(2, 7). o(3, -2).
		c("sales", N) :- o(I, Q), Q > 0 |> do fn:group_by(), let N = fn:count().
		c("refunds", N) :- o(I, Q), Q < 0 |> do fn:group_by(), let N = fn:count().
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
	for _, want := range []ast.Atom{
		ast.NewAtom("c", ast.String("sales"), ast.Number(2)),
		ast.NewAtom("c", ast.String("refunds"), ast.Number(1)),
	} {
		if !store.Contains(want) {
			t.Errorf("missing %v", want)
		}
	}
}
