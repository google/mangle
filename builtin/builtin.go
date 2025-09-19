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

// Package builtin contains functions for evaluating built-in predicates.
package builtin

import (
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/google/mangle/ast"
	"github.com/google/mangle/functional"
	"github.com/google/mangle/symbols"
	"github.com/google/mangle/unionfind"
)

var (
	// Predicates has all built-in predicates.
	Predicates = map[ast.PredicateSym]ast.Mode{
		symbols.MatchPrefix:    {ast.ArgModeInput, ast.ArgModeInput},
		symbols.StartsWith:     {ast.ArgModeInput, ast.ArgModeInput},
		symbols.EndsWith:       {ast.ArgModeInput, ast.ArgModeInput},
		symbols.Contains:       {ast.ArgModeInput, ast.ArgModeInput},
		symbols.Filter:         {ast.ArgModeInput},
		symbols.Lt:             {ast.ArgModeInput, ast.ArgModeInput},
		symbols.Le:             {ast.ArgModeInput, ast.ArgModeInput},
		symbols.Gt:             {ast.ArgModeInput, ast.ArgModeInput},
		symbols.Ge:             {ast.ArgModeInput, ast.ArgModeInput},
		symbols.ListMember:     {ast.ArgModeOutput, ast.ArgModeInput},
		symbols.WithinDistance: {ast.ArgModeInput, ast.ArgModeInput, ast.ArgModeInput},
		symbols.MatchPair:      {ast.ArgModeInput, ast.ArgModeOutput, ast.ArgModeOutput},
		symbols.MatchCons:      {ast.ArgModeInput, ast.ArgModeOutput, ast.ArgModeOutput},
		symbols.MatchNil:       {ast.ArgModeInput},
		symbols.MatchField:     {ast.ArgModeInput, ast.ArgModeInput, ast.ArgModeOutput},
		symbols.MatchEntry:     {ast.ArgModeInput, ast.ArgModeInput, ast.ArgModeOutput},
	}

	varX                  = ast.Variable{"X"}
	varY                  = ast.Variable{"Y"}
	listOfX               = symbols.NewListType(varX)
	listOfNum             = symbols.NewListType(ast.NumberBound)
	listOfFloats          = symbols.NewListType(ast.Float64Bound)
	numberOrDecimal       = symbols.NewUnionType(ast.NumberBound, ast.DecimalBound)
	numberDecimalOrFloat  = symbols.NewUnionType(ast.NumberBound, ast.DecimalBound, ast.Float64Bound)
	listOfNumberOrDecimal = symbols.NewListType(numberOrDecimal)
	listOfNumOrFloats     = symbols.NewListType(numberDecimalOrFloat)
	emptyType             = symbols.NewUnionType()

	// Functions has all built-in functions.
	Functions = map[ast.FunctionSym]ast.BaseTerm{
		symbols.Div:       emptyType,
		symbols.FloatDiv:  emptyType,
		symbols.FloatMult: emptyType,
		symbols.FloatPlus: emptyType,
		symbols.Mult:      emptyType,
		symbols.Plus:      emptyType,
		symbols.Minus:     emptyType,
		symbols.Sqrt:      emptyType,

		// This is only used to start a "do-transform".
		symbols.GroupBy: emptyType,

		symbols.ListGet:      symbols.NewFunType(symbols.NewOptionType(varX) /* <= */, listOfX, ast.NumberBound),
		symbols.ListContains: symbols.NewFunType(symbols.BoolType() /* <= */, listOfX, varX),
		symbols.Append:       symbols.NewFunType(listOfX /* <= */, listOfX, varX),
		symbols.Cons:         symbols.NewFunType(listOfX /* <= */, varX, listOfX),
		symbols.Len:          symbols.NewFunType(ast.NumberBound /* <= */, listOfX),
		symbols.Pair:         symbols.NewFunType(symbols.NewPairType(varX, varY) /* <= */, varX, varY),
		symbols.Some:         symbols.NewFunType(symbols.NewOptionType(varX) /* <= */, varX),
		symbols.StringConcatenate: symbols.NewFunType(
			ast.StringBound /* <= */, ast.AnyBound),
		symbols.StringReplace: symbols.NewFunType(
			ast.StringBound /* <= */, ast.StringBound),
		symbols.StructGet:          symbols.NewFunType(ast.AnyBound /* <= */, ast.AnyBound, ast.NameBound),
		symbols.NumberToString:     symbols.NewFunType(ast.StringBound /* <= */, ast.NumberBound),
		symbols.Float64ToString:    symbols.NewFunType(ast.StringBound /* <= */, ast.Float64Bound),
		symbols.DecimalFromString:  symbols.NewFunType(ast.DecimalBound /* <= */, ast.StringBound),
		symbols.DecimalFromNumber:  symbols.NewFunType(ast.DecimalBound /* <= */, ast.NumberBound),
		symbols.DecimalFromFloat64: symbols.NewFunType(ast.DecimalBound /* <= */, ast.Float64Bound),
		symbols.DecimalToString:    symbols.NewFunType(ast.StringBound /* <= */, ast.DecimalBound),
		symbols.DecimalToNumber:    symbols.NewFunType(ast.NumberBound /* <= */, ast.DecimalBound),
		symbols.DecimalToFloat64:   symbols.NewFunType(ast.Float64Bound /* <= */, ast.DecimalBound),
		symbols.NameToString:       symbols.NewFunType(ast.StringBound /* <= */, ast.NameBound),
		symbols.DateFromString:     symbols.NewFunType(ast.DateBound /* <= */, ast.StringBound),
		symbols.DateToString:       symbols.NewFunType(ast.StringBound /* <= */, ast.DateBound),
		symbols.DateFromParts:      symbols.NewFunType(ast.DateBound /* <= */, ast.NumberBound, ast.NumberBound, ast.NumberBound),
		symbols.DateAddDays:        symbols.NewFunType(ast.DateBound /* <= */, ast.DateBound, ast.NumberBound),
		symbols.DateSubDays:        symbols.NewFunType(ast.DateBound /* <= */, ast.DateBound, ast.NumberBound),
		symbols.DateDiffDays:       symbols.NewFunType(ast.NumberBound /* <= */, ast.DateBound, ast.DateBound),
		symbols.NameRoot:           symbols.NewFunType(ast.NameBound /* <= */, ast.NameBound),
		symbols.NameTip:            symbols.NewFunType(ast.NameBound /* <= */, ast.NameBound),
		symbols.NameList:           symbols.NewFunType(symbols.NewListType(ast.NameBound) /* <= */, ast.NameBound),

		// These "functions" (constructors) need special handling due to varargs.
		symbols.List:   symbols.NewFunType(symbols.NewListType(varX) /* <= */, varX),
		symbols.Map:    emptyType,
		symbols.Tuple:  emptyType,
		symbols.Struct: emptyType,
	}

	// ReducerFunctions has those built-in functions with are reducers.
	ReducerFunctions = map[ast.FunctionSym]ast.BaseTerm{
		symbols.Collect:         symbols.NewFunType(listOfX /* <= */, listOfX),
		symbols.CollectDistinct: symbols.NewFunType(listOfX /* <= */, listOfX),
		symbols.PickAny:         symbols.NewFunType(varX /* <= */, listOfX),
		symbols.Max:             symbols.NewFunType(numberOrDecimal /* <= */, listOfNumberOrDecimal),
		symbols.Min:             symbols.NewFunType(numberOrDecimal /* <= */, listOfNumberOrDecimal),
		symbols.Sum:             symbols.NewFunType(numberOrDecimal /* <= */, listOfNumberOrDecimal),
		symbols.FloatMax:        symbols.NewFunType(ast.Float64Bound /* <= */, listOfFloats),
		symbols.FloatMin:        symbols.NewFunType(ast.Float64Bound /* <= */, listOfFloats),
		symbols.FloatSum:        symbols.NewFunType(ast.Float64Bound /* <= */, listOfNumOrFloats),
		symbols.Count:           symbols.NewFunType(ast.NumberBound /* <= */, listOfX),
		symbols.Avg:             symbols.NewFunType(ast.Float64Bound /* <= */, listOfFloats),
	}

	// errFound is used for exiting a loop
	errFound = errors.New("found")
)

