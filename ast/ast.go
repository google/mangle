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

// Package ast contains abstract syntax tree representations of Mangle code.
package ast

import (
	"encoding/binary"
	"fmt"
	"hash/fnv"
	"math"
	"regexp"
	"sort"
	"strings"
)

// AnyBound is a type expression that has all values as elements.
var AnyBound Constant

// Float64Bound is a type expression that has all float64s as elements.
var Float64Bound Constant

// NameBound is a type expression that has all names as elements.
var NameBound Constant

// StringBound is a type expression that has all strings as elements.
var StringBound Constant

// NumberBound is a type expression that has all numbers as elements.
var NumberBound Constant

// TruePredicate is a predicate symbol to represent an
// "unconditionally true" proposition.
var TruePredicate = PredicateSym{"true", 0}

// FalsePredicate is a predicate symbol to represent an
// "unconditionally false" proposition.
var FalsePredicate = PredicateSym{"false", 0}

// TrueAnd represents true() in empty conjunction form.
var TrueAnd = And{}

func init() {
	AnyBound, _ = Name("/any")
	Float64Bound, _ = Name("/float64")
	NameBound, _ = Name("/name")
	NumberBound, _ = Name("/number")
	StringBound, _ = Name("/string")
}

// FormatNumber turns a number constant into a string.
func FormatNumber(num int64) string {
	return fmt.Sprintf("%d", num)
}

// FormatFloat64 turns a float64 constant into a string.
func FormatFloat64(floatNum float64) string {
	return fmt.Sprintf("%f", floatNum)
}

// Term represents the building blocks of datalog programs, namely constants, variables, atoms,
// and also negated atoms, equality and inequality.
//
// Some minor differences to formal mathematical logic and prolog-style logic programming:
// - an Atom in datalog is always a predicate symbol applied to constants or variables ("base terms"
// - an equality and inequality would be called a "formula" but we do not need a separate interface
// - Atom and NegatedAtom instances would be called "literals", but again one interface is enough.
//
// Note that constants start with "/" whereas variables start with a capital letter  name. This
// convention gives us a 1:1 mapping between term objects and their string representations, and
// this is useful because Atom is not a hashable type (use String() to use Term as key in maps).
type Term interface {
	// Marker method.
	isTerm()

	// Returns a string representation.
	String() string

	// Syntactic (or structural) equality.
	// If Equals returns true, terms have the same string representation.
	Equals(Term) bool

	// Returns a new term.
	ApplySubst(s Subst) Term
}

// BaseTerm represents a subset of terms: constant or variables.
// Every BaseTerm will implement Term.
// TODO: Rename to Expr.
type BaseTerm interface {
	Term

	// Marker method.
	isBaseTerm()

	Hash() uint64

	// Returns a new base term.
	ApplySubstBase(s Subst) BaseTerm
}

// Subst is the interface for substitutions.
type Subst interface {
	// Returns the term the given variable maps to, or nil if the variable is not in domain.
	Get(Variable) BaseTerm
}

// ConstSubstMap is a substitution backed by a map from variables to constants.
type ConstSubstMap map[Variable]Constant

// Get implements the Get method from Subst.
func (m ConstSubstMap) Get(v Variable) BaseTerm {
	return m[v]
}

// ConstSubstPair represents a (variable, constant) pair.
type ConstSubstPair struct {
	v Variable
	c Constant
}

// ConstSubstList is a substitution backed by a slice of (variable, constant) pairs.
type ConstSubstList []ConstSubstPair

// Get implements the Get method from Subst.
func (c ConstSubstList) Get(v Variable) BaseTerm {
	for _, x := range c {
		if x.v == v {
			return x.c
		}
	}
	return nil
}

// Extend extends this substitution with a new binding.
func (c ConstSubstList) Extend(v Variable, con Constant) ConstSubstList {
	return append(c, ConstSubstPair{v, con})
}

// ConstantType describes the primitive type or shape of a constant.
type ConstantType int

const (
	// NameType is the type of name constants.
	NameType ConstantType = iota
	// StringType is the type of string constants.
	StringType
	// NumberType is the type of number (int64) constants.
	NumberType
	// Float64Type is the type of float64 constants.
	Float64Type
	// PairShape indicates that the constant is a pair.
	PairShape
	// ListShape indicates that the constant is a list.
	ListShape
	// MapShape indicates that the constant is a map.
	// Internally, we just represent this as list of ($key, $value) pairs.
	MapShape
	// StructShape indicates that the constant is a struct.
	// Internally, we just represent this as a list of (/field/$name, $value) pairs.
	StructShape
)

var number = regexp.MustCompile(`-?\d+`)

