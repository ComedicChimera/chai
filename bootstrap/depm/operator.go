package depm

import (
	"chai/report"
	"chai/typing"
	"fmt"
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

// CheckOperatorCollisions checks for colliding definitions of operators.
func CheckOperatorCollisions(pkg *ChaiPackage) {
	// check globally defined operator collisions
	for _, op := range pkg.OperatorTable {
		for i, overload := range op.Overloads {
		searchloop:
			for j, o2 := range op.Overloads {
				if i != j {
					if len(overload.Signature.Args) == len(o2.Signature.Args) {
						for k, arg := range overload.Signature.Args {
							if !arg.Equiv(o2.Signature.Args[k]) {
								continue searchloop
							}
						}

						report.ReportCompileError(
							overload.Context,
							overload.Position,
							fmt.Sprintf("conflicting definitions for `%s`: `%s` v `%s`", op.OpName, overload.Signature.Repr(), o2.Signature.Repr()),
						)
					}
				}
			}
		}
	}

	// TODO: check for imported operator conflicts
}
