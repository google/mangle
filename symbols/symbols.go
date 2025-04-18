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
	// MatchPrefix matches name constants that have a given prefix.
	MatchPrefix = ast.PredicateSym{":match_prefix", 2}

	// StartsWith matches string constants that have a given prefix.
	StartsWith = ast.PredicateSym{":string:starts_with", 2}

	// EndsWith matches string constants that have a given suffix.
	EndsWith = ast.PredicateSym{":string:ends_with", 2}

	// Contains matches string constants that contain the given string.
	Contains = ast.PredicateSym{":string:contains", 2}

	// Filter is turning a boolean function into a predicate.
	Filter = ast.PredicateSym{":filter", 1}

	// Lt is the less-than relation on numbers.
	Lt = ast.PredicateSym{":lt", 2}

	// Le is the less-than-or-equal relation on numbers.
	Le = ast.PredicateSym{":le", 2}

	// Gt is the greater-than relation on numbers.
	Gt = ast.PredicateSym{":gt", 2}

	// Ge is the greater-than-or-equal relation on numbers.
	Ge = ast.PredicateSym{":ge", 2}

	// MatchPair mode(+, -, -) matches a pair to its elements.
	MatchPair = ast.PredicateSym{":match_pair", 3}

	// MatchCons mode(+, -, -) matches a list to head and tail.
	MatchCons = ast.PredicateSym{":match_cons", 3}

	// MatchNil matches the empty list.
	MatchNil = ast.PredicateSym{":match_nil", 1}

	// MatchEntry mode(+, +, -) matches an entry in a map.
	MatchEntry = ast.PredicateSym{":match_entry", 3}

	// MatchField mode(+, +, -) matches a field in a struct.
	MatchField = ast.PredicateSym{":match_field", 3}

	// ListMember mode(+, -) either checks membership or binds var to every element.
	ListMember = ast.PredicateSym{":list:member", 2}

	// WithinDistance is a relation on numbers X, Y, Z satisfying |X - Y| < Z.
	WithinDistance = ast.PredicateSym{":within_distance", 3}

	// Div is a family of functions mapping integer division: X,Y1,.. to (X / Y1) / Y2 ... DIV(X) is 1/x.
	Div = ast.FunctionSym{"fn:div", -1}
	// FloatDiv is a family of functions mapping division: X,Y1,.. to (X / Y1) / Y2 ... FloatDiv(X) is 1/x.
	FloatDiv = ast.FunctionSym{"fn:float:div", -1}
	// FloatMult is a family of functions mapping multiplication: X,Y1,.. to (X * Y1) * Y2 ... FloatMult(x) is x.
	FloatMult = ast.FunctionSym{"fn:float:mult", -1}
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
	// FloatMax reduces a set of float64 { x_1,...x_n } to { x_i } that is maximal.
	FloatMax = ast.FunctionSym{"fn:float:max", 1}
	// Min reduces a set { x_1,...x_n } to { x_i } that is minimal.
	Min = ast.FunctionSym{"fn:min", 1}
	// FloatMin reduces a set of float64 { x_1,...x_n } to { x_i } that is minimal.
	FloatMin = ast.FunctionSym{"fn:float:min", 1}
	// Sum reduces a set { x_1,...x_n } to { x_1 + ... + x_n }.
	Sum = ast.FunctionSym{"fn:sum", 1}
	// FloatSum reduces a set of float64 { x_1,...x_n } to { x_1 + ... + x_n }.
	FloatSum = ast.FunctionSym{"fn:float:sum", 1}
	// Count reduces a set { x_1,...x_n } to { n }.
	Count = ast.FunctionSym{"fn:count", 0}
	// CountDistinct reduces a set { x_1,...x_n } to { m } with m being the
	// number of unique elements.
	CountDistinct = ast.FunctionSym{"fn:count_distinct", 0}
	// Avg reduces a set { x_1,...x_n } to { fn:sum(x_1,...,x_n) /  n }.
	Avg = ast.FunctionSym{"fn:avg", 1}

	// GroupBy groups all tuples by the values of key variables, e.g. 'group_by(X)'.
	// An empty group_by() treats the whole relation as a group.
	GroupBy = ast.FunctionSym{"fn:group_by", -1}

	// Append appends a element to a list.
	Append = ast.FunctionSym{"fn:list:append", 2}

	// ListGet is a function (List, Number) which returns element at index 'Number'.
	ListGet = ast.FunctionSym{"fn:list:get", 2}

	// ListContains is a function (List, Member) which returns /true if Member is contained in list.
	ListContains = ast.FunctionSym{"fn:list:contains", 2}

	// Len returns length of a list.
	Len = ast.FunctionSym{"fn:list:len", 1}
	// Cons constructs a pair.
	Cons = ast.FunctionSym{"fn:list:cons", 2}
	// Pair constructs a pair.
	Pair = ast.FunctionSym{"fn:pair", 2}
	// MapGet is a function (Map, Key) which returns element at key.
	MapGet = ast.FunctionSym{"fn:map:get", 2}
	// StructGet is a function (Struct, Field) which returns specified field.
	StructGet = ast.FunctionSym{"fn:struct:get", 2}
	// Tuple acts either as identity (one argument), pair (two arguments) or nested pair (more).
	Tuple = ast.FunctionSym{"fn:tuple", -1}
	// Some constructs an element of an option type.
	Some = ast.FunctionSym{"fn:some", 1}
	// List constructs a list.
	List = ast.FunctionSym{"fn:list", -1}
	// Map constructs a map.
	Map = ast.FunctionSym{"fn:map", -1}
	// Struct constructs a struct.
	Struct = ast.FunctionSym{"fn:struct", -1}

	// FunType is a constructor for a function type.
	// fn:Fun(Res, Arg1, ..., ArgN) is Res <= Arg1, ..., ArgN
	FunType = ast.FunctionSym{"fn:Fun", -1}
	// RelType is a constructor for a relation type.
	RelType = ast.FunctionSym{"fn:Rel", -1}
	// SingletonType is a constructor for a singleton type.
	SingletonType = ast.FunctionSym{"fn:Singleton", 1}
	// NumberToString converts from ast.NumberType to ast.StringType
	NumberToString = ast.FunctionSym{"fn:number:to_string", 1}
	// Float64ToString converts from ast.Float64Type to ast.StringType
	Float64ToString = ast.FunctionSym{"fn:float64:to_string", 1}
	// NameToString converts from ast.NameType to ast.StringType
	NameToString = ast.FunctionSym{"fn:name:to_string", 1}
	// NameRoot returns the first name part of a name.
	NameRoot = ast.FunctionSym{"fn:name:root", 1}
	// NameTip returns the last name part of a name.
	NameTip = ast.FunctionSym{"fn:name:tip", 1}
	// NameList turns a name into a list of name parts.
	NameList = ast.FunctionSym{"fn:name:list", 1}
	// StringConcatenate concatenates the arguments into a single string constant.
	StringConcatenate = ast.FunctionSym{"fn:string:concat", -1}
	// StringReplace replaces old with new in the first n occurrences of a string.
	StringReplace = ast.FunctionSym{"fn:string:replace", 4}
	// PairType is a constructor for a pair type.
	PairType = ast.FunctionSym{"fn:Pair", 2}
	// TupleType is a type-level function that returns a tuple type out of pair types.
	TupleType = ast.FunctionSym{"fn:Tuple", -1}
	// OptionType is a constructor for an option type.
	// A value of fn:Option(T) is either fn:some(c) for c:T, or fn:none().
	// TODO: Implement runtime representation.
	OptionType = ast.FunctionSym{"fn:Option", 1}
	// ListType is a constructor for a list type.
	ListType = ast.FunctionSym{"fn:List", 1}
	// MapType is a constructor for a map type.
	MapType = ast.FunctionSym{"fn:Map", 2}
	// StructType is a constructor for a struct type.
	StructType = ast.FunctionSym{"fn:Struct", -1}
	// UnionType is a constructor for a union type.
	UnionType = ast.FunctionSym{"fn:Union", -1}

	// Optional may appear inside StructType to indicate optional fields.
	Optional = ast.FunctionSym{"fn:opt", -1}

	// Package is an improper symbol, used to represent package declaration.
	Package = ast.PredicateSym{"Package", 0}
	// Use is an improper symbol, used to represent use declaration.
	Use = ast.PredicateSym{"Use", 0}

	// TypeConstructors is a list of function symbols used in structured type expressions.
	// Each name is mapped to the corresponding type constructor (a function at the level of types).
	TypeConstructors = map[string]ast.FunctionSym{
		UnionType.Symbol:     UnionType,
		SingletonType.Symbol: SingletonType,
		ListType.Symbol:      ListType,
		OptionType.Symbol:    OptionType,
		PairType.Symbol:      PairType,
		TupleType.Symbol:     TupleType,
		MapType.Symbol:       MapType,
		StructType.Symbol:    StructType,
		FunType.Symbol:       FunType,
		RelType.Symbol:       RelType,
	}

	// EmptyType is a type without members.
	// TODO: replace with /bot
	EmptyType = ast.ApplyFn{UnionType, nil}

	// BuiltinRelations maps each builtin predicate to its argument range list
	BuiltinRelations = map[ast.PredicateSym]ast.BaseTerm{
		MatchPrefix: NewRelType(ast.NameBound, ast.NameBound),
		StartsWith:  NewRelType(ast.StringBound, ast.StringBound),
		EndsWith:    NewRelType(ast.StringBound, ast.StringBound),
		Contains:    NewRelType(ast.StringBound, ast.StringBound),
		Filter:      NewRelType(BoolType()),
		// TODO: support float64
		Lt:       NewRelType(ast.NumberBound, ast.NumberBound),
		Le:       NewRelType(ast.NumberBound, ast.NumberBound),
		Gt:       NewRelType(ast.NumberBound, ast.NumberBound),
		Ge:       NewRelType(ast.NumberBound, ast.NumberBound),
		MatchNil: NewRelType(NewListType(ast.Variable{"X"})),
		MatchCons: NewRelType(
			NewListType(ast.Variable{"X"}), ast.Variable{"X"}, NewListType(ast.Variable{"X"})),
		MatchPair: NewRelType(
			NewPairType(ast.Variable{"X"}, ast.Variable{"Y"}), ast.Variable{"X"}, ast.Variable{"Y"}),
		MatchEntry: NewRelType(
			NewMapType(ast.AnyBound, ast.AnyBound), ast.AnyBound),
		// Note: :match_field is treated specially, the following type is only a fallback.
		MatchField: NewRelType(
			ast.AnyBound, ast.NameBound, ast.AnyBound),
		ListMember: NewRelType(ast.Variable{"X"}, NewListType(ast.Variable{"X"})),
	}

	errTypeMismatch = errors.New("type mismatch")
)