func init() {
	for fn, tpe := range ReducerFunctions {
		Functions[fn] = tpe
	}
}

// GetBuiltinFunctionType returns the type of a builtin function.
// The type may contain type variables.
func GetBuiltinFunctionType(sym ast.FunctionSym) (ast.BaseTerm, bool) {
	if tpe, ok := Functions[sym]; ok {
		return tpe, true
	}
	if tpe, ok := Functions[ast.FunctionSym{sym.Symbol, -1}]; ok {
		return tpe, true // variable arity
	}
	return nil, false
}

// IsBuiltinFunction returns true if sym is a builtin function.
func IsBuiltinFunction(sym ast.FunctionSym) bool {
	if _, ok := Functions[sym]; ok {
		return true
	}
	if _, ok := Functions[ast.FunctionSym{sym.Symbol, -1}]; ok {
		return true // variable arity
	}
	return false
}

// IsReducerFunction returns true if sym is a reducer function.
func IsReducerFunction(sym ast.FunctionSym) bool {
	if _, ok := ReducerFunctions[sym]; ok {
		return true
	}
	if _, ok := ReducerFunctions[ast.FunctionSym{sym.Symbol, -1}]; ok {
		return true // variable arity
	}
	return false
}

// Decide evaluates an atom of a built-in predicate. The atom must no longer contain any
// apply-expressions or variables.
func Decide(atom ast.Atom, subst *unionfind.UnionFind) (bool, []*unionfind.UnionFind, error) {
	switch atom.Predicate.Symbol {
	case symbols.StartsWith.Symbol:
		fallthrough
	case symbols.EndsWith.Symbol:
		fallthrough
	case symbols.Contains.Symbol:
		fallthrough
	case symbols.MatchPrefix.Symbol:
		fallthrough
	case symbols.MatchPair.Symbol:
		fallthrough
	case symbols.MatchCons.Symbol:
		fallthrough
	case symbols.MatchEntry.Symbol:
		fallthrough
	case symbols.MatchField.Symbol:
		fallthrough
	case symbols.MatchNil.Symbol:
		ok, nsubst, err := match(atom, subst)
		if err != nil {
			return false, nil, err
		}
		if !ok {
			return false, nil, nil
		}
		return ok, []*unionfind.UnionFind{nsubst}, nil
	}
	switch atom.Predicate.Symbol {
	case symbols.Filter.Symbol:
		if len(atom.Args) != 1 {
			return false, nil, fmt.Errorf("wrong number of arguments for built-in predicate 'filter': %v", atom.Args)
		}
		evaluatedArg, err := functional.EvalExpr(atom.Args[0], subst)
		if err != nil {
			return false, nil, err
		}
		if evaluatedArg == ast.TrueConstant {
			return true, []*unionfind.UnionFind{subst}, nil
		}
		return false, nil, nil

	case symbols.Lt.Symbol:
		if len(atom.Args) != 2 {
			return false, nil, fmt.Errorf("wrong number of arguments for built-in predicate '<': %v", atom.Args)
		}
		result, err := compareOrderedArgs(atom.Args, orderRelationLess)
		if err != nil {
			return false, nil, err
		}
		return result, []*unionfind.UnionFind{subst}, nil
	case symbols.Le.Symbol:
		if len(atom.Args) != 2 {
			return false, nil, fmt.Errorf("wrong number of arguments for built-in predicate '<=': %v", atom.Args)
		}
		result, err := compareOrderedArgs(atom.Args, orderRelationLessEqual)
		if err != nil {
			return false, nil, err
		}
		return result, []*unionfind.UnionFind{subst}, nil

	case symbols.Gt.Symbol:
		if len(atom.Args) != 2 {
			return false, nil, fmt.Errorf("wrong number of arguments for built-in predicate '>': %v", atom.Args)
		}
		result, err := compareOrderedArgs(atom.Args, orderRelationGreater)
		if err != nil {
			return false, nil, err
		}
		return result, []*unionfind.UnionFind{subst}, nil
	case symbols.Ge.Symbol:
		if len(atom.Args) != 2 {
			return false, nil, fmt.Errorf("wrong number of arguments for built-in predicate '>=': %v", atom.Args)
		}
		result, err := compareOrderedArgs(atom.Args, orderRelationGreaterEqual)
		if err != nil {
			return false, nil, err
		}
		return result, []*unionfind.UnionFind{subst}, nil

	case symbols.ListMember.Symbol: // :list:member(Member, List)
		evaluatedArg, err := functional.EvalExpr(atom.Args[1], subst)
		if err != nil {
			return false, nil, err
		}
		c, ok := evaluatedArg.(ast.Constant)
		if !ok {
			return false, nil, fmt.Errorf("not a constant: %v %T", evaluatedArg, evaluatedArg)
		}
		evaluatedMember := atom.Args[0]
		memberVar, memberIsVar := evaluatedMember.(ast.Variable)
		if memberIsVar && subst != nil {
			evaluatedMember = subst.Get(memberVar)
			_, memberIsVar = evaluatedMember.(ast.Variable)
		}
		if !memberIsVar { // We are looking for a member
			res, err := functional.EvalExpr(
				ast.ApplyFn{symbols.ListContains, []ast.BaseTerm{evaluatedArg, evaluatedMember}}, nil)
			if err != nil {
				return false, nil, err
			}
			return res.Equals(ast.TrueConstant), []*unionfind.UnionFind{subst}, nil
		}
		if c.Type != ast.ListShape {
			return false, nil, nil // If expanding fails, this is not an error.
		}
		var values []ast.Constant
		c.ListValues(func(elem ast.Constant) error {
			values = append(values, elem)
			return nil
		}, func() error { return nil })
		if len(values) > 0 {
			var nsubsts []*unionfind.UnionFind
			for _, elem := range values {
				nsubst, err := unionfind.UnifyTermsExtend([]ast.BaseTerm{memberVar}, []ast.BaseTerm{elem}, *subst)
				if err != nil {
					return false, nil, err
				}
				nsubsts = append(nsubsts, &nsubst)
			}
			return true, nsubsts, nil
		}
		return false, nil, nil

	case symbols.WithinDistance.Symbol:
		if len(atom.Args) != 3 {
			return false, nil, fmt.Errorf("wrong number of arguments for built-in predicate 'within_distance': %v", atom.Args)
		}
		nums, err := getNumericRats(atom.Args)
		if err != nil {
			return false, nil, err
		}
		diff := new(big.Rat).Sub(nums[0], nums[1])
		return absRat(diff).Cmp(nums[2]) < 0, []*unionfind.UnionFind{subst}, nil
	default:
		return false, nil, fmt.Errorf("not a builtin predicate: %s", atom.Predicate.Symbol)
	}
}

