package codegen

import (
	"chaic/ast"
	"chaic/report"
	"chaic/syntax"
	"chaic/types"

	"github.com/llir/llvm/ir/constant"
	lltypes "github.com/llir/llvm/ir/types"
	llvalue "github.com/llir/llvm/ir/value"
)

// generateVarDecl generates a variable declaration.
func (g *Generator) generateVarDecl(vd *ast.VarDecl) {
	for _, varList := range vd.VarLists {
		// Allocate all the variables.
		for _, v := range varList.Vars {
			v.Sym.LLValue = g.varBlock.NewAlloca(g.convAllocType(v.Type()))
		}

		// Get the initializer.
		var initExpr llvalue.Value

		if varList.Initializer == nil {
			// TODO: handle more complex null values
			initExpr = g.generateNullValue(varList.Vars[0].Type())
		} else {
			initExpr = g.generateExpr(varList.Initializer)
		}

		// Initialize all the variables.
		for _, v := range varList.Vars {
			g.block.NewStore(initExpr, v.Sym.LLValue)
		}
	}
}

// generateAssignment generates an assignment statement.
func (g *Generator) generateAssignment(assign *ast.Assignment) {
	if len(assign.LHSVars) == len(assign.RHSExprs) {
		// Generate the LHS variables.
		llLhsVars := make([]llvalue.Value, len(assign.LHSVars))
		for i, lhsVar := range assign.LHSVars {
			llLhsVars[i] = g.generateLHSExpr(lhsVar)
		}

		// Generate the RHS values.
		llRhsValues := make([]llvalue.Value, len(assign.RHSExprs))
		for i, rhsExpr := range assign.RHSExprs {
			llRhsValues[i] = g.generateExpr(rhsExpr)
		}

		// Apply compound operators if necessary.
		if assign.CompoundOp != nil {
			// TODO: handle short-circuiting binary operators?
			for i, llRhsValue := range llRhsValues {
				llLhsValue := g.block.NewLoad(g.convType(assign.LHSVars[i].Type()), llLhsVars[i])

				llRhsValues[i] = g.applyNonSSBinaryOp(assign.CompoundOp.Overload, llLhsValue, llRhsValue)
			}
		}

		// Store the RHS values into the LHS values.
		for i, llLhsVar := range llLhsVars {
			g.block.NewStore(llRhsValues[i], llLhsVar)
		}
	} else {
		report.ReportICE("codegen for pattern matching assignment not implemented")
	}
}

// generateIncDec generates an increment/decrement statement.
func (g *Generator) generateIncDec(incdec *ast.IncDecStmt) {
	// Generate the LHS operand.
	llLHSOperand := g.generateLHSExpr(incdec.LHSOperand)

	// Generate the one-value to use the increment/decrement the operand.
	var oneValue llvalue.Value
	if pt, ok := types.InnerType(incdec.LHSOperand.Type()).(types.PrimitiveType); ok {
		if pt.IsFloating() {
			oneValue = constant.NewFloat(g.convPrimType(pt, false).(*lltypes.FloatType), 1)
		} else {
			oneValue = constant.NewInt(g.convPrimType(pt, false).(*lltypes.IntType), 1)
		}
	} else {
		report.ReportICE("codegen for non-primitive increment not implemented")
	}

	// Extract the value of the LHS operand.
	llLHSValue := g.block.NewLoad(g.convType(incdec.LHSOperand.Type()), llLHSOperand)

	// Apply the operator.
	llResult := g.applyNonSSBinaryOp(incdec.Op.Overload, llLHSValue, oneValue)

	// Store the result into the LHS operand.
	g.block.NewStore(llResult, llLHSOperand)
}

/* -------------------------------------------------------------------------- */

// generateReturnStmt generates a return statement,
func (g *Generator) generateReturnStmt(ret *ast.ReturnStmt) {
	switch len(ret.Exprs) {
	case 0:
		g.block.NewRet(nil)
		return
	case 1:
		g.block.NewRet(g.generateExpr(ret.Exprs[0]))
		return
	default:
		report.ReportICE("codegen for multi-return not implemented")
	}
}

// generateKeywordStmt generates a keyword control statement.
func (g *Generator) generateKeywordStmt(keyStmt *ast.KeywordStmt) {
	switch keyStmt.Kind {
	case syntax.TOK_BREAK:
		g.block.NewBr(g.currLoopContext().breakDest)
	case syntax.TOK_CONTINUE:
		g.block.NewBr(g.currLoopContext().continueDest)
	}
}
