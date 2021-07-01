package typing

import (
	"strconv"
	"strings"
)

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

func (pt PrimType) Copy() DataType {
	return pt
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
		if arg.Variadic {
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

func (ft *FuncType) Copy() DataType {
	newArgs := make([]*FuncArg, len(ft.Args))
	for i, arg := range ft.Args {
		newArgs[i] = arg.copy()
	}

	return &FuncType{
		Args:          newArgs,
		ReturnType:    ft.ReturnType.Copy(),
		Async:         ft.Async,
		IntrinsicName: ft.IntrinsicName,
		Boxed:         ft.Boxed,
	}
}

// FuncArg represents an argument to a Chai function
type FuncArg struct {
	Name               string
	Type               DataType
	Optional, Variadic bool
	ByReference        bool
}

func (fa *FuncArg) equals(ofa *FuncArg) bool {
	return (fa.Name == "" || ofa.Name == "" || fa.Name == ofa.Name) &&
		Equals(fa.Type, ofa.Type) &&
		fa.Optional == ofa.Optional &&
		fa.Variadic == ofa.Variadic &&
		fa.ByReference == ofa.ByReference
}

func (fa *FuncArg) copy() *FuncArg {
	return &FuncArg{
		Name:        fa.Name,
		Type:        fa.Type.Copy(),
		Optional:    fa.Optional,
		Variadic:    fa.Variadic,
		ByReference: fa.ByReference,
	}
}

// -----------------------------------------------------------------------------

// VectorType represents a fixed-size array of similarly-typed elements
type VectorType struct {
	// ElemType is the type of the elements to the vector
	ElemType DataType

	// Size is either the fixed size of the vector or `-1` if the vector's
	// size is unknown at compile-time (ie. vector generic parameters)
	Size int

	// IsRow indicates whether or not this vector is a row vector
	IsRow bool
}

func (vt *VectorType) equals(other DataType) bool {
	if ovt, ok := other.(*VectorType); ok {
		return vt.ElemType.equals(ovt.ElemType) &&
			vt.IsRow == ovt.IsRow &&
			(vt.Size == ovt.Size || vt.Size == -1 || ovt.Size == -1)
	}

	return false
}

func (vt *VectorType) Repr() string {
	b := strings.Builder{}
	if vt.IsRow {
		b.WriteRune('R')
	}

	b.WriteString("Vec[")
	b.WriteString(vt.ElemType.Repr())
	b.WriteString("](")

	if vt.Size == -1 {
		b.WriteRune('_')
	} else {
		b.WriteString(strconv.Itoa(vt.Size))
	}

	b.WriteRune(')')
	return b.String()
}

func (vt *VectorType) Copy() DataType {
	return &VectorType{
		ElemType: vt.ElemType.Copy(),
		Size:     vt.Size,
		IsRow:    vt.IsRow,
	}
}

// -----------------------------------------------------------------------------

// // ConstraintSet is a polymorphic constraint set (ie. a user-defined
// // constraint). These are primarily used in generics to facilitate situations
// // where multiple type parameter values are legal, but there is no logical type
// // class grouping them. They are also used for numeric literals.
// type ConstraintSet struct {
// 	// Name is the name of the defined type constraint this poly cons set
// 	// corresponds to.  It can be empty if this is a "synthetic" poly cons set.
// 	Name string

// 	// SrcPackageID is the ID of the package this poly cons set was created in.
// 	SrcPackageID uint

// 	// Set is the list of possible types for this constraint set.  This cannot
// 	// contain type variables.
// 	Set []DataType
// }

// func (cs *ConstraintSet) equals(other DataType) bool {
// 	// because this function is only ever called at the top level --
// 	// there is no possibility for our poly cons set to be synthetic;
// 	// therefore, we can just do a name and package comparison
// 	if opcs, ok := other.(*ConstraintSet); ok {
// 		return cs.Name == opcs.Name && cs.SrcPackageID == opcs.SrcPackageID
// 	}

// 	return false
// }

// // contains is used as the equivalence test for poly cons sets in place of
// // `equals` which is only used for pure equality
// func (cs *ConstraintSet) contains(other DataType) bool {
// 	for _, item := range cs.Set {
// 		if Equivalent(other, item) {
// 			return true
// 		}
// 	}

// 	return false
// }

// func (cs *ConstraintSet) Repr() string {
// 	if cs.Name == "" {
// 		reprs := make([]string, len(cs.Set))
// 		for i, dt := range cs.Set {
// 			reprs[i] = dt.Repr()
// 		}

// 		return "(" + strings.Join(reprs, " | ") + ")"
// 	}

// 	return cs.Name
// }

// func (cs *ConstraintSet) Copy() DataType {
// 	newSet := make([]DataType, len(cs.Set))
// 	for i, item := range cs.Set {
// 		newSet[i] = item.Copy()
// 	}

// 	return &ConstraintSet{
// 		Name:         cs.Name,
// 		SrcPackageID: cs.SrcPackageID,
// 		Set:          newSet,
// 	}
// }
