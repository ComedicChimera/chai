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

// TupleType represents a tuple type.
type TupleType struct {
	// The element types of the tuple.
	ElementTypes []Type

	// The memoized size and alignment.
	size  int
	align int
}

func (tt *TupleType) equals(other Type) bool {
	if ott, ok := other.(*TupleType); ok {
		if len(tt.ElementTypes) == len(ott.ElementTypes) {
			for i, elemType := range tt.ElementTypes {
				if !Equals(elemType, ott.ElementTypes[i]) {
					return false
				}
			}

			return true
		}
	}

	return false
}

func (tt *TupleType) Size() int {
	// Use the memoized size if possible.  Tuples can never be empty.
	if tt.size != 0 {
		return tt.size
	}

	size := 0

	// Calculate the size of struct such that all the fields are aligned.
	for _, elemType := range tt.ElementTypes {
		elemAlign := elemType.Align()

		if size%elemAlign != 0 {
			size += elemAlign - size%elemAlign
		}

		size += elemType.Size()
	}

	// TODO: do we need to pad the end of the struct for alignment?

	// Memoize the calculated size.
	tt.size = size

	return size
}

func (tt *TupleType) Align() int {
	// Use the memoized alignment if possible.
	if tt.align != 0 {
		return tt.align
	}

	// The alignment of the tuple is simply its maximum element alignment.
	maxAlign := 0

	for _, elemType := range tt.ElementTypes {
		elemAlign := elemType.Align()

		if elemAlign > maxAlign {
			maxAlign = elemAlign
		}
	}

	// Make sure we don't give it zero alignment.
	if maxAlign == 0 {
		maxAlign = 1
	}

	// Memoize the alignment.
	tt.align = maxAlign

	return maxAlign
}

func (tt *TupleType) Repr() string {
	sb := strings.Builder{}

	sb.WriteRune('(')

	for i, elemType := range tt.ElementTypes {
		if i > 0 {
			sb.WriteString(", ")
		}

		sb.WriteString(elemType.Repr())
	}

	sb.WriteRune(')')

	return sb.String()
}

/* -------------------------------------------------------------------------- */

// ArrayType represents an array type.
type ArrayType struct {
	// The element type of the array.
	ElemType Type

	// Whether the array is a view.
	Const bool
}

func (at *ArrayType) equals(other Type) bool {
	if oat, ok := other.(*ArrayType); ok {
		return Equals(at.ElemType, oat.ElemType) && at.Const == oat.Const
	}

	return false
}

func (at *ArrayType) Size() int {
	return 2 * util.PointerSize
}

func (at *ArrayType) Align() int {
	return util.PointerSize
}

func (at *ArrayType) Repr() string {
	if at.Const {
		return "[]const " + at.ElemType.Repr()
	} else {
		return "[]" + at.ElemType.Repr()
	}
}

/* -------------------------------------------------------------------------- */

// NamedType represents a user-defined type associated with a symbol.
type NamedType interface {
	Type

	// The named type's name.
	Name() string

	// The package ID that the named type is defined in.
	ParentID() uint64

	// The index of the declaration which creates this named type within its
	// parent file's list of definitions.  This effectively functions as a back
	// reference to the declaring AST without actually referencing the AST to
	// get around Go's import rules.
	DeclIndex() int

	// Color returns the color associated with this named type.  See the
	// infinite type checker for more information.
	Color() Color

	// SetColor sets the color of the named type.
	SetColor(color Color)

	// LLType returns the LLVM type currently associated with this named type.
	LLType() lltypes.Type

	// SetLLType sets the LLVM type of the named type.
	SetLLType(llType lltypes.Type)
}

// Color is a special value associated with every named type.  Must be one of
// the enumerated color values.  See the infinite type checker for more
// information.
type Color byte

// Enumeration of possible named type colors.
const (
	ColorWhite Color = iota
	ColorGrey
	ColorBlack
)

/* -------------------------------------------------------------------------- */

// NamedTypeBase is the base type for all named types: structs, enums, etc.
type NamedTypeBase struct {
	// The named type's name.
	name string

	// The package ID that the named type is defined in.
	parentID uint64

	// The index of the declaration which creates this named type within its
	// parent file's list of definitions.  This effectively functions as a back
	// reference to the declaring AST without actually referencing the AST to
	// get around Go's import rules.
	declIndex int

	// The color associated with the named type.
	color Color

	// The LLVM type reference associated with this named type.
	// TODO: figure out how to handle this across multiple packages...
	llType lltypes.Type
}

// NewNamedTypeBase creates a new named type base.
func NewNamedTypeBase(name string, parentID uint64, declIndex int) NamedTypeBase {
	return NamedTypeBase{
		name:      name,
		parentID:  parentID,
		declIndex: declIndex,
		color:     ColorWhite,
		llType:    nil,
	}
}

func (nt *NamedTypeBase) equals(other Type) bool {
	if ont, ok := other.(NamedType); ok {
		return nt.name == ont.Name() && nt.parentID == ont.ParentID()
	}

	return false
}

func (nt *NamedTypeBase) Repr() string {
	return nt.name
}

func (nt *NamedTypeBase) Size() int {
	report.ReportICE("Size() not overridden on NamedType")
	return 0
}

func (nt *NamedTypeBase) Align() int {
	report.ReportICE("Align() not overridden on NamedType")
	return 0
}

func (nt *NamedTypeBase) Name() string {
	return nt.name
}

func (nt *NamedTypeBase) ParentID() uint64 {
	return nt.parentID
}

func (nt *NamedTypeBase) DeclIndex() int {
	return nt.declIndex
}

func (nt *NamedTypeBase) Color() Color {
	return nt.color
}

func (nt *NamedTypeBase) SetColor(color Color) {
	nt.color = color
}

func (nt *NamedTypeBase) LLType() lltypes.Type {
	return nt.llType
}

func (nt *NamedTypeBase) SetLLType(lltype lltypes.Type) {
	nt.llType = lltype
}

/* -------------------------------------------------------------------------- */

// OpaqueType represents a (possibly unresolved) reference to a named type.
type OpaqueType struct {
	// The name of the type referenced by the opaque type.
	Name string

	// The ID of the package containing the visible definition of the opaque
	// type. Note that this may not be the same as the type of the underlying
	// named type.
	ContainingID uint64

	// The span over which the named type reference occurs.
	Span *report.TextSpan

	// The resolved type of this opaque type.
	Value NamedType
}

func (ot *OpaqueType) equals(other Type) bool {
	report.ReportICE("called equals() on an unresolved opaque type")
	return false
}

func (ot *OpaqueType) Repr() string {
	return ot.Name
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
	NamedTypeBase

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
