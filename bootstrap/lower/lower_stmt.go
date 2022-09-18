package lower

import (
	"chaic/ast"
	"chaic/mir"
	"chaic/report"
	"chaic/types"
)

// lowerVarDecl lowers a variable declaration.
func (l *Lowerer) lowerVarDecl(vd *ast.VarDecl) {
	for _, vlist := range vd.VarLists {
		init := l.lowerExpr(vlist.Initializer)

		// TODO: pattern matching
		if len(vlist.Vars) > 1 {
			report.ReportICE("pattern matching variable declaration not implemented")
		}

		// Struct literal variable initializers.
		if _, ok := vlist.Initializer.(*ast.StructLiteral); ok {
			for _, ident := range vlist.Vars {
				ident.Sym.MIRSymbol = init.(*mir.Identifier).Symbol
			}
		} else {
			for _, ident := range vlist.Vars {
				vd := &mir.VarDecl{
					StmtBase: mir.NewStmtBase(vd.Span()),
					Ident: &mir.Identifier{
						ExprBase: mir.NewExprBase(ident.Span()),
						Symbol: &mir.MSymbol{
							Name:              ident.Name,
							Type:              types.Simplify(ident.Type()),
							IsImplicitPointer: true,
						},
					},
					Initializer: init,
				}

				l.appendStmt(vd)
			}
		}
	}
}

// lowerAssignment lowers an assignment statement.
func (l *Lowerer) lowerAssignment(assign *ast.Assignment) {
	if len(assign.LHSVars) == len(assign.RHSExprs) {
		mRHSVals := make([]mir.Expr, len(assign.RHSExprs))
		for i, rhsExpr := range assign.RHSExprs {
			tempVar := &mir.VarDecl{
				StmtBase: mir.NewStmtBase(rhsExpr.Span()),
				Ident: &mir.Identifier{
					ExprBase: mir.NewExprBase(rhsExpr.Span()),
					Symbol: &mir.MSymbol{
						Type:              types.Simplify(rhsExpr.Type()),
						IsImplicitPointer: false,
					},
				},
				Initializer: l.lowerExpr(rhsExpr),
				Temporary:   true,
			}
			l.appendStmt(tempVar)

			mRHSVals[i] = tempVar.Ident
		}

		// if assign.CompoundOp == nil {
		// 	for i, lhsVar := range assign.LHSVars {
		// 		mLHSVar := l.lowerExpr(lhsVar)

		// 		l.appendStmt(&mir.Assignment{
		// 			StmtBase: mir.NewStmtBase(assign.Span()),
		// 			LHS:      mLHSVar,
		// 			RHS:      mRHSVals[i],
		// 		})
		// 	}
		// } else {
		// 	for i, lhsVar := range assign.LHSVars {
		// 		mLHSVar := l.lowerExpr(lhsVar)
		// 		lhsTempVar := &mir.VarDecl{
		// 			StmtBase: mir.NewStmtBase(lhsVar.Span()),
		// 			Ident: &mir.Identifier{
		// 				ExprBase: mir.NewExprBase(lhsVar.Span()),
		// 				Symbol: &mir.MSymbol{
		// 					Type:              mLHSVar.Type(),
		// 					IsImplicitPointer: false,
		// 				},
		// 			},
		// 		}
		// 		l.appendStmt(lhsTempVar)

		// 		opResult := l.buildBinaryOpApp(
		// 			assign.Span(),
		// 			assign.CompoundOp,
		// 			lhsTempVar.Ident,
		// 			mRHSVals[i],
		// 			mLHSVar.Type(),
		// 		)

		// 		l.appendStmt(&mir.Assignment{
		// 			StmtBase: mir.NewStmtBase(assign.Span()),
		// 			LHS:      mLHSVar,
		// 			RHS:      opResult,
		// 		})
		// 	}
		// }
	} else {
		// TODO: pattern matching
		report.ReportICE("pattern matching assignment not implemented")
	}
}

// lowerIncDecStmt lowers an increment/decrement statement.
func (l *Lowerer) lowerIncDecStmt(incdec *ast.IncDecStmt) mir.Statement {
	// lhsOp := l.lowerExpr(incdec.LHSOperand)

	// opResult := l.buildBinaryOpApp(
	// 	incdec.Span(),
	// 	incdec.Op,
	// 	lhsOp,
	// 	&mir.ConstInt{
	// 		ValueBase: mir.NewValueBase(incdec.Span(), lhsOp.Type()),
	// 		IntValue:  0,
	// 	},
	// 	lhsOp.Type(),
	// )

	// return &mir.Assignment{
	// 	StmtBase: mir.NewStmtBase(incdec.Span()),
	// 	LHS:      lhsOp,
	// 	RHS:      opResult,
	// }
	return nil
}
