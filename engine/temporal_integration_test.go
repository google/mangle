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

package engine

import (
	"strings"
	"testing"
	"time"

	"github.com/google/mangle/ast"
	"github.com/google/mangle/factstore"
	"github.com/google/mangle/parse"
	"github.com/google/mangle/unionfind"
)

// End-to-end integration tests for temporal reasoning in Mangle.
// These tests verify the complete flow from parsing to evaluation.

func TestIntegration_TemporalFactParsing(t *testing.T) {
	tests := []struct {
		name        string
		program     string
		wantFacts   int
		wantTemporal bool
	}{
		{
			name:        "simple temporal fact",
			program:     "foo(/bar)@[2024-01-15, 2024-06-30].",
			wantFacts:   1,
			wantTemporal: true,
		},
		{
			name:        "point interval fact",
			program:     "event(/login)@[2024-03-15].",
			wantFacts:   1,
			wantTemporal: true,
		},
		{
			name:        "non-temporal fact",
			program:     "regular(/fact).",
			wantFacts:   1,
			wantTemporal: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			unit, err := parse.Unit(strings.NewReader(test.program))
			if err != nil {
				t.Fatalf("Failed to parse program: %v", err)
			}

			if len(unit.Clauses) != test.wantFacts {
				t.Errorf("Got %d clauses, want %d", len(unit.Clauses), test.wantFacts)
			}

			if len(unit.Clauses) > 0 {
				clause := unit.Clauses[0]
				hasTemporal := clause.HeadTime != nil
				if hasTemporal != test.wantTemporal {
					t.Errorf("Clause has temporal = %v, want %v", hasTemporal, test.wantTemporal)
				}
			}
		})
	}
}

func TestIntegration_TemporalDeclarations(t *testing.T) {
	tests := []struct {
		name         string
		program      string
		predName     string
		wantTemporal bool
	}{
		{
			name:         "temporal predicate declaration",
			program:      "Decl employee(X) temporal bound [/name].",
			predName:     "employee",
			wantTemporal: true,
		},
		{
			name:         "non-temporal predicate declaration",
			program:      "Decl config(X) bound [/string].",
			predName:     "config",
			wantTemporal: false,
		},
		{
			name: "temporal with documentation",
			program: `Decl status(X, Y) temporal
				descr [doc("Employee status over time")]
				bound [/name, /string].`,
			predName:     "status",
			wantTemporal: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			unit, err := parse.Unit(strings.NewReader(test.program))
			if err != nil {
				t.Fatalf("Failed to parse program: %v", err)
			}

			// Find the declaration (skip Package decl)
			var decl ast.Decl
			for _, d := range unit.Decls {
				if d.DeclaredAtom.Predicate.Symbol == test.predName {
					decl = d
					break
				}
			}

			if decl.DeclaredAtom.Predicate.Symbol != test.predName {
				t.Fatalf("Declaration not found for predicate %s", test.predName)
			}

			if decl.IsTemporal() != test.wantTemporal {
				t.Errorf("Decl.IsTemporal() = %v, want %v", decl.IsTemporal(), test.wantTemporal)
			}
		})
	}
}

func TestIntegration_TemporalOperators(t *testing.T) {
	tests := []struct {
		name      string
		program   string
		wantOpType ast.TemporalOperatorType
	}{
		{
			name:      "diamond minus operator",
			program:   "recently_active(X) :- <-[0d, 7d] active(X).",
			wantOpType: ast.DiamondMinus,
		},
		{
			name:      "box minus operator",
			program:   "stable(X) :- [-[0d, 30d] employed(X).",
			wantOpType: ast.BoxMinus,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			unit, err := parse.Unit(strings.NewReader(test.program))
			if err != nil {
				t.Fatalf("Failed to parse program: %v", err)
			}

			if len(unit.Clauses) != 1 {
				t.Fatalf("Expected 1 clause, got %d", len(unit.Clauses))
			}

			clause := unit.Clauses[0]
			if len(clause.Premises) != 1 {
				t.Fatalf("Expected 1 premise, got %d", len(clause.Premises))
			}

			tempLit, ok := clause.Premises[0].(ast.TemporalLiteral)
			if !ok {
				t.Fatalf("Premise is %T, want TemporalLiteral", clause.Premises[0])
			}

			if tempLit.Operator == nil {
				t.Fatal("TemporalLiteral.Operator is nil")
			}

			if tempLit.Operator.Type != test.wantOpType {
				t.Errorf("Operator.Type = %v, want %v", tempLit.Operator.Type, test.wantOpType)
			}
		})
	}
}

