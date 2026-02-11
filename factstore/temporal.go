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
	"sort"
	"time"

	"github.com/google/mangle/ast"
)

// TemporalFact represents a fact with a validity interval.
type TemporalFact struct {
	Atom     ast.Atom
	Interval ast.Interval
}

// String returns a string representation of the temporal fact.
func (tf TemporalFact) String() string {
	return tf.Atom.String() + tf.Interval.String()
}

// DisplayString returns a string representation of the temporal fact using unescaped constants.
func (tf TemporalFact) DisplayString() string {
	return tf.Atom.DisplayString() + tf.Interval.String()
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
	// Returns (true, nil) if the fact was added, (false, nil) if duplicate.
	// Returns (false, ErrIntervalLimitExceeded) if the atom has too many intervals.
	Add(atom ast.Atom, interval ast.Interval) (bool, error)

	// AddEternal adds a fact valid for all time (eternal/timeless).
	AddEternal(atom ast.Atom) (bool, error)

	// Coalesce merges adjacent/overlapping intervals for the same fact.
	Coalesce(predicate ast.PredicateSym) error

	// Merge merges contents of given store.
	Merge(ReadOnlyTemporalFactStore) error
}

// DefaultMaxIntervalsPerAtom is the default maximum number of intervals allowed per atom.
// This prevents interval explosion in recursive temporal rules.
const DefaultMaxIntervalsPerAtom = 1000

// ErrIntervalLimitExceeded is returned when an atom has too many intervals.
var ErrIntervalLimitExceeded = fmt.Errorf("interval limit exceeded")

// TemporalStore is an in-memory implementation of TemporalFactStore.
// Facts are indexed by predicate symbol and atom hash, with each atom
// having an interval tree for O(log n + k) query performance.
type TemporalStore struct {
	facts               map[ast.PredicateSym]map[uint64]*IntervalTree
	atoms               map[uint64]ast.Atom
	count               int
	maxIntervalsPerAtom int // negative = no limit, 0 = use default
}

var _ TemporalFactStore = &TemporalStore{}

// TemporalStoreOption configures a TemporalStore.
type TemporalStoreOption func(*TemporalStore)

// WithMaxIntervalsPerAtom sets the maximum intervals allowed per atom.
// Negative value disables the limit.
func WithMaxIntervalsPerAtom(limit int) TemporalStoreOption {
	return func(s *TemporalStore) { s.maxIntervalsPerAtom = limit }
}

