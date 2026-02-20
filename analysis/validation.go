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

	"go.uber.org/multierr"
	"codeberg.org/TauCeti/mangle-go/ast"
	"codeberg.org/TauCeti/mangle-go/functional"
	"codeberg.org/TauCeti/mangle-go/packages"
	"codeberg.org/TauCeti/mangle-go/parse"
	"codeberg.org/TauCeti/mangle-go/symbols"
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
// EdbPredicates and IdbPredicates are disjoint.
type ProgramInfo struct {
	// Extensional predicate symbols; those that do not appear in the head of a clause with a body.
	EdbPredicates map[ast.PredicateSym]struct{}
	// Intensional predicate symbols; those that do appear in the head of a clause with a body.
	IdbPredicates map[ast.PredicateSym]struct{}
	// Heads of rules without a body.
	InitialFacts []ast.Atom
	// Validity intervals for InitialFacts (parallel slice, nil for eternal facts).
	InitialFactTimes []*ast.Interval
	// Rules that have a body.
	Rules []ast.Clause
	// Desugared declarations for all predicates, possibly synthetic.
	Decls map[ast.PredicateSym]*ast.Decl
	// Warnings collected during analysis.
	Warnings []TemporalWarning
}

// Analyzer is a struct providing built-in predicates and functions for name analysis.
type Analyzer struct {
	// Predicates that can be referenced by rules though we do not have declarations in source.
	// Keys are disjoint from builtin.Predicates.
	extraPredicates map[ast.PredicateSym]ast.Decl
	// Additional functions.
	// Keys are disjoint from builtin functions.
	extraFunctions map[ast.FunctionSym]ast.BaseTerm
	// Declaration of predicates to be analyzed.
	// Keys are disjoint from extraPredicates and builtin.Predicates.
	decl map[ast.PredicateSym]ast.Decl
	// Whether mismatch in bounds leads to an error
	boundsCheckingMode BoundsCheckingMode
}

// AnalyzeOneUnit is a convenience method to analyze a program consisting of a single source unit.
func AnalyzeOneUnit(unit parse.SourceUnit, extraPredicates map[ast.PredicateSym]ast.Decl) (*ProgramInfo, error) {
	return Analyze([]parse.SourceUnit{unit}, extraPredicates)
}

// Analyze identifies the extensional and intensional predicates of a program and checks every rule.
func Analyze(program []parse.SourceUnit, extraPredicates map[ast.PredicateSym]ast.Decl) (*ProgramInfo, error) {
	return AnalyzeAndCheckBounds(program, extraPredicates, NoBoundsChecking)
}

// ExtractPackages turns source units into merged source packages.
func ExtractPackages(program []parse.SourceUnit) (map[string]*packages.Package, error) {
	pkgs := map[string]*packages.Package{}
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
	return pkgs, nil
}

