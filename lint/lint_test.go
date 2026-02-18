// Copyright 2024 Google LLC
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

package lint

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/mangle/parse"
)

func lintSource(t *testing.T, source string, config ...LintConfig) []LintResult {
	t.Helper()
	unit, err := parse.Unit(strings.NewReader(source))
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	cfg := DefaultConfig()
	if len(config) > 0 {
		cfg = config[0]
	}
	linter := NewLinter(cfg)
	results, err := linter.LintUnit("test.mg", unit)
	if err != nil {
		t.Fatalf("lint error: %v", err)
	}
	return results
}

func filterByRule(results []LintResult, rule string) []LintResult {
	var filtered []LintResult
	for _, r := range results {
		if r.RuleName == rule {
			filtered = append(filtered, r)
		}
	}
	return filtered
}

// --- Complexity Rule Tests ---

func TestOverlyComplexRule_Triggers(t *testing.T) {
	// A rule with 3 premises, threshold set to 2.
	source := `
bar(/x).
baz(/y).
qux(/z).
foo(X) :- bar(X), baz(X), qux(X).
`
	cfg := DefaultConfig()
	cfg.MaxPremises = 2
	results := lintSource(t, source, cfg)
	filtered := filterByRule(results, "overly-complex-rule")
	if len(filtered) != 1 {
		t.Errorf("got %d findings, want 1: %v", len(filtered), filtered)
	}
}

func TestOverlyComplexRule_NoTrigger(t *testing.T) {
	source := `
bar(/x).
foo(X) :- bar(X).
`
	results := lintSource(t, source)
	filtered := filterByRule(results, "overly-complex-rule")
	if len(filtered) != 0 {
		t.Errorf("got %d findings, want 0: %v", len(filtered), filtered)
	}
}

// --- Missing Doc Rule Tests ---

func TestMissingDoc_Triggers(t *testing.T) {
	source := `
Decl bar(X) bound [/name].
bar(/x).
`
	results := lintSource(t, source)
	filtered := filterByRule(results, "missing-doc")
	if len(filtered) != 1 {
		t.Errorf("got %d findings, want 1: %v", len(filtered), filtered)
	}
}

func TestMissingDoc_WithDoc(t *testing.T) {
	source := `
Decl bar(X) descr[doc("a documented predicate")] bound [/name].
bar(/x).
`
	results := lintSource(t, source)
	filtered := filterByRule(results, "missing-doc")
	if len(filtered) != 0 {
		t.Errorf("got %d findings, want 0: %v", len(filtered), filtered)
	}
}

// --- Naming Convention Rule Tests ---

func TestNamingConvention_BadPredicate(t *testing.T) {
	source := `
Decl pointsTo(X, Y) bound [/name, /name].
pointsTo(/a, /b).
`
	results := lintSource(t, source)
	filtered := filterByRule(results, "naming-convention")
	if len(filtered) != 1 {
		t.Errorf("got %d naming findings, want 1: %v", len(filtered), filtered)
	}
}

func TestNamingConvention_GoodPredicate(t *testing.T) {
	source := `
Decl points_to(X, Y) bound [/name, /name].
points_to(/a, /b).
`
	results := lintSource(t, source)
	filtered := filterByRule(results, "naming-convention")
	if len(filtered) != 0 {
		t.Errorf("got %d naming findings, want 0: %v", len(filtered), filtered)
	}
}

// --- Singleton Variable Rule Tests ---

func TestSingletonVariable_Triggers(t *testing.T) {
	source := `
bar(/x, /y).
foo(X) :- bar(X, Typo).
`
	results := lintSource(t, source)
	filtered := filterByRule(results, "singleton-variable")
	// "Typo" appears only in one premise and nowhere else.
	if len(filtered) != 1 {
		t.Errorf("got %d findings, want 1: %v", len(filtered), filtered)
	}
	if len(filtered) > 0 && !strings.Contains(filtered[0].Message, "Typo") {
		t.Errorf("expected message to mention 'Typo', got: %s", filtered[0].Message)
	}
}

func TestSingletonVariable_Wildcard(t *testing.T) {
	// Wildcards should not trigger.
	source := `
bar(/x, /y).
foo(X) :- bar(X, _).
`
	results := lintSource(t, source)
	filtered := filterByRule(results, "singleton-variable")
	if len(filtered) != 0 {
		t.Errorf("got %d findings, want 0: %v", len(filtered), filtered)
	}
}

func TestSingletonVariable_TransformOutput(t *testing.T) {
	// Transform output variables used in head should not be singletons.
	source := `
Decl score(X, Y) bound [/name, /number].
score(/alice, 10).
score(/alice, 20).
Decl total_score(X, Y) bound [/name, /number].
total_score(Name, Total) :- score(Name, _) |> do fn:group_by(Name), let Total = fn:sum(Name).
`
	results := lintSource(t, source)
	filtered := filterByRule(results, "singleton-variable")
	// Total is in transform output AND head, Name is in premises+head+transform.
	// None should be singletons.
	for _, f := range filtered {
		if strings.Contains(f.Message, "Total") {
			t.Errorf("transform output variable 'Total' should not be flagged: %s", f.Message)
		}
	}
}

// --- Unused Predicate Rule Tests ---

func TestUnusedPredicate_Triggers(t *testing.T) {
	source := `
Decl unused_pred(X) bound [/name].
Decl used_pred(X) bound [/name].
used_pred(/x).
foo(X) :- used_pred(X).
`
	results := lintSource(t, source)
	filtered := filterByRule(results, "unused-predicate")
	if len(filtered) != 1 {
		t.Errorf("got %d findings, want 1: %v", len(filtered), filtered)
	}
	if len(filtered) > 0 && !strings.Contains(filtered[0].Message, "unused_pred") {
		t.Errorf("expected message to mention 'unused_pred', got: %s", filtered[0].Message)
	}
}

