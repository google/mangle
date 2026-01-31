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

package builtin

import (
	"testing"
	"time"

	"github.com/google/mangle/ast"
	"github.com/google/mangle/symbols"
	"github.com/google/mangle/unionfind"
)

func makeIntervalConstant(startNano, endNano int64) ast.Constant {
	startConst := ast.Number(startNano)
	endConst := ast.Number(endNano)
	return ast.Pair(&startConst, &endConst)
}

func TestIntervalBefore(t *testing.T) {
	jan1 := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC).UnixNano()
	jan5 := time.Date(2024, 1, 5, 0, 0, 0, 0, time.UTC).UnixNano()
	jan10 := time.Date(2024, 1, 10, 0, 0, 0, 0, time.UTC).UnixNano()
	jan15 := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC).UnixNano()

	tests := []struct {
		name string
		t1   ast.Constant
		t2   ast.Constant
		want bool
	}{
		{
			name: "before",
			t1:   makeIntervalConstant(jan1, jan5),
			t2:   makeIntervalConstant(jan10, jan15),
			want: true,
		},
		{
			name: "after",
			t1:   makeIntervalConstant(jan10, jan15),
			t2:   makeIntervalConstant(jan1, jan5),
			want: false,
		},
		{
			name: "overlapping",
			t1:   makeIntervalConstant(jan1, jan10),
			t2:   makeIntervalConstant(jan5, jan15),
			want: false,
		},
		{
			name: "meets",
			t1:   makeIntervalConstant(jan1, jan10),
			t2:   makeIntervalConstant(jan10, jan15),
			want: false, // meets is not before
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			atom := ast.NewAtom(symbols.IntervalBefore.Symbol, tt.t1, tt.t2)
			atom.Predicate = symbols.IntervalBefore
			subst := unionfind.New()

			ok, _, err := DecideTemporalPredicate(atom, &subst)
			if err != nil {
				t.Fatalf("DecideTemporalPredicate error: %v", err)
			}
			if ok != tt.want {
				t.Errorf("intervalBefore = %v, want %v", ok, tt.want)
			}
		})
	}
}

func TestIntervalOverlaps(t *testing.T) {
	jan1 := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC).UnixNano()
	jan5 := time.Date(2024, 1, 5, 0, 0, 0, 0, time.UTC).UnixNano()
	jan10 := time.Date(2024, 1, 10, 0, 0, 0, 0, time.UTC).UnixNano()
	jan15 := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC).UnixNano()
	jan20 := time.Date(2024, 1, 20, 0, 0, 0, 0, time.UTC).UnixNano()

	tests := []struct {
		name string
		t1   ast.Constant
		t2   ast.Constant
		want bool
	}{
		{
			name: "overlapping",
			t1:   makeIntervalConstant(jan1, jan10),
			t2:   makeIntervalConstant(jan5, jan15),
			want: true,
		},
		{
			name: "disjoint",
			t1:   makeIntervalConstant(jan1, jan5),
			t2:   makeIntervalConstant(jan10, jan15),
			want: false,
		},
		{
			name: "contained",
			t1:   makeIntervalConstant(jan1, jan20),
			t2:   makeIntervalConstant(jan5, jan15),
			want: true,
		},
		{
			name: "same",
			t1:   makeIntervalConstant(jan1, jan10),
			t2:   makeIntervalConstant(jan1, jan10),
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			atom := ast.NewAtom(symbols.IntervalOverlaps.Symbol, tt.t1, tt.t2)
			atom.Predicate = symbols.IntervalOverlaps
			subst := unionfind.New()

			ok, _, err := DecideTemporalPredicate(atom, &subst)
			if err != nil {
				t.Fatalf("DecideTemporalPredicate error: %v", err)
			}
			if ok != tt.want {
				t.Errorf("intervalOverlaps = %v, want %v", ok, tt.want)
			}
		})
	}
}

