// Copyright 2022 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package ast

import (
	"fmt"
	"strings"
)

const (
	// DescrExtensional is a descriptor for extensional predicates.
	DescrExtensional = "extensional"
	// DescrMode is a descriptor for a supported mode of a predicate.
	DescrMode = "mode"
	// DescrReflects is a descriptor for predicates that test ("reflect") name prefixes
	DescrReflects = "reflects"
	// DescrSynthetic is a descriptor for synthetic declarations.
	DescrSynthetic = "synthetic"
	// DescrPrivate is a descriptor for a predicate with package-private visibility.
	DescrPrivate = "private"
	// DescrDoc is a descriptor containing documentation.
	DescrDoc = "doc"
	// DescrArg is a descriptor containing documentation for an argument.
	DescrArg = "arg"
	// DescrName is a descriptor used internally for naming a declared object.
	DescrName = "name"
	// DescrDesugared is a descriptor used internally to mark desugared declarations.
	DescrDesugared = "desugared"
)

// Decl is a declaration.
type Decl struct {
	// Predicate we are defining, with arguments.
	DeclaredAtom Atom

	// Description atoms for this predicate.
	Descr []Atom

	// Upper bounds (may be empty). Each BoundDecl has a matching arity.
	// For example a foo(X,Y) may have bounds (/string x /string) and
	// (foo x /number), which means that X may be /string or foo(X).
	// The idea is that there are a few special unary predicates describing
	// something that could be considered a "type."
	// The bounds form a union (join) so only one of the needs to hold.
	Bounds []BoundDecl

	// Either nil, or an inclusion constraint that holds.
	Constraints *InclusionConstraint
}

// Iterates over all descriptor atoms if cb is non-nil.
// Otherwise returns true if found.
func (d Decl) findDescr(descr string, cb func(Atom)) bool {
	var found bool
	for _, a := range d.Descr {
		if a.Predicate.Symbol == descr {
			found = true
			if cb != nil {
				cb(a)
			} else {
				return true
			}
		}
	}
	return found
}

// Doc returns the doc strings from the Decl's description atoms.
func (d Decl) Doc() []string {
	var res []string
	d.findDescr(DescrDoc, func(a Atom) {
		for _, arg := range a.Args {
			c, ok := arg.(Constant)
			if !ok {
				return
			}
			str, err := c.StringValue()
			if err != nil {
				return
			}
			res = append(res, str)
		}
	})
	return res
}

// IsExtensional returns true if decl is for an extensional predicate.
func (d Decl) IsExtensional() bool {
	return d.findDescr(DescrExtensional, nil)
}

// IsDesugared returns true if decl has been desugared.
func (d Decl) IsDesugared() bool {
	return d.findDescr(DescrDesugared, nil)
}

// ArgMode is an enum specifying whether an argument is input, output, or both.
type ArgMode int

const (
	// ArgModeInput indicates an input to the predicate "+"
	ArgModeInput ArgMode = 1
	// ArgModeOutput indicates an output to the predicate "-"
	ArgModeOutput = 2
	// ArgModeInputOutput indicates that an argument can be either input or output.
	ArgModeInputOutput = 3
)

const (
	// InputString indicates an argument that is input.
	InputString = "+"
	// OutputString indicates an argument that is output.
	OutputString = "-"
	// InputOutputString indicates an argument that can either be input or output.
	InputOutputString = "?"
)

// Mode specifies the mode of the predicate. A tuple of argument modes.
type Mode []ArgMode

// Check checks a goal against this mode and returns an error if it is incompatible.
func (m Mode) Check(goal Atom, boundVars map[Variable]bool) error {
	isFree := func(v Variable) bool {
		return boundVars == nil || !boundVars[v]
	}
	if len(m) != len(goal.Args) {
		return fmt.Errorf("number of arguments, %v, does not match the mode %v", goal.Args, m)
	}
	for i, argMode := range m {
		arg := goal.Args[i]
		switch argMode {
		case ArgModeInput:
			if v, ok := arg.(Variable); ok && isFree(v) {
				return fmt.Errorf("for goal %q expected %v (arg %d) to be constant or bound variable", goal, arg, i)
			}
		case ArgModeOutput:
			if v, ok := arg.(Variable); !ok || !isFree(v) {
				return fmt.Errorf("for goal %q expected %v (arg %d) to be a free variable", goal, arg, i)
			}
		}
	}
	return nil
}

