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

	"codeberg.org/TauCeti/mangle-go/ast"
)

const debug = false

// IsBaseTypeExpression returns true if c is a base type expression.
// A name constant is /foo is not a base type expression.
func IsBaseTypeExpression(c ast.Constant) bool {
	switch c {
	case ast.AnyBound:
		return true
	case ast.BotBound:
		return true
	case ast.Float64Bound:
		return true
	case ast.NumberBound:
		return true
	case ast.StringBound:
		return true
	case ast.BytesBound:
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

// NewOptionType returns a new ListType.
func NewOptionType(elem ast.BaseTerm) ast.ApplyFn {
	return newTypeExpr(OptionType, elem)
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

// NewSingletonType returns a new SingletonType.
func NewSingletonType(d ast.Constant) ast.ApplyFn {
	return newTypeExpr(SingletonType, d)
}

// BoolType returns a type named in honor of George Boole.
func BoolType() ast.ApplyFn {
	return NewUnionType(NewSingletonType(ast.TrueConstant), NewSingletonType(ast.FalseConstant))
}

// NewTaggedUnionType returns a new TaggedUnionType.
// variantPairs must be alternating tag constants and struct types.
func NewTaggedUnionType(tagField ast.Constant, variantPairs ...ast.BaseTerm) ast.ApplyFn {
	args := make([]ast.BaseTerm, 0, 1+len(variantPairs))
	args = append(args, tagField)
	args = append(args, variantPairs...)
	return newTypeExpr(TaggedUnionType, args...)
}

// IsTaggedUnionTypeExpression returns true if tpe is a TaggedUnionType.
func IsTaggedUnionTypeExpression(tpe ast.BaseTerm) bool {
	op := typeOp(tpe)
	return op != nil && op.Symbol == TaggedUnionType.Symbol
}

// TaggedUnionTagField returns the tag field name of a TaggedUnionType.
func TaggedUnionTagField(tpe ast.BaseTerm) (ast.Constant, error) {
	args := typeArgs(tpe)
	if args == nil || len(args) < 3 {
		return ast.Constant{}, fmt.Errorf("not a tagged union type expression: %v", tpe)
	}
	c, ok := args[0].(ast.Constant)
	if !ok {
		return ast.Constant{}, fmt.Errorf("tag field is not a constant: %v", args[0])
	}
	return c, nil
}

// TaggedUnionVariants returns the variant tags and their struct types.
func TaggedUnionVariants(tpe ast.BaseTerm) ([]ast.Constant, []ast.BaseTerm, error) {
	args := typeArgs(tpe)
	if args == nil || len(args) < 3 || len(args)%2 != 1 {
		return nil, nil, fmt.Errorf("not a tagged union type expression: %v", tpe)
	}
	n := (len(args) - 1) / 2
	tags := make([]ast.Constant, n)
	types := make([]ast.BaseTerm, n)
	for i := 0; i < n; i++ {
		c, ok := args[1+2*i].(ast.Constant)
		if !ok {
			return nil, nil, fmt.Errorf("variant tag is not a constant: %v", args[1+2*i])
		}
		tags[i] = c
		types[i] = args[2+2*i]
	}
	return tags, types, nil
}

// ExpandTaggedUnionType expands a TaggedUnionType into the equivalent Union of Struct types.
// fn:TaggedUnion(/kind, /a, fn:Struct(/x: /number), /b, fn:Struct(/y: /string))
// becomes:
// fn:Union(fn:Struct(/kind, fn:Singleton(/a), /x, /number), fn:Struct(/kind, fn:Singleton(/b), /y, /string))
func ExpandTaggedUnionType(tpe ast.BaseTerm) (ast.ApplyFn, error) {
	tagField, err := TaggedUnionTagField(tpe)
	if err != nil {
		return ast.ApplyFn{}, err
	}
	tags, structTypes, err := TaggedUnionVariants(tpe)
	if err != nil {
		return ast.ApplyFn{}, err
	}
	alternatives := make([]ast.BaseTerm, len(tags))
	for i, tag := range tags {
		variantArgs := typeArgs(structTypes[i])
		// Build struct: tag_field, Singleton(tag), ...variant fields
		structArgs := make([]ast.BaseTerm, 0, 2+len(variantArgs))
		structArgs = append(structArgs, tagField, NewSingletonType(tag))
		structArgs = append(structArgs, variantArgs...)
		alternatives[i] = NewStructType(structArgs...)
	}
	return NewUnionType(alternatives...), nil
}

// expandTaggedUnionForBounds is like ExpandTaggedUnionType but uses /name for the
// tag field type instead of fn:Singleton(tag). This is less precise but compatible
// with the bounds checker's type inference, which infers /name for name constants.
func expandTaggedUnionForBounds(tpe ast.BaseTerm) (ast.ApplyFn, error) {
	tagField, err := TaggedUnionTagField(tpe)
	if err != nil {
		return ast.ApplyFn{}, err
	}
	_, structTypes, err := TaggedUnionVariants(tpe)
	if err != nil {
		return ast.ApplyFn{}, err
	}
	alternatives := make([]ast.BaseTerm, len(structTypes))
	for i := range structTypes {
		variantArgs := typeArgs(structTypes[i])
		structArgs := make([]ast.BaseTerm, 0, 2+len(variantArgs))
		structArgs = append(structArgs, tagField, ast.NameBound)
		structArgs = append(structArgs, variantArgs...)
		alternatives[i] = NewStructType(structArgs...)
	}
	return NewUnionType(alternatives...), nil
}

// IsListTypeExpression returns true if tpe is a ListType.
func IsListTypeExpression(tpe ast.BaseTerm) bool {
	op := typeOp(tpe)
	return op != nil && *op == ListType
}

// IsMapTypeExpression returns true if tpe is a MapType.
func IsMapTypeExpression(tpe ast.BaseTerm) bool {
	op := typeOp(tpe)
	return op != nil && *op == MapType
}

// IsStructTypeExpression returns true if tpe is a StructType.
func IsStructTypeExpression(tpe ast.BaseTerm) bool {
	op := typeOp(tpe)
	return op != nil && op.Symbol == StructType.Symbol
}

// IsFunTypeExpression returns true if tpe is a UnionType.
func IsFunTypeExpression(tpe ast.BaseTerm) bool {
	op := typeOp(tpe)
	return op != nil && op.Symbol == FunType.Symbol
}

// IsUnionTypeExpression returns true if tpe is a UnionType.
func IsUnionTypeExpression(tpe ast.BaseTerm) bool {
	op := typeOp(tpe)
	return op != nil && op.Symbol == UnionType.Symbol
}

// IsRelTypeExpression returns true if tpe is a RelType.
func IsRelTypeExpression(tpe ast.BaseTerm) bool {
	op := typeOp(tpe)
	return op != nil && op.Symbol == RelType.Symbol
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
	if IsTaggedUnionTypeExpression(tpe) {
		expanded, err := ExpandTaggedUnionType(tpe)
		if err != nil {
			return nil, err
		}
		return StructTypeField(expanded, field)
	}
	if IsUnionTypeExpression(tpe) { // Project
		src, _ := UnionTypeArgs(tpe)
		alternatives := []ast.BaseTerm{}
		for _, s := range src {
			projected, err := StructTypeField(s, field)
			if err != nil {
				continue
			}
			alternatives = append(alternatives, projected)
		}
		if len(alternatives) == 1 {
			return alternatives[0], nil
		}
		return NewUnionType(alternatives...), nil
	}
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

// FunTypeResult returns result type of function type.
func FunTypeResult(tpe ast.BaseTerm) (ast.BaseTerm, error) {
	if debug && !IsFunTypeExpression(tpe) {
		return nil, fmt.Errorf("not a function type expression: %v", tpe)
	}
	return typeArgs(tpe)[0], nil
}

// FunTypeArgs returns function arguments of function type.
func FunTypeArgs(tpe ast.BaseTerm) ([]ast.BaseTerm, error) {
	if debug && !IsFunTypeExpression(tpe) {
		return nil, fmt.Errorf("not a function type expression: %v", tpe)
	}
	return typeArgs(tpe)[1:], nil
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
		if SetConforms(nil /*TODO*/, arg, tpeToRemove) {
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

// GetTypeContext returns type context containing all type vars, with /any bound.
func GetTypeContext(typeExpr ast.BaseTerm) map[ast.Variable]ast.BaseTerm {
	typeCtx := map[ast.Variable]ast.BaseTerm{}
	typeVars := map[ast.Variable]bool{}
	ast.AddVars(typeExpr, typeVars)
	for v := range typeVars {
		typeCtx[v] = ast.AnyBound
	}
	return typeCtx
}
