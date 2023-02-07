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

// Package analysis contains methods that check whether each datalog clause is valid, and
// whether a set of valid clauses forms a valid datalog program.
package analysis

import (
	"fmt"
	"sort"
	"strings"

	"github.com/google/mangle/ast"
	"github.com/google/mangle/builtin"
	"github.com/google/mangle/functional"
	"github.com/google/mangle/packages"
	"github.com/google/mangle/parse"
	"github.com/google/mangle/symbols"
	"github.com/google/mangle/unionfind"
)

// BoundsCheckingMode represents a mode for bounds checking.
type BoundsCheckingMode int

const (
	// NoBoundsChecking means there is no bounds checking of any kind.
	NoBoundsChecking BoundsCheckingMode = iota
	// LogBoundsMismatch means we log mismatch.
	LogBoundsMismatch
	// ErrorForBoundsMismatch means bounds mismatch leads to error.
	ErrorForBoundsMismatch
)

// ProgramInfo represents the result of program analysis.
type ProgramInfo struct {
	// Extensional predicate symbols; those that do not appear in the head of a clause with a body.
	EdbPredicates map[ast.PredicateSym]struct{}
	// Intensional predicate symbols; those that do appear in the head of a clause with a body.
	IdbPredicates map[ast.PredicateSym]struct{}
	// Heads of rules without a body.
	InitialFacts []ast.Atom
	// Rules that have a body.
	Rules []ast.Clause
	// Desugared declarations for all predicates, possibly synthetic.
	Decls map[ast.PredicateSym]*ast.Decl
}

// Analyzer is a struct providing built-in predicates and functions for name analysis.
type Analyzer struct {
	// Predicates that are provided by this datalog implementation.
	builtInPredicates map[ast.PredicateSym]struct{}
	// Functions (only used in transforms) provided by this datalog implementation.
	builtInFunctions map[ast.FunctionSym]struct{}
	// Predicates that are defined previously and can safely be referenced by rules.
	knownPredicates map[ast.PredicateSym]ast.Decl
	// Declaration of predicates to be analyzed. Must not overlap with knownPredicates.
	decl map[ast.PredicateSym]ast.Decl
	// Whether mismatch in bounds leads to an error
	boundsCheckingMode BoundsCheckingMode
}

// BoundsAnalyzer to infer and check bounds.
type BoundsAnalyzer struct {
	programInfo *ProgramInfo
	// Trie of name constants from declarations different from /any,/name,/number,/string.
	nameTrie nametrie
	RulesMap map[ast.PredicateSym][]ast.Clause
	// maps `foo`` to either RelType[...] or fn:Union(RelType[...]...RelType[...])
	relTypeMap map[ast.PredicateSym]ast.BaseTerm
	// maps `foo`` to either RelType[...] or fn:Union(RelType[...]...RelType[...])
	initialFactMap map[ast.PredicateSym]ast.BaseTerm
	// maps `foo`` to either RelType[...] or fn:Union(RelType[...]...RelType[...])
	inferred map[ast.PredicateSym]ast.BaseTerm
	visiting map[ast.PredicateSym]bool
}

// AnalyzeOneUnit is a convenience method to analyze a program consisting of a single source unit.
func AnalyzeOneUnit(unit parse.SourceUnit, knownPredicates map[ast.PredicateSym]ast.Decl) (*ProgramInfo, error) {
	return Analyze([]parse.SourceUnit{unit}, knownPredicates)
}

// Analyze identifies the extensional and intensional predicates of a program and checks every rule.
func Analyze(program []parse.SourceUnit, knownPredicates map[ast.PredicateSym]ast.Decl) (*ProgramInfo, error) {
	return AnalyzeAndCheckBounds(program, knownPredicates, NoBoundsChecking)
}

// AnalyzeAndCheckBounds checks every rule, including bounds.
func AnalyzeAndCheckBounds(program []parse.SourceUnit, knownPredicates map[ast.PredicateSym]ast.Decl, boundsChecking BoundsCheckingMode) (*ProgramInfo, error) {
	pkgs := map[string]*packages.Package{}
	var clauses []ast.Clause
	var decls []ast.Decl
	for _, unit := range program {
		p, err := packages.Extract(unit)
		if err != nil {
			return nil, err
		}
		pkg, ok := pkgs[p.Name]
		if ok {
			pkg.Merge(p)
		} else {
			pkgs[p.Name] = &p
		}
	}
	for _, p := range pkgs {
		ds, err := p.Decls()
		if err != nil {
			return nil, err
		}
		decls = append(decls, ds...)
		cs, err := p.Clauses()
		if err != nil {
			return nil, err
		}
		clauses = append(clauses, cs...)
	}

	analyzer, err := New(knownPredicates, decls, boundsChecking)
	if err != nil {
		return nil, err
	}
	if err := analyzer.EnsureDecl(clauses); err != nil {
		return nil, err
	}
	return analyzer.Analyze(clauses)
}

func byName(decls map[ast.PredicateSym]ast.Decl) map[string]ast.Decl {
	byName := make(map[string]ast.Decl, len(decls))
	for sym, decl := range decls {
		byName[sym.Symbol] = decl
	}
	return byName
}

