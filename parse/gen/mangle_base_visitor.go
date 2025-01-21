// Code generated from parse/gen/Mangle.g4 by ANTLR 4.13.1. DO NOT EDIT.

package gen // Mangle
import "github.com/antlr4-go/antlr/v4"

type BaseMangleVisitor struct {
	*antlr.BaseParseTreeVisitor
}

func (v *BaseMangleVisitor) VisitStart(ctx *StartContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseMangleVisitor) VisitProgram(ctx *ProgramContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseMangleVisitor) VisitPackageDecl(ctx *PackageDeclContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseMangleVisitor) VisitUseDecl(ctx *UseDeclContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseMangleVisitor) VisitDecl(ctx *DeclContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseMangleVisitor) VisitDescrBlock(ctx *DescrBlockContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseMangleVisitor) VisitBoundsBlock(ctx *BoundsBlockContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseMangleVisitor) VisitConstraintsBlock(ctx *ConstraintsBlockContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseMangleVisitor) VisitClause(ctx *ClauseContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseMangleVisitor) VisitClauseBody(ctx *ClauseBodyContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseMangleVisitor) VisitTransform(ctx *TransformContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseMangleVisitor) VisitLetStmt(ctx *LetStmtContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseMangleVisitor) VisitLiteralOrFml(ctx *LiteralOrFmlContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseMangleVisitor) VisitVar(ctx *VarContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseMangleVisitor) VisitConst(ctx *ConstContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseMangleVisitor) VisitNum(ctx *NumContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseMangleVisitor) VisitFloat(ctx *FloatContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseMangleVisitor) VisitStr(ctx *StrContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseMangleVisitor) VisitBStr(ctx *BStrContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseMangleVisitor) VisitList(ctx *ListContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseMangleVisitor) VisitMap(ctx *MapContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseMangleVisitor) VisitStruct(ctx *StructContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseMangleVisitor) VisitDotType(ctx *DotTypeContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseMangleVisitor) VisitAppl(ctx *ApplContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseMangleVisitor) VisitMember(ctx *MemberContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseMangleVisitor) VisitAtom(ctx *AtomContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseMangleVisitor) VisitAtoms(ctx *AtomsContext) interface{} {
	return v.VisitChildren(ctx)
}
