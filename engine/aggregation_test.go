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

package engine

import (
	"testing"

	"github.com/google/mangle/ast"
)

func TestAggregateBy(t *testing.T) {
	// Test the new aggregate_by functionality
	tests := []testCase{
		{
			initialFacts: []ast.Atom{
				atom("project(/p1)"),
				atom("project(/p2)"),
			},
			clause:        clause("count_all(N) :- project(P) |> do fn:aggregate_by(), let N = fn:group_size()."),
			expectedFacts: []ast.Atom{atom("count_all(2)")}, // Test new group_size function
		},
		{
			initialFacts: []ast.Atom{
				atom("has_dependency(/p1, /log4j)"),
				atom("has_dependency(/p1, /junit)"),
				atom("has_dependency(/p2, /log4j)"),
			},
			clause:        clause("project_deps(ProjMap) :- has_dependency(P, D) |> do fn:aggregate_by(), let ProjMap = fn:group_map(P, D)."),
			expectedFacts: []ast.Atom{atom("project_deps({/p1:[/junit,/log4j],/p2:[/log4j]})")},
		},
	}

	for _, test := range tests {
		runEval(test, t)
	}
}

func TestGroupMap(t *testing.T) {
	// Test the group_map function specifically
	tests := []testCase{
		{
			initialFacts: []ast.Atom{
				atom("has_dependency(/p1, /log4j)"),
				atom("has_dependency(/p1, /junit)"),
				atom("has_dependency(/p2, /log4j)"),
			},
			clause:        clause("project_deps(ProjMap) :- has_dependency(P, D) |> do fn:aggregate_by(), let ProjMap = fn:group_map(P, D)."),
			expectedFacts: []ast.Atom{atom("project_deps({/p1:[/junit,/log4j],/p2:[/log4j]})")},
		},
	}

	for _, test := range tests {
		runEval(test, t)
	}
}