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

package analysis

import (
	"fmt"
	"sort"
	"strings"

	"codeberg.org/TauCeti/mangle-go/ast"
	"codeberg.org/TauCeti/mangle-go/builtin"
	"codeberg.org/TauCeti/mangle-go/symbols"
	"codeberg.org/TauCeti/mangle-go/unionfind"
)

// BoundsAnalyzer to infer and check bounds.
type BoundsAnalyzer struct {
	programInfo *ProgramInfo
	// Trie of name constants from declarations different from /any,/name,/number,/string.
	nameTrie symbols.NameTrie
	RulesMap map[ast.PredicateSym][]ast.Clause
	// maps `foo`` to either RelType[...] or fn:Union(RelType[...]...RelType[...])
	relTypeMap map[ast.PredicateSym]ast.BaseTerm
	// maps `foo`` to either RelType[...] or fn:Union(RelType[...]...RelType[...])
	initialFactMap map[ast.PredicateSym]ast.BaseTerm
	// maps `foo`` to either RelType[...] or fn:Union(RelType[...]...RelType[...])
	inferred map[ast.PredicateSym]ast.BaseTerm
	visiting map[ast.PredicateSym]bool
}

// We extract name constants from declarations. We build a trie of names
// that play the role of type expressions. so we can later map name constants to their
// corresponding, most precise (longest prefix) type expression.
func collectNames(extraPredicates map[ast.PredicateSym]ast.Decl, decls map[ast.PredicateSym]*ast.Decl) symbols.NameTrie {
	nameTrie := symbols.NewNameTrie()
	handleDecl := func(d ast.Decl) {
		for _, bs := range d.Bounds {
			for _, typeExpr := range bs.Bounds {
				nameTrie.Collect(typeExpr)
			}
		}
	}
	for _, d := range extraPredicates {
		handleDecl(d)
	}
	for _, d := range decls {
		handleDecl(*d)
	}
	return nameTrie
}

func newBoundsAnalyzer(programInfo *ProgramInfo, nameTrie symbols.NameTrie, initialFacts []ast.Atom, rulesMap map[ast.PredicateSym][]ast.Clause) (*BoundsAnalyzer, error) {
	var err error
	relTypeMap := make(map[ast.PredicateSym]ast.BaseTerm)

	initialFactTypes := make(map[ast.PredicateSym]map[uint64]ast.BaseTerm)
	for _, f := range initialFacts {
		// From a unit clause (initial fact) `foo(2, 'a')` we
		// derive an "observation" [/int /string].
		observation := make([]ast.BaseTerm, len(f.Args))
		for i, arg := range f.Args {
			observation[i] = boundOfArg(arg, nil, nameTrie)
		}
		relType := symbols.NewRelType(observation...)
		m, ok := initialFactTypes[f.Predicate]
		if !ok {
			m = make(map[uint64]ast.BaseTerm)
			initialFactTypes[f.Predicate] = m
		}
		m[relType.Hash()] = relType
	}

	// We gather all relation types.
	initialFactMap := make(map[ast.PredicateSym]ast.BaseTerm, len(initialFactTypes))
	for pred, relTypeMap := range initialFactTypes {
		relTypes := make([]ast.BaseTerm, 0, len(relTypeMap))
		for _, relType := range relTypeMap {
			relTypes = append(relTypes, relType)
		}
		if len(relTypes) == 1 {
			initialFactMap[pred] = relTypes[0]
		} else {
			initialFactMap[pred] = symbols.NewUnionType(relTypes...)
		}
	}

	for _, decl := range programInfo.Decls {
		if decl.IsSynthetic() && !decl.IsExtensional() {
			continue
		}
		// Populate relTypes with type info from declarations.
		// For intensional predicates, we include only user-supplied declaration.
		// For extensional predicates, we include declarations whether or not they
		// are synthetic.
		pred := decl.DeclaredAtom.Predicate
		relTypeMap[pred], err = symbols.RelTypeExprFromDecl(*programInfo.Decls[pred])
		if err != nil {
			return nil, err
		}
	}

	// Populate relTypes with all builtin relation types.
	for pred, relTypeExpr := range symbols.BuiltinRelations {
		relTypeMap[pred] = relTypeExpr
	}

	visiting := make(map[ast.PredicateSym]bool)
	return &BoundsAnalyzer{
		programInfo, nameTrie, rulesMap, relTypeMap, initialFactMap,
		make(map[ast.PredicateSym]ast.BaseTerm), visiting}, nil
}

