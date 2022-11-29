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

// Package symbols contains symbols for built-in functions and predicates.
package symbols

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/google/mangle/ast"
)

var (
	// Lt is the less-than relation on numbers.
	Lt = ast.PredicateSym{":lt", 2}

	// Le is the less-than-or-equal relation on numbers.
	Le = ast.PredicateSym{":le", 2}

	// MatchPair matches a pair to its elements.
	MatchPair = ast.PredicateSym{":match_pair", 3}

	// MatchCons matches a cons to head and tail.
	MatchCons = ast.PredicateSym{":match_cons", 3}

	// MatchNil matches the empty list.
	MatchNil = ast.PredicateSym{":match_nil", 1}

	// WithinDistance is a relation on numbers X, Y, Z satisfying |X - Y| < Z.
	WithinDistance = ast.PredicateSym{":within_distance", 3}

	// Div is a family of functions mapping X,Y1,.. to (X / Y1) / Y2 ... DIV(X) is 1/x.
	Div = ast.FunctionSym{"fn:div", -1}
	// Mult is a family of functions mapping X,Y1,.. to (X * Y1) * Y2 ... MULT(x) is x.
	Mult = ast.FunctionSym{"fn:mult", -1}
	// Plus is a family of functions mapping X,Y1,.. to (X + Y1) + Y2 ... PLUS(x) is x.
	Plus = ast.FunctionSym{"fn:plus", -1}
	// Minus is a family of functions mapping X,Y1,.. to (X - Y1) - Y2 ...MINUS(x) is -X.
	Minus = ast.FunctionSym{"fn:minus", -1}

	// Collect turns a collection { tuple_1,...tuple_n } into a list [tuple_1, ..., tuple_n].
	Collect = ast.FunctionSym{"fn:collect", -1}
	// CollectDistinct turns a collection { tuple_1,...tuple_n } into a list with distinct elements [tuple_1, ..., tuple_n].
	CollectDistinct = ast.FunctionSym{"fn:collect_distinct", -1}
	// PickAny reduces a set { x_1,...x_n } to a single { x_i },
	PickAny = ast.FunctionSym{"fn:pick_any", 1}
	// Max reduces a set { x_1,...x_n } to { x_i } that is maximal.
	Max = ast.FunctionSym{"fn:max", 1}
	// Min reduces a set { x_1,...x_n } to { x_i } that is minimal.
	Min = ast.FunctionSym{"fn:min", 1}
	// Sum reduces a set { x_1,...x_n } to { x_1 + ... + x_n }.
	Sum = ast.FunctionSym{"fn:sum", 1}
	// Count reduces a set { x_1,...x_n } to { n }.
	Count = ast.FunctionSym{"fn:count", 0}

	// GroupBy groups all tuples by the values of key variables, e.g. 'group_by(X)'.
	// An empty group_by() treats the whole relation as a group.
	GroupBy = ast.FunctionSym{"fn:group_by", -1}

	// Append appends a element to a list.
	Append = ast.FunctionSym{"fn:list:append", 2}
	// ListGet is a function (List, Number) which returns element at index 'Number'.
	ListGet = ast.FunctionSym{"fn:list:get", 2}
	// Len returns length of a list.
	Len = ast.FunctionSym{"fn:list:len", 1}
	// Cons constructs a pair.
	Cons = ast.FunctionSym{"fn:list:cons", 2}
	// Pair constructs a pair.
	Pair = ast.FunctionSym{"fn:pair", 2}
	// Tuple acts either as identity (one argument), pair (two arguments) or nested pair (more).
	Tuple = ast.FunctionSym{"fn:tuple", -1}
	// List constructs a list.
	List = ast.FunctionSym{"fn:list", -1}

	// PairType is a constructor for a pair type.
	PairType = ast.FunctionSym{"fn:Pair", 2}
	// TupleType is a type-level function that returns a tuple type out of pair types.
	TupleType = ast.FunctionSym{"fn:Tuple", -1}
	// ListType is a constructor for a list type.
	ListType = ast.FunctionSym{"fn:List", 1}
	// UnionType is a constructor for a union type.
	UnionType = ast.FunctionSym{"fn:Union", -1}

	// Package is an improper symbol, used to represent package declaration.
	Package = ast.PredicateSym{"Package", 0}
	// Use is an improper symbol, used to represent use declaration.
	Use = ast.PredicateSym{"Use", 0}

	// TypeConstructors is a list of function symbols used in structured type expressions.
	TypeConstructors = map[string]ast.FunctionSym{
		UnionType.Symbol: UnionType,
		ListType.Symbol:  ListType,
		PairType.Symbol:  PairType,
		TupleType.Symbol: TupleType,
	}

	// EmptyType is a type without members.
	EmptyType = ast.ApplyFn{UnionType, nil}

	argumentRange = map[ast.PredicateSym][]ast.BaseTerm{
		Lt:       []ast.BaseTerm{ast.NumberBound, ast.NumberBound},
		Le:       []ast.BaseTerm{ast.NumberBound, ast.NumberBound},
		MatchNil: []ast.BaseTerm{ast.ApplyFn{ListType, []ast.BaseTerm{ast.Number(0)}}},
		MatchCons: []ast.BaseTerm{
			ast.ApplyFn{ListType, []ast.BaseTerm{ast.Number(0)}},
			ast.Number(0),
			ast.ApplyFn{ListType, []ast.BaseTerm{ast.Number(0)}}},
		MatchPair: []ast.BaseTerm{
			ast.ApplyFn{PairType, []ast.BaseTerm{ast.Number(0), ast.Number(1)}},
			ast.Number(0),
			ast.Number(1)},
	}

	errTypeMismatch = errors.New("type mismatch")
)

