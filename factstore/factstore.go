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

// Package factstore contains the interface and a simple implementation for access
// to facts (atoms that are ground, i.e. contain no variables).
package factstore

import (
	"strings"

	"github.com/google/mangle/ast"
)

// ReadOnlyFactStore provides read access to a set of facts.
type ReadOnlyFactStore interface {
	// Returns a stream of facts that match a given atom. It takes a callback
	// to process results. If the callback returns an error, or it encounters
	// a malformed atom, scanning stops and that error is returned.
	GetFacts(ast.Atom, func(ast.Atom) error) error

	// Contains returns true if given atom is already present in store.
	// This is a convenience method that has a straightforward implementation
	// in terms of GetFacts. It does not return error and treats any
	// error condition as "false". Clients who distinguish "absent" from "error"
	// should call GetFacts directly.
	Contains(ast.Atom) bool

	// ListPredicates lists predicates available in this store.
	ListPredicates() []ast.PredicateSym

	// EstimateFactCount returns the estimated number of facts in the store.
	EstimateFactCount() int
}

// FactStore provides access to a set of facts.
type FactStore interface {
	ReadOnlyFactStore

	// Add adds an atom to a store and returns true if the fact didn't exist before.
	Add(ast.Atom) bool

	// Merge merges contents of given store.
	Merge(ReadOnlyFactStore)
}

// InMemoryStore provides a simple implementation backed by a map from each predicate sym to a T value.
type InMemoryStore[T any] struct {
	constants         map[ast.PredicateSym]ast.Atom
	shardsByPredicate map[ast.PredicateSym]T
}

// NewInMemoryStore constructs a new InMemoryStore.
func NewInMemoryStore[T any]() InMemoryStore[T] {
	return InMemoryStore[T]{
		make(map[ast.PredicateSym]ast.Atom),
		make(map[ast.PredicateSym]T),
	}
}

// SimpleInMemoryStore provides a simple implementation backed by a two-level map.
// For each predicate sym, we have a separate map, using numeric hash as key.
type SimpleInMemoryStore struct {
	InMemoryStore[map[uint64]ast.Atom]
}

// String returns a readable debug string for this store.
func (s SimpleInMemoryStore) String() string {
	var sb strings.Builder
	for _, m := range s.shardsByPredicate {
		for _, v := range m {
			sb.WriteString(v.String())
			sb.WriteRune(' ')
		}
		sb.WriteRune('\n')
	}
	return sb.String()
}

// ListPredicates returns a list of predicates.
func (s SimpleInMemoryStore) ListPredicates() []ast.PredicateSym {
	var r []ast.PredicateSym
	for p := range s.shardsByPredicate {
		r = append(r, p)
	}
	return r
}

// NewSimpleInMemoryStore constructs a new SimpleInMemoryStore.
func NewSimpleInMemoryStore() SimpleInMemoryStore {
	return SimpleInMemoryStore{
		NewInMemoryStore[map[uint64]ast.Atom](),
	}
}

// GetFacts implementation that looks up facts from an in-memory map.
func (s SimpleInMemoryStore) GetFacts(a ast.Atom, fn func(ast.Atom) error) error {
	for _, fact := range s.shardsByPredicate[a.Predicate] {
		if Matches(a.Args, fact.Args) {
			if err := fn(fact); err != nil {
				return err
			}
		}
	}
	return nil
}

// EstimateFactCount returns the number of facts.
func (s SimpleInMemoryStore) EstimateFactCount() int {
	c := 0
	for _, m := range s.shardsByPredicate {
		c += len(m)
	}
	return c
}

// Add implements the FactStore interface by adding the fact to the backing map.
func (s SimpleInMemoryStore) Add(a ast.Atom) bool {
	key := a.Hash()
	if atoms, ok := s.shardsByPredicate[a.Predicate]; ok {
		_, ok := atoms[key]
		if !ok {
			atoms[key] = a
		}
		return !ok
	}
	s.shardsByPredicate[a.Predicate] = map[uint64]ast.Atom{key: a}
	return true
}