// BoundsCheck checks whether the rules respect the bounds.
func (bc *BoundsAnalyzer) BoundsCheck() error {
	predMap := make(map[string]ast.PredicateSym)
	for pred := range bc.programInfo.IdbPredicates {
		predMap[pred.Symbol] = pred
	}
	for pred := range bc.initialFactMap {
		predMap[pred.Symbol] = pred // overwrite ok
	}
	preds := make([]ast.PredicateSym, 0, len(predMap))
	for _, v := range predMap {
		preds = append(preds, v)
	}
	// Fix the order in which we do our checks.
	sort.Slice(preds, func(i, j int) bool { return preds[i].Symbol < preds[j].Symbol })
	for _, pred := range preds {
		if err := bc.inferAndCheckBounds(pred); err != nil {
			return err
		}
	}

	return nil
}

// Entry point for bounds checking.
func (bc *BoundsAnalyzer) inferAndCheckBounds(pred ast.PredicateSym) error {
	decl, ok := bc.programInfo.Decls[pred]
	if !ok {
		return nil // This should not happen.
	}
	if !decl.IsSynthetic() {
		return bc.checkClauses(decl)
	}
	_, err := bc.inferRelTypes(pred)
	return err
}

// checkRelTypes takes a pred that has a declaration supplied by the user.
// It checks every clause, including unit clauses ("initial facts").
func (bc *BoundsAnalyzer) checkClauses(decl *ast.Decl) error {
	pred := decl.DeclaredAtom.Predicate
	clauses := bc.RulesMap[pred]

	declaredRelTypeExpr, err := symbols.RelTypeExprFromDecl(*decl)
	if err != nil {
		return err
	}

	// Handle unit clauses (initial facts). We inferred some relation types, each must
	// conform to the declaration (at least one alternative among declared ones).
	if initialFactRelTypes, ok := bc.initialFactMap[pred]; ok {
		for _, inferred := range symbols.RelTypeAlternatives(initialFactRelTypes) {
			typeCtx := symbols.GetTypeContext(declaredRelTypeExpr)
			if !symbols.SetConforms(typeCtx, inferred, declaredRelTypeExpr) {
				return fmt.Errorf("found unit clause with %v that does not conform to any decl %v", inferred, declaredRelTypeExpr)
			}
		}
	}

	for _, clause := range clauses {
		ic := newInferContext(bc, decl, &clause)
		inferredRelTypeExpr, err := ic.inferRelTypesFromClause()
		if err != nil {
			return err
		}
		typeCtx := symbols.GetTypeContext(declaredRelTypeExpr)
		if !symbols.SetConforms(typeCtx, inferredRelTypeExpr, declaredRelTypeExpr) {
			var rules strings.Builder
			for _, r := range bc.RulesMap[pred] {
				rules.WriteString(r.String() + "\n")
			}
			return fmt.Errorf("type mismatch for pred %v rule: %s \n inferred: %v\nvs declared %v",
				decl.DeclaredAtom, clause, inferredRelTypeExpr, declaredRelTypeExpr)
		}
	}
	return nil
}

