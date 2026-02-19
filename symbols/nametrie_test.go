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

package symbols

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"codeberg.org/TauCeti/mangle-go/ast"
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
			allparts: [][]string{{"foo"}},
			query:    []string{"foo"},
			want:     -1,
		},
		{
			allparts: [][]string{{"foo"}},
			query:    []string{"foo", "bar"},
			want:     0,
		},
		{
			allparts: [][]string{{"foo", "bar"}},
			query:    []string{"foo", "bar"},
			want:     -1,
		},
		{
			allparts: [][]string{{"foo", "bar"}},
			query:    []string{"foo", "bar", "baz"},
			want:     1,
		},
	}

	for _, test := range tests {
		n := NewNameTrie()
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

func TestPrefixTrie(t *testing.T) {
	tests := []struct {
		nameTrie NameTrie
		query    string
		want     ast.Constant
	}{
		{
			nameTrie: NewNameTrie().Add([]string{"foo"}).Add([]string{"foo", "bar"}),
			query:    "/foo",
			want:     ast.NameBound,
		},
		{
			nameTrie: NewNameTrie().Add([]string{"foo"}).Add([]string{"foo", "bar"}),
			query:    "/foo/bar",
			want:     name("/foo"),
		},
		{
			nameTrie: NewNameTrie().Add([]string{"foo"}).Add([]string{"foo", "bar"}),
			query:    "/foo/baz",
			want:     name("/foo"),
		},
		{
			nameTrie: NewNameTrie().Add([]string{"foo"}).Add([]string{"foo", "bar"}),
			query:    "",
			want:     ast.NameBound,
		},
		{
			nameTrie: NewNameTrie().Add([]string{"foo"}).Add([]string{"foo", "bar"}),
			query:    "/foo/bar/baz",
			want:     name("/foo/bar"),
		},
	}
	for _, test := range tests {
		got := test.nameTrie.PrefixName(test.query)
		if !cmp.Equal(got, test.want, cmp.AllowUnexported(ast.Constant{})) {
			t.Errorf("prefixType(%v, %v) = %v, want %v", test.nameTrie, test.query, got, test.want)
		}
	}
}

func TestCollect(t *testing.T) {
	nameTrie := NewNameTrie()
	someType, err := ast.Name("/something/something")
	if err != nil {
		t.Fatalf("ast.Name(%v) failed: %v", "/something/something", err)
	}
	nameTrie.Collect(ast.ApplyFn{Function: List, Args: []ast.BaseTerm{
		ast.ApplyFn{Function: Pair, Args: []ast.BaseTerm{ast.StringBound, someType}},
	}})
	got := nameTrie.PrefixName("/something/something/foo")
	// The type of `/something/something/foo` is `/something/something`.
	if !someType.Equals(got) {
		t.Errorf("TestCollect prefix: got %v from trie, want %v", got, someType)
	}
}
