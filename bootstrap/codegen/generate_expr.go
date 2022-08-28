package codegen

import (
	"chaic/ast"
	"chaic/common"
	"chaic/report"
	"chaic/syntax"
	"chaic/types"
	"fmt"

	"github.com/llir/llvm/ir"
	"github.com/llir/llvm/ir/constant"
	"github.com/llir/llvm/ir/enum"
	lltypes "github.com/llir/llvm/ir/types"
	llvalue "github.com/llir/llvm/ir/value"
)

// generateExpr generates an expression.
func (g *Generator) generateExpr(expr ast.ASTExpr) llvalue.Value {
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

			args := make([]llvalue.Value, len(v.Args))
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
			return nil
		}
	case *ast.Deref:
		return g.block.NewLoad(g.convType(v.Ptr.Type()), g.generateExpr(v.Ptr))
	case *ast.Literal:
		return g.generateLiteral(v)
	case *ast.Null:
		return g.generateNullValue(v.NodeType)
	case *ast.Identifier:
		if v.Sym.DefKind == common.DefKindFunc {
			return v.Sym.LLValue
		} else {
			return g.block.NewLoad(v.Sym.LLValue.Type().(*lltypes.PointerType).ElemType, v.Sym.LLValue)
		}
	default:
		report.ReportICE("codegen for expr not implemented")
		return nil
	}
}

