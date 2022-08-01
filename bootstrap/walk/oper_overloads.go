package walk

import (
	"chaic/ast"
	"chaic/common"
	"chaic/report"
	"chaic/types"
	"chaic/util"
	"fmt"
)

// checkOperApp walks an operator application.  The operands should already be walked.
func (w *Walker) checkOperApp(op *common.AppliedOperator, span *report.TextSpan, operands ...ast.ASTExpr) types.Type {
	// Find the matching overloads.
	overloads := w.getOverloads(op, span, len(operands))

	// Type checking for operator applications appears as follows:
	//
	//   (+) := {overloads}
	//   r := {return type}
	//   for each operand:
	//      a_i := {any}
	//   assert (+) == (a...) -> r
	//   for each operand:
	//      assert operand == a_i
	//
	// This method gives better error messages for the argument types.

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

	// Create type variables corresponding to each of the operands.
	operandTypeVars := make([]types.Type, len(operands))
	for i, operand := range operands {
		operandTypeVars[i] = w.solver.NewTypeVar(fmt.Sprintf("operand%d", i), operand.Span())
	}

	// Create a new type to extract the operand types.
	operandFuncType := &types.FuncType{
		ParamTypes: operandTypeVars,
		ReturnType: returnType,
	}

	// Unify it with the operator overloads.
	w.solver.MustEqual(opVar, operandFuncType, span)

	// TODO: simplify the traceback in this case

	// Match the operand types to the extracted operand types of the overloads.
	for i, operand := range operands {
		w.solver.MustEqual(operand.Type(), operandTypeVars[i], operand.Span())
	}

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