// AnalyzeAndCheckBounds checks every rule, including bounds.
func AnalyzeAndCheckBounds(program []parse.SourceUnit, extraPredicates map[ast.PredicateSym]ast.Decl, boundsChecking BoundsCheckingMode) (*ProgramInfo, error) {
	pkgs, err := ExtractPackages(program)
	if err != nil {
		return nil, err
	}
	var clauses []ast.Clause
	var decls []ast.Decl
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

	analyzer, err := New(extraPredicates, decls, boundsChecking)
	if err != nil {
		return nil, err
	}
	if err := analyzer.EnsureDecl(clauses); err != nil {
		return nil, err
	}

	// Resolve "MaybeTemporal" declarations.
	// If a synthetic declaration is marked MaybeTemporal, it means we saw at least
	// one temporal usage. We promote it to Temporal.
	// We also filter out the internal MaybeTemporal descriptor.
	for pred, decl := range analyzer.decl {
		if decl.IsMaybeTemporal() {
			// Remove MaybeTemporal descriptor
			var newDescr []ast.Atom
			for _, d := range decl.Descr {
				if d.Predicate.Symbol != ast.DescrMaybeTemporal {
					newDescr = append(newDescr, d)
				}
			}
			// Add Temporal descriptor
			newDescr = append(newDescr, ast.NewAtom(ast.DescrTemporal))
			decl.Descr = newDescr
			analyzer.decl[pred] = decl
		}
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

// New creates a new analyzer based on declarations and extra predicates.
func New(extraPredicates map[ast.PredicateSym]ast.Decl, decls []ast.Decl, boundsChecking BoundsCheckingMode) (*Analyzer, error) {
	extraByName := byName(extraPredicates)
	declMap := make(map[ast.PredicateSym]ast.Decl)
	for _, decl := range decls {
		pred := decl.DeclaredAtom.Predicate
		if pred == symbols.Package || pred == symbols.Use {
			continue
		}
		if existing, ok := declMap[pred]; ok {
			return nil, fmt.Errorf("predicate %v declared more than once, previous was %v", pred, existing)
		}
		declMap[pred] = decl

		if extraDecl, ok := extraByName[pred.Symbol]; ok {
			// We have a user declaration for a symbol that is also known via extraPredicates.
			if !extraDecl.IsSynthetic() {
				return nil, fmt.Errorf("cannot redeclare %v, previous Decl %v", decl, extraDecl)
			}
			// We can override a synthetic decl, but arity should still match.
			if extraDecl.DeclaredAtom.Predicate.Arity != pred.Arity {
				return nil, fmt.Errorf("declared arity %v conflicts with extra decl %v", decl, extraDecl)
			}
			// Override the synthetic decl with the one we were requested to use.
			delete(extraPredicates, pred)
		}
	}
	return &Analyzer{extraPredicates, nil /* extraFunctions */, declMap, boundsChecking}, nil
}

// EnsureDecl ensures that every predicate in the program is declared.
// It also ensures consistency of temporal usage.
func (a *Analyzer) EnsureDecl(clauses []ast.Clause) error {
	extraByName := byName(a.extraPredicates)
	declByName := byName(a.decl)
	for _, c := range clauses {
		pred := c.Head.Predicate
		name := pred.Symbol
		// Check that the name was not defined previously (in a separate source).
		// We may permit "distributing" definitions over source files later.
		if decl, ok := a.extraPredicates[pred]; ok {
			if decl.IsExtensional() && len(c.Premises) == 0 {
				continue
			}
			return fmt.Errorf("predicate %v was defined previously %v", decl.DeclaredAtom.Predicate, decl)
		}
		if decl, ok := extraByName[name]; ok { // different arity
			return fmt.Errorf(
				"predicate %v was defined previously", decl.DeclaredAtom.Predicate)
		}

		// If we already have a declaration, check if it is compatible.
		// Note that we may have multiple declarations for the same predicate
		// (e.g. one from use and one from package).
		if decl, ok := a.decl[pred]; ok {
			// If we have an existing declaration, check for temporal consistency
			if c.HeadTime != nil && !c.HeadTime.IsEternal() {
				if !decl.IsTemporal() && !decl.IsMaybeTemporal() {
					return fmt.Errorf("predicate %v is not declared temporal but used with temporal annotation in %v", pred, c)
				}
			}
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

		// If this is a synthetic declaration and we have a temporal annotation,
		// mark it as "MaybeTemporal" so we can check consistency later.
		if c.HeadTime != nil && !c.HeadTime.IsEternal() {
			synthDecl.Descr = append(synthDecl.Descr, ast.NewAtom(ast.DescrMaybeTemporal))
		}

		a.decl[pred] = synthDecl
		declByName[pred.Symbol] = a.decl[pred]
	}
	return nil
}

// Analyze identifies the extensional and intensional predicates of a program, checks every rule and that
// all references to built-in predicates and functions used in transforms are valid.
func (a *Analyzer) Analyze(program []ast.Clause) (*ProgramInfo, error) {
	if err := a.EnsureDecl(program); err != nil {
		return nil, err
	}
	// Resolve MaybeTemporal descriptors to Temporal.
	for sym, decl := range a.decl {
		if decl.IsMaybeTemporal() {
			newDescr := make([]ast.Atom, 0, len(decl.Descr))
			temporalAtom := ast.NewAtom(ast.DescrTemporal)
			maybeTemporalAtom := ast.NewAtom(ast.DescrMaybeTemporal)
			for _, d := range decl.Descr {
				if !d.Equals(maybeTemporalAtom) {
					newDescr = append(newDescr, d)
				}
			}
			newDescr = append(newDescr, temporalAtom)
			decl.Descr = newDescr
			a.decl[sym] = decl
		}
	}
	globalDecls := make(map[ast.PredicateSym]ast.Decl)
	for p, d := range a.extraPredicates {
		globalDecls[p] = d
	}
	for p, d := range a.decl {
		globalDecls[p] = d
		if errs := CheckDecl(d); errs != nil {
			return nil, multierr.Combine(errs...)
		}
	}
	desugaredDecls, err := symbols.CheckAndDesugar(globalDecls)
	if err != nil {
		return nil, err
	}
	edbSymbols := make(map[ast.PredicateSym]struct{})
	idbSymbols := make(map[ast.PredicateSym]struct{})
	var initialFacts []ast.Atom
	var initialFactTimes []*ast.Interval
	var rules []ast.Clause
	rulesMap := make(map[ast.PredicateSym][]ast.Clause)
	for _, clause := range program {
		clause = RewriteClause(desugaredDecls, clause)

		// Normalize TemporalAtom to TemporalLiteral (or bare Atom) for consistent analysis
		if clause.Premises != nil {
			for i, premise := range clause.Premises {
				if ta, ok := premise.(ast.TemporalAtom); ok {
					if ta.Interval == nil {
						clause.Premises[i] = ta.Atom
					} else {
						clause.Premises[i] = ast.TemporalLiteral{
							Literal:  ta.Atom,
							Interval: ta.Interval,
						}
					}
				}
			}
		}

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
			initialFactTimes = append(initialFactTimes, clause.HeadTime)
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

	programInfo := ProgramInfo{edbSymbols, idbSymbols, initialFacts, initialFactTimes, rules, desugaredDecls, nil}

	// Check for temporal recursion issues
	if warnings := CheckTemporalRecursion(&programInfo); len(warnings) > 0 {
		var errs error
		for _, w := range warnings {
			if w.Severity == SeverityCritical {
				errs = multierr.Append(errs, fmt.Errorf("temporal analysis error: %v", w))
			} else {
				programInfo.Warnings = append(programInfo.Warnings, w)
			}
		}
		if errs != nil {
			return nil, errs
		}
	}

	if a.boundsCheckingMode != NoBoundsChecking {
		nameTrie := collectNames(a.extraPredicates, programInfo.Decls)
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
