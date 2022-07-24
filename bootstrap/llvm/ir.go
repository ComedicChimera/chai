package llvm

/*
#include <stdlib.h>

#include "llvm-c/Core.h"
*/
import "C"

import (
	"fmt"
	"math"
	"unsafe"
)

// OpCode represents an LLVM instruction opcode.
type OpCode C.LLVMOpcode

// Enumeration of different LLVM opcodes.
const (
	RetOpCode            OpCode = C.LLVMRet
	BrOpCode             OpCode = C.LLVMBr
	SwitchOpCode         OpCode = C.LLVMSwitch
	IndirectBrOpCode     OpCode = C.LLVMIndirectBr
	InvokeOpCode         OpCode = C.LLVMInvoke
	UnreachableOpCode    OpCode = C.LLVMUnreachable
	CallBrOpCode         OpCode = C.LLVMCallBr
	FNegOpCode           OpCode = C.LLVMFNeg
	AddOpCode            OpCode = C.LLVMAdd
	FAddOpCode           OpCode = C.LLVMFAdd
	SubOpCode            OpCode = C.LLVMSub
	FSubOpCode           OpCode = C.LLVMFSub
	MulOpCode            OpCode = C.LLVMMul
	FMulOpCode           OpCode = C.LLVMFMul
	UDivOpCode           OpCode = C.LLVMUDiv
	SDivOpCode           OpCode = C.LLVMSDiv
	FDivOpCode           OpCode = C.LLVMFDiv
	URemOpCode           OpCode = C.LLVMURem
	SRemOpCode           OpCode = C.LLVMSRem
	FRemOpCode           OpCode = C.LLVMFRem
	ShlOpCode            OpCode = C.LLVMShl
	LShrOpCode           OpCode = C.LLVMLShr
	AShrOpCode           OpCode = C.LLVMAShr
	AndOpCode            OpCode = C.LLVMAnd
	OrOpCode             OpCode = C.LLVMOr
	XorOpCode            OpCode = C.LLVMXor
	AllocaOpCode         OpCode = C.LLVMAlloca
	LoadOpCode           OpCode = C.LLVMLoad
	StoreOpCode          OpCode = C.LLVMStore
	GetElementPtrOpCode  OpCode = C.LLVMGetElementPtr
	TruncOpCode          OpCode = C.LLVMTrunc
	ZExtOpCode           OpCode = C.LLVMZExt
	SExtOpCode           OpCode = C.LLVMSExt
	FPToUIOpCode         OpCode = C.LLVMFPToUI
	FPToSIOpCode         OpCode = C.LLVMFPToSI
	UIToFPOpCode         OpCode = C.LLVMUIToFP
	SIToFPOpCode         OpCode = C.LLVMSIToFP
	FPTruncOpCode        OpCode = C.LLVMFPTrunc
	FPExtOpCode          OpCode = C.LLVMFPExt
	PtrToIntOpCode       OpCode = C.LLVMPtrToInt
	IntToPtrOpCode       OpCode = C.LLVMIntToPtr
	BitCastOpCode        OpCode = C.LLVMBitCast
	AddrSpaceCastOpCode  OpCode = C.LLVMAddrSpaceCast
	ICmpOpCode           OpCode = C.LLVMICmp
	FCmpOpCode           OpCode = C.LLVMFCmp
	PHIOpCode            OpCode = C.LLVMPHI
	CallOpCode           OpCode = C.LLVMCall
	SelectOpCode         OpCode = C.LLVMSelect
	UserOp1OpCode        OpCode = C.LLVMUserOp1
	UserOp2OpCode        OpCode = C.LLVMUserOp2
	VAArgOpCode          OpCode = C.LLVMVAArg
	ExtractElementOpCode OpCode = C.LLVMExtractElement
	InsertElementOpCode  OpCode = C.LLVMInsertElement
	ShuffleVectorOpCode  OpCode = C.LLVMShuffleVector
	ExtractValueOpCode   OpCode = C.LLVMExtractValue
	InsertValueOpCode    OpCode = C.LLVMInsertValue
	FreezeOpCode         OpCode = C.LLVMFreeze
	FenceOpCode          OpCode = C.LLVMFence
	AtomicCmpXchgOpCode  OpCode = C.LLVMAtomicCmpXchg
	AtomicRMWOpCode      OpCode = C.LLVMAtomicRMW
	ResumeOpCode         OpCode = C.LLVMResume
	LandingPadOpCode     OpCode = C.LLVMLandingPad
	CleanupRetOpCode     OpCode = C.LLVMCleanupRet
	CatchRetOpCode       OpCode = C.LLVMCatchRet
	CatchPadOpCode       OpCode = C.LLVMCatchPad
	CleanupPadOpCode     OpCode = C.LLVMCleanupPad
	CatchSwitchOpCode    OpCode = C.LLVMCatchSwitch
)

