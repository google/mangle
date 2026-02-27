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
	"strings"
)

// SingletonVariableRule flags variables that appear in only one term within a clause.
type SingletonVariableRule struct{}

func (r *SingletonVariableRule) Name() string            { return "singleton-variable" }
func (r *SingletonVariableRule) Description() string      { return "Flags variables appearing only once in a clause (likely typos)" }
func (r *SingletonVariableRule) DefaultSeverity() Severity { return SeverityWarning }

func (r *SingletonVariableRule) Check(input *LintInput, config LintConfig) []LintResult {
	var results []LintResult
	for _, clause := range input.ProgramInfo.Rules {
		counts := collectVariablesByTerm(clause)
		for v, count := range counts {
			// Skip wildcards and variables prefixed with underscore.
			if v.Symbol == "_" || strings.HasPrefix(v.Symbol, "_") {
				continue
			}
			if count == 1 {
				results = append(results, LintResult{
					RuleName:  r.Name(),
					Severity:  r.DefaultSeverity(),
					Message:   fmt.Sprintf("variable %q appears only once in rule for %q; use _ if intentionally unused", v.Symbol, clause.Head.Predicate.Symbol),
					Predicate: clause.Head.Predicate.Symbol,
				})
			}
		}
	}
	return results
}