func match(pattern ast.Atom, subst *unionfind.UnionFind) (bool, *unionfind.UnionFind, error) {
	evaluatedArg, err := functional.EvalExpr(pattern.Args[0], subst)
	if err != nil {
		return false, nil, err
	}
	scrutinee, ok := evaluatedArg.(ast.Constant)
	if !ok {
		return false, nil, fmt.Errorf("not a constant: %v %T", evaluatedArg, evaluatedArg)
	}
	switch pattern.Predicate.Symbol {
	case symbols.MatchPrefix.Symbol:
		if len(pattern.Args) != 2 {
			return false, nil, fmt.Errorf("wrong number of arguments for built-in predicate ':match_prefix': %v", pattern.Args)
		}
		pat, ok := pattern.Args[1].(ast.Constant)
		if !ok || pat.Type != ast.NameType {
			return false, nil, fmt.Errorf("2nd arguments must be name constant for ':match_prefix': %v", pattern)
		}
		name, ok := evaluatedArg.(ast.Constant)
		if !ok || name.Type != ast.NameType {
			return false, nil, nil
		}
		return strings.HasPrefix(name.Symbol, pat.Symbol) && len(name.Symbol) > len(pat.Symbol), subst, nil

	case symbols.StartsWith.Symbol:
		if len(pattern.Args) != 2 {
			return false, nil, fmt.Errorf("wrong number of arguments for built-in predicate ':string:starts_with': %v", pattern.Args)
		}
		pat, ok := pattern.Args[1].(ast.Constant)
		if !ok || pat.Type != ast.StringType {
			return false, nil, fmt.Errorf("2nd arguments must be string constant for ':string:starts_with': %v", pattern)
		}
		str, ok := evaluatedArg.(ast.Constant)
		if !ok || str.Type != ast.StringType {
			return false, nil, nil
		}
		return strings.HasPrefix(str.Symbol, pat.Symbol), subst, nil

	case symbols.EndsWith.Symbol:
		if len(pattern.Args) != 2 {
			return false, nil, fmt.Errorf("wrong number of arguments for built-in predicate ':string:ends_with': %v", pattern.Args)
		}
		pat, ok := pattern.Args[1].(ast.Constant)
		if !ok || pat.Type != ast.StringType {
			return false, nil, fmt.Errorf("2nd arguments must be string constant for ':string:ends_with': %v", pattern)
		}
		str, ok := evaluatedArg.(ast.Constant)
		if !ok || str.Type != ast.StringType {
			return false, nil, nil
		}
		return strings.HasSuffix(str.Symbol, pat.Symbol), subst, nil

	case symbols.Contains.Symbol:
		if len(pattern.Args) != 2 {
			return false, nil, fmt.Errorf("wrong number of arguments for built-in predicate ':string:contains': %v", pattern.Args)
		}
		pat, ok := pattern.Args[1].(ast.Constant)
		if !ok || pat.Type != ast.StringType {
			return false, nil, fmt.Errorf("2nd arguments must be string constant for ':string:contains': %v", pattern)
		}
		str, ok := evaluatedArg.(ast.Constant)
		if !ok || str.Type != ast.StringType {
			return false, nil, nil
		}
		return strings.Contains(str.Symbol, pat.Symbol), subst, nil

	case symbols.MatchPair.Symbol:
		if len(pattern.Args) != 3 {
			return false, nil, fmt.Errorf("wrong number of arguments for built-in predicate ':match_pair': %v", pattern.Args)
		}
		leftVar, leftOK := pattern.Args[1].(ast.Variable)
		rightVar, rightOk := pattern.Args[2].(ast.Variable)
		if !leftOK || !rightOk {
			return false, nil, fmt.Errorf("2nd and 3rd arguments must be variables for ':match_pair': %v", pattern)
		}

		fst, snd, err := scrutinee.PairValue()
		if err != nil {
			return false, nil, nil // failing match is not an error
		}
		// First argument is indeed a pair. Bind.
		nsubst, err := unionfind.UnifyTermsExtend([]ast.BaseTerm{leftVar, rightVar}, []ast.BaseTerm{fst, snd}, *subst)
		if err != nil {
			return false, nil, fmt.Errorf("This should never happen for %v", pattern)
		}
		return true, &nsubst, nil

	case symbols.MatchCons.Symbol:
		if len(pattern.Args) != 3 {
			return false, nil, fmt.Errorf("wrong number of arguments for built-in predicate ':match_cons': %v", pattern.Args)
		}
		leftVar, leftOK := pattern.Args[1].(ast.Variable)
		rightVar, rightOk := pattern.Args[2].(ast.Variable)
		if !leftOK || !rightOk {
			return false, nil, fmt.Errorf("2nd and 3rd arguments must be variables for ':match_cons': %v", pattern)
		}

		scrutineeList, err := getListValue(scrutinee)
		if err != nil {
			return false, nil, nil // failing match is not an error
		}
		hd, tail, err := scrutineeList.ConsValue()
		if err != nil {
			return false, nil, nil // failing match is not an error
		}
		// First argument is indeed a cons. Bind.
		nsubst, err := unionfind.UnifyTermsExtend([]ast.BaseTerm{leftVar, rightVar}, []ast.BaseTerm{hd, tail}, *subst)
		if err != nil {
			return false, nil, fmt.Errorf("This should never happen for %v", pattern)
		}
		return true, &nsubst, nil

	case symbols.MatchNil.Symbol:
		if len(pattern.Args) != 1 {
			return false, nil, fmt.Errorf("wrong number of arguments for built-in predicate ':match_nil': %v", pattern.Args)
		}
		if !scrutinee.IsListNil() {
			return false, nil, nil
		}
		return true, subst, nil

	case symbols.MatchEntry.Symbol:
		if scrutinee.Type != ast.MapShape || scrutinee.IsMapNil() {
			return false, nil, nil
		}
		patternKey, ok := pattern.Args[1].(ast.Constant)
		if !ok {
			return false, nil, fmt.Errorf("bad pattern %v", pattern) // This should not happen
		}
		patternVal := pattern.Args[2]
		var found *ast.Constant
		e, err := scrutinee.MapValues(func(key ast.Constant, val ast.Constant) error {
			if key.Equals(patternKey) {
				found = &val
				return errFound
			}
			return nil
		}, func() error { return nil })
		if e != nil {
			return false, nil, e // This should not happen
		}
		if errors.Is(err, errFound) {
			if nsubst, errUnify := unionfind.UnifyTermsExtend([]ast.BaseTerm{patternVal}, []ast.BaseTerm{*found}, *subst); errUnify == nil { // if NO error
				return true, &nsubst, nil
			}
		}
		return false, nil, nil

	case symbols.MatchField.Symbol:
		if scrutinee.Type != ast.StructShape || scrutinee.IsStructNil() {
			return false, nil, nil
		}
		patternKey, ok := pattern.Args[1].(ast.Constant)
		if !ok {
			return false, nil, fmt.Errorf("bad pattern %v", pattern) // This should not happen
		}
		patternVal := pattern.Args[2]
		var found *ast.Constant
		e, err := scrutinee.StructValues(func(key ast.Constant, val ast.Constant) error {
			if key.Equals(patternKey) {
				found = &val
				return errFound
			}
			return nil
		}, func() error { return nil })
		if e != nil {
			return false, nil, nil // This should not happen
		}

		if errors.Is(err, errFound) {
			if nsubst, errUnify := unionfind.UnifyTermsExtend([]ast.BaseTerm{patternVal}, []ast.BaseTerm{*found}, *subst); errUnify == nil { // if NO error
				return true, &nsubst, nil
			}
		}
		return false, nil, nil

	default:
		return false, nil, fmt.Errorf("unexpected case: %v", pattern.Predicate.Symbol)
	}
}

