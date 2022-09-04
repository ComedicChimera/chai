package codegen

import (
	"chaic/ast"
	"chaic/common"
	"chaic/depm"
	"chaic/llc"
	"chaic/report"
	"chaic/types"
	"chaic/util"
	"fmt"

	"github.com/llir/llvm/ir"
	"github.com/llir/llvm/ir/constant"
	lltypes "github.com/llir/llvm/ir/types"
	llvalue "github.com/llir/llvm/ir/value"
)

// Generator is responsible for converting Chai to LLVM IR.
type Generator struct {
	// The LLVM module being generated.
	mod *ir.Module

	// The prefix to prepend before all global symbols to prevent definition
	// collisions between symbols in different packages.
	pkgPrefix string

	// The list of body predicate extracted from definitions.
	bodyPredicates []bodyPredicate

	// The variable block of the LLVM function being generated.
	varBlock *ir.Block

	// The current block instructions are being inserted in.
	block *ir.Block

	// The table of declared LLVM intrinsics.
	llvmIntrinsics map[string]llvalue.Value

	// The loop context stack.
	loopContextStack []loopContext

	// A global reference to the frequently used `memcpy` intrinsic.
	memcpy llvalue.Value
}

// bodyPredicate represents the predicate of a function or operator body.
type bodyPredicate struct {
	// The LLVM function.
	LLFunc *ir.Func

	// The parameter symbols.
	Params []*common.Symbol

	// The return type of the function.
	ReturnType types.Type

	// The AST body.
	Body ast.ASTNode
}

// loopContext stores the contextual block destinations for the break and
// continue statements.
type loopContext struct {
	breakDest, continueDest *ir.Block
}

// Generate generates a Chai package into an LLVM module.
func Generate(ctx *llc.Context, pkg *depm.ChaiPackage) *llc.Module {
	// The LLVM name of the package.
	llPkgName := fmt.Sprintf("pkg%d", pkg.ID)

	// Create the LLVM module for the package.
	mod := ir.NewModule()
	mod.SourceFilename = pkg.AbsPath

	// Create the code generator for the package.
	g := Generator{
		mod:            mod,
		pkgPrefix:      llPkgName + ".",
		llvmIntrinsics: make(map[string]llvalue.Value),
	}

	// Generate the `memcpy` intrinsic.
	g.memcpy = g.getIntrinsic(
		"llvm.memcpy.p0i8.p0i8.i64",
		lltypes.Void,
		lltypes.I8Ptr,
		lltypes.I8Ptr,
		lltypes.I64,
		lltypes.I1,
	)

	// TODO: generate imports.

	// Generate all the definitions.
	for _, file := range pkg.Files {
		for _, def := range file.Definitions {
			g.generateDef(def)
		}
	}

	// Generate all the body predicates.
	for _, bodyPred := range g.bodyPredicates {
		g.generateBodyPredicate(bodyPred)
	}

	// Convert the LLIR module into an LLVM module.
	llMod, err := ctx.NewModuleFromIR(g.mod.String())
	if err != nil {
		report.ReportICE("failed to convert LLIR module to LLVM module:\n%s\n%s", err.Error(), g.mod.String())
		return nil
	}

	return llMod
}

/* -------------------------------------------------------------------------- */

// generateBodyPrediate generates a body predicate.
func (g *Generator) generateBodyPredicate(pred bodyPredicate) {
	// Add the variable block.
	g.varBlock = pred.LLFunc.NewBlock("entry")

	// Generate all the function parameters.
	for _, param := range pred.Params {
		if !types.IsPtrWrappedType(param.Type) {
			paramVar := g.varBlock.NewAlloca(g.convAllocType(param.Type))
			g.varBlock.NewStore(param.LLValue, paramVar)
			param.LLValue = paramVar
		}
	}

	// Generate the function body itself.
	firstBlock := pred.LLFunc.NewBlock("body")
	g.block = firstBlock

	if predExpr, ok := pred.Body.(ast.ASTExpr); ok {
		// Body is an expression.
		predValue := g.generateExpr(predExpr)

		// If the value is non-unit, return it.
		if !types.IsUnit(predExpr.Type()) {
			g.block.NewRet(predValue)
		}
	} else {
		// Body is a block.
		g.generateBlock(pred.Body.(*ast.Block))
	}

	// If the block we are now positioned in (the last block of the function)
	// is missing a terminator, then we know the function must return void
	// and so we add in the implicit `ret void` at the end.
	if g.block.Term == nil {
		g.block.NewRet(nil)
	}

	// Build a terminator for the var block to the first code block if the
	// variable block is non-empty.  Otherwise, remove it.
	if len(g.varBlock.Insts) > 0 {
		g.varBlock.NewBr(firstBlock)
	} else {
		pred.LLFunc.Blocks = pred.LLFunc.Blocks[1:]
	}
}

/* -------------------------------------------------------------------------- */