// TypeHandle provides functionality related to type expression.
type TypeHandle struct {
	expr ast.BaseTerm
	ctx  map[ast.Variable]ast.BaseTerm
}

// NewBoundHandle constructs a TypeHandle for a bound (may contain type variables).
func NewBoundHandle(expr ast.BaseTerm) (TypeHandle, error) {
	typeVars := make(map[ast.Variable]bool)
	ast.AddVars(expr, typeVars)
	ctx := make(map[ast.Variable]ast.BaseTerm)
	for v := range typeVars {
		// TODO: This is wrong. At this stage, we would need to not only collect
		// type variables but also infer upper and lower type bounds. The context
		// has to map type variables to type constraints.
		ctx[v] = ast.Variable{"_"}
	}
	return NewTypeHandle(ctx, expr)
}

// NewSetHandle constructs a TypeHandle for a (simple) monotype.
func NewSetHandle(expr ast.BaseTerm) (TypeHandle, error) {
	return NewTypeHandle(nil, expr)
}

// NewTypeHandle constructs a TypeHandle.
func NewTypeHandle(ctx map[ast.Variable]ast.BaseTerm, expr ast.BaseTerm) (TypeHandle, error) {
	if err := WellformedType(ctx, expr); err != nil {
		return TypeHandle{}, err
	}
	return TypeHandle{expr, ctx}, nil
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
	tpe, ok := t.expr.(ast.ApplyFn)
	if !ok {
		return false // This never happens.
	}
	switch tpe.Function {
	case PairType:
		fst, snd, err := c.PairValue()
		if err != nil {
			return false
		}
		return TypeHandle{tpe.Args[0], t.ctx}.HasType(fst) &&
			TypeHandle{tpe.Args[1], t.ctx}.HasType(snd)
	case ListType:
		elementType := TypeHandle{tpe.Args[0], t.ctx}
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
		return TypeHandle{expandTupleType(tpe.Args), t.ctx}.HasType(c)
	case MapType:
		if c.IsMapNil() {
			return true
		}
		keyTpe := TypeHandle{tpe.Args[0], t.ctx}
		valTpe := TypeHandle{tpe.Args[1], t.ctx}
		e, err := c.MapValues(func(key ast.Constant, val ast.Constant) error {
			if keyTpe.HasType(key) && valTpe.HasType(val) {
				return nil
			}
			return errTypeMismatch
		}, func() error {
			return nil
		})
		return e == nil && err == nil
	case StructType:
		if c.IsStructNil() {
			return len(tpe.Args) == 0
		}
		fieldTpeMap := make(map[ast.Constant]TypeHandle)
		requiredArgs, err := StructTypeRequiredArgs(tpe)
		if err != nil {
			return false
		}
		for i := 0; i < len(requiredArgs); i++ {
			key := requiredArgs[i].(ast.Constant)
			i++
			val := requiredArgs[i]
			fieldTpeMap[key] = TypeHandle{val, t.ctx}
		}
		optArgs, err := StructTypeOptionaArgs(tpe)
		if err != nil {
			return false
		}
		for _, optArg := range optArgs {
			f := optArg.(ast.ApplyFn)
			fieldTpeMap[f.Args[0].(ast.Constant)] = TypeHandle{f.Args[1], t.ctx}
		}
		seen := make(map[ast.Constant]bool)
		e, err := c.StructValues(func(key ast.Constant, val ast.Constant) error {
			fieldTpe, ok := fieldTpeMap[key]
			if !ok {
				return errTypeMismatch
			}
			seen[key] = true
			if !fieldTpe.HasType(val) {
				return errTypeMismatch
			}
			return nil
		}, func() error {
			return nil
		})
		return e == nil && err == nil && len(fieldTpeMap) == len(seen)
	case UnionType:
		for _, arg := range tpe.Args {
			alt := TypeHandle{arg, t.ctx}
			if alt.HasType(c) {
				return true
			}
		}
		return false
	case SingletonType:
		d := tpe.Args[0]
		return c.Equals(d)
	}
	return false
}