func getStringValue(baseTerm ast.BaseTerm) (string, error) {
	constant, ok := baseTerm.(ast.Constant)
	if !ok || constant.Type != ast.StringType {
		return "", fmt.Errorf("value %v (%T) is not a string", baseTerm, baseTerm)
	}
	return constant.StringValue()
}

func getNumberValue(b ast.BaseTerm) (int64, error) {
	c, ok := b.(ast.Constant)
	if !ok {
		return 0, fmt.Errorf("not a value %v (%T)", b, b)
	}
	if c.Type != ast.NumberType {
		return 0, fmt.Errorf("value %v (%v) is not a number", c, c.Type)
	}
	return c.NumberValue()
}

func getNumericRat(b ast.BaseTerm) (*big.Rat, error) {
	c, ok := b.(ast.Constant)
	if !ok {
		return nil, fmt.Errorf("not a value %v (%T)", b, b)
	}
	switch c.Type {
	case ast.NumberType:
		return big.NewRat(c.NumValue, 1), nil
	case ast.DecimalType:
		return c.DecimalValue()
	default:
		return nil, fmt.Errorf("value %v (%v) is not a decimal or number", c, c.Type)
	}
}

func getFloatValue(b ast.BaseTerm) (float64, error) {
	c, ok := b.(ast.Constant)
	if !ok {
		return 0, fmt.Errorf("not a value %v (%T)", b, b)
	}
	if c.Type != ast.Float64Type {
		return 0, fmt.Errorf("value %v (%v) is not a number", c, c.Type)
	}
	return c.Float64Value()
}

