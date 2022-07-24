package llvm

/*
#include <stdlib.h>

#include "llvm-c/Core.h"
*/
import "C"

// IRBuilder represents an LLVM IR builder.
type IRBuilder struct {
	c C.LLVMBuilderRef
}

// NewBuilder creates a new IR builder in the given context.
func (c Context) NewBuilder() (irb IRBuilder) {
	irb.c = C.LLVMCreateBuilderInContext(c.c)

	c.takeOwnership(irb)
	return
}

// dispose disposes of the builder.
func (irb IRBuilder) dispose() {
	C.LLVMDisposeBuilder(irb.c)
}

// -----------------------------------------------------------------------------

// Block returns the current basic block the builder is positioned over.
func (irb IRBuilder) Block() BasicBlock {
	return BasicBlock{c: C.LLVMGetInsertBlock(irb.c)}
}

// MoveBefore moves the builder before instr.
func (irb IRBuilder) MoveBefore(instr Instruction) {
	C.LLVMPositionBuilderBefore(irb.c, instr.c)
}

// MoveAfter moves the builder after instr.
func (irb IRBuilder) MoveAfter(instr Instruction) {
	C.LLVMPositionBuilder(irb.c, instr.Parent().c, instr.c)
}

// MoveToStart moves the builder to the start of bb.
func (irb IRBuilder) MoveToStart(bb BasicBlock) {
	first, ok := bb.First()

	if ok {
		C.LLVMPositionBuilderBefore(irb.c, first.c)
	} else {
		C.LLVMPositionBuilderAtEnd(irb.c, bb.c)
	}
}

// MoveToEnd moves the builder to the end of bb.
func (irb IRBuilder) MoveToEnd(bb BasicBlock) {
	C.LLVMPositionBuilderAtEnd(irb.c, bb.c)
}

// -----------------------------------------------------------------------------

// BuildRet builds a `ret` instruction.
func (irb IRBuilder) BuildRet(values ...Value) (ret Terminator) {
	switch len(values) {
	case 0:
		ret.c = C.LLVMBuildRetVoid(irb.c)
	case 1:
		ret.c = C.LLVMBuildRet(irb.c, values[0].ptr())
	default:
		valPtrs := make([]C.LLVMValueRef, len(values))
		for i, value := range values {
			valPtrs[i] = value.ptr()
		}

		ret.c = C.LLVMBuildAggregateRet(irb.c, byref(&valPtrs[0]), (C.uint)(len(values)))
	}

	return
}

// BuildBr builds an unconditional `br` instruction.
func (irb IRBuilder) BuildBr(dest BasicBlock) (br Terminator) {
	br.c = C.LLVMBuildBr(irb.c, dest.c)
	return
}

// BuildCondBr builds a conditional `br` instruction.
func (irb IRBuilder) BuildCondBr(cond Value, thenBlock, elseBlock BasicBlock) (cbr ConditionalTerminator) {
	cbr.c = C.LLVMBuildCondBr(irb.c, cond.ptr(), thenBlock.c, elseBlock.c)
	return
}

// BuildSwitch builds a `switch` instruction.
func (irb IRBuilder) BuildSwitch(v Value, defaultBlock BasicBlock, expectedNumCases int) (sw SwitchInstruction) {
	sw.c = C.LLVMBuildSwitch(irb.c, v.ptr(), defaultBlock.c, (C.uint)(expectedNumCases))
	return
}

// BuildUnreachable builds an `unreachable` instruction.
func (irb IRBuilder) BuildUnreachable() (un Terminator) {
	un.c = C.LLVMBuildUnreachable(irb.c)
	return
}

// BuildAdd builds an `add` instruction.
func (irb IRBuilder) BuildAdd(lhs, rhs Value) (in Instruction) {
	in.c = C.LLVMBuildAdd(irb.c, lhs.ptr(), rhs.ptr(), nil)
	return
}