// feasibleAlternatives returns feasible alternatives for a subgoal p(e1...eN).
// We consider a relation type expression (from decl), the given list of
// arguments, the type assignments to variable symbols, and a type context
// that collects type variables with associated type constrains.
// Returns list of alternatives and list of type contexts with new type constraints.
func (bc *BoundsAnalyzer) feasibleAlternatives(
	pred ast.PredicateSym, relTypeExpr ast.BaseTerm, args []ast.BaseTerm,
	varRanges map[ast.Variable]ast.BaseTerm,
	typeCtx map[ast.Variable]ast.BaseTerm) ([]ast.BaseTerm, []map[ast.Variable]ast.BaseTerm, error) {
	if pred.Symbol == symbols.ListMember.Symbol {
		tpe := boundOfArg(args[1], varRanges, bc.nameTrie)
		if tpe == ast.AnyBound {
			return []ast.BaseTerm{symbols.NewRelType(ast.AnyBound, symbols.NewListType(ast.AnyBound))}, nil, nil
		}
		if symbols.IsListTypeExpression(tpe) {
			elemTpe, err := symbols.ListTypeArg(tpe)
			if err != nil {
				return nil, nil, err
			}
			var bound ast.BaseTerm
			if v, ok := args[0].(ast.Variable); ok {
				if _, ok := varRanges[v]; ok {
					bound = varRanges[v]
				} else {
					bound = elemTpe // a new variable binding
				}
			} else {
				bound = boundOfArg(args[0], varRanges, bc.nameTrie)
			}
			// TODO: typing context must give lower bounds as well.
			meet := symbols.LowerBound(typeCtx, []ast.BaseTerm{bound, elemTpe})
			if !meet.Equals(symbols.EmptyType) {
				return []ast.BaseTerm{symbols.NewRelType(meet, tpe)}, []map[ast.Variable]ast.BaseTerm{typeCtx}, nil
			}
			return nil, nil, fmt.Errorf("pred %v on args %v cannot succeed var ranges %v", pred, args, varRanges)
		}
	}
	if pred.Symbol == symbols.MatchPrefix.Symbol {
		tpe := boundOfArg(args[0], varRanges, bc.nameTrie)
		prefix := args[1]
		meet := symbols.LowerBound(nil /*TODO*/, []ast.BaseTerm{tpe, prefix})
		if !meet.Equals(symbols.EmptyType) {
			return []ast.BaseTerm{symbols.NewRelType(meet, ast.NameBound)}, []map[ast.Variable]ast.BaseTerm{typeCtx}, nil
		}
		return nil, nil, fmt.Errorf("pred %v cannot succeed: type %v is incompatible with %v", pred, tpe, prefix)
	}

	if pred.Symbol == symbols.MatchEntry.Symbol {
		tpe := boundOfArg(args[0], varRanges, bc.nameTrie)
		if symbols.IsMapTypeExpression(tpe) {
			keyType, valTpe, err := symbols.MapTypeArgs(tpe)
			if err != nil {
				return nil, nil, err
			}
			var bound ast.BaseTerm
			if v, ok := args[1].(ast.Variable); ok {
				if _, ok := varRanges[v]; ok {
					bound = varRanges[v]
				} else {
					bound = tpe // a new variable binding
				}
			} else {
				bound = boundOfArg(args[1], varRanges, bc.nameTrie)
			}
			// TODO: typing context must give lower bounds as well.
			meet := symbols.LowerBound(typeCtx, []ast.BaseTerm{bound, keyType})
			if !meet.Equals(symbols.EmptyType) {
				var valbound ast.BaseTerm
				if v, ok := args[2].(ast.Variable); ok {
					if _, ok := varRanges[v]; ok {
						valbound = varRanges[v]
					} else {
						valbound = valTpe // a new variable binding
					}
				} else {
					valbound = boundOfArg(args[2], varRanges, bc.nameTrie)
				}

				valmeet := symbols.LowerBound(nil /*DOTO*/, []ast.BaseTerm{valbound, valTpe})
				if !valmeet.Equals(symbols.EmptyType) {
					return []ast.BaseTerm{symbols.NewRelType(tpe, keyType, valmeet)}, []map[ast.Variable]ast.BaseTerm{typeCtx}, nil
				}
				return nil, nil, fmt.Errorf("pred %v on args %v val type mismatch got %v want %v", pred, args, valbound, valTpe)
			}
			return nil, nil, fmt.Errorf("pred %v on args %v key type mismatch got %v want %v", pred, args, bound, keyType)
		}
	}
	if pred.Symbol == symbols.MatchField.Symbol {
		tpe := boundOfArg(args[0], varRanges, bc.nameTrie)
		if symbols.IsStructTypeExpression(tpe) {
			fieldTpe, err := symbols.StructTypeField(tpe, args[1].(ast.Constant))
			if err != nil {
				return nil, nil, err
			}

			var bound ast.BaseTerm
			if v, ok := args[2].(ast.Variable); ok {
				if _, ok := varRanges[v]; ok {
					bound = varRanges[v]
				} else {
					bound = fieldTpe // a new variable binding
				}
			} else {
				bound = boundOfArg(args[2], varRanges, bc.nameTrie)
			}

			// TODO: typing context must give lower bounds as well.
			meet := symbols.LowerBound(typeCtx, []ast.BaseTerm{bound, fieldTpe})
			if !meet.Equals(symbols.EmptyType) {
				return []ast.BaseTerm{symbols.NewRelType(ast.AnyBound, ast.NameBound, meet)}, []map[ast.Variable]ast.BaseTerm{typeCtx}, nil
			}
			return nil, nil, fmt.Errorf("pred %v on args %v cannot succeed var ranges %v", pred, args, varRanges)
		}
	}
	alternatives := symbols.RelTypeAlternatives(relTypeExpr)

	// Construct an "incoming" relation type from preceding.
	// E.g. for p(X,Y,Z), we produce a tuple Tx,Ty,Tz with types (bounds.)
	argBoundForAlternative := func(alternative ast.BaseTerm) (ast.BaseTerm, ast.SubstMap, error) {
		usedTypeVars := make(map[ast.Variable]bool, len(typeCtx))
		for v := range typeCtx {
			usedTypeVars[v] = true
		}

		argBound := make([]ast.BaseTerm, len(args))
		relTypeArgs, err := symbols.RelTypeArgs(alternative)
		if err != nil {
			return nil, nil, err
		}
		for i, arg := range args {
			v, isVar := arg.(ast.Variable)
			if !isVar {
				argBound[i] = boundOfArg(arg, varRanges, bc.nameTrie)
			}
			if _, ok := varRanges[v]; ok {
				argBound[i] = varRanges[v]
			} else {
				argBound[i] = relTypeArgs[i] // a new variable binding
			}
		}
		relTpe := symbols.NewRelType(argBound...)
		// skolemize
		typeVars := map[ast.Variable]bool{}
		ast.AddVars(relTpe, typeVars)
		var freshTypeSubst ast.SubstMap = map[ast.Variable]ast.BaseTerm{}
		for v := range typeVars {
			freshTypeSubst[v] = ast.FreshVariable(usedTypeVars)
		}
		return relTpe.ApplySubstBase(freshTypeSubst), freshTypeSubst, nil
	}

	var feasible []ast.BaseTerm
	var feasibleSubst []map[ast.Variable]ast.BaseTerm
	for _, alternative := range alternatives {
		// For a polymorphic type forall A1...An, TyExpr[A1...An],
		// we add a "skolem" type variables ?A1...?An, i.e.
		// variable symbols are used a placeholders for a fixed,
		// but unknown types that have various lower and upper bounds.
		argBound, newTypeSubst, err := argBoundForAlternative(alternative)
		if err != nil {
			return nil, nil, err
		}
		alternative = alternative.ApplySubstBase(newTypeSubst)
		newTypeCtx := map[ast.Variable]ast.BaseTerm{}
		// Copy old bounds.
		for v, tpe := range typeCtx {
			newTypeCtx[v] = tpe
		}
		// Add fresh vars.
		for _, v := range newTypeSubst {
			if typeVar, ok := v.(ast.Variable); ok {
				newTypeCtx[typeVar] = ast.AnyBound
			}
		}
		tpe := symbols.LowerBound(newTypeCtx, []ast.BaseTerm{argBound, alternative})
		if !tpe.Equals(symbols.EmptyType) {
			feasible = append(feasible, alternative)
			feasibleSubst = append(feasibleSubst, newTypeSubst)
		}
	}
	if len(feasible) == 0 {
		return nil, nil, fmt.Errorf("no feasible alternative reltypes %v args %v var ranges %v", relTypeExpr, args, varRanges)
	}
	return feasible, feasibleSubst, nil
}