func getListValue(c ast.Constant) (ast.Constant, error) {
	if c.Type != ast.ListShape {
		return ast.Constant{}, fmt.Errorf("value %v (%v) is not a list", c, c.Type)
	}
	return c, nil
}

func getMapValue(c ast.Constant) (ast.Constant, error) {
	if c.Type != ast.MapShape {
		return ast.Constant{}, fmt.Errorf("value %v (%v) is not a map", c, c.Type)
	}
	return c, nil
}

func getStructValue(c ast.Constant) (ast.Constant, error) {
	if c.Type != ast.StructShape {
		return ast.Constant{}, fmt.Errorf("value %v (%v) is not a struct", c, c.Type)
	}
	return c, nil
}

type orderKind int

const (
	orderKindNumeric orderKind = iota
	orderKindDate
)

type orderRelation int

const (
	orderRelationLess orderRelation = iota
	orderRelationLessEqual
	orderRelationGreater
	orderRelationGreaterEqual
)

type orderValue struct {
	kind     orderKind
	numeric  *big.Rat
	date     time.Time
	constant ast.Constant
}

func getOrderValue(term ast.BaseTerm) (orderValue, error) {
	c, ok := term.(ast.Constant)
	if !ok {
		return orderValue{}, fmt.Errorf("not a value %v (%T)", term, term)
	}
	switch c.Type {
	case ast.NumberType:
		number, err := c.NumberValue()
		if err != nil {
			return orderValue{}, err
		}
		return orderValue{kind: orderKindNumeric, numeric: big.NewRat(number, 1), constant: c}, nil
	case ast.DecimalType:
		rat, err := c.DecimalValue()
		if err != nil {
			return orderValue{}, err
		}
		return orderValue{kind: orderKindNumeric, numeric: rat, constant: c}, nil
	case ast.DateType:
		t, err := c.DateValue()
		if err != nil {
			return orderValue{}, err
		}
		return orderValue{kind: orderKindDate, date: t, constant: c}, nil
	default:
		return orderValue{}, fmt.Errorf("value %v (%v) is not orderable", c, c.Type)
	}
}

