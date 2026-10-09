// Copyright 2026 The Mangle Authors
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
	"strings"
	"testing"

	"codeberg.org/TauCeti/mangle-go/analysis"
	"codeberg.org/TauCeti/mangle-go/ast"
	"codeberg.org/TauCeti/mangle-go/factstore"
	"codeberg.org/TauCeti/mangle-go/parse"
	"codeberg.org/TauCeti/mangle-go/unionfind"
)

// Premise order does not matter for inequalities: a `!=` whose variables are
// not bound yet is delayed until they are, like a negation.
func TestInequalityOrderIsIrrelevant(t *testing.T) {
	for _, tt := range []struct {
		src  string
		want []ast.Atom
	}{
		{
			src:  "e(1). e(2). e(3). p(X) :- X != 1, e(X).",
			want: []ast.Atom{ast.NewAtom("p", ast.Number(2)), ast.NewAtom("p", ast.Number(3))},
		},
		{
			src:  "e(1). e(2). e(3). p(X) :- e(X), X != 1.",
			want: []ast.Atom{ast.NewAtom("p", ast.Number(2)), ast.NewAtom("p", ast.Number(3))},
		},
		{
			src:  "e(1). e(2). f(1). p(X) :- e(X), X != Y, f(Y).",
			want: []ast.Atom{ast.NewAtom("p", ast.Number(2))},
		},
		{
			src:  "e(1). e(2). f(1). p(X) :- X != Y, e(X), f(Y).",
			want: []ast.Atom{ast.NewAtom("p", ast.Number(2))},
		},
	} {
		unit, err := parse.Unit(strings.NewReader(tt.src))
		if err != nil {
			t.Fatal(err)
		}
		info, err := analysis.AnalyzeOneUnit(unit, nil)
		if err != nil {
			t.Fatalf("AnalyzeOneUnit(%q): %v", tt.src, err)
		}
		store := factstore.NewSimpleInMemoryStore()
		if err := EvalProgram(info, store); err != nil {
			t.Fatalf("EvalProgram(%q): %v", tt.src, err)
		}
		for _, want := range tt.want {
			if !store.Contains(want) {
				t.Errorf("EvalProgram(%q): missing %v", tt.src, want)
			}
		}
	}
}

// An inequality whose variable no premise binds is an error, not a premise
// that silently fails. The delayed premise is kept at the end of the body, so
// rule checking reports the unbound variable.
func TestInequalityWithNeverBoundVariableIsRejected(t *testing.T) {
	for _, src := range []string{
		"e(1). p(X) :- e(X), X != Y.",
		"e(1). p(X) :- X != Y, e(X).",
	} {
		unit, err := parse.Unit(strings.NewReader(src))
		if err != nil {
			t.Fatal(err)
		}
		if _, err := analysis.AnalyzeOneUnit(unit, nil); err == nil {
			t.Errorf("AnalyzeOneUnit(%q) succeeded, want an error about the unbound variable", src)
		}
	}
}

// premiseIneq itself refuses an unbound variable. After delay rewriting this
// should be unreachable from programs that pass analysis; the refusal is a
// safety net that keeps a would-be silent wrong answer loud.
func TestPremiseIneqRefusesUnboundVariable(t *testing.T) {
	if _, err := premiseIneq(ast.Variable{Symbol: "X"}, ast.Number(1), unionfind.New()); err == nil {
		t.Error("premiseIneq(X, 1) with X unbound succeeded, want an error")
	}
}
