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
	"fmt"
	"strings"

	"github.com/google/mangle/ast"
	"github.com/google/mangle/builtin"
	"github.com/google/mangle/factstore"
	"github.com/google/mangle/symbols"
	"github.com/google/mangle/unionfind"
)

// InclusionChecker checks inclusion constraints.
// It does not check type bounds.
type InclusionChecker struct {
	decls       map[ast.PredicateSym]*ast.Decl
	typeChecker *builtin.TypeChecker
}

// NewInclusionChecker returns a new InclusionChecker.
func NewInclusionChecker(decls map[ast.PredicateSym]ast.Decl) (*InclusionChecker, error) {
	desugaredDecls, err := symbols.CheckAndDesugar(decls)
	if err != nil {
		return nil, err
	}
	return NewInclusionCheckerFromDesugared(desugaredDecls), nil
}

// NewInclusionCheckerFromDesugared returns a new InclusionChecker.
// The declarations must be in desugared form.
func NewInclusionCheckerFromDesugared(decls map[ast.PredicateSym]*ast.Decl) *InclusionChecker {
	return &InclusionChecker{decls, builtin.NewTypeCheckerFromDesugared(decls)}
}

// CheckFact verifies that a store containing this fact respects inclusion constraints.
// It also checks types, since after desugaring, the argument types may affect which of the
// inclusion constraint alternatives need to be checked.
func (i InclusionChecker) CheckFact(fact ast.Atom, store factstore.FactStore) error {
	decl, ok := i.decls[fact.Predicate]
	if !ok {
		return fmt.Errorf("could not find declaration for %v", fact.Predicate)
	}
	subst, err := unionfind.UnifyTerms(fact.Args, decl.DeclaredAtom.Args)
	if err != nil {
		return fmt.Errorf("could not unify %v and %v: %w", fact, decl.DeclaredAtom, err)
	}

	var reasons []string
	// Try each disjunction and return when the first one succeeds.
	// Otherwise, we collect the reason why it fails.
disjunctions:
	for j, alternative := range decl.Constraints.Alternatives {
		if err := i.typeChecker.CheckOneBoundDecl(fact, decl.Bounds[j]); err != nil {
			reasons = append(reasons, fmt.Sprintf("%v does not match type bounds %v", fact, decl.Bounds[j]))
			continue
		}

		for _, c := range alternative {
			want := c.ApplySubst(subst)
			extraVars := make(map[ast.Variable]bool)
			ast.AddVars(want, extraVars)
			if len(extraVars) > 0 {
				return fmt.Errorf("%v found extra variables %v", want, extraVars)
			}
			switch a := want.(type) {
			case ast.Atom:
				if !store.Contains(a) {
					reasons = append(reasons, fmt.Sprintf("%v fails, store does not contain %v", alternative, want))
					continue disjunctions
				}
			case ast.NegAtom:
				if store.Contains(a.Atom) {
					reasons = append(reasons, fmt.Sprintf("%v fails, store contains %v but shouldn't", alternative, want))
					continue disjunctions
				}
			case ast.Eq:
				if !a.Left.Equals(a.Right) {
					reasons = append(reasons, fmt.Sprintf("%v fails, equality does not hold %v", alternative, want))
					continue disjunctions
				}
			case ast.Ineq:
				if a.Left.Equals(a.Right) {
					reasons = append(reasons, fmt.Sprintf("%v fails, equality %v holds but shouldn't", alternative, want))
					continue disjunctions
				}
			default:
				return fmt.Errorf("unexpected inclusion constraint %v", want)
			}
		}
		// All constraints from this disjunction element have worked out.
		return nil
	}
	return fmt.Errorf("none of the inclusion constraints are satisfied. reasons: " + strings.Join(reasons, ","))
}