// Instruction represents an LLVM instruction.
type Instruction struct {
	UserValue
}

// OpCode returns the LLVM op code of the instruction.
func (instr Instruction) OpCode() OpCode {
	return OpCode(C.LLVMGetInstructionOpcode(instr.c))
}

// IsTerminator returns whether the instruction is a terminator.
func (instr Instruction) IsTerminator() bool {
	return C.LLVMIsATerminatorInst(instr.c) != nil
}

// Parent returns the parent block of the instruction.
func (instr Instruction) Parent() BasicBlock {
	return BasicBlock{c: C.LLVMGetInstructionParent(instr.c)}
}

// -----------------------------------------------------------------------------

// IntPredicate represents the predicate of an `icmp` instruction.
type IntPredicate C.LLVMIntPredicate

// Enumeration of valid int predicates.
const (
	IntEQ  IntPredicate = C.LLVMIntEQ
	IntNE  IntPredicate = C.LLVMIntNE
	IntUGT IntPredicate = C.LLVMIntUGT
	IntUGE IntPredicate = C.LLVMIntUGE
	IntULT IntPredicate = C.LLVMIntULT
	IntULE IntPredicate = C.LLVMIntULE
	IntSGT IntPredicate = C.LLVMIntSGT
	IntSGE IntPredicate = C.LLVMIntSGE
	IntSLT IntPredicate = C.LLVMIntSLT
	IntSLE IntPredicate = C.LLVMIntSLE
)

// ICmpInstruction represents an `icmp` instruction.
type ICmpInstruction struct {
	Instruction
}

// Predicate returns the int predicate of the `icmp` instruction.
func (ici ICmpInstruction) Predicate() IntPredicate {
	return IntPredicate(C.LLVMGetICmpPredicate(ici.c))
}

// -----------------------------------------------------------------------------

// RealPredicate represents the predicate of an `fcmp` instruction.
type RealPredicate C.LLVMRealPredicate

// Enumeration of different real predicates.
const (
	RealFalse RealPredicate = C.LLVMRealPredicateFalse
	RealOEQ   RealPredicate = C.LLVMRealOEQ
	RealOGT   RealPredicate = C.LLVMRealOGT
	RealOGE   RealPredicate = C.LLVMRealOGE
	RealOLT   RealPredicate = C.LLVMRealOLT
	RealOLE   RealPredicate = C.LLVMRealOLE
	RealONE   RealPredicate = C.LLVMRealONE
	RealORD   RealPredicate = C.LLVMRealORD
	RealUNO   RealPredicate = C.LLVMRealUNO
	RealUEQ   RealPredicate = C.LLVMRealUEQ
	RealUGT   RealPredicate = C.LLVMRealUGT
	RealUGE   RealPredicate = C.LLVMRealUGE
	RealULT   RealPredicate = C.LLVMRealULT
	RealULE   RealPredicate = C.LLVMRealULE
	RealUNE   RealPredicate = C.LLVMRealUNE
	RealTrue  RealPredicate = C.LLVMRealPredicateTrue
)

// FCmpInstruction represents an LLVM `fcmp` instruction.
type FCmpInstruction struct {
	Instruction
}

// Predicate returns the real predicate of an `fcmp` instruction.
func (fci FCmpInstruction) Predicate() RealPredicate {
	return RealPredicate(C.LLVMGetFCmpPredicate(fci.c))
}

// -----------------------------------------------------------------------------

// CallInstruction represents an `call` instruction.
type CallInstruction struct {
	Instruction
}

// NumArgs returns the number of arguments passed to the call instruction.
func (ci CallInstruction) NumArgs() int {
	return int(C.LLVMGetNumArgOperands(ci.c))
}

// CallConv returns the calling convention of the `call` instruction.
func (ci CallInstruction) CallConv() CallConv {
	return CallConv(C.LLVMGetInstructionCallConv(ci.c))
}

