package typing

import "strings"

// DataType is the parent interface for all types in Chai.
type DataType interface {
	// Equals returns if two types are exactly identical.  This operation is
	// commutative.
	Equals(DataType) bool

	// Equiv returns if two types are semantic equivalent: eg. an alias is
	// equivalent to the type that it is an alias of, but it is not equal to
	// that type.  This operation is commutative.  Equivalency requires that the
	// two types compile to the same output type in LLVM.
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

// -----------------------------------------------------------------------------

// FuncType represents a function type.
type FuncType struct {
	Args       []DataType
	ReturnType DataType
}

func (ft *FuncType) Equals(other DataType) bool {
	if oft, ok := other.(*FuncType); ok {
		if len(ft.Args) != len(oft.Args) {
			return false
		}

		for i, arg := range ft.Args {
			oarg := oft.Args[i]

			if !arg.Equals(oarg) {
				return false
			}
		}

		return ft.ReturnType.Equals(oft.ReturnType)
	}

	return false
}

func (ft *FuncType) Equiv(other DataType) bool {
	if oft, ok := other.(*FuncType); ok {
		if len(ft.Args) != len(oft.Args) {
			return false
		}

		for i, arg := range ft.Args {
			oarg := oft.Args[i]

			if !arg.Equiv(oarg) {
				return false
			}
		}

		return ft.ReturnType.Equiv(oft.ReturnType)
	}

	return false
}

func (ft *FuncType) Repr() string {
	sb := strings.Builder{}

	sb.WriteRune('(')

	for i, arg := range ft.Args {
		sb.WriteString(arg.Repr())

		if i < len(ft.Args)-1 {
			sb.WriteString(", ")
		}
	}

	sb.WriteString(") -> ")
	sb.WriteString(ft.ReturnType.Repr())

	return sb.String()
}

// -----------------------------------------------------------------------------

// TupleType represents a tuple type.
type TupleType []DataType

func (tt TupleType) Equals(other DataType) bool {
	if ott, ok := other.(TupleType); ok {
		if len(tt) != len(ott) {
			return false
		}

		for i, item := range tt {
			if !item.Equals(ott[i]) {
				return false
			}
		}

		return true
	}

	return false
}

func (tt TupleType) Equiv(other DataType) bool {
	if ott, ok := other.(TupleType); ok {
		if len(tt) != len(ott) {
			return false
		}

		for i, item := range tt {
			if !item.Equiv(ott[i]) {
				return false
			}
		}

		return true
	}

	return false
}

func (tt TupleType) Repr() string {
	sb := strings.Builder{}
	sb.WriteRune('(')

	for i, typ := range tt {
		sb.WriteString(typ.Repr())

		if i < len(tt)-1 {
			sb.WriteString(", ")
		}
	}

	sb.WriteRune(')')
	return sb.String()
}

// -----------------------------------------------------------------------------

// RefType is a reference type.
type RefType struct {
	ElemType DataType
}

func (rt *RefType) Equals(other DataType) bool {
	if ort, ok := other.(*RefType); ok {
		return rt.ElemType.Equals(ort.ElemType)
	}

	return false
}

func (rt *RefType) Equiv(other DataType) bool {
	if ort, ok := other.(*RefType); ok {
		return rt.ElemType.Equiv(ort.ElemType)
	}

	return false
}

func (rt *RefType) Repr() string {
	return "&" + rt.ElemType.Repr()
}