// Contains returns true if this store contains this atom already.
func (s SimpleInMemoryStore) Contains(a ast.Atom) bool {
	key := a.Hash()
	if atoms, ok := s.shardsByPredicate[a.Predicate]; ok {
		_, ok := atoms[key]
		return ok
	}
	return false
}

// Merge adds all facts from other to this fact store.
func (s SimpleInMemoryStore) Merge(other ReadOnlyFactStore) {
	for _, pred := range other.ListPredicates() {
		other.GetFacts(ast.NewQuery(pred), func(fact ast.Atom) error {
			s.Add(fact)
			return nil
		})
	}
}

// MergedStore is an implementation of FactStore that merges multiple
// fact stores. It dispatches lookup requests to all of them but sending
// all writes to a single one. It is advisable that the read stores are
// disjoint, otherwise it may well happen that GetFacts will invoke the
// callback with a fact multiple times.
type MergedStore struct {
	readStore  []ReadOnlyFactStore
	writeStore FactStore
}

// Add implementation that adds to the write store.
func (s MergedStore) Add(atom ast.Atom) bool {
	if s.Contains(atom) {
		return false
	}
	return s.writeStore.Add(atom)
}

// Contains implementation that checks all stores.
func (s MergedStore) Contains(atom ast.Atom) bool {
	for _, store := range s.readStore {
		if store.Contains(atom) {
			return true
		}
	}
	return s.writeStore.Contains(atom)
}

// GetFacts implementation that dispatches to all stores.
func (s MergedStore) GetFacts(query ast.Atom, cb func(ast.Atom) error) error {
	for _, store := range s.readStore {
		if err := store.GetFacts(query, cb); err != nil {
			return err
		}
	}
	if err := s.writeStore.GetFacts(query, cb); err != nil {
		return err
	}
	return nil
}

// EstimateFactCount implements a FactStore method. The result
// is an overestimate, because facts may be stored multiple times.
func (s MergedStore) EstimateFactCount() int {
	var estimatedTotal int
	for _, store := range s.readStore {
		estimatedTotal += store.EstimateFactCount()
	}
	return estimatedTotal + s.writeStore.EstimateFactCount()
}

// ListPredicates returns a merged list of predicates.
func (s MergedStore) ListPredicates() []ast.PredicateSym {
	m := make(map[ast.PredicateSym]bool)
	for _, store := range s.readStore {
		for _, p := range store.ListPredicates() {
			m[p] = true
		}
	}
	for _, p := range s.writeStore.ListPredicates() {
		m[p] = true
	}
	res := make([]ast.PredicateSym, 0, len(m))
	for p := range m {
		res = append(res, p)
	}
	return res
}

// Merge forwards to writeStore.Merge
func (s MergedStore) Merge(other ReadOnlyFactStore) {
	s.writeStore.Merge(other)
}

// NewMergedStore returns a new MergedStore.
func NewMergedStore[T ReadOnlyFactStore](readStores []T, writeStore FactStore) FactStore {
	if len(readStores) == 0 {
		return writeStore
	}
	var readOnlyStores []ReadOnlyFactStore
	for _, store := range readStores {
		readOnlyStores = append(readOnlyStores, store)
	}
	return MergedStore{readOnlyStores, writeStore}
}

// Ensure that TeeingStore implements the FactStore interface.
var _ FactStore = MergedStore{nil, NewSimpleInMemoryStore()}

// TeeingStore is an implementation of FactStore that directs all writes to
// an output store, while distributing reads over a read-only base store and
// the output store.
type TeeingStore struct {
	base FactStore
	Out  FactStore
}

// Ensure that TeeingStore implements the FactStore interface.
var _ FactStore = TeeingStore{NewSimpleInMemoryStore(), NewSimpleInMemoryStore()}

// Add implementation that adds to the output store.
func (s TeeingStore) Add(atom ast.Atom) bool {
	if s.base.Contains(atom) {
		return true
	}
	return s.Out.Add(atom)
}