// SetCallConv sets the calling convention of the `call` instruction to cc.
func (ci CallInstruction) SetCallConv(cc CallConv) {
	C.LLVMSetInstructionCallConv(ci.c, (C.uint)(cc))
}

// IsTailCall returns whether the `call` instruction is a tail call.
func (ci CallInstruction) IsTailCall() bool {
	return C.LLVMIsTailCall(ci.c) == 1
}

// SetTailCall sets whether the `call` instruction is a tail call.
func (ci CallInstruction) SetTailCall(tc bool) {
	C.LLVMSetTailCall(ci.c, llvmBool(tc))
}

// FuncAttrs returns the call site function attribute set.
func (ci CallInstruction) FuncAttrs() AttributeSet {
	return callSiteAttrSet{call: ci.c, ndx: math.MaxUint32}
}

// ReturnAttrs returns the call site return value attribute set.
func (ci CallInstruction) ReturnAttrs() AttributeSet {
	return callSiteAttrSet{call: ci.c, ndx: 0}
}

// ParamAttrs returns the call site parameter value attribute set of the
// parameter at ndx.
func (ci CallInstruction) ParamAttrs(ndx int) AttributeSet {
	if 0 <= ndx && ndx < ci.NumArgs() {
		return callSiteAttrSet{call: ci.c, ndx: C.uint(ndx + 1)}
	}

	panic("error: parameter call site attribute index out of bounds")
}

// -----------------------------------------------------------------------------

// Terminator represents a terminator instruction.
type Terminator struct {
	Instruction
}

// IsConditional returns whether this is a conditional terminator.
func (term Terminator) IsConditional() bool {
	return C.LLVMIsConditional(term.c) == 1
}

// NumSuccessors returns the number of successors of this terminator.
func (term Terminator) NumSuccessors() int {
	return int(C.LLVMGetNumSuccessors(term.c))
}

// GetSuccessor gets the successor of this terminator at ndx.
func (term Terminator) GetSuccessor(ndx int) BasicBlock {
	if 0 <= ndx && ndx < term.NumSuccessors() {
		return BasicBlock{c: C.LLVMGetSuccessor(term.c, (C.uint)(ndx))}
	}

	panic("error: successor index out of bounds")
}

// SetSuccessor sets the successor of this terminator at ndx.
func (term Terminator) SetSuccessor(ndx int, bb BasicBlock) {
	if 0 <= ndx && ndx < term.NumSuccessors() {
		C.LLVMSetSuccessor(term.c, (C.uint)(ndx), bb.c)
	}

	panic("error: successor index out of bounds")
}

// -----------------------------------------------------------------------------

// ConditionalTerminator represents a conditional terminator in LLVM.
type ConditionalTerminator struct {
	Terminator
}

// Condition returns the condition of this terminator.
func (ct ConditionalTerminator) Condition() Value {
	return valueBase{c: C.LLVMGetCondition(ct.c)}
}

// SetCondition sets the condition of this terminator.
func (ct ConditionalTerminator) SetCondition(cond Value) {
	C.LLVMSetCondition(ct.c, cond.ptr())
}

// -----------------------------------------------------------------------------

// SwitchInstruction represents an LLVM `switch` instruction.
type SwitchInstruction struct {
	ConditionalTerminator
}

// AddCase adds a new case to the `switch`` instruction.
func (si SwitchInstruction) AddCase(caseVal Value, caseBlock BasicBlock) {
	C.LLVMAddCase(si.c, caseVal.ptr(), caseBlock.c)
}

// Default returns the default case block of the `switch` instruction.
func (si SwitchInstruction) Default() BasicBlock {
	return BasicBlock{c: C.LLVMGetSwitchDefaultDest(si.c)}
}

// -----------------------------------------------------------------------------

// AllocaInstruction represents an LLVM `alloca` instruction.
type AllocaInstruction struct {
	Instruction
}

// AllocatedType returns the allocated type of the instruction.
func (ai AllocaInstruction) AllocatedType() Type {
	return typeBase{c: C.LLVMGetAllocatedType(ai.c)}
}

// -----------------------------------------------------------------------------

// GEPInstruction represents an LLVM `getelementptr` instruction.
type GEPInstruction struct {
	Instruction
}

// IsInBounds returns whether or not the GEP instruction is in bounds.
func (gep GEPInstruction) IsInBounds() bool {
	return C.LLVMIsInBounds(gep.c) == 1
}