func TestIntegration_TemporalStoreAndQuery(t *testing.T) {
	// Create a temporal store
	store := factstore.NewTemporalStore()

	// Add temporal facts
	aliceActive := ast.NewAtom("active", name("/alice"))
	bobActive := ast.NewAtom("active", name("/bob"))

	// Alice was active from Jan 1-15, 2024
	store.Add(aliceActive, makeInterval(
		time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
	))

	// Bob is active from Jan 10-31, 2024
	store.Add(bobActive, makeInterval(
		time.Date(2024, 1, 10, 0, 0, 0, 0, time.UTC),
		time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC),
	))

	tests := []struct {
		name       string
		evalTime   time.Time
		wantActive []string
	}{
		{
			name:       "Jan 5: only Alice active",
			evalTime:   time.Date(2024, 1, 5, 0, 0, 0, 0, time.UTC),
			wantActive: []string{"/alice"},
		},
		{
			name:       "Jan 12: both active",
			evalTime:   time.Date(2024, 1, 12, 0, 0, 0, 0, time.UTC),
			wantActive: []string{"/alice", "/bob"},
		},
		{
			name:       "Jan 20: only Bob active",
			evalTime:   time.Date(2024, 1, 20, 0, 0, 0, 0, time.UTC),
			wantActive: []string{"/bob"},
		},
		{
			name:       "Feb 1: no one active",
			evalTime:   time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC),
			wantActive: []string{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			te := NewTemporalEvaluator(store, test.evalTime)

			query := ast.NewAtom("active", ast.Variable{Symbol: "X"})
			tl := ast.TemporalLiteral{
				Literal:  query,
				Operator: nil,
			}

			solutions, err := te.EvalTemporalLiteral(tl, unionfind.New())
			if err != nil {
				t.Fatalf("EvalTemporalLiteral failed: %v", err)
			}

			// Extract the bound names
			var gotActive []string
			for _, sol := range solutions {
				term := sol.Get(ast.Variable{Symbol: "X"})
				if term != nil {
					if c, ok := term.(ast.Constant); ok {
						gotActive = append(gotActive, c.Symbol)
					}
				}
			}

			if len(gotActive) != len(test.wantActive) {
				t.Errorf("Got %d active, want %d: got %v, want %v",
					len(gotActive), len(test.wantActive), gotActive, test.wantActive)
			}
		})
	}
}

func TestIntegration_DiamondMinusQuery(t *testing.T) {
	// Create a temporal store
	store := factstore.NewTemporalStore()

	// Add a fact: employee was on_leave from Jan 5-10, 2024
	onLeave := ast.NewAtom("on_leave", name("/alice"))
	store.Add(onLeave, makeInterval(
		time.Date(2024, 1, 5, 0, 0, 0, 0, time.UTC),
		time.Date(2024, 1, 10, 0, 0, 0, 0, time.UTC),
	))

	tests := []struct {
		name       string
		evalTime   time.Time
		lookbackDays int
		wantMatch  bool
	}{
		{
			name:       "Jan 15: within 30 day lookback",
			evalTime:   time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
			lookbackDays: 30,
			wantMatch:  true,
		},
		{
			name:       "Jan 15: within 10 day lookback (barely)",
			evalTime:   time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
			lookbackDays: 10,
			wantMatch:  true,
		},
		{
			name:       "Jan 15: 3 day lookback misses it",
			evalTime:   time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
			lookbackDays: 3,
			wantMatch:  false,
		},
		{
			name:       "Feb 15: 30 day lookback misses it",
			evalTime:   time.Date(2024, 2, 15, 0, 0, 0, 0, time.UTC),
			lookbackDays: 30,
			wantMatch:  false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			te := NewTemporalEvaluator(store, test.evalTime)

			query := ast.NewAtom("on_leave", name("/alice"))
			operator := ast.TemporalOperator{
				Type: ast.DiamondMinus,
				Interval: ast.NewInterval(
					ast.TemporalBound{Type: ast.TimestampBound, Timestamp: -1}, // ~now
					ast.TemporalBound{Type: ast.TimestampBound, Timestamp: -int64(time.Duration(test.lookbackDays) * 24 * time.Hour)},
				),
			}

			tl := ast.TemporalLiteral{
				Literal:  query,
				Operator: &operator,
			}

			solutions, err := te.EvalTemporalLiteral(tl, unionfind.New())
			if err != nil {
				t.Fatalf("EvalTemporalLiteral failed: %v", err)
			}

			gotMatch := len(solutions) > 0
			if gotMatch != test.wantMatch {
				t.Errorf("Got match = %v, want %v", gotMatch, test.wantMatch)
			}
		})
	}
}

