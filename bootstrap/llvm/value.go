package llvm

/*
#include <stdlib.h>

#include "llvm-c/Core.h"
*/
import "C"
import "unsafe"

// ValueKind represents a kind of LLVM value.
type ValueKind C.LLVMValueKind

// Enumeration of LLVM value kinds.
const (
	ArgumentValueKind              ValueKind = C.LLVMArgumentValueKind
	BasicBlockValueKind            ValueKind = C.LLVMBasicBlockValueKind
	MemoryUseValueKind             ValueKind = C.LLVMMemoryUseValueKind
	MemoryDefValueKind             ValueKind = C.LLVMMemoryDefValueKind
	MemoryPhiValueKind             ValueKind = C.LLVMMemoryPhiValueKind
	FunctionValueKind              ValueKind = C.LLVMFunctionValueKind
	GlobalAliasValueKind           ValueKind = C.LLVMGlobalAliasValueKind
	GlobalIFuncValueKind           ValueKind = C.LLVMGlobalIFuncValueKind
	GlobalVariableValueKind        ValueKind = C.LLVMGlobalVariableValueKind
	BlockAddressValueKind          ValueKind = C.LLVMBlockAddressValueKind
	ConstantExprValueKind          ValueKind = C.LLVMConstantExprValueKind
	ConstantArrayValueKind         ValueKind = C.LLVMConstantArrayValueKind
	ConstantStructValueKind        ValueKind = C.LLVMConstantStructValueKind
	ConstantVectorValueKind        ValueKind = C.LLVMConstantVectorValueKind
	UndefValueValueKind            ValueKind = C.LLVMUndefValueValueKind
	ConstantAggregateZeroValueKind ValueKind = C.LLVMConstantAggregateZeroValueKind
	ConstantDataArrayValueKind     ValueKind = C.LLVMConstantDataArrayValueKind
	ConstantDataVectorValueKind    ValueKind = C.LLVMConstantDataVectorValueKind
	ConstantIntValueKind           ValueKind = C.LLVMConstantIntValueKind
	ConstantFPValueKind            ValueKind = C.LLVMConstantFPValueKind
	ConstantPointerNullValueKind   ValueKind = C.LLVMConstantPointerNullValueKind
	ConstantTokenNoneValueKind     ValueKind = C.LLVMConstantTokenNoneValueKind
	MetadataAsValueValueKind       ValueKind = C.LLVMMetadataAsValueValueKind
	InlineAsmValueKind             ValueKind = C.LLVMInlineAsmValueKind
	InstructionValueKind           ValueKind = C.LLVMInstructionValueKind
	PoisonValueValueKind           ValueKind = C.LLVMPoisonValueValueKind
)

// Value is an interface used to represent all LLVM values.
type Value interface {
	// ptr returns the internal LLVM object pointer to the value.
	ptr() C.LLVMValueRef

	// Type returns the type of the LLVM value.
	Type() Type

	// Kind returns the kind of the LLVM value.
	Kind() ValueKind

	// Name returns the name of the value.
	Name() string

	// SetName sets the name of the value to name.
	SetName(name string)

	// IsConstant returns whether the value is constant.
	IsConstant() bool

	// IsUndef returns whether the value is `undef`.
	IsUndef() bool

	// IsPoison returns whether the value is `poison`.
	IsPoison() bool

	// Dump prints the value to standard out.
	Dump()

	// AsMetadata converts the value to metadata.
	// NB: Implementation for valueBase in metadata.go
	AsMetadata() Metadata
}

// valueBase is the base type for all values.
type valueBase struct {
	c C.LLVMValueRef
}

func (v valueBase) ptr() C.LLVMValueRef {
	return v.c
}

func (v valueBase) Type() Type {
	return typeBase{c: C.LLVMTypeOf(v.c)}
}

func (v valueBase) Kind() ValueKind {
	return ValueKind(C.LLVMGetValueKind(v.c))
}

func (v valueBase) Name() string {
	var strlen C.size_t
	return C.GoString(C.LLVMGetValueName2(v.c, byref(&strlen)))
}

func (v valueBase) SetName(name string) {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))
	C.LLVMSetValueName2(v.c, cname, (C.size_t)(len(name)))
}

func (v valueBase) IsConstant() bool {
	return C.LLVMIsConstant(v.c) == 1
}

func (v valueBase) IsUndef() bool {
	return C.LLVMIsUndef(v.c) == 1
}

func (v valueBase) IsPoison() bool {
	return C.LLVMIsPoison(v.c) == 1
}

func (v valueBase) Dump() {
	C.LLVMDumpValue(v.c)
}

// -----------------------------------------------------------------------------

// UserValue represents an LLVM value which has operands.
type UserValue struct {
	valueBase
}

// NumOperands returns the number of operands of the value.
func (uv UserValue) NumOperands() int {
	return int(C.LLVMGetNumOperands(uv.c))
}

// GetOperand retrieves the operand at index ndx.
func (uv UserValue) GetOperand(ndx int) Value {
	return valueBase{c: C.LLVMGetOperand(uv.c, (C.uint)(ndx))}
}

// SetOperand sets the operand at index ndx to val.
func (uv UserValue) SetOperand(ndx int, val Value) {
	C.LLVMSetOperand(uv.c, (C.uint)(ndx), val.ptr())
}

// -----------------------------------------------------------------------------

// Constant represents an LLVM constant value.
type Constant struct {
	valueBase
}

// ConstNull creates a new constant null value of type typ.
func ConstNull(typ Type) (c Constant) {
	c.c = C.LLVMConstNull(typ.ptr())
	return
}

