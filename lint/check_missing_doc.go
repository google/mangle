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

// MissingDocRule flags predicates that have no documentation descriptors.
type MissingDocRule struct{}

func (r *MissingDocRule) Name() string            { return "missing-doc" }
func (r *MissingDocRule) Description() string      { return "Flags predicates without documentation" }
func (r *MissingDocRule) DefaultSeverity() Severity { return SeverityInfo }

func (r *MissingDocRule) Check(input *LintInput, config LintConfig) []LintResult {
	var results []LintResult
	for pred, decl := range input.ProgramInfo.Decls {
		if !isUserPredicate(pred) {
			continue
		}
		if decl.IsSynthetic() {
			continue
		}
		docs := decl.Doc()
		if len(docs) == 0 || (len(docs) == 1 && docs[0] == "") {
			results = append(results, LintResult{
				RuleName:  r.Name(),
				Severity:  r.DefaultSeverity(),
				Message:   fmt.Sprintf("predicate %q has no documentation", pred.Symbol),
				Predicate: pred.Symbol,
			})
		}
	}
	return results
}
