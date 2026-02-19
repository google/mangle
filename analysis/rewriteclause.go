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

package analysis

import (
	"codeberg.org/TauCeti/mangle-go/ast"
	"codeberg.org/TauCeti/mangle-go/symbols"
)

func scrutineeIsInputArg(d *ast.Decl) bool {
	modes := d.Modes()
	if len(modes) == 0 {
		return false
	}
	for _, m := range modes {
		if m[0] != ast.ArgModeInput {
			return false
		}
	}
	return true
}

// RewriteClause rewrites a clause using information from declarations.
func RewriteClause(decls map[ast.PredicateSym]*ast.Decl, clause ast.Clause) ast.Clause {
	if len(clause.Premises) == 0 {
		return clause
	}
	boundVars := VarList{}
	pred := clause.Head.Predicate
	if decl, ok := decls[pred]; ok {
		mode := unifyModes(decl.Modes())
		boundVars = boundVars.Extend(
			variablesForArgMode(clause.Head, mode, ast.ArgModeInput|ast.ArgModeInputOutput))
	}
	var premises []ast.Term
	var delayNegAtom []ast.Term
	var delayVars []map[ast.Variable]bool
	for _, p := range clause.Premises {
		needsDelay := false
		switch p := p.(type) {
		case ast.Atom:
			defVarMap := make(map[ast.Variable]bool)
			ast.AddVars(p, defVarMap)
			defVars := make([]ast.Variable, 0, len(defVarMap))
			for v := range defVarMap {
				defVars = append(defVars, v)
			}
			if decl, ok := decls[p.Predicate]; ok {
				if prefix, ok := decl.Reflects(); ok {
					// A predicate that reflects a name prefix type can be rewritten when the
					// argument is:
					// - a variable that is guaranteed to have ArgModeInput, or
					// - a variable defined previously
					if v, ok := p.Args[0].(ast.Variable); ok && scrutineeIsInputArg(decl) || boundVars.Find(v) != -1 {
						premises = append(premises, ast.Atom{symbols.MatchPrefix, []ast.BaseTerm{p.Args[0], prefix}})
						continue
					}
				}
			}
			boundVars = boundVars.Extend(defVars)
		case ast.Eq:
			m := boundVars.AsMap()
			ast.AddVars(p, m)
			boundVars = NewVarList(m)

		case ast.NegAtom:
			varToBind := map[ast.Variable]bool{}
			negVars := make(map[ast.Variable]bool)
			ast.AddVars(p, negVars)
			for v := range negVars {
				if boundVars.Find(v) == -1 {
					varToBind[v] = true
				}
			}
			if len(varToBind) > 0 {
				needsDelay = true
				delayNegAtom = append(delayNegAtom, p)
				delayVars = append(delayVars, varToBind)
			}
		}
		if !needsDelay {
			var toRemove []int
			premises = append(premises, p)
		delayTerms:
			for i, vars := range delayVars {
				for v := range vars {
					if boundVars.Find(v) == -1 {
						continue delayTerms
					}
				}
				premises = append(premises, delayNegAtom[i])
				toRemove = append([]int{i}, toRemove...)
			}
			for i := range toRemove {
				negAtomTail := []ast.Term{}
				varsTail := []map[ast.Variable]bool{}
				if i+1 < len(delayNegAtom) {
					negAtomTail = delayNegAtom[i+1:]
					varsTail = delayVars[i+1:]
				}
				delayNegAtom = append(delayNegAtom[:i], negAtomTail...)
				delayVars = append(delayVars[:i], varsTail...)
			}
		}
	}
	return ast.Clause{Head: clause.Head, HeadTime: clause.HeadTime, Premises: premises, Transform: clause.Transform}
}