// While checking a rule, we want to look up possible relation types.
// If we find several applicable ones, we return the feasible ones.
func (bc *BoundsAnalyzer) getOrInferRelTypes(
	pred ast.PredicateSym,
	args []ast.BaseTerm,
	varRanges map[ast.Variable]ast.BaseTerm,
	typeCtx map[ast.Variable]ast.BaseTerm) ([]ast.BaseTerm, error) {

	if relType, ok := bc.relTypeMap[pred]; ok {
		tpe, _, err := bc.feasibleAlternatives(pred, relType, args, varRanges, typeCtx)
		return tpe, err
	}

	if relType, ok := bc.inferred[pred]; ok {
		tpe, _, err := bc.feasibleAlternatives(pred, relType, args, varRanges, typeCtx)
		return tpe, err
	}

	if bc.visiting[pred] {
		// We are asking for pred in a recursive call. Use [any ... any]
		relType, err := symbols.RelTypeExprFromDecl(*bc.programInfo.Decls[pred])
		if err != nil {
			return nil, err
		}
		return []ast.BaseTerm{relType}, nil
	}

	bc.visiting[pred] = true
	defer delete(bc.visiting, pred)

	relTypeExpr, err := bc.inferRelTypes(pred)
	if err != nil {
		return nil, err
	}
	bc.inferred[pred] = relTypeExpr
	tpe, _, err := bc.feasibleAlternatives(pred, relTypeExpr, args, varRanges, nil /*TODO map[ast.Variable]bool*/)
	return tpe, err

}

