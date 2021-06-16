package typing

// Coercion and Casting Rules
// --------------------------
// 1. A coercion is a type conversion that can never result in a loss or
// reintepretation of underlying data.  For example, `i32` won't coerce to `u32`
// because such a coercion would result in the data of `i32` being
// reinterpreted.  Coercions happen implicitly.
// 2. A cast is a type conversion that can cause data loss and/or a
// reinterpretation of the data.  For example, an `i32` must be cast to a a
// `f32` because data can be lost or reinterpreted.  Chai does not permit all
// casts, but rather only casts between data types that are logically related.
// For example, floats and ints are both numbers so casts between them make
// snese.  However, casting from an int to a string makes no sense in terms of
// the actual data manipulation involved.

func CoerceTo(src, dest DataType) bool {
	if Equivalent(src, dest) {
		return true
	}

	src = InnerType(src)
	dest = InnerType(dest)

	switch v := dest.(type) {
	case PrimType:
		return v.coerce(src)
	default:
		// no known coercion
		return false
	}
}

func CastTo(src, dest DataType) bool {
	if CoerceTo(src, dest) {
		return true
	}

	src = InnerType(src)
	dest = InnerType(dest)

	switch v := src.(type) {
	case PrimType:
		return v.cast(dest)
	default:
		// no known casts
		return false
	}
}

// -----------------------------------------------------------------------------

func (pt PrimType) coerce(from DataType) bool {
	// all types will coerce to `any`
	if pt == PrimKindAny {
		return true
	}

	// all valid coercions for primitive types involve other primitive types
	if opt, ok := from.(PrimType); ok {
		// ints can coerec to larger ints of the same signedness
		if pt < 4 && opt < 4 {
			return pt > opt
		} else if 3 < pt && pt < 8 && 3 < opt && opt < 8 {
			return pt > opt
		}

		// floats can be coerced to larger floats
		if opt == PrimKindF32 && pt == PrimKindF64 {
			return true
		}

		// runes will coerce to strings
		return opt == PrimKindRune && pt == PrimKindString
	}

	return false
}

func (pt PrimType) cast(to DataType) bool {
	// `any` can be cast to any other type
	if pt == PrimKindAny {
		return true
	}

	if opt, ok := to.(PrimType); ok {
		// all numbers can be cast between each other
		if pt < PrimKindBool && opt < PrimKindBool {
			return true
		}

		// rune to int and int to rune
		if pt < PrimKindF32 && opt == PrimKindRune || opt < PrimKindF32 && pt == PrimKindRune {
			return true
		}

		// bool to int and int to bool
		return pt < PrimKindF32 && opt == PrimKindBool || opt < PrimKindF32 && pt == PrimKindBool
	}

	return false
}
