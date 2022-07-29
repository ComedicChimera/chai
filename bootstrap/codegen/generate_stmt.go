package codegen

import (
	"chaic/ast"
	"chaic/report"

	llvalue "github.com/llir/llvm/ir/value"
)

// generateVarDecl generates a variable declaration.
func (g *Generator) generateVarDecl(vd *ast.VarDecl) {
	for _, varList := range vd.VarLists {
		// Allocate all the variables.
		for _, v := range varList.Vars {
			v.Sym.LLValue = g.varBlock.NewAlloca(g.convAllocType(v.Type()))
		}

		// Get the initializer.
		var initExpr llvalue.Value

		if varList.Initializer == nil {
			// TODO: handle more complex null values
			initExpr = g.generateNullValue(varList.Vars[0].Type())
		} else {
			initExpr = g.generateExpr(varList.Initializer)
		}

		// Initialize all the variables.
		for _, v := range varList.Vars {
			g.block.NewStore(initExpr, v.Sym.LLValue)
		}
	}
}

// generateReturnStmt generates a return statement,
func (g *Generator) generateReturnStmt(ret *ast.ReturnStmt) {
	switch len(ret.Exprs) {
	case 0:
		g.block.NewRet(nil)
		return
	case 1:
		g.block.NewRet(g.generateExpr(ret.Exprs[0]))
		return
	default:
		report.ReportICE("codegen for multi-return not implemented")
	}
}
