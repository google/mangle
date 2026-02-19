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

// Package unionfind is an implementation of Union-Find for use in unification.
package unionfind

import (
	"fmt"
	"strings"

	"codeberg.org/TauCeti/mangle-go/ast"
)

// UnionFind holds a data structure that permits fast unification.
type UnionFind struct {
	parent map[ast.BaseTerm]ast.BaseTerm
}

// New constructs a new UnionFind instance.
func New() UnionFind {
	return UnionFind{make(map[ast.BaseTerm]ast.BaseTerm)}
}

// AsConstSubstList turns this UnionFind structure into a linked list representation.
func (uf UnionFind) AsConstSubstList() ast.ConstSubstList {
	var subst ast.ConstSubstList
	for k := range uf.parent {
		v, ok := k.(ast.Variable)
		if ok {
			if c, ok := uf.find(v).(ast.Constant); ok {
				subst = subst.Extend(v, c)
			}
		}
	}
	return subst
}

// String returns a readable debug string for this UnionFind object.
func (uf UnionFind) String() string {
	var sb strings.Builder
	sb.WriteRune('{')
	for k, v := range uf.parent {
		if k.Equals(v) {
			continue
		}
		sb.WriteRune(' ')
		sb.WriteString(k.String())
		sb.WriteString("->")
		sb.WriteString(v.String())
	}
	sb.WriteString(" }")
	return sb.String()
}

// Adds an edge.
func (uf UnionFind) union(s ast.BaseTerm, t ast.BaseTerm) {
	sroot := uf.find(s)
	troot := uf.find(t)
	if _, ok := sroot.(ast.Constant); ok {
		uf.parent[troot] = sroot
	} else {
		uf.parent[sroot] = troot
	}
}

// Find the representative element from the set of s.
func (uf UnionFind) find(s ast.BaseTerm) ast.BaseTerm {
	child := s
	parent := uf.parent[child]
	if parent == nil {
		return nil
	}
	for child != parent {
		grandparent := uf.parent[parent]
		// Optimize the next lookup
		uf.parent[child] = grandparent
		child = grandparent
		parent = uf.parent[child]
	}
	return parent
}

// Returns true if v can be unified with t, updates unionfind sets.
func (uf UnionFind) unify(v ast.Variable, t ast.BaseTerm) bool {
	vroot := uf.find(v)
	if vroot == nil {
		vroot = v
	}
	troot := uf.find(t)
	if troot == nil {
		troot = t
	}
	if vroot.Equals(troot) {
		return true
	}
	_, vconst := vroot.(ast.Constant)
	_, tconst := troot.(ast.Constant)
	if vconst && tconst {
		return false
	}
	uf.parent[v] = vroot
	uf.parent[t] = troot
	uf.union(vroot, troot)
	return true
}

// Get implements the Subst interface so UnionFind can be used as substituion.
func (uf UnionFind) Get(v ast.Variable) ast.BaseTerm {
	if res := uf.find(v); res != nil {
		return res
	}
	return v
}

// InitVars initializes a unionfind with a given substitution. The caller
// needs to ensure that none of the variables is "_" and none of the variables
// appears among the terms.
func InitVars(vars []ast.Variable, ts []ast.BaseTerm) (UnionFind, error) {
	uf := UnionFind{make(map[ast.BaseTerm]ast.BaseTerm)}
	if len(vars) != len(ts) {
		return UnionFind{}, fmt.Errorf("not of equal size")
	}
	for i, v := range vars {
		t := ts[i]
		uf.parent[t] = t
		uf.parent[v] = t
	}
	return uf, nil
}

// UnifyTerms unifies two same-length lists of relational terms. It does not handle
// apply-expressions.
func UnifyTerms(xs []ast.BaseTerm, ys []ast.BaseTerm) (UnionFind, error) {
	if len(xs) != len(ys) {
		return UnionFind{}, fmt.Errorf("not of equal size")
	}
	uf := UnionFind{make(map[ast.BaseTerm]ast.BaseTerm)}
	var newXs []ast.BaseTerm
	var newYs []ast.BaseTerm
	for i, x := range xs {
		y := ys[i]
		if x.Equals(ast.Variable{"_"}) || y.Equals(ast.Variable{"_"}) {
			continue
		}
		uf.parent[x] = x
		uf.parent[y] = y
		newXs = append(newXs, x)
		newYs = append(newYs, y)
	}
	return uf, unifyTermsUpdate(newXs, newYs, uf)
}

// UnifyTermsExtend unifies two same-length lists of relational terms, returning
// an extended UnionFind. It does not handle apply-expressions.
func UnifyTermsExtend(xs []ast.BaseTerm, ys []ast.BaseTerm, base UnionFind) (UnionFind, error) {
	if len(xs) != len(ys) {
		return UnionFind{}, fmt.Errorf("not of equal size")
	}
	uf := UnionFind{make(map[ast.BaseTerm]ast.BaseTerm)}
	for k, v := range base.parent {
		uf.parent[k] = v
	}
	var newXs []ast.BaseTerm
	var newYs []ast.BaseTerm
	for i, x := range xs {
		y := ys[i]
		if x.Equals(ast.Variable{"_"}) || y.Equals(ast.Variable{"_"}) {
			continue
		}
		if uf.find(x) == nil {
			uf.parent[x] = x
		}
		newXs = append(newXs, x)
	}
	for i, y := range ys {
		x := xs[i]
		if x.Equals(ast.Variable{"_"}) || y.Equals(ast.Variable{"_"}) {
			continue
		}
		if uf.find(y) == nil {
			uf.parent[y] = y
		}
		newYs = append(newYs, y)
	}
	return uf, unifyTermsUpdate(newXs, newYs, uf)
}

func unifyTermsUpdate(xs []ast.BaseTerm, ys []ast.BaseTerm, uf UnionFind) error {
	for i, x := range xs {
		y := ys[i]
		xroot := uf.find(x)
		yroot := uf.find(y)
		if xroot.Equals(yroot) {
			continue
		}
		_, xconst := xroot.(ast.Constant)
		_, yconst := yroot.(ast.Constant)
		if xconst && yconst {
			return fmt.Errorf("cannot unify %v %v", xroot, yroot)
		}
		uf.union(xroot, yroot)
	}
	return nil
}
