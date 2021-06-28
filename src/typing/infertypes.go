package typing

import (
	"chai/logging"
	"strings"
)

// TypeVariable is a special data type that represents an unknown type to be
// determined by the solver.  It is somewhat of a placeholder, but it does
// eventually have a known value.
type TypeVariable struct {
	// s is the parent solver of this type variable
	s *Solver

	// ID is a unique value identifying this type variable
	ID int

	// EvalType is the type that this type variable has been evaluated to. This
	// field is `nil` until the type variable is determined
	EvalType DataType

	// DefaultType is the type this type variable will be substituted with if no
	// other constraints are placed on it.  This field can be `nil` indicating
	// that there is no default type.  This is primarily used for numeric
	// literals when no more specific type for them can be determined
	DefaultType DataType

	// EvalFailed indicates that this type variable failed to be deduced
	EvalFailed bool

	// HandleUndetermined is called whenever a type value cannot be determined
	// for this type variable, but there was no other type error.
	HandleUndetermined func()
}

func (tv *TypeVariable) equals(other DataType) bool {
	logging.LogFatal("`equals` called directly on unevaluated type variable")
	return false
}

func (tv *TypeVariable) Repr() string {
	if tv.EvalType != nil {
		return tv.EvalType.Repr()
	} else if sub, ok := tv.s.substitutions[tv.ID]; ok {
		if sub.equivTo != nil {
			return sub.equivTo.Repr()
		}

		b := strings.Builder{}
		b.WriteRune('{')

		if sub.superTypeOf != nil {
			b.WriteString(sub.superTypeOf.Repr() + " < ")
		}

		b.WriteRune('_')

		if sub.subTypeOf != nil {
			b.WriteString(" < " + sub.subTypeOf.Repr())
		}

		b.WriteRune('}')
		return b.String()
	}

	return "_"
}

func (tv *TypeVariable) Copy() DataType {
	logging.LogFatal("Copy called on a type variable")
	return nil
}

// -----------------------------------------------------------------------------

// TypeConstraint represents a single Hindley-Milner type constraint: a
// statement of some relation one type has to another.
type TypeConstraint struct {
	Lhs, Rhs DataType

	// kind must be one of the enumerated constraint kinds
	Kind int

	// Pos is the position of the element(s) that generated this constraint
	Pos *logging.TextPosition
}

// Enumeration of constraint kind
const (
	TCEquiv   = iota // LHS and RHS are equivalent
	TCSubType        // RHS is a subtype of LHS
)

// TypeSubstitution is a set of conjectures the solver is making about a type to
// infer for a given type variable.  It is essentially a representation of the
// solver's current knowledge of that variable.
type TypeSubstitution struct {
	// equivTo is the value that this substitution must be exactly equivalent to
	equivTo DataType

	// subTypeOf is the type that this substitution is a sub type of (upper bound)
	subTypeOf DataType

	// superTypeOf is the type that this substitution is a super type of (lower bound)
	superTypeOf DataType
}
