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

package analysis

import (
	"testing"
)

func TestTrie(t *testing.T) {
	tests := []struct {
		allparts [][]string
		query    []string
		want     int
	}{
		{
			allparts: nil,
			query:    []string{"foo"},
			want:     -1,
		},
		{
			allparts: [][]string{[]string{"foo"}},
			query:    []string{"foo"},
			want:     -1,
		},
		{
			allparts: [][]string{[]string{"foo"}},
			query:    []string{"foo", "bar"},
			want:     0,
		},
		{
			allparts: [][]string{[]string{"foo", "bar"}},
			query:    []string{"foo", "bar"},
			want:     -1,
		},
		{
			allparts: [][]string{[]string{"foo", "bar"}},
			query:    []string{"foo", "bar", "baz"},
			want:     1,
		},
	}

	for _, test := range tests {
		n := newNameTrie()
		for _, parts := range test.allparts {
			n.Add(parts)
			if !n.Contains(parts) {
				t.Errorf("Trie(%v).Contains(%v)=false want true", test.allparts, parts)
			}
		}

		if n.Contains([]string{"absent"}) {
			t.Errorf("Trie(%v).Contains(%v)=true want false", test.allparts, []string{"absent"})
		}
		got := n.LongestPrefix(test.query)
		if got != test.want {
			t.Errorf("Trie(%v).LongestPrefix(%v)=%d want %d", test.allparts, test.query, got, test.want)
		}
	}
}