func compareOrderedArgs(args []ast.BaseTerm, relation orderRelation) (bool, error) {
	left, err := getOrderValue(args[0])
	if err != nil {
		return false, err
	}
	right, err := getOrderValue(args[1])
	if err != nil {
		return false, err
	}
	return compareOrderValues(left, right, relation)
}

func compareOrderValues(left, right orderValue, relation orderRelation) (bool, error) {
	if left.kind != right.kind {
		return false, fmt.Errorf("cannot compare %v (%v) with %v (%v)", left.constant, left.constant.Type, right.constant, right.constant.Type)
	}
	switch left.kind {
	case orderKindNumeric:
		cmp := left.numeric.Cmp(right.numeric)
		switch relation {
		case orderRelationLess:
			return cmp < 0, nil
		case orderRelationLessEqual:
			return cmp <= 0, nil
		case orderRelationGreater:
			return cmp > 0, nil
		case orderRelationGreaterEqual:
			return cmp >= 0, nil
		}
	case orderKindDate:
		switch relation {
		case orderRelationLess:
			return left.date.Before(right.date), nil
		case orderRelationLessEqual:
			return !left.date.After(right.date), nil
		case orderRelationGreater:
			return left.date.After(right.date), nil
		case orderRelationGreaterEqual:
			return !left.date.Before(right.date), nil
		}
	}
	return false, fmt.Errorf("unsupported order relation %v", relation)
}

