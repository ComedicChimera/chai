package typing

import (
	"chai/logging"
	"strings"
)

// MonoConsSet is a special data type that represents a monomorphic constraint
// set.  The key trait of this type is that while it has multiple possible
// values, it can only evaluate to one of them.  This is most commonly used with
// operators: a given operator type or operand type will have one of these
// constraint sets that it will eventually collapse to a single instance of.
// Notable, these types can also influence/be linked to other monomorphic
// constraint sets.
type MonoConsSet struct {
	// Name is the name of the operator that this constraint set corresponds to.
	// If this constraint set instead corresponds to an argument to an operator,
	// this field should be an empty string
	Name string

	// Set is the list of possible types for this constraint set.  This cannot
	// contain type variables.
	Set []DataType

	// Relations is the list of mono constraint sets that are related (depend
	// on/ influence) this constraint set
	Relations []*MonoConsSet
}

func (mcs *MonoConsSet) Repr() string {
	if mcs.Name == "" {
		reprs := make([]string, len(mcs.Set))
		for i, dt := range mcs.Set {
			reprs[i] = dt.Repr()
		}

		return "(" + strings.Join(reprs, " | ") + ")"
	}

	return mcs.Name
}

func (mcs *MonoConsSet) equals(other DataType) bool {
	logging.LogFatal("`equals` called directly on monomorphic constraint set")
	return false
}

// -----------------------------------------------------------------------------

// PolyConsSet is a polymorphic constraint set (ie. a user-defined constraint).
// These are primarily used in generics to facilitate situations where multiple
// type parameter values are legal, but there is no logical type class grouping
// them. They are also used for numeric literals.
type PolyConsSet struct {
	// Name is the name of the defined type constraint this poly cons set
	// corresponds to.  It can be empty if this is a "synthetic" poly cons set.
	Name string

	// SrcPackageID is the ID of the package this poly cons set was created in.
	SrcPackageID uint

	// Set is the list of possible types for this constraint set.  This cannot
	// contain type variables.
	Set []DataType
}

func (pcs *PolyConsSet) Repr() string {
	if pcs.Name == "" {
		reprs := make([]string, len(pcs.Set))
		for i, dt := range pcs.Set {
			reprs[i] = dt.Repr()
		}

		return "(" + strings.Join(reprs, " | ") + ")"
	}

	return pcs.Name
}

func (pcs *PolyConsSet) equals(other DataType) bool {
	// because this function is only ever called at the top level --
	// there is no possibility for our poly cons set to be synthetic;
	// therefore, we can just do a name and package comparison
	if opcs, ok := other.(*PolyConsSet); ok {
		return pcs.Name == opcs.Name && pcs.SrcPackageID == opcs.SrcPackageID
	}

	return false
}

// contains is used as the equivalence test for poly cons sets in place of
// `equals` which is only used for pure equality
func (pcs *PolyConsSet) contains(other DataType) bool {
	for _, item := range pcs.Set {
		if Equivalent(other, item) {
			return true
		}
	}

	return false
}
