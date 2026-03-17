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
	"encoding/json"
	"fmt"
	"io"
)

// FormatText writes lint results in human-readable text format.
func FormatText(w io.Writer, results []LintResult) {
	for _, r := range results {
		loc := r.File
		if loc == "" {
			loc = "<stdin>"
		}
		fmt.Fprintf(w, "%s: [%s] %s: %s\n", loc, r.Severity, r.RuleName, r.Message)
	}
}

// FormatJSON writes lint results in JSON format.
func FormatJSON(w io.Writer, results []LintResult) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	// Ensure we output an empty array rather than null for zero results.
	if results == nil {
		results = []LintResult{}
	}
	return enc.Encode(results)
}
