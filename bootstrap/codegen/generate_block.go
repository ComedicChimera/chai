package codegen

import (
	"chaic/ast"
	"chaic/report"
)

// generateBlock generates a block of statements.
func (g *Generator) generateBlock(block *ast.Block) {
	for _, stmt := range block.Stmts {
		switch v := stmt.(type) {
		case *ast.IfTree:
			g.generateIfTree(v)

			// If the block we are positioned over after generated the if tree
			// has a terminator, then the if tree never reaches its exit block
			// and so we can just return here: all code after this is deadcode.
			return
		case *ast.VarDecl:
			g.generateVarDecl(v)
		case *ast.Assignment:
			g.generateAssignment(v)
		case *ast.IncDecStmt:
			g.generateIncDec(v)
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

/* -------------------------------------------------------------------------- */

// generateIfTree generates an if statement tree.
func (g *Generator) generateIfTree(ifTree *ast.IfTree) {
	// Generate the exit block of the if statement.
	exitBlock := g.appendBlock()
	exitNdx := len(g.block.Parent.Blocks)

	// Whether the if statement ever reaches the exit block.
	neverExit := true

	// Generate all of the conditional branches.
	for _, condBranch := range ifTree.CondBranches {
		// Generate the header variable if it exists.
		if condBranch.HeaderVarDecl != nil {
			g.generateVarDecl(condBranch.HeaderVarDecl)
		}

		// Generate the condition expression.
		llCond := g.generateExpr(condBranch.Condition)

		// Generate the then and else blocks.
		thenBlock := g.appendBlock()
		elseBlock := g.appendBlock()

		// Generate the conditional branch.
		g.block.NewCondBr(llCond, thenBlock, elseBlock)

		// Generate the body of the if statement.
		g.block = thenBlock
		g.generateBlock(condBranch.Body)

		// Jump to the exit block if the block is not already terminated.
		if g.block.Term == nil {
			g.block.NewBr(exitBlock)
			neverExit = false
		}

		// Position the generator in the else block for the next branch of the
		// if statement if any exists.
		g.block = elseBlock
	}

	// Generate the else block if it exists.
	if ifTree.ElseBranch != nil {
		// We are always already positioned in the else block so we can just
		// generate in place.
		g.generateBlock(ifTree.ElseBranch)

		// Jump to the exit block if the block is not already terminated.
		if g.block.Term == nil {
			g.block.NewBr(exitBlock)
			neverExit = false
		}
	} else {
		// The last else block just jumps to the exit block.  LLVM should be
		// able to remove this from the CFG since LLIR doesn't provide a good
		// way to safely remove it.
		g.block.NewBr(exitBlock)
		neverExit = false
	}

	// If the exit block is never reached, we just remove it.
	if neverExit {
		g.block.Parent.Blocks = append(g.block.Parent.Blocks[:exitNdx], g.block.Parent.Blocks[exitNdx+1:]...)
	} else {
		// Otherwise, we position the generate over it so we can keep generating
		// from the end of the if statement.
		g.block = exitBlock
	}
}