// inferRelType infers a relation type from rules when no decl is available.
// inferRelType ensures that bc.inferred[pred] is populated with the inferred relation type.
func (bc *BoundsAnalyzer) inferRelTypes(pred ast.PredicateSym) (ast.BaseTerm, error) {
	if existing, ok := bc.relTypeMap[pred]; ok {
		return existing, nil
	}
	if existing, ok := bc.inferred[pred]; ok {
		return existing, nil
	}

	var alternatives []ast.BaseTerm
	if initialFactRelTypeExpr, ok := bc.initialFactMap[pred]; ok {
		alternatives = symbols.RelTypeAlternatives(initialFactRelTypeExpr)
	}

	clauses := bc.RulesMap[pred]
	for _, clause := range clauses {
		ic := newInferContextNoDecl(bc, &pred, &clause)
		relType, err := ic.inferRelTypesFromClause()
		if err != nil {
			return nil, err
		}
		for _, alternative := range symbols.RelTypeAlternatives(relType) {
			if !symbols.SetConforms(nil /*TODO*/, alternative, symbols.RelTypeFromAlternatives(alternatives)) {
				alternatives = append(alternatives, alternative)
			}
		}
	}
	bc.inferred[pred] = symbols.RelTypeFromAlternatives(alternatives)
	return bc.inferred[pred], nil
}

