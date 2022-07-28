package codegen

import (
	"chaic/ast"
	"chaic/common"
	"chaic/depm"
	"chaic/report"
	"chaic/types"
	"fmt"

	"github.com/llir/llvm/ir"
	lltypes "github.com/llir/llvm/ir/types"
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

// Generate generates a Chai package into an LLVM module.
func Generate(pkg *depm.ChaiPackage) *ir.Module {
	// The LLVM name of the package.
	llPkgName := fmt.Sprintf("pkg%d", uint(pkg.ID))

	// Create the LLVM module for the package.
	mod := ir.NewModule()
	mod.SourceFilename = pkg.AbsPath

	// Create the code generator for the package.
	g := Generator{
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

	// Generate all the body predicates.
	for _, bodyPred := range g.bodyPredicates {
		_ = bodyPred
	}

	// TODO: verify/validate?

	return g.mod
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
			return lltypes.I1
		}
	default:
		// unreachable
		return nil
	}
}
