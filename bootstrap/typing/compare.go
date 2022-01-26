package typing

// Equals returns if two types are exactly identical.  This operation is
// commutative.
func Equals(lhs, rhs DataType) bool {
	return InnerType(lhs).equals(InnerType(rhs))
}

// Equiv returns if two types are semantic equivalent: eg. an alias is
// equivalent to the type that it is an alias of, but it is not equal to that
// type.  This operation is commutative.  Equivalency requires that the two
// types compile to the same output type in LLVM.
func Equiv(lhs, rhs DataType) bool {
	// check for aliases on the RHS
	rhInner := InnerType(rhs)
	if rhAlias, ok := rhInner.(*AliasType); ok {
		return InnerType(lhs).equiv(rhAlias.Type)
	}

	return InnerType(lhs).equiv(rhInner)
}

// -----------------------------------------------------------------------------

// InnerType extracts the inner type value of a data type. This value can then
// be used for things like cast checking.  For example, it extracts the value
// for a type variable.
func InnerType(dt DataType) DataType {
	switch v := dt.(type) {
	case *TypeVar:
		return InnerType(v.Value)
	case *OpaqueType:
		return InnerType(*v.TypePtr)
	default:
		// no inner type to extract
		return v
	}
}
