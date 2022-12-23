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
	"math"
	"strings"

	"github.com/google/mangle/ast"
	"github.com/google/mangle/symbols"
	"github.com/google/mangle/unionfind"
)

var (
	// Predicates built-in predicates.
	Predicates = map[ast.PredicateSym]struct{}{
		symbols.Lt:             {},
		symbols.Le:             {},
		symbols.ListMember:     {},
		symbols.WithinDistance: {},
		symbols.MatchPair:      {},
		symbols.MatchCons:      {},
		symbols.MatchNil:       {},
		symbols.MatchField:     {},
		symbols.MatchEntry:     {},
	}

	// Functions has all built-in functions except reducers.
	Functions = map[ast.FunctionSym]struct{}{
		symbols.Div:   {},
		symbols.Mult:  {},
		symbols.Plus:  {},
		symbols.Minus: {},

		// This is only used to start a "do-transform".
		symbols.GroupBy: {},

		symbols.ListGet: {},
		symbols.Append:  {},
		symbols.Cons:    {},
		symbols.Len:     {},
		symbols.List:    {},
		symbols.Pair:    {},
		symbols.Tuple:   {},
	}

	// ReducerFunctions has those built-in functions with are reducers.
	ReducerFunctions = map[ast.FunctionSym]struct{}{
		symbols.Collect:         {},
		symbols.CollectDistinct: {},
		symbols.PickAny:         {},
		symbols.Max:             {},
		symbols.Sum:             {},
		symbols.Count:           {},
	}

	// ErrDivisionByZero indicates a division by zero runtime error.
	ErrDivisionByZero = errors.New("div by zero")

	// errFound is used for exiting a loop
	errFound = errors.New("found")
)

func init() {
	for fn := range ReducerFunctions {
		Functions[fn] = struct{}{}
	}
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
	case symbols.Lt.Symbol:
		if len(atom.Args) != 2 {
			return false, nil, fmt.Errorf("wrong number of arguments for built-in predicate '<': %v", atom.Args)
		}
		nums, err := getNumberValues(atom.Args)
		if err != nil {
			return false, nil, err
		}
		return nums[0] < nums[1], []*unionfind.UnionFind{subst}, nil
	case symbols.Le.Symbol:
		if len(atom.Args) != 2 {
			return false, nil, fmt.Errorf("wrong number of arguments for built-in predicate '<=': %v", atom.Args)
		}
		nums, err := getNumberValues(atom.Args)
		if err != nil {
			return false, nil, err
		}
		return nums[0] <= nums[1], []*unionfind.UnionFind{subst}, nil

	case symbols.ListMember.Symbol: // :list:member(Member, List)
		evaluatedArg, err := EvalExpr(atom.Args[1], subst)
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
			res, err := EvalExpr(
				ast.ApplyFn{symbols.ListContains, []ast.BaseTerm{evaluatedArg, evaluatedMember}}, nil)
			if err != nil {
				return false, nil, err
			}
			return res.Equals(ast.TrueConstant), []*unionfind.UnionFind{subst}, nil
		}
		list, err := getListValue(c)
		if err != nil {
			return false, nil, nil // If expanding fails, this is not an error.
		}
		var values []ast.Constant
		list.ListValues(func(elem ast.Constant) error {
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
		nums, err := getNumberValues(atom.Args)
		if err != nil {
			return false, nil, err
		}
		return abs(nums[0]-nums[1]) < nums[2], []*unionfind.UnionFind{subst}, nil
	default:
		return false, nil, fmt.Errorf("not a builtin predicate: %s", atom.Predicate.Symbol)
	}
}

