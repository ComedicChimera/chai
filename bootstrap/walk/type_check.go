package walk

import (
	"chaic/ast"
	"chaic/common"
	"chaic/report"
	"chaic/types"
	"chaic/util"
	"strings"
)

// mustUnify asserts that two types must be equal or able to be made equal.
func (w *Walker) mustUnify(expected, actual types.Type, span *report.TextSpan) {
	if !types.Unify(expected, actual) {
		w.error(span, "expected type %s but got %s", expected, actual)
	}
}

// checkOperApp walks an operator application.  The operands should already be walked.
func (w *Walker) checkOperApp(op *common.AppliedOperator, span *report.TextSpan, operands ...ast.ASTExpr) types.Type {
	// Find the matching overload.
	overload := w.mustFindOverload(op, span, operands)

	// Get the overload signature.
	signature := overload.Signature.(*types.FuncType)

	// Unify the operands.
	for i, operand := range operands {
		w.mustUnify(signature.ParamTypes[i], operand.Type(), operand.Span())
	}

	// Return the resultant type of the operator application.
	return signature.ReturnType
}

// mustFindOverload attempts to find the first overload of the given
// operator matching the passed arguments.  If no such overload can be found,
// then an error is reported.
func (w *Walker) mustFindOverload(aop *common.AppliedOperator, span *report.TextSpan, operands []ast.ASTExpr) *common.OperatorOverload {
	// Lookup the operator definitions.
	// TODO: local operator overloads
	opers, ok := w.chFile.Parent.OperatorTable[aop.Kind]
	if !ok {
		w.error(aop.Span, "no visible overloads for the `%s` operator", aop.OpRepr)
	}

	// Find a set overloads matching the operator application's arity.
	var overloads []*common.OperatorOverload
	for _, oper := range opers {
		if oper.Arity == len(operands) {
			overloads = oper.Overloads
			break
		}
	}

	if len(overloads) == 0 {
		w.error(aop.Span, "no visible overloads of the `%s` operator take %d operands", aop.OpRepr, len(operands))
	}

	// Find all the overloads which match the argument types.
	var matchingOverloads []*common.OperatorOverload
matchLoop:
	for _, overload := range overloads {
		oft := overload.Signature.(*types.FuncType)

		for i, opt := range oft.ParamTypes {
			if !types.Equals(opt, operands[i].Type()) {
				continue matchLoop
			}
		}

		matchingOverloads = append(matchingOverloads, overload)
	}

	if len(matchingOverloads) == 0 {
		w.error(
			span,
			"no visible overloads of the `%s` operator match the operand types: %s",
			aop.OpRepr,
			strings.Join(util.Map(operands, func(expr ast.ASTExpr) string {
				return expr.Type().Repr()
			}), ", "),
		)
	}

	// If there is more than one overload, then we just default and pick the
	// first matching one: this won't happen all that often, but when it does,
	// it should happen relatively predictably (since the ordering of overload
	// declarations should be preserved).
	aop.Overload = matchingOverloads[0]
	return matchingOverloads[0]
}
