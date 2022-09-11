package codegen

import (
	"chaic/mir"
	"chaic/report"
	"chaic/types"
	"fmt"

	"github.com/llir/llvm/ir"
	"github.com/llir/llvm/ir/constant"
	"github.com/llir/llvm/ir/enum"
	lltypes "github.com/llir/llvm/ir/types"
	llvalue "github.com/llir/llvm/ir/value"
)

// generateExpr generates an expression.
func (g *Generator) generateExpr(expr mir.Expr) llvalue.Value {
	switch v := expr.(type) {
	case *mir.TypeCast:
		return g.generateTypeCast(v)
	case *mir.BinaryOpApp:
		return g.generateBinaryOpApp(v)
	case *mir.UnaryOpApp:
		return g.generateUnaryOpApp(v)
	case *mir.FuncCall:
		return g.generateFuncCall(v)
	case *mir.ConstInt:
		return constant.NewInt(
			g.convType(expr.Type()).(*lltypes.IntType),
			v.IntValue,
		)
	case *mir.ConstReal:
		return constant.NewFloat(
			g.convType(expr.Type()).(*lltypes.FloatType),
			v.FloatValue,
		)
	case *mir.ConstUnit:
		return constant.NewStruct(lltypes.NewStruct())
	case *mir.ConstNullPtr:
		return constant.NewNull(g.convType(expr.Type()).(*lltypes.PointerType))
	default:
		report.ReportICE("codegen for expr not implemented")
		return nil
	}
}

// generateLHSExpr generates an LHS expression (mutable expression).
func (g *Generator) generateLHSExpr(expr mir.Expr) llvalue.Value {
	switch v := expr.(type) {
	case *mir.Identifier:
		// Already a pointer: no need to dereference
		return v.Symbol.LLValue
	case *mir.Deref:
		// Already a pointer: no need to dereference
		return g.generateExpr(v.Ptr)
	default:
		report.ReportICE("LHS expression codegen not implemented")
		return nil
	}
}

/* -------------------------------------------------------------------------- */

// generateTypeCast generates a type cast.
func (g *Generator) generateTypeCast(tc *mir.TypeCast) llvalue.Value {
	// Generate the source expression.
	src := g.generateExpr(tc.SrcExpr)

	// If the two types are already equal, then no cast is needed: we can just
	// return the source value.
	if types.Equals(tc.SrcExpr.Type(), tc.DestType) {
		return src
	}

	// Convert the destination type to its LLVM type.
	destLLType := g.convType(tc.DestType)

	switch v := tc.SrcExpr.Type().(type) {
	case types.PrimitiveType:
		return g.generatePrimTypeCast(src, v, tc.DestType.(types.PrimitiveType), destLLType)
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
					),
					destLLType,
					src.Type(),
				)

				return g.block.NewCall(
					intrinsic,
					src,
				)
			} else {
				// float to signed int
				intrinsic := g.getIntrinsic(
					fmt.Sprintf(
						"llvm.fptosi.sat.%s.%s",
						destLLType.LLString(),
						src.Type().LLString(),
					),
					destLLType,
					src.Type(),
				)

				return g.block.NewCall(
					intrinsic,
					src,
				)
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
func (g *Generator) generateBinaryOpApp(bopApp *mir.BinaryOpApp) llvalue.Value {
	lhs := g.generateExpr(bopApp.LHS)

	// Handle short-circuiting binary operators.
	switch bopApp.OpName {
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

	return g.applyNonSSBinaryOp(bopApp.OpName, lhs, g.generateExpr(bopApp.RHS))
}

// applyNonSSBinaryOp applies a non-short-circuiting binary operator repreanted
// by overload to lhs and rhs.
func (g *Generator) applyNonSSBinaryOp(opName string, lhs, rhs llvalue.Value) llvalue.Value {
	switch opName {
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
	default:
		report.ReportICE("binary non-SS codegen for `%s` not implemented", opName)
		return nil
	}
}

// generateUnaryOpApp generates a unary operator application.
func (g *Generator) generateUnaryOpApp(uopApp *mir.UnaryOpApp) llvalue.Value {
	operand := g.generateExpr(uopApp.Operand)

	switch uopApp.OpName {
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
	default:
		report.ReportICE("unary intrinsic codegen not implemented")
		return nil
	}
}

/* -------------------------------------------------------------------------- */

// generateFuncCall generates a function call.
func (g *Generator) generateFuncCall(call *mir.FuncCall) llvalue.Value {
	return nil
}
