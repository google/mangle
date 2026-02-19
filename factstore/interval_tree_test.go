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
	"testing"
	"time"

	"codeberg.org/TauCeti/mangle-go/ast"
)

func TestIntervalTree_Insert(t *testing.T) {
	tree := NewIntervalTree()

	i1 := makeTestInterval(100, 200)
	i2 := makeTestInterval(150, 250)
	i3 := makeTestInterval(50, 75)

	// Insert first interval
	if !tree.Insert(i1) {
		t.Error("Expected first insert to succeed")
	}
	if tree.Size() != 1 {
		t.Errorf("Expected size 1, got %d", tree.Size())
	}

	// Insert duplicate
	if tree.Insert(i1) {
		t.Error("Expected duplicate insert to fail")
	}
	if tree.Size() != 1 {
		t.Errorf("Expected size still 1 after duplicate, got %d", tree.Size())
	}

	// Insert different intervals
	if !tree.Insert(i2) {
		t.Error("Expected second insert to succeed")
	}
	if !tree.Insert(i3) {
		t.Error("Expected third insert to succeed")
	}
	if tree.Size() != 3 {
		t.Errorf("Expected size 3, got %d", tree.Size())
	}
}

func TestIntervalTree_QueryPoint(t *testing.T) {
	tree := NewIntervalTree()

	// Insert intervals: [100, 200], [150, 250], [50, 75], [300, 400]
	tree.Insert(makeTestInterval(100, 200))
	tree.Insert(makeTestInterval(150, 250))
	tree.Insert(makeTestInterval(50, 75))
	tree.Insert(makeTestInterval(300, 400))

	tests := []struct {
		name      string
		timestamp int64
		wantCount int
	}{
		{"before all", 25, 0},
		{"in first interval only", 60, 1},
		{"at start of interval", 100, 1},
		{"in overlapping region", 175, 2},
		{"at end of first interval (overlaps with second)", 200, 2},
		{"between intervals", 275, 0},
		{"in last interval", 350, 1},
		{"after all", 500, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			count := 0
			tree.QueryPoint(tt.timestamp, func(interval ast.Interval) error {
				count++
				return nil
			})
			if count != tt.wantCount {
				t.Errorf("QueryPoint(%d) returned %d intervals, want %d", tt.timestamp, count, tt.wantCount)
			}
		})
	}
}

func TestIntervalTree_QueryRange(t *testing.T) {
	tree := NewIntervalTree()

	// Insert intervals: [100, 200], [150, 250], [50, 75], [300, 400]
	tree.Insert(makeTestInterval(100, 200))
	tree.Insert(makeTestInterval(150, 250))
	tree.Insert(makeTestInterval(50, 75))
	tree.Insert(makeTestInterval(300, 400))

	tests := []struct {
		name      string
		start     int64
		end       int64
		wantCount int
	}{
		{"before all", 0, 40, 0},
		{"overlap first", 60, 70, 1},
		{"overlap two", 175, 180, 2},
		{"overlap all middle", 100, 250, 2},
		{"gap between", 275, 290, 0},
		{"overlap last", 350, 375, 1},
		{"overlap all", 0, 500, 4},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			count := 0
			tree.QueryRange(tt.start, tt.end, func(interval ast.Interval) error {
				count++
				return nil
			})
			if count != tt.wantCount {
				t.Errorf("QueryRange(%d, %d) returned %d intervals, want %d", tt.start, tt.end, count, tt.wantCount)
			}
		})
	}
}

func TestIntervalTree_Balance(t *testing.T) {
	tree := NewIntervalTree()

	// Insert many intervals in sorted order (worst case for naive BST)
	for i := int64(0); i < 100; i++ {
		tree.Insert(makeTestInterval(i*10, i*10+5))
	}

	if tree.Size() != 100 {
		t.Errorf("Expected size 100, got %d", tree.Size())
	}

	// Tree should be balanced, so root height should be O(log n)
	// For 100 elements, height should be around 7-8 (log2(100) ≈ 6.6)
	if tree.root == nil {
		t.Fatal("Root should not be nil")
	}
	if tree.root.height > 10 {
		t.Errorf("Tree appears unbalanced: height %d for 100 elements", tree.root.height)
	}
}

func TestIntervalTree_All(t *testing.T) {
	tree := NewIntervalTree()

	intervals := []ast.Interval{
		makeTestInterval(100, 200),
		makeTestInterval(50, 75),
		makeTestInterval(150, 250),
	}

	for _, i := range intervals {
		tree.Insert(i)
	}

	count := 0
	tree.All(func(interval ast.Interval) error {
		count++
		return nil
	})

	if count != len(intervals) {
		t.Errorf("All() returned %d intervals, want %d", count, len(intervals))
	}
}

