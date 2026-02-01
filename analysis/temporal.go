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

	"github.com/google/mangle/ast"
)

// TemporalWarning represents a warning about potentially problematic temporal rules.
type TemporalWarning struct {
	// Predicate is the predicate that may cause issues.
	Predicate ast.PredicateSym
	// Message describes the potential problem.
	Message string
	// Severity indicates how serious the warning is.
	Severity WarningSeverity
}

// WarningSeverity indicates the severity of a warning.
type WarningSeverity int

const (
	// SeverityInfo is for informational warnings that may not indicate a problem.
	SeverityInfo WarningSeverity = iota
	// SeverityWarning is for warnings that may cause performance issues or unexpected behavior.
	SeverityWarning
	// SeverityCritical is for warnings that are likely to cause non-termination or errors.
	SeverityCritical
)

func (s WarningSeverity) String() string {
	switch s {
	case SeverityInfo:
		return "info"
	case SeverityWarning:
		return "warning"
	case SeverityCritical:
		return "critical"
	default:
		return "unknown"
	}
}

func (w TemporalWarning) String() string {
	return fmt.Sprintf("[%s] %s: %s", w.Severity, w.Predicate.Symbol, w.Message)
}

// CheckTemporalRecursion analyzes a program for potentially problematic recursive temporal rules.
// Returns a list of warnings about patterns that may cause non-termination or performance issues.
func CheckTemporalRecursion(programInfo *ProgramInfo) []TemporalWarning {
	var warnings []TemporalWarning

	// Build a set of temporal predicates
	temporalPreds := make(map[ast.PredicateSym]bool)
	for pred, decl := range programInfo.Decls {
		if decl.IsTemporal() {
			temporalPreds[pred] = true
		}
	}

	if len(temporalPreds) == 0 {
		return nil // No temporal predicates, nothing to check
	}

	// Build dependency graph for IDB predicates
	depGraph := buildTemporalDepGraph(programInfo, temporalPreds)

	// Find strongly connected components (recursive groups)
	sccs := findSCCs(depGraph)

	// Check each SCC for problematic patterns
	for _, scc := range sccs {
		if len(scc) == 1 {
			// Single predicate - check for self-recursion
			for pred := range scc {
				if depGraph[pred][pred] {
					// Self-recursive temporal predicate
					if temporalPreds[pred] {
						warnings = append(warnings, TemporalWarning{
							Predicate: pred,
							Message:   "self-recursive temporal predicate may cause interval explosion; ensure coalescing or use interval limits",
							Severity:  SeverityWarning,
						})
					}
				}
			}
		} else {
			// Multi-predicate SCC - mutual recursion
			hasTemporalPred := false
			for pred := range scc {
				if temporalPreds[pred] {
					hasTemporalPred = true
					break
				}
			}
			if hasTemporalPred {
				// Get first predicate for the warning message
				var firstPred ast.PredicateSym
				for pred := range scc {
					firstPred = pred
					break
				}
				warnings = append(warnings, TemporalWarning{
					Predicate: firstPred,
					Message:   fmt.Sprintf("mutual recursion through temporal predicates may cause non-termination; %d predicates in cycle", len(scc)),
					Severity:  SeverityCritical,
				})
			}
		}
	}

	// Check for future operators in recursive rules (especially problematic)
	for _, rule := range programInfo.Rules {
		headPred := rule.Head.Predicate
		if !temporalPreds[headPred] {
			continue
		}

		for _, premise := range rule.Premises {
			if tl, ok := premise.(ast.TemporalLiteral); ok {
				if tl.Operator != nil {
					// Check if this is a future operator
					if tl.Operator.Type == ast.DiamondPlus || tl.Operator.Type == ast.BoxPlus {
						// Check if the literal references a predicate in the same SCC
						var litPred ast.PredicateSym
						switch lit := tl.Literal.(type) {
						case ast.Atom:
							litPred = lit.Predicate
						case ast.NegAtom:
							litPred = lit.Atom.Predicate
						}
						if litPred.Symbol != "" && isInSameSCC(headPred, litPred, sccs) {
							warnings = append(warnings, TemporalWarning{
								Predicate: headPred,
								Message:   "future operator (<+ or [+) in recursive temporal rule may cause unbounded fact generation",
								Severity:  SeverityCritical,
							})
						}
					}
				}
			}
		}
	}

	return warnings
}

