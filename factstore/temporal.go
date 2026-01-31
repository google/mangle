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
	"sort"
	"time"

	"github.com/google/mangle/ast"
)

// TemporalFact represents a fact with a validity interval.
type TemporalFact struct {
	Atom     ast.Atom
	Interval ast.Interval
}

// ReadOnlyTemporalFactStore provides read access to temporal facts.
type ReadOnlyTemporalFactStore interface {
	// GetFactsAt returns facts valid at a specific point in time.
	GetFactsAt(query ast.Atom, t time.Time, fn func(TemporalFact) error) error

	// GetFactsDuring returns facts that overlap with the given interval.
	GetFactsDuring(query ast.Atom, interval ast.Interval, fn func(TemporalFact) error) error

	// GetAllFacts returns all facts (with their intervals) matching the query.
	GetAllFacts(query ast.Atom, fn func(TemporalFact) error) error

	// ContainsAt returns true if the atom is valid at the given time.
	ContainsAt(atom ast.Atom, t time.Time) bool

	// ListPredicates lists predicates available in this store.
	ListPredicates() []ast.PredicateSym

	// EstimateFactCount returns the estimated number of temporal facts.
	EstimateFactCount() int
}

// TemporalFactStore provides access to temporal facts.
type TemporalFactStore interface {
	ReadOnlyTemporalFactStore

	// Add adds a temporal fact to the store.
	// Returns true if the fact was added (not a duplicate).
	Add(atom ast.Atom, interval ast.Interval) bool

	// AddEternal adds a fact valid for all time (eternal/timeless).
	AddEternal(atom ast.Atom) bool

	// Coalesce merges adjacent/overlapping intervals for the same fact.
	Coalesce(predicate ast.PredicateSym) error

	// Merge merges contents of given store.
	Merge(ReadOnlyTemporalFactStore)
}

// temporalEntry stores intervals for a single fact (atom).
type temporalEntry struct {
	intervals []ast.Interval
}

// SimpleTemporalStore provides a simple in-memory implementation of TemporalFactStore.
// Facts are indexed by predicate symbol and atom hash, with each atom
// having a list of validity intervals.
type SimpleTemporalStore struct {
	// Map from predicate -> atom hash -> temporal entry
	facts map[ast.PredicateSym]map[uint64]*temporalEntry
	// Store the actual atoms by hash (for retrieval)
	atoms map[uint64]ast.Atom
	count int
}

// Ensure SimpleTemporalStore implements TemporalFactStore.
var _ TemporalFactStore = &SimpleTemporalStore{}

// NewSimpleTemporalStore creates a new SimpleTemporalStore.
func NewSimpleTemporalStore() *SimpleTemporalStore {
	return &SimpleTemporalStore{
		facts: make(map[ast.PredicateSym]map[uint64]*temporalEntry),
		atoms: make(map[uint64]ast.Atom),
		count: 0,
	}
}

// Add adds a temporal fact to the store.
func (s *SimpleTemporalStore) Add(atom ast.Atom, interval ast.Interval) bool {
	hash := atom.Hash()

	// Store the atom
	s.atoms[hash] = atom

	// Get or create the predicate map
	predMap, ok := s.facts[atom.Predicate]
	if !ok {
		predMap = make(map[uint64]*temporalEntry)
		s.facts[atom.Predicate] = predMap
	}

	// Get or create the temporal entry
	entry, ok := predMap[hash]
	if !ok {
		entry = &temporalEntry{intervals: make([]ast.Interval, 0, 1)}
		predMap[hash] = entry
	}

	// Check if this exact interval already exists
	for _, existing := range entry.intervals {
		if existing.Equals(interval) {
			return false
		}
	}

	// Add the interval
	entry.intervals = append(entry.intervals, interval)
	s.count++
	return true
}

// AddEternal adds a fact valid for all time.
func (s *SimpleTemporalStore) AddEternal(atom ast.Atom) bool {
	return s.Add(atom, ast.EternalInterval())
}

