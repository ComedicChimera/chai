package codegen

import (
	"chaic/ast"
	"chaic/common"
	"chaic/report"
	"chaic/types"
	"fmt"

	"github.com/llir/llvm/ir"
	"github.com/llir/llvm/ir/enum"
	lltypes "github.com/llir/llvm/ir/types"
)

// generateDef generates an LLVM IR definition.
func (g *Generator) generateDef(def ast.ASTNode) {
	switch v := def.(type) {
	case *ast.FuncDef:
		g.generateFuncDef(v)
	case *ast.OperDef:
		g.generateOperDef(v)
	case *ast.StructDef:
		g.generateStructDef(v)
	default:
		report.ReportICE("definition codegen not implemented")
	}
}

// generateFuncDef generates a function definition.
func (g *Generator) generateFuncDef(fd *ast.FuncDef) {
	// TODO intrinsic functions
	if hasAnnot(fd.Annotations, "intrinsic") {
		return
	}

	// Whether to apply standard name mangling
	mangle := true

	// Whether this function definition should be marked external.
	public := false

	// Both @extern and @abientry stop name mangling and make the symbol public.
	if hasAnnot(fd.Annotations, "extern") || hasAnnot(fd.Annotations, "abientry") {
		mangle = false
		public = true
	}

	// Determine the LLVM name based on the Chai name.
	var llName string
	if mangle {
		llName = g.pkgPrefix + fd.Symbol.Name
	} else {
		llName = fd.Symbol.Name
	}

	// TODO: more intelligent extraction of the return type.
	returnType := fd.Symbol.Type.(*types.FuncType).ReturnType

	// Assign the function its corresponding LLVM function value.
	fd.Symbol.LLValue = g.generateLLFunction(
		llName,
		public,
		fd.Params,
		returnType,
		fd.Body,
	)
}

// generateOperDef generates an operator definition.
func (g *Generator) generateOperDef(od *ast.OperDef) {
	// Intrinsic operator definitions aren't compiled.
	if hasAnnot(od.Annotations, "intrinsic") {
		return
	}

	// All operators have the LLVM name `oper.overload.[overload_id]`.
	llName := fmt.Sprintf("%soper.overload.%d", g.pkgPrefix, od.Overload.ID)

	// TODO: more intelligent extraction of the return type.
	returnType := od.Overload.Signature.(*types.FuncType).ReturnType

	// Assign the operator overload its corresponding LLVM function value.
	od.Overload.LLValue = g.generateLLFunction(
		llName,
		// TODO: handle public overloads
		false,
		od.Params,
		returnType,
		od.Body,
	)
}

// generateStructDef generates a struct definition.
func (g *Generator) generateStructDef(sd *ast.StructDef) {
	// Get the struct type from the struct definition.
	st := sd.Symbol.Type.(*types.StructType)

	// Convert the field types skipping any unit types.
	fieldTypes := make([]lltypes.Type, len(st.Fields))
	n := 0
	for _, field := range st.Fields {
		if !types.IsUnit(field.Type) {
			fieldTypes[n] = g.convType(field.Type)
			n++
		}
	}
	fieldTypes = fieldTypes[:n]

	// Create the LLVM struct type.
	llvmStruct := lltypes.NewStruct(fieldTypes...)

	// Add the appropriate type definition.
	g.mod.NewTypeDef(g.pkgPrefix+sd.Symbol.Name, llvmStruct)

	// Set the LLVM type of the struct type.
	st.SetLLType(llvmStruct)
}

/* -------------------------------------------------------------------------- */

// generateLLFunction generates an LLVM function with the given name, visibility
// parameters, return type, and body.  The generated function is returned.
func (g *Generator) generateLLFunction(
	name string,
	public bool,
	params []*common.Symbol,
	returnType types.Type,
	body ast.ASTNode,
) *ir.Func {
	// Create the LLVM function.
	llParams := make([]*ir.Param, len(params))
	for i, param := range params {
		llParams[i] = ir.NewParam(param.Name, g.convType(param.Type))
		param.LLValue = llParams[i]
	}

	// Add the pointer return parameter as necessary.
	if types.IsPtrWrappedType(returnType) {
		ptrRetType := g.convType(returnType)
		retParam := ir.NewParam("return", ptrRetType)
		retParam.Attrs = append(retParam.Attrs, ir.SRet{Typ: ptrRetType.(*lltypes.PointerType).ElemType})
		llParams = append([]*ir.Param{retParam}, llParams...)
	}

	llFunc := g.mod.NewFunc(name, g.convReturnType(returnType), llParams...)

	// Set the function's linkage based on its visibility.
	if public {
		llFunc.Linkage = enum.LinkageExternal
	} else {
		llFunc.Linkage = enum.LinkageInternal
	}

	// Add the function body as a predicate to generate if it exists.
	if body != nil {
		g.bodyPredicates = append(g.bodyPredicates, bodyPredicate{
			LLFunc:     llFunc,
			Params:     params,
			ReturnType: returnType,
			Body:       body,
		})
	}

	return llFunc
}

/* -------------------------------------------------------------------------- */

// hasAnnot returns whether the annotation named name exists in annots.
func hasAnnot(annots map[string]ast.AnnotValue, name string) bool {
	_, ok := annots[name]
	return ok
}
