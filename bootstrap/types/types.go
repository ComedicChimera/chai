package types

import (
	"chaic/report"
	"chaic/util"
	"strings"

	lltypes "github.com/llir/llvm/ir/types"
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
	case PrimTypeUnit:
		return 0
	case PrimTypeBool, PrimTypeI8, PrimTypeU8:
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

/* -------------------------------------------------------------------------- */

// NamedType is the base type for all named types: structs, enums, etc.
type NamedType struct {
	// The named type's full name: parent package path + type name.
	Name string

	// The package ID that the named type is defined in.
	ParentID uint64

	// The index of the declaration which creates this named type within its
	// parent file's list of definitions.  This effectively functions as a back
	// reference to the declaring AST without actually referencing the AST to
	// get around Go's import rules.
	DeclIndex int

	// The LLVM type reference associated with this named type.
	// TODO: figure out how to handle this across multiple packages...
	LLType lltypes.Type
}

func (nt *NamedType) equals(other Type) bool {
	if ont, ok := other.(*NamedType); ok {
		return nt.Name == ont.Name && nt.ParentID == ont.ParentID
	}

	return false
}

func (nt *NamedType) Repr() string {
	return nt.Name
}

func (nt *NamedType) Size() int {
	report.ReportICE("Size() not overridden on NamedType")
	return 0
}

func (nt *NamedType) Align() int {
	report.ReportICE("Align() not overridden on NamedType")
	return 0
}

// OpaqueType represents a named type whose value has not yet been inferred.
type OpaqueType struct {
	NamedType

	// The span over which the named type reference occurs.
	Span *report.TextSpan

	// The resolved type of this opaque type.
	Value Type
}

func (ot *OpaqueType) Size() int {
	if ot.Value != nil {
		return ot.Value.Size()
	}

	report.ReportICE("called Size() on an unresolved opaque type")
	return 0
}

func (ot *OpaqueType) Align() int {
	if ot.Value != nil {
		return ot.Value.Align()
	}

	report.ReportICE("called Align() on an unresolved opaque type")
	return 0
}

/* -------------------------------------------------------------------------- */

// StructType represents a structure type.
type StructType struct {
	NamedType

	// The list of fields of the struct in order.
	Fields []StructField

	// A mapping between field names and their index within the struct.
	Indices map[string]int

	// The memoized struct size.
	size int

	// The memoized struct align.
	align int
}

// StructField represents a field of a structure type.
type StructField struct {
	// The field's name.
	Name string

	// The field's type.
	Type Type
}

func (st *StructType) Size() int {
	// Use the memoized size if possible.
	// Note: size could technically be zero if the struct is empty, but in that
	// case running this function again is trivial.
	if st.size != 0 {
		return st.size
	}

	size := 0

	// Calculate the size of struct such that all the fields are aligned.
	for _, field := range st.Fields {
		fieldAlign := field.Type.Align()

		if size%fieldAlign != 0 {
			size += fieldAlign - size%fieldAlign
		}

		size += field.Type.Size()
	}

	// TODO: do we need to pad the end of the struct for alignment?

	// Memoize the calculated size.
	st.size = size

	return size
}

func (st *StructType) Align() int {
	// Use the memoized alignment if possible.
	if st.align != 0 {
		return st.align
	}

	// The alignment of the struct is simply its maximum field alignment.
	maxAlign := 0

	for _, field := range st.Fields {
		fieldAlign := field.Type.Align()

		if fieldAlign > maxAlign {
			maxAlign = fieldAlign
		}
	}

	// Make sure we don't give it zero alignment.
	if maxAlign == 0 {
		maxAlign = 1
	}

	// Memoize the alignment.
	st.align = maxAlign

	return maxAlign
}

// GetFieldByName returns the struct field corresponding to the given name if it
// exists in the struct.
func (st *StructType) GetFieldByName(name string) (StructField, bool) {
	if index, ok := st.Indices[name]; ok {
		return st.Fields[index], true
	}

	return StructField{}, false
}
