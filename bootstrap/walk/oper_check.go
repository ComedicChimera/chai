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

		if opMethod.ID == common.OP_ID_UNKNOWN {
			w.unknownOperators = append(w.unknownOperators, op)
		}

		return op.OpMethod.Signature.ReturnType
	}

	w.error(
		span,
		"no definition of %s matches argument types (%s)",
		op.OpName,
		strings.Join(util.Map(argTypes, func(typ types.Type) string { return typ.Repr() }), ", "),
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

/* -------------------------------------------------------------------------- */

// getIntrinsicBinaryOperator tries to get an intrinsic binary operator method.
func getIntrinsicBinaryOperator(kind int, lhs, rhs types.Type) *common.OperatorMethod {
	lhs = types.InnerType(lhs)
	rhs = types.InnerType(rhs)

	if lpt, ok := lhs.(types.PrimitiveType); ok {
		return findIntrinsicBinaryOperatorForPrimitive(kind, lpt, rhs)
	} else if rpt, ok := rhs.(types.PrimitiveType); ok {
		return findIntrinsicBinaryOperatorForPrimitive(kind, rpt, lhs)
	} else if _, ok := lhs.(*types.UntypedNumber); ok {
		if types.Equals(lhs, rhs) {
			types.Unify(lhs, rhs)

			return &common.OperatorMethod{
				ID: common.OP_ID_UNKNOWN,
				Signature: &types.FuncType{
					ParamTypes: []types.Type{lhs, lhs},
					ReturnType: getIntrinsicBinOpReturnType(kind, lhs),
				},
			}
		}
	} else if _, ok := rhs.(*types.UntypedNumber); ok {
		if types.Equals(lhs, rhs) {
			types.Unify(lhs, rhs)

			return &common.OperatorMethod{
				ID: common.OP_ID_UNKNOWN,
				Signature: &types.FuncType{
					ParamTypes: []types.Type{rhs, rhs},
					ReturnType: getIntrinsicBinOpReturnType(kind, rhs),
				},
			}
		}
	}

	return nil
}

// Table mapping floating-point binary operator kinds to intrinsic operator IDs.
var floatBinOpKindToIntrinsicID = map[int]uint64{
	syntax.TOK_PLUS:  common.OP_ID_FADD,
	syntax.TOK_MINUS: common.OP_ID_FSUB,
	syntax.TOK_STAR:  common.OP_ID_FADD,
	syntax.TOK_DIV:   common.OP_ID_FDIV,
	syntax.TOK_MOD:   common.OP_ID_FMOD,

	syntax.TOK_EQ:   common.OP_ID_FEQ,
	syntax.TOK_NEQ:  common.OP_ID_FNEQ,
	syntax.TOK_LT:   common.OP_ID_FLT,
	syntax.TOK_GT:   common.OP_ID_FGT,
	syntax.TOK_LTEQ: common.OP_ID_FLTEQ,
	syntax.TOK_GTEQ: common.OP_ID_FGTEQ,
}

// Table mapping signed integer binary operator kinds to intrinsic operator IDs.
var sintBinOpKindToIntrinsicID = map[int]uint64{
	syntax.TOK_DIV:  common.OP_ID_SDIV,
	syntax.TOK_MOD:  common.OP_ID_SMOD,
	syntax.TOK_LT:   common.OP_ID_SLT,
	syntax.TOK_GT:   common.OP_ID_SGT,
	syntax.TOK_LTEQ: common.OP_ID_SLTEQ,
	syntax.TOK_GTEQ: common.OP_ID_SGTEQ,
}

// Table mapping unsigned integer binary operator kinds to intrinsic operator IDs.
var uintBinOpKindToIntrinsicID = map[int]uint64{
	syntax.TOK_DIV:  common.OP_ID_UDIV,
	syntax.TOK_MOD:  common.OP_ID_UMOD,
	syntax.TOK_LT:   common.OP_ID_ULT,
	syntax.TOK_GT:   common.OP_ID_UGT,
	syntax.TOK_LTEQ: common.OP_ID_ULTEQ,
	syntax.TOK_GTEQ: common.OP_ID_UGTEQ,
}

// mapIntBinOpKindToIntrinsicID maps an integer binary operator kind to its
// corresponding intrinsic operator ID.
func mapIntBinOpKindToIntrinsicID(kind int, pt types.PrimitiveType) uint64 {
	switch kind {
	case syntax.TOK_PLUS:
		return common.OP_ID_IADD
	case syntax.TOK_MINUS:
		return common.OP_ID_ISUB
	case syntax.TOK_STAR:
		return common.OP_ID_IMUL
	case syntax.TOK_BWAND:
		return common.OP_ID_BWAND
	case syntax.TOK_BWOR:
		return common.OP_ID_BWOR
	case syntax.TOK_LSHIFT:
		return common.OP_ID_BWSHL
	case syntax.TOK_RSHIFT:
		return common.OP_ID_BWSHR
	case syntax.TOK_EQ:
		return common.OP_ID_IEQ
	case syntax.TOK_NEQ:
		return common.OP_ID_INEQ
	default:
		if pt%2 == 0 { // unsigned
			return uintBinOpKindToIntrinsicID[kind]
		} else { // signed
			return sintBinOpKindToIntrinsicID[kind]
		}
	}
}

// getIntrinsicBinOpReturnType gets the return type of an intrinsic binary
// operator based on its kind and one of its arguments.
func getIntrinsicBinOpReturnType(kind int, arg types.Type) types.Type {
	if syntax.TOK_EQ <= kind && kind <= syntax.TOK_GTEQ {
		return types.PrimTypeBool
	} else {
		return arg
	}
}

// findIntrinsicBinaryOperatorForPrimitive attempts to find an intrinsic binary
// operator matching the given primitive type and another type as its arguments.
func findIntrinsicBinaryOperatorForPrimitive(kind int, pt types.PrimitiveType, other types.Type) *common.OperatorMethod {
	signature := &types.FuncType{ParamTypes: []types.Type{pt, pt}, ReturnType: getIntrinsicBinOpReturnType(kind, pt)}

	if pt.IsIntegral() {
		if syntax.TOK_PLUS <= kind && kind <= syntax.TOK_RSHIFT {
			if types.Equals(pt, other) {
				types.Unify(pt, other)
			} else {
				return nil
			}

			return &common.OperatorMethod{
				ID:        mapIntBinOpKindToIntrinsicID(kind, pt),
				Signature: signature,
			}
		}
	} else if pt.IsFloating() {
		if syntax.TOK_PLUS <= kind && kind <= syntax.TOK_GTEQ {
			if types.Equals(pt, other) {
				types.Unify(pt, other)
			} else {
				return nil
			}

			return &common.OperatorMethod{
				ID:        floatBinOpKindToIntrinsicID[kind],
				Signature: signature,
			}
		}
	} else if pt == types.PrimTypeBool {
		var opID uint64
		switch kind {
		case syntax.TOK_LAND:
			opID = common.OP_ID_LAND
		case syntax.TOK_LOR:
			opID = common.OP_ID_LOR
		case syntax.TOK_EQ:
			opID = common.OP_ID_IEQ
		case syntax.TOK_NEQ:
			opID = common.OP_ID_INEQ
		default:
			return nil
		}

		return &common.OperatorMethod{
			ID:        opID,
			Signature: signature,
		}
	}

	return nil
}
