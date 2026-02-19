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
	"strings"

	"codeberg.org/TauCeti/mangle-go/ast"
)

// NameTrieNode is a node in NameTrie.
type NameTrieNode struct {
	next map[string]*NameTrieNode
	end  bool
}

// NameTrie is a trie for looking up name constants.
// Every node represents a unique part of a name.
// Note that the trie for {"/foo", "/foo/bar"} is different from {"/foo/bar"}: the former
// would map a constant "/foo/baz" to the type "/foo", whereas the latter would map it
// to type "/name". "/foo" appears as a node in both, but only the former treats it
// as a terminal node.
type NameTrie = *NameTrieNode

// NewNameTrie constructs a new NameTrie (representing empty prefix).
func NewNameTrie() NameTrie {
	return &NameTrieNode{make(map[string]*NameTrieNode), false}
}

// Collect traverses a type expression and extracts names.
// Base type expressions are ignored.
func (n NameTrie) Collect(typeExpr ast.BaseTerm) {
	walk := func(typeExpr ast.BaseTerm) {
		switch x := typeExpr.(type) {
		case ast.Constant:
			if IsBaseTypeExpression(x) {
				return
			}
			if x.Type == ast.NameType {
				parts := strings.Split(x.Symbol, "/")
				n.Add(parts[1:])
			}
		case ast.ApplyFn:
			for _, arg := range x.Args {
				n.Collect(arg)
			}
		default:
		}
	}
	walk(typeExpr)
}

// Add adds a part sequence to this trie.
func (n NameTrie) Add(parts []string) NameTrie {
	m := n
	for _, p := range parts {
		next := m.next[p]
		if next == nil {
			next = NewNameTrie()
			m.next[p] = next
		}
		m = next
	}
	m.end = true
	return n
}

// Contains returns true if the part sequence is contained in the trie.
func (n NameTrie) Contains(parts []string) bool {
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

// PrefixName returns, for a given name constant, the longest prefix contained in the trie.
func (n NameTrie) PrefixName(symName string) ast.Constant {
	parts := strings.Split(symName, "/")
	if len(parts) == 1 {
		return ast.NameBound
	}
	index := n.LongestPrefix(parts[1:])
	if index == -1 {
		return ast.NameBound
	}
	prefixstrlen := index + 1 // number of "/" separators
	for i := 0; i <= index; i++ {
		prefixstrlen += len(parts[i+1])
	}
	node, err := ast.Name(symName[:prefixstrlen])
	if err != nil {
		return ast.NameBound // This cannot happen
	}
	return node
}

// LongestPrefix returns index of the last element of longest prefix.
func (n NameTrie) LongestPrefix(parts []string) int {
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

func (n NameTrie) String() string {
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