func match(pattern ast.Atom, subst *unionfind.UnionFind) (bool, *unionfind.UnionFind, error) {
	evaluatedArg, err := EvalExpr(pattern.Args[0], subst)
	if err != nil {
		return false, nil, err
	}
	scrutinee, ok := evaluatedArg.(ast.Constant)
	if !ok {
		return false, nil, fmt.Errorf("not a constant: %v %T", evaluatedArg, evaluatedArg)
	}
	switch pattern.Predicate.Symbol {
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

// EvalApplyFn evaluates a built-in function application.
func EvalApplyFn(applyFn ast.ApplyFn, subst ast.Subst) (ast.Constant, error) {
	evaluatedArgs, err := EvalExprs(applyFn.Args, subst)
	if err != nil {
		return ast.Constant{}, err
	}
	switch applyFn.Function.Symbol {
	case symbols.Append.Symbol:
		list, err := getListValue(evaluatedArgs[0])
		if err != nil {
			return ast.Constant{}, fmt.Errorf("expected list got %v (%v) in %v (subst %v)", evaluatedArgs[0], evaluatedArgs[0].Type, applyFn, subst)
		}
		elem := evaluatedArgs[1]
		var res []ast.Constant
		list.ListValues(func(c ast.Constant) error {
			res = append(res, c)
			return nil
		}, func() error { return nil })
		res = append(res, elem)
		return ast.List(res), nil

	case symbols.ListContains.Symbol: // fn:list:contains(List, Member)
		list, err := getListValue(evaluatedArgs[0])
		if err != nil {
			return ast.Constant{}, err
		}
		elem := evaluatedArgs[1]
		_, loopErr := list.ListValues(func(c ast.Constant) error {
			if c.Equals(elem) {
				return errFound
			}
			return nil
		}, func() error { return nil })
		if errors.Is(loopErr, errFound) {
			return ast.TrueConstant, nil
		}
		if loopErr != nil {
			return ast.Constant{}, loopErr
		}
		return ast.FalseConstant, nil

	case symbols.Cons.Symbol:
		fst := evaluatedArgs[0]
		snd, err := getListValue(evaluatedArgs[1])
		if err != nil {
			return ast.Constant{}, err
		}
		return ast.ListCons(&fst, &snd), nil

	case symbols.Pair.Symbol:
		fst := evaluatedArgs[0]
		snd := evaluatedArgs[1]
		return ast.Pair(&fst, &snd), nil

	case symbols.Len.Symbol:
		list := evaluatedArgs[0]
		var length int64
		err, _ := list.ListValues(func(c ast.Constant) error {
			length++
			return nil
		}, func() error { return nil })
		if err != nil {
			return ast.Constant{}, err
		}
		return ast.Number(length), nil

	case symbols.List.Symbol:
		list := &ast.ListNil
		for i := len(evaluatedArgs) - 1; i >= 0; i-- {
			elem := evaluatedArgs[i]
			next := ast.ListCons(&elem, list)
			list = &next
		}
		return *list, nil

	case symbols.Map.Symbol:
		kvMap := make(map[*ast.Constant]*ast.Constant)
		for i := 0; i < len(evaluatedArgs); i++ {
			label := &evaluatedArgs[i]
			i++
			value := &evaluatedArgs[i]
			kvMap[label] = value
		}
		return *ast.Map(kvMap), nil

	case symbols.Struct.Symbol:
		kvMap := make(map[*ast.Constant]*ast.Constant)
		for i := 0; i < len(evaluatedArgs); i++ {
			label := &evaluatedArgs[i]
			i++
			value := &evaluatedArgs[i]
			kvMap[label] = value
		}
		return *ast.Struct(kvMap), nil

	case symbols.Tuple.Symbol:
		if len(evaluatedArgs) == 1 {
			return evaluatedArgs[0], nil
		}
		width := len(evaluatedArgs)
		pair, err := EvalApplyFn(ast.ApplyFn{symbols.Pair, applyFn.Args[width-2:]}, subst)
		if err != nil {
			return ast.Constant{}, fmt.Errorf("could not eval args %v (subst %v)", applyFn, subst)
		}
		tuple := &pair
		for i := width - 3; i >= 0; i-- {
			arg := evaluatedArgs[i]
			next := ast.Pair(&arg, tuple)
			tuple = &next
		}
		return *tuple, nil

	case symbols.Max.Symbol:
		list, err := getListValue(evaluatedArgs[0])
		if err != nil {
			return ast.Constant{}, err
		}
		return evalMax(func(cbNext func(ast.Constant) error, cbNil func() error) error {
			if _, err = list.ListValues(cbNext, cbNil); err != nil {
				return err
			}
			return nil
		})

	case symbols.FloatMax.Symbol:
		list, err := getListValue(evaluatedArgs[0])
		if err != nil {
			return ast.Constant{}, err
		}
		return evalFloatMax(func(cbNext func(ast.Constant) error, cbNil func() error) error {
			if _, err = list.ListValues(cbNext, cbNil); err != nil {
				return err
			}
			return nil
		})

	case symbols.Min.Symbol:
		list, err := getListValue(evaluatedArgs[0])
		if err != nil {
			return ast.Constant{}, err
		}
		return evalMin(func(cbNext func(ast.Constant) error, cbNil func() error) error {
			if _, err = list.ListValues(cbNext, cbNil); err != nil {
				return err
			}
			return nil
		})

	case symbols.FloatMin.Symbol:
		list, err := getListValue(evaluatedArgs[0])
		if err != nil {
			return ast.Constant{}, err
		}
		return evalFloatMin(func(cbNext func(ast.Constant) error, cbNil func() error) error {
			if _, err = list.ListValues(cbNext, cbNil); err != nil {
				return err
			}
			return nil
		})

	case symbols.Sum.Symbol:
		list, err := getListValue(evaluatedArgs[0])
		if err != nil {
			return ast.Constant{}, err
		}
		return evalSum(func(cbNext func(ast.Constant) error, cbNil func() error) error {
			if _, err = list.ListValues(cbNext, cbNil); err != nil {
				return err
			}
			return nil
		})

	case symbols.FloatSum.Symbol:
		list, err := getListValue(evaluatedArgs[0])
		if err != nil {
			return ast.Constant{}, err
		}
		return evalFloatSum(func(cbNext func(ast.Constant) error, cbNil func() error) error {
			if _, err = list.ListValues(cbNext, cbNil); err != nil {
				return err
			}
			return nil
		})

	case symbols.ListGet.Symbol:
		arg, err := getListValue(evaluatedArgs[0])
		if err != nil {
			return ast.Constant{}, err
		}
		indexConstant := evaluatedArgs[1]
		index, err := getNumberValue(indexConstant)
		if err != nil {
			return ast.Constant{}, err
		}
		i := 0
		var res *ast.Constant
		_, loopErr := arg.ListValues(func(c ast.Constant) error {
			if i == int(index) {
				res = &c
				return errFound
			}
			i++
			return nil
		}, func() error {
			return nil
		})
		if errors.Is(loopErr, errFound) {
			return *res, nil
		}
		return ast.Constant{}, fmt.Errorf("index out of bounds: %d", index)

	case symbols.MapGet.Symbol:
		arg, err := getMapValue(evaluatedArgs[0])
		if err != nil {
			return ast.Constant{}, err
		}
		lookupKey := evaluatedArgs[1]
		var res *ast.Constant
		_, loopErr := arg.MapValues(func(key ast.Constant, val ast.Constant) error {
			if key.Equals(lookupKey) {
				res = &val
				return errFound
			}
			return nil
		}, func() error {
			return nil
		})
		if errors.Is(loopErr, errFound) {
			return *res, nil
		}
		return ast.Constant{}, fmt.Errorf("key does not exist: %v", lookupKey)

	case symbols.StructGet.Symbol:
		arg, err := getStructValue(evaluatedArgs[0])
		if err != nil {
			return ast.Constant{}, err
		}
		lookupField := evaluatedArgs[1]
		var res *ast.Constant
		_, loopErr := arg.StructValues(func(field ast.Constant, val ast.Constant) error {
			if field.Equals(lookupField) {
				res = &val
				return errFound
			}
			return nil
		}, func() error {
			return nil
		})
		if errors.Is(loopErr, errFound) {
			return *res, nil
		}
		return ast.Constant{}, fmt.Errorf("key does not exist: %v", lookupField)

	default:
		return EvalNumericApplyFn(applyFn, subst)
	}
}

// EvalNumericApplyFn evaluates a numeric built-in function application.
func EvalNumericApplyFn(applyFn ast.ApplyFn, subst ast.Subst) (ast.Constant, error) {
	evaluatedArgs, err := EvalExprs(applyFn.Args, subst)
	if err != nil {
		return ast.Constant{}, err
	}
	args, err := getNumberValues(evaluatedArgs)
	if err != nil {
		return ast.Constant{}, err
	}
	switch applyFn.Function.Symbol {
	case symbols.Div.Symbol:
		res, err := evalDiv(args)
		if err != nil {
			return ast.Constant{}, err
		}
		return ast.Number(res), nil
	case symbols.Mult.Symbol:
		return ast.Number(evalMult(args)), nil
	case symbols.Plus.Symbol:
		return ast.Number(evalPlus(args)), nil
	case symbols.Minus.Symbol:
		return ast.Number(evalMinus(args)), nil
	default:
		return ast.Constant{}, fmt.Errorf("unknown function %s in %s", applyFn, applyFn.Function)
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

// Abs returns the absolute value of x.
func abs(x int64) int64 {
	if x < 0 {
		return -x // This is wrong for math.MinInt
	}
	return x
}

func evalDiv(args []int64) (int64, error) {
	if len(args) == 1 {
		switch args[0] {
		case 0:
			return 0, ErrDivisionByZero
		case 1:
			return 1, nil
		default:
			return 0, nil // integer division 1 / arg[0]
		}
	}
	res := args[0]
	for _, divisor := range args[1:] {
		if divisor == 0 {
			return 0, ErrDivisionByZero
		}
		res = res / divisor
		if res == 0 {
			return 0, nil
		}
	}
	return res, nil
}

func evalMult(args []int64) int64 {
	var product int64 = 1
	for _, factor := range args {
		product = product * factor
	}
	return product
}

func evalPlus(args []int64) int64 {
	var sum int64 = 0
	for _, num := range args {
		sum += num
	}
	return sum
}

func evalMinus(args []int64) int64 {
	if len(args) == 1 {
		return -args[0]
	}
	var diff int64 = args[0]
	for _, num := range args[1:] {
		diff -= num
	}
	return diff
}

// EvalReduceFn evaluates a combiner (reduce) function.
func EvalReduceFn(reduceFn ast.ApplyFn, rows []ast.ConstSubstList) (ast.Constant, error) {
	distinct := false
	switch reduceFn.Function.Symbol {
	case symbols.CollectDistinct.Symbol:
		distinct = true
		fallthrough
	case symbols.Collect.Symbol:
		tuples := &ast.ListNil
		seen := make(map[uint64][]ast.Constant)
	rowloop:
		for i := len(rows) - 1; i >= 0; i-- {
			subst := rows[i]
			tuple := make([]ast.BaseTerm, 0, len(reduceFn.Args))
			for _, v := range reduceFn.Args {
				v, err := EvalExpr(v, subst)
				if err != nil {
					continue rowloop
				}
				constant, ok := v.(ast.Constant)
				if !ok {
					continue rowloop
				}
				tuple = append(tuple, constant)
			}
			head, err := EvalApplyFn(ast.ApplyFn{symbols.Tuple, tuple}, subst)
			if err != nil {
				continue rowloop
			}
			if !distinct {
				next := ast.ListCons(&head, tuples)
				tuples = &next
				continue
			}

			previousConstants, ok := seen[head.Hash()]
			if ok {
				for _, prev := range previousConstants {
					if prev.Equals(head) {
						continue rowloop
					}
				}
				seen[head.Hash()] = append(seen[head.Hash()], head)
			} else {
				seen[head.Hash()] = []ast.Constant{head}
			}
			next := ast.ListCons(&head, tuples)
			tuples = &next
		}
		return *tuples, nil
	case symbols.Count.Symbol:
		return ast.Number(int64(len(rows))), nil
	case symbols.Max.Symbol:
		v := reduceFn.Args[0].(ast.Variable)
		return evalMax(func(cbNext func(ast.Constant) error, cbNil func() error) error {
			for _, subst := range rows {
				if num, ok := subst.Get(v).(ast.Constant); ok {
					if err := cbNext(num); err != nil {
						return err
					}
				}
			}
			if err := cbNil(); err != nil {
				return err
			}
			return nil
		})
	case symbols.FloatMax.Symbol:
		v := reduceFn.Args[0].(ast.Variable)
		return evalFloatMax(func(cbNext func(ast.Constant) error, cbNil func() error) error {
			for _, subst := range rows {
				if num, ok := subst.Get(v).(ast.Constant); ok {
					if err := cbNext(num); err != nil {
						return err
					}
				}
			}
			return cbNil()
		})
	case symbols.Min.Symbol:
		v := reduceFn.Args[0].(ast.Variable)
		return evalMin(func(cbNext func(ast.Constant) error, cbNil func() error) error {
			for _, subst := range rows {
				if num, ok := subst.Get(v).(ast.Constant); ok {
					if err := cbNext(num); err != nil {
						return err
					}
				}
			}
			return cbNil()
		})
	case symbols.FloatMin.Symbol:
		v := reduceFn.Args[0].(ast.Variable)
		return evalFloatMin(func(cbNext func(ast.Constant) error, cbNil func() error) error {
			for _, subst := range rows {
				if num, ok := subst.Get(v).(ast.Constant); ok {
					if err := cbNext(num); err != nil {
						return err
					}
				}
			}
			return cbNil()
		})
	case symbols.Sum.Symbol:
		v := reduceFn.Args[0].(ast.Variable)
		return evalSum(func(cbNext func(ast.Constant) error, cbNil func() error) error {
			for _, subst := range rows {
				if num, ok := subst.Get(v).(ast.Constant); ok {
					if err := cbNext(num); err != nil {
						return err
					}
				}
			}
			return cbNil()
		})
	case symbols.FloatSum.Symbol:
		v := reduceFn.Args[0].(ast.Variable)
		return evalFloatSum(func(cbNext func(ast.Constant) error, cbNil func() error) error {
			for _, subst := range rows {
				if num, ok := subst.Get(v).(ast.Constant); ok {
					if err := cbNext(num); err != nil {
						return err
					}
				}
			}
			return cbNil()
		})
	default:
		return ast.Constant{}, fmt.Errorf("unknown reducer %v", reduceFn.Function)
	}
}

// EvalAtom returns an atom with any apply-expressions evaluated.
func EvalAtom(a ast.Atom, subst ast.Subst) (ast.Atom, error) {
	args, err := EvalExprsBase(a.Args, subst)
	if err != nil {
		return ast.Atom{}, err
	}
	return ast.Atom{a.Predicate, args}, nil
}

// EvalBaseTermPair evaluates a pair of base terms.
func EvalBaseTermPair(left, right ast.BaseTerm, subst ast.Subst) (ast.BaseTerm, ast.BaseTerm, error) {
	var err error
	left, err = EvalExpr(left, subst)
	if err != nil {
		return nil, nil, err
	}
	right, err = EvalExpr(right, subst)
	if err != nil {
		return nil, nil, err
	}
	return left, right, nil
}

// EvalExpr evaluates any apply-expression in b and applies subst.
func EvalExpr(b ast.BaseTerm, subst ast.Subst) (ast.BaseTerm, error) {
	expr, ok := b.(ast.ApplyFn)
	if !ok {
		return b.ApplySubstBase(subst), nil
	}
	return EvalApplyFn(expr, subst)
}

// EvalExprsBase evaluates any apply-expressions in args and applies subst.
func EvalExprsBase(args []ast.BaseTerm, subst ast.Subst) ([]ast.BaseTerm, error) {
	res := make([]ast.BaseTerm, len(args))
	for i, expr := range args {
		r, err := EvalExpr(expr, subst)
		if err != nil {
			return nil, err
		}
		res[i] = r
	}
	return res, nil
}

// EvalExprs evaluates any apply-expressions in args and applies subst.
func EvalExprs(args []ast.BaseTerm, subst ast.Subst) ([]ast.Constant, error) {
	res := make([]ast.Constant, len(args))
	for i, expr := range args {
		r, err := EvalExpr(expr, subst)
		if err != nil {
			return nil, err
		}
		c, ok := r.(ast.Constant)
		if !ok {
			return nil, fmt.Errorf("evaluation produced something that is not a value: %v %T", r, r)
		}
		res[i] = c
	}
	return res, nil
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
		t, err := symbols.NewTypeHandle(bound)
		if err != nil {
			return err
		}
		if !t.HasType(c) {
			return fmt.Errorf("argument %v is not an element of |%v|", c, t)
		}
	}
	return nil
}

func evalMax(iter func(func(ast.Constant) error, func() error) error) (ast.Constant, error) {
	max := int64(math.MinInt64)
	if err := iter(func(c ast.Constant) error {
		num, err := getNumberValue(c)
		if err != nil {
			return err
		}
		if num > max {
			max = num
		}
		return nil
	}, func() error { return nil }); err != nil {
		return ast.Constant{}, err
	}
	return ast.Number(max), nil
}

func evalFloatMax(iter func(func(ast.Constant) error, func() error) error) (ast.Constant, error) {
	max := -1 * math.MaxFloat64
	if err := iter(func(c ast.Constant) error {
		num, err := getFloatValue(c)
		if err != nil {
			return err
		}
		if num > max {
			max = num
		}
		return nil
	}, func() error { return nil }); err != nil {
		return ast.Constant{}, err
	}
	return ast.Float64(max), nil
}

func evalMin(iter func(func(ast.Constant) error, func() error) error) (ast.Constant, error) {
	min := int64(math.MaxInt64)
	if err := iter(func(c ast.Constant) error {
		num, err := getNumberValue(c)
		if err != nil {
			return err
		}
		if num < min {
			min = num
		}
		return nil
	}, func() error { return nil }); err != nil {
		return ast.Constant{}, err
	}
	return ast.Number(min), nil
}

func evalFloatMin(iter func(func(ast.Constant) error, func() error) error) (ast.Constant, error) {
	min := math.MaxFloat64
	if err := iter(func(c ast.Constant) error {
		floatNum, err := getFloatValue(c)
		if err != nil {
			return err
		}
		if floatNum < min {
			min = floatNum
		}
		return nil
	}, func() error { return nil }); err != nil {
		return ast.Constant{}, err
	}
	return ast.Float64(min), nil
}

func evalSum(iter func(func(ast.Constant) error, func() error) error) (ast.Constant, error) {
	var sum int64
	if err := iter(func(c ast.Constant) error {
		num, err := getNumberValue(c)
		if err != nil {
			return err
		}
		sum += num
		return nil
	}, func() error { return nil }); err != nil {
		return ast.Constant{}, err
	}
	return ast.Number(sum), nil
}

func evalFloatSum(iter func(func(ast.Constant) error, func() error) error) (ast.Constant, error) {
	var sum float64
	if err := iter(func(c ast.Constant) error {
		num, err := getFloatValue(c)
		if err != nil {
			return err
		}
		sum += num
		return nil
	}, func() error { return nil }); err != nil {
		return ast.Constant{}, err
	}
	return ast.Float64(sum), nil
}
