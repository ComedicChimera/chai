package types

// Equals returns whether two types are equal.
func Equals(a, b Type) bool {
	return InnerType(a).equals(InnerType(b))
}

// InnerType returns the "inner" type of typ.  For most types, this is just an
// identity function; however, for types such as type variables which
// essentially just wrap other types, this method is useful.
func InnerType(typ Type) Type {
	switch v := typ.(type) {
	case *TypeVariable:
		if v.Value == nil {
			return v
		}

		return v.Value
	}

	return typ
}

// Nullable returns whether the given type is nullable.
func Nullable(typ Type) bool {
	switch InnerType(typ).(type) {
	case *FuncType:
		// TODO: make pointer types non-nullable
		return false
	default:
		return true
	}
}

// IsUnit returns whether the given type is a unit type.
func IsUnit(typ Type) bool {
	return Equals(typ, PrimTypeUnit)
}
