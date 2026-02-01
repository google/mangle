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

package factstore

import (
	"errors"
	"testing"
	"time"

	"github.com/google/mangle/ast"
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

func TestSimpleTemporalStore_Add(t *testing.T) {
	store := NewSimpleTemporalStore()

	atom := ast.NewAtom("employed", name("/alice"))
	interval := makeInterval(
		time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2023, 12, 31, 0, 0, 0, 0, time.UTC),
	)

	// First add should succeed
	added, err := store.Add(atom, interval)
	if err != nil {
		t.Errorf("First Add returned error: %v", err)
	}
	if !added {
		t.Error("First Add should return true")
	}

	// Duplicate add should fail
	added, err = store.Add(atom, interval)
	if err != nil {
		t.Errorf("Duplicate Add returned error: %v", err)
	}
	if added {
		t.Error("Duplicate Add should return false")
	}

	// Adding same atom with different interval should succeed
	interval2 := makeInterval(
		time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC),
	)
	added, err = store.Add(atom, interval2)
	if err != nil {
		t.Errorf("Add with different interval returned error: %v", err)
	}
	if !added {
		t.Error("Add with different interval should return true")
	}

	if store.EstimateFactCount() != 2 {
		t.Errorf("EstimateFactCount = %d, want 2", store.EstimateFactCount())
	}
}

func TestSimpleTemporalStore_AddEternal(t *testing.T) {
	store := NewSimpleTemporalStore()

	atom := ast.NewAtom("admin", name("/bob"))

	added, err := store.AddEternal(atom)
	if err != nil {
		t.Errorf("AddEternal returned error: %v", err)
	}
	if !added {
		t.Error("AddEternal should return true")
	}

	// Query at any time should return the fact
	testTimes := []time.Time{
		time.Date(1900, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC),
		time.Date(3000, 12, 31, 0, 0, 0, 0, time.UTC),
	}

	for _, queryTime := range testTimes {
		if !store.ContainsAt(atom, queryTime) {
			t.Errorf("ContainsAt(%v) = false, want true for eternal fact", queryTime)
		}
	}
}

func TestSimpleTemporalStore_IntervalLimit(t *testing.T) {
	// Use a small custom limit for faster testing
	customLimit := 100
	store := NewSimpleTemporalStore(WithMaxIntervalsPerAtom(customLimit))
	atom := ast.NewAtom("test", name("/foo"))

	// Add intervals up to the limit
	for i := 0; i < customLimit; i++ {
		start := time.Date(2020, 1, 1, i, 0, 0, 0, time.UTC)
		end := time.Date(2020, 1, 1, i, 59, 59, 0, time.UTC)
		added, err := store.Add(atom, makeInterval(start, end))
		if err != nil {
			t.Fatalf("Add %d returned unexpected error: %v", i, err)
		}
		if !added {
			t.Fatalf("Add %d should return true", i)
		}
	}

	// Next add should fail with ErrIntervalLimitExceeded
	start := time.Date(2020, 1, 1, customLimit, 0, 0, 0, time.UTC)
	end := time.Date(2020, 1, 1, customLimit, 59, 59, 0, time.UTC)
	added, err := store.Add(atom, makeInterval(start, end))
	if !errors.Is(err, ErrIntervalLimitExceeded) {
		t.Errorf("Add beyond limit: err = %v, want ErrIntervalLimitExceeded", err)
	}
	if added {
		t.Error("Add beyond limit should return false")
	}
}

func TestSimpleTemporalStore_NoLimit(t *testing.T) {
	// Negative limit means no limit
	store := NewSimpleTemporalStore(WithMaxIntervalsPerAtom(-1))
	atom := ast.NewAtom("test", name("/foo"))

	// Should be able to add more than default limit
	for i := 0; i < DefaultMaxIntervalsPerAtom+10; i++ {
		start := time.Date(2020, 1, 1, 0, i, 0, 0, time.UTC)
		end := time.Date(2020, 1, 1, 0, i, 59, 0, time.UTC)
		_, err := store.Add(atom, makeInterval(start, end))
		if err != nil {
			t.Fatalf("Add %d returned unexpected error with no limit: %v", i, err)
		}
	}
}

