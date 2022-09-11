package codegen

import (
	"chaic/mir"

	"github.com/llir/llvm/ir/constant"
	lltypes "github.com/llir/llvm/ir/types"
)

// generateVarDecl generates a new variable declaration.
func (g *Generator) generateVarDecl(vd *mir.VarDecl) {
	// TODO: RVO

	llInit := g.generateExpr(vd.Initializer)

	if vd.Temporary {
		vd.Ident.Symbol.LLValue = llInit
	} else {
		varAlloca := g.varBlock.NewAlloca(g.convAllocType(vd.Ident.Type()))
		vd.Ident.Symbol.LLValue = varAlloca

		// TODO: move semantics
		g.copyInto(vd.Ident.Type(), llInit, varAlloca)
	}
}

// generateStructInstDecl generates a new struct instance declaration.
func (g *Generator) generateStructInstDecl(sid *mir.StructInstanceDecl) {
	// TODO: RVO

	structAllocType := g.convAllocType(sid.Ident.Type())
	structAlloca := g.varBlock.NewAlloca(structAllocType)

	for i, fieldInit := range sid.FieldInits {
		llFieldInit := g.generateExpr(fieldInit)

		fieldPtr := g.block.NewGetElementPtr(
			structAllocType,
			structAlloca,
			constant.NewInt(lltypes.I64, 0),
			constant.NewInt(lltypes.I32, int64(i)),
		)

		// TODO: move semantics
		g.copyInto(fieldInit.Type(), llFieldInit, fieldPtr)
	}
}

// generateAssignment generates an assignment.
func (g *Generator) generateAssignment(assign *mir.Assignment) {
	lhsExpr := g.generateLHSExpr(assign.LHS)
	rhsExpr := g.generateExpr(assign.RHS)

	// TODO: move semantics
	g.copyInto(assign.LHS.Type(), rhsExpr, lhsExpr)
}

// generateReturnStmt generates a return statement.
func (g *Generator) generateReturnStmt(ret *mir.Return) {
	if ret.Value == nil {
		g.block.NewRet(nil)
	} else if g.retParam == nil {
		g.block.NewRet(g.generateExpr(ret.Value))
	} else {
		// TODO: RVO, move semantics
		g.copyInto(ret.Value.Type(), g.generateExpr(ret.Value), g.retParam)
		g.block.NewRet(nil)
	}
}
