package generate

import (
	"chai/typing"
	"log"

	"github.com/llir/llvm/ir/types"
)

func (g *Generator) convType(typ typing.DataType) types.Type {
	typ = typing.InnerType(typ)

	switch v := typ.(type) {
	case typing.PrimType:
		return g.convPrimType(v)
	case *typing.RefType:
		return types.NewPointer(g.convType(v.ElemType))
	case *typing.FuncType:
		log.Fatalln("first class functions are not supported in Alpha Chai")
	case *typing.AliasType:
		return g.convType(v.Type)
	case *typing.StructType:
		{
			st := g.globalTypes[v.Name()]

			// handle structs containing only nothing fields
			if st == nil {
				return st
			}

			// struct types are always wrapped in pointers
			return types.NewPointer(st)
		}
	}

	log.Fatalln("type not implemented yet")
	return nil
}

func (g *Generator) convPrimType(pt typing.PrimType) types.Type {
	switch pt {
	case typing.PrimI8, typing.PrimU8:
		return types.I8
	case typing.PrimI16, typing.PrimU16:
		return types.I16
	case typing.PrimI32, typing.PrimU32:
		return types.I32
	case typing.PrimI64, typing.PrimU64:
		return types.I64
	case typing.PrimF32:
		return types.Float
	case typing.PrimF64:
		return types.Double
	case typing.PrimBool:
		return types.I1
	case typing.PrimNothing:
		return types.Void
	case typing.PrimString:
		return types.NewPointer(g.stringType)
	}

	// unreachable
	return nil
}

// isPtrValue returns whether or not the type is pointer-value type such as a
// struct which is commonly wrapped in a pointer so it can be more readily
// manipulated.  Such values need to have additional semantics and tags
// generated for them to uphold Chai's strict value semantics.
func isPtrValue(typ typing.DataType) bool {
	switch v := typ.(type) {
	case typing.PrimType:
		return v == typing.PrimString
	}

	return false
}