// BuildNSWAdd builds an `add nsw` instruction.
func (irb IRBuilder) BuildNSWAdd(lhs, rhs Value) (in Instruction) {
	in.c = C.LLVMBuildNSWAdd(irb.c, lhs.ptr(), rhs.ptr(), nil)
	return
}

// BuildNUWAdd builds an `add nuw` instruction.
func (irb IRBuilder) BuildNUWAdd(lhs, rhs Value) (in Instruction) {
	in.c = C.LLVMBuildNUWAdd(irb.c, lhs.ptr(), rhs.ptr(), nil)
	return
}

// BuildFAdd builds an `fadd` instruction.
func (irb IRBuilder) BuildFAdd(lhs, rhs Value) (in Instruction) {
	in.c = C.LLVMBuildFAdd(irb.c, lhs.ptr(), rhs.ptr(), nil)
	return
}

// BuildAdd builds an `sub` instruction.
func (irb IRBuilder) BuildSub(lhs, rhs Value) (in Instruction) {
	in.c = C.LLVMBuildSub(irb.c, lhs.ptr(), rhs.ptr(), nil)
	return
}

// BuildNSWAdd builds an `sub nsw` instruction.
func (irb IRBuilder) BuildNSWSub(lhs, rhs Value) (in Instruction) {
	in.c = C.LLVMBuildNSWSub(irb.c, lhs.ptr(), rhs.ptr(), nil)
	return
}

// BuildNUWAdd builds an `sub nuw` instruction.
func (irb IRBuilder) BuildNUWSub(lhs, rhs Value) (in Instruction) {
	in.c = C.LLVMBuildNUWSub(irb.c, lhs.ptr(), rhs.ptr(), nil)
	return
}

// BuildFAdd builds an `fsub` instruction.
func (irb IRBuilder) BuildFSub(lhs, rhs Value) (in Instruction) {
	in.c = C.LLVMBuildFSub(irb.c, lhs.ptr(), rhs.ptr(), nil)
	return
}

// BuildAdd builds an `mul` instruction.
func (irb IRBuilder) BuildMul(lhs, rhs Value) (in Instruction) {
	in.c = C.LLVMBuildMul(irb.c, lhs.ptr(), rhs.ptr(), nil)
	return
}

// BuildNSWAdd builds an `mul nsw` instruction.
func (irb IRBuilder) BuildNSWMul(lhs, rhs Value) (in Instruction) {
	in.c = C.LLVMBuildNSWMul(irb.c, lhs.ptr(), rhs.ptr(), nil)
	return
}

// BuildNUWAdd builds an `mul nuw` instruction.
func (irb IRBuilder) BuildNUWMul(lhs, rhs Value) (in Instruction) {
	in.c = C.LLVMBuildNUWMul(irb.c, lhs.ptr(), rhs.ptr(), nil)
	return
}

// BuildFAdd builds an `fmul` instruction.
func (irb IRBuilder) BuildFMul(lhs, rhs Value) (in Instruction) {
	in.c = C.LLVMBuildFMul(irb.c, lhs.ptr(), rhs.ptr(), nil)
	return
}

// BuildUDiv builds a `udiv` instruction.
func (irb IRBuilder) BuildUDiv(lhs, rhs Value) (in Instruction) {
	in.c = C.LLVMBuildUDiv(irb.c, lhs.ptr(), rhs.ptr(), nil)
	return
}

// BuildSDiv builds a `sdiv` instruction.
func (irb IRBuilder) BuildSDiv(lhs, rhs Value) (in Instruction) {
	in.c = C.LLVMBuildSDiv(irb.c, lhs.ptr(), rhs.ptr(), nil)
	return
}

// BuildExactUDiv builds a `udiv exact` instruction.
func (irb IRBuilder) BuildExactUDiv(lhs, rhs Value) (in Instruction) {
	in.c = C.LLVMBuildExactUDiv(irb.c, lhs.ptr(), rhs.ptr(), nil)
	return
}