func getNumberValues[T ast.BaseTerm](cs []T) ([]int64, error) {
	var nums []int64
	for _, c := range cs {
		num, err := getNumberValue(c)
		if err != nil {
			return nil, err
		}
		nums = append(nums, num)
	}
	return nums, nil
}

func getNumericRats[T ast.BaseTerm](cs []T) ([]*big.Rat, error) {
	var nums []*big.Rat
	for _, c := range cs {
		num, err := getNumericRat(c)
		if err != nil {
			return nil, err
		}
		nums = append(nums, num)
	}
	return nums, nil
}

// Abs returns the absolute value of x.
func abs(x int64) int64 {
	if x < 0 {
		return -x // This is wrong for math.MinInt
	}
	return x
}

func absRat(r *big.Rat) *big.Rat {
	if r.Sign() >= 0 {
		return new(big.Rat).Set(r)
	}
	return new(big.Rat).Neg(r)
}

// TypeChecker checks the type of constant (run-time type).
type TypeChecker struct {
	decls map[ast.PredicateSym]*ast.Decl
}

// NewTypeChecker returns a new TypeChecker.
// The decls must be desugared so they only contain type bounds.
func NewTypeChecker(decls map[ast.PredicateSym]ast.Decl) (*TypeChecker, error) {
	desugaredDecls, err := symbols.CheckAndDesugar(decls)
	if err != nil {
		return nil, err
	}
	return NewTypeCheckerFromDesugared(desugaredDecls), nil
}