// Constant represents a constant symbol.
type Constant struct {
	// The (runtime) type of this constant.
	Type ConstantType

	// String representation of the symbol.
	// For a structured name "/foo/bar", for a string `/"content"`, for a number "/[123]",
	// For a struct (proto), the deterministically marshalled bytes.
	Symbol string

	// For NumberType, the number value. For PairShape and
	// ListShape, it contains a hash code.
	NumValue int64

	// For a pair constant, the first component.
	// For a list constant, the head of the list.
	// For a map constant, the first map entry.
	// For a struct constant, the first field.
	fst *Constant

	// For a pair constant, the second component
	// For a list constant, the tail of the list.
	// For a map constant, the tail of the entries.
	// For a struct constant, the tail of the fields.
	snd *Constant
}

// Name constructs a new name constant while checking that constant symbol starts with a '/' and
// does not contain empty parts.
func Name(symbol string) (Constant, error) {
	switch {
	case len(symbol) <= 1:
		return Constant{}, fmt.Errorf("constant symbol must be a non-empty string starting with '/'")
	case symbol[0] != '/':
		return Constant{}, fmt.Errorf("constant symbol must start with '/'")
	}
	if strings.Contains(symbol, "\"") {
		return Constant{}, fmt.Errorf("this constructor does not handle string content \"%s\"", symbol)
	}
	for _, part := range strings.Split(symbol[1:], "/") {
		if part == "" {
			return Constant{}, fmt.Errorf("constant symbol \"%s\" contains empty part", symbol)
		}
	}
	return Constant{NameType, symbol, 0, nil, nil}, nil
}

// String constructs a "constant symbol" that contains an arbitrary string.
func String(str string) Constant {
	return Constant{StringType, str, 0, nil, nil}
}

// Number constructs a constant symbol that contains a number.
func Number(num int64) Constant {
	return Constant{NumberType, "", num, nil, nil}
}

// Float64 constructs a constant symbol that contains a float64.
func Float64(floatNum float64) Constant {
	return Constant{Float64Type, "", int64(math.Float64bits(floatNum)), nil, nil}
}

// Pair constructs a pair constant. Parts can only be accessed in transforms.
func pair(tpe ConstantType, fst, snd *Constant) Constant {
	return Constant{tpe, "", hashPair(fst, snd, tpe), fst, snd}
}

// Pair constructs a pair constant. Parts can only be accessed in transforms.
func Pair(fst, snd *Constant) Constant {
	return pair(PairShape, fst, snd)
}

// ListCons constructs a list, using pairs. Parts can only be accessed in transforms.
func ListCons(fst, snd *Constant) Constant {
	return pair(ListShape, fst, snd)
}

// MapCons constructs a map, using pairs. Parts can only be accessed in transforms.
func MapCons(key, val, rest *Constant) Constant {
	e := pair(PairShape, key, val)
	return pair(MapShape, &e, rest)
}

// StructCons constructs a struct, using pairs. Parts can only be accessed in transforms.
func StructCons(label, val, rest *Constant) Constant {
	e := pair(PairShape, label, val)
	return pair(StructShape, &e, rest)
}

// ListNil represents an empty list.
var ListNil = Constant{ListShape, "", 0, nil, nil}

// MapNil represents an empty map.
var MapNil = Constant{MapShape, "", 0, nil, nil}

// StructNil represents an empty struct.
var StructNil = Constant{StructShape, "", 0, nil, nil}

// List constructs a list constant. Parts can only be accessed in transforms.
func List(constants []Constant) Constant {
	list := &ListNil
	if constants == nil {
		return *list
	}
	for i := len(constants) - 1; i >= 0; i-- {
		next := ListCons(&constants[i], list)
		list = &next
	}
	return *list
}

func canonicalOrder(left *Constant, right *Constant) {

}

// Map constructs a map constant. Parts can only be accessed in transforms.
// Keys and values must come in the order.
func Map(kvMap map[*Constant]*Constant) *Constant {
	m := &MapNil
	if len(kvMap) == 0 {
		return m
	}
	keys := make([]*Constant, len(kvMap))
	vals := make([]*Constant, len(kvMap))
	i := 0
	for k, v := range kvMap {
		keys[i] = k
		vals[i] = v
		i++
	}
	index := make([]int, len(kvMap))
	SortIndexInto(keys, index)
	for _, i := range index {
		next := MapCons(keys[i], vals[i], m)
		m = &next
	}
	return m
}

