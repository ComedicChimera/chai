package codegen

import (
	"chaic/ast"
	"chaic/llvm"
	"chaic/report"
)

// generateVarDecl generates a variable declaration.
func (g *Generator) generateVarDecl(vd *ast.VarDecl) {
	for _, varList := range vd.VarLists {
		// Allocate all the variables.
		currBlock := g.irb.Block()
		g.irb.MoveToEnd(g.varBlock)
		for _, v := range varList.Vars {
			v.Sym.LLType = g.convType(v.Type())

			v.Sym.LLValue = g.irb.BuildAlloca(g.convAllocType(v.Type()))
		}
		g.irb.MoveToEnd(currBlock)

		// Get the initializer.
		var initExpr llvm.Value

		if varList.Initializer == nil {
			// TODO: handle more complex null values
			initExpr = llvm.ConstNull(g.convType(varList.Vars[0].Type()))
		} else {
			initExpr = g.generateExpr(varList.Initializer)
		}

		// Initialize all the variables.
		for _, v := range varList.Vars {
			g.irb.BuildStore(initExpr, v.Sym.LLValue)
		}
	}
}

// generateReturnStmt generates a return statement,
func (g *Generator) generateReturnStmt(ret *ast.ReturnStmt) {
	switch len(ret.Exprs) {
	case 0:
		g.irb.BuildRet()
		return
	case 1:
		g.irb.BuildRet(g.generateExpr(ret.Exprs[0]))
		return
	default:
		report.ReportICE("codegen for multi-return not implemented")
	}
}
