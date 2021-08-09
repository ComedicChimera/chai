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
		return w.walkIfChain(blockExpr, yieldsValue)
	case "while_loop":
		return w.walkWhileLoop(blockExpr, yieldsValue)
	}

	// unreachable
	return nil, false
}

// walkIfChain walks an `if_chain` branch
func (w *Walker) walkIfChain(branch *syntax.ASTBranch, yieldsValue bool) (sem.HIRExpr, bool) {
	// create the if control frame
	w.pushControlFrame(FKIf)
	defer w.popControlFrame()

	ifChain := &sem.HIRIfChain{
		ExprBase: sem.NewExprBase(nothingType(), sem.RValue, false),
	}

	// store the positions of all the branches so we can refer to them later
	var ifBranchPos, elseBranchPos *logging.TextPosition
	var elifBranchPoses []*logging.TextPosition

	for _, item := range branch.Content {
		// only items are subbranches
		itembranch := item.(*syntax.ASTBranch)

		// push an expr context scope for each of the subbranches
		w.pushScopeContext()

		// handle else blocks
		if itembranch.Name == "else_block" {
			bodyBranch := itembranch.LastBranch()
			elseBranchPos = itembranch.Last().Position()

			if bodyBranch.Name == "block_content" {
				if body, ok := w.walkBlockContent(bodyBranch, yieldsValue); ok {
					ifChain.ElseBranch = body
				} else {
					return nil, false
				}
			} else {
				if body, ok := w.walkBlockBody(bodyBranch, yieldsValue); ok {
					ifChain.ElseBranch = body
				} else {
					return nil, false
				}
			}
		} else /* handle if and elif blocks */ {
			condBranch := &sem.HIRCondBranch{}

			// if and elif blocks are syntactically equivalent for our purposes
			// so we can just iterate through them the same way and collect all
			// relevant block information
			for _, elem := range itembranch.Content {
				if elembranch, ok := elem.(*syntax.ASTBranch); ok {
					switch elembranch.Name {
					case "variable_decl":
						if varDecl, ok := w.walkVarDecl(elembranch, false, sem.ModNone); ok {
							condBranch.HeaderDecl = varDecl
						} else {
							return nil, false
						}
					case "expr":
						if expr, ok := w.walkExpr(elembranch, true); ok {
							// force the conditional expression to be a boolean
							w.solver.AddEqConstraint(expr.Type(), boolType(), elembranch.Position())
							condBranch.HeaderCond = expr
						} else {
							return nil, false
						}
					case "block_body":
						if body, ok := w.walkBlockBody(elembranch, yieldsValue); ok {
							condBranch.BranchBody = body
						} else {
							return nil, false
						}
					}
				}
			}

			// add the cond branch and the branch position
			if itembranch.Name == "if_block" {
				ifBranchPos = itembranch.Last().Position()
				ifChain.IfBranch = condBranch
			} else /* elif_block */ {
				elifBranchPoses = append(elifBranchPoses, itembranch.Last().Position())
				ifChain.ElifBranches = append(ifChain.ElifBranches, condBranch)
			}
		}

		// exit the branch scope
		w.popExprContext()
	}

	// determine the resultant type of the if chain if it yields a value and
	// does not have any control flow effect
	if yieldsValue && w.hasNoControlEffect() {
		// create the type variable that will store the return value of the if chain
		tvar := w.solver.CreateTypeVar(nil, func() { /* additional handler should never be necessary */ })

		// constraint the type of the primary if block
		w.solver.AddSubConstraint(tvar, ifChain.IfBranch.BranchBody.Type(), ifBranchPos)

		// constraint the types of all the elif blocks
		for i, elifBranch := range ifChain.ElifBranches {
			w.solver.AddSubConstraint(tvar, elifBranch.BranchBody.Type(), elifBranchPoses[i])
		}

		if ifChain.ElseBranch == nil {
			// if chain is not exhaustive, wrap the type variable in an optional
			// type before setting it as the return type of the if branch --
			// TODO
		} else {
			// add the constraint on the else branch
			w.solver.AddSubConstraint(tvar, ifChain.ElseBranch.Type(), elseBranchPos)

			// set the return type of the else to the raw type of the type
			// variable
			ifChain.SetType(tvar)
		}
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

	// create a loop control frame
	w.pushControlFrame(FKLoop)
	defer w.popControlFrame()

	// walk through the loop content
	for _, item := range branch.Content {
		if itembranch, ok := item.(*syntax.ASTBranch); ok {
			switch itembranch.Name {
			case "variable_decl":
				// header variable declaration
				if varDecl, ok := w.walkVarDecl(itembranch, false, sem.ModNone); ok {
					loop.HeaderDecl = varDecl
				} else {
					return nil, false
				}
			case "expr":
				// header loop condition
				if loopCond, ok := w.walkExpr(itembranch, true); ok {
					w.solver.AddEqConstraint(loopCond.Type(), boolType(), itembranch.Position())
					loop.HeaderCond = loopCond
				} else {
					return nil, false
				}
			case "expr_stmt":
				// header update statement
				if updateStmt, ok := w.walkExprStmt(itembranch); ok {
					loop.HeaderUpdate = updateStmt
				} else {
					return nil, false
				}
			case "loop_body":
				if loopBody, noBreakClause, ok := w.walkLoopBody(itembranch, yieldsValue); ok {
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
	bodyBranch := branch.LastBranch()

	if bodyBranch.Name == "block_content" {
		return w.walkBlockContent(bodyBranch, yieldsValue)
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

	// branchControlIndex is the index where the first control flow statement
	// was encountered -- used to flag deadcode
	branchControlIndex := -1

	// updateBlockType handles the determination of the resultant type of the
	// block including checking for control flow changes
	updateBlockType := func(i int, sdt typing.DataType, pos *logging.TextPosition) {
		// if there is a control index, then the block never returns a value
		// directly and there is no need to update the block type
		if branchControlIndex != -1 {
			return
		} else if !w.hasNoControlEffect() {
			// control just updated, we need to set the branch control index
			branchControlIndex = i
		}

		// the value is only yielded if it is the last element inside the
		// `block_content`, and the enclosing block is supposed to yield a
		// value; we don't care about the yielded value of the block if it isn't
		// actually used or expected
		if yieldsValue && i == branch.Len()-1 {
			w.solver.AddSubConstraint(block.Type(), sdt, pos)
		}
	}

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
					updateBlockType(i, stmt.Type(), blockElem.Position())
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
					block.Statements = append(block.Statements, expr)
					updateBlockType(i, expr.Type(), blockElem.Position())
				} else {
					return nil, false
				}
			}
		case *syntax.ASTLeaf:
			if v.Kind == syntax.ELLIPSIS {
				// panic on unimplemented block
				w.updateControl(CFPanic)
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
