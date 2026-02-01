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
	"testing"
	"time"

	"github.com/google/mangle/ast"
	"github.com/google/mangle/factstore"
	"github.com/google/mangle/unionfind"
)

func name(str string) ast.Constant {
	res, _ := ast.Name(str)
	return res
}

func makeInterval(start, end time.Time) ast.Interval {
	return ast.NewInterval(
		ast.NewTimestampBound(start),
		ast.NewTimestampBound(end),
	)
}

func TestTemporalEvaluator_DiamondMinus(t *testing.T) {
	store := factstore.NewTemporalStore()

	// Set up test data: employee was active from Jan 1 to Jan 15, 2024
	activeAtom := ast.NewAtom("active", name("/alice"))
	store.Add(activeAtom, makeInterval(
		time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
	))

	// Evaluation time: Jan 20, 2024
	evalTime := time.Date(2024, 1, 20, 0, 0, 0, 0, time.UTC)
	te := NewTemporalEvaluator(store, evalTime)

	tests := []struct {
		name        string
		operator    ast.TemporalOperator
		wantSolns   int
		description string
	}{
		{
			name: "within_range",
			// <-[0d, 30d] active(/alice) - was active sometime in last 30 days
			// The operator interval uses negative timestamps to represent durations
			// Start: 0 days ago (now), End: 30 days ago
			operator: ast.TemporalOperator{
				Type: ast.DiamondMinus,
				Interval: ast.NewInterval(
					ast.TemporalBound{Type: ast.TimestampBound, Timestamp: -1}, // ~0 days ago = now (use -1 to indicate duration)
					ast.TemporalBound{Type: ast.TimestampBound, Timestamp: -int64(30 * 24 * time.Hour)}, // 30 days ago
				),
			},
			wantSolns:   1,
			description: "Alice was active within the last 30 days",
		},
		{
			name: "outside_range",
			// <-[0d, 3d] active(/alice) - was active sometime in last 3 days
			operator: ast.TemporalOperator{
				Type: ast.DiamondMinus,
				Interval: ast.NewInterval(
					ast.TemporalBound{Type: ast.TimestampBound, Timestamp: -1}, // ~0 days ago
					ast.TemporalBound{Type: ast.TimestampBound, Timestamp: -int64(3 * 24 * time.Hour)},
				),
			},
			wantSolns:   0,
			description: "Alice was NOT active within the last 3 days (she left on Jan 15)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tl := ast.TemporalLiteral{
				Literal:  activeAtom,
				Operator: &tt.operator,
			}

			solutions, err := te.EvalTemporalLiteral(tl, unionfind.New())
			if err != nil {
				t.Fatalf("EvalTemporalLiteral error: %v", err)
			}

			if len(solutions) != tt.wantSolns {
				t.Errorf("%s: got %d solutions, want %d", tt.description, len(solutions), tt.wantSolns)
			}
		})
	}
}

