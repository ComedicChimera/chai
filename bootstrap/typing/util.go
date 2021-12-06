package typing

// InnerType extracts the inner type value of a data type. This value can then
// be used for things like cast checking.  For example, it extracts the value
// for a type variable.
func InnerType(dt DataType) DataType {
	switch v := dt.(type) {
	case *TypeVar:
		return InnerType(v.Value)
	default:
		// no inner type to extract
		return v
	}
}

// -----------------------------------------------------------------------------

// NothingType returns a new `nothing` type.
func NothingType() PrimType {
	return PrimType(PrimNothing)
}

// IsNothing returns if the given data type is equivalent to `nothing`.
func IsNothing(dt DataType) bool {
	return dt.Equiv(NothingType())
}

// boolType returns a new `bool` type.
func BoolType() PrimType {
	return PrimType(PrimBool)
}
