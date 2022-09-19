package walk

import (
	"chaic/ast"
	"chaic/common"
	"chaic/report"
	"chaic/syntax"
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
	operand = types.InnerType(operand)

	signature := &types.FuncType{ParamTypes: []types.Type{operand}, ReturnType: operand}

	if pt, ok := operand.(types.PrimitiveType); ok {
		switch kind {
		case syntax.TOK_NOT:
			if pt == types.PrimTypeBool {
				return &common.OperatorMethod{
					ID:        common.OP_ID_LNOT,
					Signature: signature,
				}
			}
		case syntax.TOK_COMPL:
			if pt.IsIntegral() {
				return &common.OperatorMethod{
					ID:        common.OP_ID_BWCOMPL,
					Signature: signature,
				}
			}
		case syntax.TOK_MINUS:
			if pt.IsIntegral() {
				return &common.OperatorMethod{
					ID:        common.OP_ID_INEG,
					Signature: signature,
				}
			} else if pt.IsFloating() {
				return &common.OperatorMethod{
					ID:        common.OP_ID_FNEG,
					Signature: signature,
				}
			}
		}
	} else if un, ok := operand.(*types.UntypedNumber); ok {
		if kind == syntax.TOK_MINUS {
			return &common.OperatorMethod{
				ID:        common.OP_ID_UNKNOWN,
				Signature: signature,
			}
		} else if kind == syntax.TOK_COMPL {
			// Make sure the type isn't a float type.
			if un.DisplayName == "untyped float literal" {
				return nil
			}

			// Remove all floating types from the possible types.
			filteredPossibleTypes := make([]types.PrimitiveType, len(un.PossibleTypes))
			n := 0
			for _, pt := range un.PossibleTypes {
				if !pt.IsFloating() {
					filteredPossibleTypes[n] = pt
					n++
				}
			}

			filteredPossibleTypes = filteredPossibleTypes[:n]

			// If no types remain, then it was still a float type and thus invalid.
			if len(filteredPossibleTypes) == 0 {
				return nil
			}

			// Update the possible types.
			un.PossibleTypes = filteredPossibleTypes

			return &common.OperatorMethod{
				ID:        common.OP_ID_UNKNOWN,
				Signature: signature,
			}
		}
	}

	return nil
}

// getIntrinsicBinaryOperator tries to get an intrinsic binary operator method.
func getIntrinsicBinaryOperator(kind int, lhs, rhs types.Type) *common.OperatorMethod {
	lhs = types.InnerType(lhs)
	rhs = types.InnerType(rhs)

	// TODO

	return nil
}

/* -------------------------------------------------------------------------- */
