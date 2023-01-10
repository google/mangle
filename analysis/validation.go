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
	// maps foo(X, Y, Z) to RelType(boundX, boundY, boundZ)
	relTypeMap map[ast.PredicateSym]ast.BaseTerm
	// types observed from initial facts.
	initialFactMap map[ast.PredicateSym][]ast.BaseTerm
	visiting       map[ast.PredicateSym]bool
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
			head, err := builtin.EvalAtom(clause.Head, ast.ConstSubstList{})
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
		bc := newBoundsAnalyzer(&programInfo, nameTrie, initialFacts, rulesMap)
		if err := bc.BoundsCheck(); err != nil {
			if a.boundsCheckingMode == ErrorForBoundsMismatch {
				return nil, err
			}
			fmt.Printf("bounds error: %v\n", err)
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

func newBoundsAnalyzer(programInfo *ProgramInfo, nameTrie nametrie, initialFacts []ast.Atom, rulesMap map[ast.PredicateSym][]ast.Clause) *BoundsAnalyzer {
	initialFactUnion := make(map[ast.PredicateSym][][]ast.BaseTerm)
	for _, f := range initialFacts {
		observation := make([]ast.BaseTerm, len(f.Args))
		for i, arg := range f.Args {
			if c, ok := arg.(ast.Constant); ok {
				observation[i] = boundOfArg(c, nil, nameTrie)
			} else {
				observation[i] = ast.AnyBound // This cannot happen.
			}
		}
		initialFactUnion[f.Predicate] = append(initialFactUnion[f.Predicate], observation)
	}
	initialFactTypes := make(map[ast.PredicateSym][]ast.BaseTerm)
	for pred, obs := range initialFactUnion {
		transposed := make([][]ast.BaseTerm, pred.Arity)
		for i := 0; i < pred.Arity; i++ {
			for _, observed := range obs {
				transposed[i] = append(transposed[i], observed[i])
			}
		}
		res := make([]ast.BaseTerm, pred.Arity)
		for i := 0; i < pred.Arity; i++ {
			res[i] = symbols.UpperBound(transposed[i])
		}
		initialFactTypes[pred] = res
	}

	relTypeMap := make(map[ast.PredicateSym]ast.BaseTerm)

	// Populate relTypes with declared type for extensional relations.
	for pred := range programInfo.EdbPredicates {
		relTypeMap[pred] = symbols.ApproximateRelTypeFromDecl(*programInfo.Decls[pred])
	}

	visiting := make(map[ast.PredicateSym]bool)
	return &BoundsAnalyzer{programInfo, nameTrie, rulesMap, relTypeMap, initialFactTypes, visiting}
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
	preds := make([]ast.PredicateSym, len(bc.programInfo.IdbPredicates))
	i := 0
	for pred := range bc.programInfo.IdbPredicates {
		preds[i] = pred
		i++
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

func (bc *BoundsAnalyzer) inferAndCheckBounds(pred ast.PredicateSym) error {
	decl, ok := bc.programInfo.Decls[pred]
	if !ok {
		return nil // This should not happen.
	}

	inferredArgs, err := bc.inferArgumentRange(pred)
	if err != nil {
		return err
	}
	// Iterate over arguments
	argumentRange, _ := symbols.RelTypeArgs(symbols.ApproximateRelTypeFromDecl(*decl))
	for i, declared := range argumentRange {
		inferred := inferredArgs[i]
		if symbols.TypeConforms(inferred, declared) {
			continue
		}

		var rules strings.Builder
		for _, r := range bc.RulesMap[pred] {
			rules.WriteString(r.String() + "\n")
		}
		return fmt.Errorf("type mismatch for pred %v argument %q\ninferred: %v\nvs declared %v\nrules:\n%s",
			decl.DeclaredAtom, decl.DeclaredAtom.Args[i].String(), inferred, declared, rules.String())
	}
	return nil
}

func (bc *BoundsAnalyzer) getOrInferArgumentRange(pred ast.PredicateSym) ([]ast.BaseTerm, error) {
	if bc.visiting[pred] {
		unknownBounds := make([]ast.BaseTerm, pred.Arity)
		for i := 0; i < pred.Arity; i++ {
			unknownBounds[i] = ast.AnyBound

		}
		return unknownBounds, nil
	}

	bc.visiting[pred] = true
	defer delete(bc.visiting, pred)
	if relType, ok := bc.relTypeMap[pred]; ok {
		return symbols.RelTypeArgs(relType)
	}
	if decl, ok := bc.programInfo.Decls[pred]; ok && !decl.IsSynthetic() {
		// If a decl was provided, use that.
		relTypeFromDecl := symbols.ApproximateRelTypeFromDecl(*decl)
		bc.relTypeMap[pred] = relTypeFromDecl
		return symbols.RelTypeArgs(relTypeFromDecl)
	}

	res, err := bc.inferArgumentRange(pred)
	if err != nil {
		return nil, err
	}
	bc.relTypeMap[pred] = symbols.NewRelType(res...)
	return res, nil
}

// inferArgumentRange ensures that bc.range[pred] is populated with the inferred argument range.
// It will inspect clause bodies, only consulting decl for extensional predicates.
func (bc *BoundsAnalyzer) inferArgumentRange(pred ast.PredicateSym) ([]ast.BaseTerm, error) {
	if pred.IsBuiltin() {
		relTypeArgs, err := symbols.RelTypeArgs(symbols.BuiltinRelations[pred])
		if err != nil {
			return nil, fmt.Errorf("could not get relation type args for builtin %v : %v", pred, err)
		}
		return relTypeArgs, nil
	}
	if ranges, ok := bc.relTypeMap[pred]; ok {
		return symbols.RelTypeArgs(ranges)
	}

	// Inspect all clauses for pred
	res := make([][]ast.BaseTerm, pred.Arity)
	clauses := bc.RulesMap[pred]
	if len(clauses) == 0 && len(bc.initialFactMap[pred]) == 0 {
		return nil, fmt.Errorf("unknown predicate %v (no clauses, no decls). typo?", pred)
	}
	for _, clause := range clauses {
		// We map each variable to a type expressions (possibly a union).
		varRanges := make(map[ast.Variable]ast.BaseTerm)
		// When we add an expected type from argument range, we check.
		addToVarRange := func(v ast.Variable, expectedTpe ast.BaseTerm) error {
			if v.Symbol == "_" {
				return nil
			}
			existing, ok := varRanges[v]
			if !ok {
				varRanges[v] = expectedTpe
				return nil
			}
			tpe := symbols.LowerBound([]ast.BaseTerm{existing, expectedTpe})
			if tpe.Equals(symbols.EmptyType) {
				return fmt.Errorf("problem in rule '%v': variable %v cannot have both %v and %v", clause, v, existing, expectedTpe)
			}
			varRanges[v] = tpe
			return nil
		}

		handleAtom := func(atom ast.Atom, isBinding bool) error {
			ranges, err := bc.getOrInferArgumentRange(atom.Predicate)
			if err != nil {
				return err
			}
			for i, arg := range atom.Args {
				v, ok := arg.(ast.Variable)
				if !ok {
					tpe := boundOfArg(arg, varRanges, bc.nameTrie)
					if !symbols.TypeConforms(tpe, ranges[i]) {
						return fmt.Errorf("problem in rule '%v' subgoal %v: argument %v of type %v does not conform to type %v", clause, atom, arg, tpe, ranges[i])
					}
					continue
				}
				if isBinding { // negated subgoals do not contribute to binding.
					if err := addToVarRange(v, ranges[i]); err != nil {
						return err
					}
				}
			}
			return nil
		}
		handlePair := func(left ast.BaseTerm, right ast.BaseTerm, isBinding bool) error {
			if !isBinding {
				return nil //  check that types are comparable?
			}
			if leftVar, ok := left.(ast.Variable); ok {
				return addToVarRange(leftVar, boundOfArg(right, varRanges, bc.nameTrie))
			}

			if rightVar, ok := right.(ast.Variable); ok {
				return addToVarRange(rightVar, boundOfArg(left, varRanges, bc.nameTrie))
			}

			// This cannot happen.
			return fmt.Errorf("cannot type check pair %v %v", left, right)
		}

		for _, premise := range clause.Premises {
			switch subgoal := premise.(type) {
			case ast.Atom:
				if err := handleAtom(subgoal, true); err != nil {
					return nil, err
				}

			case ast.NegAtom:
				if err := handleAtom(subgoal.Atom, false); err != nil {
					return nil, err
				}

			case ast.Eq:
				if err := handlePair(subgoal.Left, subgoal.Right, true); err != nil {
					return nil, err
				}

			case ast.Ineq:
				if err := handlePair(subgoal.Left, subgoal.Right, false); err != nil {
					return nil, err
				}
			}
		}
		if clause.Transform != nil {
			for _, tr := range clause.Transform.Statements {
				if tr.Var != nil {
					varRanges[*tr.Var] = typeOfFn(tr.Fn, varRanges, bc.nameTrie)
				}
			}
		}
		for i, arg := range clause.Head.Args {
			v, ok := arg.(ast.Variable)
			if !ok {
				res[i] = append(res[i], boundOfArg(arg, varRanges, bc.nameTrie))
				continue
			}
			res[i] = append(res[i], varRanges[v])
		}
	}
	argRanges := make([]ast.BaseTerm, pred.Arity)
	initialFactTypes, hasInitial := bc.initialFactMap[pred]
	for i := range res {
		if hasInitial {
			res[i] = append(res[i], initialFactTypes[i])
		}
		argRanges[i] = symbols.UpperBound(res[i])
	}
	bc.relTypeMap[pred] = symbols.NewRelType(argRanges...)
	return argRanges, nil
}

func boundOfArg(x ast.BaseTerm, varRanges map[ast.Variable]ast.BaseTerm, nameTrie nametrie) ast.BaseTerm {
	switch z := x.(type) {
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
			return ast.ApplyFn{symbols.ListType, []ast.BaseTerm{symbols.UpperBound(argTypes)}}
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