// Contains implementation that checks both stores.
func (s TeeingStore) Contains(atom ast.Atom) bool {
	return s.base.Contains(atom) || s.Out.Contains(atom)
}

// GetFacts implementation that queries both stores.
func (s TeeingStore) GetFacts(query ast.Atom, cb func(ast.Atom) error) error {
	if err := s.base.GetFacts(query, cb); err != nil {
		return err
	}
	if err := s.Out.GetFacts(query, cb); err != nil {
		return err
	}
	return nil
}

// Merge implementation that adds to the output store.
func (s TeeingStore) Merge(other ReadOnlyFactStore) {
	s.Out.Merge(other)
}

// ListPredicates returns a list of predicates.
func (s TeeingStore) ListPredicates() []ast.PredicateSym {
	m := make(map[string]ast.PredicateSym)
	for _, pred := range s.base.ListPredicates() {
		m[pred.Symbol] = pred
	}
	for _, pred := range s.Out.ListPredicates() {
		m[pred.Symbol] = pred
	}
	res := make([]ast.PredicateSym, 0, len(m))
	for _, pred := range m {
		res = append(res, pred)
	}
	return res
}

// EstimateFactCount returns the number of facts. The real number can be lower in case of duplicates.
func (s TeeingStore) EstimateFactCount() int {
	return s.base.EstimateFactCount() + s.Out.EstimateFactCount()
}

// NewTeeingStore returns a new TeeingStore.
func NewTeeingStore(base FactStore) TeeingStore {
	return TeeingStore{base, NewMultiIndexedArrayInMemoryStore()}
}

func Matches(pattern []ast.BaseTerm, args []ast.BaseTerm) bool {
	for i, t := range pattern {
		if _, ok := t.(ast.Constant); ok && !t.Equals(args[i]) {
			return false
		}
	}
	return true
}

// IndexedInMemoryStore provides a simple implementation backed by a three-level map.
// For each predicate sym, we have a separate map, using hash of the first argument and then
// hash of the entire atom.
type IndexedInMemoryStore struct {
	InMemoryStore[map[uint64]map[uint64]ast.Atom]
}

// NewIndexedInMemoryStore constructs a new IndexedInMemoryStore.
func NewIndexedInMemoryStore() IndexedInMemoryStore {
	return IndexedInMemoryStore{
		NewInMemoryStore[map[uint64]map[uint64]ast.Atom](),
	}
}

