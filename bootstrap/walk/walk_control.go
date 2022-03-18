package walk

import (
	"chai/ast"
	"chai/typing"
)

// walkIfExpr walks an if expression.
func (w *Walker) walkIfExpr(ifExpr *ast.IfExpr, yieldsValue bool) bool {
	if yieldsValue {
		// if statements that yield a value must uphold type consistency.  We can check
		// this by asserting that all the types are equal to type of the first branch.
		var firstBranchType typing.DataType
		for _, condBranch := range ifExpr.CondBranches {
			// push a new scope for the branch
			w.pushScope()

			// walk the header variable as necessary
			if condBranch.HeaderVarDecl != nil && !w.walkLocalVarDecl(condBranch.HeaderVarDecl) {
				w.popScope()
				return false
			}

			// walk the condition
			if !w.walkExpr(condBranch.Cond, true) {
				w.popScope()
				return false
			}

			// constrain the condition to be a boolean
			w.solver.MustBeEquiv(condBranch.Cond.Type(), typing.BoolType(), condBranch.Cond.Position())

			// walk the body
			if !w.walkExpr(condBranch.Body, true) {
				w.popScope()
				return false
			}

			// constrain it to be the same as the firstBranchType (type consistency)
			if firstBranchType == nil {
				firstBranchType = condBranch.Body.Type()
			} else {
				w.solver.MustBeEquiv(firstBranchType, condBranch.Body.Type(), condBranch.Body.Position())
			}

			w.popScope()
		}

		// in order to be exhaustive, if statements must have an `else`
		if ifExpr.ElseBranch == nil {
			w.reportError(ifExpr.Position(), "if expression must have an else to be exhaustive")
		}

		// walk and constrain the `else` branch body
		if w.walkExpr(ifExpr.ElseBranch, true) {
			w.solver.MustBeEquiv(firstBranchType, ifExpr.ElseBranch.Type(), ifExpr.ElseBranch.Position())
		}

		// set the returned type to be the first branch type
		ifExpr.SetType(firstBranchType)
	} else {
		for _, condBranch := range ifExpr.CondBranches {
			// push a new scope for the body
			w.pushScope()

			// walk the header variable as necessary
			if condBranch.HeaderVarDecl != nil && !w.walkLocalVarDecl(condBranch.HeaderVarDecl) {
				w.popScope()
				return false
			}

			// walk the condition
			if !w.walkExpr(condBranch.Cond, true) {
				w.popScope()
				return false
			}

			// constrain the condition to be a boolean
			w.solver.MustBeEquiv(condBranch.Cond.Type(), typing.BoolType(), condBranch.Cond.Position())

			// walk the body (know it doesn't yield a value => no need to
			// perform any other checks)
			if !w.walkExpr(condBranch.Body, false) {
				w.popScope()
				return false
			}

			w.popScope()
		}

		// check else clauses
		if ifExpr.ElseBranch != nil {
			// push scope for else block
			w.pushScope()

			if !w.walkExpr(ifExpr.ElseBranch, false) {
				w.popScope()
				return false
			}
		}
	}

	return true
}

// walkWhileExpr walks a while loop.
func (w *Walker) walkWhileExpr(whileExpr *ast.WhileExpr, yieldsValue bool) bool {
	if yieldsValue {
		// NOTE: loop generators are not going to be supported in Alpha Chai
		w.reportError(whileExpr.Pos, "loop generators are not supported in Alpha Chai")
		return false
	}

	// push a new scope for the body
	w.pushScope()
	defer w.popScope()

	// walk the header variable as necessary
	if whileExpr.HeaderVarDecl != nil && !w.walkLocalVarDecl(whileExpr.HeaderVarDecl) {
		return false
	}

	// walk and constrain the condition
	if !w.walkExpr(whileExpr.Cond, true) {
		return false
	}

	w.solver.MustBeEquiv(whileExpr.Cond.Type(), typing.BoolType(), whileExpr.Cond.Position())

	// walk the body
	return w.walkExpr(whileExpr.Body, true)
}
