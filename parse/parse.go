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

// Package parse provides methods to parse datalog programs and parts thereof.
package parse

import (
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"sync"

	antlr "github.com/antlr/antlr4/runtime/Go/antlr/v4"
	"github.com/google/mangle/ast"
	"github.com/google/mangle/parse/gen"
	"github.com/google/mangle/symbols"
)

var (
	packageSym   = ast.Atom{Predicate: symbols.Package}
	useSym       = ast.Atom{Predicate: symbols.Use}
	emptyPackage = ast.Decl{DeclaredAtom: packageSym, Descr: []ast.Atom{nameAtom("")}}

	lexerPool *sync.Pool = &sync.Pool{
		New: func() any {
			l := gen.NewMangleLexer(nil)
			l.RemoveErrorListeners()
			return l
		},
	}
	parserPool *sync.Pool = &sync.Pool{
		New: func() any {
			p := gen.NewMangleParser(nil)
			p.RemoveErrorListeners()
			return p
		},
	}
)

var _ gen.MangleVisitor = (*Parser)(nil)

const commentStart = '#'

// Parser represents an object that can be used for parsing.
type Parser struct {
	gen.BaseMangleVisitor
	input  string
	lexer  *gen.MangleLexer
	parser *gen.MangleParser
	errors *errorsList
}

// Holds errors.
type errorsList struct {
	errors []Error
}

func (e *errorsList) Add(msg string, line int, col int) {
	e.errors = append(e.errors, Error{Message: msg, Line: line, Column: col})
}

func (p *Parser) init() error {
	p.lexer = lexerPool.Get().(*gen.MangleLexer)
	p.parser = parserPool.Get().(*gen.MangleParser)

	p.lexer.SetInputStream(antlr.NewInputStream(p.input))
	p.parser.SetInputStream(antlr.NewCommonTokenStream(p.lexer, 0))

	p.lexer.AddErrorListener(p)
	p.parser.AddErrorListener(p)
	return nil
}

func (p *Parser) reset() {
	// Reset the lexer and parser before putting them back in the pool.
	p.lexer.RemoveErrorListeners()
	p.parser.RemoveErrorListeners()
	p.lexer.SetInputStream(nil)
	p.parser.SetInputStream(nil)
	lexerPool.Put(p.lexer)
	parserPool.Put(p.parser)
	p.lexer = nil
	p.parser = nil
}

// SourceUnit consists of decls and clauses.
type SourceUnit struct {
	Decls   []ast.Decl
	Clauses []ast.Clause
}

// Error represents a parser error messages, without location.
type Error struct {
	Message string
	Line    int // 1-based line number within source.
	Column  int // 0-based line number within source.
}

func newParser(input string) (*Parser, error) {
	p := &Parser{input: input, errors: &errorsList{}}
	if err := p.init(); err != nil {
		return nil, err
	}
	return p, nil
}

// Unit parses a source unit (clauses and decls).
func Unit(reader io.Reader) (SourceUnit, error) {
	bytes, err := io.ReadAll(reader)
	if err != nil {
		return SourceUnit{}, err
	}
	p, err := newParser(string(bytes))
	if err != nil {
		return SourceUnit{}, err
	}
	defer p.reset()
	tree := p.parser.Start()
	if err := p.error(); err != nil {
		return SourceUnit{}, err
	}
	sourceUnit := p.Visit(tree)
	if err := p.error(); err != nil {
		return SourceUnit{}, err
	}
	return sourceUnit.(SourceUnit), nil
}

func (p *Parser) error() error {
	if len(p.errors.errors) == 0 {
		return nil
	}
	var buf strings.Builder
	for _, e := range p.errors.errors {
		buf.WriteString(fmt.Sprintf("%d:%d %s\n", e.Line, e.Column, e.Message))
	}
	return errors.New(buf.String())
}

