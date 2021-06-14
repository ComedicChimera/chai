package walk

import (
	"chai/sem"
	"chai/syntax"
	"chai/typing"
)

// Many expression functions take a boolean parameter called `yieldsValue` which
// essentially indicates whether or not the block yields a meaningful value.
// This is used to facilitate the behavior of if-blocks, match-blocks, etc. that
// can yield multiple different, ununifiable types on different branches because
// their value is not used.

// WalkExpr walks an expression node and returns a HIRExpr
func (w *Walker) WalkExpr(branch *syntax.ASTBranch, yieldsValue bool) (sem.HIRExpr, bool) {
	return nil, false
}

// WalkDoBlock walks a `do_block` node and returns a HIRExpr
func (w *Walker) WalkDoBlock(branch *syntax.ASTBranch, yieldsValue bool) (sem.HIRExpr, bool) {
	return w.walkBlockContents(branch.BranchAt(1), yieldsValue)
}

// walkBlockContents walks a `block_content` node and returns a HIRExpr
func (w *Walker) walkBlockContents(branch *syntax.ASTBranch, yieldsValue bool) (sem.HIRExpr, bool) {
	block := &sem.HIRDoBlock{
		ExprBase: sem.NewExprBase(nil, sem.RValue, false),
	}

	for i, item := range branch.Content {
		_ = i

		// only `block_element` can be a branch inside `block_content`
		if itembranch, ok := item.(*syntax.ASTBranch); ok {
			// access the inner node
			blockElem := itembranch.BranchAt(0)

			switch blockElem.Name {
			case "stmt":
				stmt := blockElem.BranchAt(0)
				switch stmt.Name {

				}
			case "expr_stmt":
			case "block_expr":
			}
		}
	}

	if block.Type() == nil {
		block.SetType(typing.PrimType(typing.PrimKindNothing))
	}

	return block, true
}
