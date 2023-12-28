// Code generated from parse/gen/Mangle.g4 by ANTLR 4.13.1. DO NOT EDIT.

package gen // Mangle
import "github.com/antlr4-go/antlr/v4"

// A complete Visitor for a parse tree produced by MangleParser.
type MangleVisitor interface {
	antlr.ParseTreeVisitor

	// Visit a parse tree produced by MangleParser#start.
	VisitStart(ctx *StartContext) interface{}

	// Visit a parse tree produced by MangleParser#program.
	VisitProgram(ctx *ProgramContext) interface{}

	// Visit a parse tree produced by MangleParser#packageDecl.
	VisitPackageDecl(ctx *PackageDeclContext) interface{}

	// Visit a parse tree produced by MangleParser#useDecl.
	VisitUseDecl(ctx *UseDeclContext) interface{}

	// Visit a parse tree produced by MangleParser#decl.
	VisitDecl(ctx *DeclContext) interface{}

	// Visit a parse tree produced by MangleParser#descrBlock.
	VisitDescrBlock(ctx *DescrBlockContext) interface{}

	// Visit a parse tree produced by MangleParser#boundsBlock.
	VisitBoundsBlock(ctx *BoundsBlockContext) interface{}

	// Visit a parse tree produced by MangleParser#constraintsBlock.
	VisitConstraintsBlock(ctx *ConstraintsBlockContext) interface{}

	// Visit a parse tree produced by MangleParser#clause.
	VisitClause(ctx *ClauseContext) interface{}

	// Visit a parse tree produced by MangleParser#clauseBody.
	VisitClauseBody(ctx *ClauseBodyContext) interface{}

	// Visit a parse tree produced by MangleParser#transform.
	VisitTransform(ctx *TransformContext) interface{}

	// Visit a parse tree produced by MangleParser#letStmt.
	VisitLetStmt(ctx *LetStmtContext) interface{}

	// Visit a parse tree produced by MangleParser#literalOrFml.
	VisitLiteralOrFml(ctx *LiteralOrFmlContext) interface{}

	// Visit a parse tree produced by MangleParser#Var.
	VisitVar(ctx *VarContext) interface{}

	// Visit a parse tree produced by MangleParser#Const.
	VisitConst(ctx *ConstContext) interface{}

	// Visit a parse tree produced by MangleParser#Num.
	VisitNum(ctx *NumContext) interface{}

	// Visit a parse tree produced by MangleParser#Float.
	VisitFloat(ctx *FloatContext) interface{}

	// Visit a parse tree produced by MangleParser#Str.
	VisitStr(ctx *StrContext) interface{}

	// Visit a parse tree produced by MangleParser#BStr.
	VisitBStr(ctx *BStrContext) interface{}

	// Visit a parse tree produced by MangleParser#Appl.
	VisitAppl(ctx *ApplContext) interface{}

	// Visit a parse tree produced by MangleParser#List.
	VisitList(ctx *ListContext) interface{}

	// Visit a parse tree produced by MangleParser#Map.
	VisitMap(ctx *MapContext) interface{}

	// Visit a parse tree produced by MangleParser#Struct.
	VisitStruct(ctx *StructContext) interface{}

	// Visit a parse tree produced by MangleParser#atom.
	VisitAtom(ctx *AtomContext) interface{}

	// Visit a parse tree produced by MangleParser#atoms.
	VisitAtoms(ctx *AtomsContext) interface{}
}