// SetInBounds sets whether or not the GEP instruction is in bounds.
func (gep GEPInstruction) SetInBounds(inb bool) {
	C.LLVMSetIsInBounds(gep.c, llvmBool(inb))
}

// -----------------------------------------------------------------------------

// PHINodeInstruction represents a `phi` instruction.
type PHINodeInstruction struct {
	Instruction
}

// PHIIncoming represents an incoming in a `phi` instruction.
type PHIIncoming struct {
	// The incoming value.
	Value Value

	// The incoming block.
	Block BasicBlock
}

// AddIncoming adds new incoming to the PHI node.
func (pni PHINodeInstruction) AddIncoming(incoming ...PHIIncoming) {
	incomingValues := make([]C.LLVMValueRef, len(incoming))
	incomingBlocks := make([]C.LLVMBasicBlockRef, len(incoming))

	for i, inc := range incoming {
		incomingValues[i] = inc.Value.ptr()
		incomingBlocks[i] = inc.Block.c
	}

	C.LLVMAddIncoming(pni.c, byref(&incomingValues[0]), byref(&incomingBlocks[0]), (C.uint)(len(incoming)))
}

// NumIncoming returns the number of incoming to the PHI node.
func (pni PHINodeInstruction) NumIncoming() int {
	return int(C.LLVMCountIncoming(pni.c))
}

// GetIncoming returns the incoming to the PHI node at ndx.
func (pni PHINodeInstruction) GetIncoming(ndx int) (incoming PHIIncoming) {
	if 0 <= ndx && ndx < pni.NumIncoming() {
		incoming.Value = valueBase{c: C.LLVMGetIncomingValue(pni.c, (C.uint)(ndx))}
		incoming.Block = BasicBlock{c: C.LLVMGetIncomingBlock(pni.c, (C.uint)(ndx))}
		return
	}

	panic("error: incoming index out of bounds")
}

// -----------------------------------------------------------------------------

// BasicBlock represents an LLVM basic block.
type BasicBlock struct {
	c C.LLVMBasicBlockRef
}

// Name returns the name of the basic block.
func (bb BasicBlock) Name() string {
	return C.GoString(C.LLVMGetBasicBlockName(bb.c))
}

// Terminator returns the terminator instruction of a basic block.
func (bb BasicBlock) Terminator() (in Instruction, exists bool) {
	termPtr := C.LLVMGetBasicBlockTerminator(bb.c)

	if termPtr != nil {
		in.c = termPtr
		exists = true
	} else {
		exists = false
	}

	return
}

// Parent returns the parent function of the basic block.
func (bb BasicBlock) Parent() (fn Function) {
	fn.c = C.LLVMGetBasicBlockParent(bb.c)
	return
}

// instrIter is an iterator over the instruction of a basic block.
type instrIter struct {
	curr, next C.LLVMValueRef
}

func (it *instrIter) Item() (instr Instruction) {
	instr.c = it.curr
	return
}

func (it *instrIter) Next() bool {
	it.curr = it.next
	it.next = C.LLVMGetNextInstruction(it.curr)
	return it.curr != nil
}

// Instructions returns an iterator over the instructions of a basic block.
func (bb BasicBlock) Instructions() Iterator[Instruction] {
	return &instrIter{next: C.LLVMGetFirstInstruction(bb.c)}
}

// First returns the first instruction in a basic block.
func (bb BasicBlock) First() (instr Instruction, exists bool) {
	instrPtr := C.LLVMGetFirstInstruction(bb.c)

	if instrPtr == nil {
		exists = false
	} else {
		instr.c = instrPtr
		exists = true
	}

	return
}

// Last returns the last instruction in a basic block.
func (bb BasicBlock) Last() (instr Instruction, exists bool) {
	instrPtr := C.LLVMGetLastInstruction(bb.c)

	if instrPtr == nil {
		exists = false
	} else {
		instr.c = instrPtr
		exists = true
	}

	return
}

// -----------------------------------------------------------------------------

// CallConv represents an LLVM calling convention.
type CallConv C.LLVMCallConv

