package codegen

import (
	"chaic/mir"
	"chaic/report"
)

// generateBlock generates a block.
func (g *Generator) generateBlock(block []mir.Statement) {
	for _, stmt := range block {
		switch v := stmt.(type) {
		case *mir.IfTree:
			g.generateIfTree(v)

			// If the block we are positioned over after generated the if tree
			// has a terminator, then the if tree never reaches its exit block
			// and so we can just return here: all code after this is deadcode.
			if g.block.Term != nil {
				return
			}
		case *mir.WhileLoop:
			g.generateWhileLoop(v)
		case *mir.DoWhileLoop:
			g.generateDoWhileLoop(v)
		case *mir.VarDecl:
			g.generateVarDecl(v)
		case *mir.StructInstanceDecl:
			g.generateStructInstDecl(v)
		case *mir.Assignment:
			g.generateAssignment(v)
		case *mir.Return:
			g.generateReturnStmt(v)
			return
		case *mir.Continue:
			g.block.NewBr(g.currLoopContext().continueDest)
			return
		case *mir.Break:
			g.block.NewBr(g.currLoopContext().breakDest)
			return
		case mir.Expr:
			g.generateExpr(v)
		default:
			report.ReportICE("codegen for statement not implemented")
		}
	}
}

// generateIfStmt generates an if tree.
func (g *Generator) generateIfTree(ifTree *mir.IfTree) {
	// Generate the exit block of the if statement.
	exitNdx := len(g.block.Parent.Blocks)
	exitBlock := g.appendBlock()

	// Whether the if statement ever reaches the exit block.
	neverExit := true

	// Generate all of the conditional branches.
	for _, condBranch := range ifTree.CondBranches {
		// Generate the header block if it exists.
		if len(condBranch.HeaderBlock) > 0 {
			g.generateBlock(condBranch.HeaderBlock)
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
	if len(ifTree.ElseBlock) > 0 {
		// We are always already positioned in the else block so we can just
		// generate in place.
		g.generateBlock(ifTree.ElseBlock)

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
func (g *Generator) generateWhileLoop(loop *mir.WhileLoop) {
	// Generate the header block if it exists.  Note that this is placed outside
	// the loop header because we don't want it to run every iteration.
	if len(loop.HeaderBlock) > 0 {
		g.generateBlock(loop.HeaderBlock)
	}

	// Create and jump to the loop header block.
	loopHeader := g.appendBlock()
	g.block.NewBr(loopHeader)

	// Create the loop body and exit blocks.
	loopBody := g.appendBlock()
	loopExit := g.appendBlock()

	// Create the loop update block if one is needed.
	loopUpdate := loopHeader
	if len(loop.UpdateBlock) > 0 {
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
	g.block = loopHeader
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
	if len(loop.UpdateBlock) > 0 {
		g.block = loopUpdate

		g.generateBlock(loop.UpdateBlock)

		// Add the branch back to the loop header.
		g.block.NewBr(loopHeader)
	}

	// Position the generator over the exit block.
	g.block = loopExit
}

// generateDoWhileLoop generates a do-while loop.
func (g *Generator) generateDoWhileLoop(loop *mir.DoWhileLoop) {
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
