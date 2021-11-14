package generate

import (
	"chai/ast"
	"chai/typing"
	"log"
	"strconv"

	"github.com/llir/llvm/ir"
	"github.com/llir/llvm/ir/constant"
	"github.com/llir/llvm/ir/types"
	"github.com/llir/llvm/ir/value"
)

// genExpr generates an expression.  It takes a basic block to append onto and
// returns the last value of block if the value if not nothing.  If the value is
// nothing, `nil` is returned.
func (g *Generator) genExpr(block *ir.Block, expr ast.Expr) value.Value {
	switch v := expr.(type) {
	case *ast.Cast:
		// there are no valid casts from nothing to another type so we don't
		// have to do any nothing pruning checks here.
		return g.genCast(block, g.genExpr(block, v.Src), v.Src.Type(), v.Type())
	case *ast.Identifier:
		{
			val, mut := g.lookup(v.Name)
			if mut {
				// load mutable values since they are always wrapped in pointers
				return block.NewLoad(val.Type().(*types.PointerType).ElemType, val)
			}

			return val
		}
	case *ast.Literal:
		return g.genLiteral(block, v)
	}

	return nil
}

// genCast generates a type cast.
func (g *Generator) genCast(block *ir.Block, srcVal value.Value, srcType, dstType typing.DataType) value.Value {
	// types are equal: no cast necessary
	if srcType.Equiv(dstType) {
		return srcVal
	}

	switch v := srcType.(type) {
	case typing.PrimType:
		// only valid casts from primitive types are to other primitive types
		dpt := dstType.(typing.PrimType)

		// float to double
		if v == typing.PrimF32 && dpt == typing.PrimF64 {
			return block.NewFPExt(srcVal, types.Double)
		}

		// double to float
		if v == typing.PrimF64 && dpt == typing.PrimF32 {
			return block.NewFPTrunc(srcVal, types.Float)
		}

		// int to float/double
		if v < typing.PrimF32 && dpt == typing.PrimF32 || dpt == typing.PrimF64 {
			// unsigned int to float/double
			if v < typing.PrimI8 {
				return block.NewUIToFP(srcVal, g.convPrimType(dpt))
			} else {
				// signed int to float/double
				return block.NewSIToFP(srcVal, g.convPrimType(dpt))
			}
		}

		// float/double to int
		if v == typing.PrimF32 || v == typing.PrimF64 && dpt < typing.PrimF32 {
			if dpt < typing.PrimI8 {
				// float/double to unsigned int
				return block.NewFPToUI(srcVal, g.convPrimType(dpt))
			} else {
				// float/double to signed int
				return block.NewFPToSI(srcVal, g.convPrimType(dpt))
			}
		}

		// bool to int
		if v == typing.PrimBool {
			// always zext (booleans are never signed)
			return block.NewZExt(srcVal, g.convPrimType(dpt))
		}

		// TODO: rune to string

		// TODO: int to int casting
	case *typing.RefType:
		// TEMPORARY: remove this cheeky bitcast once `core.unsafe` is implemented
		return block.NewBitCast(srcVal, g.convType(dstType))
	}

	log.Fatalln("cast not yet implemented")
	return nil
}

// genLiteral generates a literal constant.
func (g *Generator) genLiteral(block *ir.Block, lit *ast.Literal) value.Value {
	// handle null
	if lit.Value == "null" {
		// TODO
		return nil
	}

	// all other literals should be primitive types
	pt := lit.Type().(typing.PrimType)
	switch pt {
	case typing.PrimBool:
		if lit.Value == "true" {
			return constant.NewBool(true)
		}

		return constant.NewBool(false)
	case typing.PrimU8, typing.PrimI8:
		return g.genIntLit(lit.Value, types.I8, 8)
	case typing.PrimU16, typing.PrimI16:
		return g.genIntLit(lit.Value, types.I16, 16)
	case typing.PrimU32, typing.PrimI32:
		// TODO: handle rune literals
		return g.genIntLit(lit.Value, types.I32, 32)
	case typing.PrimU64, typing.PrimI64:
		return g.genIntLit(lit.Value, types.I64, 64)
	case typing.PrimF32:
		{
			// strconv should always succeed (parsed by lexer)
			x, _ := strconv.ParseFloat(lit.Value, 32)
			return constant.NewFloat(types.Float, x)
		}
	case typing.PrimF64:
		{
			// strconv should always succeed (parsed by lexer)
			x, _ := strconv.ParseFloat(lit.Value, 64)
			return constant.NewFloat(types.Double, x)
		}
	case typing.PrimString:
		{
			// this code just generates a new structure allocation for the
			// string struct
			str := block.NewAlloca(g.stringType)

			strBytes := g.mod.NewGlobalDef(strconv.Itoa(g.globalCounter), constant.NewCharArrayFromString(lit.Value))
			strBytesField := block.NewGetElementPtr(
				types.I8Ptr, str, constant.NewInt(types.I32, 0), constant.NewInt(types.I32, 0),
			)
			block.NewStore(strBytes, strBytesField)

			strLenField := block.NewGetElementPtr(
				types.I32, str, constant.NewInt(types.I32, 0), constant.NewInt(types.I32, 1),
			)
			block.NewStore(strLenField, constant.NewInt(types.I32, int64(len(lit.Value))))

			return str
		}
	}

	// nothing type => nil (nothing pruning)
	return nil
}

// genIntLit generates an integer literal.
func (g *Generator) genIntLit(value string, llType types.Type, bitsize int) value.Value {
	// From: https://pkg.go.dev/strconv#ParseInt
	// If the base argument is 0, the true base is implied by the string's
	// prefix following the sign (if present): 2 for "0b", 8 for "0" or "0o", 16
	// for "0x", and 10 otherwise. Also, for argument base 0 only, underscore
	// characters are permitted as defined by the Go syntax for integer
	// literals.
	// Note: should always succeed
	x, _ := strconv.ParseInt(value, 0, bitsize)
	return constant.NewInt(llType.(*types.IntType), x)
}
