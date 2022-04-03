package lower

import (
	"chai/ast"
	"chai/mir"
	"chai/typing"
)

// lowerBlock lowers a do block expression.
func (l *Lowerer) lowerBlock(block *ast.Block) *mir.Block {
	mblock := &mir.Block{
		Stmts:     make([]mir.Stmt, len(block.Stmts)),
		YieldType: typing.Simplify(block.Type()),
		TermMode:  mir.BTNone,
	}

	for i, stmt := range block.Stmts {
		if i == len(block.Stmts)-1 && !typing.IsNothing(stmt.Type()) {
			mblock.Stmts[i] = &mir.SimpleStmt{
				Kind: mir.SSKindYield,
				Arg:  l.lowerExpr(stmt),
			}
		} else {
			mstmt := l.lowerStmt(stmt)
			mblock.Stmts[i] = mstmt

			if mstmt.Term() != mir.BTNone {
				mblock.TermMode = mstmt.Term()

				// TODO: this break may not be necessary if the walker prunes
				// off deadcode.. tbd
				break
			}
		}
	}

	return mblock
}

// lowerIfBlock lowers an if expression into an if block. When used as an
// expression, this should be wrapped in a block.
func (l *Lowerer) lowerIfBlock(aif *ast.IfExpr) mir.Stmt {
	mif := &mir.IfTree{CondBranches: make([]mir.CondBranch, len(aif.CondBranches))}

	for i, cbranch := range aif.CondBranches {
		mBodyExpr := l.lowerExpr(cbranch.Body)
		var cBranchBlock *mir.Block
		if mblock, ok := mBodyExpr.(*mir.Block); ok {
			cBranchBlock = mblock
		} else if typing.IsNothing(mBodyExpr.Type()) {
			cBranchBlock = &mir.Block{
				Stmts: []mir.Stmt{&mir.SimpleStmt{
					Kind: mir.SSKindExpr,
					Arg:  mBodyExpr,
				}},
				YieldType: typing.NothingType(),
				// TODO: figure out how to get the termination mode
				// TermMode:  mBodyExpr,
			}
		} else {
			cBranchBlock = &mir.Block{
				Stmts: []mir.Stmt{&mir.SimpleStmt{
					Kind: mir.SSKindYield,
					Arg:  mBodyExpr,
				}},
				YieldType: mBodyExpr.Type(),
				TermMode:  mir.BTNone,
			}
		}

		mif.CondBranches[i] = mir.CondBranch{
			Condition: l.lowerExpr(cbranch.Cond),
			Body:      cBranchBlock,
		}

		// TODO: update the termination mode
	}

	// TODO: handle else clauses

	return mif
}
