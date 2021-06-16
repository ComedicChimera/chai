package typing

import "strings"

// PrimType represents a primitive Chai type such as an `i32` or a `string`. Its
// value must be one of the enumerated primitive kinds below
type PrimType uint

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

// equals for primitives is an integer comparison
func (pt PrimType) equals(other DataType) bool {
	if opt, ok := other.(PrimType); ok {
		return pt == opt
	}

	return false
}

// Repr of a primitive type is just its corresponding token value
func (pt PrimType) Repr() string {
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

// -----------------------------------------------------------------------------

// FuncType represents a Chai function (or operator variant) type
type FuncType struct {
	Args       []*FuncArg
	ReturnType DataType
	Async      bool

	// compiler internal properties
	IntrinsicName string
	Boxed         bool
}

func (ft *FuncType) equals(other DataType) bool {
	if oft, ok := other.(*FuncType); ok {
		if len(ft.Args) != len(oft.Args) {
			return false
		}

		for i, arg := range ft.Args {
			if !arg.equals(oft.Args[i]) {
				return false
			}
		}

		// boxed and intrinsic are compiler internal properties, so they can be
		// used to test equality
		return ft.Async == oft.Async
	}

	return false
}

func (ft *FuncType) Repr() string {
	sb := strings.Builder{}

	if ft.Async {
		sb.WriteString("async(")
	} else {
		sb.WriteString("fn(")
	}

	for i, arg := range ft.Args {
		if arg.Indefinite {
			sb.WriteString("...")
		} else if arg.Optional {
			sb.WriteRune('~')
		} else if arg.ByReference {
			sb.WriteString("&:")
		}

		sb.WriteString(arg.Type.Repr())

		if i < len(ft.Args)-1 {
			sb.WriteString(", ")
		}
	}

	sb.WriteRune(')')
	sb.WriteString(ft.ReturnType.Repr())
	return sb.String()
}

// FuncArg represents an argument to a Chai function
type FuncArg struct {
	Name                 string
	Type                 DataType
	Volatile             bool
	Optional, Indefinite bool
	ByReference          bool
}

func (fa *FuncArg) equals(ofa *FuncArg) bool {
	return fa.Name == ofa.Name &&
		Equals(fa.Type, ofa.Type) &&
		fa.Volatile == ofa.Volatile &&
		fa.Optional == ofa.Optional &&
		fa.Indefinite == ofa.Indefinite &&
		fa.ByReference == ofa.ByReference
}
