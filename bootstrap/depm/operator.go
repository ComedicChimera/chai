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

// -----------------------------------------------------------------------------

// CheckOperatorCollisions checks for colliding definitions of operators.
func CheckOperatorCollisions(pkg *ChaiPackage) {
	// TODO: handle universal operator collisions

	// check globally defined operator collisions
	for _, op := range pkg.OperatorTable {
		for _, overload := range op.Overloads {
			// no need to look through import collisions: imports collide with
			// globals not the other way around
			if o2 := searchForCollisions(overload, op.Overloads); o2 != nil {
				report.ReportCompileError(
					overload.Context,
					overload.Position,
					fmt.Sprintf("conflicting definitions for `%s`: `%s` v `%s`", op.OpName, overload.Signature.Repr(), o2.Signature.Repr()),
				)
			}
		}
	}

	// check for imported operator conflicts
	for _, file := range pkg.Files {
		for opKind, op := range file.ImportedOperators {
			for _, overload := range op.Overloads {
				// check for conflicts between imported and global operators
				if globalOperator, ok := pkg.OperatorTable[opKind]; ok {
					if o2 := searchForCollisions(overload, globalOperator.Overloads); o2 != nil {
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
				if o2 := searchForCollisions(overload, op.Overloads); o2 != nil {
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

// searchForCollisions searches a list of overloads for collisions with a given
// overload.  The overload can be contained within the overload list.  It
// returns the colliding overload if one is found.
func searchForCollisions(overload *OperatorOverload, overloadList []*OperatorOverload) *OperatorOverload {
	for _, iOverload := range overloadList {
		// check if the overloads point to the same definition before checking
		// collision: an overload can't collide with itself
		if overload != iOverload && opOverloadCollides(overload, iOverload) {
			return iOverload
		}
	}

	return nil
}

// opOverloadCollides checks if two overloads collide.
func opOverloadCollides(overloadA, overloadB *OperatorOverload) bool {
	// NOTE: A standard equivalency test won't work since two operators collide
	// if and only if they have equivalent parameters: even if they have
	// different return types.
	if len(overloadA.Signature.Args) == len(overloadB.Signature.Args) {
		for i, arg := range overloadA.Signature.Args {
			if !typing.Equiv(arg, overloadB.Signature.Args[i]) {
				return false
			}
		}

		return true
	}

	return false
}
