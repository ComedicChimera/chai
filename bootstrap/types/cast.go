package types

// Cast returns whether it is possible to cast src to dest.  `dest` should
// not be untyped (but may be a type variable).
func Cast(src, dest Type) bool {
	src = InnerType(src)
	dest = InnerType(dest)

	switch v := dest.(type) {
	case PrimitiveType:
		return castPrimitiveType(src, v)
	case *PointerType:
		if _, ok := src.(*PointerType); ok {
			// TODO: remove this kind of casting once pointer bit casting is no
			// longer needed (can be moved in `core.unsafe`).
			return true
		}
	}

	// All other casts are invalid.
	return false
}

// castPrimitiveType performs a type cast involving a primitive type.
func castPrimitiveType(src Type, dpt PrimitiveType) bool {
	// All other casts can only occur between primitive types.
	spt, ok := src.(PrimitiveType)
	if !ok {
		return false
	}

	// Casts from a one type to itself always succeed.
	if spt.equals(dpt) {
		return true
	}

	if dpt.IsFloating() { // ... to float
		// int to float and float to float
		return spt.IsIntegral() || spt.IsFloating()
	} else if dpt.IsIntegral() { // ... to int
		// float to int and int to int
		if spt.IsFloating() || spt.IsIntegral() {
			return true
		}

		// bool to int
		return spt == PrimTypeBool
	} else if dpt == PrimTypeBool { // ... to bool
		// int to bool
		return spt.IsIntegral()
	}

	// No other primitive casts are legal.
	return false
}
