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

// An inequality whose variables are not bound yet is delayed until they are,
// so premise order does not matter: `X != Y, e(X), f(Y)` and its reorderings
// evaluate the same.
func TestInequalityWithUnboundVariableIsDelayed(t *testing.T) {
	got := RewriteClause(nil, clause("p(X) :- X != Y, e(X), f(Y)."))
	want := clause("p(X) :- e(X), f(Y), X != Y.")
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

// Inequalities and negations wait together, each placed right after the
// premise that binds its last variable.
func TestDelayedInequalitiesAndNegationsKeepOrder(t *testing.T) {
	got := RewriteClause(nil, clause("p(X) :- !a(Y), X != Z, q(X), r(Y), s(Z)."))
	want := clause("p(X) :- q(X), r(Y), !a(Y), s(Z), X != Z.")
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

// A fully bound inequality stays where it is, and an inequality whose
// variables no premise binds is kept, at the end, so that checking the rule
// reports the unbound variable.
func TestNeverBoundInequalityIsKeptAndReported(t *testing.T) {
	got := RewriteClause(nil, clause("p(X) :- e(X), X != 1, X != Y."))
	want := clause("p(X) :- e(X), X != 1, X != Y.")
	if len(got.Premises) != len(want.Premises) {
		t.Fatalf("RewriteClause gave %v, want %v", got, want)
	}

	unit, err := parse.Unit(strings.NewReader("e(1). p(X) :- e(X), X != Y."))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := AnalyzeOneUnit(unit, nil); err == nil {
		t.Error("AnalyzeOneUnit succeeded, want an error about the unbound variable Y")
	}
}
