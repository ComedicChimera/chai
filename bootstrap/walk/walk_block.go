package walk

import (
	"chaic/ast"
	"chaic/report"
	"chaic/types"
)

// Enumeration of different control flow modes.  Control modes indicate how a
// block or statement affects the flow of the program.
const (
	ControlNone   = iota // No control flow change.
	ControlLoop          // Jump within a loop.
	ControlReturn        // Function return
	ControlNoExit        // Control never exits the block (eg. infinite loop)
)

// updateBranchControl produces a new control mode from the combination of a
// preexisting control mode with a new control mode assuming the control modes
// occur on different branches of execution.
func updateBranchControl(initial, next int) int {
	if initial < next {
		return initial
	}

	return next
}

// catchControl blocks to propagation of a specific control mode.
func catchControl(initial, block int) int {
	if initial == block {
		return ControlNone
	}

	return initial
}

/* -------------------------------------------------------------------------- */

// walkBlock walks an AST block and returns the control mode of the block.
func (w *Walker) walkBlock(block *ast.Block) int {
	// Keep track of where any deadcode if any starts.
	var deadcodeStart ast.ASTNode

	// The control mode of the block.
	mode := ControlNone

	for _, stmt := range block.Stmts {
		// Set the deadcode marker if there is control flow which occurs before
		// this statement in the block.
		if mode != ControlNone {
			deadcodeStart = stmt
		}

		// The new control mode produced by the statement.
		newMode := ControlNone

		switch v := stmt.(type) {
		case *ast.VarDecl:
			w.walkVarDecl(v)
		case *ast.Assignment:
			w.walkAssign(v)
		case *ast.IncDecStmt:
			w.walkIncDec(v)
		case *ast.KeywordStmt:
			newMode = w.walkKeywordStmt(v)
		case *ast.ReturnStmt:
			w.walkReturnStmt(v)
			newMode = ControlReturn
		case *ast.IfTree:
			newMode = w.walkIfTree(v)
		case *ast.WhileLoop:
			newMode = w.walkWhileLoop(v)
		case *ast.CForLoop:
			newMode = w.walkCForLoop(v)
		case *ast.DoWhileLoop:
			newMode = w.walkDoWhileLoop(v)
		default: // Must be an expression.
			w.walkExpr(stmt.(ast.ASTExpr))
		}

		// Update the block control mode.
		if mode == ControlNone {
			mode = newMode
		}
	}

	// Emit a deadcode warning as necessary.
	if deadcodeStart != nil {
		w.warn(
			report.NewSpanOver(deadcodeStart.Span(), block.Stmts[len(block.Stmts)-1].Span()),
			"unreachable code detected",
		)
	}

	return mode
}

// -----------------------------------------------------------------------------

// walkIfTree walks an if tree statement.
func (w *Walker) walkIfTree(ifTree *ast.IfTree) int {
	// The overall control mode of the if tree.
	mode := ControlNone

	// Walk all the conditional branches.
	for i, condBranch := range ifTree.CondBranches {
		w.pushScope()

		// Walk the header variable declaration.
		if condBranch.HeaderVarDecl != nil {
			w.walkVarDecl(condBranch.HeaderVarDecl)
		}

		// Walk the condition and make sure it is a boolean.
		w.walkExpr(condBranch.Condition)
		w.mustUnify(types.PrimTypeBool, condBranch.Condition.Type(), condBranch.Condition.Span())

		// Walk the block and update the overall control mode.
		newMode := w.walkBlock(condBranch.Body)
		if i == 0 {
			mode = newMode
		} else {
			mode = updateBranchControl(mode, newMode)
		}

		w.popScope()
	}

	// Walk the else branch if it exists.
	if ifTree.ElseBranch != nil {
		w.pushScope()

		// Update the overall control mode based on the else block.
		mode = updateBranchControl(mode, w.walkBlock(ifTree.ElseBranch))

		w.popScope()
	} else {
		// If there is no else block, then there is no guarantee that any
		// control flow code will run and so the resulting mode is always none.
		return ControlNone
	}

	return mode
}

// walkWhileLoop walks a while loop.
func (w *Walker) walkWhileLoop(loop *ast.WhileLoop) int {
	w.pushScope()

	// Walk the header variable declaration if it exists.
	if loop.HeaderVarDecl != nil {
		w.walkVarDecl(loop.HeaderVarDecl)
	}

	// Walk the condition and make sure it is a boolean.
	w.walkExpr(loop.Condition)
	w.mustUnify(types.PrimTypeBool, loop.Condition.Type(), loop.Condition.Span())

	// Walk the loop body.
	w.loopDepth++
	mode := catchControl(w.walkBlock(loop.Body), ControlLoop)
	w.loopDepth--

	w.popScope()

	// Walk the else block if it exists.
	if loop.ElseBlock != nil {
		w.pushScope()
		mode = updateBranchControl(mode, w.walkBlock(loop.ElseBlock))
		w.popScope()
	} else {
		// If there is no else block, then there is no guarantee that any
		// control flow code will run and so the resulting mode is always none.
		return ControlNone
	}

	return mode
}

// walkCForLoop walks a C-style for loop.
func (w *Walker) walkCForLoop(loop *ast.CForLoop) int {
	w.pushScope()

	// Walk the iterator variable declaration if it exists.
	if loop.IterVarDecl != nil {
		w.walkVarDecl(loop.IterVarDecl)
	}

	// Walk the condition and make sure it is a boolean.
	w.walkExpr(loop.Condition)
	w.mustUnify(types.PrimTypeBool, loop.Condition.Type(), loop.Condition.Span())

	// Walk the iterator update statement if it exists.
	if loop.UpdateStmt != nil {
		switch v := loop.UpdateStmt.(type) {
		case *ast.Assignment:
			w.walkAssign(v)
		case *ast.IncDecStmt:
			w.walkIncDec(v)
		default: // Must be an expression.
			w.walkExpr(v.(ast.ASTExpr))
		}
	}

	// Walk the loop body.
	w.loopDepth++
	mode := catchControl(w.walkBlock(loop.Body), ControlLoop)
	w.loopDepth--

	w.popScope()

	// Walk the else block if it exists.
	if loop.ElseBlock != nil {
		w.pushScope()
		mode = updateBranchControl(mode, w.walkBlock(loop.ElseBlock))
		w.popScope()
	} else {
		// If there is no else block, then there is no guarantee that any
		// control flow code will run and so the resulting mode is always none.
		return ControlNone
	}

	return mode
}

// walkDoWhileLoop walks a do-while loop.
func (w *Walker) walkDoWhileLoop(loop *ast.DoWhileLoop) int {
	// Walk the body of the while loop.
	w.pushScope()
	w.loopDepth++

	mode := catchControl(w.walkBlock(loop.Body), ControlLoop)

	w.loopDepth--
	w.popScope()

	// Walk the condition and make sure it is a boolean.
	w.walkExpr(loop.Condition)
	w.mustUnify(types.PrimTypeBool, loop.Condition.Type(), loop.Condition.Span())

	// Walk the else block if it exists.
	if loop.ElseBlock != nil {
		w.pushScope()
		mode = updateBranchControl(mode, w.walkBlock(loop.ElseBlock))
		w.popScope()
	}

	// The code of the body is always guaranteed to run so we cannot accumulate
	// the overall control mode to `None` in the absense of an else block.
	return mode
}
