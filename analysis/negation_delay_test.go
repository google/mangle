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

package analysis

import (
	"strings"
	"testing"

	"codeberg.org/TauCeti/mangle-go/parse"
)

// A negated premise whose variables no other premise binds must not be
// dropped: the rule would silently lose a condition. It must be rejected.
func TestNegationWithUnboundVariableIsRejected(t *testing.T) {
	for _, src := range []string{
		"e(1, 2). q(2). p(W) :- q(W), !e(Any, W).",
		"e(1, 2). q(2). p(W) :- q(W), !e(_, W).",
	} {
		unit, err := parse.Unit(strings.NewReader(src))
		if err != nil {
			t.Fatal(err)
		}
		if _, err := AnalyzeOneUnit(unit, nil); err == nil {
			t.Errorf("AnalyzeOneUnit(%q) succeeded, want an error about the unbound variable", src)
		}
	}
}

// When several negated premises wait for their variables, each is placed
// right after the premise that binds its last variable, and none is lost.
func TestEveryDelayedNegationIsKept(t *testing.T) {
	got := RewriteClause(nil, clause("p(X) :- !a(Y), !b(X), q(X), r(Y)."))
	want := clause("p(X) :- q(X), !b(X), r(Y), !a(Y).")
	if len(got.Premises) != len(want.Premises) {
		t.Fatalf("RewriteClause gave %v, want %v", got, want)
	}
	for i := range want.Premises {
		if !got.Premises[i].Equals(want.Premises[i]) {
			t.Errorf("RewriteClause gave %v, want %v", got, want)
			break
		}
	}
}
