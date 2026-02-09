package analysis

import (
	"strings"
	"testing"

	"github.com/google/mangle/ast"
	"github.com/google/mangle/parse"
)

func TestTemporalValidation(t *testing.T) {
	tests := []struct {
		name    string
		program string
		wantErr bool
		err     string
	}{
		{
			name: "valid temporal predicate",
			program: `
				Decl p(X) temporal bound [/any].
				p(1) @[2020-01-01, 2021-01-01].
			`,
			wantErr: false,
		},
		{
			name: "valid synthetic temporal predicate",
			program: `
				p(1) @[2020-01-01, 2021-01-01].
				p(2) @[2021-01-01, 2022-01-01].
			`,
			wantErr: false,
		},
		{
			name: "invalid non-temporal predicate with annotation",
			program: `
				Decl p(X) bound [/any].
				p(1) @[2020-01-01, 2021-01-01].
			`,
			wantErr: true,
			err:     "predicate p(A0) is not declared temporal",
		},
		{
			name: "invalid eternal fact for temporal predicate",
			program: `
				Decl p(X) temporal bound [/any].
				p(1).
			`,
			wantErr: true,
			err:     "defined without temporal annotation",
		},
		{
			name: "valid annotated fact for temporal predicate",
			program: `
				Decl p(X) temporal bound [/any].
				p(1) @[2020-01-01, 2021-01-01].
			`,
			wantErr: false,
		},
		{
			name: "mixed usage synthetic (invalid)",
			program: `
				p(1) @[2020-01-01, 2021-01-01].
				p(2).
			`,
			wantErr: true,
			err:     "defined without temporal annotation",
		},
		{
			name: "mixed usage explicit temporal (invalid)",
			program: `
				Decl p(X) temporal bound [/any].
				p(1) @[2020-01-01, 2021-01-01].
				p(2).
			`,
			wantErr: true,
			err:     "defined without temporal annotation",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			unit, err := parse.Unit(strings.NewReader(tt.program))
			if err != nil {
				t.Fatalf("parse.Unit() error = %v", err)
			}

			_, err = AnalyzeOneUnit(unit, nil)
			if tt.wantErr {
				if err == nil {
					t.Errorf("Analyze() expected error, got nil")
				} else if tt.err != "" {
					if !contains(err.Error(), tt.err) {
						t.Errorf("Analyze() error = %v, want %v", err, tt.err)
					}
				}
			} else {
				if err != nil {
					t.Errorf("Analyze() unexpected error = %v", err)
				}
			}
		})
	}
}

func TestIntervalVariableTypes(t *testing.T) {
	tests := []struct {
		name    string
		program string
		wantErr bool
		err     string
	}{
		{
			name: "valid interval time variable",
			program: `
				Decl p(X) temporal bound [/any].
				Decl q(T) bound [/time].
				p(1) @[T, _] :- q(T).
			`,
			wantErr: false,
		},
		{
			name: "invalid interval string variable in head",
			program: `
				Decl p(X) temporal bound [/any].
				Decl q(S) bound [/string].
				p(1) @[S, _] :- q(S).
			`,
			wantErr: true,
			err:     "HeadTime variables must be of type Time",
		},
		{
			name: "invalid interval string variable in premise",
			program: `
				Decl p(X) temporal bound [/any].
				Decl q(S) bound [/string].
				p(1) @[2020-01-01, 2021-01-01] :- p(1) @[S, _], q(S).
			`,
			wantErr: true,
			err:     "type mismatch",
		},
		{
			name: "invalid non-temporal predicate in premise temporal annotation",
			program: `
				Decl p(X) temporal bound [/any].
				Decl q(X) bound [/any].
				p(X) @[_,_] :- q(X) @[2020-01-01, 2021-01-01].
			`,
			wantErr: true,
			err:     "predicate q(A0) is not declared temporal",
		},
		{
			name: "invalid bare usage of temporal predicate",
			program: `
				Decl p(X) temporal bound [/any].
				Decl q(X) temporal bound [/any].
				p(X) @[_,_] :- q(X).
			`,
			wantErr: true,
			err:     "temporal predicate q(A0) used without temporal annotation",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			unit, err := parse.Unit(strings.NewReader(tt.program))
			if err != nil {
				t.Fatalf("parse.Unit() error = %v", err)
			}

			// We need to use ErrorForBoundsMismatch to detect type errors
			_, err = AnalyzeAndCheckBounds([]parse.SourceUnit{unit}, nil, ErrorForBoundsMismatch)
			if tt.wantErr {
				if err == nil {
					t.Errorf("Analyze() expected error, got nil")
				} else if tt.err != "" {
					if !contains(err.Error(), tt.err) {
						t.Errorf("Analyze() error = %v, want %v", err, tt.err)
					}
				}
			} else {
				if err != nil {
					t.Errorf("Analyze() unexpected error = %v", err)
				}
			}
		})
	}
}