// Modes returns the supported modes declared for this predicate.
// Returns nil if the declaration does not declare modes. This is
// always interpreted as supporting all modes i.e. "?" ... "?".
func (d Decl) Modes() []Mode {
	convertMode := func(args []BaseTerm) Mode {
		var mode Mode
		for _, arg := range args {
			m, ok := arg.(Constant)
			if !ok || m.Type != StringType {
				return nil
			}
			switch m.Symbol {
			case InputString:
				mode = append(mode, ArgModeInput)
			case OutputString:
				mode = append(mode, ArgModeOutput)
			case InputOutputString:
				mode = append(mode, ArgModeInputOutput)
			default:
				return nil
			}
		}
		return mode
	}
	var modes []Mode
	d.findDescr(DescrMode, func(a Atom) {
		mode := convertMode(a.Args)
		if len(mode) > 0 {
			modes = append(modes, mode)
		}
	})
	return modes
}

// PackageID returns the package part (dirname).
func (d Decl) PackageID() string {
	p := d.DeclaredAtom.Predicate
	if p.Symbol == "Package" {
		for _, a := range d.Descr {
			if a.Predicate.Symbol == DescrName {
				c := a.Args[0].(Constant)
				s, _ := c.StringValue()
				return s
			}
		}
	}
	if lastDot := strings.LastIndex(p.Symbol, "."); lastDot != -1 {
		return p.Symbol[:lastDot]
	}
	return ""
}

// Visible returns whether the predicate should be visible to other packages.
func (d Decl) Visible() bool {
	// TODO: Swap default value when corresponding declarations are public.
	return !d.findDescr(DescrPrivate, nil)
}

// IsSynthetic returns true if this Decl is synthetic (generated).
func (d Decl) IsSynthetic() bool {
	return d.findDescr(DescrSynthetic, nil)
}

// Reflects returns (true, prefix) if this predicate covers ("reflects") a name prefix type.
func (d Decl) Reflects() (Constant, bool) {
	var (
		found bool
		c     Constant
	)
	d.findDescr(DescrReflects, func(a Atom) {
		if len(a.Args) != 1 {
			return
		}
		if name, ok := (a.Args[0]).(Constant); ok {
			found = true
			c = name
		}
	})
	return c, found
}

// BoundDecl is a bound declaration for the arguments of a predicate.
//
// A bound declaration is either a type-expression (a type bound) or
// a reference to a unary predicate (a predicate bound).
//
// The BaseTerm t represents a set of constants |t| as follows:
// - |/any| is a type whose elements are all values
// - |/name| is a type whose elements are name constants
// - |/number| is a type whose elements are numbers
// - |/string| is a type whose elements are strings
// - "$pred" if the model satisfies $pred(X). This is not a type but
//
//	a short way to add an inclusion constraint. Predicate bounds
//	must not be recursive and resolvable to a type bound t which
//	can be used for type-checking. This |t| is used as an
//	upper bound for the elements.
//
// (In the following all subexpression must be type bounds:)
// - fn:Pair(s, t) if the argument is a pair whose first
//
//	element is in s and whose second element is in t.
//
// - fn:List(t) if the argument is a list
// - fn:Tuple(s1,...,sN) for N > 2 is a shorthand for type
//
//	fn:Pair(s1, fn:Pair(...fn:Pair(N-1,sN)...))
//
// - fn:Union(s1,...,sN) is the type whose elements are
//
//	included in at least one of types s1, ..., sN.
//
// This list is incomplete in two ways:
// - There are type expressions that are not permitted in bound
// declarations, and
// - extensions may provide further type expressions that are permitted
// in bound declarations.
type BoundDecl struct {
	Bounds []BaseTerm
}

