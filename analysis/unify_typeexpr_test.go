package analysis

import (
	"strings"
	"testing"

	"codeberg.org/TauCeti/mangle-go/parse"
)

// Regression tests for type expression unification in bounds checking
// (unionfind.UnifyTypeExpr via checkFunApply). The unifier used to
// silently accept mismatching constants inside type constructor
// applications, and could not unify types with union type expressions.
func TestUnifyTypeExprChecking(t *testing.T) {
	tests := []struct {
		name    string
		src     string
		wantErr bool
	}{
		// fn:float:sum over floats works (its signature relates the
		// result to the element type of the list argument).
		{"float:sum of floats", "result(X) :- X = fn:float:sum([1.5, 2.5]).", false},
		{"float:max of floats", "result(X) :- X = fn:float:max([1.5, 2.5]).", false},
		// The runtime rejects number elements in float reducers, so
		// the type checker must reject them as well.
		{"float:sum of numbers", "result(X) :- X = fn:float:sum([1, 2]).", true},
		{"float:max of numbers", "result(X) :- X = fn:float:max([1, 2]).", true},
		// Mismatching element types are rejected (the unification
		// error used to be dropped inside the unifier).
		{"list:contains over mismatched element type",
			"result(X) :- X = fn:list:contains([1, 2], 'a').", true},
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

func TestUnifyTypeExprCheckingAnyBound(t *testing.T) {
	// The /any bound must not conflict with concrete types: a variable
	// holding a value of an /any-typed predicate can be passed to a
	// function whose other argument fixes the type variable.
	src := `
Decl p(X)
  bound [/any].

result(X) :- p(Y), X = fn:list:contains([1, 2], Y).
`
	unit, err := parse.Unit(strings.NewReader(src))
	if err != nil {
		t.Fatalf("parse.Unit() error = %v", err)
	}
	if _, err := AnalyzeAndCheckBounds([]parse.SourceUnit{unit}, nil, ErrorForBoundsMismatch); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}
