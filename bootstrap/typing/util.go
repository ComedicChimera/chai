package typing

// NothingType returns a new `nothing` type.
func NothingType() PrimType {
	return PrimType(PrimNothing)
}

// IsNothing returns if the given data type is equivalent to `nothing`.
func IsNothing(dt DataType) bool {
	return Equiv(dt, NothingType())
}

// boolType returns a new `bool` type.
func BoolType() PrimType {
	return PrimType(PrimBool)
}
