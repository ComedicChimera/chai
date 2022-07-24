package codegen

import (
	"chaic/ast"
	"chaic/common"
	"chaic/llvm"
	"chaic/report"
	"chaic/syntax"
	"chaic/types"
)

// generateExpr generates an expression.
func (g *Generator) generateExpr(expr ast.ASTExpr) llvm.Value {
	switch v := expr.(type) {
	case *ast.TypeCast:
		return g.generateTypeCast(v)
	case *ast.BinaryOpApp:
		return g.generateBinaryOpApp(v)
	case *ast.UnaryOpApp:
		return g.generateUnaryOpApp(v)
	case *ast.FuncCall:
		{
			fn := g.generateExpr(v.Func)

			args := make([]llvm.Value, len(v.Args))
			for i, arg := range v.Args {
				args[i] = g.generateExpr(arg)
			}

			return g.callFunc(v.Type(), fn, args...)
		}
	case *ast.Indirect:
		if v.Elem.Category() == ast.LVALUE {
			// For l-value, indirection simply corresponds to returning an LHS
			// expression (a pointer to the value).
			return g.generateLHSExpr(v.Elem)
		} else {
			// TODO: not implemented
			report.ReportICE("codegen for r-value indirect not implemented")
		}
	case *ast.Deref:
		return g.irb.BuildLoad(g.convType(v.Ptr.Type()), g.generateExpr(v.Ptr))
	case *ast.Literal:
		return g.generateLiteral(v)
	case *ast.Null:
		// TODO: more complex null values
		return llvm.ConstNull(g.convType(v.Type()))
	case *ast.Identifier:
		if v.Sym.DefKind == common.DefKindFunc {
			return v.Sym.LLValue
		} else {
			return g.irb.BuildLoad(v.Sym.LLType, v.Sym.LLValue)
		}
	}

	report.ReportICE("expression codegen not implemented")
	return nil
}

// generateLHSExpr generates an LHS expression (mutable expression).
func (g *Generator) generateLHSExpr(expr ast.ASTExpr) llvm.Value {
	switch v := expr.(type) {
	case *ast.Identifier:
		// Already a pointer: no need to dereference
		return v.Sym.LLValue
	case *ast.Deref:
		// Already a pointer: no need to dereference
		return g.generateExpr(v.Ptr)
	default:
		report.ReportICE("LHS expression codegen not implemented")
		return nil
	}
}

// -----------------------------------------------------------------------------

// generateTypeCast generates a type cast.
func (g *Generator) generateTypeCast(tc *ast.TypeCast) llvm.Value {
	src := g.generateExpr(tc.SrcExpr)

	srcType := types.InnerType(tc.SrcExpr.Type())
	destType := types.InnerType(tc.Type())

	if types.Equals(srcType, destType) {
		return src
	}

	destLLType := g.convInnerType(destType, false, false)

	switch v := srcType.(type) {
	case types.PrimitiveType:
		return g.generatePrimTypeCast(src, v, destType.(types.PrimitiveType), destLLType)
	case *types.PointerType:
		// TODO: remove once it is no longer needed.
		return g.irb.BuildBitCast(src, destLLType)
	}

	report.ReportICE("codegen for cast not implemented")
	return nil
}

