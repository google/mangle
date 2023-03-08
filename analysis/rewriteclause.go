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
	"github.com/google/mangle/ast"
	"github.com/google/mangle/symbols"
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
	var premises []ast.Term
	for _, p := range clause.Premises {
		switch p := p.(type) {
		case ast.Atom:
			_, _, _, defVars := RectifyAtom(p, boundVars)
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
		}
		premises = append(premises, p)
	}
	return ast.Clause{Head: clause.Head, Premises: premises, Transform: clause.Transform}
}