func TestTemporalEvaluator_BoxMinus(t *testing.T) {
	store := factstore.NewTemporalStore()

	// Set up test data: service was running from Dec 1, 2023 to Feb 1, 2024
	runningAtom := ast.NewAtom("running", name("/service"))
	store.Add(runningAtom, makeInterval(
		time.Date(2023, 12, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC),
	))

	// Evaluation time: Jan 15, 2024
	evalTime := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	te := NewTemporalEvaluator(store, evalTime)

	tests := []struct {
		name        string
		operator    ast.TemporalOperator
		wantSolns   int
		description string
	}{
		{
			name: "continuously_true",
			// [-[0d, 30d] running(/service) - was running continuously for last 30 days
			operator: ast.TemporalOperator{
				Type: ast.BoxMinus,
				Interval: ast.NewInterval(
					ast.TemporalBound{Type: ast.TimestampBound, Timestamp: -0},
					ast.TemporalBound{Type: ast.TimestampBound, Timestamp: -int64(30 * 24 * time.Hour)},
				),
			},
			wantSolns:   1,
			description: "Service was running continuously for the last 30 days",
		},
		{
			name: "not_continuously_true",
			// [-[0d, 60d] running(/service) - was running continuously for last 60 days
			// This should fail because service started on Dec 1, which is only ~45 days ago
			operator: ast.TemporalOperator{
				Type: ast.BoxMinus,
				Interval: ast.NewInterval(
					ast.TemporalBound{Type: ast.TimestampBound, Timestamp: -0},
					ast.TemporalBound{Type: ast.TimestampBound, Timestamp: -int64(60 * 24 * time.Hour)},
				),
			},
			wantSolns:   0,
			description: "Service was NOT running for the full 60 days (only started Dec 1)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tl := ast.TemporalLiteral{
				Literal:  runningAtom,
				Operator: &tt.operator,
			}

			solutions, err := te.EvalTemporalLiteral(tl, unionfind.New())
			if err != nil {
				t.Fatalf("EvalTemporalLiteral error: %v", err)
			}

			if len(solutions) != tt.wantSolns {
				t.Errorf("%s: got %d solutions, want %d", tt.description, len(solutions), tt.wantSolns)
			}
		})
	}
}

func TestTemporalEvaluator_WithVariable(t *testing.T) {
	store := factstore.NewTemporalStore()

	// Add multiple employees with different active periods
	store.Add(ast.NewAtom("active", name("/alice")), makeInterval(
		time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
	))
	store.Add(ast.NewAtom("active", name("/bob")), makeInterval(
		time.Date(2024, 1, 10, 0, 0, 0, 0, time.UTC),
		time.Date(2024, 1, 25, 0, 0, 0, 0, time.UTC),
	))
	store.Add(ast.NewAtom("active", name("/charlie")), makeInterval(
		time.Date(2023, 6, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2023, 12, 31, 0, 0, 0, 0, time.UTC),
	))

	// Evaluation time: Jan 20, 2024
	evalTime := time.Date(2024, 1, 20, 0, 0, 0, 0, time.UTC)
	te := NewTemporalEvaluator(store, evalTime)

	// Query: <-[0d, 30d] active(X) - find who was active in last 30 days
	queryAtom := ast.NewAtom("active", ast.Variable{Symbol: "X"})
	operator := ast.TemporalOperator{
		Type: ast.DiamondMinus,
		Interval: ast.NewInterval(
			ast.TemporalBound{Type: ast.TimestampBound, Timestamp: -1}, // ~0 days ago
			ast.TemporalBound{Type: ast.TimestampBound, Timestamp: -int64(30 * 24 * time.Hour)},
		),
	}

	tl := ast.TemporalLiteral{
		Literal:  queryAtom,
		Operator: &operator,
	}

	solutions, err := te.EvalTemporalLiteral(tl, unionfind.New())
	if err != nil {
		t.Fatalf("EvalTemporalLiteral error: %v", err)
	}

	// Should find Alice, Bob, AND Charlie
	// (Alice: Jan 1-15, Bob: Jan 10-25, Charlie: Jun-Dec 31 2023)
	// At Jan 20, looking back 30 days (Dec 21 - Jan 20), all three overlap
	if len(solutions) != 3 {
		t.Errorf("Expected 3 solutions (Alice, Bob, and Charlie), got %d", len(solutions))
	}
}