func checkFunApply(z ast.ApplyFn, fnTpe ast.BaseTerm, varRanges map[ast.Variable]ast.BaseTerm, nameTrie symbols.NameTrie) (ast.BaseTerm, error) {
	if fnTpe.Equals(symbols.EmptyType) {
		return nil, fmt.Errorf("type checking for %v not implemented", z)
	}
	argTypes, err := symbols.FunTypeArgs(fnTpe)
	if err != nil {
		return nil, fmt.Errorf("not a function type: %v", fnTpe)
	}
	if len(argTypes) != len(z.Args) {
		return nil, fmt.Errorf("wrong number of arguments: expected %d got %d", len(argTypes), len(z.Args))
	}
	actualTpes := make([]ast.BaseTerm, len(argTypes))
	for i, arg := range z.Args {
		actualTpes[i] = boundOfArg(arg, varRanges, nameTrie)
	}
	subst, err := unionfind.UnifyTypeExpr(actualTpes, argTypes)
	if err != nil {
		return nil, fmt.Errorf("could not unify %v and %v: %v", actualTpes, argTypes, err)
	}
	res, err := symbols.FunTypeResult(fnTpe)
	if err != nil {
		return nil, fmt.Errorf("not a function type: %v", fnTpe)
	}
	return res.ApplySubstBase(ast.SubstMap(subst)), nil
}

func boundOfArg(x ast.BaseTerm, varRanges map[ast.Variable]ast.BaseTerm, nameTrie symbols.NameTrie) ast.BaseTerm {
	switch z := x.(type) {
	case ast.Variable:
		if bound, ok := varRanges[z]; ok {
			return bound
		}
		return ast.AnyBound

	case ast.Constant:
		switch z.Type {
		case ast.NumberType:
			return ast.NumberBound
		case ast.Float64Type:
			return ast.Float64Bound
		case ast.StringType:
			return ast.StringBound
		case ast.TimeType:
			return ast.TimeBound
		case ast.DurationType:
			return ast.DurationBound
		case ast.NameType:
			// Find a name prefix type, or fall back to /name.
			return nameTrie.PrefixName(z.Symbol)
		case ast.ListShape:
			var args []ast.BaseTerm
			z.ListValues(func(arg ast.Constant) error {
				args = append(args, arg)
				return nil
			}, func() error {
				return nil
			})

			return boundOfArg(ast.ApplyFn{symbols.List, args}, varRanges, nameTrie)
		case ast.MapShape:
			var args []ast.BaseTerm
			z.MapValues(func(keyArg, valArg ast.Constant) error {
				args = append(args, keyArg)
				args = append(args, valArg)
				return nil
			}, func() error {
				return nil
			})
			return boundOfArg(ast.ApplyFn{symbols.Map, args}, varRanges, nameTrie)
		case ast.StructShape:
			var args []ast.BaseTerm
			z.StructValues(func(fieldArg, valArg ast.Constant) error {
				args = append(args, fieldArg)
				args = append(args, valArg)
				return nil
			}, func() error {
				return nil
			})
			return boundOfArg(ast.ApplyFn{symbols.Struct, args}, varRanges, nameTrie)

		default:
			return ast.AnyBound // This cannot happen
		}

	case ast.ApplyFn:
		// TODO: Less special cases. Add support repeated arguments to function types.
		switch z.Function.Symbol {
		case symbols.List.Symbol:
			if len(z.Args) == 0 {
				return ast.ApplyFn{symbols.ListType, []ast.BaseTerm{ast.BotBound}}
			}
			var argTypes []ast.BaseTerm
			for _, arg := range z.Args {
				argTypes = append(argTypes, boundOfArg(arg, varRanges, nameTrie))
			}
			return symbols.NewListType(symbols.UpperBound(nil /*TODO*/, argTypes))

		case symbols.Map.Symbol:
			var keyTpes []ast.BaseTerm
			var valTpes []ast.BaseTerm
			for i := 0; i < len(z.Args); i++ {
				keyTpes = append(keyTpes, boundOfArg(z.Args[i], varRanges, nameTrie))
				i++
				valTpes = append(valTpes, boundOfArg(z.Args[i], varRanges, nameTrie))
			}
			return symbols.NewMapType(symbols.UpperBound(nil /*TODO*/, keyTpes), symbols.UpperBound(nil /*TODO*/, valTpes))

		case symbols.Struct.Symbol:
			var fields []ast.BaseTerm
			for i := 0; i < len(z.Args); i++ {
				fields = append(fields, z.Args[i])
				i++
				fields = append(fields, boundOfArg(z.Args[i], varRanges, nameTrie))
			}

			return symbols.NewStructType(fields...)

		case symbols.StructGet.Symbol:
			structTpe := boundOfArg(z.Args[0], varRanges, nameTrie)
			if !symbols.IsStructTypeExpression(structTpe) {
				return symbols.EmptyType
			}
			field, ok := z.Args[1].(ast.Constant)
			if !ok || field.Type != ast.NameType {
				return symbols.EmptyType
			}
			fieldTpe, err := symbols.StructTypeField(structTpe, field)
			if err != nil {
				return symbols.EmptyType
			}
			return fieldTpe

		case symbols.Tuple.Symbol:
			var argTypes []ast.BaseTerm
			for _, arg := range z.Args {
				argTypes = append(argTypes, boundOfArg(arg, varRanges, nameTrie))
			}
			return symbols.NewTupleType(argTypes...)

		case symbols.StringConcatenate.Symbol:
			return ast.StringBound

		case symbols.Plus.Symbol:
			fallthrough
		case symbols.Minus.Symbol:
			fallthrough
		case symbols.Mult.Symbol:
			fallthrough
		case symbols.Div.Symbol:
			for _, arg := range z.Args {
				if ast.NumberBound != boundOfArg(arg, varRanges, nameTrie) {
					return symbols.EmptyType
				}
			}
			return ast.NumberBound
		case symbols.FloatDiv.Symbol:
			for _, arg := range z.Args {
				if ast.Float64Bound != boundOfArg(arg, varRanges, nameTrie) {
					return symbols.EmptyType
				}
			}
			return ast.Float64Bound
		}

		if fnTpe, ok := builtin.GetBuiltinFunctionType(z.Function); ok {
			res, err := checkFunApply(z, fnTpe, varRanges, nameTrie)
			if err != nil {
				return symbols.EmptyType // TODO: return error
			}
			return res
		}
		return ast.AnyBound

	default:
		return ast.AnyBound
	}
}

