package depm

import (
	"chai/report"
	"chai/typing"
)

// Operator is a defined operator in a specific package.
type Operator struct {
	// OpName is the token name of the operaotr.
	OpName string

	// Overloads is a list of the different overloading definitions.
	Overloads []*OperatorOverload
}

// OperatorOverload is a single overloading definition for an operator.
type OperatorOverload struct {
	Signature *typing.FuncType

	// Context and Position refer to the actual operator token of the definition
	// of this operator.
	Context  *report.CompilationContext
	Position *report.TextPosition

	// Public indicates whether or not this specific overload is exported.
	Public bool
}