// Visit is the Visitor implementation.
func (p *Parser) Visit(tree antlr.ParseTree) any {
	switch tree.(type) {
	case *gen.StartContext:
		return p.VisitStart(tree.(*gen.StartContext))
	case *gen.ProgramContext:
		return p.VisitProgram(tree.(*gen.ProgramContext))
	case *gen.PackageDeclContext:
		return p.VisitPackageDecl(tree.(*gen.PackageDeclContext))
	case *gen.UseDeclContext:
		return p.VisitUseDecl(tree.(*gen.UseDeclContext))
	case *gen.DeclContext:
		return p.VisitDecl(tree.(*gen.DeclContext))
	case *gen.DescrBlockContext:
		return p.VisitDescrBlock(tree.(*gen.DescrBlockContext))
	case *gen.BoundsBlockContext:
		return p.VisitBoundsBlock(tree.(*gen.BoundsBlockContext))
	case *gen.ConstraintsBlockContext:
		return p.VisitConstraintsBlock(tree.(*gen.ConstraintsBlockContext))
	case *gen.ClauseContext:
		return p.VisitClause(tree.(*gen.ClauseContext))
	case *gen.ClauseBodyContext:
		return p.VisitClauseBody(tree.(*gen.ClauseBodyContext))
	case *gen.TransformContext:
		return p.VisitTransform(tree.(*gen.TransformContext))
	case *gen.LetStmtContext:
		return p.VisitLetStmt(tree.(*gen.LetStmtContext))
	case *gen.LiteralOrFmlContext:
		return p.VisitLiteralOrFml(tree.(*gen.LiteralOrFmlContext))
	case *gen.VarContext:
		return p.VisitVar(tree.(*gen.VarContext))
	case *gen.ConstContext:
		return p.VisitConst(tree.(*gen.ConstContext))
	case *gen.NumContext:
		return p.VisitNum(tree.(*gen.NumContext))
	case *gen.FloatContext:
		return p.VisitFloat(tree.(*gen.FloatContext))
	case *gen.StrContext:
		return p.VisitStr(tree.(*gen.StrContext))
	case *gen.ApplContext:
		return p.VisitAppl(tree.(*gen.ApplContext))
	case *gen.MapContext:
		return p.VisitMap(tree.(*gen.MapContext))
	case *gen.StructContext:
		return p.VisitStruct(tree.(*gen.StructContext))
	case *gen.ListContext:
		return p.VisitList(tree.(*gen.ListContext))
	case *gen.AtomContext:
		return p.VisitAtom(tree.(*gen.AtomContext))
	case *gen.AtomsContext:
		return p.VisitAtoms(tree.(*gen.AtomsContext))
	}
	p.errors.Add(fmt.Sprintf("parse error: %q", tree.GetText()), 0, 0)
	return nil
}

// VisitStart visits a parse tree produced by MangleParser#start.
func (p Parser) VisitStart(ctx *gen.StartContext) any {
	return p.Visit(ctx.Program())
}

// VisitProgram visits a parse tree produced by MangleParser#program.
func (p Parser) VisitProgram(ctx *gen.ProgramContext) any {
	var decls []ast.Decl
	if packageDeclCtx := ctx.PackageDecl(); packageDeclCtx != nil {
		decls = append(decls, p.Visit(packageDeclCtx).(ast.Decl))
	} else {
		decls = append(decls, emptyPackage)
	}
	for _, u := range ctx.AllUseDecl() {
		decls = append(decls, p.Visit(u).(ast.Decl))
	}
	for _, d := range ctx.AllDecl() {
		decls = append(decls, p.Visit(d).(ast.Decl))
	}
	var clauses []ast.Clause
	for _, c := range ctx.AllClause() {
		clauses = append(clauses, p.Visit(c).(ast.Clause))
	}
	return SourceUnit{decls, clauses}
}

// VisitPackageDecl visits a parse tree produced by MangleParser#packageDecl.
func (p Parser) VisitPackageDecl(ctx *gen.PackageDeclContext) any {
	text := ctx.NAME().GetText()
	if text != strings.ToLower(text) {
		p.errors.Add(fmt.Sprintf("package names have to be lower-case: %s", text), ctx.GetStart().GetLine(), ctx.GetStart().GetColumn())
	}
	name := ast.NewAtom("name", ast.String(text))
	atoms := []ast.Atom{name}
	if atomsCtx := ctx.Atoms(); atomsCtx != nil {
		atoms = append(atoms, p.Visit(atomsCtx).([]ast.Atom)...)
	}
	return ast.Decl{DeclaredAtom: ast.Atom{Predicate: symbols.Package}, Descr: atoms}
}

// VisitUseDecl visits a parse tree produced by MangleParser#useDecl.
func (p Parser) VisitUseDecl(ctx *gen.UseDeclContext) any {
	name := ast.NewAtom("name", ast.String(ctx.NAME().GetText()))
	atoms := []ast.Atom{name}
	if atomsCtx := ctx.Atoms(); atomsCtx != nil {
		atoms = append(atoms, p.Visit(atomsCtx).([]ast.Atom)...)
	}
	return ast.Decl{DeclaredAtom: ast.Atom{Predicate: symbols.Use}, Descr: atoms}
}

