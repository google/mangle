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
	"fmt"
	"time"

	"github.com/google/mangle/ast"
)

// IndexedTemporalStore is an implementation of TemporalFactStore that uses
// an augmented interval tree (AVL-based) for efficient point and range queries.
// It provides O(log n + k) query performance where n is the number of intervals
// and k is the number of matching results.
type IndexedTemporalStore struct {
	// facts maps predicate -> atom hash -> interval tree
	facts map[ast.PredicateSym]map[uint64]*IntervalTree
	// atoms maps hash to atom for reconstruction
	atoms               map[uint64]ast.Atom
	count               int
	maxIntervalsPerAtom int
}

var _ TemporalFactStore = &IndexedTemporalStore{}

// NewIndexedTemporalStore creates a new IndexedTemporalStore.
func NewIndexedTemporalStore(opts ...TemporalStoreOption) *IndexedTemporalStore {
	// Create a temporary TemporalStore just to apply options
	tempStore := &TemporalStore{maxIntervalsPerAtom: DefaultMaxIntervalsPerAtom}
	for _, opt := range opts {
		opt(tempStore)
	}

	return &IndexedTemporalStore{
		facts:               make(map[ast.PredicateSym]map[uint64]*IntervalTree),
		atoms:               make(map[uint64]ast.Atom),
		maxIntervalsPerAtom: tempStore.maxIntervalsPerAtom,
	}
}

// Add adds a temporal fact to the store.
func (s *IndexedTemporalStore) Add(atom ast.Atom, interval ast.Interval) (bool, error) {
	hash := atom.Hash()

	// Store the atom
	s.atoms[hash] = atom

	// Get or create the predicate map
	predMap, ok := s.facts[atom.Predicate]
	if !ok {
		predMap = make(map[uint64]*IntervalTree)
		s.facts[atom.Predicate] = predMap
	}

	// Get or create the interval tree
	tree, ok := predMap[hash]
	if !ok {
		tree = NewIntervalTree()
		predMap[hash] = tree
	}

	// Check interval limit before inserting
	if s.maxIntervalsPerAtom > 0 && tree.Size() >= s.maxIntervalsPerAtom {
		return false, fmt.Errorf("%w: maximum %d intervals per atom", ErrIntervalLimitExceeded, s.maxIntervalsPerAtom)
	}

	// Insert returns false if duplicate
	if !tree.Insert(interval) {
		return false, nil
	}

	s.count++
	return true, nil
}

// AddEternal adds a fact valid for all time.
func (s *IndexedTemporalStore) AddEternal(atom ast.Atom) (bool, error) {
	return s.Add(atom, ast.EternalInterval())
}

// GetFactsAt returns facts valid at a specific point in time.
func (s *IndexedTemporalStore) GetFactsAt(query ast.Atom, t time.Time, fn func(TemporalFact) error) error {
	predMap, ok := s.facts[query.Predicate]
	if !ok {
		return nil
	}

	timestamp := t.UnixNano()

	for hash, tree := range predMap {
		atom := s.atoms[hash]
		if !Matches(query.Args, atom.Args) {
			continue
		}

		err := tree.QueryPoint(timestamp, func(interval ast.Interval) error {
			return fn(TemporalFact{Atom: atom, Interval: interval})
		})
		if err != nil {
			return err
		}
	}
	return nil
}

// GetFactsDuring returns facts that overlap with the given interval.
func (s *IndexedTemporalStore) GetFactsDuring(query ast.Atom, interval ast.Interval, fn func(TemporalFact) error) error {
	predMap, ok := s.facts[query.Predicate]
	if !ok {
		return nil
	}

	start := getStartTime(interval)
	end := getEndTime(interval)

	for hash, tree := range predMap {
		atom := s.atoms[hash]
		if !Matches(query.Args, atom.Args) {
			continue
		}

		err := tree.QueryRange(start, end, func(factInterval ast.Interval) error {
			return fn(TemporalFact{Atom: atom, Interval: factInterval})
		})
		if err != nil {
			return err
		}
	}
	return nil
}

// GetAllFacts returns all facts matching the query with their intervals.
func (s *IndexedTemporalStore) GetAllFacts(query ast.Atom, fn func(TemporalFact) error) error {
	predMap, ok := s.facts[query.Predicate]
	if !ok {
		return nil
	}

	for hash, tree := range predMap {
		atom := s.atoms[hash]
		if !Matches(query.Args, atom.Args) {
			continue
		}

		err := tree.All(func(interval ast.Interval) error {
			return fn(TemporalFact{Atom: atom, Interval: interval})
		})
		if err != nil {
			return err
		}
	}
	return nil
}

// ContainsAt returns true if the atom is valid at the given time.
func (s *IndexedTemporalStore) ContainsAt(atom ast.Atom, t time.Time) bool {
	predMap, ok := s.facts[atom.Predicate]
	if !ok {
		return false
	}

	tree, ok := predMap[atom.Hash()]
	if !ok {
		return false
	}

	found := false
	timestamp := t.UnixNano()
	tree.QueryPoint(timestamp, func(interval ast.Interval) error {
		found = true
		return nil
	})
	return found
}

// ListPredicates returns all predicates in the store.
func (s *IndexedTemporalStore) ListPredicates() []ast.PredicateSym {
	result := make([]ast.PredicateSym, 0, len(s.facts))
	for pred := range s.facts {
		result = append(result, pred)
	}
	return result
}

// EstimateFactCount returns the number of temporal facts.
func (s *IndexedTemporalStore) EstimateFactCount() int {
	return s.count
}

// Coalesce merges adjacent or overlapping intervals for the same fact.
func (s *IndexedTemporalStore) Coalesce(predicate ast.PredicateSym) error {
	predMap, ok := s.facts[predicate]
	if !ok {
		return nil
	}

	for hash, tree := range predMap {
		if tree.Size() <= 1 {
			continue
		}

		// Collect all intervals
		var intervals []ast.Interval
		tree.All(func(interval ast.Interval) error {
			intervals = append(intervals, interval)
			return nil
		})

		// Coalesce them
		coalesced := coalesceIntervals(intervals)

		// Rebuild the tree with coalesced intervals
		s.count -= len(intervals) - len(coalesced)
		tree.Rebuild(coalesced)
		predMap[hash] = tree
	}

	return nil
}

// Merge merges contents of another temporal store into this one.
func (s *IndexedTemporalStore) Merge(other ReadOnlyTemporalFactStore) error {
	var mergeErr error
	for _, pred := range other.ListPredicates() {
		query := ast.NewQuery(pred)
		other.GetAllFacts(query, func(tf TemporalFact) error {
			_, err := s.Add(tf.Atom, tf.Interval)
			if err != nil {
				mergeErr = err
				return err
			}
			return nil
		})
		if mergeErr != nil {
			return mergeErr
		}
	}
	return nil
}
