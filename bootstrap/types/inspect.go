package types

import "chaic/util"

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

		return InnerType(v.Value)
	case *OpaqueType:
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

// IsPtrWrappedType returns whether or not the given type should be wrapped in a
// pointer when used as a value in the backend.
func IsPtrWrappedType(typ Type) bool {
	switch InnerType(typ).(type) {
	case *StructType:
		return typ.Size() <= 2*util.PointerSize
	}

	return false
}
