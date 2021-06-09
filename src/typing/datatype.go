package typing

// DataType is the interface for all data types in Chai.
type DataType interface {
	// Repr returns a string representing the data type
	Repr() string

	// equals takes in another DataType returns if the two data types are equal.
	// This method should return exact/true equality with no considerations for
	// other types (such as type parameters).  It is meant to only be called
	// internally.
	equals(other DataType) bool

	// coerce checks whether the argument type can be coerced to this data type.
	// It should only check for coercion not equality.
	coerce(from DataType) bool

	// cast checks whether this data type can cast to the argument type. It
	// should only check for casting not coercion or equality.
	cast(to DataType) bool
}

// -----------------------------------------------------------------------------

// Equals computes effective equality between two data types as opposed to true
// equality.  This handles things like type parameters and other such types that
// "act" like another type while actually being of a different type.  This
// function does refer to the internal `equals` method of the types passed in.
func Equals(a, b DataType) bool {
	return InnerType(a).equals(InnerType(b))
}

// InnerType returns the type "stored" by another data type (such as the value
// of a type parameter).  This is used for quickly unwrapping types before
// comparison and analysis.
func InnerType(dt DataType) DataType {
	return dt
}
