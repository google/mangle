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

// Package functional provides evaluation of function expressions.
package functional

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/google/mangle/ast"
	"github.com/google/mangle/symbols"
)

var (
	// ErrDivisionByZero indicates a division by zero runtime error.
	ErrDivisionByZero = errors.New("div by zero")

	// errFound is used for exiting a loop
	errFound = errors.New("found")
)

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

// EvalApplyFn evaluates a built-in function application.
func EvalApplyFn(applyFn ast.ApplyFn, subst ast.Subst) (ast.Constant, error) {
	evaluatedArgs, err := EvalExprs(applyFn.Args, subst)
	if err != nil {
		return ast.Constant{}, err
	}
	switch applyFn.Function.Symbol {
	case symbols.Append.Symbol:
		list := evaluatedArgs[0]
		if list.Type != ast.ListShape {
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
		list := evaluatedArgs[0]
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
		snd := evaluatedArgs[1]
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
		list := evaluatedArgs[0]
		return evalMax(func(cbNext func(ast.Constant) error, cbNil func() error) error {
			if _, err = list.ListValues(cbNext, cbNil); err != nil {
				return err
			}
			return nil
		})

	case symbols.FloatMax.Symbol:
		list := evaluatedArgs[0]
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
		list := evaluatedArgs[0]
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
		list := evaluatedArgs[0]
		return evalFloatMin(func(cbNext func(ast.Constant) error, cbNil func() error) error {
			if _, err = list.ListValues(cbNext, cbNil); err != nil {
				return err
			}
			return nil
		})

	case symbols.NumberToString.Symbol:
		val := evaluatedArgs[0]
		if val.Type != ast.NumberType {
			return ast.Constant{}, fmt.Errorf("cannot convert to string: fn:number:to_string() only converts ast.NumberType type")
		}

		return ast.String(ast.FormatNumber(val.NumValue)), nil

	case symbols.Float64ToString.Symbol:
		val := evaluatedArgs[0]
		if val.Type != ast.Float64Type {
			return ast.Constant{}, fmt.Errorf("cannot convert to string: fn:float64:to_string() only converts ast.Float64Type type")
		}

		f, err := val.Float64Value()
		if err != nil {
			return ast.Constant{}, err
		}

		return ast.String(strconv.FormatFloat(f, 'f', -1, 64)), nil

	case symbols.NameToString.Symbol:
		val := evaluatedArgs[0]
		if val.Type != ast.NameType {
			return ast.Constant{}, fmt.Errorf("cannot convert to string: fn:name:to_string() only converts ast.NameType type")
		}

		return ast.String(val.Symbol), nil

	case symbols.StringConcatenate.Symbol:
		var values []string
		for i, val := range evaluatedArgs {
			if val.Type != ast.StringType {
				return ast.Constant{}, fmt.Errorf("cannot string concatenate: value at position %v cannot be converted to string", i)
			}
			values = append(values, val.Symbol)
		}
		return ast.String(strings.Join(values, "")), nil

	case symbols.Sum.Symbol:
		list := evaluatedArgs[0]
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
		list := evaluatedArgs[0]
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
		arg := evaluatedArgs[0]
		if err != nil {
			return ast.Constant{}, err
		}
		indexConstant := evaluatedArgs[1]
		index, err := indexConstant.NumberValue()
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
		arg := evaluatedArgs[0] // map value
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
		arg := evaluatedArgs[0] // struct value
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
	args, err := EvalExprs(applyFn.Args, subst)
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

func evalDiv(args []ast.Constant) (int64, error) {
	if len(args) == 1 {
		switch args[0].NumValue {
		case 0:
			return 0, ErrDivisionByZero
		case 1:
			return 1, nil
		default:
			return 0, nil // integer division 1 / arg[0]
		}
	}
	res := args[0].NumValue
	for _, divisorConst := range args[1:] {
		divisor := divisorConst.NumValue
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

func evalMult(args []ast.Constant) int64 {
	var product int64 = 1
	for _, factor := range args {
		product = product * factor.NumValue
	}
	return product
}

func evalPlus(args []ast.Constant) int64 {
	var sum int64 = 0
	for _, num := range args {
		sum += num.NumValue
	}
	return sum
}

func evalMinus(args []ast.Constant) int64 {
	diff := args[0].NumValue
	if len(args) == 1 {
		return -diff // "unary minus"
	}
	for _, num := range args[1:] {
		diff -= num.NumValue
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

func evalMax(iter func(func(ast.Constant) error, func() error) error) (ast.Constant, error) {
	max := int64(math.MinInt64)
	if err := iter(func(c ast.Constant) error {
		num, err := c.NumberValue()
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
		num, err := c.Float64Value()
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
		num, err := c.NumberValue()
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
		floatNum, err := c.Float64Value()
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
		num, err := c.NumberValue()
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
		num, err := c.Float64Value()
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