// BuildExactSDiv builds a `sdiv exact` instruction.
func (irb IRBuilder) BuildExactSDiv(lhs, rhs Value) (in Instruction) {
	in.c = C.LLVMBuildExactSDiv(irb.c, lhs.ptr(), rhs.ptr(), nil)
	return
}

// BuildFDiv builds a `fdiv` instruction.
func (irb IRBuilder) BuildFDiv(lhs, rhs Value) (in Instruction) {
	in.c = C.LLVMBuildFDiv(irb.c, lhs.ptr(), rhs.ptr(), nil)
	return
}

// BuildURem builds a `urem` instruction.
func (irb IRBuilder) BuildURem(lhs, rhs Value) (in Instruction) {
	in.c = C.LLVMBuildURem(irb.c, lhs.ptr(), rhs.ptr(), nil)
	return
}

// BuildSRem builds a `srem` instruction.
func (irb IRBuilder) BuildSRem(lhs, rhs Value) (in Instruction) {
	in.c = C.LLVMBuildSRem(irb.c, lhs.ptr(), rhs.ptr(), nil)
	return
}

// BuildFRem builds a `frem` instruction.
func (irb IRBuilder) BuildFRem(lhs, rhs Value) (in Instruction) {
	in.c = C.LLVMBuildFRem(irb.c, lhs.ptr(), rhs.ptr(), nil)
	return
}

// BuildShl builds a `shl` instruction.
func (irb IRBuilder) BuildShl(lhs, rhs Value) (in Instruction) {
	in.c = C.LLVMBuildShl(irb.c, lhs.ptr(), rhs.ptr(), nil)
	return
}

// BuildLShr builds a `lshr` instruction.
func (irb IRBuilder) BuildLShr(lhs, rhs Value) (in Instruction) {
	in.c = C.LLVMBuildLShr(irb.c, lhs.ptr(), rhs.ptr(), nil)
	return
}

// BuildAShr builds a `ashr` instruction.
func (irb IRBuilder) BuildAShr(lhs, rhs Value) (in Instruction) {
	in.c = C.LLVMBuildAShr(irb.c, lhs.ptr(), rhs.ptr(), nil)
	return
}

// BuildAnd builds a `and` instruction.
func (irb IRBuilder) BuildAnd(lhs, rhs Value) (in Instruction) {
	in.c = C.LLVMBuildAnd(irb.c, lhs.ptr(), rhs.ptr(), nil)
	return
}

// BuildOr builds a `or` instruction.
func (irb IRBuilder) BuildOr(lhs, rhs Value) (in Instruction) {
	in.c = C.LLVMBuildOr(irb.c, lhs.ptr(), rhs.ptr(), nil)
	return
}

// BuildXor builds a `xor` instruction.
func (irb IRBuilder) BuildXor(lhs, rhs Value) (in Instruction) {
	in.c = C.LLVMBuildXor(irb.c, lhs.ptr(), rhs.ptr(), nil)
	return
}

// BuildNeg builds a `neg` instruction.
func (irb IRBuilder) BuildNeg(v Value) (in Instruction) {
	in.c = C.LLVMBuildNeg(irb.c, v.ptr(), nil)
	return
}

// BuildNSWNeg builds a `neg nsw` instruction.
func (irb IRBuilder) BuildNSWNeg(v Value) (in Instruction) {
	in.c = C.LLVMBuildNSWNeg(irb.c, v.ptr(), nil)
	return
}

// BuildNUWNeg builds a `neg nuw` instruction.
func (irb IRBuilder) BuildNUWNeg(v Value) (in Instruction) {
	in.c = C.LLVMBuildNUWNeg(irb.c, v.ptr(), nil)
	return
}

// BuildFNeg builds a `fneg` instruction.
func (irb IRBuilder) BuildFNeg(v Value) (in Instruction) {
	in.c = C.LLVMBuildFNeg(irb.c, v.ptr(), nil)
	return
}

