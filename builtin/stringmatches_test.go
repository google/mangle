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

package builtin

// :string:matches — the RE2 predicate, and the two refusals that make it trustworthy.
//
// A membership test over an approved SET is expressible today with :list:member; a test over an
// approved SHAPE is not. That is what this adds: "every environment name is lower-case letters and
// hyphens", "every migration file is NNN_name.sql", stated once as a pattern rather than as a list
// somebody keeps in step by hand.
//
// THE ONE REFUSAL THAT MATTERS is an invalid pattern. regexp.MatchString returns (false, err), and a
// caller that drops the error reports "did not match" — indistinguishable from a clean answer and
// meaning the opposite. There is no correct `false` for a pattern that does not compile: the program
// is wrong, not the datum. A non-string SUBJECT is a silent false, as it is for every sibling.

import (
	"strings"
	"testing"

	"codeberg.org/TauCeti/mangle-go/ast"
	"codeberg.org/TauCeti/mangle-go/symbols"
	"codeberg.org/TauCeti/mangle-go/unionfind"
)

func decideMatches(t *testing.T, subject, pattern ast.Constant) (bool, error) {
	t.Helper()
	atom := ast.NewAtom(symbols.StringMatches.Symbol, subject, pattern)
	atom.Predicate = symbols.StringMatches
	subst := unionfind.New()
	ok, _, err := Decide(atom, &subst)
	return ok, err
}

func TestStringMatches(t *testing.T) {
	for _, tt := range []struct {
		name    string
		subject string
		pattern string
		want    bool
	}{
		{"anchored, matching", "dev", "^[a-z-]+$", true},
		{"anchored, not matching", "Dev1", "^[a-z-]+$", false},
		{"unanchored matches a substring", "a-migration-file", "migration", true},
		{"a pattern matching nothing in the subject", "abc", "xyz", false},
		{"the empty pattern matches everything", "anything", "", true},
		{"the empty subject matches an anchored empty pattern", "", "^$", true},
		{"alternation", "stg", "^(dev|qa|stg|prd)$", true},
		{"alternation, absent branch", "sandbox", "^(dev|qa|stg|prd)$", false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got, err := decideMatches(t, ast.String(tt.subject), ast.String(tt.pattern))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf(":string:matches(%q, %q) = %v, want %v", tt.subject, tt.pattern, got, tt.want)
			}
		})
	}
}

// AN INVALID PATTERN IS AN ERROR, NOT A FALSE. This is the case the implementation comment is about:
// a swallowed compile error reads as "did not match", which is a clean-looking wrong answer.
func TestStringMatchesRefusesAnInvalidPattern(t *testing.T) {
	_, err := decideMatches(t, ast.String("anything"), ast.String("([unclosed"))
	if err == nil {
		t.Fatal("an uncompilable pattern was answered rather than refused — a swallowed compile error reads as 'did not match'")
	}
	if !strings.Contains(err.Error(), "does not compile") {
		t.Errorf("the refusal does not say the pattern failed to compile: %v", err)
	}
}

// A NON-STRING SUBJECT IS A SILENT false, which is what :string:contains, :string:starts_with and
// :string:ends_with all do. Pinned so the convention is not broken by accident: a predicate over
// heterogeneous data filters rather than refusing, and only the PATTERN — where no `false` is
// correct — is an error.
func TestStringMatchesFiltersANonStringSubject(t *testing.T) {
	ok, err := decideMatches(t, ast.Number(42), ast.String("^[0-9]+$"))
	if err != nil {
		t.Fatalf("a non-string subject was refused rather than filtered, unlike its siblings: %v", err)
	}
	if ok {
		t.Error("a non-string subject matched")
	}
}

func TestStringMatchesRefusesANonStringPattern(t *testing.T) {
	_, err := decideMatches(t, ast.String("42"), ast.Number(42))
	if err == nil {
		t.Fatal("a number was used as a pattern and answered rather than refused")
	}
	if !strings.Contains(err.Error(), "2nd argument") {
		t.Errorf("the refusal does not name the argument at fault: %v", err)
	}
}

// THE CACHE MUST NOT CHANGE AN ANSWER, which is the only thing about it worth asserting: the same
// pattern is compiled once and every later call is a hit, so a cache that returned the WRONG
// compiled pattern would be invisible until two patterns were used together.
func TestTheCompiledPatternCacheDoesNotConfuseTwoPatterns(t *testing.T) {
	for i := 0; i < 3; i++ {
		if ok, err := decideMatches(t, ast.String("dev"), ast.String("^dev$")); err != nil || !ok {
			t.Fatalf("round %d: ^dev$ against \"dev\" = %v, %v", i, ok, err)
		}
		if ok, err := decideMatches(t, ast.String("dev"), ast.String("^prd$")); err != nil || ok {
			t.Fatalf("round %d: ^prd$ against \"dev\" = %v, %v — the cache returned another pattern", i, ok, err)
		}
	}
}

// AND AN INVALID PATTERN IS NOT CACHED AS A FAILURE: it refuses every time, so a typo keeps
// reporting itself rather than being answered from a cache of failures.
func TestAnInvalidPatternRefusesEveryTime(t *testing.T) {
	for i := 0; i < 3; i++ {
		if _, err := decideMatches(t, ast.String("x"), ast.String("([unclosed")); err == nil {
			t.Fatalf("round %d: an invalid pattern stopped refusing", i)
		}
	}
}
