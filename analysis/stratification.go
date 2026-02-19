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
	"fmt"

	"codeberg.org/TauCeti/mangle-go/ast"
	"codeberg.org/TauCeti/mangle-go/builtin"
)

// edgeMap represents the dependencies, i.e. those IDB predicate symbols q that
// appear in the body of a rule p :- ... q ..., possibly negated.
// The boolean indicates whether the target appears negated. If there is
// both a positive and negated dependency, we only keep the negative one.
type edgeMap map[ast.PredicateSym]bool

// depGraph maps each predicate symbol p to its edge map.
type depGraph map[ast.PredicateSym]edgeMap

// Program represents a set of rules that we may or may not be able to stratify.
type Program struct {
	// Extensional predicate symbols; those that do not appear in the head of a clause with a body.
	EdbPredicates map[ast.PredicateSym]struct{}
	// Intensional predicate symbols; those that do appear in the head of a clause with a body.
	IdbPredicates map[ast.PredicateSym]struct{}
	// Rules that have a body.
	Rules []ast.Clause
}

func makeDepGraph(program Program) depGraph {
	dep := make(depGraph)
	for _, rule := range program.Rules {
		s := rule.Head.Predicate
		dep.initNode(s)
		for _, premise := range rule.Premises {
			switch p := premise.(type) {
			case ast.Atom:
				if _, ok := builtin.Predicates[p.Predicate]; ok {
					continue
				}
				if _, ok := program.EdbPredicates[p.Predicate]; !ok {
					if rule.Transform == nil || rule.Transform.IsLetTransform() {
						dep.addEdge(s, p.Predicate, false)
					} else {
						// Recursion through a do-transform is not permitted.
						// We treat this as if it was a negation.
						dep.addEdge(s, p.Predicate, true)
					}
				}
			case ast.NegAtom:
				if _, ok := program.EdbPredicates[p.Atom.Predicate]; !ok {
					dep.addEdge(s, p.Atom.Predicate, true)
				}
			}
		}
	}
	return dep
}

// Stratify checks whether a program can be stratified. It returns strongly-connected components
// and a map from predicate to stratum in the affirmative case, an error otherwise.
// The result list of strata is topologically sorted.
// Stratification means separating a list of intensional predicate symbols with rules
// into strata (layers) such that each layer only depends on the
// evaluation of negated atoms for idb predicates in lower layers.
func Stratify(program Program) ([]Nodeset, map[ast.PredicateSym]int, error) {
	dep := makeDepGraph(program)
	strata := dep.sccs()
	predToStratum := make(map[ast.PredicateSym]int)
	for i, c := range strata {
		for sym := range c {
			predToStratum[sym] = i
		}
		for sym := range c {
			for dest, negated := range dep[sym] {
				if !negated {
					continue
				}
				// "Negative" dependency in same stratum indicates recursion through negation.
				if destStratum, ok := predToStratum[dest]; ok && destStratum == i {
					return nil, nil, fmt.Errorf("program cannot be stratified")
				}
			}
		}
	}
	strata, predToStratum = dep.sortResult(strata, predToStratum)
	return strata, predToStratum, nil
}

func (dep depGraph) initNode(src ast.PredicateSym) {
	if _, ok := dep[src]; !ok {
		dep[src] = make(edgeMap)
	}
}

func (dep depGraph) addEdge(src ast.PredicateSym, dest ast.PredicateSym, negated bool) {
	edges := dep[src]
	if negated {
		edges[dest] = negated
		return
	}
	if wasNegated, ok := edges[dest]; !ok || !wasNegated {
		edges[dest] = false
	}
}

func (dep depGraph) transpose() depGraph {
	rev := make(depGraph)
	for src, edges := range dep {
		for dest, negated := range edges {
			rev.initNode(dest)
			rev.addEdge(dest, src, negated)
		}
	}
	return rev
}

type nodelist []ast.PredicateSym

// Nodeset represents a set of nodes in the dependency graph.
type Nodeset map[ast.PredicateSym]struct{}

func (dep depGraph) sccs() []Nodeset {
	// Kosaraju's algorithm
	// Forward pass.
	S := make(nodelist, 0, len(dep)) // postorder stack
	seen := make(Nodeset)
	var visit func(node ast.PredicateSym)
	visit = func(node ast.PredicateSym) {
		if _, ok := seen[node]; !ok {
			seen[node] = struct{}{}
			for e := range dep[node] {
				visit(e)
			}
			S = append(S, node)
		}
	}
	for node := range dep {
		visit(node)
	}

	// Reverse pass.
	rev := dep.transpose()
	var scc Nodeset
	seen = make(Nodeset)
	var rvisit func(node ast.PredicateSym)
	rvisit = func(node ast.PredicateSym) {
		if _, ok := seen[node]; !ok {
			seen[node] = struct{}{}
			scc[node] = struct{}{}
			for e := range rev[node] {
				rvisit(e)
			}
		}
	}
	var sccs []Nodeset
	for len(S) > 0 {
		top := S[len(S)-1]
		S = S[:len(S)-1] // pop
		if _, ok := seen[top]; !ok {
			scc = make(Nodeset)
			rvisit(top)
			sccs = append(sccs, scc)
		}
	}
	return sccs
}

// sortResult sorts the strata topologically (ignoring cycles).
func (dep depGraph) sortResult(strata []Nodeset, predToStratumMap map[ast.PredicateSym]int) ([]Nodeset, map[ast.PredicateSym]int) {
	var sorted []int
	seen := make(map[int]struct{})
	var visitStratum func(index int)
	visitStratum = func(index int) {
		if _, ok := seen[index]; ok {
			return
		}
		seen[index] = struct{}{}
		for sym := range strata[index] {
			for d := range dep[sym] {
				visitStratum(predToStratumMap[d])
			}
		}
		sorted = append(sorted, index)
	}

	for i := 0; i < len(strata); i++ {
		visitStratum(i)
	}
	newstrata := make([]Nodeset, len(strata), len(strata))
	oldToNew := make(map[int]int)
	for i := 0; i < len(strata); i++ {
		newstrata[i] = strata[sorted[i]]
		oldToNew[sorted[i]] = i
	}
	newPredToStratumMap := make(map[ast.PredicateSym]int, len(predToStratumMap))
	for sym := range predToStratumMap {
		newPredToStratumMap[sym] = oldToNew[predToStratumMap[sym]]
	}
	return newstrata, newPredToStratumMap
}