// New creates a new analyzer that uses builtin predicates and functions.
func New(knownPredicates map[ast.PredicateSym]ast.Decl, decls []ast.Decl, boundsChecking BoundsCheckingMode) (*Analyzer, error) {
	knownByName := byName(knownPredicates)
	declMap := make(map[ast.PredicateSym]ast.Decl)
	for _, decl := range decls {
		pred := decl.DeclaredAtom.Predicate
		if pred == symbols.Package || pred == symbols.Use {
			continue
		}
		if knownDecl, ok := knownByName[pred.Symbol]; ok {
			if !knownDecl.IsSynthetic() {
				return nil, fmt.Errorf("cannot redeclare %v, previous Decl %v", decl, knownDecl)
			}
			// We can override a synthetic decl, but arity should still match.
			if knownDecl.DeclaredAtom.Predicate.Arity != pred.Arity {
				return nil, fmt.Errorf("declared arity %v conflicts with known decl %v", decl, knownDecl)
			}
			// Override the synthetic decl with the one we were requested to use.
			delete(knownPredicates, pred)
		}
		declMap[pred] = decl
	}
	return &Analyzer{builtin.Predicates, builtin.Functions, knownPredicates, declMap, boundsChecking}, nil
}

func isExtensional(decl ast.Decl) bool {
	for _, a := range decl.Descr {
		if a.Predicate.Symbol == "extensional" {
			return true
		}
	}
	return false
}

// EnsureDecl will ensure there is a declaration for each head of a rule,
// creating one if necessary.
func (a *Analyzer) EnsureDecl(clauses []ast.Clause) error {
	knownByName := byName(a.knownPredicates)
	declByName := byName(a.decl)
	for _, c := range clauses {
		pred := c.Head.Predicate
		name := pred.Symbol
		// Check that the name was not defined previously (in a separate source).
		// We may permit "distributing" definitions over source files later.
		if decl, ok := a.knownPredicates[pred]; ok {
			if isExtensional(decl) && len(c.Premises) == 0 {
				continue
			}
			return fmt.Errorf("predicate %v was defined previously %v", decl.DeclaredAtom.Predicate, decl)
		}
		if decl, ok := knownByName[name]; ok { // different arity
			return fmt.Errorf(
				"predicate %v was defined previously", decl.DeclaredAtom.Predicate)
		}
		if _, ok := a.decl[pred]; ok {
			continue
		}
		// Check that the name was not defined in the same source with a different arity.
		if decl, ok := declByName[name]; ok {
			return fmt.Errorf("%v does not match arity of %v", c.Head, decl.DeclaredAtom)
		}
		var (
			synthDecl ast.Decl
			err       error
		)
		if c.Premises != nil {
			synthDecl, err = ast.NewSyntheticDecl(c.Head) // preserve variable names.
		} else {
			synthDecl = ast.NewSyntheticDeclFromSym(pred)
		}
		if err != nil {
			return err
		}
		a.decl[pred] = synthDecl
		declByName[pred.Symbol] = a.decl[pred]
	}
	return nil
}

// Analyze identifies the extensional and intensional predicates of a program, checks every rule and that
// all references to built-in predicates and functions used in transforms are valid.
func (a *Analyzer) Analyze(program []ast.Clause) (*ProgramInfo, error) {
	globalDecls := make(map[ast.PredicateSym]ast.Decl)
	for p, d := range a.knownPredicates {
		globalDecls[p] = d
	}
	for p, d := range a.decl {
		globalDecls[p] = d
	}
	desugaredDecls, err := symbols.CheckAndDesugar(globalDecls)
	if err != nil {
		return nil, err
	}
	edbSymbols := make(map[ast.PredicateSym]struct{})
	idbSymbols := make(map[ast.PredicateSym]struct{})
	var initialFacts []ast.Atom
	var rules []ast.Clause
	rulesMap := make(map[ast.PredicateSym][]ast.Clause)
	for _, clause := range program {
		// Check each rule.
		if err := a.CheckRule(clause); err != nil {
			return nil, err
		}
		// Is it an extensional or intensional predicate?
		if clause.Premises == nil {
			head, err := functional.EvalAtom(clause.Head, ast.ConstSubstList{})
			if err != nil {
				return nil, err
			}
			initialFacts = append(initialFacts, head)
			edbSymbols[clause.Head.Predicate] = struct{}{}
		} else {
			rules = append(rules, clause)
			rulesMap[clause.Head.Predicate] = append(rulesMap[clause.Head.Predicate], clause)
			idbSymbols[clause.Head.Predicate] = struct{}{}
			for _, premise := range clause.Premises {
				switch p := premise.(type) {
				case ast.Atom:
					if !p.Predicate.IsBuiltin() {
						edbSymbols[p.Predicate] = struct{}{}
					}
				case ast.NegAtom:
					edbSymbols[p.Atom.Predicate] = struct{}{}
				}
			}
		}
	}
	// If it is "both" (has rules with premises and without premises),
	// treat is an intensional rule since we will have to evaluate.
	for s := range idbSymbols {
		delete(edbSymbols, s)
	}

	programInfo := ProgramInfo{edbSymbols, idbSymbols, initialFacts, rules, desugaredDecls}
	if a.boundsCheckingMode != NoBoundsChecking {
		nameTrie, err := collectNamePrefixes(programInfo)
		if err != nil {
			return nil, err
		}
		bc, err := newBoundsAnalyzer(&programInfo, nameTrie, initialFacts, rulesMap)
		if err != nil {
			return nil, err
		}
		if err := bc.BoundsCheck(); err != nil {
			if a.boundsCheckingMode == ErrorForBoundsMismatch {
				return nil, err
			}
		}
	}

	return &programInfo, nil
}