func TestTemporalEvaluator_EternalFact(t *testing.T) {
	store := factstore.NewTemporalStore()

	// Add an eternal fact (valid for all time)
	store.AddEternal(ast.NewAtom("admin", name("/root")))

	evalTime := time.Date(2024, 1, 20, 0, 0, 0, 0, time.UTC)
	te := NewTemporalEvaluator(store, evalTime)

	// Query with any past operator should succeed for eternal facts
	queryAtom := ast.NewAtom("admin", name("/root"))
	operator := ast.TemporalOperator{
		Type: ast.DiamondMinus,
		Interval: ast.NewInterval(
			ast.TemporalBound{Type: ast.TimestampBound, Timestamp: -0},
			ast.TemporalBound{Type: ast.TimestampBound, Timestamp: -int64(365 * 24 * time.Hour)},
		),
	}

	tl := ast.TemporalLiteral{
		Literal:  queryAtom,
		Operator: &operator,
	}

	solutions, err := te.EvalTemporalLiteral(tl, unionfind.New())
	if err != nil {
		t.Fatalf("EvalTemporalLiteral error: %v", err)
	}

	if len(solutions) != 1 {
		t.Errorf("Eternal fact should match any temporal query, got %d solutions", len(solutions))
	}
}

func TestTemporalEvaluator_NoOperator(t *testing.T) {
	store := factstore.NewTemporalStore()

	// Add a temporal fact
	store.Add(ast.NewAtom("status", name("/ok")), makeInterval(
		time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC),
	))

	tests := []struct {
		name      string
		evalTime  time.Time
		wantSolns int
	}{
		{
			name:      "within_validity",
			evalTime:  time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
			wantSolns: 1,
		},
		{
			name:      "before_validity",
			evalTime:  time.Date(2023, 12, 15, 0, 0, 0, 0, time.UTC),
			wantSolns: 0,
		},
		{
			name:      "after_validity",
			evalTime:  time.Date(2024, 2, 15, 0, 0, 0, 0, time.UTC),
			wantSolns: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			te := NewTemporalEvaluator(store, tt.evalTime)

			// Temporal literal without operator - just checks if fact is valid at eval time
			tl := ast.TemporalLiteral{
				Literal:  ast.NewAtom("status", name("/ok")),
				Operator: nil,
			}

			solutions, err := te.EvalTemporalLiteral(tl, unionfind.New())
			if err != nil {
				t.Fatalf("EvalTemporalLiteral error: %v", err)
			}

			if len(solutions) != tt.wantSolns {
				t.Errorf("At %v: got %d solutions, want %d", tt.evalTime, len(solutions), tt.wantSolns)
			}
		})
	}
}

func TestTemporalEvaluator_DiamondPlus(t *testing.T) {
	store := factstore.NewTemporalStore()

	// Set up test data: scheduled maintenance from Feb 1 to Feb 15, 2024
	maintenanceAtom := ast.NewAtom("maintenance", name("/server"))
	store.Add(maintenanceAtom, makeInterval(
		time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2024, 2, 15, 0, 0, 0, 0, time.UTC),
	))

	// Evaluation time: Jan 20, 2024
	evalTime := time.Date(2024, 1, 20, 0, 0, 0, 0, time.UTC)
	te := NewTemporalEvaluator(store, evalTime)

	tests := []struct {
		name        string
		operator    ast.TemporalOperator
		wantSolns   int
		description string
	}{
		{
			name: "within_future_range",
			// <+[0d, 30d] maintenance(/server) - will there be maintenance in the next 30 days?
			operator: ast.TemporalOperator{
				Type: ast.DiamondPlus,
				Interval: ast.NewInterval(
					ast.TemporalBound{Type: ast.TimestampBound, Timestamp: -1},
					ast.TemporalBound{Type: ast.TimestampBound, Timestamp: -int64(30 * 24 * time.Hour)},
				),
			},
			wantSolns:   1,
			description: "Maintenance is scheduled within the next 30 days",
		},
		{
			name: "outside_future_range",
			// <+[0d, 5d] maintenance(/server) - will there be maintenance in next 5 days?
			operator: ast.TemporalOperator{
				Type: ast.DiamondPlus,
				Interval: ast.NewInterval(
					ast.TemporalBound{Type: ast.TimestampBound, Timestamp: -1},
					ast.TemporalBound{Type: ast.TimestampBound, Timestamp: -int64(5 * 24 * time.Hour)},
				),
			},
			wantSolns:   0,
			description: "No maintenance in the next 5 days (starts Feb 1)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tl := ast.TemporalLiteral{
				Literal:  maintenanceAtom,
				Operator: &tt.operator,
			}

			solutions, err := te.EvalTemporalLiteral(tl, unionfind.New())
			if err != nil {
				t.Fatalf("EvalTemporalLiteral error: %v", err)
			}

			if len(solutions) != tt.wantSolns {
				t.Errorf("%s: got %d solutions, want %d", tt.description, len(solutions), tt.wantSolns)
			}
		})
	}
}