// VisitDecl visits a parse tree produced by MangleParser#decl.
func (p Parser) VisitDecl(ctx *gen.DeclContext) any {
	atom := p.Visit(ctx.Atom()).(ast.Atom)
	var descrAtoms []ast.Atom
	if ctx.DescrBlock() != nil {
		descrAtoms = p.Visit(ctx.DescrBlock()).([]ast.Atom)
	}
	var bounds []ast.BoundDecl
	for _, b := range ctx.AllBoundsBlock() {
		bounds = append(bounds, p.Visit(b).(ast.BoundDecl))
	}
	var incl *ast.InclusionConstraint
	if ctx.ConstraintsBlock() != nil {
		gotIncl := p.Visit(ctx.ConstraintsBlock()).(ast.InclusionConstraint)
		incl = &gotIncl
	}
	decl, err := ast.NewDecl(atom, descrAtoms, bounds, incl)
	if err != nil {
		p.errors.Add(err.Error(), ctx.GetStart().GetLine(), ctx.GetStart().GetColumn())
		return ast.Decl{DeclaredAtom: atom}
	}
	return decl
}

// VisitDescrBlock visits a parse tree produced by MangleParser#descrBlock.
func (p Parser) VisitDescrBlock(ctx *gen.DescrBlockContext) any {
	return p.Visit(ctx.Atoms()).([]ast.Atom)
}

// VisitBoundsBlock visits a parse tree produced by MangleParser#boundsBlock.
func (p Parser) VisitBoundsBlock(ctx *gen.BoundsBlockContext) any {
	var baseTerms []ast.BaseTerm
	for _, t := range ctx.AllTerm() {
		term := p.Visit(t).(ast.Term)
		baseTerm, ok := term.(ast.BaseTerm)
		if !ok {
			p.errors.Add(fmt.Sprintf("expected base term got %v", term), ctx.GetStart().GetLine(), ctx.GetStart().GetColumn())
			baseTerms = append(baseTerms, ast.AnyBound)
		}
		baseTerms = append(baseTerms, baseTerm)
	}
	return ast.NewBoundDecl(baseTerms...)
}

// VisitConstraintsBlock visits a parse tree produced by MangleParser#constraintsBlock.
func (p Parser) VisitConstraintsBlock(ctx *gen.ConstraintsBlockContext) any {
	atoms := p.Visit(ctx.Atoms()).([]ast.Atom)
	return ast.NewInclusionConstraint(atoms)
}

// VisitClause visits a parse tree produced by MangleParser#clause.
func (p Parser) VisitClause(ctx *gen.ClauseContext) any {
	head := p.Visit(ctx.Atom()).(ast.Atom)
	if ctx.COLONDASH() != nil {
		body := p.Visit(ctx.ClauseBody()).(ast.Clause)
		body.Head = head
		return body
	}
	return ast.NewClause(head, nil)
}

// VisitClauseBody visits a parse tree produced by MangleParser#clauseBody.
func (p Parser) VisitClauseBody(ctx *gen.ClauseBodyContext) any {
	var premises []ast.Term
	for _, litOrFml := range ctx.AllLiteralOrFml() {
		premises = append(premises, p.Visit(litOrFml).(ast.Term))
	}
	clause := ast.NewClause(ast.Atom{}, premises)
	if ctx.PIPEGREATER() != nil {
		t := p.Visit(ctx.Transform()).(ast.Transform)
		clause.Transform = &t
	}
	return clause
}

// VisitTransform visits a parse tree produced by MangleParser#transform.
func (p Parser) VisitTransform(ctx *gen.TransformContext) any {
	var stmts []ast.TransformStmt
	if ctx.DO() != nil {
		term := p.Visit(ctx.Term()).(ast.Term)
		apply, ok := term.(ast.ApplyFn)
		if !ok {
			p.errors.Add(fmt.Sprintf("expected fn application got %v", term), ctx.Term().GetStart().GetLine(), ctx.Term().GetStart().GetColumn())
			return ast.Transform{}
		}
		stmts = append(stmts, ast.TransformStmt{nil, apply})
	}
	for _, l := range ctx.AllLetStmt() {
		stmts = append(stmts, p.Visit(l).(ast.TransformStmt))
	}
	return ast.Transform{stmts}
}

