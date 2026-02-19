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
	"fmt"

	"codeberg.org/TauCeti/mangle-go/ast"
	"codeberg.org/TauCeti/mangle-go/symbols"
	"codeberg.org/TauCeti/mangle-go/unionfind"
)

// DecideTemporalPredicate evaluates temporal interval predicates.
// Returns (ok, substitions, error). If ok is false with no error, the predicate failed.
func DecideTemporalPredicate(atom ast.Atom, subst *unionfind.UnionFind) (bool, []*unionfind.UnionFind, error) {
	if len(atom.Args) != 2 {
		return false, nil, fmt.Errorf("temporal predicate %s requires 2 arguments, got %d", atom.Predicate.Symbol, len(atom.Args))
	}

	// Extract intervals from arguments
	interval1, err := getIntervalValue(atom.Args[0])
	if err != nil {
		return false, nil, err
	}
	interval2, err := getIntervalValue(atom.Args[1])
	if err != nil {
		return false, nil, err
	}

	var result bool
	switch atom.Predicate.Symbol {
	case symbols.IntervalBefore.Symbol:
		result = intervalBefore(interval1, interval2)
	case symbols.IntervalAfter.Symbol:
		result = intervalAfter(interval1, interval2)
	case symbols.IntervalMeets.Symbol:
		result = intervalMeets(interval1, interval2)
	case symbols.IntervalOverlaps.Symbol:
		result = intervalOverlaps(interval1, interval2)
	case symbols.IntervalDuring.Symbol:
		result = intervalDuring(interval1, interval2)
	case symbols.IntervalContains.Symbol:
		result = intervalContains(interval1, interval2)
	case symbols.IntervalStarts.Symbol:
		result = intervalStarts(interval1, interval2)
	case symbols.IntervalFinishes.Symbol:
		result = intervalFinishes(interval1, interval2)
	case symbols.IntervalEquals.Symbol:
		result = intervalEquals(interval1, interval2)
	default:
		return false, nil, fmt.Errorf("unknown temporal predicate: %s", atom.Predicate.Symbol)
	}

	if result {
		return true, []*unionfind.UnionFind{subst}, nil
	}
	return false, nil, nil
}

// getIntervalValue extracts an interval from a constant.
// Intervals are represented as pairs of timestamps (start, end).
func getIntervalValue(term ast.BaseTerm) (ast.Interval, error) {
	c, ok := term.(ast.Constant)
	if !ok {
		return ast.Interval{}, fmt.Errorf("expected constant for interval, got %T", term)
	}

	// Intervals are stored as pairs of numbers (nanoseconds since epoch)
	if c.Type == ast.PairShape {
		fst, snd, err := c.PairValue()
		if err != nil {
			return ast.Interval{}, fmt.Errorf("invalid interval pair: %w", err)
		}
		startNano, err := fst.NumberValue()
		if err != nil {
			return ast.Interval{}, fmt.Errorf("invalid interval start: %w", err)
		}
		endNano, err := snd.NumberValue()
		if err != nil {
			return ast.Interval{}, fmt.Errorf("invalid interval end: %w", err)
		}
		return ast.Interval{
			Start: ast.TemporalBound{Type: ast.TimestampBound, Timestamp: startNano},
			End:   ast.TemporalBound{Type: ast.TimestampBound, Timestamp: endNano},
		}, nil
	}

	return ast.Interval{}, fmt.Errorf("expected pair for interval, got %v", c.Type)
}

// Allen's Interval Algebra implementations

// intervalBefore: T1 ends before T2 starts
func intervalBefore(t1, t2 ast.Interval) bool {
	if t1.End.Type != ast.TimestampBound || t2.Start.Type != ast.TimestampBound {
		return false
	}
	return t1.End.Timestamp < t2.Start.Timestamp
}

// intervalAfter: T1 starts after T2 ends
func intervalAfter(t1, t2 ast.Interval) bool {
	return intervalBefore(t2, t1)
}

// intervalMeets: T1 ends exactly when T2 starts
func intervalMeets(t1, t2 ast.Interval) bool {
	if t1.End.Type != ast.TimestampBound || t2.Start.Type != ast.TimestampBound {
		return false
	}
	return t1.End.Timestamp == t2.Start.Timestamp
}

// intervalOverlaps: T1 and T2 share some time
func intervalOverlaps(t1, t2 ast.Interval) bool {
	return t1.Overlaps(t2)
}

// intervalDuring: T1 is contained within T2
func intervalDuring(t1, t2 ast.Interval) bool {
	// T1.start >= T2.start AND T1.end <= T2.end
	if t1.Start.Type != ast.TimestampBound || t1.End.Type != ast.TimestampBound {
		return false
	}
	if t2.Start.Type != ast.TimestampBound || t2.End.Type != ast.TimestampBound {
		// Check for unbounded
		if t2.Start.Type == ast.NegativeInfinityBound {
			if t2.End.Type == ast.PositiveInfinityBound {
				return true // T2 is eternal
			}
			if t2.End.Type == ast.TimestampBound {
				return t1.End.Timestamp <= t2.End.Timestamp
			}
		}
		if t2.End.Type == ast.PositiveInfinityBound {
			if t2.Start.Type == ast.TimestampBound {
				return t1.Start.Timestamp >= t2.Start.Timestamp
			}
		}
		return false
	}
	return t1.Start.Timestamp >= t2.Start.Timestamp && t1.End.Timestamp <= t2.End.Timestamp
}

// intervalContains: T1 contains T2
func intervalContains(t1, t2 ast.Interval) bool {
	return intervalDuring(t2, t1)
}

// intervalStarts: T1 and T2 start at the same time
func intervalStarts(t1, t2 ast.Interval) bool {
	if t1.Start.Type != ast.TimestampBound || t2.Start.Type != ast.TimestampBound {
		if t1.Start.Type == ast.NegativeInfinityBound && t2.Start.Type == ast.NegativeInfinityBound {
			return true
		}
		if t1.Start.Type == ast.PositiveInfinityBound && t2.Start.Type == ast.PositiveInfinityBound {
			return true
		}
		return false
	}
	return t1.Start.Timestamp == t2.Start.Timestamp
}

// intervalFinishes: T1 and T2 end at the same time
func intervalFinishes(t1, t2 ast.Interval) bool {
	if t1.End.Type != ast.TimestampBound || t2.End.Type != ast.TimestampBound {
		if t1.End.Type == ast.NegativeInfinityBound && t2.End.Type == ast.NegativeInfinityBound {
			return true
		}
		if t1.End.Type == ast.PositiveInfinityBound && t2.End.Type == ast.PositiveInfinityBound {
			return true
		}
		return false
	}
	return t1.End.Timestamp == t2.End.Timestamp
}

// intervalEquals: T1 and T2 are identical
func intervalEquals(t1, t2 ast.Interval) bool {
	return t1.Equals(t2)
}

// IsTemporalPredicate returns true if the predicate is a temporal interval predicate.
func IsTemporalPredicate(pred ast.PredicateSym) bool {
	switch pred.Symbol {
	case symbols.IntervalBefore.Symbol,
		symbols.IntervalAfter.Symbol,
		symbols.IntervalMeets.Symbol,
		symbols.IntervalOverlaps.Symbol,
		symbols.IntervalDuring.Symbol,
		symbols.IntervalContains.Symbol,
		symbols.IntervalStarts.Symbol,
		symbols.IntervalFinishes.Symbol,
		symbols.IntervalEquals.Symbol:
		return true
	}
	return false
}
