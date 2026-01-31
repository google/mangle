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
	"iter"
	"math"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

// AnyBound is a type expression that has all values as elements.
var AnyBound Constant

// BotBound is a type expression that has no elements.
var BotBound Constant

// Float64Bound is a type expression that has all float64s as elements.
var Float64Bound Constant

// NameBound is a type expression that has all names as elements.
var NameBound Constant

// StringBound is a type expression that has all strings as elements.
var StringBound Constant

// BytesBound is a type expression that has all bytestrings as elements.
var BytesBound Constant

// NumberBound is a type expression that has all numbers as elements.
var NumberBound Constant

// TimeBound is a type expression that has all time instants as elements.
var TimeBound Constant

// DurationBound is a type expression that has all durations as elements.
var DurationBound Constant

// TruePredicate is a predicate symbol to represent an
// "unconditionally true" proposition.
var TruePredicate = PredicateSym{"true", 0}

// FalsePredicate is a predicate symbol to represent an
// "unconditionally false" proposition.
var FalsePredicate = PredicateSym{"false", 0}

// TrueConstant is the "/true" name constant.
var TrueConstant Constant

// FalseConstant is the "/false" name constant.
var FalseConstant Constant

func init() {
	AnyBound, _ = Name("/any")
	BotBound, _ = Name("/bot")
	Float64Bound, _ = Name("/float64")
	NameBound, _ = Name("/name")
	NumberBound, _ = Name("/number")
	StringBound, _ = Name("/string")
	BytesBound, _ = Name("/bytes")
	TimeBound, _ = Name("/time")
	DurationBound, _ = Name("/duration")
	TrueConstant, _ = Name("/true")
	FalseConstant, _ = Name("/false")
}

// FormatNumber turns a number constant into a string.
func FormatNumber(num int64) string {
	return fmt.Sprintf("%d", num)
}

// FormatFloat64 turns a float64 constant into a string.
func FormatFloat64(floatNum float64) string {
	return strconv.FormatFloat(floatNum, 'f', -1, 64)
}

// FormatTime formats a time instant (nanoseconds since Unix epoch) as an ISO 8601 string.
func FormatTime(nanos int64) string {
	t := time.Unix(0, nanos).UTC()
	return t.Format(time.RFC3339Nano)
}