func TestIntegration_BoxMinusQuery(t *testing.T) {
	// Create a temporal store
	store := factstore.NewTemporalStore()

	// Add a fact: employee was employed from Jan 1, 2023 to Dec 31, 2024
	employed := ast.NewAtom("employed", name("/alice"))
	store.Add(employed, makeInterval(
		time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC),
	))

	tests := []struct {
		name         string
		evalTime     time.Time
		lookbackDays int
		wantMatch    bool
	}{
		{
			name:         "Mid-2024: employed continuously for last 365 days",
			evalTime:     time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC),
			lookbackDays: 365,
			wantMatch:    true,
		},
		{
			name:         "Early 2023: employed continuously for last 30 days",
			evalTime:     time.Date(2023, 2, 15, 0, 0, 0, 0, time.UTC),
			lookbackDays: 30,
			wantMatch:    true,
		},
		{
			name:         "Early 2023: NOT employed continuously for last 365 days",
			evalTime:     time.Date(2023, 2, 15, 0, 0, 0, 0, time.UTC),
			lookbackDays: 365,
			wantMatch:    false, // Alice wasn't employed before Jan 1, 2023
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			te := NewTemporalEvaluator(store, test.evalTime)

			query := ast.NewAtom("employed", name("/alice"))
			operator := ast.TemporalOperator{
				Type: ast.BoxMinus,
				Interval: ast.NewInterval(
					ast.TemporalBound{Type: ast.TimestampBound, Timestamp: -1}, // ~now
					ast.TemporalBound{Type: ast.TimestampBound, Timestamp: -int64(time.Duration(test.lookbackDays) * 24 * time.Hour)},
				),
			}

			tl := ast.TemporalLiteral{
				Literal:  query,
				Operator: &operator,
			}

			solutions, err := te.EvalTemporalLiteral(tl, unionfind.New())
			if err != nil {
				t.Fatalf("EvalTemporalLiteral failed: %v", err)
			}

			gotMatch := len(solutions) > 0
			if gotMatch != test.wantMatch {
				t.Errorf("Got match = %v, want %v", gotMatch, test.wantMatch)
			}
		})
	}
}

func TestIntegration_EternalFacts(t *testing.T) {
	// Create a temporal store
	store := factstore.NewTemporalStore()

	// Add an eternal fact (valid for all time)
	admin := ast.NewAtom("admin", name("/root"))
	store.AddEternal(admin)

	tests := []struct {
		name     string
		evalTime time.Time
		wantMatch bool
	}{
		{
			name:      "Past: eternal fact is valid",
			evalTime:  time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC),
			wantMatch: true,
		},
		{
			name:      "Present: eternal fact is valid",
			evalTime:  time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			wantMatch: true,
		},
		{
			name:      "Future: eternal fact is valid",
			evalTime:  time.Date(2100, 1, 1, 0, 0, 0, 0, time.UTC),
			wantMatch: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			te := NewTemporalEvaluator(store, test.evalTime)

			query := ast.NewAtom("admin", name("/root"))
			tl := ast.TemporalLiteral{
				Literal:  query,
				Operator: nil,
			}

			solutions, err := te.EvalTemporalLiteral(tl, unionfind.New())
			if err != nil {
				t.Fatalf("EvalTemporalLiteral failed: %v", err)
			}

			gotMatch := len(solutions) > 0
			if gotMatch != test.wantMatch {
				t.Errorf("Got match = %v, want %v", gotMatch, test.wantMatch)
			}
		})
	}
}

func TestIntegration_IntervalFunctions(t *testing.T) {
	// Test the interval extraction functions
	program := `
		# Extract interval components
		Decl status(X) temporal.
	`

	_, err := parse.Unit(strings.NewReader(program))
	if err != nil {
		t.Fatalf("Failed to parse program: %v", err)
	}

	// Test interval to constant conversion
	interval := makeInterval(
		time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC),
	)

	intervalConst := intervalToConstant(interval)

	// Verify it's a pair
	if intervalConst.Type != ast.PairShape {
		t.Errorf("intervalToConstant returned type %v, want PairShape", intervalConst.Type)
	}
}