// ConstNullptr creates a new constant null pointer with element type elemType.
func ConstNullptr(elemType Type) (c Constant) {
	c.c = C.LLVMConstPointerNull(elemType.ptr())
	return
}

// ConstInt creates a new integer constant of type intType, with value n, and
// signedness signed.
func ConstInt(intType IntegerType, n uint64, signed bool) (c Constant) {
	c.c = C.LLVMConstInt(intType.c, (C.ulonglong)(n), llvmBool(signed))
	return
}

// ConstReal creates a new real constant of type floatType with value n.
func ConstReal(floatType Type, n float64) (c Constant) {
	c.c = C.LLVMConstReal(floatType.ptr(), (C.double)(n))
	return
}

// IsNull returns whether or not the given constant is null.
func (c Constant) IsNull() bool {
	return C.LLVMIsNull(c.c) == 1
}

// -----------------------------------------------------------------------------

// Linkage represents an LLVM linkage.
type Linkage C.LLVMLinkage

// Enumeration of the different linkages.
const (
	ExternalLinkage            Linkage = C.LLVMExternalLinkage
	AvailableExternallyLinkage Linkage = C.LLVMAvailableExternallyLinkage
	LinkOnceAnyLinkage         Linkage = C.LLVMLinkOnceAnyLinkage
	LinkOnceODRLinkage         Linkage = C.LLVMLinkOnceODRLinkage
	WeakAnyLinkage             Linkage = C.LLVMWeakAnyLinkage
	WeakODRLinkage             Linkage = C.LLVMWeakODRLinkage
	AppendingLinkage           Linkage = C.LLVMAppendingLinkage
	InternalLinkage            Linkage = C.LLVMInternalLinkage
	PrivateLinkage             Linkage = C.LLVMPrivateLinkage
	ExternalWeakLinkage        Linkage = C.LLVMExternalWeakLinkage
	CommonLinkage              Linkage = C.LLVMCommonLinkage
)

// Visibility represents an LLVM visibility style.
type Visibility C.LLVMVisibility

// Enumeration of the different visibility styles
const (
	DefaultVisibility   Visibility = C.LLVMDefaultVisibility
	HiddenVisibility    Visibility = C.LLVMHiddenVisibility
	ProtectedVisibility Visibility = C.LLVMProtectedVisibility
)

// DLLStorageClass represents an LLVM DLL storage class.
type DLLStorageClass C.LLVMDLLStorageClass

// Enumeration of the different DLL storage classes.
const (
	DefaultStorageClass   DLLStorageClass = C.LLVMDefaultStorageClass
	DLLImportStorageClass DLLStorageClass = C.LLVMDLLImportStorageClass
	DLLExportStorageClass DLLStorageClass = C.LLVMDLLExportStorageClass
)

// UnnamedAddr represents an LLVM `unnamed_addr` attribute.
type UnnamedAddr C.LLVMUnnamedAddr

// Enumeration of different `unnamed_addr` attributes.
const (
	NoUnnamedAddr     UnnamedAddr = C.LLVMNoUnnamedAddr
	LocalUnnamedAddr  UnnamedAddr = C.LLVMLocalUnnamedAddr
	GlobalUnnamedAddr UnnamedAddr = C.LLVMGlobalUnnamedAddr
)

// GlobalValue represents an LLVM global value.
type GlobalValue struct {
	valueBase
}

// IsDeclaration returns whether the global value is a declaration.
func (gv GlobalValue) IsDeclaration() bool {
	return C.LLVMIsDeclaration(gv.c) == 1
}

// ValueType returns the value type of the global value.
func (gv GlobalValue) ValueType() Type {
	return typeBase{c: C.LLVMGlobalGetValueType(gv.c)}
}

// Linkage returns the linkage of the global value.
func (gv GlobalValue) Linkage() Linkage {
	return Linkage(C.LLVMGetLinkage(gv.c))
}

// SetLinkage sets the linkage of the global value to linkage.
func (gv GlobalValue) SetLinkage(linkage Linkage) {
	C.LLVMSetLinkage(gv.c, (C.LLVMLinkage)(linkage))
}

// Visibility returns the visibility style of the global value.
func (gv GlobalValue) Visibility() Visibility {
	return Visibility(C.LLVMGetVisibility(gv.c))
}

// SetVisibility sets the visibility of the global value to vis
func (gv GlobalValue) SetVisibility(vis Visibility) {
	C.LLVMSetVisibility(gv.c, (C.LLVMVisibility)(vis))
}

// DLLStorageClass returns the DLL storage class of the global value.
func (gv GlobalValue) DLLStorageClass() DLLStorageClass {
	return DLLStorageClass(C.LLVMGetDLLStorageClass(gv.c))
}

// SetDLLStorageClass sets the DLL storage class of the global value to dsc.
func (gv GlobalValue) SetDLLStorageClass(dsc DLLStorageClass) {
	C.LLVMSetDLLStorageClass(gv.c, (C.LLVMDLLStorageClass)(dsc))
}

// UnnamedAddr returns the `unnamed_addr` attribute of the global value.
func (gv GlobalValue) UnnamedAddr() UnnamedAddr {
	return UnnamedAddr(C.LLVMGetUnnamedAddress(gv.c))
}

// SetUnnamedAddr sets the `unnamed_addr` attribute of the global value to ua.
func (gv GlobalValue) SetUnnamedAddr(ua UnnamedAddr) {
	C.LLVMSetUnnamedAddress(gv.c, (C.LLVMUnnamedAddr)(ua))
}
