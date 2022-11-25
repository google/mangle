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
	"strings"
)

type nametrienode struct {
	next map[string]*nametrienode
	end  bool
}

type nametrie = *nametrienode

func newNameTrie() nametrie {
	return &nametrienode{make(map[string]*nametrienode), false}
}

// Adds part sequence to this trie.
func (n nametrie) Add(parts []string) nametrie {
	m := n
	for _, p := range parts {
		next := m.next[p]
		if next == nil {
			next = newNameTrie()
			m.next[p] = next
		}
		m = next
	}
	m.end = true
	return n
}

// Contains returns true if the part sequence is contained in the trie.
func (n nametrie) Contains(parts []string) bool {
	m := n
	for _, p := range parts {
		next := m.next[p]
		if next == nil {
			return false
		}
		m = next
	}
	return m.end
}

// LongestPrefix returns index of the last element of longest prefix.
func (n nametrie) LongestPrefix(parts []string) int {
	last := -1
	current := n
	i := -1
	for _, p := range parts {
		if current.end {
			last = i
		}
		next := current.next[p]
		if next == nil {
			break
		}
		current = next
		i++
	}
	return last
}

func (n nametrie) String() string {
	var sb strings.Builder
	sb.WriteString("{")
	for k, v := range n.next {
		sb.WriteString(k)
		sb.WriteString("=>")
		sb.WriteString(v.String())
	}
	sb.WriteString("} ")
	if n.end {
		sb.WriteString("(end)")
	}
	return sb.String()
}