// generatePrimTypeCast generates a cast between primitive types.
func (g *Generator) generatePrimTypeCast(src llvm.Value, srcType, destType types.PrimitiveType, destLLType llvm.Type) llvm.Value {
	if srcType.IsIntegral() {
		if destType.IsFloating() {
			if srcType%2 == 0 {
				// unsigned to float
				return g.irb.BuildUIToFP(src, destLLType)
			} else {
				// signed to float
				return g.irb.BuildSIToFP(src, destLLType)
			}
		} else if destType.IsIntegral() {
			if srcType < destType {
				if destType-srcType == 1 {
					// signed to unsigned
					return src
				} else if srcType%2 == 1 && destType%2 == 1 {
					// small signed to large signed
					return g.irb.BuildSExt(src, destLLType)
				} else {
					// small signed to large unsigned
					// small unsigned to large signed
					// small unsigned to large unsigned
					return g.irb.BuildZExt(src, destLLType)
				}
			} else if srcType-destType == 1 {
				// unsigned to signed
				return src
			} else {
				// large int to small int
				return g.irb.BuildTrunc(src, destLLType)
			}
		} else if destType == types.PrimTypeBool {
			// int to bool
			return g.irb.BuildTrunc(src, destLLType)
		}
	} else if srcType.IsFloating() {
		if destType.IsFloating() {
			if srcType < destType {
				// small float to big float
				return g.irb.BuildFPExt(src, destLLType)
			} else {
				// big float to small float
				return g.irb.BuildFPTrunc(src, destLLType)
			}
		} else {
			if destType%2 == 0 {
				// float to unsigned int
				intrinsic := g.getIntrinsic("llvm.fptoui.sat", destLLType, src.Type())
				return g.callFunc(destType, intrinsic, src)
			} else {
				// float to signed int
				intrinsic := g.getIntrinsic("llvm.fptosi.sat", destLLType, src.Type())
				return g.callFunc(destType, intrinsic, src)
			}
		}
	} else if srcType == types.PrimTypeBool {
		// bool to int
		return g.irb.BuildZExt(src, destLLType)
	}

	report.ReportICE("codegen for primitive cast not implemented")
	return nil
}

// generateBinaryOpApp generates a binary operator application.
func (g *Generator) generateBinaryOpApp(bopApp *ast.BinaryOpApp) llvm.Value {
	lhs := g.generateExpr(bopApp.LHS)

	switch bopApp.Op.Overload.IntrinsicName {
	case "land":
		// TODO
	case "lor":
		// TODO
	}

	rhs := g.generateExpr(bopApp.RHS)

	switch bopApp.Op.Overload.IntrinsicName {
	default:
		// TODO: remaining intrinsic binary operators
		return g.callFunc(bopApp.Type(), bopApp.Op.Overload.LLValue, lhs, rhs)
	}
}

// generateUnaryOpApp generates a unary operator application.
func (g *Generator) generateUnaryOpApp(uopApp *ast.UnaryOpApp) llvm.Value {
	operand := g.generateExpr(uopApp.Operand)

	switch uopApp.Op.Overload.IntrinsicName {
	case "lnot", "compl":
		return g.irb.BuildNot(operand)
	case "ineg":
		return g.irb.BuildNeg(operand)
	case "fneg":
		return g.irb.BuildFNeg(operand)
	default:
		return g.callFunc(uopApp.Type(), uopApp.Op.Overload.LLValue, operand)
	}
}

// generateLiteral generates an LLVM literal.
func (g *Generator) generateLiteral(lit *ast.Literal) llvm.Value {
	switch lit.Kind {
	case syntax.TOK_RUNELIT:
		return llvm.ConstInt(g.ctx.Int32Type(), lit.Value.(uint64), false)
	case syntax.TOK_INTLIT:
		return llvm.ConstInt(g.convType(lit.Type()).(llvm.IntegerType), lit.Value.(uint64), false)
	case syntax.TOK_FLOATLIT:
		return llvm.ConstReal(g.convType(lit.Type()), lit.Value.(float64))
	case syntax.TOK_NUMLIT:
		if lit.Type().(types.PrimitiveType).IsFloating() {
			var fv float64
			if _fv, ok := lit.Value.(float64); ok {
				fv = _fv
			} else {
				fv = float64(lit.Value.(uint64))
			}

			return llvm.ConstReal(g.convType(lit.Type()), fv)
		} else {
			return llvm.ConstInt(g.convType(lit.Type()).(llvm.IntegerType), lit.Value.(uint64), false)
		}
	case syntax.TOK_BOOLLIT:
		if lit.Value.(bool) {
			return llvm.ConstInt(g.ctx.Int1Type(), 1, false)
		} else {
			return llvm.ConstInt(g.ctx.Int1Type(), 0, false)
		}
	default:
		report.ReportICE("literal codegen not implemented")
		return nil
	}
}
