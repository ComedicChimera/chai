package lower

import (
	"chaic/ast"
	"chaic/common"
	"chaic/mir"
	"chaic/report"
	"chaic/syntax"
	"chaic/types"
	"chaic/util"
)

func (l *Lowerer) lowerExpr(expr ast.ASTExpr) mir.Expr {
	switch v := expr.(type) {
	case *ast.TypeCast:
		return &mir.TypeCast{
			ExprBase: mir.NewExprBase(v.Span()),
			SrcExpr:  l.lowerExpr(v.SrcExpr),
			DestType: types.Simplify(v.Type()),
		}
	case *ast.BinaryOpApp:
		return l.buildBinaryOpApp(
			v.Span(),
			v.Op,
			l.lowerExpr(v.LHS),
			l.lowerExpr(v.RHS),
			types.Simplify(v.Type()),
		)
	case *ast.UnaryOpApp:
		if v.Op.OpMethod.ID >= common.OP_ID_END_RESERVED {
			// TODO
			// return &mir.FuncCall{
			// 	ExprBase: mir.NewExprBase(v.Span()),
			// 	Func: &mir.Identifier{
			// 		ExprBase: mir.NewExprBase(v.Op.Span),
			// 		Symbol:   v.Op.Overload.MIRSymbol,
			// 	},
			// 	Args: []mir.Expr{l.lowerExpr(v.Operand)},
			// }
		} else {
			return &mir.UnaryOpApp{
				ExprBase:   mir.NewExprBase(v.Span()),
				Op:         v.Op.OpMethod.ID,
				Operand:    l.lowerExpr(v.Operand),
				ResultType: types.Simplify(v.Type()),
			}
		}
	case *ast.PropertyAccess:
		return l.lowerPropAccess(v)
	case *ast.FuncCall:
		return &mir.FuncCall{
			ExprBase: mir.NewExprBase(v.Span()),
			Func:     l.lowerExpr(v.Func),
			Args:     util.Map(v.Args, l.lowerExpr),
		}
	case *ast.Indirect:
		if v.Elem.Category() == ast.LVALUE {
			return &mir.AddressOf{
				ExprBase: mir.NewExprBase(v.Span()),
				Element:  l.lowerExpr(v.Elem),
			}
		}
	case *ast.Deref:
		return &mir.Deref{
			ExprBase: mir.NewExprBase(v.Span()),
			Ptr:      l.lowerExpr(v.Ptr),
		}
	case *ast.StructLiteral:
		return l.lowerStructLiteral(v)
	case *ast.Identifier:
		return &mir.Identifier{
			ExprBase: mir.NewExprBase(v.Span()),
			Symbol:   v.Sym.MIRSymbol,
		}
	case *ast.Literal:
		return l.lowerLiteral(v)
	case *ast.Null:
		return l.lowerNull(v)
	}

	report.ReportICE("lowering not implemented for expr")
	return nil
}

/* -------------------------------------------------------------------------- */

// buildBinaryOpApp lowers a binary operator application.
func (l *Lowerer) buildBinaryOpApp(span *report.TextSpan, op *ast.AppliedOperator, lhs, rhs mir.Expr, resultType types.Type) mir.Expr {
	if op.OpMethod.ID >= common.OP_ID_END_RESERVED {
		// TODO
		report.ReportICE("lowering not implemented for non-intrinsic binary operators")
		return nil
		// return &mir.FuncCall{
		// 	ExprBase: mir.NewExprBase(span),
		// 	Func: &mir.Identifier{
		// 		ExprBase: mir.NewExprBase(op.Span),
		// 		// Symbol:   op.Overload.MIRSymbol,
		// 	},
		// 	Args: []mir.Expr{lhs, rhs},
		// }
	} else {
		return &mir.BinaryOpApp{
			ExprBase:   mir.NewExprBase(span),
			Op:         op.OpMethod.ID,
			LHS:        lhs,
			RHS:        rhs,
			ResultType: resultType,
		}
	}
}

// lowerPropAccess lowers an AST property access.
func (l *Lowerer) lowerPropAccess(prop *ast.PropertyAccess) mir.Expr {
	mroot := l.lowerExpr(prop.Root)

	// Automatic dereferencing.
	if _, ok := mroot.Type().(*types.PointerType); ok {
		mroot = &mir.Deref{
			ExprBase: mir.NewExprBase(mroot.Span()),
			Ptr:      mroot,
		}
	}

	switch prop.PropKind {
	case types.PropStructField:
		{
			st := mroot.Type().(*types.StructType)
			fieldNdx := st.Indices[prop.PropName]

			return &mir.FieldAccess{
				ExprBase:    mir.NewExprBase(prop.Span()),
				Struct:      mroot,
				FieldNumber: fieldNdx,
				FieldType:   st.Fields[fieldNdx].Type,
			}
		}
	default:
		report.ReportICE("lowering not implemented for property access")
		return nil
	}
}