func TestSimpleTemporalStore_GetFactsAt(t *testing.T) {
	store := NewSimpleTemporalStore()

	// Alice employed from 2020-2023
	aliceEmployed := ast.NewAtom("employed", name("/alice"))
	store.Add(aliceEmployed, makeInterval(
		time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2023, 12, 31, 0, 0, 0, 0, time.UTC),
	))

	// Bob employed from 2022 onwards
	bobEmployed := ast.NewAtom("employed", name("/bob"))
	store.Add(bobEmployed, makeInterval(
		time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2099, 12, 31, 0, 0, 0, 0, time.UTC),
	))

	tests := []struct {
		name      string
		queryTime time.Time
		wantCount int
	}{
		{
			name:      "before both",
			queryTime: time.Date(2019, 6, 15, 0, 0, 0, 0, time.UTC),
			wantCount: 0,
		},
		{
			name:      "alice only",
			queryTime: time.Date(2021, 6, 15, 0, 0, 0, 0, time.UTC),
			wantCount: 1,
		},
		{
			name:      "both employed",
			queryTime: time.Date(2022, 6, 15, 0, 0, 0, 0, time.UTC),
			wantCount: 2,
		},
		{
			name:      "bob only (after alice left)",
			queryTime: time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC),
			wantCount: 1,
		},
	}

	query := ast.NewQuery(ast.PredicateSym{Symbol: "employed", Arity: 1})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			count := 0
			err := store.GetFactsAt(query, tt.queryTime, func(tf TemporalFact) error {
				count++
				return nil
			})
			if err != nil {
				t.Fatalf("GetFactsAt error: %v", err)
			}
			if count != tt.wantCount {
				t.Errorf("GetFactsAt count = %d, want %d", count, tt.wantCount)
			}
		})
	}
}

func TestSimpleTemporalStore_GetFactsDuring(t *testing.T) {
	store := NewSimpleTemporalStore()

	// Event from Jan 1-15
	event1 := ast.NewAtom("event", name("/conference"))
	store.Add(event1, makeInterval(
		time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
	))

	// Event from Jan 20-30
	event2 := ast.NewAtom("event", name("/workshop"))
	store.Add(event2, makeInterval(
		time.Date(2024, 1, 20, 0, 0, 0, 0, time.UTC),
		time.Date(2024, 1, 30, 0, 0, 0, 0, time.UTC),
	))

	tests := []struct {
		name      string
		interval  ast.Interval
		wantCount int
	}{
		{
			name: "before all events",
			interval: makeInterval(
				time.Date(2023, 12, 1, 0, 0, 0, 0, time.UTC),
				time.Date(2023, 12, 31, 0, 0, 0, 0, time.UTC),
			),
			wantCount: 0,
		},
		{
			name: "overlaps first event",
			interval: makeInterval(
				time.Date(2024, 1, 10, 0, 0, 0, 0, time.UTC),
				time.Date(2024, 1, 18, 0, 0, 0, 0, time.UTC),
			),
			wantCount: 1,
		},
		{
			name: "overlaps both events",
			interval: makeInterval(
				time.Date(2024, 1, 10, 0, 0, 0, 0, time.UTC),
				time.Date(2024, 1, 25, 0, 0, 0, 0, time.UTC),
			),
			wantCount: 2,
		},
		{
			name: "between events (no overlap)",
			interval: makeInterval(
				time.Date(2024, 1, 16, 0, 0, 0, 0, time.UTC),
				time.Date(2024, 1, 19, 0, 0, 0, 0, time.UTC),
			),
			wantCount: 0,
		},
	}

	query := ast.NewQuery(ast.PredicateSym{Symbol: "event", Arity: 1})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			count := 0
			err := store.GetFactsDuring(query, tt.interval, func(tf TemporalFact) error {
				count++
				return nil
			})
			if err != nil {
				t.Fatalf("GetFactsDuring error: %v", err)
			}
			if count != tt.wantCount {
				t.Errorf("GetFactsDuring count = %d, want %d", count, tt.wantCount)
			}
		})
	}
}

func TestSimpleTemporalStore_Coalesce(t *testing.T) {
	store := NewSimpleTemporalStore()

	atom := ast.NewAtom("active", name("/service"))

	// Add overlapping intervals
	store.Add(atom, makeInterval(
		time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
	))
	store.Add(atom, makeInterval(
		time.Date(2024, 1, 10, 0, 0, 0, 0, time.UTC),
		time.Date(2024, 1, 25, 0, 0, 0, 0, time.UTC),
	))
	store.Add(atom, makeInterval(
		time.Date(2024, 1, 20, 0, 0, 0, 0, time.UTC),
		time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC),
	))

	// Before coalescing: 3 intervals
	if store.EstimateFactCount() != 3 {
		t.Errorf("Before coalesce: count = %d, want 3", store.EstimateFactCount())
	}

	// Coalesce
	pred := ast.PredicateSym{Symbol: "active", Arity: 1}
	if err := store.Coalesce(pred); err != nil {
		t.Fatalf("Coalesce error: %v", err)
	}

	// After coalescing: should be 1 interval covering Jan 1-31
	if store.EstimateFactCount() != 1 {
		t.Errorf("After coalesce: count = %d, want 1", store.EstimateFactCount())
	}

	// Verify the merged interval covers the full range
	query := ast.NewQuery(pred)
	var resultInterval ast.Interval
	store.GetAllFacts(query, func(tf TemporalFact) error {
		resultInterval = tf.Interval
		return nil
	})

	expectedStart := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	expectedEnd := time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC)

	// Compare in UTC to avoid timezone issues
	if !resultInterval.Start.Time().UTC().Equal(expectedStart) {
		t.Errorf("Coalesced start = %v, want %v", resultInterval.Start.Time().UTC(), expectedStart)
	}
	if !resultInterval.End.Time().UTC().Equal(expectedEnd) {
		t.Errorf("Coalesced end = %v, want %v", resultInterval.End.Time().UTC(), expectedEnd)
	}
}

