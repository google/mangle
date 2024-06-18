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

// Package packages provides functionality for creating datalog packages.
package packages

import (
	"fmt"
	"strings"

	"bitbucket.org/creachadair/stringset"
	"github.com/google/mangle/ast"
	"github.com/google/mangle/parse"
	"github.com/google/mangle/symbols"
)

var (
	nameSym = ast.PredicateSym{"name", 1}
)

// Package represents a Melange package.
type Package struct {
	Name  string
	Atoms []ast.Atom
	units []parse.SourceUnit
}

func (p *Package) declarationMappings() (stringset.Set, map[ast.PredicateSym]bool, error) {
	usedPackages := stringset.New(p.Name)
	definedIdentifier := map[ast.PredicateSym]bool{}
	for _, u := range p.units {
		for _, clause := range u.Clauses {
			oldSym := clause.Head.Predicate
			definedIdentifier[oldSym] = true
		}
		for _, decl := range u.Decls {
			if decl.DeclaredAtom.Predicate == symbols.Package {
				continue
			}
			if decl.DeclaredAtom.Predicate == symbols.Use {
				for _, desc := range decl.Descr {
					if desc.Predicate == nameSym {
						if len(desc.Args) != 1 {
							return nil, nil, fmt.Errorf("unexpected length %v", len(desc.Args))
						}
						n, ok := desc.Args[0].(ast.Constant)
						if !ok {
							return nil, nil, fmt.Errorf("unexpected type for name, expected constant")
						}
						v, err := n.StringValue()
						if err != nil {
							return nil, nil, err
						}
						if v == p.Name {
							return nil, nil, fmt.Errorf("used package %q is same as current package", v)
						}
						usedPackages.Add(v)
						break
					}
				}
			}
			oldSym := decl.DeclaredAtom.Predicate
			definedIdentifier[oldSym] = true
		}
	}
	return usedPackages, definedIdentifier, nil
}

// Decls returns Decls of the package with rewritten identifiers.
func (p *Package) Decls() ([]ast.Decl, error) {
	usedPackages, definedIdentifier, err := p.declarationMappings()
	if err != nil {
		return nil, err
	}

	decls := []ast.Decl{}
	for _, u := range p.units {
		for _, decl := range u.Decls {
			// Skip Package and Import Decls.
			if decl.DeclaredAtom.Predicate == symbols.Package {
				continue
			}
			if decl.DeclaredAtom.Predicate == symbols.Use {
				continue
			}

			if p.Name != "" {
				decl.DeclaredAtom.Predicate.Symbol = fmt.Sprintf("%s.%s", p.Name, decl.DeclaredAtom.Predicate.Symbol)
			}
			for i, bd := range decl.Bounds {
				for j, b := range bd.Bounds {
					if err := symbols.WellformedBound(b); err == nil { // if no error
						continue
					}

					if c, ok := b.(ast.Constant); ok {
						if c.Type != ast.StringType {
							return nil, fmt.Errorf("cannot handle bound %v part of %v for decl %v", c, bd, decl.DeclaredAtom.Predicate.Symbol)
						}
						sv, err := c.StringValue()
						if err != nil {
							return nil, err
						}
						if _, ok := definedIdentifier[ast.PredicateSym{Symbol: sv, Arity: 1}]; ok {
							if p.Name != "" {
								decl.Bounds[i].Bounds[j] = ast.String(fmt.Sprintf("%s.%s", p.Name, sv))
							}
							continue
						}
						u := strings.LastIndex(sv, ".")
						if u == -1 {
							continue
						}
						if !usedPackages.Contains(sv[:u]) {
							return nil, fmt.Errorf("in package %q, 'Use' declaration for %v not found", p.Name, sv)
						}
					}
				}
			}
			decls = append(decls, decl)
		}
	}
	return decls, nil
}

func (p *Package) updatedAtom(a ast.Atom, definedIdentifier map[ast.PredicateSym]bool, usedPackages stringset.Set) (ast.Atom, error) {
	if _, ok := definedIdentifier[a.Predicate]; ok {
		if p.Name == "" {
			return a, nil
		}
		a.Predicate.Symbol = fmt.Sprintf("%s.%s", p.Name, a.Predicate.Symbol)
		return a, nil
	}
	u := strings.LastIndex(a.Predicate.Symbol, ".")
	// TODO: We handle this case in the transition stage. We can remove this later and return an error instead.
	if u == -1 {
		return a, nil
	}

	pkgName := a.Predicate.Symbol[:u]

	if !usedPackages.Contains(pkgName) {
		return ast.Atom{}, fmt.Errorf("in package %q, 'Use' declaration for %v not found", p.Name, a.Predicate)
	}
	return a, nil
}

// Clauses returns Clauses of the package with rewritten identifiers.
func (p *Package) Clauses() ([]ast.Clause, error) {
	clauses := []ast.Clause{}

	usedPackages, definedIdentifier, err := p.declarationMappings()
	if err != nil {
		return nil, err
	}

	for _, u := range p.units {
		for _, clause := range u.Clauses {
			if p.Name != "" {
				clause.Head.Predicate.Symbol = fmt.Sprintf("%s.%s", p.Name, clause.Head.Predicate.Symbol)
			}
			for i, t := range clause.Premises {
				switch a := t.(type) {
				case ast.Atom:
					na, err := p.updatedAtom(a, definedIdentifier, usedPackages)
					if err != nil {
						return nil, err
					}
					clause.Premises[i] = na
				case ast.NegAtom:
					ia, err := p.updatedAtom(a.Atom, definedIdentifier, usedPackages)
					if err != nil {
						return nil, err
					}
					clause.Premises[i] = ast.NegAtom{Atom: ia}
				default:
					continue
				}
			}
			clauses = append(clauses, clause)
		}
	}
	return clauses, nil
}

func findPackage(decls []ast.Decl) (Package, error) {
	name := ""
	atoms := []ast.Atom{}
	for _, decl := range decls {
		if decl.DeclaredAtom.Predicate != symbols.Package {
			continue
		}
		for _, desc := range decl.Descr {
			if desc.Predicate == nameSym {
				if len(desc.Args) != 1 {
					return Package{}, fmt.Errorf("unexpected length %v", len(desc.Args))
				}
				n, ok := desc.Args[0].(ast.Constant)
				if !ok {
					return Package{}, fmt.Errorf("invalid description %v for name", desc)
				}

				var err error
				name, err = n.StringValue()
				if err != nil {
					return Package{}, err
				}
			} else {
				atoms = append(atoms, desc)
			}
		}
		break
	}
	return Package{Name: name, Atoms: atoms}, nil
}

// Merge merges two packages.
func (p *Package) Merge(other Package) error {
	if other.Name != p.Name {
		return fmt.Errorf("other package name %q does not match this package name %q", other.Name, p.Name)
	}
	p.Atoms = append(p.Atoms, other.Atoms...)
	p.units = append(p.units, other.units...)
	return nil
}

// Extract components from a parse.SourceUnit type.
func Extract(su parse.SourceUnit) (Package, error) {
	pkg, err := findPackage(su.Decls)
	if err != nil {
		return Package{}, err
	}

	pkg.units = append(pkg.units, su)
	return pkg, nil
}