func TestTemporalAtomNormalization(t *testing.T) {
	// Identify predicates.
	p := ast.PredicateSym{"p", 1}
	q := ast.PredicateSym{"q", 1}

	// Declare p and q as temporal.
	// We use a custom analyzer to bypass the parser and manually inject declarations and clauses.
	analyzer, err := New(nil, nil, ErrorForBoundsMismatch)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	analyzer.decl[p] = ast.Decl{
		DeclaredAtom: ast.Atom{p, []ast.BaseTerm{ast.Variable{"X"}}},
		Descr:        []ast.Atom{ast.NewAtom(ast.DescrTemporal)},
	}
	analyzer.decl[q] = ast.Decl{
		DeclaredAtom: ast.Atom{q, []ast.BaseTerm{ast.Variable{"X"}}},
		Descr:        []ast.Atom{ast.NewAtom(ast.DescrTemporal)},
	}

	// Helper to create *ast.Interval
	mkInterval := func(start, end string) *ast.Interval {
		i := ast.NewInterval(ast.NewVariableBound(ast.Variable{start}), ast.NewVariableBound(ast.Variable{end}))
		return &i
	}

	// Case 1: Valid TemporalAtom (should be normalized and pass)
	// p(X) :- q(X) @[S, E]
	validClause := ast.Clause{
		Head:     ast.Atom{p, []ast.BaseTerm{ast.Variable{"X"}}},
		HeadTime: mkInterval("S", "E"),
		Premises: []ast.Term{
			ast.TemporalAtom{
				Atom:     ast.Atom{q, []ast.BaseTerm{ast.Variable{"X"}}},
				Interval: mkInterval("S", "E"),
			},
		},
	}

	// Analyze the clause.
	_, err = analyzer.Analyze([]ast.Clause{validClause})
	if err != nil {
		t.Errorf("Analyze(validClause) unexpected error: %v", err)
	}

	// Verify normalization
	info, err := analyzer.Analyze([]ast.Clause{validClause})
	if err != nil {
		t.Errorf("Analyze(validClause) unexpected error: %v", err)
	}
	if len(info.Rules) != 1 {
		t.Fatalf("Expected 1 rule, got %d", len(info.Rules))
	}
	rule := info.Rules[0]
	if _, ok := rule.Premises[0].(ast.TemporalLiteral); !ok {
		t.Errorf("Premise was not normalized to TemporalLiteral: %T", rule.Premises[0])
	}

	// Case 2: Invalid TemporalAtom (bare usage, missing interval for temporal predicate)
	// p(X) :- q(X)
	// represented as TemporalAtom with nil Interval.
	invalidClause := ast.Clause{
		Head:     ast.Atom{p, []ast.BaseTerm{ast.Variable{"X"}}},
		HeadTime: mkInterval("S", "E"),
		Premises: []ast.Term{
			ast.TemporalAtom{
				Atom:     ast.Atom{q, []ast.BaseTerm{ast.Variable{"X"}}},
				Interval: nil,
			},
		},
	}

	_, err = analyzer.Analyze([]ast.Clause{invalidClause})
	if err == nil {
		t.Errorf("Analyze(invalidClause) expected error, got nil")
	} else if !contains(err.Error(), "used without temporal annotation") {
		t.Errorf("Analyze(invalidClause) error = %v, want 'used without temporal annotation'", err)
	}
}

func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}
