package ast

import (
	"chaic/common"
	"chaic/report"
)

// The value of an annotation.
type AnnotValue struct {
	Value    string
	NameSpan *report.TextSpan
	ValSpan  *report.TextSpan
}

// -----------------------------------------------------------------------------

// The AST node representing a function definition.
type FuncDef struct {
	ASTBase

	// The symbol corresponding to the function.
	Symbol *common.Symbol

	// The parameters to the function.
	Params []*common.Symbol

	// The body of the function.  This may be nil if the function has no body.
	Body ASTNode

	// The function's annotations.
	Annotations map[string]AnnotValue
}

// -----------------------------------------------------------------------------

// The AST node representing an operator definition.
type OperDef struct {
	ASTBase

	// The representative string for the operator being overloaded.
	OpRepr string

	// The operator overload defined by this definition.
	Overload *common.OperatorOverload

	// The parameters to the operator overload function.
	Params []*common.Symbol

	// The body of the operator overload function.  This may be nil if the
	// function has no body.
	Body ASTNode

	// The operator function's annotations.
	Annotations map[string]AnnotValue
}