// Enumeration of different calling conventions.
const (
	CCallConv             CallConv = C.LLVMCCallConv
	FastCallConv          CallConv = C.LLVMFastCallConv
	ColdCallConv          CallConv = C.LLVMColdCallConv
	GHCCallConv           CallConv = C.LLVMGHCCallConv
	HiPECallConv          CallConv = C.LLVMHiPECallConv
	WebKitJSCallConv      CallConv = C.LLVMWebKitJSCallConv
	AnyRegCallConv        CallConv = C.LLVMAnyRegCallConv
	PreserveMostCallConv  CallConv = C.LLVMPreserveMostCallConv
	PreserveAllCallConv   CallConv = C.LLVMPreserveAllCallConv
	SwiftCallConv         CallConv = C.LLVMSwiftCallConv
	CXXFASTTLSCallConv    CallConv = C.LLVMCXXFASTTLSCallConv
	X86StdcallCallConv    CallConv = C.LLVMX86StdcallCallConv
	X86FastcallCallConv   CallConv = C.LLVMX86FastcallCallConv
	ARMAPCSCallConv       CallConv = C.LLVMARMAPCSCallConv
	ARMAAPCSCallConv      CallConv = C.LLVMARMAAPCSCallConv
	ARMAAPCSVFPCallConv   CallConv = C.LLVMARMAAPCSVFPCallConv
	MSP430INTRCallConv    CallConv = C.LLVMMSP430INTRCallConv
	X86ThisCallCallConv   CallConv = C.LLVMX86ThisCallCallConv
	PTXKernelCallConv     CallConv = C.LLVMPTXKernelCallConv
	PTXDeviceCallConv     CallConv = C.LLVMPTXDeviceCallConv
	SPIRFUNCCallConv      CallConv = C.LLVMSPIRFUNCCallConv
	SPIRKERNELCallConv    CallConv = C.LLVMSPIRKERNELCallConv
	IntelOCLBICallConv    CallConv = C.LLVMIntelOCLBICallConv
	X8664SysVCallConv     CallConv = C.LLVMX8664SysVCallConv
	Win64CallConv         CallConv = C.LLVMWin64CallConv
	X86VectorCallCallConv CallConv = C.LLVMX86VectorCallCallConv
	HHVMCallConv          CallConv = C.LLVMHHVMCallConv
	HHVMCCallConv         CallConv = C.LLVMHHVMCCallConv
	X86INTRCallConv       CallConv = C.LLVMX86INTRCallConv
	AVRINTRCallConv       CallConv = C.LLVMAVRINTRCallConv
	AVRSIGNALCallConv     CallConv = C.LLVMAVRSIGNALCallConv
	AVRBUILTINCallConv    CallConv = C.LLVMAVRBUILTINCallConv
	AMDGPUVSCallConv      CallConv = C.LLVMAMDGPUVSCallConv
	AMDGPUGSCallConv      CallConv = C.LLVMAMDGPUGSCallConv
	AMDGPUPSCallConv      CallConv = C.LLVMAMDGPUPSCallConv
	AMDGPUCSCallConv      CallConv = C.LLVMAMDGPUCSCallConv
	AMDGPUKERNELCallConv  CallConv = C.LLVMAMDGPUKERNELCallConv
	X86RegCallCallConv    CallConv = C.LLVMX86RegCallCallConv
	AMDGPUHSCallConv      CallConv = C.LLVMAMDGPUHSCallConv
	MSP430BUILTINCallConv CallConv = C.LLVMMSP430BUILTINCallConv
	AMDGPULSCallConv      CallConv = C.LLVMAMDGPULSCallConv
	AMDGPUESCallConv      CallConv = C.LLVMAMDGPUESCallConv
)

// Function represents an LLVM function.
type Function struct {
	GlobalValue

	// The parent module context of the function.
	mctx C.LLVMContextRef
}

// NumParams returns the number of parameters of the function.
func (f Function) NumParams() int {
	return int(C.LLVMCountParams(f.c))
}

// GetParam returns the function parameter at index ndx.
func (f Function) GetParam(ndx int) (fp FuncParam) {
	if 0 <= ndx && ndx < f.NumParams() {
		fp.c = C.LLVMGetParam(f.c, (C.uint)(ndx))
		fp.fa = funcAttrSet{fn: f.c, ndx: (C.uint)(ndx + 1)}
	} else {
		panic("error: parameter index out of bounds")
	}

	return
}

// Attrs returns the function attribute set.
func (f Function) Attrs() AttributeSet {
	return funcAttrSet{fn: f.c, ndx: math.MaxUint32}
}

// ReturnAttrs returns the return value attribute set.
func (f Function) ReturnAttrs() AttributeSet {
	return funcAttrSet{fn: f.c, ndx: C.LLVMAttributeReturnIndex}
}