func TestTemporalEvaluator_BoxPlus(t *testing.T) {
	store := factstore.NewTemporalStore()

	// Set up test data: contract valid from Jan 1, 2024 to Dec 31, 2024
	contractAtom := ast.NewAtom("contract", name("/customer"))
	store.Add(contractAtom, makeInterval(
		time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC),
	))

	// Evaluation time: Jan 15, 2024
	evalTime := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	te := NewTemporalEvaluator(store, evalTime)

	tests := []struct {
		name        string
		operator    ast.TemporalOperator
		wantSolns   int
		description string
	}{
		{
			name: "continuously_true_future",
			// [+[0d, 30d] contract(/customer) - will contract be valid for all of next 30 days?
			operator: ast.TemporalOperator{
				Type: ast.BoxPlus,
				Interval: ast.NewInterval(
					ast.TemporalBound{Type: ast.TimestampBound, Timestamp: -1},
					ast.TemporalBound{Type: ast.TimestampBound, Timestamp: -int64(30 * 24 * time.Hour)},
				),
			},
			wantSolns:   1,
			description: "Contract is valid for the entire next 30 days",
		},
		{
			name: "not_continuously_true_future",
			// [+[0d, 365d] contract(/customer) - will contract be valid for all of next 365 days?
			operator: ast.TemporalOperator{
				Type: ast.BoxPlus,
				Interval: ast.NewInterval(
					ast.TemporalBound{Type: ast.TimestampBound, Timestamp: -1},
					ast.TemporalBound{Type: ast.TimestampBound, Timestamp: -int64(365 * 24 * time.Hour)},
				),
			},
			wantSolns:   0,
			description: "Contract expires before 365 days (ends Dec 31)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tl := ast.TemporalLiteral{
				Literal:  contractAtom,
				Operator: &tt.operator,
			}

			solutions, err := te.EvalTemporalLiteral(tl, unionfind.New())
			if err != nil {
				t.Fatalf("EvalTemporalLiteral error: %v", err)
			}

			if len(solutions) != tt.wantSolns {
				t.Errorf("%s: got %d solutions, want %d", tt.description, len(solutions), tt.wantSolns)
			}
		})
	}
}

func TestTemporalEvaluator_IntervalVariableBinding(t *testing.T) {
	store := factstore.NewTemporalStore()

	// Add a fact with a known interval
	startTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	endTime := time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC)
	store.Add(ast.NewAtom("event", name("/meeting")), makeInterval(startTime, endTime))

	evalTime := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	te := NewTemporalEvaluator(store, evalTime)

	// Query with interval variable binding
	intervalVar := ast.Variable{Symbol: "T"}
	tl := ast.TemporalLiteral{
		Literal:     ast.NewAtom("event", name("/meeting")),
		Operator:    nil,
		IntervalVar: &intervalVar,
	}

	solutions, err := te.EvalTemporalLiteral(tl, unionfind.New())
	if err != nil {
		t.Fatalf("EvalTemporalLiteral error: %v", err)
	}

	if len(solutions) != 1 {
		t.Fatalf("Expected 1 solution, got %d", len(solutions))
	}

	// Check that the interval variable was bound
	boundValue := solutions[0].Get(intervalVar)
	if boundValue == nil {
		t.Error("Interval variable T was not bound")
	}

	// The bound value should be a pair constant
	if c, ok := boundValue.(ast.Constant); ok {
		if c.Type != ast.PairShape {
			t.Errorf("Expected pair constant for interval, got %v", c.Type)
		}
	}
}

