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

// M = X does not give X a value. The negation and the inequality below must
// wait for x(X): run before it, the negation matched any q(1, _) and removed
// p(1), and the inequality failed with "X is not bound".
func TestEqualityOfUnboundVariablesWaitsForABinding(t *testing.T) {
	for _, test := range []struct{ program, want string }{
		{"z(1). x(1). x(2). q(1, 2).\np(M) :- M = X, z(Z), !q(Z, X), x(X).", "p(1)"},
		{"x(1). x(2).\np(M) :- M = X, X != 1, x(X).", "p(2)"},
	} {
		unit, err := parse.Unit(strings.NewReader(test.program))
		if err != nil {
			t.Fatal(err)
		}
		pi, err := analysis.AnalyzeOneUnit(unit, nil)
		if err != nil {
			t.Fatal(err)
		}
		store := factstore.NewSimpleInMemoryStore()
		if err := EvalProgram(pi, store); err != nil {
			t.Errorf("%s: %v", test.program, err)
			continue
		}
		var got []string
		store.GetFacts(ast.NewQuery(ast.PredicateSym{Symbol: "p", Arity: 1}), func(a ast.Atom) error {
			got = append(got, a.String())
			return nil
		})
		if strings.Join(got, " ") != test.want {
			t.Errorf("%s: got %v, want %s", test.program, got, test.want)
		}
	}
}
