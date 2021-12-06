package generate

import (
	"chai/ast"
	"chai/typing"
	"fmt"
	"log"
	"strconv"

	"github.com/llir/llvm/ir"
	"github.com/llir/llvm/ir/constant"
	"github.com/llir/llvm/ir/enum"
	"github.com/llir/llvm/ir/types"
	"github.com/llir/llvm/ir/value"
)

// genExpr generates an expression.  It takes a basic block to append onto and
// returns the last value of block if the value if not nothing.  If the value is
// nothing, `nil` is returned.
func (g *Generator) genExpr(expr ast.Expr) value.Value {
	switch v := expr.(type) {
	case *ast.Block:
		return g.genBlock(v)
	case *ast.Cast:
		// there are no valid casts from nothing to another type so we don't
		// have to do any nothing pruning checks here.
		return g.genCast(g.genExpr(v.Src), v.Src.Type(), v.Type())
	case *ast.Call:
		return g.genCall(v)
	case *ast.BinaryOp:
		return g.genOpCall(v.Op, v.Lhs, v.Rhs)
	case *ast.UnaryOp:
		return g.genOpCall(v.Op, v.Operand)
	case *ast.Identifier:
		{
			// test for pruned identifiers
			if typing.IsNothing(v.Type()) {
				return nil
			}

			val, mut := g.lookup(v.Name)
			if mut {
				// load mutable values since they are always wrapped in pointers
				return g.block.NewLoad(val.Type().(*types.PointerType).ElemType, val)
			}

			return val
		}
	case *ast.Literal:
		return g.genLiteral(v)
	case *ASTWrappedLLVMVal:
		return v.Val
	}

	log.Fatalln("AST node generation not yet supported")
	return nil
}

// genCast generates a type cast.
func (g *Generator) genCast(srcVal value.Value, srcType, dstType typing.DataType) value.Value {
	// types are equal: no cast necessary
	if srcType.Equiv(dstType) {
		return srcVal
	}

	// make sure to extract the inner type -- that is all we care about
	srcType = typing.InnerType(srcType)
	dstType = typing.InnerType(dstType)

	switch v := srcType.(type) {
	case typing.PrimType:
		// only valid casts from primitive types are to other primitive types
		dpt := dstType.(typing.PrimType)

		// float to double
		if v == typing.PrimF32 && dpt == typing.PrimF64 {
			return g.block.NewFPExt(srcVal, types.Double)
		}

		// double to float
		if v == typing.PrimF64 && dpt == typing.PrimF32 {
			return g.block.NewFPTrunc(srcVal, types.Float)
		}

		// int to float/double
		if v < typing.PrimF32 && dpt == typing.PrimF32 || dpt == typing.PrimF64 {
			// unsigned int to float/double
			if v < typing.PrimI8 {
				return g.block.NewUIToFP(srcVal, g.convPrimType(dpt))
			} else {
				// signed int to float/double
				return g.block.NewSIToFP(srcVal, g.convPrimType(dpt))
			}
		}

		// float/double to int
		if v == typing.PrimF32 || v == typing.PrimF64 && dpt < typing.PrimF32 {
			if dpt < typing.PrimI8 {
				// float/double to unsigned int
				return g.block.NewFPToUI(srcVal, g.convPrimType(dpt))
			} else {
				// float/double to signed int
				return g.block.NewFPToSI(srcVal, g.convPrimType(dpt))
			}
		}

		// bool to int
		if v == typing.PrimBool {
			// always zext (booleans are never signed)
			return g.block.NewZExt(srcVal, g.convPrimType(dpt))
		}

		// TODO: rune to string

		// int to int casting
		if v < typing.PrimF32 && dpt < typing.PrimF32 {
			// signed to unsigned or unsigned to signed => nop
			if v-dpt == 4 || v-dpt == -4 {
				return srcVal
			}

			// TODO: rest
		}
	case *typing.RefType:
		// TEMPORARY: remove this cheeky bitcast once `core.unsafe` is implemented
		return g.block.NewBitCast(srcVal, g.convType(dstType))
	}

	log.Fatalln("cast not yet implemented")
	return nil
}

// -----------------------------------------------------------------------------

// genOpCall generates an operator application.
func (g *Generator) genOpCall(op ast.Oper, operands ...ast.Expr) value.Value {
	// test for intrinsics
	if iname := typing.InnerType(op.Signature).(*typing.FuncType).IntrinsicName; iname != "" {
		return g.genIntrinsic(iname, operands)
	}

	// TODO: non-intrinsic operator definitions
	log.Fatalln("non-intrinsic operators not yet supported")
	return nil
}

// genCall generates a function call.
func (g *Generator) genCall(call *ast.Call) value.Value {
	// test for intrinsics
	if iname := call.Func.Type().(*typing.FuncType).IntrinsicName; iname != "" {
		return g.genIntrinsic(iname, call.Args)
	}

	llFunc := g.genExpr(call.Func)

	var llExprs []value.Value

	// we don't include arguments that compile to nothing in the function call.
	// They are still evaluated since they may have side-effects.
	for _, expr := range call.Args {
		if val := g.genExpr(expr); val != nil {
			llExprs = append(llExprs, val)
		}
	}

	return g.block.NewCall(llFunc, llExprs...)
}

