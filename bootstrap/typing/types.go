package typing

// DataType is the parent interface for all types in Chai.
type DataType interface {
	// Equals returns if two types are exactly identical.
	Equals(DataType) bool

	// Equiv returns if two types are semantic equivalent: eg. an alias is
	// equivalent to the type that it is an alias of, but it is not equal to
	// that type.
	Equiv(DataType) bool

	// Repr returns a representative string of the type for purposes of error
	// reporting.
	Repr() string
}

// -----------------------------------------------------------------------------

// PrimType represents a primitive type.  It should be one of the enumerated
// primitive types.
type PrimType int

// Enumeration of different primitive types.
const (
	PrimU8 = iota
	PrimU16
	PrimU32
	PrimU64
	PrimI8
	PrimI16
	PrimI32
	PrimI64
	PrimF32
	PrimF64
	PrimBool
	PrimString
	PrimNothing
)

func (pt PrimType) Equals(other DataType) bool {
	if opt, ok := other.(PrimType); ok {
		return pt == opt
	}

	return false
}

func (pt PrimType) Equiv(other DataType) bool {
	return pt.Equals(other)
}

func (pt PrimType) Repr() string {
	switch pt {
	case PrimU8:
		return "u8"
	case PrimU16:
		return "u16"
	case PrimU32:
		return "u32"
	case PrimU64:
		return "u64"
	case PrimI8:
		return "i8"
	case PrimI16:
		return "i16"
	case PrimI32:
		return "i32"
	case PrimI64:
		return "i64"
	case PrimF32:
		return "f32"
	case PrimF64:
		return "f64"
	case PrimBool:
		return "bool"
	case PrimString:
		return "string"
	default:
		// PrimNothing
		return "nothing"
	}
}