// generateLHSExpr generates an LHS expression (mutable expression).
func (g *Generator) generateLHSExpr(expr ast.ASTExpr) llvalue.Value {
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

/* -------------------------------------------------------------------------- */

// generateTypeCast generates a type cast.
func (g *Generator) generateTypeCast(tc *ast.TypeCast) llvalue.Value {
	// Generate the source expression.
	src := g.generateExpr(tc.SrcExpr)

	// Extract the inner types we are converting between.
	srcType := types.InnerType(tc.SrcExpr.Type())
	destType := types.InnerType(tc.Type())

	// If the two types are already equal, then no cast is needed: we can just
	// return the source value.
	if types.Equals(srcType, destType) {
		return src
	}

	// Convert the destination type to its LLVM type.
	destLLType := g.convInnerType(destType, false, false)

	switch v := srcType.(type) {
	case types.PrimitiveType:
		return g.generatePrimTypeCast(src, v, destType.(types.PrimitiveType), destLLType)
	case *types.PointerType:
		// TODO: remove once it is no longer needed.
		return g.block.NewBitCast(src, destLLType)
	}

	report.ReportICE("codegen for cast not implemented")
	return nil
}

// generatePrimTypeCast generates a cast between primitive types.
func (g *Generator) generatePrimTypeCast(src llvalue.Value, srcType, destType types.PrimitiveType, destLLType lltypes.Type) llvalue.Value {
	if srcType.IsIntegral() {
		if destType.IsFloating() {
			if srcType%2 == 0 {
				// unsigned to float
				return g.block.NewUIToFP(src, destLLType)
			} else {
				// signed to float
				return g.block.NewSIToFP(src, destLLType)
			}
		} else if destType.IsIntegral() {
			if srcType < destType {
				if destType-srcType == 1 {
					// signed to unsigned
					return src
				} else if srcType%2 == 1 && destType%2 == 1 {
					// small signed to large signed
					return g.block.NewSExt(src, destLLType)
				} else {
					// small signed to large unsigned
					// small unsigned to large signed
					// small unsigned to large unsigned
					return g.block.NewZExt(src, destLLType)
				}
			} else if srcType-destType == 1 {
				// unsigned to signed
				return src
			} else {
				// large int to small int
				return g.block.NewTrunc(src, destLLType)
			}
		} else if destType == types.PrimTypeBool {
			// int to bool
			return g.block.NewTrunc(src, destLLType)
		}
	} else if srcType.IsFloating() {
		if destType.IsFloating() {
			if srcType < destType {
				// small float to big float
				return g.block.NewFPExt(src, destLLType)
			} else {
				// big float to small float
				return g.block.NewFPTrunc(src, destLLType)
			}
		} else {
			if destType%2 == 0 {
				// float to unsigned int
				intrinsic := g.getIntrinsic(
					fmt.Sprintf(
						"llvm.fptoui.sat.%s.%s",
						destLLType.LLString(),
						src.Type().LLString(),
					), destLLType, src.Type())
				return g.callFunc(destType, intrinsic, src)
			} else {
				// float to signed int
				intrinsic := g.getIntrinsic(
					fmt.Sprintf(
						"llvm.fptosi.sat.%s.%s",
						destLLType.LLString(),
						src.Type().LLString(),
					), destLLType, src.Type())
				return g.callFunc(destType, intrinsic, src)
			}
		}
	} else if srcType == types.PrimTypeBool {
		// bool to int
		return g.block.NewZExt(src, destLLType)
	}

	report.ReportICE("codegen for primitive cast not implemented")
	return nil
}

/* -------------------------------------------------------------------------- */

// generateBinaryOpApp generates a binary operator application.
func (g *Generator) generateBinaryOpApp(bopApp *ast.BinaryOpApp) llvalue.Value {
	lhs := g.generateExpr(bopApp.LHS)

	// Handle short-circuiting binary operators.
	switch bopApp.Op.Overload.IntrinsicName {
	case "land":
		// The LLVM IR for logical AND is roughly:
		//
		//   some_block:
		//     ...
		//     %lhs = ...
		//     br i1 %lhs, label %land_maybe, label %land_exit
		//
		//   land_maybe:
		//     %rhs = ...
		//     br label %land_exit
		//
		//   land_exit:
		//     %land = phi i1 [i1 false, label %some_block], [i1 %rhs, label %land_maybe]

		{
			landEntry := g.block

			landMaybe := g.appendBlock()
			landExit := g.appendBlock()

			g.block.NewCondBr(lhs, landMaybe, landExit)

			g.block = landMaybe
			rhs := g.generateExpr(bopApp.RHS)
			g.block.NewBr(landExit)

			// Handle the case where generating the RHS changes the block.
			landMaybe = g.block

			g.block = landExit
			return g.block.NewPhi(
				ir.NewIncoming(constant.NewBool(false), landEntry),
				ir.NewIncoming(rhs, landMaybe),
			)
		}

	case "lor":
		// The LLVM IR for logical OR is roughyl:
		//
		//   some_block:
		//     ...
		//     %lhs = ...
		//     br i1 %lhs, label %lor_exit, label %lor_maybe
		//
		//   lor_maybe:
		//     %rhs = ...
		//     br label %lor_exit
		//
		//   lor_exit:
		//     %lor = phi i1 [i1 true, label %some_block], [i1 %rhs, label %lor_maybe]

		{
			lorEntry := g.block

			lorMaybe := g.appendBlock()
			lorExit := g.appendBlock()

			g.block.NewCondBr(lhs, lorExit, lorMaybe)

			g.block = lorMaybe
			rhs := g.generateExpr(bopApp.RHS)
			g.block.NewBr(lorExit)

			// Handle the case when the RHS changes the block.
			lorMaybe = g.block

			g.block = lorExit
			return g.block.NewPhi(
				ir.NewIncoming(constant.NewBool(true), lorEntry),
				ir.NewIncoming(rhs, lorMaybe),
			)
		}
	}

	return g.applyNonSSBinaryOp(bopApp.Op.Overload, lhs, g.generateExpr(bopApp.RHS))
}

// applyNonSSBinaryOp applies a non-short-circuiting binary operator repreanted
// by overload to lhs and rhs.
func (g *Generator) applyNonSSBinaryOp(overload *common.OperatorOverload, lhs, rhs llvalue.Value) llvalue.Value {
	switch overload.IntrinsicName {
	case "iadd":
		return g.block.NewAdd(lhs, rhs)
	case "isub":
		return g.block.NewSub(lhs, rhs)
	case "imul":
		return g.block.NewMul(lhs, rhs)
	case "sdiv":
		return g.block.NewSDiv(lhs, rhs)
	case "smod":
		return g.block.NewSRem(lhs, rhs)
	case "ieq":
		return g.block.NewICmp(enum.IPredEQ, lhs, rhs)
	case "ineq":
		return g.block.NewICmp(enum.IPredNE, lhs, rhs)
	case "slt":
		return g.block.NewICmp(enum.IPredSLT, lhs, rhs)
	case "sgt":
		return g.block.NewICmp(enum.IPredSGT, lhs, rhs)
	case "slteq":
		return g.block.NewICmp(enum.IPredSLE, lhs, rhs)
	case "sgteq":
		return g.block.NewICmp(enum.IPredSGE, lhs, rhs)
	case "":
		return g.block.NewCall(overload.LLValue, lhs, rhs)
	default:
		report.ReportICE("binary non-SS intrinsic codegen for `%s` not implemented", overload.IntrinsicName)
		return nil
	}
}

/* -------------------------------------------------------------------------- */

// generateUnaryOpApp generates a unary operator application.
func (g *Generator) generateUnaryOpApp(uopApp *ast.UnaryOpApp) llvalue.Value {
	operand := g.generateExpr(uopApp.Operand)

	switch uopApp.Op.Overload.IntrinsicName {
	case "lnot", "compl":
		{
			// ~op == op ^ 0b111'1111
			operandIntType := operand.Type().(*lltypes.IntType)
			return g.block.NewXor(operand, constant.NewInt(
				operandIntType,
				^(^int64(0)<<int64(operandIntType.BitSize)),
			))
		}
	case "ineg":
		// -op = 0 - op
		return g.block.NewSub(constant.NewInt(operand.Type().(*lltypes.IntType), 0), operand)
	case "fneg":
		return g.block.NewFNeg(operand)
	case "":
		return g.callFunc(uopApp.Type(), uopApp.Op.Overload.LLValue, operand)
	default:
		report.ReportICE("unary intrinsic codegen not implemented")
		return nil
	}
}

// generateLiteral generates an LLVM literal.
func (g *Generator) generateLiteral(lit *ast.Literal) llvalue.Value {
	switch lit.Kind {
	case syntax.TOK_RUNELIT:
		return constant.NewInt(lltypes.I32, int64(lit.Value.(int32)))
	case syntax.TOK_INTLIT:
		return constant.NewInt(g.convType(lit.Type()).(*lltypes.IntType), lit.Value.(int64))
	case syntax.TOK_FLOATLIT:
		return constant.NewFloat(g.convType(lit.Type()).(*lltypes.FloatType), lit.Value.(float64))
	case syntax.TOK_NUMLIT:
		if types.InnerType(lit.Type()).(types.PrimitiveType).IsFloating() {
			var fv float64
			if _fv, ok := lit.Value.(float64); ok {
				fv = _fv
			} else {
				fv = float64(uint64(lit.Value.(int64)))
			}

			return constant.NewFloat(g.convType(lit.Type()).(*lltypes.FloatType), fv)
		} else {
			return constant.NewInt(g.convType(lit.Type()).(*lltypes.IntType), lit.Value.(int64))
		}
	case syntax.TOK_BOOLLIT:
		return constant.NewBool(lit.Value.(bool))
	default:
		report.ReportICE("literal codegen not implemented")
		return nil
	}
}

// generateNullValue generates a LLVM null value.
func (g *Generator) generateNullValue(typ types.Type) llvalue.Value {
	switch v := types.InnerType(typ).(type) {
	case types.PrimitiveType:
		if v.IsFloating() {
			return constant.NewFloat(g.convPrimType(v, false).(*lltypes.FloatType), 0)
		} else if v == types.PrimTypeUnit {
			return constant.NewStruct(lltypes.NewStruct())
		} else {
			// All other primitives compile as int constants.
			return constant.NewInt(g.convPrimType(v, false).(*lltypes.IntType), 0)
		}
	case *types.PointerType:
		return constant.NewNull(g.convInnerType(v, false, false).(*lltypes.PointerType))
	}

	report.ReportICE("null codegen not implemented")
	return nil
}