// VisitLetStmt visits a parse tree produced by MangleParser#letStmt.
func (p Parser) VisitLetStmt(ctx *gen.LetStmtContext) any {
	v := ast.Variable{ctx.VARIABLE().GetText()}
	rhs := p.Visit(ctx.Term()).(ast.Term)
	apply, ok := rhs.(ast.ApplyFn)
	if !ok {
		p.errors.Add(fmt.Sprintf("expected fn application got %v", rhs), ctx.Term().GetStart().GetLine(), ctx.Term().GetStart().GetColumn())
		return ast.TransformStmt{&v, ast.ApplyFn{ast.FunctionSym{"fn:b0rk", 0}, nil}}
	}
	return ast.TransformStmt{&v, apply}
}

// VisitLiteralOrFml visits a parse tree produced by literalOrFml
func (p Parser) VisitLiteralOrFml(ctx *gen.LiteralOrFmlContext) any {
	term := p.Visit(ctx.Term(0)).(ast.Term)
	atom, ok := term.(ast.Atom)
	if ctx.BANG() != nil {
		if !ok {
			p.errors.Add(fmt.Sprintf("not a literal or fml: %v", ctx.Term(0).GetText()), ctx.Term(0).GetStart().GetLine(), ctx.Term(0).GetStart().GetColumn())
		}
		return ast.NegAtom{atom}
	}
	if ctx.EQ() == nil && ctx.BANGEQ() == nil && ctx.LESS() == nil && ctx.LESSEQ() == nil && ctx.GREATER() == nil && ctx.GREATEREQ() == nil {
		if ok {
			return atom
		}
		p.errors.Add(fmt.Sprintf("parse error: %v", ctx.GetText()), ctx.GetStart().GetLine(), ctx.GetStart().GetColumn())
		return ast.NewAtom("br0ken")
	}
	left := p.Visit(ctx.Term(0)).(ast.Term)
	leftBase, ok := left.(ast.BaseTerm)
	if !ok {
		p.errors.Add(fmt.Sprintf("not a base term: %v", ctx.Term(0).GetText()), ctx.Term(0).GetStart().GetLine(), ctx.Term(0).GetStart().GetColumn())
	}
	right := p.Visit(ctx.Term(1)).(ast.Term)
	rightBase, ok := right.(ast.BaseTerm)
	if !ok {
		p.errors.Add(fmt.Sprintf("not a base term: %v", ctx.Term(1).GetText()), ctx.Term(1).GetStart().GetLine(), ctx.Term(1).GetStart().GetColumn())
	}

	if ctx.EQ() != nil {
		return ast.Eq{leftBase, rightBase}
	}
	if ctx.BANGEQ() != nil {
		return ast.Ineq{leftBase, rightBase}
	}
	if ctx.LESS() != nil {
		return ast.NewAtom(":lt", leftBase, rightBase)
	}
	if ctx.LESSEQ() != nil {
		return ast.NewAtom(":le", leftBase, rightBase)
	}
	if ctx.GREATER() != nil {
		return ast.NewAtom(":gt", leftBase, rightBase)
	}
	return ast.NewAtom(":ge", leftBase, rightBase)
}

// VisitVar visits a parse tree produced by MangleParser#Var.
func (p Parser) VisitVar(ctx *gen.VarContext) any {
	return ast.Variable{ctx.VARIABLE().GetText()}
}

// VisitConst visits a parse tree produced by MangleParser#Const.
func (p Parser) VisitConst(ctx *gen.ConstContext) any {
	nameConstant, err := ast.Name(ctx.CONSTANT().GetText())
	if err != nil {
		p.errors.Add(err.Error(), ctx.GetStart().GetLine(), ctx.GetStart().GetColumn())
		return ast.AnyBound
	}
	return nameConstant
}

// VisitNum visits a parse tree produced by MangleParser#Num.
func (p Parser) VisitNum(ctx *gen.NumContext) any {
	num, err := strconv.ParseInt(ctx.NUMBER().GetText(), 10, 64)
	if err != nil {
		p.errors.Add(err.Error(), ctx.GetStart().GetLine(), ctx.GetStart().GetColumn())
		return ast.Number(-1)
	}
	return ast.Number(num)
}

