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

package symbols

import (
	"fmt"

	"github.com/google/mangle/ast"
)

const debug = false

// IsBaseTypeExpression returns true if c is a base type expression.
// A name constant is /foo is not a base type expression.
func IsBaseTypeExpression(c ast.Constant) bool {
	switch c {
	case ast.AnyBound:
		return true
	case ast.Float64Bound:
		return true
	case ast.NumberBound:
		return true
	case ast.StringBound:
		return true
	default:
		return false
	}
}

func newTypeExpr(typeOp ast.FunctionSym, typeArgs ...ast.BaseTerm) ast.ApplyFn {
	return ast.ApplyFn{typeOp, typeArgs}
}

func typeOp(typeExpr ast.BaseTerm) *ast.FunctionSym {
	if expr, ok := typeExpr.(ast.ApplyFn); ok {
		return &expr.Function
	}
	return nil
}

func typeArgs(typeExpr ast.BaseTerm) []ast.BaseTerm {
	if expr, ok := typeExpr.(ast.ApplyFn); ok {
		return expr.Args
	}
	return nil
}

// NewPairType returns a new PairType.
func NewPairType(left, right ast.BaseTerm) ast.ApplyFn {
	return newTypeExpr(PairType, left, right)
}

// NewTupleType returns a new TupleType.
func NewTupleType(parts ...ast.BaseTerm) ast.ApplyFn {
	return newTypeExpr(TupleType, parts...)
}

// NewListType returns a new ListType.
func NewListType(elem ast.BaseTerm) ast.ApplyFn {
	return newTypeExpr(ListType, elem)
}

// NewMapType returns a new MapType.
func NewMapType(keyType, valueType ast.BaseTerm) ast.ApplyFn {
	return newTypeExpr(MapType, keyType, valueType)
}

// NewOpt wraps a label-type pair inside a StructType.
// fn:optional(/foo, /string)
func NewOpt(label, tpe ast.BaseTerm) ast.ApplyFn {
	return newTypeExpr(Optional, label, tpe)
}

// NewStructType returns a new StructType.
func NewStructType(args ...ast.BaseTerm) ast.ApplyFn {
	return newTypeExpr(StructType, args...)
}

// NewFunType returns a new function type.
// Res <- Arg1, ..., ArgN
func NewFunType(res ast.BaseTerm, args ...ast.BaseTerm) ast.ApplyFn {
	return newTypeExpr(FunType, append([]ast.BaseTerm{res}, args...)...)
}

// NewRelType returns a new relation type.
func NewRelType(args ...ast.BaseTerm) ast.ApplyFn {
	return newTypeExpr(RelType, args...)
}

// NewUnionType returns a new UnionType.
func NewUnionType(elems ...ast.BaseTerm) ast.ApplyFn {
	return newTypeExpr(UnionType, elems...)
}

// IsListTypeExpression returns true if tpe is a ListType.
func IsListTypeExpression(tpe ast.BaseTerm) bool {
	return *typeOp(tpe) == ListType
}

// IsMapTypeExpression returns true if tpe is a MapType.
func IsMapTypeExpression(tpe ast.BaseTerm) bool {
	return *typeOp(tpe) == MapType
}

// IsStructTypeExpression returns true if tpe is a StructType.
func IsStructTypeExpression(tpe ast.BaseTerm) bool {
	return (*typeOp(tpe)).Symbol == StructType.Symbol
}

// IsUnionTypeExpression returns true if tpe is a UnionType.
func IsUnionTypeExpression(tpe ast.BaseTerm) bool {
	return (*typeOp(tpe)).Symbol == UnionType.Symbol
}

// IsRelTypeExpression returns true if tpe is a RelType.
func IsRelTypeExpression(tpe ast.BaseTerm) bool {
	return (*typeOp(tpe)).Symbol == RelType.Symbol
}

// ListTypeArg returns the type argument of a ListType.
func ListTypeArg(tpe ast.BaseTerm) (ast.BaseTerm, error) {
	if debug && !IsListTypeExpression(tpe) {
		return nil, fmt.Errorf("not a list type expression: %v", tpe)
	}
	return typeArgs(tpe)[0], nil
}

// MapTypeArgs returns the type arguments of a MapType.
func MapTypeArgs(tpe ast.BaseTerm) (ast.BaseTerm, ast.BaseTerm, error) {
	if debug && !IsMapTypeExpression(tpe) {
		return nil, nil, fmt.Errorf("not a map type expression: %v", tpe)
	}
	args := typeArgs(tpe)
	return args[0], args[1], nil
}

// IsOptional returns true if an argument of fn:Struct is an optional field.
func IsOptional(structElem ast.BaseTerm) bool {
	opt, ok := structElem.(ast.ApplyFn)
	return ok && opt.Function.Symbol == Optional.Symbol
}

