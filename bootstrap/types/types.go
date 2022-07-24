package types

import (
	"chaic/util"
	"strings"
)

// Type represents a Chai data type.
type Type interface {
	// Returns whether this type is equal to the other type. This does not
	// account for inner types/type unwrapping: it should only be called within
	// methods of type instances.
	equals(other Type) bool

	// Returns the size of this type in bytes.
	Size() int

	// Returns the alignment of this type in bytes.
	Align() int

	// Returns the representative string for this type.
	Repr() string
}

// -----------------------------------------------------------------------------

// PrimitiveType represents a primitive type.  This must be one of the enumerated
// primitive type values below.
type PrimitiveType int

// Enumeration of the different primitive types.  The values of all of the
// integral types correspond to the usable width of the type: eg. i32 has usable
// width of 31.
const (
	PrimTypeUnit = PrimitiveType(0)
	PrimTypeBool = PrimitiveType(1)
	PrimTypeI8   = PrimitiveType(7)
	PrimTypeU8   = PrimitiveType(8)
	PrimTypeI16  = PrimitiveType(15)
	PrimTypeU16  = PrimitiveType(16)
	PrimTypeI32  = PrimitiveType(31)
	PrimTypeU32  = PrimitiveType(32)
	PrimTypeI64  = PrimitiveType(63)
	PrimTypeU64  = PrimitiveType(64)
	PrimTypeF32  = PrimitiveType(2)
	PrimTypeF64  = PrimitiveType(3)
)

func (pt PrimitiveType) equals(other Type) bool {
	if opt, ok := other.(PrimitiveType); ok {
		return pt == opt
	}

	return false
}

func (pt PrimitiveType) Size() int {
	switch pt {
	case PrimTypeUnit, PrimTypeBool, PrimTypeI8, PrimTypeU8:
		return 1
	case PrimTypeI16, PrimTypeU16:
		return 2
	case PrimTypeI32, PrimTypeU32, PrimTypeF32:
		return 4
	default:
		return 8
	}
}

func (pt PrimitiveType) Align() int {
	return pt.Size()
}

func (pt PrimitiveType) Repr() string {
	switch pt {
	case PrimTypeUnit:
		return "unit"
	case PrimTypeBool:
		return "bool"
	case PrimTypeI8:
		return "i8"
	case PrimTypeU8:
		return "u8"
	case PrimTypeI16:
		return "i16"
	case PrimTypeU16:
		return "u16"
	case PrimTypeI32:
		return "i32"
	case PrimTypeU32:
		return "u32"
	case PrimTypeI64:
		return "i64"
	case PrimTypeU64:
		return "u64"
	case PrimTypeF32:
		return "f32"
	default:
		return "f64"
	}
}

// IsIntegral returns whether this primitive is an integral type.
func (pt PrimitiveType) IsIntegral() bool {
	return PrimTypeI8 <= pt && pt <= PrimTypeU64
}

// IsFloating returns whether this primitive type is a floating-point type.
func (pt PrimitiveType) IsFloating() bool {
	return pt == PrimTypeF32 || pt == PrimTypeF64
}

// -----------------------------------------------------------------------------

// PointerType represents a pointer type.
type PointerType struct {
	// The element (content) type of the pointer.
	ElemType Type

	// Whether the pointer points to an immutable value.
	Const bool
}

func (pt *PointerType) equals(other Type) bool {
	if opt, ok := other.(*PointerType); ok {
		return Equals(pt.ElemType, opt.ElemType)
	}

	return false
}

func (pt *PointerType) Size() int {
	return util.PointerSize
}

func (pt *PointerType) Align() int {
	return util.PointerSize
}

func (pt *PointerType) Repr() string {
	if pt.Const {
		return "*const " + pt.ElemType.Repr()
	} else {
		return "*" + pt.ElemType.Repr()
	}
}

// -----------------------------------------------------------------------------

// Represents a function type.
type FuncType struct {
	// The parameter types of the function.
	ParamTypes []Type

	// The return type of the function.
	ReturnType Type
}

func (ft *FuncType) equals(other Type) bool {
	if oft, ok := other.(*FuncType); ok {
		if len(ft.ParamTypes) != len(oft.ParamTypes) {
			return false
		}

		for i, paramtyp := range ft.ParamTypes {
			if !Equals(paramtyp, oft.ParamTypes[i]) {
				return false
			}
		}

		return Equals(ft.ReturnType, oft.ReturnType)
	}

	return false
}

func (ft *FuncType) Size() int {
	return util.PointerSize
}

func (ft *FuncType) Align() int {
	return util.PointerSize
}

func (ft *FuncType) Repr() string {
	sb := strings.Builder{}

	if len(ft.ParamTypes) != 1 {
		sb.WriteRune('(')

		for i, paramtyp := range ft.ParamTypes {
			if i != 0 {
				sb.WriteString(", ")
			}

			sb.WriteString(paramtyp.Repr())
		}

		sb.WriteRune(')')
	} else {
		sb.WriteString(ft.ParamTypes[0].Repr())
	}

	sb.WriteString(" -> ")
	sb.WriteString(ft.ReturnType.Repr())

	return sb.String()
}
