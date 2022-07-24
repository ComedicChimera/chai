package types

// Cast returns whether it is possible to cast src to dest.  `dest` should not
// be untyped (but may be a type variable).  It will also use casting to make
// type deductions: eg. `0 as i32` infers 0 as an i32.
func Cast(src, dest Type) bool {
	src = InnerType(src)
	dest = InnerType(dest)

	// If src is still an untyped null, then it has no inferred type and so its
	// type can freely become that of `dest`.
	if un, ok := src.(*UntypedNull); ok {
		un.InferredType = dest
		return true
	}

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
	// If the source type is still an untyped number, then it has no inferred
	// type, and we can check cast validity by just searching the valid types.
	if un, ok := src.(*UntypedNumber); ok {
		for _, vtype := range un.ValidTypes {
			if Equals(vtype, dpt) {
				un.InferredType = dpt
				return true
			}
		}

		return false
	}

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
