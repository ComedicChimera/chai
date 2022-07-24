package walk

import (
	"chaic/ast"
	"chaic/report"
	"chaic/types"
)

// Enumeration of different control flow modes.  Control modes indicate how a
// block or statement affects the flow of the program.
const (
	ControlNone     = iota // No control flow change.
	ControlBreak           // Loop break
	ControlContinue        // Loop continue
	ControlReturn          // Loop return
	ControlNoExit          // Control never exits the block (eg. infinite loop)
)

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

		switch v := stmt.(type) {
		case *ast.VarDecl:
			w.walkVarDecl(v)
		case *ast.Assignment:
			w.walkAssign(v)
		case *ast.IncDecStmt:
			w.walkIncDec(v)
		case *ast.KeywordStmt:
			newMode := w.walkKeywordStmt(v)

			if mode == ControlNone {
				mode = newMode
			}
		case *ast.ReturnStmt:
			w.walkReturnStmt(v)

			if mode == ControlNone {
				mode = ControlReturn
			}
		case *ast.IfTree:
			w.walkIfTree(v)
		case *ast.WhileLoop:
			w.walkWhileLoop(v)
		case *ast.CForLoop:
			w.walkCForLoop(v)
		default: // Must be an expression.
			w.walkExpr(stmt.(ast.ASTExpr))
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
func (w *Walker) walkIfTree(ifTree *ast.IfTree) {
	// Walk all the conditional branches.
	for _, condBranch := range ifTree.CondBranches {
		w.pushScope()

		// Walk the header variable declaration.
		if condBranch.HeaderVarDecl != nil {
			w.walkVarDecl(condBranch.HeaderVarDecl)
		}

		// Walk the condition and make sure it is a boolean.
		w.walkExpr(condBranch.Condition)
		w.mustUnify(types.PrimTypeBool, condBranch.Condition.Type(), condBranch.Condition.Span())

		w.walkBlock(condBranch.Body)

		w.popScope()
	}

	// Walk the else branch if it exists.
	if ifTree.ElseBranch != nil {
		w.pushScope()

		w.walkBlock(ifTree.ElseBranch)

		w.popScope()
	}
}

// walkWhileLoop walks a while loop.
func (w *Walker) walkWhileLoop(loop *ast.WhileLoop) {
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
	w.walkBlock(loop.Body)
	w.loopDepth--

	w.popScope()

	// Walk the else block if it exists.
	if loop.ElseBlock != nil {
		w.pushScope()
		w.walkBlock(loop.ElseBlock)
		w.popScope()
	}
}

// walkCForLoop walks a C-style for loop.
func (w *Walker) walkCForLoop(loop *ast.CForLoop) {
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
	w.walkBlock(loop.Body)
	w.loopDepth--

	w.popScope()

	// Walk the else block if it exists.
	if loop.ElseBlock != nil {
		w.pushScope()
		w.walkBlock(loop.ElseBlock)
		w.popScope()
	}
}