func hasBaseType(typeExpr ast.Constant, c ast.Constant) bool {
	switch typeExpr {
	case ast.AnyBound:
		return true
	case ast.Float64Bound:
		return c.Type == ast.Float64Type
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

// WellformedBound returns an error if expr is not a valid bound expression.
//
// A bound expression is well formed if the type-in-context that is formed by
// closing over its type variables is a well-formed type expression.
//
// An unconstrained type variable still counts as a well-formed bound expression.
// If it is not constrained by type expressions in the other bounds, it will
// be treated as /any.
func WellformedBound(expr ast.BaseTerm) error {
	typeVars := make(map[ast.Variable]bool)
	ast.AddVars(expr, typeVars)
	ctx := make(map[ast.Variable]ast.BaseTerm)
	for v := range typeVars {
		// We get away with this because well-formedness checking does not
		// consider constraints.
		ctx[v] = ast.Variable{"_"}
	}
	return WellformedType(ctx, expr)
}

// WellformedType returns an error if expr is not a well-formed type-in-context.
// expr is by convention not a reltype.
func WellformedType(ctx map[ast.Variable]ast.BaseTerm, expr ast.BaseTerm) error {
	switch expr := expr.(type) {
	case ast.Constant:
		if IsBaseTypeExpression(expr) || expr.Type == ast.NameType {
			return nil
		}
		return fmt.Errorf("not a base type expression: %v", expr)
	case ast.Variable:
		if ctx == nil || ctx[expr] == nil {
			return fmt.Errorf("unexpected type variable: %v context: %v", expr, ctx)
		}
		return nil

	case ast.ApplyFn:
		fn, ok := TypeConstructors[expr.Function.Symbol]
		if !ok {
			return fmt.Errorf("not a structured type expression: %v", expr)
		}
		args := expr.Args
		if fn == FunType {
			return CheckFunTypeExpression(ctx, expr)
		}
		if fn.Arity != -1 && len(args) != fn.Arity {
			return fmt.Errorf("expected %d arguments in type expression %v ", fn.Arity, expr)
		}
		if fn == UnionType && len(args) <= 0 {
			return fmt.Errorf("union type must not be empty %v ", expr)
		}
		if fn == TupleType && len(args) <= 2 {
			return fmt.Errorf("tuple type must have more than 2 args %v ", expr)
		}
		if fn == StructType {
			requiredArgs, err := StructTypeRequiredArgs(expr)
			if err != nil {
				return err
			}
			if len(requiredArgs)%2 != 0 {
				return fmt.Errorf("struct type must have even number of required arguments %v ", expr)
			}
			for i := 0; i < len(requiredArgs); i++ {
				key := requiredArgs[i]
				if c, ok := key.(ast.Constant); !ok || c.Type != ast.NameType {
					return fmt.Errorf("in a struct type expression, odd arguments must be name constants, argument %d (%v) is not %v ", i, key, expr)
				}
				i++
				tpe := requiredArgs[i]
				if err := WellformedType(ctx, tpe); err != nil {
					return fmt.Errorf("in a struct type expression %v : %w", expr, err)
				}
			}
			return nil
		}

		for _, arg := range args {
			if err := WellformedType(ctx, arg); err != nil {
				return err
			}
		}
		return nil
	default:
		return fmt.Errorf("CheckTypeExpression: unexpected case %v %T", expr, expr)
	}
}

// CheckFunTypeExpression checks a function type expression.
func CheckFunTypeExpression(ctx map[ast.Variable]ast.BaseTerm, expr ast.ApplyFn) error {
	if len(expr.Args) == 0 {
		return fmt.Errorf("expected at least 1 argument in function type expression %v ", expr)
	}
	codomain := expr.Args[0]
	argTpes := expr.Args[1:]
	vars := make(map[ast.Variable]bool)
	for _, arg := range argTpes {
		ast.AddVars(arg, vars)
	}
	codomainVars := make(map[ast.Variable]bool)
	ast.AddVars(codomain, codomainVars)
	// Check inclusion.
	for v := range codomainVars {
		if vars[v] {
			continue
		}
		return fmt.Errorf("type variable %v not in domain vars %v ", v, vars)
	}

	ctxMap := make(map[ast.Variable]ast.BaseTerm)
	for v := range vars {
		ctxMap[v] = ast.AnyBound
	}
	for _, argTpe := range argTpes {
		if err := WellformedType(ctxMap, argTpe); err != nil {
			return err
		}
	}
	return WellformedType(ctxMap, codomain)
}

// SetConforms returns true if |- left <: right for set expression.
func SetConforms(typeCtx map[ast.Variable]ast.BaseTerm, left ast.BaseTerm, right ast.BaseTerm) bool {
	if left.Equals(right) || right.Equals(ast.AnyBound) || left.Equals(ast.BotBound) {
		return true
	}
	if leftTuple, ok := left.(ast.ApplyFn); ok && leftTuple.Function.Symbol == RelType.Symbol {
		if rightTuple, ok := right.(ast.ApplyFn); ok && rightTuple.Function.Symbol == RelType.Symbol {
			for i, leftArg := range leftTuple.Args {
				if !SetConforms(typeCtx, leftArg, rightTuple.Args[i]) {
					return false
				}
			}
			return true
		}
	}
	leftApply, leftApplyOk := left.(ast.ApplyFn)
	rightApply, rightApplyOk := right.(ast.ApplyFn)
	if leftApplyOk && leftApply.Function.Symbol == UnionType.Symbol {
		for _, leftItem := range leftApply.Args {
			if !SetConforms(typeCtx, leftItem, right) {
				return false
			}
		}
		return true
	}
	if rightApplyOk && rightApply.Function.Symbol == UnionType.Symbol {
		for _, rightItem := range rightApply.Args {
			if SetConforms(typeCtx, left, rightItem) {
				return true
			}
		}
	}

	return TypeConforms(typeCtx, left, right)
}

// TypeConforms returns true if ctx |- left <: right.
// The arguments left and right cannot be RelType or UnionType
func TypeConforms(ctx map[ast.Variable]ast.BaseTerm, left ast.BaseTerm, right ast.BaseTerm) bool {
	if left.Equals(right) || right.Equals(ast.AnyBound) || left.Equals(ast.BotBound) {
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
	if leftVar, ok := left.(ast.Variable); ok {
		return TypeConforms(ctx, ctx[leftVar], right)
	}
	if rightVar, ok := right.(ast.Variable); ok {
		return TypeConforms(ctx, left, ctx[rightVar])
	}
	leftApply, leftApplyOk := left.(ast.ApplyFn)
	rightApply, rightApplyOk := right.(ast.ApplyFn)
	if leftApplyOk && leftApply.Function.Symbol == FunType.Symbol &&
		rightApplyOk && rightApply.Function.Symbol == FunType.Symbol {
		// FunType subtyping is covariant in codomain, contravariant in domain
		// E.g. /genus_species <: /name and /animal/bird <: /animal
		// therefore FunType(/genus_species <= /animal) <: FunType(/name <= /animal/bird)
		leftCodomain, rightCodomain := leftApply.Args[0], rightApply.Args[0]
		if !TypeConforms(ctx, leftCodomain, rightCodomain) {
			return false
		}
		leftDomain, rightDomain := leftApply.Args[1:], rightApply.Args[1:]

		for i, leftArg := range leftDomain {
			if !TypeConforms(ctx, rightDomain[i], leftArg) {
				return false
			}
		}
		return true
	}

	if leftApplyOk && leftApply.Function.Symbol == ListType.Symbol {
		if rightApplyOk && rightApply.Function.Symbol == ListType.Symbol {
			return TypeConforms(ctx, leftApply.Args[0], rightApply.Args[0])
		}
	}
	if leftApplyOk && leftApply.Function.Symbol == MapType.Symbol {
		if rightApplyOk && rightApply.Function.Symbol == MapType.Symbol {
			return TypeConforms(ctx, rightApply.Args[0], leftApply.Args[0]) &&
				TypeConforms(ctx, leftApply.Args[1], rightApply.Args[1])
		}
	}
	if leftApplyOk && leftApply.Function.Symbol == StructType.Symbol {
		if rightApplyOk && rightApply.Function.Symbol == StructType.Symbol {
			leftRequired, err := StructTypeRequiredArgs(left)
			if err != nil {
				return false
			}
			rightRequired, err := StructTypeRequiredArgs(right)
			if err != nil {
				return false
			}
			if len(leftRequired) < len(rightRequired) {
				return false
			}
			leftMap := make(map[string]ast.BaseTerm)
			for i := 0; i < len(leftRequired); i++ {
				leftKey, _ := leftRequired[i].(ast.Constant)
				i++
				leftMap[leftKey.Symbol] = leftRequired[i]
			}

			for j := 0; j < len(rightRequired); j++ {
				rightKey, _ := rightRequired[j].(ast.Constant)
				j++
				rightTpe := rightRequired[j]
				leftTpe, ok := leftMap[rightKey.Symbol]
				if !ok || !TypeConforms(ctx, leftTpe, rightTpe) {
					return false
				}
			}
			leftOpt, err := StructTypeOptionaArgs(left)
			if err != nil {
				return false
			}
			rightOpt, err := StructTypeOptionaArgs(right)
			if err != nil {
				return false
			}
			leftOptMap := make(map[string]ast.BaseTerm)
			for _, opt := range leftOpt {
				optApply, ok := opt.(ast.ApplyFn)
				if !ok {
					return false
				}
				leftOptMap[optApply.Args[0].(ast.Constant).Symbol] = optApply.Args[1]
			}
			for _, opt := range rightOpt {
				optApply, ok := opt.(ast.ApplyFn)
				if !ok {
					return false
				}
				key := optApply.Args[0].(ast.Constant).Symbol
				rightTpe := optApply.Args[1]
				leftTpe, ok := leftMap[key]
				if ok && !TypeConforms(ctx, leftTpe, rightTpe) {
					return false
				}
				if !ok {
					leftTpe, ok := leftOptMap[key]
					if ok && !TypeConforms(ctx, leftTpe, rightTpe) {
						return false
					}
				}
			}
			return true
		}
	}

	if leftApplyOk && leftApply.Function.Symbol == PairType.Symbol {
		if rightApplyOk && rightApply.Function.Symbol == PairType.Symbol {
			return TypeConforms(ctx, leftApply.Args[0], rightApply.Args[0]) &&
				TypeConforms(ctx, leftApply.Args[1], rightApply.Args[1])
		}
	}
	if leftTuple, ok := left.(ast.ApplyFn); ok && leftTuple.Function.Symbol == TupleType.Symbol {
		if rightTuple, ok := right.(ast.ApplyFn); ok && rightTuple.Function.Symbol == TupleType.Symbol {
			for i, leftArg := range leftTuple.Args {
				if !TypeConforms(ctx, leftArg, rightTuple.Args[i]) {
					return false
				}
			}
			return true
		}
	}

	return false
}

func expandTupleType(args []ast.BaseTerm) ast.BaseTerm {
	res := NewPairType(args[len(args)-2], args[len(args)-1])
	for j := len(args) - 3; j >= 0; j-- {
		res = NewPairType(args[j], res)
	}
	return res
}

// UpperBound returns upper bound of set expressions.
func UpperBound(typeCtx map[ast.Variable]ast.BaseTerm, typeExprs []ast.BaseTerm) ast.BaseTerm {
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
		return EmptyType
	}
	reduced := []ast.BaseTerm{worklist[0]}
	worklist = worklist[1:]
typeExprLoop:
	for _, typeExpr := range worklist {
		for i, existing := range reduced {
			if SetConforms(typeCtx, typeExpr, existing) {
				continue typeExprLoop
			}
			if SetConforms(typeCtx, existing, typeExpr) {
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

func intersectType(typeCtx map[ast.Variable]ast.BaseTerm, a, b ast.BaseTerm) ast.BaseTerm {
	if a.Equals(b) {
		return a
	}
	if a.Equals(ast.AnyBound) {
		return b
	}
	if b.Equals(ast.AnyBound) {
		return a
	}
	if typeVar, ok := a.(ast.Variable); ok {
		bound, ok := typeCtx[typeVar]
		if !ok {
			return EmptyType
		}
		return intersectType(typeCtx, bound, b)
	}
	if typeVar, ok := b.(ast.Variable); ok {
		bound, ok := typeCtx[typeVar]
		if !ok {
			return EmptyType
		}
		return intersectType(typeCtx, a, bound)
	}
	if SetConforms(typeCtx, a, b) {
		return a
	}
	if SetConforms(typeCtx, b, a) {
		return b
	}
	if aUnion, ok := a.(ast.ApplyFn); ok && aUnion.Function == UnionType {
		var res []ast.BaseTerm
		for _, elem := range aUnion.Args {
			if u := intersectType(typeCtx, elem, b); !u.Equals(EmptyType) {
				res = append(res, u)
			}
		}
		return UpperBound(typeCtx, res)
	}
	if bUnion, ok := b.(ast.ApplyFn); ok && bUnion.Function == UnionType {
		var res []ast.BaseTerm
		for _, elem := range bUnion.Args {
			if SetConforms(typeCtx, a, elem) {
				res = append(res, a)
			} else if SetConforms(typeCtx, elem, a) {
				res = append(res, elem)
			}
		}
		return UpperBound(typeCtx, res)
	}

	return EmptyType
}

// LowerBound returns a lower bound of set expressions.
func LowerBound(typeCtx map[ast.Variable]ast.BaseTerm, typeExprs []ast.BaseTerm) ast.BaseTerm {
	var typeExpr ast.BaseTerm = ast.AnyBound
	for _, t := range typeExprs {
		if typeExpr = intersectType(typeCtx, typeExpr, t); typeExpr.Equals(EmptyType) {
			return EmptyType
		}
	}
	return typeExpr
}

// CreateListType applies given type to a list.
func CreateListType(bound ast.Constant) ast.ApplyFn {
	return ast.ApplyFn{ListType, []ast.BaseTerm{bound}}
}
