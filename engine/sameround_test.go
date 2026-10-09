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

// Two premises of one recursive rule that first hold in the same incremental
// round must still be joined. s, l and r are mutually recursive: round 1
// derives s(1), the next round derives l(1) and r(1) together, and s(2) needs
// the join of those two facts.
func TestJoinOfFactsNewInTheSameRound(t *testing.T) {
	unit, err := parse.Unit(strings.NewReader(`
		start(1).
		next(1, 2).
		s(X) :- start(X).
		l(X) :- s(X).
		r(X) :- s(X).
		s(Y) :- l(X), r(X), next(X, Y).
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
	want := ast.NewAtom("s", ast.Number(2))
	if !store.Contains(want) {
		t.Errorf("missing %v", want)
	}
}