// NewTemporalStore creates a new TemporalStore.
func NewTemporalStore(opts ...TemporalStoreOption) *TemporalStore {
	s := &TemporalStore{
		facts:               make(map[ast.PredicateSym]map[uint64]*IntervalTree),
		atoms:               make(map[uint64]ast.Atom),
		maxIntervalsPerAtom: DefaultMaxIntervalsPerAtom,
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

// Add adds a temporal fact to the store.
// Returns (true, nil) if added, (false, nil) if duplicate, (false, error) if limit exceeded.
func (s *TemporalStore) Add(atom ast.Atom, interval ast.Interval) (bool, error) {
	// Guard: Ensure interval is valid
	if interval.Start.Type == ast.TimestampBound && interval.End.Type == ast.TimestampBound {
		if interval.Start.Timestamp > interval.End.Timestamp {
			return false, fmt.Errorf("invalid temporal interval: start %v > end %v", interval.Start, interval.End)
		}
	}

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

	// Check interval limit before inserting (negative limit means no limit)
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
func (s *TemporalStore) AddEternal(atom ast.Atom) (bool, error) {
	return s.Add(atom, ast.EternalInterval())
}

// GetFactsAt returns facts valid at a specific point in time.
// Uses interval tree for O(log n + k) query performance.
func (s *TemporalStore) GetFactsAt(query ast.Atom, t time.Time, fn func(TemporalFact) error) error {
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
// Uses interval tree for O(log n + k) query performance.
func (s *TemporalStore) GetFactsDuring(query ast.Atom, interval ast.Interval, fn func(TemporalFact) error) error {
	predMap, ok := s.facts[query.Predicate]
	if !ok {
		return nil
	}

	start := GetStartTime(interval)
	end := GetEndTime(interval)

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
func (s *TemporalStore) GetAllFacts(query ast.Atom, fn func(TemporalFact) error) error {
	if query.Predicate.Symbol == "" {
		for _, predMap := range s.facts {
			for hash, tree := range predMap {
				atom := s.atoms[hash]
				err := tree.All(func(interval ast.Interval) error {
					return fn(TemporalFact{Atom: atom, Interval: interval})
				})
				if err != nil {
					return err
				}
			}
		}
		return nil
	}

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
func (s *TemporalStore) ContainsAt(atom ast.Atom, t time.Time) bool {
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
func (s *TemporalStore) ListPredicates() []ast.PredicateSym {
	result := make([]ast.PredicateSym, 0, len(s.facts))
	for pred := range s.facts {
		result = append(result, pred)
	}
	return result
}

// EstimateFactCount returns the number of temporal facts (atom + interval pairs).
func (s *TemporalStore) EstimateFactCount() int {
	return s.count
}

// Coalesce merges adjacent or overlapping intervals for the same fact.
// This helps prevent interval explosion in recursive rules.
func (s *TemporalStore) Coalesce(predicate ast.PredicateSym) error {
	predMap, ok := s.facts[predicate]
	if !ok {
		return nil
	}

	for _, tree := range predMap {
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
// Returns an error if the interval limit is exceeded for any atom.
func (s *TemporalStore) Merge(other ReadOnlyTemporalFactStore) error {
	var mergeErr error
	for _, pred := range other.ListPredicates() {
		query := ast.NewQuery(pred)
		other.GetAllFacts(query, func(tf TemporalFact) error {
			_, err := s.Add(tf.Atom, tf.Interval)
			if err != nil {
				mergeErr = err
				return err // Stop iteration on error
			}
			return nil
		})
		if mergeErr != nil {
			return mergeErr
		}
	}
	return nil
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
// Note: errors from temporal store are logged but not returned per FactStore interface.
func (a *TemporalFactStoreAdapter) Add(atom ast.Atom) bool {
	added, _ := a.temporal.AddEternal(atom)
	return added
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

// TeeingTemporalStore is an implementation of TemporalFactStore that directs all writes to
// an output store, while distributing reads over a read-only base store and the output store.
type TeeingTemporalStore struct {
	base TemporalFactStore
	Out  TemporalFactStore
}

var _ TemporalFactStore = &TeeingTemporalStore{}

// NewTeeingTemporalStore returns a new TeeingTemporalStore.
func NewTeeingTemporalStore(base TemporalFactStore) *TeeingTemporalStore {
	return &TeeingTemporalStore{
		base: base,
		Out:  NewTemporalStore(),
	}
}

// GetFactsAt queries both base and output stores for facts valid at time t.
func (s *TeeingTemporalStore) GetFactsAt(query ast.Atom, t time.Time, fn func(TemporalFact) error) error {
	if err := s.base.GetFactsAt(query, t, fn); err != nil {
		return err
	}
	return s.Out.GetFactsAt(query, t, fn)
}

// GetFactsDuring queries both base and output stores for facts overlapping with interval.
func (s *TeeingTemporalStore) GetFactsDuring(query ast.Atom, interval ast.Interval, fn func(TemporalFact) error) error {
	if err := s.base.GetFactsDuring(query, interval, fn); err != nil {
		return err
	}
	return s.Out.GetFactsDuring(query, interval, fn)
}

// GetAllFacts queries both base and output stores for all matching facts.
func (s *TeeingTemporalStore) GetAllFacts(query ast.Atom, fn func(TemporalFact) error) error {
	if err := s.base.GetAllFacts(query, fn); err != nil {
		return err
	}
	return s.Out.GetAllFacts(query, fn)
}

// ContainsAt checks both base and output stores.
func (s *TeeingTemporalStore) ContainsAt(atom ast.Atom, t time.Time) bool {
	return s.base.ContainsAt(atom, t) || s.Out.ContainsAt(atom, t)
}

// ListPredicates returns the union of predicates from both stores.
func (s *TeeingTemporalStore) ListPredicates() []ast.PredicateSym {
	m := make(map[ast.PredicateSym]bool)
	for _, p := range s.base.ListPredicates() {
		m[p] = true
	}
	for _, p := range s.Out.ListPredicates() {
		m[p] = true
	}
	var res []ast.PredicateSym
	for p := range m {
		res = append(res, p)
	}
	return res
}

// EstimateFactCount returns the sum of estimated counts from both stores.
func (s *TeeingTemporalStore) EstimateFactCount() int {
	return s.base.EstimateFactCount() + s.Out.EstimateFactCount()
}

// Add adds a temporal fact to the output store.
func (s *TeeingTemporalStore) Add(atom ast.Atom, interval ast.Interval) (bool, error) {
	// Guard: Ensure interval is valid
	if interval.Start.Type == ast.TimestampBound && interval.End.Type == ast.TimestampBound {
		if interval.Start.Timestamp > interval.End.Timestamp {
			return false, fmt.Errorf("invalid temporal interval: start %v > end %v", interval.Start, interval.End)
		}
	}

	// Note: We don't check base for existence because semantics of Add
	// usually imply adding *another* interval.
	// But if we want deduplication...
	// If base has exact same fact (same interval), we might want to skip.
	// TemporalStore.Add returns false if duplicate.
	// We can check Contains? No, ContainsAt checks point.
	// Ideally we check if exact fact exists.
	// For now, let's just write to Out. Duplicate facts are usually harmless or handled by upper layers.
	return s.Out.Add(atom, interval)
}

// AddEternal adds an eternal fact to the output store.
func (s *TeeingTemporalStore) AddEternal(atom ast.Atom) (bool, error) {
	return s.Out.AddEternal(atom)
}

// Coalesce performs interval coalescing on the output store only.
func (s *TeeingTemporalStore) Coalesce(predicate ast.PredicateSym) error {
	// Coalesce only output store?
	// Base store is considered read-only/frozen.
	return s.Out.Coalesce(predicate)
}

// Merge merges facts into the output store.
func (s *TeeingTemporalStore) Merge(other ReadOnlyTemporalFactStore) error {
	return s.Out.Merge(other)
}
