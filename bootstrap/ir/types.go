package ir

import "strings"

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
)

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
	default: // PrimKindBool
		return "bool"
	}
}

func (pt PrimType) Size() uint {
	switch pt {
	case PrimKindU8, PrimKindBool, PrimKindI8:
		return 1
	case PrimKindI16, PrimKindU16:
		return 2
	case PrimKindU32, PrimKindI32, PrimKindF32:
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

func (pt *PointerType) Repr() string {
	return "ptr[" + pt.ElemType.Repr() + "]"
}

func (pt *PointerType) Size() uint {
	// we are only targeting amd64 => ptr size = 8
	return 8
}

func (pt *PointerType) Align() uint {
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
		// field is inserted at an offset that is a multiple of its alignment
		if alignMod := field.Align() % offset; alignMod != 0 {
			offset += field.Align() - alignMod
		}

		sfields = append(sfields, StructField{Typ: field, Offset: offset})

		// increment the offset by the size of the type inserted
		offset += field.Size()
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
