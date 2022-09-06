package lower

import (
	"chaic/ast"
	"chaic/mir"
	"chaic/report"
	"chaic/syntax"
)

func (l *Lowerer) lowerBlock(block *ast.Block) {
	for _, stmt := range block.Stmts {
		var mstmt mir.Statement

		switch v := stmt.(type) {
		case *ast.VarDecl:
			l.lowerVarDecl(v)
			continue
		case *ast.Assignment:
			l.lowerAssignment(v)
			continue
		case *ast.IncDecStmt:
			mstmt = l.lowerIncDecStmt(v)
		case *ast.KeywordStmt:
			switch v.Kind {
			case syntax.TOK_BREAK:
				mstmt = &mir.Break{StmtBase: mir.NewStmtBase(v.Span())}
			case syntax.TOK_CONTINUE:
				mstmt = &mir.Continue{StmtBase: mir.NewStmtBase(v.Span())}
			}
		case *ast.ReturnStmt:
			{
				var expr mir.Expr
				switch len(v.Exprs) {
				case 0:
					// Nothing
				case 1:
					expr = l.lowerExpr(v.Exprs[0])
				default:
					// TODO
				}

				mstmt = &mir.Return{
					StmtBase: mir.NewStmtBase(v.Span()),
					Value:    expr,
				}
			}
		case *ast.IfTree:
			// TODO
		case *ast.WhileLoop:
			// TODO
		case *ast.DoWhileLoop:
			// TODO
		case *ast.CForLoop:
			// TODO
		case ast.ASTExpr:
			mstmt = l.lowerExpr(v)
		default:
			report.ReportICE("lowering for statement not implemented")
		}

		l.appendStmt(mstmt)
	}
}