// CallConv returns the calling convention of the function.
func (f Function) CallConv() CallConv {
	return CallConv(C.LLVMGetFunctionCallConv(f.c))
}

// SetCallConv sets the calling convention of the function to cc.
func (f Function) SetCallConv(cc CallConv) {
	C.LLVMSetFunctionCallConv(f.c, (C.uint)(cc))
}

// IntrinsicID returns the intrinsic ID of the function if it is intrinsic.
func (f Function) IntrinsicID() (id uint, isIntrinsic bool) {
	id = uint(C.LLVMGetIntrinsicID(f.c))
	isIntrinsic = id == 0
	return
}

// IntrinsicName returns the intrinsic name of the function if it is intrinsic.
func (f Function) IntrinsicName() (string, bool) {
	id := C.LLVMGetIntrinsicID(f.c)
	if id != 0 {
		var strlen C.size_t
		cname := C.LLVMIntrinsicGetName(id, byref(&strlen))
		return C.GoStringN(cname, (C.int)(strlen)), true
	}

	return "", false

}

// Body returns the body of the function.
func (f Function) Body() FuncBody {
	return FuncBody{mctx: f.mctx, c: f.c}
}

// -----------------------------------------------------------------------------

// FuncParam represents a function parameter.
type FuncParam struct {
	valueBase
	fa funcAttrSet
}

// Attrs returns the parameter attribute set.
func (fp FuncParam) Attrs() AttributeSet {
	return fp.fa
}

// -----------------------------------------------------------------------------

// FuncBody represents an LLVM function body.
type FuncBody struct {
	c           C.LLVMValueRef
	mctx        C.LLVMContextRef
	nameCounter int
}

// First returns the first basic block in the function body.
func (fb FuncBody) First() (bb BasicBlock, exists bool) {
	bbPtr := C.LLVMGetFirstBasicBlock(fb.c)

	if bbPtr == nil {
		exists = false
	} else {
		bb.c = bbPtr
		exists = true
	}

	return
}

// Last returns the last basic block in the function body.
func (fb FuncBody) Last() (bb BasicBlock, exists bool) {
	bbPtr := C.LLVMGetLastBasicBlock(fb.c)

	if bbPtr == nil {
		exists = false
	} else {
		bb.c = bbPtr
		exists = true
	}

	return
}

// Len returns the number of basic blocks in the function body.
func (fb FuncBody) Len() int {
	return int(C.LLVMCountBasicBlocks(fb.c))
}

// bodyIter is an iterator over the body of a function.
type bodyIter struct {
	curr, next C.LLVMBasicBlockRef
}

func (it *bodyIter) Item() (bb BasicBlock) {
	bb.c = it.curr
	return
}

func (it *bodyIter) Next() bool {
	it.curr = it.next
	it.next = C.LLVMGetNextBasicBlock(it.curr)
	return it.curr != nil
}

// Blocks returns an iterator over the blocks of a function body.
func (fb FuncBody) Blocks() Iterator[BasicBlock] {
	return &bodyIter{next: C.LLVMGetFirstBasicBlock(fb.c)}
}

// Append appends a new block to the function.
func (fb FuncBody) Append() (bb BasicBlock) {
	cname := C.CString(fmt.Sprintf("bb%d", fb.nameCounter))
	defer C.free(unsafe.Pointer(cname))
	fb.nameCounter++

	bb.c = C.LLVMAppendBasicBlockInContext(fb.mctx, fb.c, cname)
	return
}

// AppendNamed appends a new block to the function named name.
func (fb FuncBody) AppendNamed(name string) (bb BasicBlock) {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))

	bb.c = C.LLVMAppendBasicBlockInContext(fb.mctx, fb.c, cname)
	return
}

// Insert inserts a new block into the function before ibb.
func (fb FuncBody) Insert(ibb BasicBlock) (bb BasicBlock) {
	cname := C.CString(fmt.Sprintf("bb%d", fb.nameCounter))
	defer C.free(unsafe.Pointer(cname))
	fb.nameCounter++

	bb.c = C.LLVMInsertBasicBlockInContext(fb.mctx, ibb.c, cname)
	return
}

// InsertNamed inserts a new block into the function named name before ibb.
func (fb FuncBody) InsertNamed(ibb BasicBlock, name string) (bb BasicBlock) {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))

	bb.c = C.LLVMInsertBasicBlockInContext(fb.mctx, ibb.c, cname)
	return
}
