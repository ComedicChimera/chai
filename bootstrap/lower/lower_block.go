package lower

import (
	"chai/ast"
	"chai/mir"
	"chai/typing"
)

// lowerBlock lowers a do block expression.
func (l *Lowerer) lowerBlock(block *ast.Block) *mir.Block {
	l.pushScope()
	defer l.popScope()

	mblock := &mir.Block{YieldType: typing.Simplify(block.Type())}

	for i, stmt := range block.Stmts {
		if i == len(block.Stmts)-1 && !typing.IsNothing(stmt.Type()) {
			mblock.Stmts = append(mblock.Stmts, &mir.SimpleStmt{
				Kind: mir.SSKindYield,
				Arg:  l.lowerExpr(stmt),
			})
		} else {
			l.lowerStmt(mblock, stmt)
		}
	}

	return mblock
}

// lowerIfBlock lowers an if expression into an if block. When used as an
// expression, this should be wrapped in a block.
func (l *Lowerer) lowerIfBlock(aif *ast.IfExpr) mir.Stmt {
	mif := &mir.IfTree{CondBranches: make([]mir.CondBranch, len(aif.CondBranches))}

	for i, cbranch := range aif.CondBranches {
		l.pushScope()

		if cbranch.HeaderVarDecl != nil {
			// TODO: header variables
		}

		mif.CondBranches[i] = mir.CondBranch{
			Condition: l.lowerExpr(cbranch.Cond),
			Body:      l.exprToBlock(cbranch.Body),
		}

		l.popScope()
	}

	if aif.ElseBranch != nil {
		mif.ElseBranch = l.exprToBlock(aif.ElseBranch)
	}

	return mif
}

// lowerWhileLoop lowers a while loop into a loop block.
// NOTE: Does not handle loop generators.
func (l *Lowerer) lowerWhileLoop(awhile *ast.WhileExpr) mir.Stmt {
	if awhile.HeaderVarDecl != nil {
		// TODO: header variables
	}

	mloop := &mir.Loop{
		Condition: l.lowerExpr(awhile.Cond),
		Body:      l.exprToBlock(awhile.Body),
	}

	// TODO: post iteration variables

	return mloop
}

// exprToBlock converts an expressions to a block (ensures that an expression
// always takes the form of a block).
func (l *Lowerer) exprToBlock(expr ast.Expr) *mir.Block {
	mExpr := l.lowerExpr(expr)
	if mblock, ok := mExpr.(*mir.Block); ok {
		return mblock
	} else if typing.IsNothing(mExpr.Type()) {
		return &mir.Block{
			Stmts: []mir.Stmt{&mir.SimpleStmt{
				Kind: mir.SSKindExpr,
				Arg:  mExpr,
			}},
			YieldType: typing.NothingType(),
		}
	} else {
		return &mir.Block{
			Stmts: []mir.Stmt{&mir.SimpleStmt{
				Kind: mir.SSKindYield,
				Arg:  mExpr,
			}},
			YieldType: mExpr.Type(),
		}
	}
}