func TestIntervalTree_Clear(t *testing.T) {
	tree := NewIntervalTree()

	tree.Insert(makeTestInterval(100, 200))
	tree.Insert(makeTestInterval(50, 75))

	if tree.Size() != 2 {
		t.Errorf("Expected size 2 before clear, got %d", tree.Size())
	}

	tree.Clear()

	if tree.Size() != 0 {
		t.Errorf("Expected size 0 after clear, got %d", tree.Size())
	}
}

func TestIntervalTree_Rebuild(t *testing.T) {
	tree := NewIntervalTree()

	// Initial intervals
	tree.Insert(makeTestInterval(100, 200))
	tree.Insert(makeTestInterval(50, 75))

	// Rebuild with different intervals
	newIntervals := []ast.Interval{
		makeTestInterval(300, 400),
		makeTestInterval(500, 600),
		makeTestInterval(700, 800),
	}

	tree.Rebuild(newIntervals)

	if tree.Size() != 3 {
		t.Errorf("Expected size 3 after rebuild, got %d", tree.Size())
	}

	// Old intervals should be gone
	count := 0
	tree.QueryPoint(100, func(interval ast.Interval) error {
		count++
		return nil
	})
	if count != 0 {
		t.Error("Old interval should not be found after rebuild")
	}

	// New intervals should be present
	count = 0
	tree.QueryPoint(350, func(interval ast.Interval) error {
		count++
		return nil
	})
	if count != 1 {
		t.Error("New interval should be found after rebuild")
	}
}

func TestIntervalTree_EternalInterval(t *testing.T) {
	tree := NewIntervalTree()

	// Insert an eternal interval (negative to positive infinity)
	eternal := ast.EternalInterval()
	tree.Insert(eternal)

	// Should contain any timestamp
	tests := []int64{-1000000000000, 0, 1000000000000}
	for _, ts := range tests {
		count := 0
		tree.QueryPoint(ts, func(interval ast.Interval) error {
			count++
			return nil
		})
		if count != 1 {
			t.Errorf("Eternal interval should contain timestamp %d", ts)
		}
	}
}

func TestTemporalStore_WithIntervalTree(t *testing.T) {
	store := NewTemporalStore()

	atom := ast.Atom{
		Predicate: ast.PredicateSym{Symbol: "test", Arity: 1},
		Args:      []ast.BaseTerm{ast.String("alice")},
	}

	interval := makeTestInterval(100, 200)

	// Add fact
	added, err := store.Add(atom, interval)
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}
	if !added {
		t.Error("Expected Add to return true")
	}

	// Check count
	if store.EstimateFactCount() != 1 {
		t.Errorf("Expected count 1, got %d", store.EstimateFactCount())
	}

	// Query at valid time
	query := ast.NewQuery(atom.Predicate)
	foundAt := false
	store.GetFactsAt(query, time.Unix(0, 150), func(tf TemporalFact) error {
		foundAt = true
		return nil
	})
	if !foundAt {
		t.Error("Expected to find fact at valid time")
	}

	// Query at invalid time
	foundAt = false
	store.GetFactsAt(query, time.Unix(0, 50), func(tf TemporalFact) error {
		foundAt = true
		return nil
	})
	if foundAt {
		t.Error("Did not expect to find fact at invalid time")
	}
}

func TestTemporalStore_IntervalTreeLimit(t *testing.T) {
	store := NewTemporalStore(WithMaxIntervalsPerAtom(3))

	atom := ast.Atom{
		Predicate: ast.PredicateSym{Symbol: "test", Arity: 1},
		Args:      []ast.BaseTerm{ast.String("alice")},
	}

	// Add 3 intervals (should succeed)
	for i := int64(0); i < 3; i++ {
		_, err := store.Add(atom, makeTestInterval(i*100, i*100+50))
		if err != nil {
			t.Fatalf("Add %d failed: %v", i, err)
		}
	}

	// 4th should fail
	_, err := store.Add(atom, makeTestInterval(300, 350))
	if err == nil {
		t.Error("Expected error when exceeding interval limit")
	}
}

// makeInterval creates an interval from start and end nanoseconds.
func makeTestInterval(startNano, endNano int64) ast.Interval {
	return ast.NewInterval(
		ast.TemporalBound{Type: ast.TimestampBound, Timestamp: startNano},
		ast.TemporalBound{Type: ast.TimestampBound, Timestamp: endNano},
	)
}
