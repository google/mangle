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

import "fmt"

// OverlyComplexRule flags rules with too many premises.
type OverlyComplexRule struct{}

func (r *OverlyComplexRule) Name() string            { return "overly-complex-rule" }
func (r *OverlyComplexRule) Description() string      { return "Flags rules with too many premises" }
func (r *OverlyComplexRule) DefaultSeverity() Severity { return SeverityInfo }

func (r *OverlyComplexRule) Check(input *LintInput, config LintConfig) []LintResult {
	var results []LintResult
	threshold := config.MaxPremises
	if threshold <= 0 {
		threshold = 8
	}
	for _, clause := range input.ProgramInfo.Rules {
		if len(clause.Premises) > threshold {
			results = append(results, LintResult{
				RuleName:  r.Name(),
				Severity:  r.DefaultSeverity(),
				Message:   fmt.Sprintf("rule for %q has %d premises (threshold: %d); consider breaking into intermediate predicates", clause.Head.Predicate.Symbol, len(clause.Premises), threshold),
				Predicate: clause.Head.Predicate.Symbol,
			})
		}
	}
	return results
}
