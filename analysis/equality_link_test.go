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

import "testing"

// An equality between two variables that have no value yet binds neither, so
// a negation or an inequality on them still waits for a premise that binds
// them. The equality is compiled away by renaming one variable to the other,
// so once one of them has a value, so does the other.
func TestEqualityOfUnboundVariablesDoesNotBind(t *testing.T) {
	for _, test := range []struct{ rule, want string }{
		{"p(M) :- M = X, X != 1, x(X).", "p(M) :- x(M), M != 1."},
		{"p(M) :- M = X, z(Z), !q(Z, X), x(X).", "p(M) :- z(Z), x(M), !q(Z, M)."},
		// The renaming works in both directions, and along a chain of equalities.
		{"p(M) :- M = X, Y = M, !q(Y), x(X).", "p(M) :- x(M), !q(M)."},
		{"p(M) :- X = M, !q(M), x(X).", "p(M) :- x(M), !q(M)."},
		// An equality with a side that has a value binds the other side.
		{"p(M) :- x(X), M = X, !q(M).", "p(M) :- x(M), !q(M)."},
		{"p(M) :- M = 1, !q(M), x(M).", "p(M) :- M = 1, !q(M), x(M)."},
	} {
		got := RewriteClause(nil, clause(test.rule))
		want := clause(test.want)
		if got.String() != want.String() {
			t.Errorf("RewriteClause(%s) = %v, want %v", test.rule, got, want)
		}
	}
}
