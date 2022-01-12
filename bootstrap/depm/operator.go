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
			// no need to look through import collisions: imports collide with
			// globals not the other way around
			for j, o2 := range op.Overloads {
				if i != j && operatorCollides(overload, o2) {
					report.ReportCompileError(
						overload.Context,
						overload.Position,
						fmt.Sprintf("conflicting definitions for `%s`: `%s` v `%s`", op.OpName, overload.Signature.Repr(), o2.Signature.Repr()),
					)
				}
			}
		}
	}

	// check for imported operator conflicts
	for _, file := range pkg.Files {
		for _, op := range file.ImportedOperators {
			for i, overload := range op.Overloads {
				// check for conflicts between imported and global operators
				for _, o2 := range op.Overloads {
					if operatorCollides(overload, o2) {
						report.ReportCompileError(
							overload.Context,
							overload.Position,
							fmt.Sprintf("imported operator definitions for `%s` conflict with global definitions: `%s` v `%s`",
								op.OpName,
								overload.Signature.Repr(),
								o2.Signature.Repr(),
							),
						)
					}
				}

				// check for conflicts between imported operators in the same file
				for j, o2 := range op.Overloads {
					if i != j && operatorCollides(overload, o2) {
						report.ReportCompileError(
							overload.Context,
							overload.Position,
							fmt.Sprintf("imported operator definitions for `%s` conflict with other imported definitions: `%s` v `%s`",
								op.OpName,
								overload.Signature.Repr(),
								o2.Signature.Repr(),
							),
						)
					}
				}
			}
		}
	}
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
