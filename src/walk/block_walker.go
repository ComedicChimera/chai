package walk

import (
	"chai/logging"
	"chai/sem"
	"chai/syntax"
	"chai/typing"
)

// walkBlockExpr walks a block expression (`block_expr`)
func (w *Walker) walkBlockExpr(branch *syntax.ASTBranch, yieldsValue bool) (sem.HIRExpr, bool) {
	blockExpr := branch.BranchAt(0)

	switch blockExpr.Name {
	case "if_chain":
		return w.walkIfChain(branch, yieldsValue)
	case "while_loop":
		return w.walkWhileLoop(branch, yieldsValue)
	}

	// unreachable
	return nil, false
}

// walkIfChain walks an `if_chain` branch
func (w *Walker) walkIfChain(branch *syntax.ASTBranch, yieldsValue bool) (sem.HIRExpr, bool) {
	ifChain := &sem.HIRIfChain{
		ExprBase: sem.NewExprBase(nothingType(), sem.RValue, false),
	}

	for _, item := range branch.Content {
		// only items are subbranches
		itembranch := item.(*syntax.ASTBranch)

		switch itembranch.Name {
		case "if_block":
		case "elif_block":
		case "else_block":
		}
	}

	// TODO: fix this broken control flow logic

	// determine the resultant type of the if chain if it yields a value, and
	// the control flow effect of the block
	if yieldsValue {
		// TODO
	} else /* only need to find control flow */ {
		control := ifChain.IfBranch.BranchBody.Control()

		// never going to have a control flow effect unless the first branch
		// actually causes a change in control flow
		if control != sem.CFNone {
			for _, elifBranch := range ifChain.ElifBranches {
				if elifBranch.BranchBody.Control() == sem.CFNone {
					control = sem.CFNone
					break
				} else {

				}
			}

			if ifChain.ElseBranch != nil && control != sem.CFNone {
				control = ifChain.ElseBranch.Control()
			}
		}

		ifChain.SetControl(control)
	}

	// return the final constructed if chain
	return ifChain, true
}

// -----------------------------------------------------------------------------

// walkWhileLoop walks a `while_loop` branch
func (w *Walker) walkWhileLoop(branch *syntax.ASTBranch, yieldsValue bool) (sem.HIRExpr, bool) {
	loop := &sem.HIRWhileLoop{
		ExprBase: sem.NewExprBase(nothingType(), sem.RValue, false),
	}

	// create a new loop expr context
	w.pushLoopContext()
	defer w.popExprContext()

	// walk through the loop content
	for _, item := range branch.Content {
		if itembranch, ok := item.(*syntax.ASTBranch); ok {
			switch itembranch.Name {
			case "variable_decl":
				// header variable declaration
				if varDecl, ok := w.walkVarDecl(branch, false, sem.ModNone); ok {
					loop.HeaderDecl = varDecl
				} else {
					return nil, false
				}
			case "expr":
				// header loop condition
				if loopCond, ok := w.walkExpr(branch, true); ok {
					w.solver.AddEqConstraint(loopCond.Type(), boolType(), itembranch.Position())
					loop.HeaderCond = loopCond
				} else {
					return nil, false
				}
			case "expr_stmt":
				// header update statement
				if updateStmt, ok := w.walkExprStmt(branch); ok {
					loop.HeaderUpdate = updateStmt
				} else {
					return nil, false
				}
			case "loop_body":
				if loopBody, noBreakClause, ok := w.walkLoopBody(branch, yieldsValue); ok {
					loop.LoopBody = loopBody
					loop.NoBreakClause = noBreakClause
				} else {
					return nil, false
				}
			}
		}
	}

	// update loop return type if the loop yields a value
	if yieldsValue {
		// TODO: Seq[loopBody.Type()]
	}

	return loop, true
}

// -----------------------------------------------------------------------------

// walkLoopBody walks a `loop_body` node.  It returns both the loop body itself
// and the `nobreak` body if one exists
func (w *Walker) walkLoopBody(branch *syntax.ASTBranch, yieldsValue bool) (sem.HIRExpr, sem.HIRExpr, bool) {
	var loopBody, noBreakClause sem.HIRExpr

	for _, item := range branch.Content {
		itembranch := item.(*syntax.ASTBranch)

		if itembranch.Name == "block_body" {
			if blockBody, ok := w.walkBlockBody(itembranch, yieldsValue); ok {
				loopBody = blockBody
			} else {
				return nil, nil, false
			}
		} else /* `nobreak_clause` */ {
			noBreakBody := itembranch.BranchAt(1)
			if noBreakBody.Name == "block_body" {
				if blockBody, ok := w.walkBlockBody(noBreakBody, yieldsValue); ok {
					noBreakClause = blockBody
				} else {
					return nil, nil, false
				}
			} else /* `block_content` */ {
				if blockBody, ok := w.walkBlockContent(noBreakBody, yieldsValue); ok {
					noBreakClause = blockBody
				} else {
					return nil, nil, false
				}
			}
		}
	}

	return loopBody, noBreakClause, true
}

