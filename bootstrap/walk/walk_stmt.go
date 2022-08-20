package walk

import (
	"chaic/ast"
	"chaic/common"
	"chaic/report"
	"chaic/syntax"
	"chaic/types"
)

// walkVarDecl walks a local variable declaration.
func (w *Walker) walkVarDecl(vd *ast.VarDecl) {
	for _, varList := range vd.VarLists {
		if varList.Initializer != nil { // Has initializer.
			// Walk the initializer.
			w.walkExpr(varList.Initializer)

			if len(varList.Vars) == 1 {
				// Single variable: no pattern matching.
				ident := varList.Vars[0]

				if ident.Sym.Type == nil { // Identifier type is to be inferred.
					ident.Sym.Type = varList.Initializer.Type()
				} else { // Specified type => unify.
					w.solver.MustEqual(ident.Sym.Type, varList.Initializer.Type(), varList.Initializer.Span())
				}
			} else {
				// TODO: pattern matching
				report.ReportFatal("pattern matching not implemented yet")
			}
		}

		// Declare variable symbols.
		for _, ident := range varList.Vars {
			w.defineLocal(ident.Sym)
		}
	}
}

// walkAssign walks an assignment statement.
func (w *Walker) walkAssign(as *ast.Assignment) {
	// Walk the LHS vars and the RHS expressions.
	for _, lhsVar := range as.LHSVars {
		w.walkLHSExpr(lhsVar)
	}

	for _, rhsExpr := range as.RHSExprs {
		w.walkExpr(rhsExpr)
	}

	if len(as.LHSVars) == len(as.RHSExprs) { // No pattern matching
		if as.CompoundOp == nil { // No compound operator.
			for i, lhsVar := range as.LHSVars {
				w.solver.MustEqual(lhsVar.Type(), as.RHSExprs[i].Type(), as.RHSExprs[i].Span())
			}
		} else {
			for i, lhsVar := range as.LHSVars {
				operSpan := report.NewSpanOver(lhsVar.Span(), as.RHSExprs[i].Span())

				rhsTyp := w.checkOperApp(
					as.CompoundOp,
					operSpan,
					lhsVar,
					as.RHSExprs[i],
				)

				w.solver.MustEqual(lhsVar.Type(), rhsTyp, operSpan)
			}
		}
	} else { // Pattern matching
		// TODO: pattern matching
		report.ReportFatal("pattern matching not implemented yet")
	}
}

// walkIncDec walks an increment/decrement statement
func (w *Walker) walkIncDec(incdec *ast.IncDecStmt) {
	w.walkLHSExpr(incdec.LHSOperand)

	rhsType := w.checkOperApp(
		incdec.Op,
		incdec.Span(),
		incdec.LHSOperand,
		&ast.Literal{ExprBase: ast.NewTypedExprBase(incdec.Span(), incdec.LHSOperand.Type())},
	)

	w.solver.MustEqual(incdec.LHSOperand.Type(), rhsType, incdec.Span())
}

// walkLHSExpr walks an LHS expression and marks it as mutable.
func (w *Walker) walkLHSExpr(expr ast.ASTExpr) {
	switch v := expr.(type) {
	case *ast.Identifier:
		{
			sym := w.lookup(v.Name, v.Span())

			if sym.DefKind == common.DefKindType {
				w.error(v.Span(), "%s cannot be used as a value", sym.Name)
			}

			if sym.Constant {
				w.recError(v.Span(), "cannot mutate an immutable value")
			}

			v.Sym = sym
		}
	case *ast.Deref:
		{
			w.walkExpr(v.Ptr)

			elemType := w.solver.NewTypeVar("element type", v.Span())
			w.solver.MustEqual(&types.PointerType{ElemType: elemType, Const: false}, v.Ptr.Type(), v.Ptr.Span())

			v.NodeType = elemType
		}
	default:
		report.ReportICE("invalid LHS expression")
	}
}

// walkKeywordStmt walks a keyword statement (like `break`).  The resulting
// control mode of the statement is returned.
func (w *Walker) walkKeywordStmt(ks *ast.KeywordStmt) int {
	switch ks.Kind {
	case syntax.TOK_BREAK:
		if w.loopDepth == 0 {
			w.error(ks.Span(), "cannot use break outside a loop")
		}

		return ControlLoop
	case syntax.TOK_CONTINUE:
		if w.loopDepth == 0 {
			w.error(ks.Span(), "cannot use continue outside a loop")
		}

		return ControlLoop
	}

	// unreachable
	return ControlNone
}

// walkReturnStmt walks a return statement.
func (w *Walker) walkReturnStmt(rs *ast.ReturnStmt) {
	for _, expr := range rs.Exprs {
		w.walkExpr(expr)
	}

	switch len(rs.Exprs) {
	case 0:
		if !types.IsUnit(w.enclosingReturnType) {
			w.error(rs.Span(), "must return a value")
		}
	case 1:
		w.solver.MustEqual(w.enclosingReturnType, rs.Exprs[0].Type(), rs.Exprs[0].Span())
	default:
		report.ReportFatal("multiple return values not implemented yet")
	}
}
