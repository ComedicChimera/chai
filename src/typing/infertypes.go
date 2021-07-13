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

	// Required indicates whether or not this type variable's value must be
	// determined in order to deduction to succeed.  This is used for
	// conditional type variables such as those used in generic operator
	// overloading
	Required bool

	// HandleUndetermined is called whenever a type value cannot be determined
	// for this type variable, but there was no other type error.
	HandleUndetermined func()
}

func (tv *TypeVariable) equals(other DataType) bool {
	if otv, ok := other.(*TypeVariable); ok {
		return tv.ID == otv.ID
	}

	return false
}

func (tv *TypeVariable) Repr() string {
	if tv.EvalType != nil {
		return tv.EvalType.Repr()
	} else if sub, ok := tv.s.getSubstitution(tv.ID); ok {
		if sub.equivTo != nil {
			return sub.equivTo.Repr()
		}

		b := strings.Builder{}
		b.WriteRune('{')

		if len(sub.lowerBounds) == 1 {
			b.WriteString(sub.lowerBounds[0].Repr() + " < ")
		} else if len(sub.lowerBounds) > 1 {
			b.WriteRune('(')
			for i, bound := range sub.lowerBounds {
				b.WriteString(bound.Repr())

				if i < len(sub.lowerBounds)-1 {
					b.WriteString(" | ")
				}
			}

			b.WriteString(") < ")
		}

		b.WriteRune('_')

		if sub.upperBound != nil {
			b.WriteString(" < " + sub.upperBound.Repr())
		}

		b.WriteRune('}')
		return b.String()
	} else if overload, ok := tv.s.getOverload(tv.ID); ok {
		b := strings.Builder{}
		b.WriteRune('{')

		for i, val := range overload.Values {
			b.WriteString(val.Repr())

			if i < len(overload.Values)-1 {
				b.WriteString(" | ")
			}
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

	// upperBound is the type that this substitution is a sub type of.  One
	// bound here is sufficient because sub typing is transitive: a < b and b <
	// c => a < c.  Therefore, any time we narrow the bounds on this field, we
	// know that the previous upper bound is still valid
	upperBound DataType

	// lowerBounds is the set of types that this substitution is a super type
	// of. We need to store multiple bounds here since sub typing is only
	// transitive upward.  For example, if we know a type is lower bounded by
	// `rune` and we see that it is also lower bounded by `i32`, then it can't
	// be either `rune` or `i32` but rather it must be a general type of both of
	// them: eg. `Showable`.  We can only infer that general type, however,
	// based on context: `Showable` is only one possible generalization
	lowerBounds []DataType
}

// TypeOverload represents the various overloads for a type variable
type TypeOverload struct {
	// Values are all the data types that the variable could be unified to: the
	// actual overloads themselves
	Values []DataType

	// Corresponds is the list of all the other type overloads that must be
	// reduced in tandem with this one -- the value is the ID of the type
	// variable influenced.  This is used to facilitate things like operator
	// overloading where different type variables generalize to related by not
	// identical values
	Corresponds []int
}

// copy creates a new copy of this type overload for an inner unification state
func (to *TypeOverload) copy() *TypeOverload {
	nto := &TypeOverload{
		Values:      make([]DataType, len(to.Values)),
		Corresponds: to.Corresponds,
	}

	copy(nto.Values, to.Values)
	return nto
}

// -----------------------------------------------------------------------------

// TypeAssertion is an assertion that makes a claim about the types the solver
// deduces.  These assertions are applied after types are deduced because they
// don't really communicate any usable information about the types during
// deduction.
type TypeAssertion struct {
	// AssertKind is one of the enumerated assertion kinds below
	AssertKind int

	// Operand is the type the assertion is applied to
	Operand DataType

	// Data is any additional data needed to check the assertion
	Data interface{}

	// Pos is the position used for error reporting if the assertion fails
	Pos *logging.TextPosition
}

// Enumeration of assertion kinds
const (
	AssertCast   = iota // Assert that the operand is castable to another type
	AssertNonRef        // Assert that the operand is not a reference
)
