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

package symbols

import (
	"fmt"
	"strings"

	"codeberg.org/TauCeti/mangle-go/ast"
)

type desugar struct {
	// Declarations as supplied by the user.
	decls map[ast.PredicateSym]ast.Decl
	// Which ones we have visited.
	seen map[ast.PredicateSym]bool
	// Desugared declarations, the result we want to obtain.
	desugared map[ast.PredicateSym]*ast.Decl
	// Any errors we encounter along the way.
	errors []error
}

// CheckAndDesugar rewrites a complete set of decls so that bound declarations contain
// only type bounds and unary predicate references are added to inclusion constraints.
func CheckAndDesugar(decls map[ast.PredicateSym]ast.Decl) (map[ast.PredicateSym]*ast.Decl, error) {
	d := &desugar{decls, make(map[ast.PredicateSym]bool), make(map[ast.PredicateSym]*ast.Decl), nil}
	d.Desugar()
	if len(d.errors) > 0 {
		return nil, d.AllErrors()
	}
	return d.desugared, nil
}

func (d *desugar) AllErrors() error {
	var s []string
	for _, e := range d.errors {
		s = append(s, e.Error())
	}
	return fmt.Errorf("%s", strings.Join(s, "\n"))
}

func (d *desugar) saveError(err error) {
	d.errors = append(d.errors, err)
}

func (d *desugar) Desugar() {
	for sym, decl := range d.decls {
		if decl.IsDesugared() {
			d.seen[sym] = true
			d.desugared[sym] = &decl
		}
	}
	for sym := range d.decls {
		if err := d.desugarOneDecl(sym); err != nil {
			d.saveError(err)
		}
	}
}

type circularDepError struct {
	names []string
}

var _ error = &circularDepError{}

func (c circularDepError) Error() string {
	return fmt.Sprintf("circular dependency: %s", strings.Join(c.names, "->"))
}

func newCircularDependencyError(pred ast.PredicateSym, parent *circularDepError) circularDepError {
	return circularDepError{names: []string{pred.Symbol}}
}

func (c circularDepError) extend(pred ast.PredicateSym) circularDepError {
	return circularDepError{names: append(c.names, pred.Symbol)}
}

// Desugars one decl, including any dependencies.
// When this returns, d.decl[sym] points to a desugared decl.
// The desugared decl will have a type bound for each argument (possibly /any).
func (d *desugar) desugarOneDecl(sym ast.PredicateSym) error {
	if _, ok := d.desugared[sym]; ok {
		return nil
	}
	if _, ok := d.seen[sym]; ok {
		return newCircularDependencyError(sym, nil)
	}
	decl, ok := d.decls[sym]
	if !ok {
		return fmt.Errorf("could not find decl for %s", sym)
	}
	d.seen[sym] = true
	// Bound declarations are the rows of a m x n matrix, where n
	//                     is the arity of the predicate. When the
	// [T_11, ..., T_1n]   cell Tij is a unary predicate, it may
	//   ...        ...    refer to a union of k type expressions.
	// [T_m1, ..., T_mn]   Instead of expanding the i-th row into k new rows,
	//                     we replace the unary predicate with a union type
	// expression fn:union(s_1,...,s_k).
	type BoundInfo struct {
		bounds         []ast.BaseTerm
		inclusionAtoms []ast.Atom
	}

	var boundInfos []*BoundInfo
	if len(decl.Bounds) == 0 && sym.Arity > 0 {
		bounds := make([]ast.BaseTerm, sym.Arity)
		for i := 0; i < sym.Arity; i++ {
			bounds[i] = ast.AnyBound
		}
		boundInfos = []*BoundInfo{
			{bounds, nil},
		}
	} else {
		boundInfos = make([]*BoundInfo, len(decl.Bounds))
		for i, boundDecl := range decl.Bounds {
			boundInfo := &BoundInfo{
				bounds: make([]ast.BaseTerm, sym.Arity),
			}
			boundInfos[i] = boundInfo
			for j, b := range boundDecl.Bounds {
				if err := WellformedBound(b); err == nil { // if NO error
					boundInfo.bounds[j] = b
					continue
				}
				b, ok := b.(ast.Constant)
				if !ok || b.Type != ast.StringType {
					return fmt.Errorf("not a bound expression: %v %T in %v", b, b, decl)
				}
				// A bound like "foo" refers to a unary predicate foo.
				// It must have been desugared before, otherwise we don't
				// know its type bound.
				predicate := ast.PredicateSym{b.Symbol, 1}
				if err := d.desugarOneDecl(predicate); err != nil {
					switch err := err.(type) {
					case circularDepError: // Give up and show dependency cyle.
						return err.extend(sym)
					default:
						d.saveError(fmt.Errorf("while desugaring decl %v bound %v: %v", predicate, b, err))
						boundInfo.bounds[j] = ast.AnyBound
						continue
					}
				}
				// Separate into type bound and inclusion constraint.
				typeExpr, err := typeBoundForPredicate(d.desugared[predicate])
				if err != nil {
					d.saveError(err)
					boundInfo.bounds[j] = ast.AnyBound
					continue
				}
				boundInfo.bounds[j] = typeExpr

				arg := decl.DeclaredAtom.Args[j]
				v, ok := arg.(ast.Variable)
				if !ok {
					return fmt.Errorf("expected variable in declared atom: %v %v %T", decl, arg, arg)
				}
				boundInfo.inclusionAtoms = append(boundInfo.inclusionAtoms, ast.Atom{predicate, []ast.BaseTerm{v}})
			}
		}
	}
	var existingInclusionAtoms []ast.Atom
	if decl.Constraints != nil {
		existingInclusionAtoms = decl.Constraints.Consequences
	}
	newBoundDecls := make([]ast.BoundDecl, len(boundInfos))
	alternatives := make([][]ast.Atom, len(boundInfos))
	for i, boundInfo := range boundInfos {
		newBoundDecls[i] = ast.NewBoundDecl(boundInfo.bounds...)
		alternatives[i] = unique(existingInclusionAtoms, boundInfo.inclusionAtoms)
	}

	d.desugared[decl.DeclaredAtom.Predicate] = &ast.Decl{
		decl.DeclaredAtom,
		append(decl.Descr, ast.NewAtom("desugared")),
		newBoundDecls,
		&ast.InclusionConstraint{nil, alternatives}}

	return nil
}

// typeBoundForPredicate takes a unary predicate and returns a single
// type bound that describes its argument.
func typeBoundForPredicate(d *ast.Decl) (ast.BaseTerm, error) {
	if len(d.Bounds) == 1 {
		return d.Bounds[0].Bounds[0], nil
	}
	typeExprs := make([]ast.BaseTerm, len(d.Bounds))
	for i, b := range d.Bounds {
		typeExprs[i] = b.Bounds[0]
	}
	return UpperBound(nil /*TODO*/, typeExprs), nil

}

func unique(existing []ast.Atom, atoms []ast.Atom) []ast.Atom {
	hashes := make(map[uint64]ast.Atom)
	for _, a := range atoms {
		hashes[a.Hash()] = a
	}
	for _, a := range existing {
		hashes[a.Hash()] = a
	}
	if len(hashes) == len(atoms) {
		return atoms
	}
	// Retain unique atoms.
	var args []ast.Atom
	for _, arg := range hashes {
		args = append(args, arg)
	}
	return args
}
