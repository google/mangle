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
	"strings"

	"github.com/google/mangle/ast"
	"github.com/google/mangle/symbols"
	"github.com/google/mangle/unionfind"
)

var (
	// Predicates built-in predicates.
	Predicates = map[ast.PredicateSym]struct{}{
		symbols.Lt:             struct{}{},
		symbols.Le:             struct{}{},
		symbols.WithinDistance: struct{}{},
		symbols.MatchPair:      struct{}{},
		symbols.MatchCons:      struct{}{},
		symbols.MatchNil:       struct{}{},
	}

	// Functions has all built-in functions except reducers.
	Functions = map[ast.FunctionSym]struct{}{
		symbols.Div:   struct{}{},
		symbols.Mult:  struct{}{},
		symbols.Plus:  struct{}{},
		symbols.Minus: struct{}{},

		// This is only used to start a "do-transform".
		symbols.GroupBy: struct{}{},

		symbols.ListGet: struct{}{},
		symbols.Append:  struct{}{},
		symbols.Cons:    struct{}{},
		symbols.Len:     struct{}{},
		symbols.List:    struct{}{},
		symbols.Pair:    struct{}{},
		symbols.Tuple:   struct{}{},
	}

	// ReducerFunctions has those built-in functions with are reducers.
	ReducerFunctions = map[ast.FunctionSym]struct{}{
		symbols.Collect:         struct{}{},
		symbols.CollectDistinct: struct{}{},
		symbols.PickAny:         struct{}{},
		symbols.Max:             struct{}{},
		symbols.Sum:             struct{}{},
		symbols.Count:           struct{}{},
	}

	// ErrDivisionByZero indicates a division by zero runtime error.
	ErrDivisionByZero = errors.New("div by zero")
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
func Decide(atom ast.Atom, subst *unionfind.UnionFind) (bool, *unionfind.UnionFind, error) {
	switch atom.Predicate.Symbol {
	case symbols.MatchPair.Symbol:
		fallthrough
	case symbols.MatchCons.Symbol:
		fallthrough
	case symbols.MatchNil.Symbol:
		return match(atom, subst)
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
		return nums[0] < nums[1], subst, nil
	case symbols.Le.Symbol:
		if len(atom.Args) != 2 {
			return false, nil, fmt.Errorf("wrong number of arguments for built-in predicate '<=': %v", atom.Args)
		}
		nums, err := getNumberValues(atom.Args)
		if err != nil {
			return false, nil, err
		}
		return nums[0] <= nums[1], subst, nil
	case symbols.WithinDistance.Symbol:
		if len(atom.Args) != 3 {
			return false, nil, fmt.Errorf("wrong number of arguments for built-in predicate 'within_distance': %v", atom.Args)
		}
		nums, err := getNumberValues(atom.Args)
		if err != nil {
			return false, nil, err
		}
		return abs(nums[0]-nums[1]) < nums[2], subst, nil
	default:
		return false, nil, fmt.Errorf("not a builtin predicate: %s", atom.Predicate.Symbol)
	}
}

func match(atom ast.Atom, subst *unionfind.UnionFind) (bool, *unionfind.UnionFind, error) {
	evaluatedArg, err := EvalExpr(atom.Args[0], subst)
	if err != nil {
		return false, nil, err
	}
	scrutinee, ok := evaluatedArg.(ast.Constant)
	if !ok {
		return false, nil, fmt.Errorf("not a constant: %v %T", evaluatedArg, evaluatedArg)
	}
	switch atom.Predicate.Symbol {
	case symbols.MatchPair.Symbol:
		if len(atom.Args) != 3 {
			return false, nil, fmt.Errorf("wrong number of arguments for built-in predicate ':match_pair': %v", atom.Args)
		}
		leftVar, leftOK := atom.Args[1].(ast.Variable)
		rightVar, rightOk := atom.Args[2].(ast.Variable)
		if !leftOK || !rightOk {
			return false, nil, fmt.Errorf("2nd and 3rd arguments must be variables for ':match_pair': %v", atom)
		}

		fst, snd, err := scrutinee.PairValue()
		if err != nil {
			return false, nil, nil // failing match is not an error
		}
		// First argument is indeed a pair. Bind.
		nsubst, err := unionfind.UnifyTermsExtend([]ast.BaseTerm{leftVar, rightVar}, []ast.BaseTerm{fst, snd}, *subst)
		if err != nil {
			return false, nil, fmt.Errorf("This should never happen for %v", atom)
		}
		return true, &nsubst, nil

	case symbols.MatchCons.Symbol:
		if len(atom.Args) != 3 {
			return false, nil, fmt.Errorf("wrong number of arguments for built-in predicate ':match_cons': %v", atom.Args)
		}
		leftVar, leftOK := atom.Args[1].(ast.Variable)
		rightVar, rightOk := atom.Args[2].(ast.Variable)
		if !leftOK || !rightOk {
			return false, nil, fmt.Errorf("2nd and 3rd arguments must be variables for ':match_cons': %v", atom)
		}

		scrutineeList, err := getListValue(scrutinee)
		if err != nil {
			return false, nil, err
		}
		hd, tail, err := scrutineeList.ConsValue()
		if err != nil {
			return false, nil, nil // failing match is not an error
		}
		// First argument is indeed a cons. Bind.
		nsubst, err := unionfind.UnifyTermsExtend([]ast.BaseTerm{leftVar, rightVar}, []ast.BaseTerm{hd, tail}, *subst)
		if err != nil {
			return false, nil, fmt.Errorf("This should never happen for %v", atom)
		}
		return true, &nsubst, nil

	case symbols.MatchNil.Symbol:
		if len(atom.Args) != 1 {
			return false, nil, fmt.Errorf("wrong number of arguments for built-in predicate ':match_nil': %v", atom.Args)
		}
		if !scrutinee.IsListNil() {
			return false, nil, nil
		}
		return true, subst, nil

	default:
		return false, nil, fmt.Errorf("unexpected case: %v", atom.Predicate.Symbol)
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
		var errBreak = errors.New("break")
		var res *ast.Constant
		_, loopErr := arg.ListValues(func(c ast.Constant) error {
			if i == int(index) {
				res = &c
				return errBreak
			}
			i++
			return nil
		}, func() error {
			return nil
		})
		if errors.Is(loopErr, errBreak) {
			return *res, nil
		}
		return ast.Constant{}, fmt.Errorf("index out of bounds: %d", index)

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

func getListValue(c ast.Constant) (ast.Constant, error) {
	if c.Type != ast.ListShape {
		return ast.Constant{}, fmt.Errorf("value %v (%v) is not a list", c, c.Type)
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
		var max int64
		for _, subst := range rows {
			num, err := getNumberValue(subst.Get(v).(ast.Constant))
			if err != nil {
				continue
			}
			if num > max {
				max = num
			}
		}
		return ast.Number(max), nil
	case symbols.Min.Symbol:
		v := reduceFn.Args[0].(ast.Variable)
		var min int64
		for _, subst := range rows {
			num, err := getNumberValue(subst.Get(v).(ast.Constant))
			if err != nil {
				continue
			}
			if num < min {
				min = num
			}
		}
		return ast.Number(min), nil
	case symbols.Sum.Symbol:
		v := reduceFn.Args[0].(ast.Variable)
		var sum int64
		for _, subst := range rows {
			num, err := getNumberValue(subst.Get(v).(ast.Constant))
			if err != nil {
				continue
			}
			sum += num
		}
		return ast.Number(sum), nil
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
