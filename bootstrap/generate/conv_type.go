package generate

import (
	"chai/typing"
	"log"

	"github.com/llir/llvm/ir/types"
)

func (g *Generator) convType(typ typing.DataType) types.Type {
	typ = typing.InnerType(typ)
	llTyp := g.pureConvType(typ)

	// handle pointer types (like structs)
	if llTyp != nil && isPtrType(typ) {
		return types.NewPointer(llTyp)
	}

	return llTyp
}

func (g *Generator) convTypeNoPtr(typ typing.DataType) types.Type {
	return g.pureConvType(typing.InnerType(typ))
}

func (g *Generator) pureConvType(typ typing.DataType) types.Type {
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
		return g.globalTypes[v.Name()]
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

// -----------------------------------------------------------------------------

// isPtrType returns true if a type is wrapped in a pointer.
func isPtrType(typ typing.DataType) bool {
	typ = typing.InnerType(typ)

	switch typ.(type) {
	case *typing.StructType:
		return true
	}

	return false
}