// Struct constructs a struct constant. Parts can only be accessed in transforms.
// labels must be sorted .
func Struct(kvMap map[*Constant]*Constant) *Constant {
	m := &StructNil
	if len(kvMap) == 0 {
		return m
	}
	labels := make([]*Constant, len(kvMap))
	vals := make([]*Constant, len(kvMap))
	i := 0
	for k, v := range kvMap {
		labels[i] = k
		vals[i] = v
		i++
	}
	index := make([]int, len(kvMap))
	SortIndexInto(labels, index)
	for _, i := range index {
		next := StructCons(labels[i], vals[i], m)
		m = &next
	}
	return m
}

// StringValue returns the string value of this constant, if it is of type string.
func (c Constant) StringValue() (string, error) {
	if c.Type != StringType {
		return "", fmt.Errorf("not a string constant %v", c)
	}
	return c.Symbol, nil
}

// NumberValue returns the number(int64) value of this constant, if it is of type number.
func (c Constant) NumberValue() (int64, error) {
	if c.Type != NumberType {
		return 0, fmt.Errorf("not a number constant %v", c)
	}
	return c.NumValue, nil
}

// Float64Value returns the float64 value of this constant, if it is of type float64.
func (c Constant) Float64Value() (float64, error) {
	if c.Type != Float64Type {
		return 0, fmt.Errorf("not a number constant %v", c)
	}
	return math.Float64frombits(uint64(c.NumValue)), nil
}

// PairValue returns the two constants that make up this pair.
func (c Constant) PairValue() (Constant, Constant, error) {
	if c.Type != PairShape {
		return Constant{}, Constant{}, fmt.Errorf("not a pair value %v", c)
	}
	return *c.fst, *c.snd, nil
}

// ConsValue returns the two constants that make up this cons.
func (c Constant) ConsValue() (Constant, Constant, error) {
	if c.Type != ListShape || c.IsListNil() {
		return Constant{}, Constant{}, fmt.Errorf("not a cons value %v", c)
	}
	return *c.fst, *c.snd, nil
}

// ListValues provides the constants that make up the list via callback.
func (c Constant) ListValues(cbCons func(Constant) error, cbNil func() error) (error, error) {
	if c.Type != ListShape {
		return fmt.Errorf("not a list constant %v", c), nil
	}
	for ; !c.IsListNil(); c = *c.snd {
		if err := cbCons(*c.fst); err != nil {
			return nil, err
		}
	}
	return nil, cbNil()
}

// MapValues provides the entries of the map via key-value callback.
func (c Constant) MapValues(cbCons func(Constant, Constant) error, cbNil func() error) (error, error) {
	if c.Type != MapShape {
		return fmt.Errorf("not a map constant %v", c), nil
	}
	for ; !c.IsMapNil(); c = *c.snd {
		p := c.fst
		if p.Type != PairShape {
			return fmt.Errorf("not a struct field %v", p), nil
		}
		if err := cbCons(*p.fst, *p.snd); err != nil {
			return nil, err
		}
	}
	return nil, cbNil()
}

// StructValues provides the entries that make up the struct via label-value callback.
func (c Constant) StructValues(cbCons func(Constant, Constant) error, cbNil func() error) (error, error) {
	if c.Type != StructShape {
		return fmt.Errorf("not a struct constant %v", c), nil
	}
	for ; !c.IsStructNil(); c = *c.snd {
		p := c.fst
		if p.Type != PairShape {
			return fmt.Errorf("not a map entry %v", p), nil
		}
		if err := cbCons(*p.fst, *p.snd); err != nil {
			return nil, err
		}
	}
	return nil, cbNil()
}

func (c Constant) isBaseTerm() {
}

func (c Constant) isTerm() {
}

