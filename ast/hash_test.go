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

package ast

import "testing"

// Lists, maps and structs that differ in one value must not share a hash,
// whether that value is nested deepest or not. Szudzik's pairing overflowed 64
// bits on nested pairs, so lists of seven or more items, and maps of eight or
// more entries, all had the same hash.
func TestLongCompoundsHashApart(t *testing.T) {
	label, err := Name("/label")
	if err != nil {
		t.Fatal(err)
	}
	for n := 1; n <= 12; n++ {
		for _, pos := range []int{0, n - 1} {
			lists, maps, structs := map[uint64]string{}, map[uint64]string{}, map[uint64]string{}
			for v := int64(0); v < 50; v++ {
				items := make([]Constant, n)
				for i := range items {
					items[i] = Number(int64(i + 1))
				}
				items[pos] = Number(v)
				list := List(items)
				m, s := &MapNil, &StructNil
				for i := range items {
					key := Number(int64(i))
					next := MapCons(&key, &items[i], m)
					m = &next
					nextStruct := StructCons(&label, &items[i], s)
					s = &nextStruct
				}
				for _, c := range []struct {
					seen  map[uint64]string
					value Constant
				}{{lists, list}, {maps, *m}, {structs, *s}} {
					if prev, ok := c.seen[c.value.Hash()]; ok {
						t.Fatalf("%v and %v have the same hash", prev, c.value)
					}
					c.seen[c.value.Hash()] = c.value.String()
				}
			}
		}
	}
}