// buildTemporalDepGraph builds a dependency graph for the program,
// focusing on temporal predicate relationships.
func buildTemporalDepGraph(programInfo *ProgramInfo, temporalPreds map[ast.PredicateSym]bool) map[ast.PredicateSym]map[ast.PredicateSym]bool {
	graph := make(map[ast.PredicateSym]map[ast.PredicateSym]bool)

	// Initialize nodes for all IDB predicates
	for pred := range programInfo.IdbPredicates {
		if graph[pred] == nil {
			graph[pred] = make(map[ast.PredicateSym]bool)
		}
	}

	// Build edges from rule dependencies
	for _, rule := range programInfo.Rules {
		headPred := rule.Head.Predicate
		if graph[headPred] == nil {
			graph[headPred] = make(map[ast.PredicateSym]bool)
		}

		for _, premise := range rule.Premises {
			var bodyPred ast.PredicateSym
			switch p := premise.(type) {
			case ast.Atom:
				bodyPred = p.Predicate
			case ast.NegAtom:
				bodyPred = p.Atom.Predicate
			case ast.TemporalLiteral:
				switch lit := p.Literal.(type) {
				case ast.Atom:
					bodyPred = lit.Predicate
				case ast.NegAtom:
					bodyPred = lit.Atom.Predicate
				}
			default:
				continue
			}

			// Only track dependencies to IDB predicates
			if _, isIDB := programInfo.IdbPredicates[bodyPred]; isIDB {
				graph[headPred][bodyPred] = true
			}
		}
	}

	return graph
}

// findSCCs finds strongly connected components using Kosaraju's algorithm.
func findSCCs(graph map[ast.PredicateSym]map[ast.PredicateSym]bool) []map[ast.PredicateSym]bool {
	// First pass: compute finish order
	visited := make(map[ast.PredicateSym]bool)
	var finishOrder []ast.PredicateSym

	var dfs1 func(node ast.PredicateSym)
	dfs1 = func(node ast.PredicateSym) {
		if visited[node] {
			return
		}
		visited[node] = true
		for neighbor := range graph[node] {
			dfs1(neighbor)
		}
		finishOrder = append(finishOrder, node)
	}

	for node := range graph {
		dfs1(node)
	}

	// Build reverse graph
	reverseGraph := make(map[ast.PredicateSym]map[ast.PredicateSym]bool)
	for node := range graph {
		if reverseGraph[node] == nil {
			reverseGraph[node] = make(map[ast.PredicateSym]bool)
		}
	}
	for src, edges := range graph {
		for dest := range edges {
			if reverseGraph[dest] == nil {
				reverseGraph[dest] = make(map[ast.PredicateSym]bool)
			}
			reverseGraph[dest][src] = true
		}
	}

	// Second pass: find SCCs in reverse finish order
	visited = make(map[ast.PredicateSym]bool)
	var sccs []map[ast.PredicateSym]bool

	var dfs2 func(node ast.PredicateSym, scc map[ast.PredicateSym]bool)
	dfs2 = func(node ast.PredicateSym, scc map[ast.PredicateSym]bool) {
		if visited[node] {
			return
		}
		visited[node] = true
		scc[node] = true
		for neighbor := range reverseGraph[node] {
			dfs2(neighbor, scc)
		}
	}

	for i := len(finishOrder) - 1; i >= 0; i-- {
		node := finishOrder[i]
		if !visited[node] {
			scc := make(map[ast.PredicateSym]bool)
			dfs2(node, scc)
			sccs = append(sccs, scc)
		}
	}

	return sccs
}

// isInSameSCC checks if two predicates are in the same strongly connected component.
func isInSameSCC(pred1, pred2 ast.PredicateSym, sccs []map[ast.PredicateSym]bool) bool {
	for _, scc := range sccs {
		if scc[pred1] && scc[pred2] {
			return true
		}
	}
	return false
}
