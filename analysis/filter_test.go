package analysis

import (
	"strings"
	"testing"

	"codeberg.org/TauCeti/mangle-go/parse"
)

// Tests for type-checking :filter premises. The function argument of
// :filter must be able to evaluate to /true: a type error in the
// argument, or an argument of a non-boolean type, means the filter can
// never succeed and is rejected. Unknown types (such as /any) are
// admitted, since the value may or may not be /true at runtime.
func TestFilterPremise(t *testing.T) {
	tests := []struct {
		name    string
		src     string
		wantErr bool
	}{
		{"boolean function", "result() :- :filter(fn:list:contains([1, 2], 3)).", false},
		{"empty list", "result() :- :filter(fn:list:contains([], 23)).", false},
		{"type error in argument", "result() :- :filter(fn:list:contains([1, 2], 'a')).", true},
		{"non-boolean function", "result() :- :filter(fn:string:concat('a', 'b')).", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			unit, err := parse.Unit(strings.NewReader(tt.src))
			if err != nil {
				t.Fatalf("parse.Unit() error = %v", err)
			}
			_, err = AnalyzeAndCheckBounds([]parse.SourceUnit{unit}, nil, ErrorForBoundsMismatch)
			if tt.wantErr && err == nil {
				t.Errorf("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestFilterPremiseAnyTypedArgument(t *testing.T) {
	// The member argument of fn:list:contains is polymorphic; a value
	// of unknown type (/any) may well be a number, so this must pass.
	src := `
Decl p(X)
  bound [/any].

result() :- p(X), :filter(fn:list:contains([1, 2], X)).
`
	unit, err := parse.Unit(strings.NewReader(src))
	if err != nil {
		t.Fatalf("parse.Unit() error = %v", err)
	}
	if _, err := AnalyzeAndCheckBounds([]parse.SourceUnit{unit}, nil, ErrorForBoundsMismatch); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}