func TestIntegration_BackwardCompatibility(t *testing.T) {
	// These programs should parse and run correctly without temporal features
	tests := []struct {
		name    string
		program string
	}{
		{
			name: "simple facts and rules",
			program: `
				node(/a).
				node(/b).
				edge(/a, /b).
				path(X, Y) :- edge(X, Y).
				path(X, Z) :- edge(X, Y), path(Y, Z).
			`,
		},
		{
			name: "negation",
			program: `
				all(/a).
				all(/b).
				excluded(/a).
				included(X) :- all(X), !excluded(X).
			`,
		},
		{
			name: "transforms",
			program: `
				item(1).
				item(2).
				item(3).
				total(N) :- item(X) |> do fn:group_by(), let N = fn:count().
			`,
		},
		{
			name: "comparisons",
			program: `
				age(/alice, 30).
				age(/bob, 25).
				adult(Name) :- age(Name, Age), Age >= 18.
			`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			unit, err := parse.Unit(strings.NewReader(test.program))
			if err != nil {
				t.Fatalf("Failed to parse program: %v", err)
			}

			// Verify no clauses have temporal annotations
			for i, clause := range unit.Clauses {
				if clause.HeadTime != nil && !clause.HeadTime.IsEternal() {
					t.Errorf("Clause %d has unexpected temporal annotation: %v", i, clause.HeadTime)
				}
			}
		})
	}
}

func TestIntegration_NowKeyword(t *testing.T) {
	tests := []struct {
		name    string
		program string
	}{
		{
			name:    "now as end bound",
			program: "active(/alice)@[2024-01-01, now].",
		},
		{
			name:    "now as point interval",
			program: "logged_in(/bob)@[now].",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			unit, err := parse.Unit(strings.NewReader(test.program))
			if err != nil {
				t.Fatalf("Failed to parse program: %v", err)
			}

			if len(unit.Clauses) != 1 {
				t.Fatalf("Expected 1 clause, got %d", len(unit.Clauses))
			}

			clause := unit.Clauses[0]
			if clause.HeadTime == nil {
				t.Fatal("Expected temporal annotation")
			}

			// Verify 'now' bound is present
			hasNow := clause.HeadTime.Start.Type == ast.NowBound ||
				clause.HeadTime.End.Type == ast.NowBound
			if !hasNow {
				t.Errorf("Expected 'now' bound in interval, got %v", clause.HeadTime)
			}
		})
	}
}

func TestIntegration_TemporalCoalesce(t *testing.T) {
	// Test interval coalescing for overlapping facts
	store := factstore.NewTemporalStore()

	status, _ := ast.Name("/active")

	// Add overlapping intervals
	store.Add(ast.NewAtom("status", status), makeInterval(
		time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
	))
	store.Add(ast.NewAtom("status", status), makeInterval(
		time.Date(2024, 1, 10, 0, 0, 0, 0, time.UTC),
		time.Date(2024, 1, 25, 0, 0, 0, 0, time.UTC),
	))
	store.Add(ast.NewAtom("status", status), makeInterval(
		time.Date(2024, 1, 20, 0, 0, 0, 0, time.UTC),
		time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC),
	))

	// Coalesce
	store.Coalesce(ast.PredicateSym{Symbol: "status", Arity: 1})

	// Query should return a single coalesced interval
	query := ast.NewAtom("status", status)
	var factCount int
	store.GetAllFacts(query, func(tf factstore.TemporalFact) error {
		factCount++
		// Verify the coalesced interval spans the full range
		if tf.Interval.Start.Type == ast.TimestampBound {
			startTime := time.Unix(0, tf.Interval.Start.Timestamp).UTC()
			expectedStart := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
			if !startTime.Equal(expectedStart) {
				t.Errorf("Coalesced start = %v, want %v", startTime, expectedStart)
			}
		}
		if tf.Interval.End.Type == ast.TimestampBound {
			endTime := time.Unix(0, tf.Interval.End.Timestamp).UTC()
			expectedEnd := time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC)
			if !endTime.Equal(expectedEnd) {
				t.Errorf("Coalesced end = %v, want %v", endTime, expectedEnd)
			}
		}
		return nil
	})

	if factCount != 1 {
		t.Errorf("After coalesce, got %d facts, want 1", factCount)
	}
}