func addNamesToTrie(b ast.BaseTerm, nameTrie nametrie) {
	if b == ast.AnyBound || b == ast.NameBound || b == ast.NumberBound || b == ast.StringBound {
		return
	}
	if apply, ok := b.(ast.ApplyFn); ok {
		for _, arg := range apply.Args {
			addNamesToTrie(arg, nameTrie)
		}
		return
	}
	c, ok := b.(ast.Constant)
	if !ok || c.Type != ast.NameType {
		// At this point, string constants "predicate" have been replaced
		// with the appropriate type expression. So this should not happen.
		return
	}
	parts := strings.Split(c.Symbol, "/")
	nameTrie.Add(parts[1:])
}

// Extracts name constants from declarations. We build a trie of names used as type expressions.
// so we can later map a name to the the most precise (longest prefix). Note that the trie for
// {"/foo", "/foo/bar"} is different from {"/foo/bar"}: the former would map a constant "/foo/baz"
// to the type "/foo", whereas the latter would map it to type "/name".
func collectNamePrefixes(programInfo ProgramInfo) (nametrie, error) {
	nameTrie := newNameTrie()

	for _, d := range programInfo.Decls {
		for _, bs := range d.Bounds {
			for _, b := range bs.Bounds {
				addNamesToTrie(b, nameTrie)
			}
		}
	}
	return nameTrie, nil
}

