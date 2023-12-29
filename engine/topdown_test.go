// Copyright 2023 Google LLC
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
	"github.com/google/mangle/factstore"
	"github.com/google/mangle/functional"
	"github.com/google/mangle/unionfind"
)

func TestTopDown(t *testing.T) {
	store := factstore.NewSimpleInMemoryStore()
	clauses := []ast.Clause{
		clause(`
missing_required(RequiredList, EnabledList, Witness) :- 
		:list:member(Witness, RequiredList),
		!:list:member(Witness, EnabledList).`),
	}

	missingRequiredDecl, err := ast.NewDecl(atom("missing_required(EnabledList, RequiredList, Witness)"), []ast.Atom{
		atom("mode('+', '+', '-')"),
		atom("deferred()"),
	}, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	decls := []ast.Decl{missingRequiredDecl}
	if err := analyzeAndEvalProgramWithDecls(t, clauses, decls, store); err != nil {
		t.Errorf("Program evaluation failed %v program %v", err, clauses)
		return
	}
	query, err := functional.EvalAtom(atom("missing_required(['foo'], ['bar'], Witness)"), nil)
	if err != nil {
		t.Fatal(err)
	}
	want, err := functional.EvalAtom(atom("missing_required(['foo'], ['bar'],'foo')"), nil)
	if err != nil {
		t.Fatal(err)
	}
	context := QueryContext{PredToRules: map[ast.PredicateSym][]ast.Clause{
		ast.PredicateSym{"missing_required", 3}: clauses,
	},
		PredToDecl: map[ast.PredicateSym]*ast.Decl{
			ast.PredicateSym{"missing_required", 3}: &missingRequiredDecl,
		},
		Store: store,
	}

	uf := unionfind.New()
	if err != nil {
		t.Fatal(err)
	}
	found := false
	err = context.EvalQuery(query, []ast.ArgMode{ast.ArgModeInput, ast.ArgModeInput, ast.ArgModeOutput}, uf, func(got ast.Atom) error {
		found = true
		if !got.Equals(want) {
			t.Errorf("got: %v want: %v", got, want)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if !found {
		t.Errorf("got nothing want: %v", want)
	}
}