// NewBoundDecl returns a new BoundDecl.
func NewBoundDecl(bounds ...BaseTerm) BoundDecl {
	return BoundDecl{bounds}
}

// InclusionConstraint expresses e.g. that if foo(X, Y) holds,
// then also bar(X) and baz(Y) and xyz(X,Y) hold. This can be a
// stronger constraint than bound declarations.
//
// Such an inclusion constraint may be entered by the user like this:
//
// Decl foo(X, Y)
//
//	bound ["even", /number]
//	bound [/number, "odd"]
//	inclusion [bar(X), baz(Y), xyz(X,Y)].
//
// In order to account for bound declarations with predicate references,
// which are also inclusion constraints, declarations are desugared.
// After desugaring, there is one alternative for each bound declaration,
// all inclusion constraints appear in Alternatives and Consequences
// is empty, like so (this is not real syntax:)
//
// Decl foo(X, Y)
//
//	bound [/number, /number]
//	bound [/number, /number]
//	inclusion-alternative[ even(X), bar(X), baz(Y), xyz(X,Y)].
//	inclusion-alternative[ odd(Y), bar(X), baz(Y), xyz(X,Y)].
type InclusionConstraint struct {
	// All of these must hold.
	Consequences []Atom
	// In addition to Consequences, at least one of these must hold.
	Alternatives [][]Atom
}

// NewInclusionConstraint returns a new InclusionConstraint.
func NewInclusionConstraint(consequences []Atom) InclusionConstraint {
	return InclusionConstraint{consequences, nil}
}

// NewDecl returns a new Decl.
func NewDecl(atom Atom, descrAtoms []Atom, bounds []BoundDecl, constraints *InclusionConstraint) (Decl, error) {
	if descrAtoms == nil {
		descrAtoms = []Atom{NewAtom(DescrDoc, String(""))}
	}
	for i, arg := range atom.Args {
		if _, ok := arg.(Variable); ok {
			continue
		}
		return Decl{}, fmt.Errorf("argument %d must be a variable, found %v", i, arg)
	}
	return Decl{atom, descrAtoms, bounds, constraints}, nil
}

// NewSyntheticDecl returns a new Decl from an atom.
// The decl has an empty doc, an explicit mode supporting all arguments as input-output,
// and is marked as being synthetic.
func NewSyntheticDecl(declaredAtom Atom) (Decl, error) {
	modeArgs := make([]BaseTerm, declaredAtom.Predicate.Arity)
	unknownBounds := make([]BaseTerm, declaredAtom.Predicate.Arity)
	for i := 0; i < declaredAtom.Predicate.Arity; i++ {
		unknownBounds[i] = AnyBound
		modeArgs[i] = String(InputOutputString)
	}
	descrAtoms := []Atom{
		NewAtom(DescrDoc, String("")), NewAtom(DescrMode, modeArgs...), NewAtom(DescrSynthetic)}

	used := make(map[Variable]bool)
	AddVars(declaredAtom, used)
	args := make([]BaseTerm, len(declaredAtom.Args))
	for i, arg := range declaredAtom.Args {
		if _, ok := arg.(Variable); ok {
			args[i] = arg
		} else {
			args[i] = FreshVariable(used)
		}
	}
	return NewDecl(Atom{declaredAtom.Predicate, args}, descrAtoms, []BoundDecl{{unknownBounds}}, nil)
}

// NewSyntheticDeclFromSym returns a new Decl from a predicate symbol.
func NewSyntheticDeclFromSym(sym PredicateSym) Decl {
	// NewQuery creates an atom with only variable arguments, cannot throw an error.
	// In general, creating a synthetic declaration from a symbol must always
	// be possible without error, but we are reusing a more general func here.
	decl, _ := NewSyntheticDecl(NewQuery(sym))
	return decl
}