// lowerStructLiteral lowers a struct literal.
func (l *Lowerer) lowerStructLiteral(slit *ast.StructLiteral) mir.Expr {
	st := types.Simplify(slit.Type()).(*types.StructType)

	if slit.SpreadInit == nil {
		// TODO: handle imported structs
		sym := l.pkg.SymbolTable[st.Name()]
		sdef := l.pkg.Files[sym.FileNumber].Definitions[st.DeclIndex()].(*ast.StructDef)

		fieldInits := make([]mir.Expr, len(st.Fields))
		for i, field := range st.Fields {
			if astFieldInit, ok := slit.FieldInits[field.Name]; ok {
				fieldInits[i] = l.lowerExpr(astFieldInit.InitExpr)
			} else if astDefaultInit, ok := sdef.FieldInits[field.Name]; ok {
				fieldInits[i] = l.lowerExpr(astDefaultInit)
			} else {
				fieldInits[i] = l.lowerNull(&ast.Null{
					ExprBase: ast.NewTypedExprBase(slit.Span(), field.Type),
				})
			}
		}

		// TODO: figure out how to get the Msymbol and struct def
		sid := &mir.StructInstanceDecl{
			StmtBase: mir.NewStmtBase(slit.Span()),
			Ident: &mir.Identifier{
				ExprBase: mir.NewExprBase(slit.Span()),
				Symbol:   sym.MIRSymbol,
			},
			FieldInits: fieldInits,
		}

		l.appendStmt(sid)

		return sid.Ident
	} else {
		vdecl := &mir.VarDecl{
			StmtBase: mir.NewStmtBase(slit.Span()),
			Ident: &mir.Identifier{
				ExprBase: mir.NewExprBase(slit.Span()),
				Symbol: &mir.MSymbol{
					Name:              "",
					Type:              slit.Type(),
					IsImplicitPointer: types.IsPtrWrappedType(slit.Type()),
				},
			},
		}

		l.appendStmt(vdecl)

		for _, fieldInit := range slit.FieldInits {
			fieldAccess := &mir.FieldAccess{
				ExprBase:    mir.NewExprBase(fieldInit.NameSpan),
				Struct:      vdecl.Ident,
				FieldNumber: st.Indices[fieldInit.Name],
				FieldType:   st.Fields[st.Indices[fieldInit.Name]].Type,
			}

			assign := &mir.Assignment{
				StmtBase: mir.NewStmtBase(fieldInit.NameSpan),
				LHS:      fieldAccess,
				RHS:      l.lowerExpr(fieldInit.InitExpr),
			}

			l.appendStmt(assign)
		}

		return vdecl.Ident
	}
}

// lowerLiteral lowers an AST literal to a constant.
func (l *Lowerer) lowerLiteral(lit *ast.Literal) mir.Expr {
	vb := mir.NewValueBase(lit.Span(), types.Simplify(lit.Type()))

	switch lit.Kind {
	case syntax.TOK_RUNELIT:
		return &mir.ConstInt{ValueBase: vb, IntValue: int64(lit.Value.(int32))}
	case syntax.TOK_INTLIT:
		return &mir.ConstInt{ValueBase: vb, IntValue: lit.Value.(int64)}
	case syntax.TOK_FLOATLIT:
		return &mir.ConstReal{ValueBase: vb, FloatValue: lit.Value.(float64)}
	case syntax.TOK_NUMLIT:
		if types.InnerType(lit.Type()).(types.PrimitiveType).IsFloating() {
			var fv float64
			if _fv, ok := lit.Value.(float64); ok {
				fv = _fv
			} else {
				fv = float64(uint64(lit.Value.(int64)))
			}

			return &mir.ConstReal{ValueBase: vb, FloatValue: fv}
		} else {
			return &mir.ConstInt{ValueBase: vb, IntValue: lit.Value.(int64)}
		}
	case syntax.TOK_BOOLLIT:
		{
			boolIntValue := 0
			if lit.Value.(bool) {
				boolIntValue = 1
			}

			return &mir.ConstInt{ValueBase: vb, IntValue: int64(boolIntValue)}
		}
	default:
		report.ReportICE("lowering not implemented for literal")
		return nil
	}
}

// lowerNull lowers an AST null to a MIR expression.
func (l *Lowerer) lowerNull(null *ast.Null) mir.Expr {
	mtype := types.Simplify(null.Type())

	vb := mir.NewValueBase(null.Span(), mtype)

	switch v := mtype.(type) {
	case types.PrimitiveType:
		if v.IsIntegral() || v == types.PrimTypeBool {
			return &mir.ConstInt{
				ValueBase: vb,
				IntValue:  0,
			}
		} else if v.IsFloating() {
			return &mir.ConstReal{
				ValueBase:  vb,
				FloatValue: 0,
			}
		} else {
			return &mir.ConstUnit{
				ValueBase: mir.NewValueBase(null.Span(), v),
			}
		}
	case *types.PointerType:
		return &mir.ConstNullPtr{ValueBase: vb}
	case *types.StructType:
		return l.lowerStructLiteral(&ast.StructLiteral{ExprBase: ast.NewExprBase(null.Span())})
	}

	report.ReportICE("null generation for type not implemented")
	return nil
}