// FormatDuration formats a duration (nanoseconds) as a human-readable string.
// Uses the most appropriate unit: ns, us, ms, s, m, h, or combinations.
func FormatDuration(nanos int64) string {
	d := time.Duration(nanos)
	// Use Go's standard duration formatting for precision
	return d.String()
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

// BaseTerm represents a subset of terms: constant, variables or ApplyFn.
// Every BaseTerm will implement Term.
type BaseTerm interface {
	Term

	// Marker method.
	isBaseTerm()

	Hash() uint64

	// Returns a new base term.
	ApplySubstBase(s Subst) BaseTerm
}

// Subst is the interface for substitutions.
// This interface provides mapping from a variable to BaseTerm.
type Subst interface {
	// Returns the term the given variable maps to, or nil if the variable is not in domain.
	Get(Variable) BaseTerm
}

// SubstMap is a substitution backed by a map from variables to constants.
type SubstMap map[Variable]BaseTerm

// Get implements the Get method from Subst.
func (m SubstMap) Get(v Variable) BaseTerm {
	return m[v]
}

// ConstSubstMap is a substitution backed by a map from variables to constants.
type ConstSubstMap map[Variable]Constant

// Get implements the Get method from Subst.
func (m ConstSubstMap) Get(v Variable) BaseTerm {
	return m[v]
}

// Domain returns the domain of this substitution.
func (m ConstSubstMap) Domain() []Variable {
	var domain []Variable
	for v := range m {
		domain = append(domain, v)
	}
	return domain
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

// Domain returns a slice of variables that form the domain of this substitution.
func (c ConstSubstList) Domain() []Variable {
	var domain []Variable
	for _, x := range c {
		domain = append(domain, x.v)
	}
	return domain
}

// GetRow turns this substitution into a tuple.
func (c ConstSubstList) GetRow(domain []Variable) []Constant {
	result := make([]Constant, len(domain))
	for i, x := range domain {
		result[i] = c.Get(x).(Constant)
	}
	return result
}

// ConstantType describes the primitive type or shape of a constant.
type ConstantType int

const (
	// NameType is the type of name constants.
	NameType ConstantType = iota
	// StringType is the type of string constants.
	StringType
	// BytesType is the type of byte strings.
	BytesType
	// NumberType is the type of number (int64) constants.
	NumberType
	// Float64Type is the type of float64 constants.
	Float64Type
	// TimeType is the type of time instant constants (nanoseconds since Unix epoch UTC).
	TimeType
	// DurationType is the type of duration constants (nanoseconds).
	DurationType
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

// Constant represents a constant symbol and other structures (e.g. pair, list map, struct).
type Constant struct {
	// The (runtime) type of this constant.
	Type ConstantType

	// If Type \in {StringType, BytesType}, the data itself.
	// Otherwise, a string representation.
	Symbol string

	// For NumberType, the number value (int64 or the bytes of a float64).
	// For other types it contains a hash code of the value.
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
	return Constant{NameType, symbol, int64(hashBytes([]byte(symbol))), nil, nil}, nil
}

// String constructs a string constant.
func String(str string) Constant {
	return Constant{StringType, str, int64(hashBytes([]byte(str))), nil, nil}
}

// Bytes constructs a byte string constant.
func Bytes(bytes []byte) Constant {
	return Constant{BytesType, string(bytes), int64(hashBytes(bytes)), nil, nil}
}

// Number constructs a constant symbol that contains a number.
func Number(num int64) Constant {
	return Constant{NumberType, "", num, nil, nil}
}

// Float64 constructs a constant symbol that contains a float64.
func Float64(floatNum float64) Constant {
	return Constant{Float64Type, "", int64(math.Float64bits(floatNum)), nil, nil}
}

// Time constructs a time instant constant from nanoseconds since Unix epoch (1970-01-01 00:00:00 UTC).
// The valid range is approximately 1678 to 2262 CE.
func Time(nanos int64) Constant {
	return Constant{TimeType, "", nanos, nil, nil}
}

// Duration constructs a duration constant from nanoseconds.
// Positive values represent forward durations, negative values represent backward durations.
func Duration(nanos int64) Constant {
	return Constant{DurationType, "", nanos, nil, nil}
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

// Map constructs a map constant. Parts can only be accessed in transforms.
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

// NameValue returns the name value of this constant, if it is of type name.
func (c Constant) NameValue() (string, error) {
	if c.Type != NameType {
		return "", fmt.Errorf("not a name constant %v", c)
	}
	return c.Symbol, nil
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
		return 0, fmt.Errorf("not a float64 constant %v", c)
	}
	return math.Float64frombits(uint64(c.NumValue)), nil
}

// TimeValue returns the time value (nanoseconds since Unix epoch) of this constant.
func (c Constant) TimeValue() (int64, error) {
	if c.Type != TimeType {
		return 0, fmt.Errorf("not a time constant %v", c)
	}
	return c.NumValue, nil
}

// DurationValue returns the duration value (nanoseconds) of this constant.
func (c Constant) DurationValue() (int64, error) {
	if c.Type != DurationType {
		return 0, fmt.Errorf("not a duration constant %v", c)
	}
	return c.NumValue, nil
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

// ListSeq returns an iterator over the list elements.
func (c Constant) ListSeq() (iter.Seq[Constant], error) {
	if c.Type != ListShape {
		return nil, fmt.Errorf("not a list constant %v", c)
	}
	return func(yield func(Constant) bool) {
		for ; !c.IsListNil(); c = *c.snd {
			if ok := yield(*c.fst); !ok {
				return
			}
		}
	}, nil
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
		return string(c.Symbol)
	case StringType:
		s, err := Escape(c.Symbol, false /* isBytes */)
		if err != nil {
			return "<bad>"
		}
		return fmt.Sprintf(`"%s"`, s)
	case BytesType:
		s, err := Escape(c.Symbol, true /* isBytes */)
		if err != nil {
			return "<bad>"
		}
		return fmt.Sprintf(`b"%s"`, s)
	case NumberType:
		return FormatNumber(c.NumValue)
	case Float64Type:
		return FormatFloat64(math.Float64frombits(uint64(c.NumValue)))
	case TimeType:
		return FormatTime(c.NumValue)
	case DurationType:
		return FormatDuration(c.NumValue)
	case PairShape:
		fst := *c.fst
		snd := *c.snd
		return fmt.Sprintf("fn:pair(%s, %s)", fst.String(), snd.String())
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

// DisplayString returns a string representation of the constant without escaping Unicode characters.
// Note: If the string contains quote characters ("), the display string won't look like a source string anymore.
func (c Constant) DisplayString() string {
	switch c.Type {
	case NameType:
		return string(c.Symbol)
	case StringType:
		return fmt.Sprintf("\"%s\"", c.Symbol)
	case BytesType:
		return fmt.Sprintf("b\"%s\"", c.Symbol)
	case NumberType:
		return FormatNumber(c.NumValue)
	case Float64Type:
		return FormatFloat64(math.Float64frombits(uint64(c.NumValue)))
	case TimeType:
		return FormatTime(c.NumValue)
	case DurationType:
		return FormatDuration(c.NumValue)
	case PairShape:
		fst := *c.fst
		snd := *c.snd
		return fmt.Sprintf("fn:pair(%s, %s)", fst.DisplayString(), snd.DisplayString())
	case ListShape:
		if c.IsListNil() {
			return "[]"
		}
		var s strings.Builder
		s.WriteRune('[')
		s.WriteString((*c.fst).DisplayString())
		c2 := *c.snd
		for !c2.IsListNil() {
			s.WriteString(", ")
			s.WriteString((*c2.fst).DisplayString())
			c2 = *c2.snd
		}
		s.WriteRune(']')
		return s.String()
	case MapShape:
		if c.IsMapNil() {
			return "fn:map()"
		}
		var s strings.Builder
		s.WriteRune('[')
		s.WriteString((*c.fst.fst).DisplayString())
		s.WriteString(" : ")
		s.WriteString((*c.fst.snd).DisplayString())
		c2 := *c.snd
		for !c2.IsMapNil() {
			s.WriteString(", ")
			s.WriteString((*c2.fst.fst).DisplayString())
			s.WriteString(" : ")
			s.WriteString((*c2.fst.snd).DisplayString())
			c2 = *c2.snd
		}
		s.WriteRune(']')
		return s.String()
	case StructShape:
		if c.IsStructNil() {
			return "{}"
		}
		var s strings.Builder
		s.WriteRune('{')
		s.WriteString((*c.fst.fst).DisplayString())
		s.WriteString(" : ")
		s.WriteString((*c.fst.snd).DisplayString())
		c2 := *c.snd
		for !c2.IsStructNil() {
			s.WriteString(", ")
			s.WriteString((*c2.fst.fst).DisplayString())
			s.WriteString(" : ")
			s.WriteString((*c2.fst.snd).DisplayString())
			c2 = *c2.snd
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
	} else if v, ok := u.(Constant); ok {
		uconst = v
	} else {
		return false
	}
	if c.Type != uconst.Type || c.NumValue != uconst.NumValue {
		return false
	}
	// At this point, we know that constants have the same hash.
	switch c.Type {
	case NameType:
		fallthrough
	case StringType:
		return c.Symbol == uconst.Symbol
	case BytesType:
		return c.Symbol == uconst.Symbol
	case NumberType:
		fallthrough
	case Float64Type:
		fallthrough
	case TimeType:
		fallthrough
	case DurationType:
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
	left = left << tpe
	if snd == nil {
		return int64(left)
	}
	right := snd.Hash()
	return int64(szudzikElegantPair(left, right))
}

// Implements Szudzik's elegant pairing function (http://szudzik.com/ElegantPairing.pdf).
func szudzikElegantPair(fst, snd uint64) uint64 {
	if fst >= snd {
		return fst*fst + fst + snd
	}
	return snd*snd + fst
}

// HashConstants hashes a slice of constants.
func HashConstants(constants []Constant) uint64 {
	if len(constants) == 0 {
		return 0
	}
	h := constants[0].Hash()
	for _, snd := range constants[1:] {
		h = szudzikElegantPair(h, snd.Hash())
	}
	return h
}

// EqualsConstants compares two slices of constants.
func EqualsConstants(left []Constant, right []Constant) bool {
	if len(left) != len(right) {
		return false
	}
	for i := 0; i < len(left); i++ {
		if !left[i].Equals(right[i]) {
			return false
		}
	}
	return true
}

// Hash returns a hash code for this constant
func (c Constant) Hash() uint64 {
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
	var sb strings.Builder
	sb.WriteString(p.Symbol)
	sb.WriteRune('(')
	for i := 0; i < p.Arity; i++ {
		if i > 0 {
			sb.WriteString(", ")
		}
		fmt.Fprintf(&sb, "A%d", i)
	}
	sb.WriteRune(')')
	return sb.String()
}

// IsBuiltin returns true if this predicate symbol is for a built-in predicate.
func (p PredicateSym) IsBuiltin() bool {
	return strings.HasPrefix(p.Symbol, ":")
}

// Variable represents a variable by the name.
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

// ApplySubst returns the result of applying the given substitution.
func (v Variable) ApplySubst(s Subst) Term {
	return v.ApplySubstBase(s)
}

// ApplySubstBase returns the result of applying the given substitution.
func (v Variable) ApplySubstBase(s Subst) BaseTerm {
	if s == nil {
		return v
	}
	if t := s.Get(v); t != nil {
		return t
	}
	return v
}

// Atom represents an atom (a predicate symbol applied to base term arguments). e.g: parent(A, B)
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

// DisplayString returns a string representation for this atom using unescaped constants.
func (a Atom) DisplayString() string {
	var sb strings.Builder
	sb.WriteString(a.Predicate.Symbol)
	sb.WriteString("(")
	for i, arg := range a.Args {
		if i > 0 {
			sb.WriteString(",")
		}
		// Use DisplayString for Constant, fallback to String otherwise
		if c, ok := arg.(Constant); ok {
			sb.WriteString(c.DisplayString())
		} else {
			sb.WriteString(arg.String())
		}
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
	return hashTerm(a.Predicate.Symbol, a.Args)
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
	var sb strings.Builder
	sb.WriteString(f.Symbol)
	sb.WriteRune('(')
	for i := 0; i < f.Arity; i++ {
		if i > 0 {
			sb.WriteString(", ")
		}
		fmt.Fprintf(&sb, "V%d", i)
	}
	sb.WriteRune(')')
	return sb.String()
}

// Transform represents a transformation of the relation of a clause. e.g. fn:max or fn:group_by
type Transform struct {
	Statements []TransformStmt
	Next       *Transform
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
// Clauses may optionally have a temporal annotation on the head.
type Clause struct {
	Head      Atom
	HeadTime  *Interval // Optional temporal annotation (nil means eternal/timeless)
	Premises  []Term
	Transform *Transform
}

func (c Clause) String() string {
	headStr := c.Head.String()
	if c.HeadTime != nil && !c.HeadTime.IsEternal() {
		headStr += c.HeadTime.String()
	}
	if c.Premises == nil {
		return fmt.Sprintf("%s.", headStr)
	}
	var premises strings.Builder
	for i, p := range c.Premises {
		if i > 0 {
			premises.WriteString(", ")
		}
		premises.WriteString(p.String())
	}
	if c.Transform == nil {
		return fmt.Sprintf("%s :- %s.", headStr, premises.String())
	}
	return fmt.Sprintf("%s :- %s |> %s.", headStr, premises.String(), c.Transform.String())
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
	return Clause{c.Head, c.HeadTime, newPremises, c.Transform}
}

// NewClause constructs a new clause.
func NewClause(head Atom, premises []Term) Clause {
	return Clause{head, nil, premises, nil}
}

// NewTemporalClause constructs a new clause with a temporal annotation.
func NewTemporalClause(head Atom, headTime *Interval, premises []Term) Clause {
	return Clause{head, headTime, premises, nil}
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
		used[v] = true
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
