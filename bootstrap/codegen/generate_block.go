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
		case *ast.WhileLoop:
			g.generateWhileLoop(v)
		case *ast.CForLoop:
			g.generateCForLoop(v)
		case *ast.DoWhileLoop:
			g.generateDoWhileLoop(v)
		case *ast.VarDecl:
			g.generateVarDecl(v)
		case *ast.Assignment:
			g.generateAssignment(v)
		case *ast.IncDecStmt:
			g.generateIncDec(v)
		case *ast.ReturnStmt:
			g.generateReturnStmt(v)
			return
		case *ast.KeywordStmt:
			g.generateKeywordStmt(v)
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
		// Otherwise, we position the generator over the exit block.
		g.block = exitBlock
	}
}

// generateWhileLoop generates a while loop.
func (g *Generator) generateWhileLoop(loop *ast.WhileLoop) {
	// Create and jump to the loop header block.
	loopHeader := g.appendBlock()
	g.block.NewBr(loopHeader)

	// Create the loop body and exit blocks.
	loopBody := g.appendBlock()
	loopExit := g.appendBlock()

	// Generate the loop else block if it exists.
	loopElse := loopExit
	if loop.ElseBlock != nil {
		loopElse = g.appendBlock()

		g.block = loopElse

		g.generateBlock(loop.ElseBlock)

		// Only create a terminator to the else block if one is needed.
		if loopElse.Term == nil {
			g.block.NewBr(loopExit)
		}
	}

	// Generate the header variable declaration if it exists.  Note that we put
	// this inside the loop header because we want it to run every iteration of
	// the loop.
	if loop.HeaderVarDecl != nil {
		g.generateVarDecl(loop.HeaderVarDecl)
	}

	// Generate the main loop conditional branch.
	llCond := g.generateExpr(loop.Condition)
	g.block.NewCondBr(llCond, loopBody, loopElse)

	// Generate the loop body.
	g.pushLoopContext(loopExit, loopHeader)

	g.block = loopBody
	g.generateBlock(loop.Body)

	// We only want to add the jump back to the loop header if the loop body
	// itself is not already terminated.
	if g.block.Term == nil {
		g.block.NewBr(loopHeader)
	}

	g.popLoopContext()

	// Position the generator over the exit block.
	g.block = loopExit
}

// generateCForLoop generates a C-style for loop.
func (g *Generator) generateCForLoop(loop *ast.CForLoop) {
	// Generate the iterator variable declaration if it exists.  Note that this
	// is placed outside the loop header because we don't want it to run every
	// iteration.
	if loop.IterVarDecl != nil {
		g.generateVarDecl(loop.IterVarDecl)
	}

	// Create and jump to the loop header block.
	loopHeader := g.appendBlock()
	g.block.NewBr(loopHeader)

	// Create the loop body and exit blocks.
	loopBody := g.appendBlock()
	loopExit := g.appendBlock()

	// Create the loop update block if one is needed.
	loopUpdate := loopHeader
	if loop.UpdateStmt != nil {
		loopUpdate = g.appendBlock()
	}

	// Generate the loop else block if it exists.
	loopElse := loopExit
	if loop.ElseBlock != nil {
		loopElse = g.appendBlock()

		g.block = loopElse

		g.generateBlock(loop.ElseBlock)

		// Only create a terminator to the else block if one is needed.
		if loopElse.Term == nil {
			g.block.NewBr(loopExit)
		}
	}

	// Generate the main loop conditional branch.
	llCond := g.generateExpr(loop.Condition)
	g.block.NewCondBr(llCond, loopBody, loopElse)

	// Generate the loop body.
	g.pushLoopContext(loopExit, loopHeader)

	g.block = loopBody
	g.generateBlock(loop.Body)

	// We only want to add the jump back to the loop header if the loop body
	// itself is not already terminated.
	if g.block.Term == nil {
		g.block.NewBr(loopUpdate)
	}

	g.popLoopContext()

	// Generate the loop update if it exists.
	if loop.UpdateStmt != nil {
		g.block = loopUpdate

		switch v := loop.UpdateStmt.(type) {
		case *ast.Assignment:
			g.generateAssignment(v)
		case *ast.IncDecStmt:
			g.generateIncDec(v)
		case ast.ASTExpr:
			g.generateExpr(v)
		default:
			report.ReportICE("invalid tripartite for loop update statement detected in codegen")
		}

		// Add the branch back to the loop header.
		g.block.NewBr(loopHeader)
	}

	// Position the generator over the exit block.
	g.block = loopExit
}

// generateDoWhileLoop generates a do-while loop.
func (g *Generator) generateDoWhileLoop(loop *ast.DoWhileLoop) {
	// Create the loop body, footer, and exit blocks.
	loopBody := g.appendBlock()
	loopFooter := g.appendBlock()
	loopExit := g.appendBlock()

	// Generate the loop body.
	g.pushLoopContext(loopExit, loopFooter)

	g.block = loopBody
	g.generateBlock(loop.Body)

	// We only want to add the jump down to the loop footer if the loop body
	// itself is not already terminated.
	if g.block.Term == nil {
		g.block.NewBr(loopFooter)
	}

	g.popLoopContext()

	// Generate the loop else block if it exists.
	loopElse := loopExit
	if loop.ElseBlock != nil {
		g.block = loopElse

		g.generateBlock(loop.ElseBlock)

		// Only create a terminator to the else block if one is needed.
		if loopElse.Term == nil {
			g.block.NewBr(loopExit)
		}
	}

	// Generate the loop footer condition.
	g.block = loopFooter
	llCond := g.generateExpr(loop.Condition)
	g.block.NewCondBr(llCond, loopBody, loopElse)

	// Position the generator over the exit block.
	g.block = loopExit
}
