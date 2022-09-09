package lower

import (
	"chaic/ast"
	"chaic/mir"
	"chaic/report"
	"chaic/syntax"
	"chaic/util"
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
			mstmt = l.lowerIfTree(v)
		case *ast.WhileLoop:
			mstmt = l.lowerWhileLoop(v)
		case *ast.DoWhileLoop:
			mstmt = l.lowerDoWhileLoop(v)
		case *ast.CForLoop:
			mstmt = l.lowerCForLoop(v)
		case ast.ASTExpr:
			mstmt = l.lowerExpr(v)
		default:
			report.ReportICE("lowering for statement not implemented")
		}

		l.appendStmt(mstmt)
	}
}

/* -------------------------------------------------------------------------- */

// lowerIfTree lowers an if tree expression.
func (l *Lowerer) lowerIfTree(ifTree *ast.IfTree) mir.Statement {
	outerBlock := l.block

	mIfTree := &mir.IfTree{
		StmtBase: mir.NewStmtBase(ifTree.Span()),
		CondBranches: util.Map(ifTree.CondBranches, func(cb ast.CondBranch) mir.CondBranch {
			mcb := mir.CondBranch{}
			l.block = &mcb.HeaderBlock

			if cb.HeaderVarDecl != nil {
				l.lowerVarDecl(cb.HeaderVarDecl)
			}

			mcb.Condition = l.lowerExpr(cb.Condition)

			l.block = &mcb.Body
			l.lowerBlock(cb.Body)

			return mcb
		}),
	}

	if ifTree.ElseBranch != nil {
		l.block = &mIfTree.ElseBlock
		l.lowerBlock(ifTree.ElseBranch)
	}

	l.block = outerBlock

	return mIfTree
}

// lowerWhileLoop lowers a while loop.
func (l *Lowerer) lowerWhileLoop(loop *ast.WhileLoop) mir.Statement {
	outerBlock := l.block

	mLoop := &mir.WhileLoop{StmtBase: mir.NewStmtBase(loop.Span())}

	l.block = &mLoop.HeaderBlock
	if loop.HeaderVarDecl != nil {
		l.lowerVarDecl(loop.HeaderVarDecl)
	}

	mLoop.Condition = l.lowerExpr(loop.Condition)

	l.block = &mLoop.Body
	l.lowerBlock(loop.Body)

	if loop.ElseBlock != nil {
		l.block = &mLoop.ElseBlock
		l.lowerBlock(loop.ElseBlock)
	}

	l.block = outerBlock

	return mLoop
}

// lowerDoWhileLoop lowers a do-while loop.
func (l *Lowerer) lowerDoWhileLoop(loop *ast.DoWhileLoop) mir.Statement {
	outerBlock := l.block

	mLoop := &mir.DoWhileLoop{StmtBase: mir.NewStmtBase(loop.Span())}

	l.block = &mLoop.HeaderBlock
	mLoop.Condition = l.lowerExpr(loop.Condition)

	l.block = &mLoop.Body
	l.lowerBlock(loop.Body)

	if loop.ElseBlock != nil {
		l.block = &mLoop.ElseBlock
		l.lowerBlock(loop.ElseBlock)
	}

	l.block = outerBlock

	return mLoop
}

// lowerCForLoop lowers a C-style for loop.
func (l *Lowerer) lowerCForLoop(loop *ast.CForLoop) mir.Statement {
	outerBlock := l.block

	mLoop := &mir.WhileLoop{StmtBase: mir.NewStmtBase(loop.Span())}

	l.block = &mLoop.HeaderBlock
	if loop.IterVarDecl != nil {
		l.lowerVarDecl(loop.IterVarDecl)
	}

	mLoop.Condition = l.lowerExpr(loop.Condition)

	l.block = &mLoop.Body
	l.lowerBlock(loop.Body)

	if loop.UpdateStmt != nil {
		l.block = &mLoop.UpdateBlock

		switch v := loop.UpdateStmt.(type) {
		case *ast.Assignment:
			l.lowerAssignment(v)
		case *ast.IncDecStmt:
			l.appendStmt(l.lowerIncDecStmt(v))
		case ast.ASTExpr:
			l.appendStmt(l.lowerExpr(v))
		}
	}

	if loop.ElseBlock != nil {
		l.block = &mLoop.ElseBlock
		l.lowerBlock(loop.ElseBlock)
	}

	l.block = outerBlock

	return mLoop
}