// String returns a string representation of the constant.
func (c Constant) String() string {
	switch c.Type {
	case NameType:
		return c.Symbol
	case StringType:
		str := c.Symbol
		if strings.ContainsRune(str, '\n') {
			return fmt.Sprintf("`%s`", str)
		}
		str = strings.ReplaceAll(str, `\`, `\\`)
		str = strings.ReplaceAll(str, `"`, `\"`)
		return fmt.Sprintf(`"%s"`, str)
	case NumberType:
		return FormatNumber(c.NumValue)
	case Float64Type:
		return FormatFloat64(math.Float64frombits(uint64(c.NumValue)))
	case PairShape:
		fst := *c.fst
		snd := *c.snd
		return fmt.Sprintf("<%s; %s>", fst.String(), snd.String())
	case ListShape:
		if c.IsListNil() {
			return "[]"
		}
		var s strings.Builder
		s.WriteRune('[')
		s.WriteString((*c.fst).String())
		c = *c.snd
		for !c.IsListNil() {
			s.WriteString(", ")
			s.WriteString((*c.fst).String())
			c = *c.snd
		}
		s.WriteRune(']')
		return s.String()
	case MapShape:
		if c.IsMapNil() {
			return "fn:map()"
		}
		var s strings.Builder
		s.WriteRune('[')
		s.WriteString((*c.fst.fst).String())
		s.WriteString(" : ")
		s.WriteString((*c.fst.snd).String())
		c = *c.snd
		for !c.IsMapNil() {
			s.WriteString(", ")
			s.WriteString((*c.fst.fst).String())
			s.WriteString(" : ")
			s.WriteString((*c.fst.snd).String())
			c = *c.snd
		}
		s.WriteRune(']')
		return s.String()

	case StructShape:
		if c.IsStructNil() {
			return "{}"
		}
		var s strings.Builder
		s.WriteRune('{')
		s.WriteString((*c.fst.fst).String())
		s.WriteString(" : ")
		s.WriteString((*c.fst.snd).String())
		c = *c.snd
		for !c.IsStructNil() {
			s.WriteString(", ")
			s.WriteString((*c.fst.fst).String())
			s.WriteString(" : ")
			s.WriteString((*c.fst.snd).String())
			c = *c.snd
		}
		s.WriteRune('}')
		return s.String()

	default:
		return "?" // cannot happen
	}
}

// IsListNil returns true if this constant represents the empty list.
func (c Constant) IsListNil() bool {
	return c.Type == ListShape && c.fst == nil
}

// IsMapNil returns true if this constant represents the empty map.
func (c Constant) IsMapNil() bool {
	return c.Type == MapShape && c.fst == nil
}

// IsStructNil returns true if this constant represents the empty struct.
func (c Constant) IsStructNil() bool {
	return c.Type == StructShape && c.fst == nil
}

// Equals returns true if u is the same constant.
func (c Constant) Equals(u Term) bool {
	var uconst Constant
	if v, ok := u.(*Constant); ok {
		uconst = *v
	} else if uconst, ok = u.(Constant); !ok || c.Type != uconst.Type {
		return false
	}
	if c.NumValue != uconst.NumValue {
		return false
	}
	// At this point, we know that constants have the same hash.
	switch c.Type {
	case NameType:
		fallthrough
	case StringType:
		return c.Symbol == uconst.Symbol
	case NumberType:
		fallthrough
	case Float64Type:
		return true
	case PairShape:
		return c.fst.Equals(uconst.fst) && c.snd.Equals(uconst.snd)
	case ListShape:
		if c.IsListNil() {
			return uconst.IsListNil()
		}
		if uconst.IsListNil() {
			return false
		}
		if c.NumValue != uconst.NumValue {
			return false
		}
		return c.fst.Equals(uconst.fst) && c.snd.Equals(uconst.snd)

	case MapShape:
		// same keys, same values
		if c.IsMapNil() {
			return uconst.IsMapNil()
		}
		if uconst.IsMapNil() {
			return false
		}
		if c.NumValue != uconst.NumValue {
			return false
		}
		return c.fst.Equals(uconst.fst) && c.snd.Equals(uconst.snd)

	case StructShape:
		if c.IsStructNil() {
			return uconst.IsStructNil()
		}
		if c.NumValue != uconst.NumValue {
			return false
		}
		if uconst.IsStructNil() {
			return false
		}
		return c.fst.Equals(uconst.fst) && c.snd.Equals(uconst.snd)
	}
	return false //  cannot happen, all cases covered.
}

func hashBytes(s []byte) uint64 {
	h := fnv.New64()
	h.Write(s)
	return h.Sum64()
}

// Computes a hash. The snd argument may be nil.
func hashPair(fst, snd *Constant, tpe ConstantType) int64 {
	left := fst.Hash()
	switch tpe {
	case MapShape:
		left = left << 1
	case StructShape:
		left = left << 2
	}
	left = left<<19 - left
	if snd == nil {
		return int64(left)
	}
	return int64(left) + int64(snd.Hash())
}

// Hash returns a hash code for this constant
func (c Constant) Hash() uint64 {
	if c.Type == StringType || c.Type == NameType {
		return hashBytes([]byte(c.Symbol))
	}
	return uint64(c.NumValue)
}

// ApplySubst simply returns this constant, for any substitution.
func (c Constant) ApplySubst(s Subst) Term {
	return c
}

// ApplySubstBase simply returns this constant, for any substitution.
func (c Constant) ApplySubstBase(s Subst) BaseTerm {
	return c
}

// PredicateSym represents a predicate symbol with a given arity.
type PredicateSym struct {
	Symbol string
	Arity  int
}

// InternalPredicateSuffix gets appended to all internal predicate symbol names.
const InternalPredicateSuffix = "__tmp"

// IsInternalPredicate returns true if predicate symbol belongs to a generated predicate name.
func (p PredicateSym) IsInternalPredicate() bool {
	return strings.HasSuffix(p.Symbol, InternalPredicateSuffix)
}

func (p PredicateSym) String() string {
	var args []string
	for i := 0; i < p.Arity; i++ {
		args = append(args, fmt.Sprintf("%c", 'A'+i))
	}
	return fmt.Sprintf("%s(%s)", p.Symbol, strings.Join(args, ", "))
}

// IsBuiltin returns true if this predicate symbol is for a built-in predicate.
func (p PredicateSym) IsBuiltin() bool {
	return strings.HasPrefix(p.Symbol, ":")
}

// Variable represents a variable.
type Variable struct {
	Symbol string
}

func (v Variable) isBaseTerm() {
}

func (v Variable) isTerm() {}

// Hash returns a hash code.
func (v Variable) Hash() uint64 {
	return hashTerm(v.Symbol, []BaseTerm{v})
}

// String simply returns the variable's name.
func (v Variable) String() string {
	return v.Symbol
}

// Equals provides syntactic equality for variables.
func (v Variable) Equals(u Term) bool {
	o, ok := u.(Variable)
	return ok && v.Symbol == o.Symbol
}

// ApplySubst returns the result of applying the given substituion.
func (v Variable) ApplySubst(s Subst) Term {
	return v.ApplySubstBase(s)
}

// ApplySubstBase returns the result of applying the given substituion.
func (v Variable) ApplySubstBase(s Subst) BaseTerm {
	if t := s.Get(v); t != nil {
		return t
	}
	return v
}

// Atom represents an atom (a predicate symbol applied to base term arguments).
type Atom struct {
	Predicate PredicateSym
	Args      []BaseTerm
}

func (a Atom) isTerm() {}

// String returns a string representation for this atom.
func (a Atom) String() string {
	var sb strings.Builder
	sb.WriteString(a.Predicate.Symbol)
	sb.WriteString("(")
	for i, arg := range a.Args {
		if i > 0 {
			sb.WriteString(",")
		}
		sb.WriteString(arg.String())
	}
	sb.WriteString(")")
	return sb.String()
}

// Equals provides syntactic equality for atoms.
func (a Atom) Equals(u Term) bool {
	o, ok := u.(Atom)
	if !ok {
		return false
	}
	if a.Predicate != o.Predicate ||
		len(a.Args) != len(o.Args) {
		return false
	}
	for i, arg := range a.Args {
		if !arg.Equals(o.Args[i]) {
			return false
		}
	}
	return true
}

// Hash returns a hash code for this atom.
func (a Atom) Hash() uint64 {
	return hashTerm(a.Predicate.String(), a.Args)
}

// ApplySubst returns the result of applying given substitution to this atom.
func (a Atom) ApplySubst(s Subst) Term {
	newargs := make([]BaseTerm, len(a.Args))
	for i, t := range a.Args {
		newargs[i] = t.ApplySubstBase(s)
	}
	return Atom{a.Predicate, newargs}
}

// IsGround returns true if all arguments are constants.
func (a Atom) IsGround() bool {
	for _, term := range a.Args {
		if _, ok := term.(Constant); !ok {
			return false
		}
	}
	return true
}

// NewAtom is a convenience constructor for Atom.
func NewAtom(predicateSym string, args ...BaseTerm) Atom {
	return Atom{PredicateSym{predicateSym, len(args)}, args}
}

// NewQuery is a convenience constructor for constructing a goal atom.
func NewQuery(predicate PredicateSym) Atom {
	vars := make([]BaseTerm, predicate.Arity)
	for i := 0; i < predicate.Arity; i++ {
		vars[i] = Variable{fmt.Sprintf("X%d", i)}
	}
	return Atom{predicate, vars}
}

// And represents a conjunction of atoms.
type And struct {
	Atoms []Atom
}

func (a And) isTerm() {}

// String returns a string representation for this atom.
func (a And) String() string {
	return fmt.Sprintf("And(%v)", a.Atoms)
}

// Equals returns true if u is syntactically (structurally) the same conjunction.
func (a And) Equals(u Term) bool {
	z, ok := u.(And)
	if !ok || len(a.Atoms) != len(z.Atoms) {
		return false
	}
	for i, atom := range a.Atoms {
		if !atom.Equals(z.Atoms[i]) {
			return false
		}
	}
	return true
}

// ApplySubst returns the result of applying given substitution to this atom.
func (a And) ApplySubst(s Subst) Term {
	os := make([]Atom, len(a.Atoms))
	for i, atom := range a.Atoms {
		os[i] = atom.ApplySubst(s).(Atom)
	}
	return And{os}
}

// NegAtom represents a negated atom.
type NegAtom struct {
	Atom Atom
}

func (a NegAtom) isTerm() {}

// String returns a string representation for this atom.
func (a NegAtom) String() string {
	return fmt.Sprintf("!%s", a.Atom.String())
}

// Equals returns true if u is syntactically (structurally) the same negated atom.
func (a NegAtom) Equals(u Term) bool {
	o, ok := u.(NegAtom)
	return ok && a.Atom.Equals(o.Atom)
}

// ApplySubst returns the result of applying given substitution to this atom.
func (a NegAtom) ApplySubst(s Subst) Term {
	return NegAtom{a.Atom.ApplySubst(s).(Atom)}
}

// IsGround returns true if all arguments are constants.
func (a NegAtom) IsGround() bool {
	return a.Atom.IsGround()
}

// NewNegAtom is a convenience constructor for NegAtom.
func NewNegAtom(predicateSym string, args ...BaseTerm) NegAtom {
	return NegAtom{NewAtom(predicateSym, args...)}
}

// FunctionSym represents a function symbol with a given arity.
type FunctionSym struct {
	// Symbol is the name of the function, always with "fn:" prefix.
	Symbol string
	Arity  int
}

func (f FunctionSym) String() string {
	var args []string
	for i := 0; i < f.Arity; i++ {
		args = append(args, fmt.Sprintf("%c", 'V'+i))
	}
	return fmt.Sprintf("%s(%s)", f.Symbol, strings.Join(args, ", "))
}

// Transform represents a transformation of the relation of a rule.
type Transform struct {
	Statements []TransformStmt
}

// IsLetTransform returns true if transform is a let-transform.
// The other case is a do-transform which starts with "do fn:group_by()".
func (t Transform) IsLetTransform() bool {
	return t.Statements[0].Var != nil
}

// TransformStmt describes how to transform the relation of a rule.
type TransformStmt struct {
	// The variable to which to assign the result e.g. X in "X = fn:sum(Z)", May be nil.
	Var *Variable
	// An expression that refers to some part of the input relation.
	Fn ApplyFn
}

// ApplyFn is a function application like "fn:max(X)".
type ApplyFn struct {
	Function FunctionSym
	Args     []BaseTerm
}

func (a ApplyFn) isBaseTerm() {}

func (a ApplyFn) isTerm() {}

// Hash returns a hash code for this expression.
func (a ApplyFn) Hash() uint64 {
	return hashTerm(a.Function.String(), a.Args)
}

// String returns a string representation for this atom.
func (a ApplyFn) String() string {
	var sb strings.Builder
	sb.WriteString(a.Function.Symbol)
	sb.WriteRune('(')
	for i, arg := range a.Args {
		if i > 0 {
			sb.WriteString(",")
		}
		sb.WriteString(arg.String())
	}
	sb.WriteRune(')')
	return sb.String()
}

// Equals checks equality of ApplyFn terms.
func (a ApplyFn) Equals(t Term) bool {
	o, ok := t.(ApplyFn)
	if !ok {
		return false
	}
	if a.Function.Symbol != o.Function.Symbol ||
		len(a.Args) != len(o.Args) {
		return false
	}
	for i, arg := range a.Args {
		if !arg.Equals(o.Args[i]) {
			return false
		}
	}
	return true
}

// ApplySubst returns the result of applying given substitution to this atom.
func (a ApplyFn) ApplySubst(s Subst) Term {
	newargs := make([]BaseTerm, len(a.Args))
	for i, t := range a.Args {
		newargs[i] = t.ApplySubstBase(s)
	}
	return ApplyFn{a.Function, newargs}
}

// ApplySubstBase simply returns this constant, for any substitution.
func (a ApplyFn) ApplySubstBase(s Subst) BaseTerm {
	return a.ApplySubst(s).(BaseTerm)
}

// Eq represents an equality (identity constraint) X = Y or X = c or c = X.
type Eq struct {
	Left  BaseTerm
	Right BaseTerm
}

func (e Eq) isTerm() {}

// String returns a string representation for this atom.
func (e Eq) String() string {
	return fmt.Sprintf("%s = %s", e.Left, e.Right)
}

// Equals provides syntactic (structural) equality for Eq(left, right) terms.
func (e Eq) Equals(u Term) bool {
	o, ok := u.(Eq)
	return ok && e.Left.Equals(o.Left) && e.Right.Equals(o.Right)
}

// ApplySubst returns the result of applying given substitution to this equality.
func (e Eq) ApplySubst(s Subst) Term {
	return Eq{e.Left.ApplySubst(s).(BaseTerm), e.Right.ApplySubst(s).(BaseTerm)}
}

// Ineq represents an inequality (apartness constraint) X != Y or X != c or c != X.
type Ineq struct {
	Left  BaseTerm
	Right BaseTerm
}

func (e Ineq) isTerm() {}

// String returns a string representation for this atom.
func (e Ineq) String() string {
	return fmt.Sprintf("%s != %s", e.Left, e.Right)
}

// Equals provides syntactic (structural) equality for Ineq(left, right) terms.
func (e Ineq) Equals(u Term) bool {
	o, ok := u.(Ineq)
	return ok && e.Left.Equals(o.Left) && e.Right.Equals(o.Right)
}

// ApplySubst returns the result of applying given substitution to this inequality.
func (e Ineq) ApplySubst(s Subst) Term {
	return Ineq{e.Left.ApplySubst(s).(BaseTerm), e.Right.ApplySubst(s).(BaseTerm)}
}

// Clause represents a clause (a rule of the form "A." or "A :- B1, ..., Bn.").
// When a clause has a body, the resulting relation can be transformed.
type Clause struct {
	Head      Atom
	Premises  []Term
	Transform *Transform
}

func (c Clause) String() string {
	if c.Premises == nil {
		return fmt.Sprintf("%s.", c.Head.String())
	}
	var premises strings.Builder
	for i, p := range c.Premises {
		if i > 0 {
			premises.WriteString(", ")
		}
		premises.WriteString(p.String())
	}
	if c.Transform == nil {
		return fmt.Sprintf("%s :- %s.", c.Head.String(), premises.String())
	}
	return fmt.Sprintf("%s :- %s |> %s.", c.Head.String(), premises.String(), c.Transform.String())
}

func (t Transform) String() string {
	var transformStmts strings.Builder
	for i, stmt := range t.Statements {
		if i > 0 {
			transformStmts.WriteString(", ")
		}
		if stmt.Var == nil {
			transformStmts.WriteString("do ")
			transformStmts.WriteString(stmt.Fn.String())
		} else {
			transformStmts.WriteString("let ")
			transformStmts.WriteString(stmt.Var.Symbol)
			transformStmts.WriteString(" = ")
			transformStmts.WriteString(stmt.Fn.String())
		}
	}
	return transformStmts.String()
}

// ReplaceWildcards returns a new clause where each wildcard is
// replaced with a fresh variable.
func (c Clause) ReplaceWildcards() Clause {
	vars := make(map[Variable]bool)
	AddVarsFromClause(c, vars)
	if !vars[Variable{"_"}] { // If no wildcards
		return c
	}
	newPremises := make([]Term, len(c.Premises))
	for i, p := range c.Premises {
		newPremises[i] = ReplaceWildcards(vars, p)
	}
	// Wildcards in the rule head are a programmer mistake,
	// there is no way a wildcard can be bound. This is caught
	// by validation.
	return Clause{c.Head, newPremises, c.Transform}
}

// NewClause constructs a new clause.
func NewClause(head Atom, premises []Term) Clause {
	return Clause{head, premises, nil}
}

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

// Doc returns the doc strings from the Decl's description atoms.
func (d Decl) Doc() []string {
	var res []string
	for _, a := range d.Descr {
		if a.Predicate.Symbol == "doc" {
			for _, d := range a.Args {
				c, ok := d.(Constant)
				if !ok {
					continue
				}
				str, err := c.StringValue()
				if err != nil {
					continue
				}
				res = append(res, str)
			}
		}
	}
	return res
}

// PackageID returns the package part (dirname).
func (d Decl) PackageID() string {
	p := d.DeclaredAtom.Predicate
	if p.Symbol == "Package" {
		for _, a := range d.Descr {
			if a.Predicate.Symbol == "name" {
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
	for _, a := range d.Descr {
		if a.Predicate.Symbol == "private" {
			return false
		}
	}
	return true
}

// IsSynthetic returns true if this Decl is synthetic (generated).
func (d Decl) IsSynthetic() bool {
	for _, a := range d.Descr {
		if a.Predicate.Symbol == "synthetic" {
			return true
		}
	}
	return false
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
	Alternatives []And
}

// NewInclusionConstraint returns a new InclusionConstraint.
func NewInclusionConstraint(consequences []Atom) InclusionConstraint {
	return InclusionConstraint{consequences, nil}
}

// NewDecl returns a new Decl.
func NewDecl(atom Atom, descrAtoms []Atom, bounds []BoundDecl, constraints *InclusionConstraint) (Decl, error) {
	if descrAtoms == nil {
		descrAtoms = []Atom{NewAtom("doc", String(""))}
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
func NewSyntheticDecl(declaredAtom Atom) (Decl, error) {
	descrAtoms := []Atom{NewAtom("doc", String("")), NewAtom("synthetic")}
	unknownBounds := make([]BaseTerm, declaredAtom.Predicate.Arity)
	for i := 0; i < declaredAtom.Predicate.Arity; i++ {
		unknownBounds[i] = AnyBound
	}
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

func hashTerm(s string, args []BaseTerm) uint64 {
	h := fnv.New64()
	h.Write([]byte(s))
	for _, arg := range args {
		switch c := arg.(type) {
		case Constant:
			b := make([]byte, 8)
			binary.LittleEndian.PutUint64(b, c.Hash())
			h.Write(b)
		case Variable:
			h.Write([]byte(c.String()))
		case ApplyFn:
			b := make([]byte, 8)
			binary.LittleEndian.PutUint64(b, c.Hash())
			h.Write(b)
		}
	}
	return h.Sum64()
}

// FreshVariable returns a variable different from the ones in used.
func FreshVariable(used map[Variable]bool) Variable {
	makeFresh := func(n int) Variable { return Variable{fmt.Sprintf("X%d", n)} }
	i := 0
	for {
		v := makeFresh(i)
		if used[v] {
			i++
			continue
		}
		return v
	}
}

// ReplaceWildcards returns a new term where each wildcard is replaced
// with a fresh variables. The used-variables map is modified to keep track
// of newly added variables.
func ReplaceWildcards(used map[Variable]bool, term Term) Term {
	numUsed := len(used)
	replaced := term
	switch t := term.(type) {
	case Constant:
		return t
	case Variable:
		if t.Symbol != "_" {
			return t
		}
		v := FreshVariable(used)
		used[v] = true
		return v
	case ApplyFn:
		args := make([]BaseTerm, len(t.Args))
		for i, arg := range t.Args {
			args[i] = ReplaceWildcards(used, arg).(BaseTerm)
		}
		replaced = ApplyFn{t.Function, args}
	case Atom:
		args := make([]BaseTerm, len(t.Args))
		for i, arg := range t.Args {
			args[i] = ReplaceWildcards(used, arg).(BaseTerm)
		}
		replaced = Atom{t.Predicate, args}
	case NegAtom:
		atom := ReplaceWildcards(used, t.Atom).(Atom)
		replaced = NegAtom{atom}
	case Eq:
		left := ReplaceWildcards(used, t.Left).(BaseTerm)
		right := ReplaceWildcards(used, t.Right).(BaseTerm)
		replaced = Eq{left, right}
	case Ineq:
		left := ReplaceWildcards(used, t.Left).(BaseTerm)
		right := ReplaceWildcards(used, t.Right).(BaseTerm)
		replaced = Ineq{left, right}
	}
	if numUsed == len(used) { // If no wildcard found
		return term
	}
	return replaced
}

// AddVars adds all variables from term to map, where term is either
// variable, constant or atom.
func AddVars(term Term, m map[Variable]bool) {
	switch t := term.(type) {
	case Constant:
		return
	case Variable:
		m[t] = true
	case ApplyFn:
		for _, baseTerm := range t.Args {
			AddVars(baseTerm, m)
		}
	case Atom:
		for _, baseTerm := range t.Args {
			AddVars(baseTerm, m)
		}
	case NegAtom:
		AddVars(t.Atom, m)
	case Eq:
		AddVars(t.Left, m)
		AddVars(t.Right, m)
	case Ineq:
		AddVars(t.Left, m)
		AddVars(t.Right, m)
	}
}

// AddVarsFromClause adds all variables from term to map, where term is
// either variable, constant or atom.
func AddVarsFromClause(clause Clause, m map[Variable]bool) {
	AddVars(clause.Head, m)
	for _, p := range clause.Premises {
		AddVars(p, m)
	}
}

// SortIndexInto sorts s and populates the index and hashes slice.
func SortIndexInto(keys []*Constant, index []int) {
	hashes := make([]uint64, len(keys))
	for i := 0; i < len(keys); i++ {
		index[i] = i
		hashes[i] = keys[i].Hash()
	}
	sort.Stable(&keysorter{keys, hashes, index})
}

// Helper to sort []*Constant by it's Hash().
type keysorter struct {
	keys   []*Constant
	hashes []uint64
	index  []int
}

// Len is part of sort.Interface.
func (s keysorter) Len() int {
	return len(s.keys)
}

// Swap is part of sort.Interface.
func (s *keysorter) Swap(i, j int) {
	s.index[i], s.index[j] = s.index[j], s.index[i]
}

// Swap is part of sort.Interface.
func (s *keysorter) Less(i, j int) bool {
	return s.hashes[s.index[i]] < s.hashes[s.index[j]]
}