// BuildNot builds a `not` instruction.
func (irb IRBuilder) BuildNot(v Value) (in Instruction) {
	in.c = C.LLVMBuildNot(irb.c, v.ptr(), nil)
	return
}

// BuildAlloca builds an `alloca` instruction.
func (irb IRBuilder) BuildAlloca(typ Type) (ain AllocaInstruction) {
	ain.c = C.LLVMBuildAlloca(irb.c, typ.ptr(), nil)
	return
}

// BuildLoad builds a `load` instruction.
func (irb IRBuilder) BuildLoad(loadedType Type, ptr Value) (in Instruction) {
	in.c = C.LLVMBuildLoad2(irb.c, loadedType.ptr(), ptr.ptr(), nil)
	return
}

// BuildStore builds a `store` instruction.
func (irb IRBuilder) BuildStore(val, ptr Value) (in Instruction) {
	in.c = C.LLVMBuildStore(irb.c, val.ptr(), ptr.ptr())
	return
}

// BuildGEP builds a `getelementptr` instruction.
func (irb IRBuilder) BuildGEP(pointeeTyp Type, ptr Value, indices ...Value) (gep GEPInstruction) {
	indicesArr := make([]C.LLVMValueRef, len(indices))
	for i, index := range indices {
		indicesArr[i] = index.ptr()
	}

	gep.c = C.LLVMBuildGEP2(irb.c, pointeeTyp.ptr(), ptr.ptr(), byref(&indicesArr[0]), (C.uint)(len(indices)), nil)
	return
}

// BuildInBoundsGEP builds a `getelementptr inbounds` instruction.
func (irb IRBuilder) BuildInBoundsGEP2(pointeeTyp Type, ptr Value, indices ...Value) (gep GEPInstruction) {
	indicesArr := make([]C.LLVMValueRef, len(indices))
	for i, index := range indices {
		indicesArr[i] = index.ptr()
	}

	gep.c = C.LLVMBuildInBoundsGEP2(irb.c, pointeeTyp.ptr(), ptr.ptr(), byref(&indicesArr[0]), (C.uint)(len(indices)), nil)
	return
}

// BuildStructGEP2 builds a `getelementptr` instruction for struct fields.
func (irb IRBuilder) BuildStructGEP2(pointeeTyp Type, ptr Value, ndx int) (gep GEPInstruction) {
	gep.c = C.LLVMBuildStructGEP2(irb.c, pointeeTyp.ptr(), ptr.ptr(), (C.uint)(ndx), nil)
	return
}

// BuildTrunc builds a `trunc` instruction.
func (irb IRBuilder) BuildTrunc(src Value, dest Type) (in Instruction) {
	in.c = C.LLVMBuildTrunc(irb.c, src.ptr(), dest.ptr(), nil)
	return
}

// BuildZExt builds a `zext` instruction.
func (irb IRBuilder) BuildZExt(src Value, dest Type) (in Instruction) {
	in.c = C.LLVMBuildZExt(irb.c, src.ptr(), dest.ptr(), nil)
	return
}

// BuildSExt builds a `sext` instruction.
func (irb IRBuilder) BuildSExt(src Value, dest Type) (in Instruction) {
	in.c = C.LLVMBuildSExt(irb.c, src.ptr(), dest.ptr(), nil)
	return
}

// BuildFPToUI builds a `fptoui` instruction.
func (irb IRBuilder) BuildFPToUI(src Value, dest Type) (in Instruction) {
	in.c = C.LLVMBuildFPToUI(irb.c, src.ptr(), dest.ptr(), nil)
	return
}

// BuildFPToSI builds a `fptosi` instruction.
func (irb IRBuilder) BuildFPToSI(src Value, dest Type) (in Instruction) {
	in.c = C.LLVMBuildFPToSI(irb.c, src.ptr(), dest.ptr(), nil)
	return
}

