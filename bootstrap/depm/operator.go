package depm

import (
	"chai/report"
	"chai/typing"
	"fmt"
)

// Operator is a defined operator in a specific package.
type Operator struct {
	// Pkg is the package containing the operator.
	Pkg *ChaiPackage

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

// operatorCollides checks if a specific overload collides with another
// overload.
func operatorCollides(overloadA, overloadB *OperatorOverload) bool {
	if len(overloadA.Signature.Args) == len(overloadB.Signature.Args) {
		for i, arg := range overloadA.Signature.Args {
			if !arg.Equiv(overloadB.Signature.Args[i]) {
				return false
			}
		}

		return true
	}

	return false
}
