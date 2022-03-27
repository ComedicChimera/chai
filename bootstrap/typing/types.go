package typing

import "strings"

// DataType is the parent interface for all types in Chai.
type DataType interface {
	// Repr returns a representative string of the type for purposes of error
	// reporting.
	Repr() string

	// equals and equiv are the internal, type-specific implementations of
	// Equals and Equiv.  They should NEVER be called directly except by
	// Equals and Equiv.  They do not handle special cases like comparisons
	// to aliases or wrapped types.
	equals(DataType) bool
	equiv(DataType) bool
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
	PrimNothing
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
	case PrimBool:
		return "bool"
	default:
		// PrimNothing
		return "nothing"
	}
}

func (pt PrimType) equals(other DataType) bool {
	if opt, ok := other.(PrimType); ok {
		return pt == opt
	}

	return false
}

func (pt PrimType) equiv(other DataType) bool {
	return Equals(pt, other)
}

// -----------------------------------------------------------------------------

// FuncType represents a function type.
type FuncType struct {
	Args       []DataType
	ReturnType DataType

	// IntrinsicName is a field that doesn't actually determine anything related
	// to the type but makes generation of target code easier: the generator can
	// quickly determine the intrinsic to generate if the function is intrinsic.
	// If this field is empty, the function (or operator), is not intrinsic.
	IntrinsicName string
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

func (ft *FuncType) equals(other DataType) bool {
	if oft, ok := other.(*FuncType); ok {
		if len(ft.Args) != len(oft.Args) {
			return false
		}

		for i, arg := range ft.Args {
			oarg := oft.Args[i]

			if !Equals(arg, oarg) {
				return false
			}
		}

		return Equals(ft.ReturnType, oft.ReturnType)
	}

	return false
}

func (ft *FuncType) equiv(other DataType) bool {
	if oft, ok := other.(*FuncType); ok {
		if len(ft.Args) != len(oft.Args) {
			return false
		}

		for i, arg := range ft.Args {
			oarg := oft.Args[i]

			if !Equiv(arg, oarg) {
				return false
			}
		}

		return Equiv(ft.ReturnType, oft.ReturnType)
	}

	return false
}

// -----------------------------------------------------------------------------

// TupleType represents a tuple type.
type TupleType []DataType

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

func (tt TupleType) equals(other DataType) bool {
	if ott, ok := other.(TupleType); ok {
		if len(tt) != len(ott) {
			return false
		}

		for i, item := range tt {
			if !Equals(item, ott[i]) {
				return false
			}
		}

		return true
	}

	return false
}

func (tt TupleType) equiv(other DataType) bool {
	if ott, ok := other.(TupleType); ok {
		if len(tt) != len(ott) {
			return false
		}

		for i, item := range tt {
			if !Equiv(item, ott[i]) {
				return false
			}
		}

		return true
	}

	return false
}

// -----------------------------------------------------------------------------

// RefType is a reference type.
type RefType struct {
	ElemType DataType
}

func (rt *RefType) Repr() string {
	return "&" + rt.ElemType.Repr()
}

func (rt *RefType) equals(other DataType) bool {
	if ort, ok := other.(*RefType); ok {
		return Equals(rt.ElemType, ort.ElemType)
	}

	return false
}

func (rt *RefType) equiv(other DataType) bool {
	if ort, ok := other.(*RefType); ok {
		return Equiv(rt.ElemType, ort.ElemType)
	}

	return false
}

// -----------------------------------------------------------------------------

// NOTE: For all named types, the `Name` field is always prefixed by the name
// of the package it is defined in.  However, it is not sufficient to compare
// two named types solely using this name: two different packages may have the
// smae name within one project.  The `ParentID` field should also be used.

// NamedType is an interface implemented by all named types in Chai. It can be
// implemented implicitly by embedding `namedType`.
type NamedType interface {
	Name() string
	Repr() string
	ParentID() uint64
}

// NamedTypeBase is the base for all named types in Chai.
type NamedTypeBase struct {
	pkgName  string
	name     string
	parentID uint64
}

func NewNamedTypeBase(pkgName, name string, parentID uint64) NamedTypeBase {
	return NamedTypeBase{pkgName: pkgName, name: name, parentID: parentID}
}

func (nt *NamedTypeBase) Name() string {
	return nt.name
}

func (nt *NamedTypeBase) Repr() string {
	return nt.pkgName + "." + nt.name
}

func (nt *NamedTypeBase) ParentID() uint64 {
	return nt.parentID
}

// equals for named types need only compare the name and the parent package ID
// since two named types which are not references to the same definition cannot
// be defined with the same name in the same package.
func (nt *NamedTypeBase) equals(other DataType) bool {
	if ont, ok := other.(NamedType); ok {
		return nt.name == ont.Repr() && nt.parentID == ont.ParentID()
	}

	return false
}

// equiv for most named types is simple equality.
func (nt *NamedTypeBase) equiv(other DataType) bool {
	// Since `equiv` is only called from `Equiv`, we know all necessary
	// unwrapping has already been carried out so we can just call `equals`
	// directly.
	return nt.equals(other)
}

// -----------------------------------------------------------------------------

// OpaqueType represents a reference to a named type.
type OpaqueType struct {
	Name string

	// TypePtr is a shared pointer to the internal data type of the opaque type.
	// This will be a pointer to nil (not a nil pointer) until it is resolved.
	TypePtr *DataType
}

// NOTE: For all of these methods, we want them to throw a nil-pointer error if
// the type is unresolved since these types should not be used or evaluated
// until they are resolved.

func (ot *OpaqueType) Repr() string {
	return (*ot.TypePtr).Repr()
}

func (ot *OpaqueType) equals(other DataType) bool {
	return Equals(*ot.TypePtr, other)
}

func (ot *OpaqueType) equiv(other DataType) bool {
	return Equiv(*ot.TypePtr, other)
}

// -----------------------------------------------------------------------------

// AliasType is used to represent a defined type alias.
type AliasType struct {
	NamedTypeBase

	Type DataType
}

func (at *AliasType) equiv(other DataType) bool {
	return Equiv(at.Type, other)
}

// -----------------------------------------------------------------------------

// StructType represents a pure structure type.
type StructType struct {
	NamedTypeBase

	// Fields enumerates the fields of the struct in order.
	Fields []StructField

	// FieldsByName is an auxilliary map used to look up structure fields by
	// name instead of by position.
	FieldsByName map[string]int
}

// StructField is a single field within a structure type.
type StructField struct {
	Name   string
	Type   DataType
	Public bool

	// Initialized indicates whether this field has a default initializer within
	// the struct definition.
	Initialized bool

	// TODO: field annotations?
}
