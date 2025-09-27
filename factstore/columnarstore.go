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
	"github.com/google/mangle/ast"
)

type ColumnarStore struct {
	tables map[ast.PredicateSym]*table
}

func NewColumnarStore() *ColumnarStore {
	return &ColumnarStore{
		tables: make(map[ast.PredicateSym]*table),
	}
}

// Add implements the FactStore interface by adding the fact to the backing map.
func (s *ColumnarStore) Add(a ast.Atom) bool {
	t, ok := s.tables[a.Predicate]
	if !ok {
		t = newTable(a.Predicate.Arity)
		s.tables[a.Predicate] = t
	}
	return t.add(a)
}

// Remove removes the fact from the backing map.
func (s *ColumnarStore) Remove(a ast.Atom) bool {
	t, ok := s.tables[a.Predicate]
	if ok {
		return t.remove(a)
	}
	return false
}

// GetFacts retrieves facts that match a given query.
func (s *ColumnarStore) GetFacts(a ast.Atom, fn func(ast.Atom) error) error {
	t, ok := s.tables[a.Predicate]
	if ok {
		return t.getFacts(a, fn)
	}
	return nil // should return error why coder doesn't return any error
}

// Contains checks if a fact is in the store.
func (s *ColumnarStore) Contains(a ast.Atom) bool {
	t, ok := s.tables[a.Predicate]
	if ok {
		return t.contains(a)
	}
	return false
}

// EstimateFactCount returns the estimated number of facts in the store.
func (s *ColumnarStore) EstimateFactCount() int {
	var count int
	for _, t := range s.tables {
		count += t.size()
	}
	return count
}

// Merge merges another fact store into this one.
func (s *ColumnarStore) Merge(other ReadOnlyFactStore) {
	// optimization if other = columnar? 
	// if other, ok := other.(*ColumnarStore); ok {
	// 	for pred, table := range other.tables {
	// 		s.tables[pred] = table
	// 	}
	// 	return
	// } FIXME: THIS IS WRONG
	for _, pred := range other.ListPredicates() {
		other.GetFacts(ast.NewQuery(pred), func(fact ast.Atom) error {
			s.Add(fact)
			return nil
		})
	}
}

// ListPredicates lists all predicates in the store.
func (s *ColumnarStore) ListPredicates() []ast.PredicateSym {
	preds := make([]ast.PredicateSym, 0, len(s.tables))
	for pred := range s.tables {
		preds = append(preds, pred)
	}
	return preds
}

// table stores all facts for a single predicate.
type table struct {
	facts   []ast.Atom
	primary map[uint64][]int   // Index for the entire atom hash.
	indices []map[uint64][]int // Index for each argument hash.
	free    []int              // A freelist of indices of removed facts.
}

func newTable(arity int) *table {
	indices := make([]map[uint64][]int, arity)
	for i := range indices {
		indices[i] = make(map[uint64][]int)
	}
	return &table{
		facts:   make([]ast.Atom, 0, 1024),
		primary: make(map[uint64][]int),
		indices: indices,
		free:    make([]int, 0, 16),
	}
}

func (t *table) size() int {
	return len(t.facts) - len(t.free)
}

func (t *table) add(a ast.Atom) bool {
	key := a.Hash()
	if candidates, ok := t.primary[key]; ok {
		for _, idx := range candidates {
			if !isNil(t.facts[idx]) && t.facts[idx].Equals(a) {
				return false // Already exists.
			}
		}
	}

	// maybe let lines 131 - 149 run concurrently at all?
	// a semaphore for table to control throughput?
	var idx int
	if len(t.free) > 0 {
		idx = t.free[len(t.free)-1]
		// add a lock & update free concurrently??
		// maybe try a pool of indices and update in intervals?
		// overhead for writes tbh
		t.free = t.free[:len(t.free)-1]
		t.facts[idx] = a
	} else {
		idx = len(t.facts)
		t.facts = append(t.facts, a)
	}

	t.primary[key] = append(t.primary[key], idx)
	for i, arg := range a.Args {
		h := arg.Hash()
		t.indices[i][h] = append(t.indices[i][h], idx)
	}
	return true
}

func (t *table) remove(a ast.Atom) bool {
	aHash := a.Hash()
	candidates, ok := t.primary[aHash]
	if !ok {
		return false
	}

	factIdx := -1
	candidateIdx := -1
	for i, idx := range candidates {
		if !isNil(t.facts[idx]) && t.facts[idx].Equals(a) {
			factIdx = idx
			candidateIdx = i
			break
		}
	}

	if factIdx == -1 {
		return false
	}

	candidates[candidateIdx] = candidates[len(candidates)-1]
	newCandidates := candidates[:len(candidates)-1]
	if len(newCandidates) == 0 {
		delete(t.primary, aHash)
	} else {
		t.primary[aHash] = newCandidates
	}

	// Remove from columnar indices.
	for i, arg := range a.Args {
		h := arg.Hash()
		list := t.indices[i][h]
		for j, idx := range list {
			if idx == factIdx {
				list[j] = list[len(list)-1]
				t.indices[i][h] = list[:len(list)-1]
				break
			}
		}
	}

	// Mark the slot as free.
	t.facts[factIdx] = ast.Atom{} // Zero out the atom.
	t.free = append(t.free, factIdx)
	return true
}

func (t *table) getFacts(a ast.Atom, fn func(ast.Atom) error) error {
	// Find the smallest set of candidate indices to check.
	var candidates []int
	var foundConst bool
	for i, arg := range a.Args {
		if _, isVar := arg.(ast.Variable); !isVar {
			h := arg.Hash()
			if list, ok := t.indices[i][h]; ok {
				if !foundConst || len(list) < len(candidates) {
					candidates = list
					foundConst = true
				}
			} else {
				// Hash not found, so no facts can match.
				return nil
			}
		}
	}

	if !foundConst { // No constants in query, scan all facts.
		for _, fact := range t.facts {
			if !isNil(fact) && Matches(a.Args, fact.Args) {
				if err := fn(fact); err != nil {
					return err
				}
			}
		}
		return nil
	}

	for _, idx := range candidates {
		fact := t.facts[idx]
		if !isNil(fact) && Matches(a.Args, fact.Args) {
			if err := fn(fact); err != nil {
				return err
			}
		}
	}
	return nil
}

func (t *table) contains(a ast.Atom) bool {
	aHash := a.Hash()
	if candidates, ok := t.primary[aHash]; ok {
		for _, idx := range candidates {
			if !isNil(t.facts[idx]) && t.facts[idx].Equals(a) {
				return true
			}
		}
	}
	return false
}

// isNil checks if an atom is the zero value.
func isNil(a ast.Atom) bool {
	return a.Predicate.Symbol == ""
}
