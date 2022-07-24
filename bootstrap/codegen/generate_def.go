package codegen

import (
	"chaic/ast"
	"chaic/common"
	"chaic/llvm"
	"chaic/report"
	"chaic/types"
	"fmt"
)

// generateDef generates an LLVM IR definition.
func (g *Generator) generateDef(def ast.ASTNode) {
	switch v := def.(type) {
	case *ast.FuncDef:
		g.generateFuncDef(v)
	case *ast.OperDef:
		g.generateOperDef(v)
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

// -----------------------------------------------------------------------------

// generateLLFunction generates an LLVM function with the given name, visibility
// parameters, return type, and body.  The generated function is returned.
func (g *Generator) generateLLFunction(
	name string,
	public bool,
	params []*common.Symbol,
	returnType types.Type,
	body ast.ASTNode,
) llvm.Function {
	// Create the LLVM function type.
	llParamTypes := make([]llvm.Type, len(params))
	for i, param := range params {
		llParamTypes[i] = g.convType(param.Type)
	}

	llFuncType := llvm.NewFunctionType(
		g.convReturnType(returnType),
		llParamTypes...,
	)

	// Create the LLVM function.
	llFunc := g.mod.AddFunction(name, llFuncType)

	// Set the function's linkage based on its visibility.
	if public {
		llFunc.SetLinkage(llvm.ExternalLinkage)
	} else {
		llFunc.SetLinkage(llvm.InternalLinkage)
	}

	// Assign all the function parameters their corresponding LLVM types and
	// update the LLVM parameters with their appropriate names.
	for i, param := range params {
		llParam := llFunc.GetParam(i)

		llParam.SetName(param.Name)

		param.LLType = llParamTypes[i]
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

	// Return the created function.
	return llFunc
}

// -----------------------------------------------------------------------------

// hasAnnot returns whether the annotation named name exists in annots.
func hasAnnot(annots map[string]ast.AnnotValue, name string) bool {
	_, ok := annots[name]
	return ok
}
