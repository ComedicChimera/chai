package lower

import (
	"chai/ir"
	"chai/typing"
)

// lowerType converts a top-level type into an IR type.
func (l *Lowerer) lowerType(typ typing.DataType) ir.Type {
	switch v := typ.(type) {
	case typing.PrimType:
		return l.lowerPrimType(v)
	case *typing.RefType:
		return &ir.PointerType{ElemType: l.lowerType(v.ElemType)}
	}

	// unreachable
	return nil
}

// lowerPrimType lowers a primitive type
func (l *Lowerer) lowerPrimType(pt typing.PrimType) ir.Type {
	switch pt {
	case typing.PrimString:
		return &ir.PointerType{ElemType: ir.NewStruct([]ir.Type{
			&ir.PointerType{ElemType: ir.PrimType(ir.PrimU8)},
			ir.PrimType(ir.PrimU32),
		})}
	case typing.PrimNothing:
		// nothings should be completely pruned and should not end up in the
		// resulting IR
		return nil
	default:
		// all of the integral and boolean types are identical in terms of value
		// between the AST and the IR
		return ir.PrimType(pt)
	}
}