func (s IndexedInMemoryStore) getFactsOfFirstVariable(a ast.Atom, fn func(ast.Atom) error) error {
	for _, shard := range s.shardsByPredicate[a.Predicate] {
		for _, fact := range shard {
			if Matches(a.Args[1:], fact.Args[1:]) {
				if err := fn(fact); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// GetFacts implementation that looks up facts from an in-memory map.
func (s IndexedInMemoryStore) GetFacts(a ast.Atom, fn func(ast.Atom) error) error {
	if a.Predicate.Arity == 0 {
		if a, ok := s.constants[a.Predicate]; ok {
			return fn(a)
		}
		return nil
	}
	if _, ok := a.Args[0].(ast.Variable); ok {
		return s.getFactsOfFirstVariable(a, fn)
	}
	h := a.Args[0].Hash()
	for _, fact := range s.shardsByPredicate[a.Predicate][h] {
		if Matches(a.Args, fact.Args) {
			if err := fn(fact); err != nil {
				return err
			}
		}
	}
	return nil
}

// Add implements the FactStore interface by adding the fact to the backing map.
func (s IndexedInMemoryStore) Add(a ast.Atom) bool {
	if a.Predicate.Arity == 0 {
		_, ok := s.constants[a.Predicate]
		if !ok {
			s.constants[a.Predicate] = a
			return true
		}
		return false
	}
	h := a.Args[0].Hash()
	shard, ok := s.shardsByPredicate[a.Predicate]
	if !ok {
		shard = map[uint64]map[uint64]ast.Atom{h: {a.Hash(): a}}
		s.shardsByPredicate[a.Predicate] = shard
		return true
	}
	key := a.Hash()
	atoms, ok := shard[h]
	if !ok {
		shard[h] = map[uint64]ast.Atom{a.Hash(): a}
		return true
	}
	if _, ok := atoms[key]; !ok {
		atoms[key] = a
		return true
	}
	return false
}

// Contains returns true if this store contains this atom already.
func (s IndexedInMemoryStore) Contains(a ast.Atom) bool {
	if a.Predicate.Arity == 0 {
		_, ok := s.constants[a.Predicate]
		return ok
	}
	shard, ok := s.shardsByPredicate[a.Predicate]
	if !ok {
		return false
	}
	h := a.Args[0].Hash()
	atoms, ok := shard[h]
	if !ok {
		return false
	}
	_, exists := atoms[a.Hash()]
	return exists
}

// EstimateFactCount returns the number of facts.
func (s IndexedInMemoryStore) EstimateFactCount() int {
	c := len(s.constants)
	for _, s := range s.shardsByPredicate {
		for _, m := range s {
			c += len(m)
		}
	}
	return c
}

// Merge adds all facts from other to this fact store.
func (s IndexedInMemoryStore) Merge(other ReadOnlyFactStore) {
	for _, pred := range other.ListPredicates() {
		other.GetFacts(ast.NewQuery(pred), func(fact ast.Atom) error {
			s.Add(fact)
			return nil
		})
	}
}

// ListPredicates returns a list of predicates.
func (s IndexedInMemoryStore) ListPredicates() []ast.PredicateSym {
	var r []ast.PredicateSym
	for p := range s.constants {
		r = append(r, p)
	}
	for p := range s.shardsByPredicate {
		r = append(r, p)
	}
	return r
}

// MultiIndexedInMemoryStore provides a simple implementation backed by a four-level map.
// For each predicate sym, we have a separate map, using the index and the hash of the nth argument
// and then hash of the entire atom.
type MultiIndexedInMemoryStore struct {
	InMemoryStore[map[uint16]map[uint64]map[uint64]*ast.Atom]
}

// NewMultiIndexedInMemoryStore constructs a new MultiIndexedInMemoryStore.
func NewMultiIndexedInMemoryStore() MultiIndexedInMemoryStore {
	return MultiIndexedInMemoryStore{
		NewInMemoryStore[map[uint16]map[uint64]map[uint64]*ast.Atom](),
	}
}

func (s MultiIndexedInMemoryStore) getFactsOfFirstVariable(a ast.Atom, fn func(ast.Atom) error) error {
	for _, shard := range s.shardsByPredicate[a.Predicate][0] {
		for _, fact := range shard {
			if Matches(a.Args[1:], fact.Args[1:]) {
				if err := fn(*fact); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// GetFacts implementation that looks up facts from an in-memory map.
func (s MultiIndexedInMemoryStore) GetFacts(a ast.Atom, fn func(ast.Atom) error) error {
	if a.Predicate.Arity == 0 {
		if a, ok := s.constants[a.Predicate]; ok {
			return fn(a)
		}
		return nil
	}
	for i := 0; i < a.Predicate.Arity; i++ {
		// Find a non variable parameter.
		if _, ok := a.Args[i].(ast.Variable); !ok {
			h := a.Args[i].Hash()
			for _, fact := range s.shardsByPredicate[a.Predicate][uint16(i)][h] {
				if Matches(a.Args, fact.Args) {
					if err := fn(*fact); err != nil {
						return err
					}
				}
			}
			return nil
		}
	}
	return s.getFactsOfFirstVariable(a, fn)
}

// Add implements the FactStore interface by adding the fact to the backing map.
func (s MultiIndexedInMemoryStore) Add(a ast.Atom) bool {
	if a.Predicate.Arity == 0 {
		_, ok := s.constants[a.Predicate]
		if !ok {
			s.constants[a.Predicate] = a
			return true
		}
		return false
	}
	aHash := a.Hash()
	shard, ok := s.shardsByPredicate[a.Predicate]
	if !ok {
		shard = make(map[uint16]map[uint64]map[uint64]*ast.Atom)
		s.shardsByPredicate[a.Predicate] = shard
		for i := 0; i < a.Predicate.Arity; i++ {
			iHash := a.Args[i].Hash()
			shard[uint16(i)] = map[uint64]map[uint64]*ast.Atom{iHash: {aHash: &a}}
		}
		return true
	}
	added := false
	for i := 0; i < a.Predicate.Arity; i++ {
		iHash := a.Args[i].Hash()
		params, ok := shard[uint16(i)]
		if !ok {
			shard[uint16(i)] = map[uint64]map[uint64]*ast.Atom{iHash: {aHash: &a}}
			added = true
			continue
		}
		atoms, ok := params[iHash]
		if !ok {
			params[iHash] = map[uint64]*ast.Atom{aHash: &a}
			added = true
		} else if _, ok := atoms[aHash]; !ok {
			atoms[aHash] = &a
			added = true
		}
	}
	return added
}

// Contains returns true if this store contains this atom already.
func (s MultiIndexedInMemoryStore) Contains(a ast.Atom) bool {
	if a.Predicate.Arity == 0 {
		_, ok := s.constants[a.Predicate]
		return ok
	}
	shard, ok := s.shardsByPredicate[a.Predicate]
	if !ok {
		return false
	}
	params, ok := shard[0]
	if !ok {
		return false
	}
	h := a.Args[0].Hash()
	atoms, ok := params[h]
	if !ok {
		return false
	}
	_, exists := atoms[a.Hash()]
	return exists
}

// EstimateFactCount returns the number of facts.
func (s MultiIndexedInMemoryStore) EstimateFactCount() int {
	c := len(s.constants)
	for _, s := range s.shardsByPredicate {
		for _, m := range s[0] {
			c += len(m)
		}
	}
	return c
}

// Merge adds all facts from other to this fact store.
func (s MultiIndexedInMemoryStore) Merge(other ReadOnlyFactStore) {
	for _, pred := range other.ListPredicates() {
		other.GetFacts(ast.NewQuery(pred), func(fact ast.Atom) error {
			s.Add(fact)
			return nil
		})
	}
}

// ListPredicates returns a list of predicates.
func (s MultiIndexedInMemoryStore) ListPredicates() []ast.PredicateSym {
	var r []ast.PredicateSym
	for p := range s.constants {
		r = append(r, p)
	}
	for p := range s.shardsByPredicate {
		r = append(r, p)
	}
	return r
}

// MultiIndexedArrayInMemoryStore provides a simple implementation backed by a four-level map.
// For each predicate sym, we have a separate map, using the index and the hash of the nth argument
// and then hash of the entire atom, with the ultimate value being an array of Atoms.
type MultiIndexedArrayInMemoryStore struct {
	InMemoryStore[map[uint16]map[uint64]map[uint64][]*ast.Atom]
}

// NewMultiIndexedArrayInMemoryStore constructs a new MultiIndexedArrayInMemoryStore.
func NewMultiIndexedArrayInMemoryStore() MultiIndexedArrayInMemoryStore {
	return MultiIndexedArrayInMemoryStore{
		NewInMemoryStore[map[uint16]map[uint64]map[uint64][]*ast.Atom](),
	}
}

func (s MultiIndexedArrayInMemoryStore) getFactsOfFirstVariable(a ast.Atom, fn func(ast.Atom) error) error {
	for _, shard := range s.shardsByPredicate[a.Predicate][0] {
		for _, facts := range shard {
			for _, fact := range facts {
				if Matches(a.Args, fact.Args) {
					if err := fn(*fact); err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}

// GetFacts implementation that looks up facts from an in-memory map.
func (s MultiIndexedArrayInMemoryStore) GetFacts(a ast.Atom, fn func(ast.Atom) error) error {
	if a.Predicate.Arity == 0 {
		if a, ok := s.constants[a.Predicate]; ok {
			return fn(a)
		}
		return nil
	}
	for i := 0; i < a.Predicate.Arity; i++ {
		// Find a non variable parameter.
		if _, ok := a.Args[i].(ast.Variable); !ok {
			h := a.Args[i].Hash()
			for _, facts := range s.shardsByPredicate[a.Predicate][uint16(i)][h] {
				for _, fact := range facts {
					if Matches(a.Args, fact.Args) {
						if err := fn(*fact); err != nil {
							return err
						}
					}
				}
			}
			return nil
		}
	}
	return s.getFactsOfFirstVariable(a, fn)
}

// Add implements the FactStore interface by adding the fact to the backing map.
func (s MultiIndexedArrayInMemoryStore) Add(a ast.Atom) bool {
	if a.Predicate.Arity == 0 {
		_, ok := s.constants[a.Predicate]
		if !ok {
			s.constants[a.Predicate] = a
			return true
		}
		return false
	}
	aHash := a.Hash()
	shard, ok := s.shardsByPredicate[a.Predicate]
	if !ok {
		shard = make(map[uint16]map[uint64]map[uint64][]*ast.Atom)
		s.shardsByPredicate[a.Predicate] = shard
		for i := 0; i < a.Predicate.Arity; i++ {
			shard[uint16(i)] = make(map[uint64]map[uint64][]*ast.Atom)
			iHash := a.Args[i].Hash()
			shard[uint16(i)][iHash] = make(map[uint64][]*ast.Atom)
			shard[uint16(i)][iHash][aHash] = append(shard[uint16(i)][iHash][aHash], &a)
		}
		return true
	}
	added := false
	for i := 0; i < a.Predicate.Arity; i++ {
		iHash := a.Args[i].Hash()
		params, ok := shard[uint16(i)]
		if !ok {
			shard[uint16(i)] = make(map[uint64]map[uint64][]*ast.Atom)
			shard[uint16(i)][iHash] = make(map[uint64][]*ast.Atom)
			shard[uint16(i)][iHash][aHash] = append(shard[uint16(i)][iHash][aHash], &a)
			added = true
			continue
		}
		atoms, ok := params[iHash]
		if !ok {
			params[iHash] = make(map[uint64][]*ast.Atom)
			params[iHash][aHash] = append(params[iHash][aHash], &a)
			added = true
		} else if _, ok := atoms[aHash]; !ok {
			atoms[aHash] = append(atoms[aHash], &a)
			added = true
		}
	}
	return added
}

// Contains returns true if this store contains this atom already.
func (s MultiIndexedArrayInMemoryStore) Contains(a ast.Atom) bool {
	if a.Predicate.Arity == 0 {
		_, ok := s.constants[a.Predicate]
		return ok
	}
	shard, ok := s.shardsByPredicate[a.Predicate]
	if !ok {
		return false
	}
	params, ok := shard[0]
	if !ok {
		return false
	}
	h := a.Args[0].Hash()
	atoms, ok := params[h]
	if !ok {
		return false
	}
	for _, fact := range atoms[a.Hash()] {
		if a.Equals(*fact) {
			return true
		}
	}
	return false
}

// EstimateFactCount returns the number of facts.
func (s MultiIndexedArrayInMemoryStore) EstimateFactCount() int {
	c := len(s.constants)
	for _, s := range s.shardsByPredicate {
		for _, m := range s[0] {
			for _, facts := range m {
				c += len(facts)
			}
		}
	}
	return c
}

// Merge adds all facts from other to this fact store.
func (s MultiIndexedArrayInMemoryStore) Merge(other ReadOnlyFactStore) {
	for _, pred := range other.ListPredicates() {
		other.GetFacts(ast.NewQuery(pred), func(fact ast.Atom) error {
			s.Add(fact)
			return nil
		})
	}
}

// ListPredicates returns a list of predicates.
func (s MultiIndexedArrayInMemoryStore) ListPredicates() []ast.PredicateSym {
	var r []ast.PredicateSym
	for p := range s.constants {
		r = append(r, p)
	}
	for p := range s.shardsByPredicate {
		r = append(r, p)
	}
	return r
}
