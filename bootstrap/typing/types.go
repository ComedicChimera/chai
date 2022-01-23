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

	// IntrinsicName is a field that doesn't actually determine anything related
	// to the type but makes generation of target code easier: the generator can
	// quickly determine the intrinsic to generate if the function is intrinsic.
	// If this field is empty, the function (or operator), is not intrinsic.
	IntrinsicName string
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

// -----------------------------------------------------------------------------

// NOTE: For all named types, the `Name` field is always prefixed by the name
// of the package it is defined in.  However, it is not sufficient to compare
// two named types solely using this name: two different packages may have the
// smae name within one project.  The `ParentID` field should also be used.

// NamedType is the base for all named types in Chai.
type NamedType struct {
	Name     string
	ParentID uint64
}

func NewNamedType(pkgName, name string, parentID uint64) NamedType {
	return NamedType{Name: pkgName + "." + name, ParentID: parentID}
}

func (nt *NamedType) Repr() string {
	return nt.Name
}

// -----------------------------------------------------------------------------

// AliasType is used to represent a defined type alias.
type AliasType struct {
	NamedType

	Type DataType
}

func (at *AliasType) Equals(other DataType) bool {
	if oat, ok := other.(*AliasType); ok {
		return at.Name == oat.Name && at.ParentID == oat.ParentID
	}

	return false
}

func (at *AliasType) Equiv(other DataType) bool {
	return InnerType(at.Type).Equiv(InnerType(other))
}

// -----------------------------------------------------------------------------

// StructType represents a pure structure type.
type StructType struct {
	NamedType

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

func (st *StructType) Equals(other DataType) bool {
	if ost, ok := other.(*StructType); ok {
		return st.Name == ost.Name && st.ParentID == ost.ParentID
	}

	return false
}

func (st *StructType) Equiv(other DataType) bool {
	// equivalency for structures is the same as equality
	return st.Equals(other)
}