// getIntrinsic gets an LLVM intrinsic function.  The given name is the the full
// name of the intrinsic included any overload types.
func (g *Generator) getIntrinsic(name string, returnType lltypes.Type, paramTypes ...lltypes.Type) llvalue.Value {

	// Check if such a name already exists.
	if llvmIntrinsic, ok := g.llvmIntrinsics[name]; ok {
		// Return the declared value if it already exists.
		return llvmIntrinsic
	} else {
		// Otherwise, create a new intrinsic declaration.
		params := make([]*ir.Param, len(paramTypes))
		for i, paramType := range paramTypes {
			params[i] = ir.NewParam(fmt.Sprintf("p%d", i), paramType)
		}
		llvmIntrinsic := g.mod.NewFunc(name, returnType, params...)

		// Add the intrinsic to the intrinsics table
		g.llvmIntrinsics[name] = llvmIntrinsic

		// Return the created intrinsic
		return llvmIntrinsic
	}
}

// callFunc calls an LLVM function with args.
func (g *Generator) callFunc(returnType types.Type, fn llvalue.Value, args ...llvalue.Value) llvalue.Value {
	// TODO: handle copying arguments: I need to get the type and category some
	// how

	// copiedArgs := make([]llvalue.Value, len(args))
	// for i, arg := range args {

	// 	copiedArgs[i] = arg
	// }

	var result llvalue.Value
	if types.IsPtrWrappedType(returnType) {
		// TODO: return argument allocation elision
		result = g.varBlock.NewAlloca(g.convAllocType(returnType))
		g.block.NewCall(fn, append([]llvalue.Value{result}, args...)...)
	} else {
		result = g.block.NewCall(fn, args...)
	}

	if types.IsUnit(returnType) {
		return constant.NewInt(lltypes.I1, 0)
	}

	return result
}

// appendBlock adds a new block to the current function.
func (g *Generator) appendBlock() *ir.Block {
	return g.block.Parent.NewBlock("")
}

// copyInto copies one value into another based on its type.
func (g *Generator) copyInto(typ types.Type, src, dest llvalue.Value) {
	if types.IsPtrWrappedType(typ) {
		destI8Ptr := g.block.NewBitCast(dest, lltypes.I8Ptr)
		srcI8Ptr := g.block.NewBitCast(src, lltypes.I8Ptr)

		g.block.NewCall(g.memcpy, destI8Ptr, srcI8Ptr, constant.NewInt(lltypes.I64, int64(typ.Size())), constant.False)
	} else {
		g.block.NewStore(src, dest)
	}
}

/* -------------------------------------------------------------------------- */

// convType converts the typ to its LLVM type.
func (g *Generator) convType(typ types.Type) lltypes.Type {
	return g.convInnerType(types.InnerType(typ), false, false)
}

// convReturnType converts typ to its LLVM type assuming typ is a return type.
func (g *Generator) convReturnType(typ types.Type) lltypes.Type {
	return g.convInnerType(types.InnerType(typ), true, false)
}

// convAllocType converts typ to its LLVM type assuming typ is used in an
// `alloca` instruction: some types which are pointers normally are not passed
// as pointers to `alloca` -- this function handles that case.
func (g *Generator) convAllocType(typ types.Type) lltypes.Type {
	return g.convInnerType(types.InnerType(typ), false, true)
}

// convInnerType converts the given Chai inner type to its LLVM type.  This
// should generally not be called except from within other type conversion
// functions.
func (g *Generator) convInnerType(typ types.Type, isReturnType, isAllocType bool) lltypes.Type {
	switch v := typ.(type) {
	case types.PrimitiveType:
		return g.convPrimType(v, isReturnType)
	case *types.PointerType:
		return lltypes.NewPointer(g.convType(v.ElemType))
	case *types.StructType:
		if isReturnType {
			return lltypes.Void
		} else if typ.Size() <= 2*util.PointerSize || isAllocType {
			return v.LLType
		} else {
			return lltypes.NewPointer(v.LLType)
		}
	default:
		report.ReportICE("type codegen not implemented")
		return nil
	}
}

// convPrimType converts a Chai primitive type to its LLVM type.
func (g *Generator) convPrimType(pt types.PrimitiveType, isReturnType bool) lltypes.Type {
	switch pt {
	case types.PrimTypeBool:
		return lltypes.I1
	case types.PrimTypeI8, types.PrimTypeU8:
		return lltypes.I8
	case types.PrimTypeI16, types.PrimTypeU16:
		return lltypes.I16
	case types.PrimTypeI32, types.PrimTypeU32:
		return lltypes.I32
	case types.PrimTypeI64, types.PrimTypeU64:
		return lltypes.I64
	case types.PrimTypeF32:
		return lltypes.Float
	case types.PrimTypeF64:
		return lltypes.Double
	case types.PrimTypeUnit:
		if isReturnType {
			return lltypes.Void
		} else {
			return lltypes.NewStruct()
		}
	default:
		// unreachable
		return nil
	}
}

/* -------------------------------------------------------------------------- */

// currLoopContext returns the current loop context.
func (g *Generator) currLoopContext() loopContext {
	return g.loopContextStack[len(g.loopContextStack)-1]
}

// pushLoopContext pushes a new loop context onto the loop context stack.
func (g *Generator) pushLoopContext(breakDest, continueDest *ir.Block) {
	g.loopContextStack = append(g.loopContextStack, loopContext{breakDest: breakDest, continueDest: continueDest})
}

// popLoopContext pops a loop context off of the loop context stack.
func (g *Generator) popLoopContext() {
	g.loopContextStack = g.loopContextStack[:len(g.loopContextStack)-1]
}