// NewTypeCheckerFromDesugared returns a new TypeChecker.
// The declarations must be in desugared form.
func NewTypeCheckerFromDesugared(decls map[ast.PredicateSym]*ast.Decl) *TypeChecker {
	return &TypeChecker{decls}
}

// CheckTypeBounds checks whether there fact is consistent with at least one of the bound decls.
func (t TypeChecker) CheckTypeBounds(fact ast.Atom) error {
	decl, ok := t.decls[fact.Predicate]
	if !ok {
		return fmt.Errorf("could not find declaration for %v", fact.Predicate)
	}
	var errs []string
	for _, boundDecl := range decl.Bounds {
		err := t.CheckOneBoundDecl(fact, boundDecl)
		if err == nil { // if NO error
			return nil
		}
		errs = append(errs, fmt.Sprintf("bound decl %v fails with %v", boundDecl, err))
	}
	return fmt.Errorf("fact %v matches none of the bound decls: %v", fact, strings.Join(errs, ","))
}

// CheckOneBoundDecl checks whether a fact is consistent with a given type bounds tuple.
func (t TypeChecker) CheckOneBoundDecl(fact ast.Atom, boundDecl ast.BoundDecl) error {
	for j, bound := range boundDecl.Bounds {
		c, ok := fact.Args[j].(ast.Constant)
		if !ok {
			return fmt.Errorf("fact %v could not check fact argument %v", fact, fact.Args[j])
		}
		t, err := symbols.NewBoundHandle(bound)
		if err != nil {
			return err
		}
		if !t.HasType(c) {
			return fmt.Errorf("argument %v is not an element of |%v|", c, t)
		}
	}
	return nil
}
