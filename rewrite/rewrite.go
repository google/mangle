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

// Package rewrite rewrites rules of a layer (stratum) of a datalog program.
package rewrite

import (
	"fmt"

	"github.com/google/mangle/analysis"
	"github.com/google/mangle/ast"
)

// Rewrite transforms each clause of a given layer (stratum) of a program to another one where
// transforms only appear on clauses with a single atom that defines all variables.
func Rewrite(stratum analysis.Program) analysis.Program {
	gen := freshNameGenerator()

	var newRules []ast.Clause
	for _, clause := range stratum.Rules {
		if (clause.Transform == nil) || clause.Transform.IsLetTransform() || len(clause.Premises) == 1 {
			newRules = append(newRules, clause)
			continue
		}

		// Make a new predicate that captures all variables.
		p := clause.Head.Predicate

		vars := make(map[ast.Variable]bool)
		for _, premise := range clause.Premises {
			getVars(premise, vars)
		}
		// Wildcard variables _ do not correspond to any column in the result relation.
		delete(vars, ast.Variable{"_"})

		internalPred := gen.freshPredicateName(p, len(vars))

		newHead := makeHead(internalPred, vars)
		newRules = append(newRules, ast.Clause{newHead, clause.Premises, nil})
		newRules = append(newRules, ast.Clause{clause.Head, []ast.Term{newHead}, clause.Transform})
	}
	return analysis.Program{stratum.EdbPredicates, stratum.IdbPredicates, newRules}
}

type nameGen struct {
	n int
}

func freshNameGenerator() nameGen {
	return nameGen{0}
}

func (gen nameGen) freshPredicateName(sym ast.PredicateSym, arity int) ast.PredicateSym {
	gen.n++
	internalName := fmt.Sprintf("%s%d%s", sym.Symbol, gen.n, ast.InternalPredicateSuffix)
	return ast.PredicateSym{internalName, arity}
}

func makeHead(sym ast.PredicateSym, vars map[ast.Variable]bool) ast.Atom {
	var args []ast.BaseTerm
	for v := range vars {
		args = append(args, v)
	}
	return ast.Atom{sym, args}
}

// Get the variables. It is enough to consider those terms where variables
// can actually be bound.
func getVars(term ast.Term, vars map[ast.Variable]bool) {
	switch t := term.(type) {
	case ast.Atom:
		ast.AddVars(t, vars)
	case ast.Eq:
		ast.AddVars(t.Left, vars)
		ast.AddVars(t.Right, vars)
	}
}