// GetFactsAt returns facts valid at a specific point in time.
func (s *SimpleTemporalStore) GetFactsAt(query ast.Atom, t time.Time, fn func(TemporalFact) error) error {
	predMap, ok := s.facts[query.Predicate]
	if !ok {
		return nil
	}

	for hash, entry := range predMap {
		atom := s.atoms[hash]
		if !Matches(query.Args, atom.Args) {
			continue
		}

		for _, interval := range entry.intervals {
			if interval.Contains(t) {
				if err := fn(TemporalFact{Atom: atom, Interval: interval}); err != nil {
					return err
				}
				// Don't break - report all matching intervals
			}
		}
	}
	return nil
}

// GetFactsDuring returns facts that overlap with the given interval.
func (s *SimpleTemporalStore) GetFactsDuring(query ast.Atom, interval ast.Interval, fn func(TemporalFact) error) error {
	predMap, ok := s.facts[query.Predicate]
	if !ok {
		return nil
	}

	for hash, entry := range predMap {
		atom := s.atoms[hash]
		if !Matches(query.Args, atom.Args) {
			continue
		}

		for _, factInterval := range entry.intervals {
			if factInterval.Overlaps(interval) {
				if err := fn(TemporalFact{Atom: atom, Interval: factInterval}); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// GetAllFacts returns all facts matching the query with their intervals.
func (s *SimpleTemporalStore) GetAllFacts(query ast.Atom, fn func(TemporalFact) error) error {
	predMap, ok := s.facts[query.Predicate]
	if !ok {
		return nil
	}

	for hash, entry := range predMap {
		atom := s.atoms[hash]
		if !Matches(query.Args, atom.Args) {
			continue
		}

		for _, interval := range entry.intervals {
			if err := fn(TemporalFact{Atom: atom, Interval: interval}); err != nil {
				return err
			}
		}
	}
	return nil
}

// ContainsAt returns true if the atom is valid at the given time.
func (s *SimpleTemporalStore) ContainsAt(atom ast.Atom, t time.Time) bool {
	predMap, ok := s.facts[atom.Predicate]
	if !ok {
		return false
	}

	entry, ok := predMap[atom.Hash()]
	if !ok {
		return false
	}

	for _, interval := range entry.intervals {
		if interval.Contains(t) {
			return true
		}
	}
	return false
}

// ListPredicates returns all predicates in the store.
func (s *SimpleTemporalStore) ListPredicates() []ast.PredicateSym {
	result := make([]ast.PredicateSym, 0, len(s.facts))
	for pred := range s.facts {
		result = append(result, pred)
	}
	return result
}

// EstimateFactCount returns the number of temporal facts (atom + interval pairs).
func (s *SimpleTemporalStore) EstimateFactCount() int {
	return s.count
}

// Coalesce merges adjacent or overlapping intervals for the same fact.
// This helps prevent interval explosion in recursive rules.
func (s *SimpleTemporalStore) Coalesce(predicate ast.PredicateSym) error {
	predMap, ok := s.facts[predicate]
	if !ok {
		return nil
	}

	for hash, entry := range predMap {
		if len(entry.intervals) <= 1 {
			continue
		}

		coalesced := coalesceIntervals(entry.intervals)
		s.count -= len(entry.intervals) - len(coalesced)
		predMap[hash].intervals = coalesced
	}
	return nil
}

// coalesceIntervals merges overlapping or adjacent intervals.
// Intervals must have concrete timestamps (not variables or unbounded).
func coalesceIntervals(intervals []ast.Interval) []ast.Interval {
	if len(intervals) <= 1 {
		return intervals
	}

	// Separate concrete intervals from those with variables/unbounded
	var concrete []ast.Interval
	var other []ast.Interval

	for _, i := range intervals {
		if i.Start.Type == ast.TimestampBound && i.End.Type == ast.TimestampBound {
			concrete = append(concrete, i)
		} else {
			other = append(other, i)
		}
	}

	if len(concrete) <= 1 {
		return append(concrete, other...)
	}

	// Sort by start time
	sort.Slice(concrete, func(i, j int) bool {
		return concrete[i].Start.Timestamp < concrete[j].Start.Timestamp
	})

	// Merge overlapping/adjacent intervals
	result := []ast.Interval{concrete[0]}
	for i := 1; i < len(concrete); i++ {
		last := &result[len(result)-1]
		curr := concrete[i]

		// Check if current overlaps or is adjacent to last
		// Adjacent means end of last + 1 nanosecond = start of current
		if last.End.Timestamp >= curr.Start.Timestamp-1 {
			// Merge: extend the end if needed
			if curr.End.Timestamp > last.End.Timestamp {
				last.End = curr.End
			}
		} else {
			result = append(result, curr)
		}
	}

	return append(result, other...)
}

// Merge merges contents of another temporal store into this one.
func (s *SimpleTemporalStore) Merge(other ReadOnlyTemporalFactStore) {
	for _, pred := range other.ListPredicates() {
		query := ast.NewQuery(pred)
		other.GetAllFacts(query, func(tf TemporalFact) error {
			s.Add(tf.Atom, tf.Interval)
			return nil
		})
	}
}

// TemporalFactStoreAdapter wraps a TemporalFactStore to provide
// a standard FactStore interface. Facts are returned without
// temporal information (using the underlying atom only).
// This allows temporal stores to be used where a regular FactStore is expected.
type TemporalFactStoreAdapter struct {
	temporal TemporalFactStore
	queryAt  *time.Time // If set, only return facts valid at this time
}

// Ensure TemporalFactStoreAdapter implements FactStore.
var _ FactStore = &TemporalFactStoreAdapter{}

// NewTemporalFactStoreAdapter creates an adapter that exposes all temporal facts
// as regular facts (ignoring time constraints).
func NewTemporalFactStoreAdapter(temporal TemporalFactStore) *TemporalFactStoreAdapter {
	return &TemporalFactStoreAdapter{temporal: temporal, queryAt: nil}
}

// NewTemporalFactStoreAdapterAt creates an adapter that only exposes facts
// valid at the specified time.
func NewTemporalFactStoreAdapterAt(temporal TemporalFactStore, t time.Time) *TemporalFactStoreAdapter {
	return &TemporalFactStoreAdapter{temporal: temporal, queryAt: &t}
}

// Add adds a fact as eternal (valid for all time).
func (a *TemporalFactStoreAdapter) Add(atom ast.Atom) bool {
	return a.temporal.AddEternal(atom)
}

// Contains returns true if the fact exists (respecting queryAt if set).
// Per the FactStore interface contract, errors are treated as "false".
// Clients who need to distinguish "absent" from "error" should use GetFacts.
func (a *TemporalFactStoreAdapter) Contains(atom ast.Atom) bool {
	if a.queryAt != nil {
		return a.temporal.ContainsAt(atom, *a.queryAt)
	}
	// Without a query time, check if the fact exists with any interval
	found := false
	_ = a.temporal.GetAllFacts(atom, func(tf TemporalFact) error {
		if tf.Atom.Equals(atom) {
			found = true
		}
		return nil
	})
	return found
}

// GetFacts returns facts matching the query (respecting queryAt if set).
func (a *TemporalFactStoreAdapter) GetFacts(query ast.Atom, fn func(ast.Atom) error) error {
	seen := make(map[uint64]bool)

	callback := func(tf TemporalFact) error {
		hash := tf.Atom.Hash()
		if seen[hash] {
			return nil
		}
		seen[hash] = true
		return fn(tf.Atom)
	}

	if a.queryAt != nil {
		return a.temporal.GetFactsAt(query, *a.queryAt, callback)
	}
	return a.temporal.GetAllFacts(query, callback)
}

// ListPredicates returns all predicates in the store.
func (a *TemporalFactStoreAdapter) ListPredicates() []ast.PredicateSym {
	return a.temporal.ListPredicates()
}

// EstimateFactCount returns an estimate of the number of facts.
func (a *TemporalFactStoreAdapter) EstimateFactCount() int {
	return a.temporal.EstimateFactCount()
}

// Merge merges another store into this one.
func (a *TemporalFactStoreAdapter) Merge(other ReadOnlyFactStore) {
	for _, pred := range other.ListPredicates() {
		other.GetFacts(ast.NewQuery(pred), func(fact ast.Atom) error {
			a.Add(fact)
			return nil
		})
	}
}
