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

package analysis

import (
	"strings"
	"testing"

	"github.com/google/mangle/parse"
)

func TestCheckTemporalRecursion(t *testing.T) {
	tests := []struct {
		name            string
		program         string
		wantWarnings    int
		wantSeverity    WarningSeverity
		wantMsgContains string
	}{
		{
			name: "no temporal predicates",
			program: `
				foo(X) :- bar(X).
				bar(1).
			`,
			wantWarnings: 0,
		},
		{
			name: "non-recursive temporal predicate",
			program: `
				Decl active(X) temporal bound [/name].
				Decl base_active(X) bound [/name].
				active(X) :- base_active(X).
				base_active(/alice).
			`,
			wantWarnings: 0,
		},
		{
			name: "self-recursive temporal predicate",
			program: `
				Decl derived(X) temporal bound [/name].
				Decl some_condition(X) bound [/name].
				derived(X) :- derived(X), some_condition(X).
				some_condition(/alice).
			`,
			wantWarnings:    1,
			wantSeverity:    SeverityWarning,
			wantMsgContains: "self-recursive",
		},
		{
			name: "mutual recursion with temporal",
			program: `
				Decl foo(X) temporal bound [/name].
				Decl bar(X) temporal bound [/name].
				foo(X) :- bar(X).
				bar(X) :- foo(X).
				bar(/alice).
			`,
			wantWarnings:    1,
			wantSeverity:    SeverityCritical,
			wantMsgContains: "mutual recursion",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			unit, err := parse.Unit(strings.NewReader(tt.program))
			if err != nil {
				t.Fatalf("parse error: %v", err)
			}

			programInfo, err := AnalyzeOneUnit(unit, nil)
			if err != nil {
				t.Fatalf("analysis error: %v", err)
			}

			warnings := CheckTemporalRecursion(programInfo)

			if len(warnings) != tt.wantWarnings {
				t.Errorf("got %d warnings, want %d", len(warnings), tt.wantWarnings)
				for _, w := range warnings {
					t.Logf("warning: %v", w)
				}
			}

			if tt.wantWarnings > 0 && len(warnings) > 0 {
				if warnings[0].Severity != tt.wantSeverity {
					t.Errorf("got severity %v, want %v", warnings[0].Severity, tt.wantSeverity)
				}
				if tt.wantMsgContains != "" && !strings.Contains(warnings[0].Message, tt.wantMsgContains) {
					t.Errorf("warning message %q does not contain %q", warnings[0].Message, tt.wantMsgContains)
				}
			}
		})
	}
}

func TestWarningSeverityString(t *testing.T) {
	tests := []struct {
		severity WarningSeverity
		want     string
	}{
		{SeverityInfo, "info"},
		{SeverityWarning, "warning"},
		{SeverityCritical, "critical"},
	}

	for _, tt := range tests {
		if got := tt.severity.String(); got != tt.want {
			t.Errorf("WarningSeverity(%d).String() = %q, want %q", tt.severity, got, tt.want)
		}
	}
}
