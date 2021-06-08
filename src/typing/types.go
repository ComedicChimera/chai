package typing

// PrimitiveType represents a primitive Chai type such as an `i32` or a `string`.
// Its value must be one of the enumerated primitive kinds below
type PrimitiveType uint

// Enumeration of primitive types
const (
	PrimKindU8 = iota
	PrimKindU16
	PrimKindU32
	PrimKindU64
	PrimKindI8
	PrimKindI16
	PrimKindI32
	PrimKindI64
	PrimKindF32
	PrimKindF64
	PrimKindBool
	PrimKindRune
	PrimKindString
	PrimKindAny
	PrimKindNothing
)

// equals for integers is an integer comparison
func (pt PrimitiveType) equals(other DataType) bool {
	if opt, ok := other.(PrimitiveType); ok {
		return pt == opt
	}

	return false
}

// Repr of a primitive type is just its corresponding token value
func (pt PrimitiveType) Repr() string {
	switch pt {
	case PrimKindU8:
		return "u8"
	case PrimKindU16:
		return "u16"
	case PrimKindU32:
		return "u32"
	case PrimKindU64:
		return "u64"
	case PrimKindI8:
		return "i8"
	case PrimKindI16:
		return "i16"
	case PrimKindI32:
		return "i32"
	case PrimKindI64:
		return "i64"
	case PrimKindF32:
		return "f32"
	case PrimKindF64:
		return "f64"
	case PrimKindBool:
		return "bool"
	case PrimKindRune:
		return "rune"
	case PrimKindString:
		return "string"
	case PrimKindAny:
		return "any"
	default:
		return "nothing"
	}
}
