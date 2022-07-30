package common

import (
	"chaic/report"
	"chaic/types"

	llvalue "github.com/llir/llvm/ir/value"
)

// Operator represents a set of visible definitions for an operator of a given
// arity.
type Operator struct {
	// The token kind of the operator.
	Kind int

	// The representative string of the operator.
	OpRepr string

	// The arity of the overloads associated with this operator.
	Arity int

	// The overloads defined for this operator.
	Overloads []*OperatorOverload
}

// A counter used to generate new overload IDs.
var overloadIDCounter = 0

// OperatorOverload represents a single overload of a particular operator.
type OperatorOverload struct {
	// The unique ID of the overload.
	ID int

	// The ID of the parent package to this overload.
	ParentID uint64

	// The number identifying the file this overload is defined in.
	FileNumber int

	// The signature of the operator overload.
	Signature types.Type

	// The span containing the operator symbol.
	DefSpan *report.TextSpan

	// The instrinsic generator associated with this overload if any.
	IntrinsicName string

	// The LLVM value of the operator.
	LLValue llvalue.Value
}

// GetNewOverloadID gets a new unique ID for an operator overload.
func GetNewOverloadID() int {
	overloadIDCounter++
	return overloadIDCounter
}

// -----------------------------------------------------------------------------

// AppliedOperator represents a particular operator application.
type AppliedOperator struct {
	// The operator kind being applied.
	Kind int

	// The operator repr string associated with this operator application.
	OpRepr string

	// The span of the operator token.
	Span *report.TextSpan

	// The overload being used in the application.
	Overload *OperatorOverload
}
