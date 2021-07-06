package typing

// DataType is the interface for all data types in Chai.
type DataType interface {
	// Repr returns a string representing the data type
	Repr() string

	// Copy returns a fresh copy of a data type
	Copy() DataType

	// equals takes in another DataType returns if the two data types are equal.
	// This method should return exact/true equality with no considerations for
	// other types (such as type parameters).  It is meant to only be called
	// internally.
	equals(other DataType) bool
}

// -----------------------------------------------------------------------------

// Equals computes effective equality between two data types as opposed to true
// equality.  This handles things like type parameters and other such types that
// "act" like another type while actually being of a different type.  This
// function does refer to the internal `equals` method of the types passed in.
func Equals(a, b DataType) bool {
	return InnerType(a).equals(InnerType(b))
}

// Equivalent checks if two types are effectively/logically equivalent
func Equivalent(a, b DataType) bool {
	a = InnerType(a)
	b = InnerType(b)

	// handle alias-based equivalency
	if at, ok := a.(*AliasType); ok {
		return Equivalent(at.Type, b)
	} else if bt, ok := b.(*AliasType); ok {
		return Equivalent(a, bt.Type)
	}

	return a.equals(b)
}

// InnerType returns the type "stored" by another data type (such as the value
// of a type parameter).  This is used for quickly unwrapping types before
// comparison and analysis.
func InnerType(dt DataType) DataType {
	switch v := dt.(type) {
	case *TypeVariable:
		if v.EvalType == nil {
			return v
		}

		return v.EvalType
	default:
		return v
	}
}