// StructTypeRequiredArgs returns type arguments of a StructType.
func StructTypeRequiredArgs(tpe ast.BaseTerm) ([]ast.BaseTerm, error) {
	if debug && !IsStructTypeExpression(tpe) {
		return nil, fmt.Errorf("not a struct type expression: %v", tpe)
	}
	var required []ast.BaseTerm
	for _, arg := range typeArgs(tpe) {
		if !IsOptional(arg) {
			required = append(required, arg)
		}
	}
	return required, nil
}

// StructTypeOptionaArgs returns type arguments of a StructType.
func StructTypeOptionaArgs(tpe ast.BaseTerm) ([]ast.BaseTerm, error) {
	if debug && !IsStructTypeExpression(tpe) {
		return nil, fmt.Errorf("not a struct type expression: %v", tpe)
	}
	var optional []ast.BaseTerm
	for _, arg := range typeArgs(tpe) {
		if IsOptional(arg) {
			optional = append(optional, arg)
		}
	}
	return optional, nil
}

// StructTypeField returns field type for given field.
func StructTypeField(tpe ast.BaseTerm, field ast.Constant) (ast.BaseTerm, error) {
	if debug && !IsStructTypeExpression(tpe) {
		return nil, fmt.Errorf("not a struct type expression: %v", tpe)
	}
	elems := typeArgs(tpe)
	for i := 0; i < len(elems); i++ {
		var key ast.BaseTerm
		arg := elems[i]
		if IsOptional(arg) {
			key = arg.(ast.ApplyFn).Args[0]
		} else {
			key = arg
		}
		if key.Equals(field) {
			if IsOptional(arg) {
				return arg.(ast.ApplyFn).Args[1], nil
			}
			i++
			return elems[i], nil
		}
	}
	return nil, fmt.Errorf("no field %v in %v", field, tpe)
}

// UnionTypeArgs returns type arguments of a UnionType.
func UnionTypeArgs(tpe ast.BaseTerm) ([]ast.BaseTerm, error) {
	if debug && !IsUnionTypeExpression(tpe) {
		return nil, fmt.Errorf("not a union type expression: %v", tpe)
	}
	return typeArgs(tpe), nil
}

// RemoveFromUnionType given T, removes S from a union type {..., S, ...} if S<:T.
func RemoveFromUnionType(tpeToRemove, unionTpe ast.BaseTerm) (ast.BaseTerm, error) {
	if debug && !IsUnionTypeExpression(unionTpe) {
		return nil, fmt.Errorf("not a union type expression: %v", unionTpe)
	}
	var newArgs []ast.BaseTerm
	for _, arg := range typeArgs(unionTpe) {
		if SetConforms(arg, tpeToRemove) {
			continue
		}
		newArgs = append(newArgs, arg)
	}
	return NewUnionType(newArgs...), nil
}

// RelTypeArgs returns type arguments of a RelType.
func RelTypeArgs(tpe ast.BaseTerm) ([]ast.BaseTerm, error) {
	if debug && !IsRelTypeExpression(tpe) {
		return nil, fmt.Errorf("not a relation type expression: %v", tpe)
	}
	return typeArgs(tpe), nil
}

// relTypesFromDecl converts bounds a list of RelTypes.
func relTypesFromDecl(decl ast.Decl) ([]ast.BaseTerm, error) {
	if len(decl.Bounds) == 0 {
		return nil, fmt.Errorf("no bound decls in %v", decl)
	}
	relTypes := make([]ast.BaseTerm, len(decl.Bounds))
	for i, boundDecl := range decl.Bounds {
		relTypes[i] = NewRelType(boundDecl.Bounds...)
	}
	return relTypes, nil
}

// RelTypeFromAlternatives converts list of rel types bounds to union of relation types.
// It is assumed that each alternatives is a RelType.
// An empty list of alternatives corresponds to the empty type fn:Union().
func RelTypeFromAlternatives(alternatives []ast.BaseTerm) ast.BaseTerm {
	if len(alternatives) == 1 {
		return alternatives[0]
	}
	// Could be reduced to a single alternative in some cases.
	return NewUnionType(alternatives...)
}

// RelTypeExprFromDecl converts bounds to relation type expression.
func RelTypeExprFromDecl(decl ast.Decl) (ast.BaseTerm, error) {
	alts, err := relTypesFromDecl(decl)
	if err != nil {
		return nil, err
	}
	return RelTypeFromAlternatives(alts), nil
}

// RelTypeAlternatives converts a relation type expression to a list of alternatives relTypes.
func RelTypeAlternatives(relTypeExpr ast.BaseTerm) []ast.BaseTerm {
	if IsUnionTypeExpression(relTypeExpr) {
		relTypes, _ := UnionTypeArgs(relTypeExpr)
		return relTypes
	}
	return []ast.BaseTerm{relTypeExpr}
}
