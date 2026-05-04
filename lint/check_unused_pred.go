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

	"github.com/google/mangle/ast"
)

// UnusedPredicateRule flags predicates that are declared but never referenced.
type UnusedPredicateRule struct{}

func (r *UnusedPredicateRule) Name() string            { return "unused-predicate" }
func (r *UnusedPredicateRule) Description() string      { return "Flags declared predicates that are never referenced" }
func (r *UnusedPredicateRule) DefaultSeverity() Severity { return SeverityWarning }

func (r *UnusedPredicateRule) Check(input *LintInput, config LintConfig) []LintResult {
	// Build set of all referenced predicates.
	referenced := make(map[ast.PredicateSym]bool)

	// From rules: head and premises.
	for _, clause := range input.ProgramInfo.Rules {
		referenced[clause.Head.Predicate] = true
		for _, p := range clause.Premises {
			for _, pred := range extractPredicatesFromTerm(p) {
				referenced[pred] = true
			}
		}
	}

	// From initial facts.
	for _, fact := range input.ProgramInfo.InitialFacts {
		referenced[fact.Predicate] = true
	}

	var results []LintResult
	for pred, decl := range input.ProgramInfo.Decls {
		if !isUserPredicate(pred) {
			continue
		}
		if decl.IsSynthetic() {
			continue
		}
		if !referenced[pred] {
			results = append(results, LintResult{
				RuleName:  r.Name(),
				Severity:  r.DefaultSeverity(),
				Message:   fmt.Sprintf("predicate %q is declared but never referenced", pred.Symbol),
				Predicate: pred.Symbol,
			})
		}
	}
	return results
}
