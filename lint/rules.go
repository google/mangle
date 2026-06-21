// Copyright 2024 Google LLC
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

package lint

import (
	"github.com/google/mangle/ast"
)

// LintRule is the interface every lint check implements.
type LintRule interface {
	// Name returns the unique, hyphen-separated rule name.
	Name() string
	// Description returns a one-line description suitable for --list-rules.
	Description() string
	// DefaultSeverity returns the severity level when no override is configured.
	DefaultSeverity() Severity
	// Check runs the lint check against the input and returns zero or more findings.
	Check(input *LintInput, config LintConfig) []LintResult
}

// AllRules returns all built-in lint rules.
func AllRules() []LintRule {
	return []LintRule{
		&UnusedPredicateRule{},
		&MissingDocRule{},
		&NamingConventionRule{},
		&SingletonVariableRule{},
		&OverlyComplexRule{},
		&DeadCodeRule{},
	}
}

// extractPredicatesFromTerm returns all predicate symbols referenced in a term.
func extractPredicatesFromTerm(term ast.Term) []ast.PredicateSym {
	switch p := term.(type) {
	case ast.Atom:
		return []ast.PredicateSym{p.Predicate}
	case ast.NegAtom:
		return []ast.PredicateSym{p.Atom.Predicate}
	case ast.TemporalLiteral:
		return extractPredicatesFromTerm(p.Literal)
	case ast.TemporalAtom:
		return []ast.PredicateSym{p.Atom.Predicate}
	default:
		return nil
	}
}

// isUserPredicate returns true if the predicate is user-defined.
func isUserPredicate(pred ast.PredicateSym) bool {
	return !pred.IsBuiltin() && !pred.IsInternalPredicate() &&
		pred.Symbol != "Package" && pred.Symbol != "Use"
}

// collectVariablesByTerm counts how many distinct terms reference each variable
// in a clause. This counts at the term level: head counts as 1, each premise
// counts as 1, each transform statement counts as 1.
func collectVariablesByTerm(clause ast.Clause) map[ast.Variable]int {
	counts := make(map[ast.Variable]int)

	// Count variables in head.
	headVars := make(map[ast.Variable]bool)
	ast.AddVars(clause.Head, headVars)
	if clause.HeadTime != nil {
		addIntervalVars(*clause.HeadTime, headVars)
	}
	for v := range headVars {
		counts[v]++
	}

	// Count variables in each premise separately.
	for _, p := range clause.Premises {
		premVars := make(map[ast.Variable]bool)
		ast.AddVars(p, premVars)
		for v := range premVars {
			counts[v]++
		}
	}

	// Count variables in transform.
	if clause.Transform != nil {
		addTransformVars(clause.Transform, counts)
	}

	return counts
}

func addTransformVars(t *ast.Transform, counts map[ast.Variable]int) {
	for _, stmt := range t.Statements {
		// Count the output variable (e.g., Sum in "Sum = fn:sum(Z)").
		if stmt.Var != nil {
			counts[*stmt.Var]++
		}
		// Count the input variables in the function arguments.
		stmtVars := make(map[ast.Variable]bool)
		for _, arg := range stmt.Fn.Args {
			ast.AddVars(arg, stmtVars)
		}
		for v := range stmtVars {
			counts[v]++
		}
	}
	if t.Next != nil {
		addTransformVars(t.Next, counts)
	}
}

func addIntervalVars(interval ast.Interval, m map[ast.Variable]bool) {
	if interval.Start.Type == ast.VariableBound {
		m[interval.Start.Variable] = true
	}
	if interval.End.Type == ast.VariableBound {
		m[interval.End.Variable] = true
	}
}