func TestIntervalContains(t *testing.T) {
	te := &TemporalEvaluator{}

	jan1 := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	jan15 := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	jan10 := time.Date(2024, 1, 10, 0, 0, 0, 0, time.UTC)
	jan20 := time.Date(2024, 1, 20, 0, 0, 0, 0, time.UTC)
	jan31 := time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name string
		a    ast.Interval // container
		b    ast.Interval // contained
		want bool
	}{
		{
			name: "exact_match",
			a:    makeInterval(jan1, jan31),
			b:    makeInterval(jan1, jan31),
			want: true,
		},
		{
			name: "a_contains_b",
			a:    makeInterval(jan1, jan31),
			b:    makeInterval(jan10, jan20),
			want: true,
		},
		{
			name: "a_does_not_contain_b_start",
			a:    makeInterval(jan10, jan31),
			b:    makeInterval(jan1, jan20),
			want: false,
		},
		{
			name: "a_does_not_contain_b_end",
			a:    makeInterval(jan1, jan15),
			b:    makeInterval(jan10, jan31),
			want: false,
		},
		{
			name: "eternal_contains_any",
			a:    ast.EternalInterval(),
			b:    makeInterval(jan1, jan31),
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := te.intervalContains(tt.a, tt.b)
			if got != tt.want {
				t.Errorf("intervalContains(%v, %v) = %v, want %v", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

// Benchmarks

func BenchmarkTemporalEvaluator_DiamondMinus(b *testing.B) {
	store := factstore.NewTemporalStore()

	// Add many facts
	for i := 0; i < 1000; i++ {
		atomName, _ := ast.Name("/user" + string(rune('0'+i%10)))
		store.Add(ast.NewAtom("active", atomName), makeInterval(
			time.Date(2024, 1, 1+i%28, 0, 0, 0, 0, time.UTC),
			time.Date(2024, 1, 15+i%15, 0, 0, 0, 0, time.UTC),
		))
	}

	evalTime := time.Date(2024, 1, 20, 0, 0, 0, 0, time.UTC)
	te := NewTemporalEvaluator(store, evalTime)

	queryAtom := ast.NewAtom("active", ast.Variable{Symbol: "X"})
	operator := ast.TemporalOperator{
		Type: ast.DiamondMinus,
		Interval: ast.NewInterval(
			ast.TemporalBound{Type: ast.TimestampBound, Timestamp: -1},
			ast.TemporalBound{Type: ast.TimestampBound, Timestamp: -int64(30 * 24 * time.Hour)},
		),
	}

	tl := ast.TemporalLiteral{
		Literal:  queryAtom,
		Operator: &operator,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := te.EvalTemporalLiteral(tl, unionfind.New())
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkTemporalStore_GetFactsDuring(b *testing.B) {
	store := factstore.NewTemporalStore()

	// Add many facts
	for i := 0; i < 10000; i++ {
		atomName, _ := ast.Name("/item" + string(rune('0'+i%100)))
		store.Add(ast.NewAtom("available", atomName), makeInterval(
			time.Date(2024, 1, 1+i%28, 0, 0, 0, 0, time.UTC),
			time.Date(2024, 2, 1+i%28, 0, 0, 0, 0, time.UTC),
		))
	}

	query := ast.NewAtom("available", ast.Variable{Symbol: "X"})
	interval := makeInterval(
		time.Date(2024, 1, 10, 0, 0, 0, 0, time.UTC),
		time.Date(2024, 1, 20, 0, 0, 0, 0, time.UTC),
	)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		count := 0
		store.GetFactsDuring(query, interval, func(tf factstore.TemporalFact) error {
			count++
			return nil
		})
	}
}

func BenchmarkIntervalCoalesce(b *testing.B) {
	for i := 0; i < b.N; i++ {
		store := factstore.NewTemporalStore()

		// Add overlapping intervals
		pred, _ := ast.Name("/test")
		for j := 0; j < 100; j++ {
			store.Add(ast.NewAtom("status", pred), makeInterval(
				time.Date(2024, 1, 1+j, 0, 0, 0, 0, time.UTC),
				time.Date(2024, 1, 3+j, 0, 0, 0, 0, time.UTC),
			))
		}

		store.Coalesce(ast.PredicateSym{Symbol: "status", Arity: 1})
	}
}

// Tests for derived temporal facts (rules that produce temporal facts)

func TestResolveHeadTime_NilHeadTime(t *testing.T) {
	evalTime := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	subst := unionfind.New()

	result, err := ResolveHeadTime(nil, subst, evalTime)
	if err != nil {
		t.Fatalf("ResolveHeadTime(nil) returned error: %v", err)
	}
	if result != nil {
		t.Errorf("ResolveHeadTime(nil) = %v, want nil", result)
	}
}

func TestResolveHeadTime_TimestampBounds(t *testing.T) {
	evalTime := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	subst := unionfind.New()

	// Create an interval with timestamp bounds
	start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, 6, 30, 0, 0, 0, 0, time.UTC)
	headTime := ast.NewInterval(
		ast.NewTimestampBound(start),
		ast.NewTimestampBound(end),
	)

	result, err := ResolveHeadTime(&headTime, subst, evalTime)
	if err != nil {
		t.Fatalf("ResolveHeadTime returned error: %v", err)
	}
	if result == nil {
		t.Fatal("ResolveHeadTime returned nil")
	}
	if !result.Equals(headTime) {
		t.Errorf("ResolveHeadTime = %v, want %v", result, headTime)
	}
}

func TestResolveHeadTime_NowBound(t *testing.T) {
	evalTime := time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)
	subst := unionfind.New()

	// Create an interval with 'now' as the end bound
	start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	headTime := ast.NewInterval(
		ast.NewTimestampBound(start),
		ast.Now(), // 'now' bound
	)

	result, err := ResolveHeadTime(&headTime, subst, evalTime)
	if err != nil {
		t.Fatalf("ResolveHeadTime returned error: %v", err)
	}
	if result == nil {
		t.Fatal("ResolveHeadTime returned nil")
	}

	// The 'now' bound should be resolved to evalTime
	expectedEnd := ast.NewTimestampBound(evalTime)
	if result.End.Type != ast.TimestampBound || result.End.Timestamp != expectedEnd.Timestamp {
		t.Errorf("ResolveHeadTime end = %v, want %v", result.End, expectedEnd)
	}
}

func TestResolveHeadTime_VariableBound(t *testing.T) {
	evalTime := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)

	// Create a substitution with a bound variable
	subst := unionfind.New()
	startTime := time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC)
	endTime := time.Date(2024, 2, 28, 0, 0, 0, 0, time.UTC)

	// Bind T1 and T2 to timestamp values (as nanoseconds)
	t1Var := ast.Variable{Symbol: "T1"}
	t2Var := ast.Variable{Symbol: "T2"}
	subst, _ = unionfind.UnifyTermsExtend(
		[]ast.BaseTerm{t1Var},
		[]ast.BaseTerm{ast.Number(startTime.UnixNano())},
		subst,
	)
	subst, _ = unionfind.UnifyTermsExtend(
		[]ast.BaseTerm{t2Var},
		[]ast.BaseTerm{ast.Number(endTime.UnixNano())},
		subst,
	)

	// Create an interval with variable bounds
	headTime := ast.NewInterval(
		ast.NewVariableBound(t1Var),
		ast.NewVariableBound(t2Var),
	)

	result, err := ResolveHeadTime(&headTime, subst, evalTime)
	if err != nil {
		t.Fatalf("ResolveHeadTime returned error: %v", err)
	}
	if result == nil {
		t.Fatal("ResolveHeadTime returned nil")
	}

	// Check that the variables were resolved
	if result.Start.Type != ast.TimestampBound {
		t.Errorf("ResolveHeadTime start type = %v, want TimestampBound", result.Start.Type)
	}
	if result.Start.Timestamp != startTime.UnixNano() {
		t.Errorf("ResolveHeadTime start = %v, want %v", result.Start.Timestamp, startTime.UnixNano())
	}
	if result.End.Type != ast.TimestampBound {
		t.Errorf("ResolveHeadTime end type = %v, want TimestampBound", result.End.Type)
	}
	if result.End.Timestamp != endTime.UnixNano() {
		t.Errorf("ResolveHeadTime end = %v, want %v", result.End.Timestamp, endTime.UnixNano())
	}
}

func TestEvalClauseWithTemporalHead(t *testing.T) {
	evalTime := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name       string
		clause     ast.Clause
		solutions  []unionfind.UnionFind
		wantFacts  int
		wantHasInt bool // whether we expect intervals
	}{
		{
			name: "clause_without_temporal_head",
			clause: ast.Clause{
				Head:     ast.NewAtom("derived", ast.Variable{Symbol: "X"}),
				HeadTime: nil,
				Premises: []ast.Term{},
			},
			solutions: []unionfind.UnionFind{
				func() unionfind.UnionFind {
					s := unionfind.New()
					s, _ = unionfind.UnifyTermsExtend(
						[]ast.BaseTerm{ast.Variable{Symbol: "X"}},
						[]ast.BaseTerm{name("/alice")},
						s,
					)
					return s
				}(),
			},
			wantFacts:  1,
			wantHasInt: false,
		},
		{
			name: "clause_with_temporal_head_timestamp",
			clause: func() ast.Clause {
				interval := ast.NewInterval(
					ast.NewTimestampBound(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)),
					ast.NewTimestampBound(time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC)),
				)
				return ast.Clause{
					Head:     ast.NewAtom("active", ast.Variable{Symbol: "X"}),
					HeadTime: &interval,
					Premises: []ast.Term{},
				}
			}(),
			solutions: []unionfind.UnionFind{
				func() unionfind.UnionFind {
					s := unionfind.New()
					s, _ = unionfind.UnifyTermsExtend(
						[]ast.BaseTerm{ast.Variable{Symbol: "X"}},
						[]ast.BaseTerm{name("/bob")},
						s,
					)
					return s
				}(),
			},
			wantFacts:  1,
			wantHasInt: true,
		},
		{
			name: "clause_with_temporal_head_now",
			clause: func() ast.Clause {
				interval := ast.NewInterval(
					ast.NewTimestampBound(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)),
					ast.Now(),
				)
				return ast.Clause{
					Head:     ast.NewAtom("current", ast.Variable{Symbol: "X"}),
					HeadTime: &interval,
					Premises: []ast.Term{},
				}
			}(),
			solutions: []unionfind.UnionFind{
				func() unionfind.UnionFind {
					s := unionfind.New()
					s, _ = unionfind.UnifyTermsExtend(
						[]ast.BaseTerm{ast.Variable{Symbol: "X"}},
						[]ast.BaseTerm{name("/charlie")},
						s,
					)
					return s
				}(),
			},
			wantFacts:  1,
			wantHasInt: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			results, err := EvalClauseWithTemporalHead(test.clause, test.solutions, evalTime)
			if err != nil {
				t.Fatalf("EvalClauseWithTemporalHead returned error: %v", err)
			}

			if len(results) != test.wantFacts {
				t.Errorf("got %d facts, want %d", len(results), test.wantFacts)
			}

			for i, result := range results {
				hasInterval := result.Interval != nil
				if hasInterval != test.wantHasInt {
					t.Errorf("result[%d] has interval = %v, want %v", i, hasInterval, test.wantHasInt)
				}
			}
		})
	}
}