// IsBaseTypeExpression returns true if c is a base type expression constants.
func IsBaseTypeExpression(c ast.Constant) bool {
	switch c {
	case ast.AnyBound:
		return true
	case ast.NumberBound:
		return true
	case ast.StringBound:
		return true
	default:
		return c.Type == ast.NameType
	}
}

// GetBuiltinArgumentRange returns argument range for a builtin predicate.
// The returning type expression may contain type variable (ast.Number).
func GetBuiltinArgumentRange(pred ast.PredicateSym) ([]ast.BaseTerm, error) {
	if res, ok := argumentRange[pred]; ok {
		return res, nil
	}
	return nil, fmt.Errorf("not a builtin predicate: %v", pred)
}

// TypeHandle provides functionality related to type expression.
type TypeHandle struct {
	expr ast.BaseTerm
}

// NewTypeHandle constructs a TypeHandle.
func NewTypeHandle(expr ast.BaseTerm) (TypeHandle, error) {
	if err := CheckTypeExpression(expr); err != nil {
		return TypeHandle{}, err
	}
	return TypeHandle{expr}, nil
}

// String returns a string represented of this type expression.
func (t TypeHandle) String() string {
	return t.expr.String()
}

// HasType returns true if c has type represented by this TypeHandle.
func (t TypeHandle) HasType(c ast.Constant) bool {
	if baseType, ok := t.expr.(ast.Constant); ok {
		return hasBaseType(baseType, c)
	}
	structuredType, ok := t.expr.(ast.ApplyFn)
	if !ok {
		return false // This never happens.
	}
	switch structuredType.Function {
	case PairType:
		fst, snd, err := c.PairValue()
		if err != nil {
			return false
		}
		return TypeHandle{structuredType.Args[0]}.HasType(fst) &&
			TypeHandle{structuredType.Args[1]}.HasType(snd)
	case ListType:
		elementType := TypeHandle{structuredType.Args[0]}
		shapeErr, err := c.ListValues(func(e ast.Constant) error {
			if !elementType.HasType(e) {
				return errTypeMismatch
			}
			return nil
		}, func() error {
			return nil
		})
		if shapeErr != nil {
			return false // not a list.
		}
		if errors.Is(err, errTypeMismatch) {
			return false
		}
		return true
	case TupleType:
		return TypeHandle{expandTupleType(structuredType.Args)}.HasType(c)
	case UnionType:
		for _, arg := range structuredType.Args {
			alt := TypeHandle{arg}
			if alt.HasType(c) {
				return true
			}
		}
		return false
	}
	return false
}