// BuildUIToFP builds a `uitofp` instruction.
func (irb IRBuilder) BuildUIToFP(src Value, dest Type) (in Instruction) {
	in.c = C.LLVMBuildUIToFP(irb.c, src.ptr(), dest.ptr(), nil)
	return
}

// BuildSIToFP builds a `sitofp` instruction.
func (irb IRBuilder) BuildSIToFP(src Value, dest Type) (in Instruction) {
	in.c = C.LLVMBuildSIToFP(irb.c, src.ptr(), dest.ptr(), nil)
	return
}

// BuildFPTrunc builds a `fptrunc` instruction.
func (irb IRBuilder) BuildFPTrunc(src Value, dest Type) (in Instruction) {
	in.c = C.LLVMBuildFPTrunc(irb.c, src.ptr(), dest.ptr(), nil)
	return
}

// BuildFPExt builds a `fpext` instruction.
func (irb IRBuilder) BuildFPExt(src Value, dest Type) (in Instruction) {
	in.c = C.LLVMBuildFPExt(irb.c, src.ptr(), dest.ptr(), nil)
	return
}

// BuildPtrToInt builds a `ptrtoint` instruction.
func (irb IRBuilder) BuildPtrToInt(src Value, dest Type) (in Instruction) {
	in.c = C.LLVMBuildPtrToInt(irb.c, src.ptr(), dest.ptr(), nil)
	return
}

// BuildIntToPtr builds a `inttoptr` instruction.
func (irb IRBuilder) BuildIntToPtr(src Value, dest Type) (in Instruction) {
	in.c = C.LLVMBuildIntToPtr(irb.c, src.ptr(), dest.ptr(), nil)
	return
}

// BuildBitCast builds a `bitcast` instruction.
func (irb IRBuilder) BuildBitCast(src Value, dest Type) (in Instruction) {
	in.c = C.LLVMBuildBitCast(irb.c, src.ptr(), dest.ptr(), nil)
	return
}

// BuildAddrSpaceCast builds a `addrspacecast` instruction.
func (irb IRBuilder) BuildAddrSpaceCast(src Value, dest Type) (in Instruction) {
	in.c = C.LLVMBuildAddrSpaceCast(irb.c, src.ptr(), dest.ptr(), nil)
	return
}

// BuildICmp builds an `icmp` instruction.
func (irb IRBuilder) BuildICmp(pred IntPredicate, lhs, rhs Value) (icmp ICmpInstruction) {
	icmp.c = C.LLVMBuildICmp(irb.c, (C.LLVMIntPredicate)(pred), lhs.ptr(), rhs.ptr(), nil)
	return
}

// BuildFCmp builds an `fcmp` instruction.
func (irb IRBuilder) BuildFCmp(pred RealPredicate, lhs, rhs Value) (fcmp FCmpInstruction) {
	fcmp.c = C.LLVMBuildFCmp(irb.c, (C.LLVMRealPredicate)(pred), lhs.ptr(), rhs.ptr(), nil)
	return
}

// BuildPhi builds a `phi` instruction.
func (irb IRBuilder) BuildPhi(incomingTyp Type) (phi PHINodeInstruction) {
	phi.c = C.LLVMBuildPhi(irb.c, incomingTyp.ptr(), nil)
	return
}

// BuildCall builds a `call` instruction.
func (irb IRBuilder) BuildCall(rtType Type, fn Value, args ...Value) (call CallInstruction) {
	var argsArrPtr *C.LLVMValueRef
	if len(args) > 0 {
		argsArr := make([]C.LLVMValueRef, len(args))
		for i, arg := range args {
			argsArr[i] = arg.ptr()
		}

		argsArrPtr = &argsArr[0]
	}

	call.c = C.LLVMBuildCall2(irb.c, rtType.ptr(), fn.ptr(), byref(argsArrPtr), (C.uint)(len(args)), nil)
	return
}
