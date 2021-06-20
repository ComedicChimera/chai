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
			w.solver.AddConstraint(block.Type(), sdt, typing.TCCoerce, pos)
		}
	}

	for i, item := range branch.Content {
		// only `block_element` can be a branch inside `block_content`
		if itembranch, ok := item.(*syntax.ASTBranch); ok {
			// access the inner node
			blockElem := itembranch.BranchAt(0)

			switch blockElem.Name {
			case "stmt":
				if stmt, ok := w.walkStmt(blockElem); ok {
					updateBlockType(i, stmt.Type(), blockElem.Position())
					block.Statements = append(block.Statements, stmt)

					// TODO: handle yield statements
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
					updateBlockType(i, expr.Type(), blockElem.Position())
					block.Statements = append(block.Statements, expr)
				} else {
					return nil, false
				}
			}
		}
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