func TestSimpleTemporalStore_CoalesceAdjacent(t *testing.T) {
	store := NewSimpleTemporalStore()

	atom := ast.NewAtom("shift", name("/worker"))

	// Add adjacent intervals (end of one = start of next - 1 nanosecond)
	store.Add(atom, makeInterval(
		time.Date(2024, 1, 1, 8, 0, 0, 0, time.UTC),
		time.Date(2024, 1, 1, 16, 0, 0, 0, time.UTC),
	))
	store.Add(atom, makeInterval(
		time.Date(2024, 1, 1, 16, 0, 0, 1, time.UTC), // 1 nanosecond after
		time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
	))

	pred := ast.PredicateSym{Symbol: "shift", Arity: 1}
	if err := store.Coalesce(pred); err != nil {
		t.Fatalf("Coalesce error: %v", err)
	}

	// Should be coalesced into 1 interval
	if store.EstimateFactCount() != 1 {
		t.Errorf("After coalesce: count = %d, want 1", store.EstimateFactCount())
	}
}

func TestSimpleTemporalStore_CoalesceNonOverlapping(t *testing.T) {
	store := NewSimpleTemporalStore()

	atom := ast.NewAtom("vacation", name("/alice"))

	// Add non-overlapping intervals
	store.Add(atom, makeInterval(
		time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2024, 1, 7, 0, 0, 0, 0, time.UTC),
	))
	store.Add(atom, makeInterval(
		time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2024, 6, 14, 0, 0, 0, 0, time.UTC),
	))

	pred := ast.PredicateSym{Symbol: "vacation", Arity: 1}
	if err := store.Coalesce(pred); err != nil {
		t.Fatalf("Coalesce error: %v", err)
	}

	// Should remain as 2 separate intervals
	if store.EstimateFactCount() != 2 {
		t.Errorf("After coalesce: count = %d, want 2", store.EstimateFactCount())
	}
}

func TestTemporalFactStoreAdapter(t *testing.T) {
	temporal := NewSimpleTemporalStore()

	// Add some temporal facts
	alice := ast.NewAtom("employed", name("/alice"))
	bob := ast.NewAtom("employed", name("/bob"))

	temporal.Add(alice, makeInterval(
		time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2023, 12, 31, 0, 0, 0, 0, time.UTC),
	))
	temporal.Add(bob, makeInterval(
		time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2099, 12, 31, 0, 0, 0, 0, time.UTC),
	))

	// Test adapter without time constraint
	adapter := NewTemporalFactStoreAdapter(temporal)

	query := ast.NewQuery(ast.PredicateSym{Symbol: "employed", Arity: 1})
	count := 0
	adapter.GetFacts(query, func(a ast.Atom) error {
		count++
		return nil
	})

	if count != 2 {
		t.Errorf("Adapter without time: count = %d, want 2", count)
	}

	// Test adapter with time constraint
	queryTime := time.Date(2021, 6, 15, 0, 0, 0, 0, time.UTC)
	adapterAt := NewTemporalFactStoreAdapterAt(temporal, queryTime)

	count = 0
	adapterAt.GetFacts(query, func(a ast.Atom) error {
		count++
		return nil
	})

	if count != 1 {
		t.Errorf("Adapter at 2021-06-15: count = %d, want 1 (only alice)", count)
	}
}

func TestTemporalFactStoreAdapter_Add(t *testing.T) {
	temporal := NewSimpleTemporalStore()
	adapter := NewTemporalFactStoreAdapter(temporal)

	atom := ast.NewAtom("admin", name("/charlie"))

	// Add through adapter should create eternal fact
	if !adapter.Add(atom) {
		t.Error("Adapter.Add should return true")
	}

	// Should be queryable at any time
	testTime := time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC)
	if !temporal.ContainsAt(atom, testTime) {
		t.Error("Fact added through adapter should be eternal")
	}
}

func TestSimpleTemporalStore_ListPredicates(t *testing.T) {
	store := NewSimpleTemporalStore()

	store.Add(ast.NewAtom("foo", name("/a")), ast.EternalInterval())
	store.Add(ast.NewAtom("bar", name("/b")), ast.EternalInterval())
	store.Add(ast.NewAtom("baz", name("/c")), ast.EternalInterval())

	preds := store.ListPredicates()
	if len(preds) != 3 {
		t.Errorf("ListPredicates returned %d predicates, want 3", len(preds))
	}

	// Check that all predicates are present
	found := make(map[string]bool)
	for _, p := range preds {
		found[p.Symbol] = true
	}

	for _, expected := range []string{"foo", "bar", "baz"} {
		if !found[expected] {
			t.Errorf("ListPredicates missing %s", expected)
		}
	}
}
