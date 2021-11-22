package ir

import (
	"fmt"
	"strings"
)

// Type represents a type that can be used in IR.  The IR type system is a more
// simplified, "machine-like" version of Chai's type system.
type Type interface {
	// Repr returns the string representation of the IR type.
	Repr() string

	// Size returns the size of the type in bytes.
	Size() uint

	Align() uint
}

// -----------------------------------------------------------------------------

// PrimType represents an IR primitive type: a number, boolean, or pointer.  It
// must be one of the enumerated IR primitive types.
type PrimType int

// Enumeration of IR PrimTypes
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
)

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
	default: // PrimKindBool
		return "bool"
	}
}

func (pt PrimType) Size() uint {
	switch pt {
	case PrimU8, PrimBool, PrimI8:
		return 1
	case PrimI16, PrimU16:
		return 2
	case PrimU32, PrimI32, PrimF32:
		return 4
	default: // u64, f64, i64
		return 8
	}
}

func (pt PrimType) Align() uint {
	return pt.Size()
}

// -----------------------------------------------------------------------------

// PointerType is a pointer to a piece of memory of some type.
type PointerType struct {
	ElemType Type
}

func (pt PointerType) Repr() string {
	return "ptr[" + pt.ElemType.Repr() + "]"
}

func (pt PointerType) Size() uint {
	// we are only targeting amd64 => ptr size = 8
	return 8
}

func (pt PointerType) Align() uint {
	return pt.Size()
}

// -----------------------------------------------------------------------------

// StructType represents any type that it is made of multiple associated fields
// placed contiguously in memory.
type StructType struct {
	Fields      []StructField
	size, align uint
}

type StructField struct {
	Typ    Type
	Offset uint
}

// NewStruct creates a new struct based on the field types.
func NewStruct(fields []Type) *StructType {
	// the alignment of a struct is simply the largest alignment of their fields
	// since structs are stored contiguously in memory
	var maxAlign uint

	// offset is used to keep track of field offsets as they are generated
	var offset uint

	// create the fields
	var sfields []StructField
	for _, field := range fields {
		// to determine the offset of a field, we first need to ensure that the
		// field is inserted at an offset that is a multiple of its alignment.
		if alignMod := offset % field.Align(); alignMod != 0 {
			offset += field.Align() - alignMod
		}

		sfields = append(sfields, StructField{Typ: field, Offset: offset})

		// increment the offset by the size of the type inserted
		offset += field.Size()

		// update the maximum alignment
		if field.Align() > maxAlign {
			maxAlign = field.Align()
		}
	}

	// the size of a struct is: the offset of its last field + the size of that
	// last field + padding necessary to make the size of the struct a multiple
	// of its alignment
	size := sfields[len(sfields)-1].Offset + sfields[len(sfields)-1].Typ.Size()
	size += maxAlign - size%maxAlign

	return &StructType{
		Fields: sfields,
		align:  maxAlign,
		size:   size,
	}
}

func (st StructType) Repr() string {
	sb := strings.Builder{}
	sb.WriteString("struct[")

	for i, field := range st.Fields {
		sb.WriteString(field.Typ.Repr())

		if i < len(st.Fields)-1 {
			sb.WriteString(", ")
		}
	}

	sb.WriteString("]")
	return sb.String()
}

func (st StructType) Size() uint {
	return st.size
}

func (st StructType) Align() uint {
	return st.align
}

// -----------------------------------------------------------------------------

// ArrayType represents a contiguous block of memory of the same time.  It has a
// fixed length but is not able (or needed) to be used as a vector.  Note that
// this ArrayType is NOT the same as the high-level `Array` type although
// `Array` will occasionally be compiled as an ArrayType.  An example usage of
// such a type would a string literal.
type ArrayType struct {
	ElemType Type
	Len      uint
}

func (at *ArrayType) Repr() string {
	return fmt.Sprintf("[%s * %d]", at.ElemType.Repr(), at.Len)
}

func (at *ArrayType) Size() uint {
	return at.ElemType.Size() * at.Len
}

func (at *ArrayType) Align() uint {
	// arrays need only be aligned based on their element type alignment
	return at.ElemType.Align()
}