func TestIntervalDuring(t *testing.T) {
	jan1 := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC).UnixNano()
	jan5 := time.Date(2024, 1, 5, 0, 0, 0, 0, time.UTC).UnixNano()
	jan10 := time.Date(2024, 1, 10, 0, 0, 0, 0, time.UTC).UnixNano()
	jan15 := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC).UnixNano()
	jan20 := time.Date(2024, 1, 20, 0, 0, 0, 0, time.UTC).UnixNano()

	tests := []struct {
		name string
		t1   ast.Constant
		t2   ast.Constant
		want bool
	}{
		{
			name: "t1_during_t2",
			t1:   makeIntervalConstant(jan5, jan15),
			t2:   makeIntervalConstant(jan1, jan20),
			want: true,
		},
		{
			name: "t1_not_during_t2",
			t1:   makeIntervalConstant(jan1, jan20),
			t2:   makeIntervalConstant(jan5, jan15),
			want: false,
		},
		{
			name: "same",
			t1:   makeIntervalConstant(jan1, jan10),
			t2:   makeIntervalConstant(jan1, jan10),
			want: true,
		},
		{
			name: "partial_overlap",
			t1:   makeIntervalConstant(jan1, jan10),
			t2:   makeIntervalConstant(jan5, jan15),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			atom := ast.NewAtom(symbols.IntervalDuring.Symbol, tt.t1, tt.t2)
			atom.Predicate = symbols.IntervalDuring
			subst := unionfind.New()

			ok, _, err := DecideTemporalPredicate(atom, &subst)
			if err != nil {
				t.Fatalf("DecideTemporalPredicate error: %v", err)
			}
			if ok != tt.want {
				t.Errorf("intervalDuring = %v, want %v", ok, tt.want)
			}
		})
	}
}

func TestIntervalMeets(t *testing.T) {
	jan1 := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC).UnixNano()
	jan10 := time.Date(2024, 1, 10, 0, 0, 0, 0, time.UTC).UnixNano()
	jan15 := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC).UnixNano()

	tests := []struct {
		name string
		t1   ast.Constant
		t2   ast.Constant
		want bool
	}{
		{
			name: "meets",
			t1:   makeIntervalConstant(jan1, jan10),
			t2:   makeIntervalConstant(jan10, jan15),
			want: true,
		},
		{
			name: "gap",
			t1:   makeIntervalConstant(jan1, jan10-1),
			t2:   makeIntervalConstant(jan10, jan15),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			atom := ast.NewAtom(symbols.IntervalMeets.Symbol, tt.t1, tt.t2)
			atom.Predicate = symbols.IntervalMeets
			subst := unionfind.New()

			ok, _, err := DecideTemporalPredicate(atom, &subst)
			if err != nil {
				t.Fatalf("DecideTemporalPredicate error: %v", err)
			}
			if ok != tt.want {
				t.Errorf("intervalMeets = %v, want %v", ok, tt.want)
			}
		})
	}
}

func TestIntervalEquals(t *testing.T) {
	jan1 := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC).UnixNano()
	jan10 := time.Date(2024, 1, 10, 0, 0, 0, 0, time.UTC).UnixNano()
	jan15 := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC).UnixNano()

	tests := []struct {
		name string
		t1   ast.Constant
		t2   ast.Constant
		want bool
	}{
		{
			name: "equal",
			t1:   makeIntervalConstant(jan1, jan10),
			t2:   makeIntervalConstant(jan1, jan10),
			want: true,
		},
		{
			name: "not_equal",
			t1:   makeIntervalConstant(jan1, jan10),
			t2:   makeIntervalConstant(jan1, jan15),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			atom := ast.NewAtom(symbols.IntervalEquals.Symbol, tt.t1, tt.t2)
			atom.Predicate = symbols.IntervalEquals
			subst := unionfind.New()

			ok, _, err := DecideTemporalPredicate(atom, &subst)
			if err != nil {
				t.Fatalf("DecideTemporalPredicate error: %v", err)
			}
			if ok != tt.want {
				t.Errorf("intervalEquals = %v, want %v", ok, tt.want)
			}
		})
	}
}

func TestIsTemporalPredicate(t *testing.T) {
	tests := []struct {
		pred ast.PredicateSym
		want bool
	}{
		{symbols.IntervalBefore, true},
		{symbols.IntervalAfter, true},
		{symbols.IntervalMeets, true},
		{symbols.IntervalOverlaps, true},
		{symbols.IntervalDuring, true},
		{symbols.IntervalContains, true},
		{symbols.IntervalStarts, true},
		{symbols.IntervalFinishes, true},
		{symbols.IntervalEquals, true},
		{symbols.Lt, false},
		{symbols.Filter, false},
	}

	for _, tt := range tests {
		t.Run(tt.pred.Symbol, func(t *testing.T) {
			got := IsTemporalPredicate(tt.pred)
			if got != tt.want {
				t.Errorf("IsTemporalPredicate(%v) = %v, want %v", tt.pred, got, tt.want)
			}
		})
	}
}
