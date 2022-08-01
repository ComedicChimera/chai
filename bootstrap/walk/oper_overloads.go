package walk

import (
	"chaic/ast"
	"chaic/common"
	"chaic/report"
	"chaic/types"
	"chaic/util"
)

// checkOperApp walks an operator application.  The operands should already be walked.
func (w *Walker) checkOperApp(op *common.AppliedOperator, span *report.TextSpan, operands ...ast.ASTExpr) types.Type {
	// Find the matching overloads.
	overloads := w.getOverloads(op, span, len(operands))

	// Create a new type variable for the operator overloads.
	opVar := w.solver.NewTypeVar(op.OpRepr, op.Span)
	w.solver.AddOperatorOverloads(
		opVar,
		util.Map(overloads, func(overload *common.OperatorOverload) types.Type { return overload.Signature }),
		func(ndx int) {
			op.Overload = overloads[ndx]
		},
	)

	// Create a new type variable for the return type.
	returnType := w.solver.NewTypeVar("return type", span)

	// Create a new type to represent the operands.
	operandFuncType := &types.FuncType{
		ParamTypes: util.Map(operands, func(expr ast.ASTExpr) types.Type { return expr.Type() }),
		ReturnType: returnType,
	}

	// Unify it with the operands.
	w.solver.MustEqual(opVar, operandFuncType, span)

	// Return the resultant type of the operator application.
	return returnType
}

// getOverloads attempts to find the first overload of the given
// operator matching the passed arguments.  If no such overload can be found,
// then an error is reported.
func (w *Walker) getOverloads(aop *common.AppliedOperator, span *report.TextSpan, arity int) []*common.OperatorOverload {
	// Lookup the operator definitions.
	// TODO: local operator overloads
	opers, ok := w.chFile.Parent.OperatorTable[aop.Kind]
	if !ok {
		w.error(aop.Span, "no visible overloads for the `%s` operator", aop.OpRepr)
	}

	// Find a set overloads matching the operator application's arity.
	var overloads []*common.OperatorOverload
	for _, oper := range opers {
		if oper.Arity == arity {
			overloads = oper.Overloads
			break
		}
	}

	if len(overloads) == 0 {
		w.error(aop.Span, "no visible overloads of the `%s` operator take %d operands", aop.OpRepr, arity)
	}

	return overloads
}
