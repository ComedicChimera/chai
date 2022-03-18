package typing

import (
	"chai/report"
	"strings"
)

// overloadSet is an interface representing a set of overloads for a particular
// type variable.  This is an inteface because several different kinds of
// overloads have extra semantics.
type overloadSet interface {
	// Prune creates a new overload set consisting only of the overloads which
	// can be unified with the given type.
	Prune(s *Solver, cstate *solutionState, dt DataType, pos *report.TextPosition) overloadSet

	// Len returns the number of overloads in the overload set.
	Len() int

	// Repr returns the string representation of the type set represented by
	// this overload set.
	Repr() string

	// Copy returns a copy of this overload set.
	Copy() overloadSet

	// Substitution returns the type to substitute after resolving this
	// overload.  This should only be called after all but one overload has been
	// effectively pruned.  This can be called after each round of overload
	// reduction to get a resulting substitution.
	Substitution() DataType

	// Finalize should be called once all constraints which could affect this
	// overload set have been applied.  It performs all "closing" logic for the
	// overload set.  It can also return a type to substitute in the case that
	// the overload set is able to "self-prune" based on its current state. Note
	// that this function returning a substitution is a sufficient but not
	// necessary condition for determining a type: a pre-existing substitution
	// is also satisfactory.
	Finalize() DataType
}

// -----------------------------------------------------------------------------

// operatorOverloadSet is an overload set created from an operator overload.
type operatorOverloadSet struct {
	// OperName is the name/symbol of the operator (eg. `+`).
	OperName string

	// Overloads is the slice of individual overloads in this overload set.
	Overloads []OperatorOverload

	// ResultID is the field in which to store the resulting overload ID.
	ResultID *uint64
}

// OperatorOverload represents a single operator overload in an operator
// overload set.
type OperatorOverload struct {
	// PkgID is the ID of the package containing this overload definition.
	PkgID uint64

	// Signature is the function type of the operator overload.
	Signature *FuncType
}

func (oos *operatorOverloadSet) Prune(s *Solver, cstate *solutionState, dt DataType, pos *report.TextPosition) overloadSet {
	var newOverloads []OperatorOverload

	for _, overload := range oos.Overloads {
		if s.unify(overload.Signature, dt, pos) {
			newOverloads = append(newOverloads, overload)
		}

		// reset state for next test unification; since we always reset this at
		// the end of the loop, we know that we don't need to restore it at the
		// end of this function.
		s.localState = cstate.copyState()
	}

	return &operatorOverloadSet{
		OperName:  oos.OperName,
		Overloads: newOverloads,
		ResultID:  oos.ResultID,
	}
}

func (oos *operatorOverloadSet) Len() int {
	return len(oos.Overloads)
}

func (oos *operatorOverloadSet) Repr() string {
	return oos.OperName
}

func (oos *operatorOverloadSet) Copy() overloadSet {
	newOverloads := make([]OperatorOverload, len(oos.Overloads))
	copy(newOverloads, oos.Overloads)

	return &operatorOverloadSet{
		OperName:  oos.OperName,
		Overloads: newOverloads,
		ResultID:  oos.ResultID,
	}
}

func (oos *operatorOverloadSet) Substitution() DataType {
	return oos.Overloads[0].Signature
}

func (oos *operatorOverloadSet) Finalize() DataType {
	if len(oos.Overloads) == 1 {
		*oos.ResultID = oos.Overloads[0].PkgID
	}

	return nil
}

// -----------------------------------------------------------------------------

// literalOverloadSet is an overload set for a literal.
type literalOverloadSet struct {
	LitTypeName string
	Overloads   []LiteralOverload
}

// LiteralOverload is a single type overload for a literal.
type LiteralOverload struct {
	Name string
	Type DataType
}

func (los *literalOverloadSet) Prune(s *Solver, cstate *solutionState, dt DataType, pos *report.TextPosition) overloadSet {
	var newOverloads []LiteralOverload

	for _, overload := range los.Overloads {
		if s.unify(overload.Type, dt, pos) {
			newOverloads = append(newOverloads, overload)
		}

		// reset state for next test unification; since we always reset this at
		// the end of the loop, we know that we don't need to restore it at the
		// end of this function.
		s.localState = cstate.copyState()
	}

	newName := los.LitTypeName
	if len(newOverloads) != len(los.Overloads) {
		useBuilder := false
		namesBuilder := strings.Builder{}

		namesBuilder.WriteRune('{')
		for i, overload := range newOverloads {
			if newName == los.LitTypeName {
				newName = overload.Name
			} else if overload.Name != newName {
				useBuilder = true
			}

			namesBuilder.WriteString(overload.Name)

			if i < len(newOverloads)-1 {
				namesBuilder.WriteString(" | ")
			}
		}
		namesBuilder.WriteRune('}')

		if useBuilder {
			newName = namesBuilder.String()
		}
	}

	return &literalOverloadSet{
		LitTypeName: newName,
		Overloads:   newOverloads,
	}
}

func (los *literalOverloadSet) Len() int {
	return len(los.Overloads)
}

func (los *literalOverloadSet) Repr() string {
	return los.LitTypeName
}

func (los *literalOverloadSet) Copy() overloadSet {
	newOverloads := make([]LiteralOverload, len(los.Overloads))
	copy(newOverloads, los.Overloads)

	return &literalOverloadSet{
		LitTypeName: los.LitTypeName,
		Overloads:   newOverloads,
	}
}

func (los *literalOverloadSet) Substitution() DataType {
	return los.Overloads[0].Type
}

func (los *literalOverloadSet) Finalize() DataType {
	if len(los.Overloads) > 0 {
		return los.Overloads[0].Type
	}

	return nil
}
