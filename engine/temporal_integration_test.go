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
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/google/mangle/analysis"
	"github.com/google/mangle/ast"
	"github.com/google/mangle/factstore"
	"github.com/google/mangle/parse"
	"github.com/google/mangle/unionfind"
)

// End-to-end integration tests for temporal reasoning in Mangle.
// These tests verify the complete flow from parsing to evaluation.

func TestIntegration_TemporalFactParsing(t *testing.T) {
	tests := []struct {
		name         string
		program      string
		wantFacts    int
		wantTemporal bool
	}{
		{
			name:         "simple temporal fact",
			program:      "foo(/bar)@[2024-01-15, 2024-06-30].",
			wantFacts:    1,
			wantTemporal: true,
		},
		{
			name:         "point interval fact",
			program:      "event(/login)@[2024-03-15].",
			wantFacts:    1,
			wantTemporal: true,
		},
		{
			name:         "non-temporal fact",
			program:      "regular(/fact).",
			wantFacts:    1,
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
		name       string
		program    string
		wantOpType ast.TemporalOperatorType
	}{
		{
			name:       "diamond minus operator",
			program:    "recently_active(X) :- <-[0s, 168h] active(X).",
			wantOpType: ast.DiamondMinus,
		},
		{
			name:       "box minus operator",
			program:    "stable(X) :- [-[0s, 720h] employed(X).",
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
	store.Add(aliceActive, ast.TimeInterval(
		ast.Date(2024, 1, 1),
		ast.Date(2024, 1, 15),
	))

	// Bob is active from Jan 10-31, 2024
	store.Add(bobActive, ast.TimeInterval(
		ast.Date(2024, 1, 10),
		ast.Date(2024, 1, 31),
	))

	tests := []struct {
		name       string
		evalTime   time.Time
		wantActive []string
	}{
		{
			name:       "Jan 5: only Alice active",
			evalTime:   ast.Date(2024, 1, 5),
			wantActive: []string{"/alice"},
		},
		{
			name:       "Jan 12: both active",
			evalTime:   ast.Date(2024, 1, 12),
			wantActive: []string{"/alice", "/bob"},
		},
		{
			name:       "Jan 20: only Bob active",
			evalTime:   ast.Date(2024, 1, 20),
			wantActive: []string{"/bob"},
		},
		{
			name:       "Feb 1: no one active",
			evalTime:   ast.Date(2024, 2, 1),
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
				Interval: nil,
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
	store.Add(onLeave, ast.TimeInterval(
		ast.Date(2024, 1, 5),
		ast.Date(2024, 1, 10),
	))

	tests := []struct {
		name         string
		evalTime     time.Time
		lookbackDays int
		wantMatch    bool
	}{
		{
			name:         "Jan 15: within 30 day lookback",
			evalTime:     ast.Date(2024, 1, 15),
			lookbackDays: 30,
			wantMatch:    true,
		},
		{
			name:         "Jan 15: within 10 day lookback (barely)",
			evalTime:     ast.Date(2024, 1, 15),
			lookbackDays: 10,
			wantMatch:    true,
		},
		{
			name:         "Jan 15: 3 day lookback misses it",
			evalTime:     ast.Date(2024, 1, 15),
			lookbackDays: 3,
			wantMatch:    false,
		},
		{
			name:         "Feb 15: 30 day lookback misses it",
			evalTime:     ast.Date(2024, 2, 15),
			lookbackDays: 30,
			wantMatch:    false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			te := NewTemporalEvaluator(store, test.evalTime)

			query := ast.NewAtom("on_leave", name("/alice"))
			operator := ast.TemporalOperator{
				Type: ast.DiamondMinus,
				Interval: ast.NewInterval(
					ast.NewDurationBound(0), // ~now
					ast.NewDurationBound(time.Duration(test.lookbackDays)*24*time.Hour),
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
	store.Add(employed, ast.TimeInterval(
		ast.Date(2023, 1, 1),
		ast.Date(2024, 12, 31),
	))

	tests := []struct {
		name         string
		evalTime     time.Time
		lookbackDays int
		wantMatch    bool
	}{
		{
			name:         "Mid-2024: employed continuously for last 365 days",
			evalTime:     ast.Date(2024, 6, 15),
			lookbackDays: 365,
			wantMatch:    true,
		},
		{
			name:         "Early 2023: employed continuously for last 30 days",
			evalTime:     ast.Date(2023, 2, 15),
			lookbackDays: 30,
			wantMatch:    true,
		},
		{
			name:         "Early 2023: NOT employed continuously for last 365 days",
			evalTime:     ast.Date(2023, 2, 15),
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
					ast.NewDurationBound(0), // ~now
					ast.NewDurationBound(time.Duration(test.lookbackDays)*24*time.Hour),
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
		name      string
		evalTime  time.Time
		wantMatch bool
	}{
		{
			name:      "Past: eternal fact is valid",
			evalTime:  ast.Date(1990, 1, 1),
			wantMatch: true,
		},
		{
			name:      "Present: eternal fact is valid",
			evalTime:  ast.Date(2024, 1, 1),
			wantMatch: true,
		},
		{
			name:      "Future: eternal fact is valid",
			evalTime:  ast.Date(2100, 1, 1),
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
				Interval: nil,
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
	interval := ast.TimeInterval(
		ast.Date(2024, 1, 1),
		ast.Date(2024, 12, 31),
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
	store.Add(ast.NewAtom("status", status), ast.TimeInterval(
		ast.Date(2024, 1, 1),
		ast.Date(2024, 1, 15),
	))
	store.Add(ast.NewAtom("status", status), ast.TimeInterval(
		ast.Date(2024, 1, 10),
		ast.Date(2024, 1, 25),
	))
	store.Add(ast.NewAtom("status", status), ast.TimeInterval(
		ast.Date(2024, 1, 20),
		ast.Date(2024, 1, 31),
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
			expectedStart := ast.Date(2024, 1, 1)
			if !startTime.Equal(expectedStart) {
				t.Errorf("Coalesced start = %v, want %v", startTime, expectedStart)
			}
		}
		if tf.Interval.End.Type == ast.TimestampBound {
			endTime := time.Unix(0, tf.Interval.End.Timestamp).UTC()
			expectedEnd := ast.Date(2024, 1, 31)
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

func TestIntegration_DurationBoundScenarios(t *testing.T) {
	// Setup: Reference time is 2024-01-20T12:00:00Z
	refTime := time.Date(2024, 1, 20, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name  string
		decls []string
		facts []string
		rules []string
		query string
		want  []string
	}{
		{
			name:  "diamond minus - within last 7 days",
			decls: []string{"Decl event(Action) temporal."},
			facts: []string{
				"event(/login) @[2024-01-15T10:00:00Z].",
				"event(/logout) @[2024-01-10T10:00:00Z].",
			},
			rules: []string{
				"recent(Action) :- <-[0s, 7d] event(Action).",
			},
			query: "recent(X)",
			want:  []string{"recent(/login)"},
		},
		{
			name:  "box minus - active for all of last 24 hours",
			decls: []string{"Decl active(S) temporal."},
			facts: []string{
				"active(/srv1) @[2024-01-19T00:00:00Z, 2024-01-21T00:00:00Z].",
				"active(/srv2) @[2024-01-20T00:00:00Z, 2024-01-20T06:00:00Z].",
			},
			rules: []string{
				"stable(S) :- [-[0s, 24h] active(S).",
			},
			query: "stable(X)",
			want:  []string{"stable(/srv1)"},
		},
		{
			name:  "diamond plus - upcoming maintenance in next 2 days",
			decls: []string{"Decl maint(N) temporal."},
			facts: []string{
				"maint(/node1) @[2024-01-21T10:00:00Z].",
				"maint(/node2) @[2024-01-25T10:00:00Z].",
			},
			rules: []string{
				"warn(N) :- <+[0s, 2d] maint(N).",
			},
			query: "warn(X)",
			want:  []string{"warn(/node1)"},
		},
		{
			name:  "box plus - reserved for next 1 hour",
			decls: []string{"Decl reserved(C) temporal."},
			facts: []string{
				"reserved(/cpu1) @[2024-01-20T12:00:00Z, 2024-01-20T14:00:00Z].",
				"reserved(/cpu2) @[2024-01-20T12:00:00Z, 2024-01-20T12:30:00Z].",
			},
			rules: []string{
				"unavailable(C) :- [+[0s, 1h] reserved(C).",
			},
			query: "unavailable(X)",
			want:  []string{"unavailable(/cpu1)"},
		},
		{
			name:  "temporal recursion - reachability",
			decls: []string{"Decl reachable(X, Y) temporal.", "Decl link(X, Y) temporal."},
			facts: []string{
				"link(/a, /b)@[2024-01-01].",
				"link(/b, /c)@[2024-01-01].",
				"link(/c, /d)@[2024-01-02].",
			},
			rules: []string{
				"reachable(X, Y)@[T] :- link(X, Y)@[T].",
				"reachable(X, Z)@[T] :- reachable(X, Y)@[T], link(Y, Z)@[T].",
			},
			query: "reachable(X, Y)",
			want: []string{
				"reachable(/a,/b)",
				"reachable(/b,/c)",
				"reachable(/a,/c)",
				"reachable(/c,/d)",
			},
		},
		{
			name:  "mixing duration and now - recent session starting after Jan 18",
			decls: []string{"Decl session(U) temporal."},
			facts: []string{
				"session(/user1) @[2024-01-19T10:00:00Z, 2024-01-20T13:00:00Z].",
				"session(/user2) @[2024-01-10T10:00:00Z, 2024-01-12T10:00:00Z].",
			},
			rules: []string{
				// 1705536000000000000 is 2024-01-18T00:00:00Z in Unix nanoseconds
				// 1705752000000000000 is 2024-01-20T12:00:00Z (refTime)
				"current_session(U) :- session(U) @[TStart, TEnd], LimitStart = fn:time:from_unix_nanos(1705536000000000000), LimitEnd = fn:time:from_unix_nanos(1705752000000000000), :time:ge(TStart, LimitStart), :time:ge(TEnd, LimitEnd).",
			},
			query: "current_session(X)",
			want:  []string{"current_session(/user1)"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := factstore.NewTemporalStore()
			var allClauses []ast.Clause
			var allDecls []ast.Decl

			// Parse decls
			for _, d := range tt.decls {
				unit, err := parse.Unit(strings.NewReader(d))
				if err != nil {
					t.Fatalf("Failed to parse decl %q: %v", d, err)
				}
				allDecls = append(allDecls, unit.Decls...)
			}

			// Parse facts and add to store
			for _, f := range tt.facts {
				unit, err := parse.Unit(strings.NewReader(f))
				if err != nil {
					t.Fatalf("Failed to parse fact %q: %v", f, err)
				}
				clause := unit.Clauses[0]
				if clause.HeadTime == nil {
					store.AddEternal(clause.Head)
				} else {
					store.Add(clause.Head, *clause.HeadTime)
				}
				allClauses = append(allClauses, clause)
			}

			// Parse rules
			for _, r := range tt.rules {
				unit, err := parse.Unit(strings.NewReader(r))
				if err != nil {
					t.Fatalf("Failed to parse rule %q: %v", r, err)
				}
				allClauses = append(allClauses, unit.Clauses...)
			}

			// Run evaluation
			programInfo, err := analysis.AnalyzeOneUnit(parse.SourceUnit{Decls: allDecls, Clauses: allClauses}, nil)
			if err != nil {
				t.Fatalf("Analysis failed: %v", err)
			}

			err = EvalProgram(programInfo, factstore.NewTemporalFactStoreAdapter(store),
				WithTemporalStore(store), WithEvaluationTime(refTime))
			if err != nil {
				t.Fatalf("Evaluation failed: %v", err)
			}

			// Query results
			queryAtom, err := parse.Atom(tt.query)
			if err != nil {
				t.Fatalf("Failed to parse query %q: %v", tt.query, err)
			}

			var got []string
			store.GetAllFacts(queryAtom, func(tf factstore.TemporalFact) error {
				got = append(got, tf.Atom.String())
				return nil
			})

			sort.Strings(got)
			sort.Strings(tt.want)

			if len(got) != len(tt.want) {
				t.Errorf("Got %d results, want %d. Got: %v, Want: %v", len(got), len(tt.want), got, tt.want)
			} else {
				for i := range got {
					if got[i] != tt.want[i] {
						t.Errorf("Result %d: got %q, want %q", i, got[i], tt.want[i])
					}
				}
			}
		})
	}
}
