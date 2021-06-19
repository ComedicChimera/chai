package walk

import (
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
		ExprBase: sem.NewExprBase(nil, sem.RValue, false),
	}

	for i, item := range branch.Content {
		// only `block_element` can be a branch inside `block_content`
		if itembranch, ok := item.(*syntax.ASTBranch); ok {
			// access the inner node
			blockElem := itembranch.BranchAt(0)

			switch blockElem.Name {
			case "stmt":
				if dt, ok := w.walkStmt(blockElem); ok {
					_ = dt
				} else {
					return nil, false
				}
			case "expr_stmt":
				if dt, ok := w.walkExprStmt(blockElem); ok {
					_ = dt
				} else {
					return nil, false
				}
			case "block_expr":
				// the value is only yielded if it is the last element inside
				// the `block_content` and the enclosing block itself is
				// supposed to yield a value
				if dt, ok := w.walkBlockExpr(blockElem, yieldsValue && i == len(branch.Content)-1); ok {
					_ = dt
				} else {
					return nil, false
				}
			}
		}
	}

	if block.Type() == nil {
		block.SetType(typing.PrimType(typing.PrimKindNothing))
	}

	return block, true
}

// walkBlockExpr walks a block expression (`block_expr`)
func (w *Walker) walkBlockExpr(branch *syntax.ASTBranch, yieldsValue bool) (sem.HIRExpr, bool) {
	return nil, false
}