func hasBaseType(typeExpr ast.Constant, c ast.Constant) bool {
	switch typeExpr {
	case ast.AnyBound:
		return true
	case ast.NameBound:
		return c.Type == ast.NameType
	case ast.NumberBound:
		return c.Type == ast.NumberType
	case ast.StringBound:
		return c.Type == ast.StringType
	default:
		return typeExpr.Type == ast.NameType && c.Type == ast.NameType && strings.HasPrefix(c.Symbol, typeExpr.Symbol+"/")
	}
}

// CheckTypeExpression returns an error if expr is not a type expression.
func CheckTypeExpression(expr ast.BaseTerm) error {
	switch expr := expr.(type) {
	case ast.Constant:
		if !IsBaseTypeExpression(expr) {
			return fmt.Errorf("not a base type expression: %v", expr)
		}
		return nil
	case ast.Variable:
		return fmt.Errorf("not a type expression: %v", expr)
	case ast.ApplyFn:
		fn, ok := TypeConstructors[expr.Function.Symbol]
		if !ok {
			return fmt.Errorf("not a structured type expression: %v", expr)
		}
		args := expr.Args
		if fn.Arity != -1 && len(args) != fn.Arity {
			return fmt.Errorf("expected %d arguments in type expression %v ", fn.Arity, expr)
		}
		if fn == UnionType && len(args) <= 0 {
			return fmt.Errorf("union type must not be empty %v ", expr)
		}
		if fn == TupleType && len(args) <= 2 {
			return fmt.Errorf("tuple type must have more than 2 args %v ", expr)
		}
		for _, arg := range args {
			if err := CheckTypeExpression(arg); err != nil {
				return err
			}
		}
		return nil
	default:
		return fmt.Errorf("CheckTypeExpression: unexpected case %v %T", expr, expr)
	}
}

// TypeConforms returns true if left <: right.
func TypeConforms(left ast.BaseTerm, right ast.BaseTerm) bool {
	if left.Equals(right) || right.Equals(ast.AnyBound) {
		return true
	}
	if leftConst, ok := left.(ast.Constant); ok {
		if rightConst, ok := right.(ast.Constant); ok {
			if strings.HasPrefix(leftConst.Symbol, rightConst.Symbol) {
				return true
			}
			return leftConst.Type == ast.NameType && rightConst.Equals(ast.NameBound)
		}
	}
	if leftUnion, ok := left.(ast.ApplyFn); ok && leftUnion.Function.Symbol == UnionType.Symbol {
		for _, leftItem := range leftUnion.Args {
			if !TypeConforms(leftItem, right) {
				return false
			}
		}
		return true
	}
	if rightUnion, ok := right.(ast.ApplyFn); ok && rightUnion.Function.Symbol == UnionType.Symbol {
		for _, rightItem := range rightUnion.Args {
			if TypeConforms(left, rightItem) {
				return true
			}
		}
	}
	if leftList, ok := left.(ast.ApplyFn); ok && leftList.Function.Symbol == ListType.Symbol {
		if rightList, ok := right.(ast.ApplyFn); ok && rightList.Function.Symbol == ListType.Symbol {
			return TypeConforms(leftList.Args[0], rightList.Args[0])
		}
	}
	if leftPair, ok := left.(ast.ApplyFn); ok && leftPair.Function.Symbol == PairType.Symbol {
		if rightPair, ok := right.(ast.ApplyFn); ok && rightPair.Function.Symbol == PairType.Symbol {
			return TypeConforms(leftPair.Args[0], rightPair.Args[0]) && TypeConforms(leftPair.Args[1], rightPair.Args[1])
		}
	}
	if leftTuple, ok := left.(ast.ApplyFn); ok && leftTuple.Function.Symbol == TupleType.Symbol {
		if rightTuple, ok := right.(ast.ApplyFn); ok && rightTuple.Function.Symbol == TupleType.Symbol {
			for i, leftArg := range leftTuple.Args {
				if !TypeConforms(leftArg, rightTuple.Args[i]) {
					return false
				}
			}
			return true
		}
	}
	return false
}

