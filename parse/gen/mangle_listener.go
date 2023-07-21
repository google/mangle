// Code generated from java-escape by ANTLR 4.11.1. DO NOT EDIT.

package gen // Mangle
import "github.com/antlr/antlr4/runtime/Go/antlr/v4"

// MangleListener is a complete listener for a parse tree produced by MangleParser.
type MangleListener interface {
	antlr.ParseTreeListener

	// EnterStart is called when entering the start production.
	EnterStart(c *StartContext)

	// EnterProgram is called when entering the program production.
	EnterProgram(c *ProgramContext)

	// EnterPackageDecl is called when entering the packageDecl production.
	EnterPackageDecl(c *PackageDeclContext)

	// EnterUseDecl is called when entering the useDecl production.
	EnterUseDecl(c *UseDeclContext)

	// EnterDecl is called when entering the decl production.
	EnterDecl(c *DeclContext)

	// EnterDescrBlock is called when entering the descrBlock production.
	EnterDescrBlock(c *DescrBlockContext)

	// EnterBoundsBlock is called when entering the boundsBlock production.
	EnterBoundsBlock(c *BoundsBlockContext)

	// EnterConstraintsBlock is called when entering the constraintsBlock production.
	EnterConstraintsBlock(c *ConstraintsBlockContext)

	// EnterClause is called when entering the clause production.
	EnterClause(c *ClauseContext)

	// EnterClauseBody is called when entering the clauseBody production.
	EnterClauseBody(c *ClauseBodyContext)

	// EnterTransform is called when entering the transform production.
	EnterTransform(c *TransformContext)

	// EnterLetStmt is called when entering the letStmt production.
	EnterLetStmt(c *LetStmtContext)

	// EnterLiteralOrFml is called when entering the literalOrFml production.
	EnterLiteralOrFml(c *LiteralOrFmlContext)

	// EnterVar is called when entering the Var production.
	EnterVar(c *VarContext)

	// EnterConst is called when entering the Const production.
	EnterConst(c *ConstContext)

	// EnterNum is called when entering the Num production.
	EnterNum(c *NumContext)

	// EnterFloat is called when entering the Float production.
	EnterFloat(c *FloatContext)

	// EnterStr is called when entering the Str production.
	EnterStr(c *StrContext)

	// EnterAppl is called when entering the Appl production.
	EnterAppl(c *ApplContext)

	// EnterList is called when entering the List production.
	EnterList(c *ListContext)

	// EnterMap is called when entering the Map production.
	EnterMap(c *MapContext)

	// EnterStruct is called when entering the Struct production.
	EnterStruct(c *StructContext)

	// EnterAtom is called when entering the atom production.
	EnterAtom(c *AtomContext)

	// EnterAtoms is called when entering the atoms production.
	EnterAtoms(c *AtomsContext)

	// ExitStart is called when exiting the start production.
	ExitStart(c *StartContext)

	// ExitProgram is called when exiting the program production.
	ExitProgram(c *ProgramContext)

	// ExitPackageDecl is called when exiting the packageDecl production.
	ExitPackageDecl(c *PackageDeclContext)

	// ExitUseDecl is called when exiting the useDecl production.
	ExitUseDecl(c *UseDeclContext)

	// ExitDecl is called when exiting the decl production.
	ExitDecl(c *DeclContext)

	// ExitDescrBlock is called when exiting the descrBlock production.
	ExitDescrBlock(c *DescrBlockContext)

	// ExitBoundsBlock is called when exiting the boundsBlock production.
	ExitBoundsBlock(c *BoundsBlockContext)

	// ExitConstraintsBlock is called when exiting the constraintsBlock production.
	ExitConstraintsBlock(c *ConstraintsBlockContext)

	// ExitClause is called when exiting the clause production.
	ExitClause(c *ClauseContext)

	// ExitClauseBody is called when exiting the clauseBody production.
	ExitClauseBody(c *ClauseBodyContext)

	// ExitTransform is called when exiting the transform production.
	ExitTransform(c *TransformContext)

	// ExitLetStmt is called when exiting the letStmt production.
	ExitLetStmt(c *LetStmtContext)

	// ExitLiteralOrFml is called when exiting the literalOrFml production.
	ExitLiteralOrFml(c *LiteralOrFmlContext)

	// ExitVar is called when exiting the Var production.
	ExitVar(c *VarContext)

	// ExitConst is called when exiting the Const production.
	ExitConst(c *ConstContext)

	// ExitNum is called when exiting the Num production.
	ExitNum(c *NumContext)

	// ExitFloat is called when exiting the Float production.
	ExitFloat(c *FloatContext)

	// ExitStr is called when exiting the Str production.
	ExitStr(c *StrContext)

	// ExitAppl is called when exiting the Appl production.
	ExitAppl(c *ApplContext)

	// ExitList is called when exiting the List production.
	ExitList(c *ListContext)

	// ExitMap is called when exiting the Map production.
	ExitMap(c *MapContext)

	// ExitStruct is called when exiting the Struct production.
	ExitStruct(c *StructContext)

	// ExitAtom is called when exiting the atom production.
	ExitAtom(c *AtomContext)

	// ExitAtoms is called when exiting the atoms production.
	ExitAtoms(c *AtomsContext)
}
