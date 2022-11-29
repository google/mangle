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

package rewrite

import (
	"fmt"
	"testing"

	"github.com/google/mangle/analysis"
	"github.com/google/mangle/ast"
	"github.com/google/mangle/parse"
)

func clause(str string) ast.Clause {
	clause, err := parse.Clause(str)
	if err != nil {
		panic(fmt.Errorf("bad syntax in test case: %s got %w", str, err))
	}
	return clause
}

func TestRewrite(t *testing.T) {
	got := Rewrite(analysis.Program{
		EdbPredicates: map[ast.PredicateSym]struct{}{
			ast.PredicateSym{"num", 1}:  struct{}{},
			ast.PredicateSym{"succ", 2}: struct{}{},
		},
		IdbPredicates: map[ast.PredicateSym]struct{}{
			ast.PredicateSym{"odd", 1}: struct{}{},
		},
		Rules: []ast.Clause{
			clause("count(A) :- odd(X), succ(Y, Z) |> do fn:group_by(), let Z = fn:count()."),
		},
	})

	want := []struct {
		headSym     ast.PredicateSym
		premiseSyms []ast.PredicateSym
	}{
		{
			headSym:     ast.PredicateSym{"count1__tmp", 3},
			premiseSyms: []ast.PredicateSym{ast.PredicateSym{"odd", 1}, ast.PredicateSym{"succ", 2}},
		},
		{
			headSym:     ast.PredicateSym{"count", 1},
			premiseSyms: []ast.PredicateSym{ast.PredicateSym{"count1__tmp", 3}},
		},
	}

ExpectedLoop:
	for _, expected := range want {
		for _, actual := range got.Rules {
			if actual.Head.Predicate == expected.headSym && samePremiseSyms(actual.Premises, expected.premiseSyms) {
				continue ExpectedLoop
			}
		}
		t.Errorf("expected to find %v actual clauses %v", expected, got.Rules)
	}
}

// Returns true if the predicates of the premises are as expected.
// This is necessary since we cannot predict the order of predicates.
func samePremiseSyms(premises []ast.Term, expectedSyms []ast.PredicateSym) bool {
	for i, premise := range premises {
		atom, ok := premise.(ast.Atom)
		if !ok {
			continue
		}
		if atom.Predicate != expectedSyms[i] {
			return false
		}
	}
	return true
}
