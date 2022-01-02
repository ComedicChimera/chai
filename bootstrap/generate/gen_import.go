package generate

import (
	"chai/depm"
	"chai/typing"
	"log"

	"github.com/llir/llvm/ir"
	"github.com/llir/llvm/ir/enum"
)

func (g *Generator) genSymbolImport(importPrefix string, sym *depm.Symbol) {
	switch sym.DefKind {
	case depm.DKFuncDef:
		{
			symFt := sym.Type.(*typing.FuncType)

			// build the base LLVM function
			var params []*ir.Param
			for _, arg := range symFt.Args {
				// prune nothing types from the arguments
				if typing.IsNothing(arg) {
					continue
				}

				params = append(params, ir.NewParam("", g.convType(arg)))
			}

			llFunc := g.mod.NewFunc(importPrefix+sym.Name, g.convType(symFt.ReturnType), params...)
			llFunc.Linkage = enum.LinkageExternal
			llFunc.FuncAttrs = append(llFunc.FuncAttrs, enum.FuncAttrNoUnwind)
		}
	case depm.DKValueDef:
		{
			glob := g.mod.NewGlobal(importPrefix+sym.Name, g.convType(sym.Type))
			glob.ExternallyInitialized = true
			glob.Linkage = enum.LinkageExternal
		}
	default:
		log.Fatalln("other kinds of imports not yet implemented")
	}
}
