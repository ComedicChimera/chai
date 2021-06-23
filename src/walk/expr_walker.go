package walk

import (
	"chai/sem"
	"chai/syntax"
	"chai/typing"
)

// WalkPredicates walks the predicates (func bodies, initializers) of a file
func (w *Walker) WalkPredicates(root *sem.HIRRoot) {
	for _, def := range root.Defs {
		switch v := def.(type) {
		case *sem.HIRFuncDef:
			ft := v.DefBase.Sym().Type.(*typing.FuncType)

			// validate the argument initializers
			for name, argInit := range v.ArgumentInitializers {
				var expectedType typing.DataType
				for _, arg := range ft.Args {
					if arg.Name == name {
						expectedType = arg.Type
						break
					}
				}

				argBranch := convertIncompleteToBranch(argInit)
				if expr, ok := w.walkExpr(argBranch, true); ok {
					w.solver.AddSubConstraint(expectedType, expr.Type(), argBranch.Position())

					// solve the argument initializer immediately -- the only
					// context should be that of the initializer
					if w.solver.Solve() {
						v.ArgumentInitializers[name] = expr
					}
				}
			}

			// validate the function body
			if expr, ok := w.walkFuncBody(convertIncompleteToBranch(v.Body), ft); ok {
				v.Body = expr
			}
		case *sem.HIROperDef:
			// validate the operator body
			if expr, ok := w.walkFuncBody(
				convertIncompleteToBranch(v.Body),
				v.DefBase.Sym().Type.(*typing.FuncType)); ok {

				v.Body = expr
			}
		}
	}
}

// -----------------------------------------------------------------------------

// Many expression functions take a boolean parameter called `yieldsValue` which
// essentially indicates whether or not the block yields a meaningful value.
// This is used to facilitate the behavior of if-blocks, match-blocks, etc. that
// can yield multiple different, ununifiable types on different branches because
// their value is not used.

// walkFuncBody walks a function body (`decl_func_body`)
func (w *Walker) walkFuncBody(branch *syntax.ASTBranch, fn *typing.FuncType) (sem.HIRExpr, bool) {
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

	// constraint the return type of the block if the function yields a value
	if yieldsValue {
		w.solver.AddSubConstraint(fn.ReturnType, hirExpr.Type(), branch.Position())
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
