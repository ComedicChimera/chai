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

// WalkFuncBody walks a function body (`decl_func_body`)
func (w *Walker) WalkFuncBody(branch *syntax.ASTBranch, fn *typing.FuncType) (sem.HIRExpr, bool) {
	w.pushFuncContext(fn)
	defer w.popExprContext()

	// maps branch of len 1 and 2 to 0, maps branch of len 3 to 1 -- always
	// giving us the correct branch for the expression
	exprBranch := branch.BranchAt(branch.Len() / 2)

	switch exprBranch.Name {
	case "block_expr":
		// TODO
	}

	return nil, false
}

// WalkExpr walks an expression node and returns a HIRExpr
func (w *Walker) WalkExpr(branch *syntax.ASTBranch, yieldsValue bool) (sem.HIRExpr, bool) {
	exprBranch := branch.BranchAt(0)

	switch exprBranch.Name {
	case "do_block":
		return w.walkDoBlock(exprBranch, yieldsValue)
	case "simple_expr":
		return w.walkSimpleExpr(exprBranch)
	case "block_expr":
		return w.walkBlockExpr(exprBranch, yieldsValue)
	}

	// unreachable
	return nil, false
}

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