// VisitFloat visits a parse tree produced by MangleParser#Float.
func (p Parser) VisitFloat(ctx *gen.FloatContext) any {
	floatNum, err := strconv.ParseFloat(ctx.FLOAT().GetText(), 64)
	if err != nil {
		p.errors.Add(err.Error(), ctx.GetStart().GetLine(), ctx.GetStart().GetColumn())
		return ast.Float64(-1)
	}
	return ast.Float64(floatNum)
}

// VisitStr visits a parse tree produced by MangleParser#Str.
func (p Parser) VisitStr(ctx *gen.StrContext) any {
	text := ctx.STRING().GetText()
	text = text[1 : len(text)-1]
	text = strings.ReplaceAll(text, `\"`, `"`)
	text = strings.ReplaceAll(text, `\\`, `\`)
	return ast.String(text)
}

// VisitAppl visits a parse tree produced by MangleParser#Appl.
func (p Parser) VisitAppl(ctx *gen.ApplContext) any {
	name := ctx.NAME().GetText()

	if strings.HasPrefix(name, "fn:") {
		// Either ast.Atom or ast.ApplyFn
		var args []ast.BaseTerm
		for _, e := range ctx.AllTerm() {
			arg := p.Visit(e)
			baseTerm, ok := arg.(ast.BaseTerm)
			if !ok {
				p.errors.Add(fmt.Sprintf("expected base term got %v", arg), e.GetStart().GetLine(), e.GetStart().GetColumn())
				continue
			}
			args = append(args, baseTerm)
		}
		fnSym := ast.FunctionSym{name, len(args)}
		return ast.ApplyFn{fnSym, args}
	}

	var args []ast.BaseTerm
	for _, e := range ctx.AllTerm() {
		arg := p.Visit(e)
		baseTerm, ok := arg.(ast.BaseTerm)
		if !ok {
			p.errors.Add(fmt.Sprintf("expected base term got %v", arg), e.GetStart().GetLine(), e.GetStart().GetColumn())
			args = append(args, ast.AnyBound)
			continue
		}
		args = append(args, baseTerm)
	}
	predicateSym := ast.PredicateSym{ctx.NAME().GetText(), len(args)}
	return ast.Atom{predicateSym, args}
}

// VisitMap visits a parse tree produced by MangleParser#Map.
func (p Parser) VisitMap(ctx *gen.MapContext) any {
	var args []ast.BaseTerm
	termCtxs := ctx.AllTerm()
	for _, e := range termCtxs {
		arg := p.Visit(e)
		baseTerm, ok := arg.(ast.BaseTerm)
		if !ok {
			p.errors.Add(fmt.Sprintf("expected base term got %v", arg), e.GetStart().GetLine(), e.GetStart().GetColumn())
			args = append(args, ast.AnyBound)
			continue
		}
		args = append(args, baseTerm)
	}
	return ast.ApplyFn{ast.FunctionSym{"fn:map", -1}, args}
}

// VisitStruct visits a parse tree produced by MangleParser#Struct.
func (p Parser) VisitStruct(ctx *gen.StructContext) any {
	var args []ast.BaseTerm
	for _, e := range ctx.AllTerm() {
		arg := p.Visit(e)
		baseTerm, ok := arg.(ast.BaseTerm)
		if !ok {
			p.errors.Add(fmt.Sprintf("expected base term got %v", arg), e.GetStart().GetLine(), e.GetStart().GetColumn())
			args = append(args, ast.AnyBound)
			continue
		}
		args = append(args, baseTerm)
	}
	return ast.ApplyFn{ast.FunctionSym{"fn:struct", -1}, args}
}

// VisitList visits a parse tree produced by MangleParser#List.
func (p Parser) VisitList(ctx *gen.ListContext) any {
	var args []ast.BaseTerm
	for _, e := range ctx.AllTerm() {
		arg := p.Visit(e)
		baseTerm, ok := arg.(ast.BaseTerm)
		if !ok {
			p.errors.Add(fmt.Sprintf("expected base term got %v", arg), e.GetStart().GetLine(), e.GetStart().GetColumn())
			args = append(args, ast.AnyBound)
			continue
		}
		args = append(args, baseTerm)
	}
	return ast.ApplyFn{ast.FunctionSym{"fn:list", -1}, args}
}

// VisitAtom visits a parse tree produced by MangleParser#atoms.
func (p Parser) VisitAtom(ctx *gen.AtomContext) any {
	term := p.Visit(ctx.Term())
	atom, ok := term.(ast.Atom)
	if !ok {
		p.errors.Add(fmt.Sprintf("expected atom got %v", term), ctx.GetStart().GetLine(), ctx.GetStart().GetColumn())
		return ast.NewAtom("b0rken")
	}
	return atom
}

// VisitAtoms visits a parse tree produced by MangleParser#atoms.
func (p Parser) VisitAtoms(ctx *gen.AtomsContext) any {
	var atoms []ast.Atom
	for _, e := range ctx.AllAtom() {
		atom := p.Visit(e).(ast.Atom)
		atoms = append(atoms, atom)
	}
	return atoms
}

// PredicateName parses a predicate name.
func PredicateName(s string) (string, error) {
	p, err := newParser(s)
	if err != nil {
		return "", err
	}
	tok := p.lexer.NextToken()
	if err := p.error(); err != nil {
		return "", err
	}
	if tok.GetTokenType() == gen.MangleLexerNAME {
		return tok.GetText(), nil
	}
	return "", nil
}

// Clause parses a single clause.
func Clause(s string) (ast.Clause, error) {
	p, err := newParser(s)
	if err != nil {
		return ast.Clause{}, err
	}
	defer p.reset()

	tree := p.parser.Clause()
	if err := p.error(); err != nil {
		return ast.Clause{}, err
	}
	clause := p.Visit(tree)
	if err := p.error(); err != nil {
		return ast.Clause{}, err
	}
	return clause.(ast.Clause), nil
}

// LiteralOrFormula parses a single Term, an equality or inequality from a given string.
func LiteralOrFormula(s string) (ast.Term, error) {
	p, err := newParser(s)
	if err != nil {
		return nil, err
	}
	defer p.reset()

	tree := p.parser.LiteralOrFml()
	if err := p.error(); err != nil {
		return nil, err
	}
	term := p.Visit(tree)
	if err := p.error(); err != nil {
		return nil, err
	}
	return term.(ast.Term), nil
}

// Term parses a Term from given string.
func Term(s string) (ast.Term, error) {
	p, err := newParser(s)
	if err != nil {
		return nil, err
	}
	defer p.reset()

	tree := p.parser.Term()
	if err := p.error(); err != nil {
		return nil, err
	}
	term := p.Visit(tree)
	if err := p.error(); err != nil {
		return nil, err
	}
	return term.(ast.Term), nil
}

// BaseTerm parses a BaseTerm from given string.
func BaseTerm(s string) (ast.BaseTerm, error) {
	term, err := Term(s)
	if err != nil {
		return nil, err
	}
	baseTerm, ok := term.(ast.BaseTerm)
	if ok {
		return baseTerm, nil
	}
	return nil, fmt.Errorf("not a base term: %v %T", term, term)
}

// Atom parses an Atom from given string.
func Atom(s string) (ast.Atom, error) {
	term, err := Term(s)
	if err != nil {
		return ast.Atom{}, err
	}
	atom, ok := term.(ast.Atom)
	if ok {
		return atom, nil
	}
	return ast.Atom{}, nil // TODO
}

func nameAtom(name string) ast.Atom {
	return ast.Atom{Predicate: ast.PredicateSym{Symbol: "name", Arity: 1}, Args: []ast.BaseTerm{ast.String(name)}}
}

// SyntaxError is called by ANTLR generated code when a syntax error is encountered.
func (p *Parser) SyntaxError(recognizer antlr.Recognizer, offendingSymbol any, line, column int, msg string, e antlr.RecognitionException) {
	p.errors.Add(msg, line, column)
}

// ReportAmbiguity implements error listener interface.
func (p *Parser) ReportAmbiguity(recognizer antlr.Parser, dfa *antlr.DFA, startIndex, stopIndex int, exact bool, ambigAlts *antlr.BitSet, configs antlr.ATNConfigSet) {
	// Intentional
}

// ReportAttemptingFullContext implements error listener interface.
func (p *Parser) ReportAttemptingFullContext(recognizer antlr.Parser, dfa *antlr.DFA, startIndex, stopIndex int, conflictingAlts *antlr.BitSet, configs antlr.ATNConfigSet) {
	// Intentional
}

// ReportContextSensitivity  implements error listener interface.
func (p *Parser) ReportContextSensitivity(recognizer antlr.Parser, dfa *antlr.DFA, startIndex, stopIndex, prediction int, configs antlr.ATNConfigSet) {
	// Intentional
}