func TestUnusedPredicate_AllUsed(t *testing.T) {
	source := `
Decl bar(X) bound [/name].
bar(/x).
foo(X) :- bar(X).
`
	results := lintSource(t, source)
	filtered := filterByRule(results, "unused-predicate")
	if len(filtered) != 0 {
		t.Errorf("got %d findings, want 0: %v", len(filtered), filtered)
	}
}

// --- Dead Code Rule Tests ---

func TestDeadCode_Triggers(t *testing.T) {
	source := `
bar(/x).
dead(X) :- bar(X).
`
	results := lintSource(t, source)
	filtered := filterByRule(results, "dead-code")
	if len(filtered) != 1 {
		t.Errorf("got %d findings, want 1: %v", len(filtered), filtered)
	}
	if len(filtered) > 0 && !strings.Contains(filtered[0].Message, "dead") {
		t.Errorf("expected message to mention 'dead', got: %s", filtered[0].Message)
	}
}

func TestDeadCode_Consumed(t *testing.T) {
	source := `
bar(/x).
intermediate(X) :- bar(X).
result(X) :- intermediate(X).
`
	results := lintSource(t, source)
	filtered := filterByRule(results, "dead-code")
	// "intermediate" is consumed by "result", so only "result" is dead code.
	for _, f := range filtered {
		if strings.Contains(f.Message, "intermediate") {
			t.Errorf("intermediate should not be flagged as dead code: %s", f.Message)
		}
	}
}

// --- Config Tests ---

func TestDisableRule(t *testing.T) {
	source := `
bar(/x, /y).
foo(X) :- bar(X, Typo).
`
	cfg := DefaultConfig()
	cfg.DisabledRules = map[string]bool{"singleton-variable": true}
	results := lintSource(t, source, cfg)
	filtered := filterByRule(results, "singleton-variable")
	if len(filtered) != 0 {
		t.Errorf("disabled rule should produce 0 findings, got %d", len(filtered))
	}
}

func TestMinSeverity(t *testing.T) {
	source := `
bar(/x).
dead(X) :- bar(X).
`
	cfg := DefaultConfig()
	cfg.MinSeverity = SeverityWarning
	results := lintSource(t, source, cfg)
	for _, r := range results {
		if r.Severity < SeverityWarning {
			t.Errorf("found result below min severity: %v", r)
		}
	}
}

// --- Output Tests ---

func TestFormatText(t *testing.T) {
	results := []LintResult{
		{RuleName: "test-rule", Severity: SeverityWarning, File: "foo.mg", Message: "test message"},
	}
	var buf bytes.Buffer
	FormatText(&buf, results)
	out := buf.String()
	if !strings.Contains(out, "foo.mg") || !strings.Contains(out, "warning") || !strings.Contains(out, "test message") {
		t.Errorf("unexpected text output: %s", out)
	}
}

func TestFormatJSON(t *testing.T) {
	results := []LintResult{
		{RuleName: "test-rule", Severity: SeverityWarning, File: "foo.mg", Message: "test message"},
	}
	var buf bytes.Buffer
	if err := FormatJSON(&buf, results); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.Contains(out, `"warning"`) || !strings.Contains(out, `"test message"`) {
		t.Errorf("unexpected JSON output: %s", out)
	}
}

func TestFormatJSON_Empty(t *testing.T) {
	var buf bytes.Buffer
	if err := FormatJSON(&buf, nil); err != nil {
		t.Fatal(err)
	}
	out := strings.TrimSpace(buf.String())
	if out != "[]" {
		t.Errorf("expected empty JSON array, got: %s", out)
	}
}

// --- Integration Tests ---

func TestExamplesNoErrors(t *testing.T) {
	// All example .mg files should parse and lint without hard errors.
	// We allow info/warning findings but not analysis errors.
	examplesDir := "../examples"
	if _, err := os.Stat(examplesDir); os.IsNotExist(err) {
		t.Skip("examples directory not found, skipping integration test")
	}
	matches, err := filepath.Glob(filepath.Join(examplesDir, "*.mg"))
	if err != nil {
		t.Fatal(err)
	}
	if len(matches) == 0 {
		t.Skip("no .mg files found in examples directory")
	}

	linter := NewLinter(DefaultConfig())
	for _, path := range matches {
		t.Run(filepath.Base(path), func(t *testing.T) {
			results, err := linter.LintFile(path)
			if err != nil {
				t.Fatalf("lint error on %s: %v", path, err)
			}
			// Check that no result is an analysis error (which would indicate
			// the linter pipeline broke on valid code).
			for _, r := range results {
				if r.RuleName == "analysis" {
					t.Errorf("analysis error on valid example %s: %s", path, r.Message)
				}
			}
		})
	}
}

// --- Severity Tests ---

func TestParseSeverity(t *testing.T) {
	tests := []struct {
		input string
		want  Severity
	}{
		{"info", SeverityInfo},
		{"warning", SeverityWarning},
		{"error", SeverityError},
		{"unknown", SeverityInfo},
	}
	for _, tt := range tests {
		got := ParseSeverity(tt.input)
		if got != tt.want {
			t.Errorf("ParseSeverity(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestSeverityString(t *testing.T) {
	tests := []struct {
		sev  Severity
		want string
	}{
		{SeverityInfo, "info"},
		{SeverityWarning, "warning"},
		{SeverityError, "error"},
	}
	for _, tt := range tests {
		got := tt.sev.String()
		if got != tt.want {
			t.Errorf("Severity(%d).String() = %q, want %q", tt.sev, got, tt.want)
		}
	}
}
