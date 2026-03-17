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
	"fmt"
	"regexp"

	"github.com/google/mangle/ast"
)

var (
	// predicateNameRe matches valid snake_case predicate names, optionally with package prefixes.
	predicateNameRe = regexp.MustCompile(`^([a-z][a-z0-9_]*\.)*[a-z][a-z0-9_]*$`)
	// variableNameRe matches valid variable names: uppercase start, alphanumeric + underscore.
	variableNameRe = regexp.MustCompile(`^[A-Z][A-Za-z0-9_]*$`)
)

// NamingConventionRule checks that predicate names are snake_case and variables
// are uppercase.
type NamingConventionRule struct{}

func (r *NamingConventionRule) Name() string            { return "naming-convention" }
func (r *NamingConventionRule) Description() string      { return "Checks predicate and variable naming conventions" }
func (r *NamingConventionRule) DefaultSeverity() Severity { return SeverityWarning }

func (r *NamingConventionRule) Check(input *LintInput, config LintConfig) []LintResult {
	var results []LintResult

	// Check predicate names.
	checked := make(map[ast.PredicateSym]bool)
	allClauses := append(input.ProgramInfo.Rules, factsAsClauses(input)...)
	for _, clause := range allClauses {
		pred := clause.Head.Predicate
		if checked[pred] || !isUserPredicate(pred) {
			continue
		}
		checked[pred] = true
		if !predicateNameRe.MatchString(pred.Symbol) {
			results = append(results, LintResult{
				RuleName:  r.Name(),
				Severity:  r.DefaultSeverity(),
				Message:   fmt.Sprintf("predicate %q does not follow snake_case naming convention", pred.Symbol),
				Predicate: pred.Symbol,
			})
		}
	}

	// Check variable names in rules.
	for _, clause := range input.ProgramInfo.Rules {
		vars := make(map[ast.Variable]bool)
		ast.AddVarsFromClause(clause, vars)
		for v := range vars {
			name := v.Symbol
			if name == "_" {
				continue
			}
			if !variableNameRe.MatchString(name) {
				results = append(results, LintResult{
					RuleName:  r.Name(),
					Severity:  r.DefaultSeverity(),
					Message:   fmt.Sprintf("variable %q in rule for %q does not follow naming convention (should start with uppercase)", name, clause.Head.Predicate.Symbol),
					Predicate: clause.Head.Predicate.Symbol,
				})
			}
		}
	}

	return results
}

// factsAsClauses wraps initial facts from ProgramInfo into Clause values.
func factsAsClauses(input *LintInput) []ast.Clause {
	var clauses []ast.Clause
	for _, fact := range input.ProgramInfo.InitialFacts {
		clauses = append(clauses, ast.Clause{Head: fact})
	}
	return clauses
}