// walkBlockBody walks a `block_body` node
func (w *Walker) walkBlockBody(branch *syntax.ASTBranch, yieldsValue bool) (sem.HIRExpr, bool) {
	// len / 2 => always maps to correct branch
	bodyBranch := branch.BranchAt(branch.Len() / 2)

	if bodyBranch.Name == "do_block" {
		return w.walkDoBlock(bodyBranch, yieldsValue)
	} else /* `expr` */ {
		return w.walkExpr(bodyBranch, yieldsValue)
	}
}

// -----------------------------------------------------------------------------

// walkDoBlock walks a `do_block` node and returns a HIRExpr
func (w *Walker) walkDoBlock(branch *syntax.ASTBranch, yieldsValue bool) (sem.HIRExpr, bool) {
	return w.walkBlockContent(branch.BranchAt(1), yieldsValue)
}

// walkBlockContent walks a `block_content` node and returns a HIRExpr
func (w *Walker) walkBlockContent(branch *syntax.ASTBranch, yieldsValue bool) (sem.HIRExpr, bool) {
	block := &sem.HIRDoBlock{
		ExprBase: sem.NewExprBase(
			w.solver.CreateTypeVar(nothingType(), func() {}),
			sem.RValue,
			false,
		),
	}

	updateBlockType := func(i int, sdt typing.DataType, pos *logging.TextPosition) {
		// the value is only yielded if it is the last element inside the
		// `block_content`, and the enclosing block is supposed to yield a
		// value; we don't care about the yielded value of the block if it isn't
		// actually used or expected
		if yieldsValue && i == branch.Len()-1 {
			w.solver.AddSubConstraint(block.Type(), sdt, pos)
		}
	}

	// branchControlIndex is the index where the first control flow statement
	// was encountered -- used to flag deadcode
	branchControlIndex := -1
	for i, item := range branch.Content {
		switch v := item.(type) {
		// only `block_element` can be a branch inside `block_content`
		case *syntax.ASTBranch:
			// access the inner node
			blockElem := v.BranchAt(0)

			switch blockElem.Name {
			case "stmt":
				if stmt, ok := w.walkStmt(blockElem); ok {
					block.Statements = append(block.Statements, stmt)

					switch stmt.Control() {
					case sem.CFNone:
						updateBlockType(i, stmt.Type(), blockElem.Position())
					default:
						// make sure not to override other control flow stmts
						if block.Control() == sem.CFNone {
							// other control flow that exits block
							block.SetControl(stmt.Control())
							branchControlIndex = i
						}
					}
				} else {
					return nil, false
				}
			case "expr_stmt":
				if stmt, ok := w.walkExprStmt(blockElem); ok {
					updateBlockType(i, stmt.Type(), blockElem.Position())
					block.Statements = append(block.Statements, stmt)
				} else {
					return nil, false
				}
			case "block_expr":
				// the same logic used in `updateBlockType` for yielding values
				if expr, ok := w.walkBlockExpr(blockElem, yieldsValue && i == branch.Len()-1); ok {
					if expr.Control() == sem.CFNone {
						updateBlockType(i, expr.Type(), blockElem.Position())
					} else {
						branchControlIndex = i
						block.SetControl(expr.Control())
					}

					block.Statements = append(block.Statements, expr)
				} else {
					return nil, false
				}
			}
		case *syntax.ASTLeaf:
			if v.Kind == syntax.ELLIPSIS {
				// panic on unimplemented block
				block.SetControl(sem.CFPanic)
			}
		}
	}

	// unreachable statements; -2 so we don't count indentation
	if branchControlIndex != -1 && branchControlIndex < len(branch.Content)-2 {
		w.logWarning(
			"unreachable code",
			logging.LMKDeadcode,
			syntax.TextPositionOfSpan(branch.Content[branchControlIndex+1], branch),
		)
	}

	// if the block type was never set (somehow), then we simply yield nothing
	if block.Type() == nil {
		block.SetType(nothingType())
	}

	return block, true
}
