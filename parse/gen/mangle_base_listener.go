// Code generated from ../parse/gen/Mangle.g4 by ANTLR 4.13.1. DO NOT EDIT.

package gen // Mangle
import "github.com/antlr4-go/antlr/v4"

// BaseMangleListener is a complete listener for a parse tree produced by MangleParser.
type BaseMangleListener struct{}

var _ MangleListener = &BaseMangleListener{}

// VisitTerminal is called when a terminal node is visited.
func (s *BaseMangleListener) VisitTerminal(node antlr.TerminalNode) {}

// VisitErrorNode is called when an error node is visited.
func (s *BaseMangleListener) VisitErrorNode(node antlr.ErrorNode) {}

// EnterEveryRule is called when any rule is entered.
func (s *BaseMangleListener) EnterEveryRule(ctx antlr.ParserRuleContext) {}

// ExitEveryRule is called when any rule is exited.
func (s *BaseMangleListener) ExitEveryRule(ctx antlr.ParserRuleContext) {}

// EnterStart is called when production start is entered.
func (s *BaseMangleListener) EnterStart(ctx *StartContext) {}

// ExitStart is called when production start is exited.
func (s *BaseMangleListener) ExitStart(ctx *StartContext) {}

// EnterProgram is called when production program is entered.
func (s *BaseMangleListener) EnterProgram(ctx *ProgramContext) {}

// ExitProgram is called when production program is exited.
func (s *BaseMangleListener) ExitProgram(ctx *ProgramContext) {}

// EnterPackageDecl is called when production packageDecl is entered.
func (s *BaseMangleListener) EnterPackageDecl(ctx *PackageDeclContext) {}

// ExitPackageDecl is called when production packageDecl is exited.
func (s *BaseMangleListener) ExitPackageDecl(ctx *PackageDeclContext) {}

// EnterUseDecl is called when production useDecl is entered.
func (s *BaseMangleListener) EnterUseDecl(ctx *UseDeclContext) {}

// ExitUseDecl is called when production useDecl is exited.
func (s *BaseMangleListener) ExitUseDecl(ctx *UseDeclContext) {}

// EnterDecl is called when production decl is entered.
func (s *BaseMangleListener) EnterDecl(ctx *DeclContext) {}

// ExitDecl is called when production decl is exited.
func (s *BaseMangleListener) ExitDecl(ctx *DeclContext) {}

// EnterDescrBlock is called when production descrBlock is entered.
func (s *BaseMangleListener) EnterDescrBlock(ctx *DescrBlockContext) {}

// ExitDescrBlock is called when production descrBlock is exited.
func (s *BaseMangleListener) ExitDescrBlock(ctx *DescrBlockContext) {}

// EnterBoundsBlock is called when production boundsBlock is entered.
func (s *BaseMangleListener) EnterBoundsBlock(ctx *BoundsBlockContext) {}

// ExitBoundsBlock is called when production boundsBlock is exited.
func (s *BaseMangleListener) ExitBoundsBlock(ctx *BoundsBlockContext) {}

// EnterConstraintsBlock is called when production constraintsBlock is entered.
func (s *BaseMangleListener) EnterConstraintsBlock(ctx *ConstraintsBlockContext) {}

// ExitConstraintsBlock is called when production constraintsBlock is exited.
func (s *BaseMangleListener) ExitConstraintsBlock(ctx *ConstraintsBlockContext) {}

// EnterClause is called when production clause is entered.
func (s *BaseMangleListener) EnterClause(ctx *ClauseContext) {}

// ExitClause is called when production clause is exited.
func (s *BaseMangleListener) ExitClause(ctx *ClauseContext) {}

// EnterClauseBody is called when production clauseBody is entered.
func (s *BaseMangleListener) EnterClauseBody(ctx *ClauseBodyContext) {}

// ExitClauseBody is called when production clauseBody is exited.
func (s *BaseMangleListener) ExitClauseBody(ctx *ClauseBodyContext) {}

// EnterTransform is called when production transform is entered.
func (s *BaseMangleListener) EnterTransform(ctx *TransformContext) {}

// ExitTransform is called when production transform is exited.
func (s *BaseMangleListener) ExitTransform(ctx *TransformContext) {}

// EnterLetStmt is called when production letStmt is entered.
func (s *BaseMangleListener) EnterLetStmt(ctx *LetStmtContext) {}

// ExitLetStmt is called when production letStmt is exited.
func (s *BaseMangleListener) ExitLetStmt(ctx *LetStmtContext) {}

// EnterLiteralOrFml is called when production literalOrFml is entered.
func (s *BaseMangleListener) EnterLiteralOrFml(ctx *LiteralOrFmlContext) {}

// ExitLiteralOrFml is called when production literalOrFml is exited.
func (s *BaseMangleListener) ExitLiteralOrFml(ctx *LiteralOrFmlContext) {}

// EnterVar is called when production Var is entered.
func (s *BaseMangleListener) EnterVar(ctx *VarContext) {}

// ExitVar is called when production Var is exited.
func (s *BaseMangleListener) ExitVar(ctx *VarContext) {}

// EnterConst is called when production Const is entered.
func (s *BaseMangleListener) EnterConst(ctx *ConstContext) {}

// ExitConst is called when production Const is exited.
func (s *BaseMangleListener) ExitConst(ctx *ConstContext) {}

// EnterNum is called when production Num is entered.
func (s *BaseMangleListener) EnterNum(ctx *NumContext) {}

// ExitNum is called when production Num is exited.
func (s *BaseMangleListener) ExitNum(ctx *NumContext) {}

// EnterFloat is called when production Float is entered.
func (s *BaseMangleListener) EnterFloat(ctx *FloatContext) {}

// ExitFloat is called when production Float is exited.
func (s *BaseMangleListener) ExitFloat(ctx *FloatContext) {}

// EnterStr is called when production Str is entered.
func (s *BaseMangleListener) EnterStr(ctx *StrContext) {}

// ExitStr is called when production Str is exited.
func (s *BaseMangleListener) ExitStr(ctx *StrContext) {}

// EnterAppl is called when production Appl is entered.
func (s *BaseMangleListener) EnterAppl(ctx *ApplContext) {}

// ExitAppl is called when production Appl is exited.
func (s *BaseMangleListener) ExitAppl(ctx *ApplContext) {}

// EnterList is called when production List is entered.
func (s *BaseMangleListener) EnterList(ctx *ListContext) {}

// ExitList is called when production List is exited.
func (s *BaseMangleListener) ExitList(ctx *ListContext) {}

// EnterMap is called when production Map is entered.
func (s *BaseMangleListener) EnterMap(ctx *MapContext) {}

// ExitMap is called when production Map is exited.
func (s *BaseMangleListener) ExitMap(ctx *MapContext) {}

// EnterStruct is called when production Struct is entered.
func (s *BaseMangleListener) EnterStruct(ctx *StructContext) {}

// ExitStruct is called when production Struct is exited.
func (s *BaseMangleListener) ExitStruct(ctx *StructContext) {}

// EnterAtom is called when production atom is entered.
func (s *BaseMangleListener) EnterAtom(ctx *AtomContext) {}

// ExitAtom is called when production atom is exited.
func (s *BaseMangleListener) ExitAtom(ctx *AtomContext) {}

// EnterAtoms is called when production atoms is entered.
func (s *BaseMangleListener) EnterAtoms(ctx *AtomsContext) {}

// ExitAtoms is called when production atoms is exited.
func (s *BaseMangleListener) ExitAtoms(ctx *AtomsContext) {}