func newBoundsAnalyzer(programInfo *ProgramInfo, nameTrie nametrie, initialFacts []ast.Atom, rulesMap map[ast.PredicateSym][]ast.Clause) (*BoundsAnalyzer, error) {
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
		if decl.IsSynthetic() && !isExtensional(*decl) {
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

// CheckRule checks that every variable is either "bound" or defined by a transform.
// A variable in a rule is bound when it appears in a positive atom, or is unified
// (via an equality) with a constant or another variable that is bound.
func (a *Analyzer) CheckRule(clause ast.Clause) error {
	clause = clause.ReplaceWildcards()
	var (
		boundVars = make(map[ast.Variable]bool)
		headVars  = make(map[ast.Variable]bool)
		seenVars  = make(map[ast.Variable]bool)
	)
	ast.AddVars(clause.Head, headVars)
	ast.AddVars(clause.Head, seenVars)
	uf := unionfind.New()
	if clause.Premises != nil {
		for _, premise := range clause.Premises {
			ast.AddVars(premise, seenVars)
			switch p := premise.(type) {
			case ast.Atom:
				if err := checkAtom(p); err != nil {
					return err
				}
				if !p.Predicate.IsBuiltin() {
					ast.AddVars(p, boundVars)
				}
				// Since evaluation will proceed from left to right, we need to ensure that the variables
				// are already assigned values by some earlier subgoal. A more sophisticated way would be
				// to rewrite the rule, automatically moving built-in predicate subgoals to a position in
				// where we are sure that variables have been assigned values.
				var builtinVars = make(map[ast.Variable]bool)
				ast.AddVars(p, builtinVars)
				if p.Predicate == symbols.MatchPair || p.Predicate == symbols.MatchCons {
					if fstVar, fstOk := p.Args[1].(ast.Variable); fstOk {
						boundVars[fstVar] = true

						if scrutinee, ok := p.Args[0].(ast.Variable); ok && scrutinee == fstVar {
							return fmt.Errorf("a variable that is matched cannot be used for binding %v", p)
						}
					} else if _, constOk := p.Args[1].(ast.Constant); !constOk {
						return fmt.Errorf("expected variable or constant as second argument to %v", p)
					}

					if sndVar, sndOk := p.Args[2].(ast.Variable); sndOk {
						boundVars[sndVar] = true

						if scrutinee, ok := p.Args[0].(ast.Variable); ok && scrutinee == sndVar {
							return fmt.Errorf("a variable that is matched cannot be used for binding %v", p)
						}
					} else if _, constOk := p.Args[2].(ast.Constant); !constOk {
						return fmt.Errorf("expected variable or constant as second argument to %v", p)
					}
				}
				if p.Predicate == symbols.MatchEntry || p.Predicate == symbols.MatchField {
					if _, keyOk := p.Args[1].(ast.Constant); !keyOk {
						return fmt.Errorf("expected constant as second argument to %v", p)
					}
					if valVar, valOk := p.Args[2].(ast.Variable); valOk {
						boundVars[valVar] = true

						if scrutinee, ok := p.Args[0].(ast.Variable); ok && scrutinee == valVar {
							return fmt.Errorf("a variable that is matched cannot be used for binding %v", p)
						}
					} else if _, constOk := p.Args[2].(ast.Constant); !constOk {
						return fmt.Errorf("expected variable or constant as third argument to %v", p)
					}
				}

				if p.Predicate == symbols.ListMember { // :list:member(Member, List)
					if memberVar, memberOk := p.Args[0].(ast.Variable); memberOk {
						boundVars[memberVar] = true

						if listArg, ok := p.Args[1].(ast.Variable); ok && listArg == memberVar {
							return fmt.Errorf("a variable whose value is expanded cannot be used for binding %v", p)
						}
					} else if _, constOk := p.Args[0].(ast.Constant); !constOk {
						return fmt.Errorf("expected variable or constant as 2nd argument to %v", p)
					}
				}

				for v := range builtinVars {
					if !boundVars[v] {
						return fmt.Errorf("variable %v in %v will not have a value yet; move the subgoal to the right", v, p)
					}
				}
			case ast.Eq:
				if _, isconst := p.Left.(ast.Constant); isconst {
					if v, isvar := p.Right.(ast.Variable); isvar {
						boundVars[v] = true
						break
					}
				}
				if _, isconst := p.Right.(ast.Constant); isconst {
					if v, isvar := p.Left.(ast.Variable); isvar {
						boundVars[v] = true
						break
					}
				}
				if _, isapply := p.Left.(ast.ApplyFn); isapply {
					vars := make(map[ast.Variable]bool)
					ast.AddVars(p.Left, vars)
					for v := range vars {
						if !boundVars[v] {
							return fmt.Errorf("variable %v in apply expression %v not bound", v, p)
						}
					}
					if v, isvar := p.Right.(ast.Variable); isvar {
						boundVars[v] = true
						break
					}
				}
				if _, isapply := p.Right.(ast.ApplyFn); isapply {
					vars := make(map[ast.Variable]bool)
					ast.AddVars(p.Right, vars)
					for v := range vars {
						if !boundVars[v] {
							return fmt.Errorf("variable %v in apply expression %v not bound", v, p)
						}
					}
					if v, isvar := p.Left.(ast.Variable); isvar {
						boundVars[v] = true
						break
					}
				}

				if _, l := p.Left.(ast.Variable); l {
					if _, r := p.Right.(ast.Variable); r {
						var err error
						uf, err = unionfind.UnifyTermsExtend([]ast.BaseTerm{p.Left}, []ast.BaseTerm{p.Right}, uf)
						if err != nil {
							return err
						}
					}
				}
			}
		}
	}

	var (
		transformVarDefs = make(map[ast.Variable]bool)
		transformVarUses = make(map[ast.Variable]bool)
	)
	// If there is no transform, these maps remain empty.
	addTransformVars(clause.Transform, transformVarDefs, transformVarUses)

	// It is ok to write "let _ = fn:foo(...)" but not "fn:foo(_)".
	delete(transformVarDefs, ast.Variable{"_"})
	for v := range transformVarDefs {
		if boundVars[v] {
			return fmt.Errorf("the transform of clause %v redefines variable %v from rule body", clause, v)
		}
	}

	// Every variable that we saw in head or body has to be bound somewhere.
	for v := range seenVars {
		if headVars[v] && transformVarDefs[v] {
			// Head variable is defined in transform.
			continue
		}
		if boundVars[v] {
			// Head variable is range-restricted by positive atom or constant.
			continue
		}
		if x := uf.Get(v); x != nil {
			// Head variable is unified.
			if _, isconst := x.(ast.Constant); isconst {
				continue
			}
			if u, isvar := x.(ast.Variable); isvar {
				if _, ok := boundVars[u]; ok {
					continue
				}
			}
		}
		return fmt.Errorf("variable %v is not bound in %v", v, clause)
	}

	// Every variable use that we saw in transform has to be bound somewhere.
	for v := range transformVarUses {
		// In a transform, we can refer to any variable that appears in the rule.
		if seenVars[v] {
			continue
		}
		return fmt.Errorf("variable %v used in transform %v does not appear in clause %v", v, *clause.Transform, clause)
	}

	if groupByVars, ok := isGroupByTransform(clause); ok {
		groupByStmt := clause.Transform.Statements[0]
		if groupByStmt.Var != nil {
			return fmt.Errorf("do-transforms cannot have a variable %v", clause)
		}
		for _, stmt := range clause.Transform.Statements[1:] {
			if stmt.Var == nil {
				return fmt.Errorf("all statements following group by have to be let-statements %v", clause)
			}
			if !builtin.IsReducerFunction(stmt.Fn.Function) {
				// We could actually permit non-reducer applications if they only involve variables from the group_by key. Left for later.
				return fmt.Errorf("for now, every statement following group has to be a reducer application, found %v in %v", stmt.Fn, clause)
			}
		}
		if len(groupByVars) != len(groupByStmt.Fn.Args) {
			return fmt.Errorf("each argument of group_by must be a distinct variable, got: %v", groupByStmt)
		}
		// All head variables have to either be part of group_by key or appear in a reducer application.
		for v := range headVars {
			if groupByVars[v] {
				continue
			}
			if transformVarDefs[v] {
				continue
			}
			return fmt.Errorf("head variable %v is neither part of group_by nor aggregated: %v", v, clause.Transform)
		}
	}
	if isLetTransform(clause) {
		for _, stmt := range clause.Transform.Statements[1:] {
			if stmt.Var == nil {
				return fmt.Errorf("all statements in a let transform have to be let-statements %v", clause)
			}
			if _, ok := builtin.ReducerFunctions[stmt.Fn.Function]; ok {
				return fmt.Errorf("reducer applications %v not allowed in a let-transform %v", stmt.Fn, clause)
			}
		}
	}

	// Check that rules only reference predicates that are defined.
	if err := a.checkPredicates(clause); err != nil {
		return err
	}
	// Check that the RHS predicates are visible from the package of this clause.
	if err := a.checkVisibility(clause); err != nil {
		return err
	}
	// Check that transforms are defined over a relation.
	if clause.Transform != nil && clause.Premises == nil {
		return fmt.Errorf("cannot have a transform without a body %v", clause)
	}
	// Check that the arity (TODO: types) of function application works out.
	if err := a.checkFunctions(clause); err != nil {
		return err
	}

	return nil
}

func isLetTransform(clause ast.Clause) bool {
	return clause.Transform != nil && clause.Transform.Statements[0].Var != nil
}

func isGroupByTransform(clause ast.Clause) (map[ast.Variable]bool, bool) {
	if clause.Transform == nil {
		return nil, false
	}
	stmt := clause.Transform.Statements[0]
	if stmt.Fn.Function.Symbol != symbols.GroupBy.Symbol {
		return nil, false
	}
	vars := make(map[ast.Variable]bool)
	for _, arg := range stmt.Fn.Args {
		if v, ok := arg.(ast.Variable); ok {
			vars[v] = true
		}
	}
	return vars, true
}

func addTransformVars(transform *ast.Transform, vardefs map[ast.Variable]bool, varuse map[ast.Variable]bool) {
	if transform == nil {
		return
	}
	for _, transformStmt := range transform.Statements {
		if transformStmt.Var != nil {
			vardefs[*transformStmt.Var] = true
		}
		for _, baseTerm := range transformStmt.Fn.Args {
			ast.AddVars(baseTerm, varuse)
		}
	}
}

func checkAtom(atom ast.Atom) error {
	if atom.Predicate.Arity != len(atom.Args) {
		return fmt.Errorf("Arity mismatch: %s expects %d arguments but has %d in %v", atom.Predicate.Symbol, atom.Predicate.Arity, len(atom.Args), atom)
	}
	return nil
}

func (a *Analyzer) check(c func(sym ast.PredicateSym) error, clause ast.Clause) error {
	for _, p := range clause.Premises {
		switch p := p.(type) {
		case ast.Atom:
			if err := c(p.Predicate); err != nil {
				return err
			}
		case ast.NegAtom:
			if err := c(p.Atom.Predicate); err != nil {
				return err
			}
		default:
			continue
		}
	}
	return nil
}

func (a *Analyzer) checkPredicates(clause ast.Clause) error {
	return a.check(func(sym ast.PredicateSym) error {
		if _, ok := a.decl[sym]; ok {
			return nil
		}
		if _, ok := a.knownPredicates[sym]; ok {
			return nil
		}
		if _, ok := a.builtInPredicates[sym]; ok {
			return nil
		}
		return fmt.Errorf("clause %v could not find predicate %v", clause, sym)
	}, clause)
}

func (a *Analyzer) checkVisibility(clause ast.Clause) error {
	var pkg string
	symbol := clause.Head.Predicate.Symbol
	if lastDot := strings.LastIndex(symbol, "."); lastDot != -1 {
		pkg = symbol[:lastDot]
	} else {
		pkg = ""
	}
	return a.check(func(sym ast.PredicateSym) error {
		d, ok := a.decl[sym]
		if !ok {
			// TODO: Invert default visibility.
			// No decl found. Assume public visibility.
			return nil
		}
		// Predicates defined in the same package are visible.
		if pkg == d.PackageID() {
			return nil
		}
		if !d.Visible() {
			return fmt.Errorf("predicate %q is not public", sym)
		}
		return nil
	}, clause)
}

func (a *Analyzer) checkFunctions(clause ast.Clause) error {
	if clause.Transform == nil {
		return nil
	}
	for _, stmt := range clause.Transform.Statements {
		sym := stmt.Fn.Function
		if _, ok := a.builtInFunctions[ast.FunctionSym{sym.Symbol, -1}]; ok {
			continue
		}
		if _, ok := a.builtInFunctions[sym]; ok {
			if sym.Arity == len(stmt.Fn.Args) {
				continue
			}
			return fmt.Errorf("function %v expects %d arguments, provided: %v", sym.Symbol, sym.Arity, stmt.Fn.Args)
		}
		return fmt.Errorf("clause %v could not find function %v", clause, sym)
	}
	return nil
}

// BoundsCheck checks whether the rules respect the bounds.
func (bc *BoundsAnalyzer) BoundsCheck() error {
	var preds []ast.PredicateSym

	for pred := range bc.programInfo.IdbPredicates {
		preds = append(preds, pred)
	}
	for pred := range bc.initialFactMap {
		preds = append(preds, pred)
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
			if !symbols.TypeConforms(inferred, declaredRelTypeExpr) {
				return fmt.Errorf("found unit clause with %v that does not conform to any decl %v", inferred, declaredRelTypeExpr)
			}
		}
	}

	for _, clause := range clauses {
		inferredRelTypeExpr, err := bc.inferRelTypesFromClause(clause)
		if err != nil {
			return err
		}
		if !symbols.TypeConforms(inferredRelTypeExpr, declaredRelTypeExpr) {
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

// feasibleAlternatives returns those alternatives from a relation type expression that
// make sense for a given list of arguments and type assignments.
func (bc *BoundsAnalyzer) feasibleAlternatives(
	pred ast.PredicateSym, relTypeExpr ast.BaseTerm, args []ast.BaseTerm,
	varRanges map[ast.Variable]ast.BaseTerm) ([]ast.BaseTerm, error) {

	if pred.Symbol == symbols.ListMember.Symbol {
		tpe := boundOfArg(args[1], varRanges, bc.nameTrie)
		if symbols.IsListTypeExpression(tpe) {
			elemTpe, err := symbols.ListTypeArg(tpe)
			if err != nil {
				return nil, err
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
			meet := symbols.LowerBound([]ast.BaseTerm{bound, elemTpe})
			if !meet.Equals(symbols.EmptyType) {
				return []ast.BaseTerm{symbols.NewRelType(meet, tpe)}, nil
			}
			return nil, fmt.Errorf("pred %v on args %v cannot succeed var ranges %v", pred, args, varRanges)
		}
	}
	if pred.Symbol == symbols.MatchEntry.Symbol {
		tpe := boundOfArg(args[0], varRanges, bc.nameTrie)
		if symbols.IsMapTypeExpression(tpe) {
			keyType, valTpe, err := symbols.MapTypeArgs(tpe)
			if err != nil {
				return nil, err
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

			meet := symbols.LowerBound([]ast.BaseTerm{bound, keyType})
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

				valmeet := symbols.LowerBound([]ast.BaseTerm{valbound, valTpe})
				if !valmeet.Equals(symbols.EmptyType) {
					return []ast.BaseTerm{symbols.NewRelType(tpe, keyType, valmeet)}, nil
				}
				return nil, fmt.Errorf("pred %v on args %v val type mismatch got %v want %v", pred, args, valbound, valTpe)
			}
			return nil, fmt.Errorf("pred %v on args %v key type mismatch got %v want %v", pred, args, bound, keyType)
		}
	}
	if pred.Symbol == symbols.MatchField.Symbol {
		tpe := boundOfArg(args[0], varRanges, bc.nameTrie)
		if symbols.IsStructTypeExpression(tpe) {
			fieldTpe, err := symbols.StructTypeField(tpe, args[1].(ast.Constant))
			if err != nil {
				return nil, err
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

			meet := symbols.LowerBound([]ast.BaseTerm{bound, fieldTpe})
			if !meet.Equals(symbols.EmptyType) {
				return []ast.BaseTerm{symbols.NewRelType(ast.AnyBound, ast.NameBound, meet)}, nil
			}
			return nil, fmt.Errorf("pred %v on args %v cannot succeed var ranges %v", pred, args, varRanges)
		}
	}
	alternatives := symbols.RelTypeAlternatives(relTypeExpr)

	// Construct a relation type from what we know.
	argBoundForAlternative := func(alternative ast.BaseTerm) (ast.BaseTerm, error) {
		argBound := make([]ast.BaseTerm, len(args))
		relTypeArgs, err := symbols.RelTypeArgs(alternative)
		if err != nil {
			return nil, err
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
		return symbols.NewRelType(argBound...), nil
	}

	var feasible []ast.BaseTerm
	for _, alternative := range alternatives {
		argBound, err := argBoundForAlternative(alternative)
		if err != nil {
			return nil, err
		}
		tpe := symbols.LowerBound([]ast.BaseTerm{argBound, alternative})
		if !tpe.Equals(symbols.EmptyType) {
			feasible = append(feasible, alternative)
		}
	}
	if len(feasible) == 0 {
		return nil, fmt.Errorf("no feasible alternative reltypes %v args %v var ranges %v", relTypeExpr, args, varRanges)
	}
	return feasible, nil
}

// While checking a rule, we want to look up possible relation types.
// If we find several applicable ones, we return the feasible ones.
func (bc *BoundsAnalyzer) getOrInferRelTypes(
	pred ast.PredicateSym,
	args []ast.BaseTerm,
	varRanges map[ast.Variable]ast.BaseTerm) ([]ast.BaseTerm, error) {

	if relType, ok := bc.relTypeMap[pred]; ok {
		return bc.feasibleAlternatives(pred, relType, args, varRanges)
	}

	if relType, ok := bc.inferred[pred]; ok {
		return bc.feasibleAlternatives(pred, relType, args, varRanges)
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
	return bc.feasibleAlternatives(pred, relTypeExpr, args, varRanges)
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
		relType, err := bc.inferRelTypesFromClause(clause)
		if err != nil {
			return nil, err
		}
		for _, alternative := range symbols.RelTypeAlternatives(relType) {
			if !symbols.TypeConforms(alternative, symbols.RelTypeFromAlternatives(alternatives)) {
				alternatives = append(alternatives, alternative)
			}
		}
	}
	bc.inferred[pred] = symbols.RelTypeFromAlternatives(alternatives)
	return bc.inferred[pred], nil
}

// inferState is state of inference while iterating over premises.
// The relation type is represented implicitly in usedVars and varTpe.
// Assigns to each var in usedVars a type (possibly union) in varTpe.
type inferState struct {
	// The index of the premise to be inspected with this state.
	index    int
	usedVars VarList
	varTpe   []ast.BaseTerm
}

func (s *inferState) String() string {
	return fmt.Sprintf("<%d; %v, %v>", s.index, s.usedVars, s.varTpe)
}

func (s *inferState) makeNext() *inferState {
	dest := make([]ast.BaseTerm, len(s.varTpe))
	for i, tpe := range s.varTpe {
		dest[i] = tpe
	}
	return &inferState{s.index + 1, s.usedVars, dest}
}

// addOrRefine either adds a binding or intersects type for an existing one.
func (s *inferState) addOrRefine(v ast.Variable, tpe ast.BaseTerm) error {
	if v.Symbol == "_" {
		return nil
	}
	i := s.usedVars.Find(v)
	if i == -1 {
		s.usedVars = s.usedVars.Extend([]ast.Variable{v})
		s.varTpe = append(s.varTpe, tpe)
		return nil
	}

	tpe = symbols.LowerBound([]ast.BaseTerm{s.varTpe[i], tpe})
	if tpe.Equals(symbols.EmptyType) {
		return fmt.Errorf("variable %v cannot have both %v and %v", v, s.varTpe[i], tpe)
	}
	s.varTpe[i] = tpe
	return nil
}

func (s *inferState) asMap() map[ast.Variable]ast.BaseTerm {
	m := make(map[ast.Variable]ast.BaseTerm, len(s.varTpe))
	for i, v := range s.usedVars.Vars {
		m[v] = s.varTpe[i]
	}
	return m
}

// inferRelTypesFromPremise is called for index \in 0..len(premises). It
// maps one state of inference to its (possibly empty) list of successors.
func (bc *BoundsAnalyzer) inferRelTypesFromPremise(premises []ast.Term, state *inferState) ([]*inferState, error) {
	var nextStates []*inferState

	premise := premises[state.index]
	switch t := premise.(type) {
	case ast.Atom:
		atom := t
		var (
			alternatives []ast.BaseTerm
			err          error
		)
		if declared, ok := bc.relTypeMap[atom.Predicate]; ok {
			alternatives, err = bc.feasibleAlternatives(atom.Predicate, declared, atom.Args, state.asMap())
		} else {
			alternatives, err = bc.getOrInferRelTypes(atom.Predicate, atom.Args, state.asMap())
		}
		if err != nil {
			return nil, fmt.Errorf("type mismatch %v : %v ", premise, err)
		}
		for _, alternative := range alternatives {
			relTypeArgs, err := symbols.RelTypeArgs(alternative)
			if err != nil {
				return nil, err // This cannot happen.
			}
			nextState := state.makeNext()
			for i, a := range atom.Args {
				if v, ok := a.(ast.Variable); ok {
					// No error-check needed - alternative is feasible.
					nextState.addOrRefine(v, relTypeArgs[i])
				}
			}
			nextStates = append(nextStates, nextState)
		}
		return nextStates, nil

	case ast.NegAtom:
		atom := t.Atom
		var (
			alternatives []ast.BaseTerm
			err          error
		)
		if declared, ok := bc.relTypeMap[atom.Predicate]; ok {
			alternatives, err = bc.feasibleAlternatives(atom.Predicate, declared, atom.Args, state.asMap())
		} else {
			alternatives, err = bc.getOrInferRelTypes(atom.Predicate, atom.Args, state.asMap())
		}
		if err != nil {
			return nil, fmt.Errorf("type mismatch %v : %v ", premise, err)
		}
		for _, alternative := range alternatives {
			relTypeArgs, err := symbols.RelTypeArgs(alternative)
			if err != nil {
				return nil, err // This cannot happen.
			}
			nextState := state.makeNext()
			for i, a := range atom.Args {
				if v, ok := a.(ast.Variable); ok {
					// No error-check needed - alternative is feasible.
					nextState.addOrRefine(v, relTypeArgs[i])
				}
			}
			nextStates = append(nextStates, nextState)
		}
		return nextStates, nil

	case ast.Eq:
		nextState := state.makeNext()
		varRanges := nextState.asMap()
		if leftVar, ok := t.Left.(ast.Variable); ok {
			tpe := boundOfArg(t.Right, varRanges, bc.nameTrie)
			if err := nextState.addOrRefine(leftVar, tpe); err != nil {
				return nil, err
			}
		}
		if rightVar, ok := t.Right.(ast.Variable); ok {
			tpe := boundOfArg(t.Left, varRanges, bc.nameTrie)
			if err := nextState.addOrRefine(rightVar, tpe); err != nil {
				return nil, err
			}
		}
		return []*inferState{nextState}, nil

	case ast.Ineq:
		nextState := state.makeNext()
		leftTpe := boundOfArg(t.Left, state.asMap(), bc.nameTrie)
		rightTpe := boundOfArg(t.Right, state.asMap(), bc.nameTrie)

		tpe := symbols.LowerBound([]ast.BaseTerm{leftTpe, rightTpe})
		if tpe.Equals(symbols.EmptyType) {
			return nil, fmt.Errorf("type mismatch %v : left type %v right type %v", premise, leftTpe, rightTpe)
		}
		if leftVar, ok := t.Left.(ast.Variable); ok {
			if err := nextState.addOrRefine(leftVar, tpe); err != nil {
				return nil, err
			}
		}
		if rightVar, ok := t.Right.(ast.Variable); ok {
			if err := nextState.addOrRefine(rightVar, tpe); err != nil {
				return nil, err
			}
		}
		return []*inferState{nextState}, nil
	}
	return nil, fmt.Errorf("unexpected state %v", premise)
}

// inferRelTypesFromClause infers possible relation types for the head predicate of a single clause.
func (bc *BoundsAnalyzer) inferRelTypesFromClause(clause ast.Clause) (ast.BaseTerm, error) {
	usedVars := VarList{}
	state := &inferState{0, usedVars, []ast.BaseTerm{}}

	levels := make([][]*inferState, len(clause.Premises)+1)
	levels[0] = []*inferState{state}
	for i := range clause.Premises {
		for _, state := range levels[i] {
			nextStates, err := bc.inferRelTypesFromPremise(clause.Premises, state)
			if err != nil {
				continue
			}
			levels[i+1] = append(levels[i+1], nextStates...)
		}
		if len(levels[i+1]) == 0 {
			return nil, fmt.Errorf("type mismatch: cannot find assignment that works for premise %v", clause.Premises[i])
		}
	}
	var relTypes []ast.BaseTerm
	for _, state := range levels[len(clause.Premises)] {
		s := state.makeNext()
		if clause.Transform != nil {
			for _, tr := range clause.Transform.Statements {
				if tr.Var != nil {
					s.addOrRefine(*tr.Var, typeOfFn(tr.Fn, s.asMap(), bc.nameTrie))
				}
			}
		}

		headTuple := make([]ast.BaseTerm, len(clause.Head.Args))
		for i, arg := range clause.Head.Args {
			headTuple[i] = boundOfArg(arg, s.asMap(), bc.nameTrie)
		}
		relTypes = append(relTypes, symbols.NewRelType(headTuple...))
	}
	return symbols.RelTypeFromAlternatives(relTypes), nil
}

func boundOfArg(x ast.BaseTerm, varRanges map[ast.Variable]ast.BaseTerm, nameTrie nametrie) ast.BaseTerm {
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
		case ast.StringType:
			return ast.StringBound
		case ast.NameType:
			if z == ast.AnyBound || z == ast.StringBound || z == ast.NumberBound || z == ast.NameBound {
				return z
			}
			return prefixType(nameTrie, z.Symbol)
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
		switch z.Function {
		case symbols.Cons:
			argType := boundOfArg(z.Args[0], varRanges, nameTrie)
			tailType := boundOfArg(z.Args[1], varRanges, nameTrie)
			return ast.ApplyFn{symbols.ListType, []ast.BaseTerm{symbols.UpperBound([]ast.BaseTerm{argType, tailType})}}
		case symbols.List:
			if len(z.Args) == 0 {
				return ast.ApplyFn{symbols.ListType, []ast.BaseTerm{ast.AnyBound}}
			}
			var argTypes []ast.BaseTerm
			for _, arg := range z.Args {
				argTypes = append(argTypes, boundOfArg(arg, varRanges, nameTrie))
			}
			return symbols.NewListType(symbols.UpperBound(argTypes))

		case symbols.Map:
			var keyTpes []ast.BaseTerm
			var valTpes []ast.BaseTerm
			for i := 0; i < len(z.Args); i++ {
				keyTpes = append(keyTpes, boundOfArg(z.Args[i], varRanges, nameTrie))
				i++
				valTpes = append(valTpes, boundOfArg(z.Args[i], varRanges, nameTrie))
			}
			return symbols.NewMapType(symbols.UpperBound(keyTpes), symbols.UpperBound(valTpes))

		case symbols.Struct:
			var fields []ast.BaseTerm
			for i := 0; i < len(z.Args); i++ {
				fields = append(fields, z.Args[i])
				i++
				fields = append(fields, boundOfArg(z.Args[i], varRanges, nameTrie))
			}

			return symbols.NewStructType(fields...)

		case symbols.Pair:
			leftTpe := boundOfArg(z.Args[0], varRanges, nameTrie)
			rightTpe := boundOfArg(z.Args[1], varRanges, nameTrie)
			return ast.ApplyFn{symbols.PairType, []ast.BaseTerm{leftTpe, rightTpe}}
		case symbols.Tuple:
			var argTypes []ast.BaseTerm
			for _, arg := range z.Args {
				argTypes = append(argTypes, boundOfArg(arg, varRanges, nameTrie))
			}
			return ast.ApplyFn{symbols.TupleType, argTypes}
		}
		return ast.AnyBound
	}
	return ast.AnyBound
}

func typeOfFn(x ast.ApplyFn, varRanges map[ast.Variable]ast.BaseTerm, nameTrie nametrie) ast.BaseTerm {
	switch x.Function.Symbol {
	case symbols.Max.Symbol:
		fallthrough
	case symbols.Min.Symbol:
		fallthrough
	case symbols.Div.Symbol:
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
	return ast.AnyBound
}

func prefixType(nameTrie nametrie, sym string) ast.Constant {
	parts := strings.Split(sym, "/")
	if len(parts) == 1 {
		return ast.NameBound
	}
	index := nameTrie.LongestPrefix(parts[1:])
	if index == -1 {
		return ast.NameBound
	}
	prefixstrlen := index + 1 // number of "/" separators
	for i := 0; i <= index; i++ {
		prefixstrlen += len(parts[i+1])
	}
	n, err := ast.Name(sym[:prefixstrlen])
	if err != nil {
		return ast.NameBound // This cannot happen
	}
	return n
}
