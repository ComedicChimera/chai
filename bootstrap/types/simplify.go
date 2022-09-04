package types

import "chaic/util"

// Simplify returns the simplified representation of a type: it returns a type that is
// semantically equivalent to the original with all unnecessary wrapping (eg. type variables)
// removed.  This is roughly equivalent to recursively calling InnerType.
func Simplify(typ Type) Type {
	switch v := InnerType(typ).(type) {
	case *PointerType:
		return &PointerType{
			ElemType: Simplify(v.ElemType),
			Const:    v.Const,
		}
	case *TupleType:
		return &TupleType{
			ElementTypes: util.Map(v.ElementTypes, Simplify),
			size:         v.size,
			align:        v.align,
		}
	case *FuncType:
		return &FuncType{
			ParamTypes: util.Map(v.ParamTypes, Simplify),
			ReturnType: Simplify(v.ReturnType),
		}
	default:
		return v
	}
}
