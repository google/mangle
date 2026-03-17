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

// DeadCodeRule flags IDB predicates whose derived facts are never consumed
// by any other rule's premises.
type DeadCodeRule struct{}

func (r *DeadCodeRule) Name() string            { return "dead-code" }
func (r *DeadCodeRule) Description() string      { return "Flags derived predicates whose results are never consumed" }
func (r *DeadCodeRule) DefaultSeverity() Severity { return SeverityInfo }

func (r *DeadCodeRule) Check(input *LintInput, config LintConfig) []LintResult {
	// Build set of predicates consumed in premises (excluding self-references).
	consumed := make(map[ast.PredicateSym]bool)
	for _, clause := range input.ProgramInfo.Rules {
		for _, p := range clause.Premises {
			for _, pred := range extractPredicatesFromTerm(p) {
				consumed[pred] = true
			}
		}
	}

	var results []LintResult
	for pred := range input.ProgramInfo.IdbPredicates {
		if !isUserPredicate(pred) {
			continue
		}
		if !consumed[pred] {
			results = append(results, LintResult{
				RuleName:  r.Name(),
				Severity:  r.DefaultSeverity(),
				Message:   fmt.Sprintf("predicate %q derives facts but results are never consumed by another rule", pred.Symbol),
				Predicate: pred.Symbol,
			})
		}
	}
	return results
}
