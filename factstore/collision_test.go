// Copyright 2026 The Mangle Authors
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
	"testing"

	"codeberg.org/TauCeti/mangle-go/ast"
)

// A constant's hash does not include its type, so p(/abc), p("/abc") and
// p(b"/abc") share a hash, and so do a number, a time and a duration with the
// same value. Every store must keep such facts apart: add each one, find each
// one, and not report one that was never added.
func TestFactsThatShareAHashAreKeptApart(t *testing.T) {
	name, err := ast.Name("/abc")
	if err != nil {
		t.Fatal(err)
	}
	groups := []struct {
		name  string
		atoms []ast.Atom
	}{
		{"name/string/bytes", []ast.Atom{ast.NewAtom("p", name), ast.NewAtom("p", ast.String("/abc")), ast.NewAtom("p", ast.Bytes([]byte("/abc")))}},
		{"number/time/duration", []ast.Atom{ast.NewAtom("p", ast.Number(5)), ast.NewAtom("p", ast.Time(5)), ast.NewAtom("p", ast.Duration(5))}},
		{"two arguments", []ast.Atom{ast.NewAtom("q", name, ast.Number(5)), ast.NewAtom("q", ast.String("/abc"), ast.Time(5))}},
	}
	stores := []func() FactStore{
		func() FactStore { return NewSimpleInMemoryStore() },
		func() FactStore { return NewIndexedInMemoryStore() },
		func() FactStore { return NewMultiIndexedInMemoryStore() },
		func() FactStore { return NewMultiIndexedArrayInMemoryStore() },
		func() FactStore {
			return NewMergedStore([]FactStore{NewSimpleInMemoryStore()}, NewSimpleInMemoryStore())
		},
		func() FactStore { return NewConcurrentFactStore(NewSimpleInMemoryStore()) },
	}
	for _, newStore := range stores {
		for _, g := range groups {
			fs, group := newStore(), g.atoms
			t.Run(fmt.Sprintf("%T/%s", fs, g.name), func(t *testing.T) {
				for i, a := range group {
					for _, later := range group[i:] {
						if fs.Contains(later) {
							t.Errorf("Contains(%v)=true before it was added", later)
						}
					}
					if !fs.Add(a) {
						t.Errorf("Add(%v)=false want true", a)
					}
					if fs.Add(a) {
						t.Errorf("2nd Add(%v)=true want false", a)
					}
				}
				var got []string
				fs.GetFacts(ast.NewQuery(group[0].Predicate), func(fact ast.Atom) error {
					got = append(got, fact.String())
					return nil
				})
				if len(got) != len(group) {
					t.Errorf("GetFacts found %v, want all of %v", got, group)
				}
				if got := fs.EstimateFactCount(); got != len(group) {
					t.Errorf("EstimateFactCount()=%d want %d", got, len(group))
				}
				fsr, ok := fs.(FactStoreWithRemove)
				if !ok {
					return
				}
				if !fsr.Remove(group[0]) {
					t.Errorf("Remove(%v)=false want true", group[0])
				}
				if fsr.Contains(group[0]) {
					t.Errorf("Contains(%v)=true after Remove", group[0])
				}
				for _, a := range group[1:] {
					if !fsr.Contains(a) {
						t.Errorf("Remove(%v) also removed %v", group[0], a)
					}
				}
			})
		}
	}
}
