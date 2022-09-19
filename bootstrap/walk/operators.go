package walk

import (
	"chaic/ast"
	"chaic/common"
	"chaic/report"
	"chaic/types"
	"chaic/util"
	"strings"
)

// checkOperApp checks an operator application.  It returns the resultant type
// of the operator application.
func (w *Walker) checkOperApp(op *ast.AppliedOperator, span *report.TextSpan, args ...ast.ASTExpr) types.Type {
	argTypes := util.Map(args, func(expr ast.ASTExpr) types.Type { return expr.Type() })

	if opMethod := w.getOperatorMethod(op.OpKind, argTypes...); opMethod != nil {
		op.OpMethod = opMethod
		return op.OpMethod.Signature.ReturnType
	}

	w.error(
		span,
		"no definition of %s matches argument types (%s)",
		op.OpName,
		strings.Join(util.Map(argTypes, func(typ types.Type) string { return typ.Repr() }), ","),
	)
	return nil
}

// getOperatorMethod gets the operator method corresponding to the operator with
// kind kind that accepts args if such an operator method exists.
func (w *Walker) getOperatorMethod(kind int, args ...types.Type) *common.OperatorMethod {
	// TODO: implement non-intrinsic operators.

	// Handle intrinsic operator calls.
	if len(args) == 1 {
		return getIntrinsicUnaryOperator(kind, args[0])
	} else if len(args) == 2 {
		return getIntrinsicBinaryOperator(kind, args[0], args[1])
	}

	return nil
}

// getIntrinsicUnaryOperator tries to get an intrinsic unary operator method.
func getIntrinsicUnaryOperator(kind int, operand types.Type) *common.OperatorMethod {
	// TODO
	return nil
}

// getIntrinsicBinaryOperator tries to get an intrinsic binary operator method.
func getIntrinsicBinaryOperator(kind int, lhs, rhs types.Type) *common.OperatorMethod {
	// TODO
	return nil
}