func expandTupleType(args []ast.BaseTerm) ast.BaseTerm {
	res := ast.ApplyFn{PairType, []ast.BaseTerm{args[len(args)-2], args[len(args)-1]}}
	for j := len(args) - 3; j >= 0; j-- {
		res = ast.ApplyFn{PairType, []ast.BaseTerm{args[j], res}}
	}
	return res
}

// UpperBound returns upper bound of type expressions.
func UpperBound(typeExprs []ast.BaseTerm) ast.BaseTerm {
	var worklist []ast.BaseTerm
	for _, typeExpr := range typeExprs {
		if ast.AnyBound.Equals(typeExpr) {
			return ast.AnyBound
		}
		if union, ok := typeExpr.(ast.ApplyFn); ok && union.Function == UnionType {
			worklist = append(worklist, union.Args...)
			continue
		}
		worklist = append(worklist, typeExpr)
	}
	if len(worklist) == 0 {
		return ast.ApplyFn{UnionType, nil}
	}
	reduced := []ast.BaseTerm{worklist[0]}
	worklist = worklist[1:]
typeExprLoop:
	for _, typeExpr := range worklist {
		for i, existing := range reduced {
			if TypeConforms(typeExpr, existing) {
				continue typeExprLoop
			}
			if TypeConforms(existing, typeExpr) {
				reduced[i] = typeExpr
				continue typeExprLoop
			}
		}
		reduced = append(reduced, typeExpr)
	}
	if len(reduced) == 1 {
		return reduced[0]
	}
	sort.Slice(reduced, func(i, j int) bool { return reduced[i].Hash() < reduced[j].Hash() })
	return ast.ApplyFn{UnionType, reduced}
}

func intersectType(a, b ast.BaseTerm) ast.BaseTerm {
	if a.Equals(b) {
		return a
	}
	if a.Equals(ast.AnyBound) {
		return b
	}
	if b.Equals(ast.AnyBound) {
		return a
	}
	if TypeConforms(a, b) {
		return a
	}
	if TypeConforms(b, a) {
		return b
	}
	if aUnion, ok := a.(ast.ApplyFn); ok && aUnion.Function == UnionType {
		var res []ast.BaseTerm
		for _, elem := range aUnion.Args {
			if u := intersectType(elem, b); !u.Equals(EmptyType) {
				res = append(res, u)
			}
		}
		return UpperBound(res)
	}
	if bUnion, ok := b.(ast.ApplyFn); ok && bUnion.Function == UnionType {
		var res []ast.BaseTerm
		for _, elem := range bUnion.Args {
			if TypeConforms(a, elem) {
				res = append(res, a)
			} else if TypeConforms(elem, a) {
				res = append(res, elem)
			}
		}
		return UpperBound(res)
	}

	return EmptyType
}

// LowerBound returns a lower bound of type expressions.
func LowerBound(typeExprs []ast.BaseTerm) ast.BaseTerm {
	var typeExpr ast.BaseTerm = ast.AnyBound
	for _, t := range typeExprs {
		if typeExpr = intersectType(typeExpr, t); typeExpr.Equals(EmptyType) {
			return EmptyType
		}
	}
	return typeExpr
}
