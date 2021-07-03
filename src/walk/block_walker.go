package walk

import (
	"chai/logging"
	"chai/sem"
	"chai/syntax"
	"chai/typing"
)

// walkDoBlock walks a `do_block` node and returns a HIRExpr
func (w *Walker) walkDoBlock(branch *syntax.ASTBranch, yieldsValue bool) (sem.HIRExpr, bool) {
	return w.walkBlockContents(branch.BranchAt(1), yieldsValue)
}

// walkBlockContents walks a `block_content` node and returns a HIRExpr
func (w *Walker) walkBlockContents(branch *syntax.ASTBranch, yieldsValue bool) (sem.HIRExpr, bool) {
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
					case sem.CFYield:
						// TODO: handle yield statements
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

// walkBlockExpr walks a block expression (`block_expr`)
func (w *Walker) walkBlockExpr(branch *syntax.ASTBranch, yieldsValue bool) (sem.HIRExpr, bool) {
	return nil, false
}
