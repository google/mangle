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

package interpreter

import (
	"bytes"
	"strings"
	"testing"

	"github.com/google/mangle/ast"
)

func TestMultiLineContinuation(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		shouldSucceed  bool
		expectedOutput string
	}{
		{
			name:          "single line rule",
			input:         "test_rule(X) :- other_pred(X).",
			shouldSucceed: true,
		},
		{
			name:          "rule with proper spacing",
			input:         "test_rule2(X) :- other_pred(X).",
			shouldSucceed: true,
		},
		{
			name:          "rule without space after :-",
			input:         "test_rule3(X) :-other_pred(X).",
			shouldSucceed: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			interpreter := New(&buf, "", nil)

			// First define the base predicate
			err := interpreter.Define("other_pred(/test).")
			if err != nil {
				t.Fatalf("Failed to define base predicate: %v", err)
			}

			// Then define the test rule
			err = interpreter.Define(tt.input)
			if tt.shouldSucceed && err != nil {
				t.Errorf("Expected success but got error: %v", err)
			}
			if !tt.shouldSucceed && err == nil {
				t.Errorf("Expected error but got success")
			}

			if tt.shouldSucceed {
				// Try to query the rule to see if it works
				results, err := interpreter.Query(parseQuery(t, interpreter, strings.Split(tt.input, "(")[0]))
				if err != nil {
					t.Errorf("Failed to query rule: %v", err)
				}
				if len(results) == 0 {
					t.Errorf("Expected results but got none")
				}
			}
		})
	}
}

func parseQuery(t *testing.T, interpreter *Interpreter, queryStr string) ast.Atom {
	query, err := interpreter.ParseQuery(queryStr)
	if err != nil {
		t.Fatalf("Failed to parse query %q: %v", queryStr, err)
	}
	return query
}