// genIntrinsic generates an intrinsic instruction based on an intrinsic name
// and some operands to the intrinsic.
func (g *Generator) genIntrinsic(iname string, operands []ast.Expr) value.Value {
	// handle intrinsics that need special generation of their operands
	switch iname {
	// short circuit evaluators generate their operands in weird order
	// (as opposed to all before the operation) so we put them here.
	case "lor":
		// the short circuit evaluation code for logical OR appears roughly as
		// follows (notice the absence of an actual usage of the `or`
		// instruction since it is encoded in the logic of the logical OR short
		// circuit evaluation)
		//
		// orbegin:
		//   %a = ...
		//   br i1 %a, label %orend, label %orfalse
		// orfalse:
		//   %b = ...
		//   br label %orend
		// orend:
		//   %result = phi i1 [ 1, %orbegin ], [ %b, %orfalse ]

		a := g.genExpr(operands[0])

		beginBlock := g.block
		falseBlock := g.appendBlock()
		endBlock := g.appendBlock()

		g.block.NewCondBr(a, endBlock, falseBlock)

		g.block = falseBlock
		b := g.genExpr(operands[1])

		g.block = endBlock
		return g.block.NewPhi(ir.NewIncoming(constant.NewBool(true), beginBlock), ir.NewIncoming(b, falseBlock))
	}

	// Note: no intrinsic (except `eq` and `neq`, handled above) accepts
	// `nothing` types so we can just naively convert our operands to LLVM
	// values

	// all other intrinsics convert their operands first and then operate upon
	// them
	llOperands := make([]value.Value, len(operands))
	for i, op := range operands {
		llOperands[i] = g.genExpr(op)
	}

	// match of the name of the intrinsic and generate he corresponding
	// instruction
	switch iname {
	case "__strbytes":
		strBytesPtr := g.block.NewBitCast(llOperands[0], types.NewPointer(types.I8Ptr))
		return g.block.NewLoad(types.I8Ptr, strBytesPtr)
	case "__strlen":
		{
			lenFieldPtr := g.block.NewGetElementPtr(
				g.stringType,
				llOperands[0],
				constant.NewInt(types.I32, 0),
				constant.NewInt(types.I32, 1),
			)

			return g.block.NewLoad(types.I32, lenFieldPtr)
		}
	case "ineg":
		// -int => 0 - int
		return g.block.NewSub(constant.NewInt(llOperands[0].Type().(*types.IntType), 0), llOperands[0])
	case "iadd":
		return g.block.NewAdd(llOperands[0], llOperands[1])
	case "isub":
		return g.block.NewSub(llOperands[0], llOperands[1])
	case "sdiv":
		return g.block.NewSDiv(llOperands[0], llOperands[1])
	case "smod":
		return g.block.NewSRem(llOperands[0], llOperands[1])
	case "ilt":
		return g.block.NewICmp(enum.IPredSLT, llOperands[0], llOperands[1])
	case "__init":
		return g.block.NewCall(g.initFunc)
	}

	log.Fatalln("intrinsic not implemented yet")
	return nil
}

// -----------------------------------------------------------------------------

// genLiteral generates a literal constant.
func (g *Generator) genLiteral(lit *ast.Literal) value.Value {
	// handle null
	if lit.Value == "null" {
		return g.genNull(lit.Type())
	}

	// all other literals should be primitive types
	pt := typing.InnerType(lit.Type()).(typing.PrimType)
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
			// string struct.  NOTE: this currently allocates the string struct
			// itself on the stack, but this may not be the best way to handle
			// string literals (the data is, of course, interned globally).
			// Improvements: TBD
			str := g.block.NewAlloca(g.stringType)

			strBytes := g.mod.NewGlobalDef(fmt.Sprintf("__strlit.%d", g.globalCounter), constant.NewCharArrayFromString(lit.Value))
			g.globalCounter++
			strBytesField := g.block.NewGetElementPtr(
				g.stringType, str, constant.NewInt(types.I32, 0), constant.NewInt(types.I32, 0),
			)
			strBytesPtr := g.block.NewBitCast(strBytes, types.I8Ptr)
			g.block.NewStore(strBytesPtr, strBytesField)

			strLenField := g.block.NewGetElementPtr(
				g.stringType, str, constant.NewInt(types.I32, 0), constant.NewInt(types.I32, 1),
			)
			g.block.NewStore(constant.NewInt(types.I32, int64(len(lit.Value))), strLenField)

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

// getNull generates an appropriate `null` value for a type.
func (g *Generator) genNull(typ typing.DataType) value.Value {
	// NOTE: references are not supposed to be nullable: this is here
	// temporarily to let me get a hello world demo up and running but will be
	// removed in favor of a more sensible intrinsic later.
	if rt, ok := typing.InnerType(typ).(*typing.RefType); ok {
		return constant.NewNull(g.convType(rt).(*types.PointerType))
	}

	log.Fatalln("null is not yet implemented for this type")
	return nil
}
