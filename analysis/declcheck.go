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

package analysis

import (
	"fmt"

	"github.com/google/mangle/ast"
	"github.com/google/mangle/parse"
	"github.com/google/mangle/symbols"
)

type declChecker struct {
	decl ast.Decl
	errs []error
}

func newDeclChecker(decl ast.Decl) *declChecker {
	return &declChecker{decl, nil}
}

// CheckDecl performs context-free checks to see whether a decl is well-formed.
func CheckDecl(decl ast.Decl) []error {
	return newDeclChecker(decl).check()
}

func (c *declChecker) check() []error {
	p := c.decl.DeclaredAtom
	var seenDocAtom bool
	expectedArgs := make(map[ast.Variable]struct{}, len(p.Args))
	for _, arg := range p.Args {
		v, ok := arg.(ast.Variable)
		if !ok {
			c.errs = append(c.errs, fmt.Errorf("Decl requires an atom with variables got %v", arg))
			continue
		}
		expectedArgs[v] = struct{}{}
	}
	if c.errs != nil {
		return c.errs
	}
	for _, descrAtom := range c.decl.Descr {
		sym := descrAtom.Predicate.Symbol
		switch sym {
		case ast.DescrDoc:
			if seenDocAtom {
				c.errs = append(c.errs, fmt.Errorf("descr[] can only have one doc atom"))
			}
			seenDocAtom = true
			if len(descrAtom.Args) == 0 {
				c.errs = append(c.errs, fmt.Errorf("descr atom must not be empty"))
				continue
			}
			for _, docArg := range descrAtom.Args {
				c.checkStringConstant(docArg)
			}
		case ast.DescrArg:
			if len(descrAtom.Args) < 2 {
				c.errs = append(c.errs, fmt.Errorf("arg atom must have at least 2 args"))
				continue
			}
			firstArg := descrAtom.Args[0]
			v, ok := firstArg.(ast.Variable)
			if !ok {
				c.errs = append(c.errs, fmt.Errorf("arg atom must have variable as arg, got%v", firstArg))
				continue
			}
			if _, ok := expectedArgs[v]; !ok {
				c.errs = append(c.errs, fmt.Errorf("arg atom for an unknown variable %v", v))
				continue
			}
			delete(expectedArgs, v)
			for _, argArg := range descrAtom.Args[1:] {
				c.checkStringConstant(argArg)
			}
		default:
			// We ignore unknown descr atoms.
		}
	}
	if !c.decl.IsSynthetic() && len(expectedArgs) > 0 && len(expectedArgs) != len(p.Args) {
		c.errs = append(c.errs, fmt.Errorf("missing arg atoms for arguments %v", expectedArgs))
	}
	for _, boundDecl := range c.decl.Bounds {
		c.checkBound(p, boundDecl)
	}
	return c.errs
}

// Checks that a boundDecl is well-formed.
// It validates that there is a bound for each argument, and
// that each bound is an approppriate bound expression.
func (c *declChecker) checkBound(p ast.Atom, boundDecl ast.BoundDecl) {
	if len(boundDecl.Bounds) != len(p.Args) {
		c.errs = append(c.errs, fmt.Errorf("in decl %v: expected %d bounds, got %d: %v ", p, len(p.Args), len(boundDecl.Bounds), boundDecl.Bounds))
	}
	for i, bound := range boundDecl.Bounds {
		if err := checkBoundExpression(bound); err != nil {
			c.errs = append(c.errs, fmt.Errorf("in decl %v: the bound for argument %d must be parseable as predicate name: %v ", p, i, bound))
		}
	}
}

func checkBoundExpression(b ast.BaseTerm) error {
	if err := symbols.CheckTypeExpression(b); err != nil {
		// Not a type expression.
		predicateBound, ok := b.(ast.Constant)
		if !ok || predicateBound.Type != ast.StringType {
			return fmt.Errorf("not a bound expression %v %T %v", b, b, err)
		}
		name := predicateBound.Symbol[1 : len(predicateBound.Symbol)-1]
		if _, err := parse.PredicateName(name); err != nil {
			return fmt.Errorf("could not parse predicate name %q", name)
		}
	}
	return nil
}

// Checks that a base term is a string constant.
func (c *declChecker) checkStringConstant(baseTerm ast.BaseTerm) {
	con, ok := baseTerm.(ast.Constant)
	if !ok {
		c.errs = append(c.errs, fmt.Errorf("expected string constant, got %v", baseTerm))
		return
	}
	if con.Type != ast.StringType {
		c.errs = append(c.errs, fmt.Errorf("expected string constant, got %v", c))
	}
}
