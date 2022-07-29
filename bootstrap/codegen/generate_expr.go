package codegen

import (
	"chaic/ast"
	"chaic/common"
	"chaic/report"
	"chaic/syntax"
	"chaic/types"

	"github.com/llir/llvm/ir/constant"
	lltypes "github.com/llir/llvm/ir/types"
	llvalue "github.com/llir/llvm/ir/value"
)

// generateExpr generates an expression.
func (g *Generator) generateExpr(expr ast.ASTExpr) llvalue.Value {
	switch v := expr.(type) {
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
			return g.block.NewLoad(v.Sym.LLValue.Type(), v.Sym.LLValue)
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

// generateBinaryOpApp generates a binary operator application.
func (g *Generator) generateBinaryOpApp(bopApp *ast.BinaryOpApp) llvalue.Value {
	lhs := g.generateExpr(bopApp.LHS)
	rhs := g.generateExpr(bopApp.RHS)

	switch bopApp.Op.Overload.IntrinsicName {
	case "":
		return g.block.NewCall(bopApp.Op.Overload.LLValue, lhs, rhs)
	default:
		report.ReportICE("binary intrinsic codegen not implemented")
		return nil
	}
}

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
		return constant.NewInt(lltypes.I32, lit.Value.(int64))
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