func typeOfFn(x ast.ApplyFn, varRanges map[ast.Variable]ast.BaseTerm, nameTrie symbols.NameTrie) ast.BaseTerm {
	switch x.Function.Symbol {
	case symbols.Max.Symbol:
		fallthrough
	case symbols.Min.Symbol:
		fallthrough
	case symbols.Count.Symbol:
		fallthrough
	case symbols.Div.Symbol:
		fallthrough
	case symbols.FloatDiv.Symbol:
		fallthrough
	case symbols.Sum.Symbol:
		fallthrough
	case symbols.Plus.Symbol:
		fallthrough
	case symbols.Minus.Symbol:
		fallthrough
	case symbols.Mult.Symbol:
		return ast.NumberBound
	case symbols.Collect.Symbol:
		fallthrough
	case symbols.CollectDistinct.Symbol:
		if len(x.Args) == 1 {
			if v, ok := x.Args[0].(ast.Variable); ok {
				return ast.ApplyFn{symbols.ListType, []ast.BaseTerm{varRanges[v]}}
			}
		}
		elemTpe := boundOfArg(x.Args[0], varRanges, nameTrie)
		return ast.ApplyFn{symbols.ListType, []ast.BaseTerm{elemTpe}}
	}
	fnTpe, ok := builtin.GetBuiltinFunctionType(x.Function)
	if !ok {
		return ast.AnyBound // TODO: return error
	}
	res, err := checkFunApply(x, fnTpe, varRanges, nameTrie)
	if err != nil {
		return ast.AnyBound
	}
	return res
}
