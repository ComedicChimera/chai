package typing

import "chai/report"

// overloadSet is an interface representing a set of overloads for a particular
// type variable.  This is an inteface because several different kinds of
// overloads have extra semantics.
type overloadSet interface {
	// Prune creates a new overload set consisting only of the overloads which
	// can be unified with the given type.
	Prune(s *Solver, dt DataType, pos *report.TextPosition) overloadSet

	// Len returns the number of overloads in the overload set.
	Len() int

	// Repr returns the string representation of the type set represented by
	// this overload set.
	Repr() string

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
	Overloads []operatorOverload

	// ResultID is the field in which to store the resulting overload ID.
	ResultID *uint64
}

// operatorOverload represents a single operator overload in an operator
// overload set.
type operatorOverload struct {
	// PkgID is the ID of the package containing this overload definition.
	PkgID uint64

	// Signature is the function type of the operator overload.
	Signature *FuncType
}

func (oos *operatorOverloadSet) Prune(s *Solver, dt DataType, pos *report.TextPosition) overloadSet {
	var newOverloads []operatorOverload

	for _, overload := range oos.Overloads {
		if s.unify(overload.Signature, dt, pos) {
			newOverloads = append(newOverloads, overload)
		}
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

func (oos *operatorOverloadSet) Substitution() DataType {
	return oos.Overloads[0].Signature
}

func (oos *operatorOverloadSet) Finalize() DataType {
	if len(oos.Overloads) == 1 {
		*oos.ResultID = oos.Overloads[0].PkgID
	}

	return nil
}
