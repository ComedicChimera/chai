package codegen

import (
	"chaic/mir"
	"chaic/types"
	"chaic/util"

	"github.com/llir/llvm/ir"
	"github.com/llir/llvm/ir/enum"
	lltypes "github.com/llir/llvm/ir/types"
)

// generateFunction generates a function definition.
func (g *Generator) generateFunction(mfunc *mir.Function) {
	// Whether to apply standard name mangling
	mangle := true

	// Whether this function definition should be marked external.
	public := mfunc.Public

	// Both @extern and @abientry stop name mangling.
	if mfunc.HasAttr(mir.AttrKindExternal) {
		mangle = false
	}

	// Determine the LLVM name based on the Chai name.
	var llName string
	if mangle {
		llName = g.pkgPrefix + mfunc.Name
	} else {
		llName = mfunc.Name
	}

	// Create the LLVM function.
	llParams := make([]*ir.Param, len(mfunc.ParamVars))
	for i, param := range mfunc.ParamVars {
		llParams[i] = ir.NewParam(param.Symbol.Name, g.convType(param.Symbol.Type))
		param.Symbol.LLValue = llParams[i]
	}

	// Get the function's return type.
	returnType := mfunc.Symbol.Type.(*types.FuncType).ReturnType

	// Add the pointer return parameter as necessary.
	var retParam *ir.Param = nil
	if types.IsPtrWrappedType(returnType) {
		ptrRetType := g.convType(returnType)
		retParam = ir.NewParam("return", ptrRetType)
		retParam.Attrs = append(retParam.Attrs, ir.SRet{Typ: ptrRetType.(*lltypes.PointerType).ElemType})
		llParams = append([]*ir.Param{retParam}, llParams...)
	}

	llFunc := g.mod.NewFunc(llName, g.convReturnType(returnType), llParams...)

	// Set the function's linkage based on its visibility.
	if public {
		llFunc.Linkage = enum.LinkageExternal
	} else {
		llFunc.Linkage = enum.LinkageInternal
	}

	// Add the function body as a predicate to generate if it exists.
	if len(mfunc.Body) > 0 {
		g.bodyPredicates = append(g.bodyPredicates, bodyPredicate{
			LLFunc:   llFunc,
			Params:   util.Map(mfunc.ParamVars, func(ident *mir.Identifier) *mir.MSymbol { return ident.Symbol }),
			RetParam: retParam,
			Body:     mfunc.Body,
		})
	}

	mfunc.Symbol.LLValue = llFunc
}

// generateStruct generates a struct definition.
func (g *Generator) generateStruct(mstruct *mir.Struct) {
	// Convert the field types skipping any unit types.
	fieldTypes := make([]lltypes.Type, len(mstruct.Type.Fields))
	n := 0
	for _, field := range mstruct.Type.Fields {
		if !types.IsUnit(field.Type) {
			fieldTypes[n] = g.convType(field.Type)
			n++
		}
	}
	fieldTypes = fieldTypes[:n]

	// Create the LLVM struct type.
	llvmStruct := lltypes.NewStruct(fieldTypes...)

	// Add the appropriate type definition.
	g.mod.NewTypeDef(g.pkgPrefix+mstruct.Type.Name(), llvmStruct)

	// Set the LLVM type of the struct type.
	mstruct.Type.SetLLType(llvmStruct)
}
