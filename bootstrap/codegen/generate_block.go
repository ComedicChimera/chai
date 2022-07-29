package codegen

import (
	"chaic/ast"
	"chaic/report"
)

// generateBlock generates a block of statements.
func (g *Generator) generateBlock(block *ast.Block) {
	for _, stmt := range block.Stmts {
		switch v := stmt.(type) {
		case *ast.VarDecl:
			g.generateVarDecl(v)
		case *ast.ReturnStmt:
			g.generateReturnStmt(v)
			return
		case ast.ASTExpr:
			g.generateExpr(v)
		default:
			report.ReportICE("codegen for statement not implemented")
		}
	}
}
