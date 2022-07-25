package codegen

import (
	"chaic/ast"
	"chaic/common"
	"chaic/depm"
	"chaic/llvm"
	"chaic/report"
	"chaic/types"
	"fmt"
)

type Generator struct {
	// The LLVM context for the generated.
	ctx *llvm.Context

	// The LLVM module being generated.
	mod llvm.Module

	// The prefix to prepend before all symbols to prevent definition collisions
	// between symbols in different packages.
	pkgPrefix string

	// The list of body predicate extracted from definitions.
	bodyPredicates []bodyPredicate

	// The IR builder for this generator.
	irb llvm.IRBuilder

	// The current LLVM function body being generated.
	body llvm.FuncBody

	// The entry/variable block of the LLVM function being generated.
	varBlock llvm.BasicBlock
}

// bodyPredicate represents the predicate of a function or operator body.
type bodyPredicate struct {
	// The LLVM function.
	LLFunc llvm.Function

	// The parameter symbols.
	Params []*common.Symbol

	// The return type of the function.
	ReturnType types.Type

	// The AST body.
	Body ast.ASTNode
}

// Generate generates a Chai package into an LLVM mdoule.
func Generate(ctx *llvm.Context, pkg *depm.ChaiPackage) llvm.Module {
	// The LLVM name of the package.
	llPkgName := fmt.Sprintf("pkg%d", uint(pkg.ID))

	// Create the LLVM module for the package.
	mod := ctx.NewModule(llPkgName)

	// Create the code generator for the package.
	g := Generator{
		ctx:       ctx,
		mod:       mod,
		pkgPrefix: llPkgName + ".",
	}

	// TODO: generate imports.

	// Generate all the definitions.
	for _, file := range pkg.Files {
		for _, def := range file.Definitions {
			g.generateDef(def)
		}
	}

	// Initialize the generator's IR builder.
	g.irb = ctx.NewBuilder()

	// Generate all the body predicates.
	for _, bodyPred := range g.bodyPredicates {
		g.generateBodyPredicate(bodyPred)
	}

	// Verify that the module we generated correctly.
	if err := mod.Verify(); err != nil {
		report.ReportICE("failed to verify LLVM module:\n%s", err)
	}

	// Return the generated module.
	return mod
}

// -----------------------------------------------------------------------------

// generateBodyPredicate generates a body predicate.
func (g *Generator) generateBodyPredicate(pred bodyPredicate) {
	// Set the function body being compiled.
	g.body = pred.LLFunc.Body()

	// Add the entry/variable block to the function.
	g.varBlock = g.body.AppendNamed("entry")

	// Generate all the function parameters.
	g.irb.MoveToEnd(g.varBlock)
	for i, param := range pred.Params {
		llParam := pred.LLFunc.GetParam(i)

		paramVar := g.irb.BuildAlloca(g.convAllocType(param.Type))
		g.irb.BuildStore(llParam, paramVar)
		param.LLValue = paramVar
	}

	// Generate the function body itself.
	firstBlock := g.body.Append()
	g.irb.MoveToEnd(firstBlock)
	if predExpr, ok := pred.Body.(ast.ASTExpr); ok {
		// Generate the expression itself.
		predValue := g.generateExpr(predExpr)

		// If the value is non-unit, return it.
		if !types.IsUnit(predExpr.Type()) {
			g.irb.BuildRet(predValue)
		}
	} else {
		// Generate the block.
		g.generateBlock(pred.Body.(*ast.Block))
	}

	// If the last block the IR builder is positioned over (the last block of
	// the function) is missing a terminator, then we know the function must
	// return void and so we add in the implicit `ret void` at the end.
	if _, exists := g.irb.Block().Terminator(); !exists {
		g.irb.BuildRet()
	}

	// Build a terminator for the vat block to the first code block.
	g.irb.MoveToEnd(g.varBlock)
	g.irb.BuildBr(firstBlock)
}

// -----------------------------------------------------------------------------

// getIntrinsics looks up an overloaded LLVM intrinsic by name.
func (g *Generator) getIntrinsic(intrinsicName string, overloadTypes ...llvm.Type) llvm.Function {
	intrinsic, exists := g.mod.GetIntrinsic(intrinsicName, overloadTypes...)

	if !exists {
		report.ReportICE("failed to find LLVM intrinsic named `%s`", intrinsicName)
	}

	return intrinsic
}

// callFunc calls an LLVM function with args.
func (g *Generator) callFunc(returnType types.Type, fn llvm.Value, args ...llvm.Value) llvm.Value {
	result := g.irb.BuildCall(g.convReturnType(returnType), fn, args...)

	if types.IsUnit(returnType) {
		return llvm.ConstInt(g.ctx.Int8Type(), 0, false)
	}

	return result
}

// -----------------------------------------------------------------------------

// convType converts the typ to its LLVM type.
func (g *Generator) convType(typ types.Type) llvm.Type {
	return g.convInnerType(types.InnerType(typ), false, false)
}

// convReturnType converts typ to its LLVM type assuming typ is a return type.
func (g *Generator) convReturnType(typ types.Type) llvm.Type {
	return g.convInnerType(types.InnerType(typ), true, false)
}

// convAllocType converts typ to its LLVM type assuming typ is used in an
// `alloca` instruction: some types which are pointers normally are not passed
// as pointers to `alloca` -- this function handles that case.
func (g *Generator) convAllocType(typ types.Type) llvm.Type {
	return g.convInnerType(types.InnerType(typ), false, true)
}

// convInnerType converts the given Chai inner type to its LLVM type.  This
// should generally not be called except from within other type conversion
// functions.
func (g *Generator) convInnerType(typ types.Type, isReturnType bool, isAllocType bool) llvm.Type {
	switch v := typ.(type) {
	case types.PrimitiveType:
		return g.convPrimType(v, isReturnType)
	case *types.PointerType:
		return llvm.NewPointerType(g.convType(v.ElemType))
	default:
		report.ReportICE("type codegen not implemented")
		return nil
	}
}

// convPrimType converts a Chai primitive type to its LLVM type.
func (g *Generator) convPrimType(pt types.PrimitiveType, isReturnType bool) llvm.Type {
	switch pt {
	case types.PrimTypeBool:
		return g.ctx.Int1Type()
	case types.PrimTypeI8, types.PrimTypeU8:
		return g.ctx.Int8Type()
	case types.PrimTypeI16, types.PrimTypeU16:
		return g.ctx.Int16Type()
	case types.PrimTypeI32, types.PrimTypeU32:
		return g.ctx.Int32Type()
	case types.PrimTypeI64, types.PrimTypeU64:
		return g.ctx.Int64Type()
	case types.PrimTypeF32:
		return g.ctx.FloatType()
	case types.PrimTypeF64:
		return g.ctx.DoubleType()
	case types.PrimTypeUnit:
		if isReturnType {
			return g.ctx.VoidType()
		} else {
			return g.ctx.Int1Type()
		}
	default:
		// unreachable
		return nil
	}
}
