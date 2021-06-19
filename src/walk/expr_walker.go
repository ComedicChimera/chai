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
	// handle function context management
	w.pushFuncContext(fn)
	defer w.popExprContext()

	// maps branch of len 1 and 2 to 0, maps branch of len 3 to 1 -- always
	// giving us the correct branch for the expression
	exprBranch := branch.BranchAt(branch.Len() / 2)

	// a function that returns nothing effectively yields no meaningful value
	yieldsValue := true
	if pt, ok := fn.ReturnType.(typing.PrimType); ok && pt == typing.PrimKindNothing {
		yieldsValue = false
	}

	// walk the function body
	var hirExpr sem.HIRExpr
	switch exprBranch.Name {
	case "block_expr":
		if e, ok := w.walkBlockExpr(exprBranch, yieldsValue); ok {
			hirExpr = e
		} else {
			return nil, false
		}
	case "simple_expr":
		if e, ok := w.walkSimpleExpr(exprBranch, yieldsValue); ok {
			hirExpr = e
		} else {
			return nil, false
		}
	case "do_block":
		if e, ok := w.walkDoBlock(exprBranch, yieldsValue); ok {
			hirExpr = e
		} else {
			return nil, false
		}
	case "expr":
		if e, ok := w.walkExpr(exprBranch, yieldsValue); ok {
			hirExpr = e
		} else {
			return nil, false
		}
	}

	// run the solver at the end of the function body
	if !w.solver.Solve() {
		return nil, false
	}

	// if we reach here, then the body was walked successfully
	return hirExpr, true
}

// -----------------------------------------------------------------------------

// walkExpr walks an expression node and returns a HIRExpr
func (w *Walker) walkExpr(branch *syntax.ASTBranch, yieldsValue bool) (sem.HIRExpr, bool) {
	exprBranch := branch.BranchAt(0)

	switch exprBranch.Name {
	case "do_block":
		return w.walkDoBlock(exprBranch, yieldsValue)
	case "simple_expr":
		return w.walkSimpleExpr(exprBranch, yieldsValue)
	case "block_expr":
		return w.walkBlockExpr(exprBranch, yieldsValue)
	}

	// unreachable
	return nil, false
}

// walkSimpleExpr walks a simple expression (`simple_expr`)
func (w *Walker) walkSimpleExpr(branch *syntax.ASTBranch, yieldsValue bool) (sem.HIRExpr, bool) {
	return nil, false
}
