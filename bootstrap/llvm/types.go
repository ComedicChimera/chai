package llvm

/*
#include <stdlib.h>

#include "llvm-c/Core.h"
*/
import "C"

// TypeKind identifies a specific kind of LLVM type.
type TypeKind C.LLVMTypeKind

// Enumeration of different possible type kinds.
const (
	VoidTypeKind           TypeKind = C.LLVMVoidTypeKind
	FloatTypeKind          TypeKind = C.LLVMFloatTypeKind
	DoubleTypeKind         TypeKind = C.LLVMDoubleTypeKind
	X86_FP80TypeKind       TypeKind = C.LLVMX86_FP80TypeKind
	FP128TypeKind          TypeKind = C.LLVMFP128TypeKind
	PPC_FP128TypeKind      TypeKind = C.LLVMPPC_FP128TypeKind
	LabelTypeKind          TypeKind = C.LLVMLabelTypeKind
	IntegerTypeKind        TypeKind = C.LLVMIntegerTypeKind
	FunctionTypeKind       TypeKind = C.LLVMFunctionTypeKind
	StructTypeKind         TypeKind = C.LLVMStructTypeKind
	ArrayTypeKind          TypeKind = C.LLVMArrayTypeKind
	PointerTypeKind        TypeKind = C.LLVMPointerTypeKind
	MetadataTypeKind       TypeKind = C.LLVMMetadataTypeKind
	TokenTypeKind          TypeKind = C.LLVMTokenTypeKind
	VectorTypeKind         TypeKind = C.LLVMVectorTypeKind
	ScalableVectorTypeKind TypeKind = C.LLVMScalableVectorTypeKind
)

// Type is an interface used to represent all LLVM types.
type Type interface {
	// ptr returns the internal LLVM object pointer to the type.
	ptr() C.LLVMTypeRef

	// Kind returns the type's type kind.
	Kind() TypeKind

	// Sized returns whether or not the type is sized.
	Sized() bool

	// Dump prints the type to standard out.
	Dump()
}

// typeBase is the base struct used to build LLVM types.
type typeBase struct {
	c C.LLVMTypeRef
}

func (tb typeBase) ptr() C.LLVMTypeRef {
	return tb.c
}

func (tb typeBase) Kind() TypeKind {
	return TypeKind(C.LLVMGetTypeKind(tb.c))
}

func (tb typeBase) Sized() bool {
	return C.LLVMTypeIsSized(tb.c) == 1
}

func (tb typeBase) Dump() {
	C.LLVMDumpType(tb.c)
}

// -----------------------------------------------------------------------------

// IntegerType represents an LLVM integer type.
type IntegerType struct {
	typeBase
}

// Width returns the bit width of the integer type.
func (it IntegerType) BitWidth() uint {
	return uint(C.LLVMGetIntTypeWidth(it.c))
}

// Int1Type returns an new `i1` type in the context.
func (c Context) Int1Type() (it IntegerType) {
	it.c = C.LLVMInt1TypeInContext(c.c)
	return
}

// Int8Type returns an new `i8` type in the context.
func (c Context) Int8Type() (it IntegerType) {
	it.c = C.LLVMInt8TypeInContext(c.c)
	return
}

// Int16Type returns an new `i16` type in the context.
func (c Context) Int16Type() (it IntegerType) {
	it.c = C.LLVMInt16TypeInContext(c.c)
	return
}

// Int32Type returns an new `i32` type in the context.
func (c Context) Int32Type() (it IntegerType) {
	it.c = C.LLVMInt32TypeInContext(c.c)
	return
}

// Int1Type returns an new `i64` type in the context.
func (c Context) Int64Type() (it IntegerType) {
	it.c = C.LLVMInt64TypeInContext(c.c)
	return
}

// -----------------------------------------------------------------------------

// PointerType represents an LLVM pointer type.
type PointerType struct {
	typeBase
}

// NewPointerType returns a new pointer type to elemType in address space 0.
func NewPointerType(elemType Type) (pt PointerType) {
	pt.c = C.LLVMPointerType(elemType.ptr(), 0)
	return
}

// NewPointerTypeInAddrSpace creates a new pointer type to elemType in the
// address space addrSpace.
func NewPointerTypeInAddrSpace(elemType Type, addrSpace int) (pt PointerType) {
	pt.c = C.LLVMPointerType(elemType.ptr(), (C.uint)(addrSpace))
	return
}

// ElemType returns the element type of the pointer.
func (pt PointerType) ElemType() Type {
	return typeBase{c: C.LLVMGetElementType(pt.c)}
}

// AddrSpace returns the address space of the pointer.
func (pt PointerType) AddrSpace() int {
	return int(C.LLVMGetPointerAddressSpace(pt.c))
}

// -----------------------------------------------------------------------------

// FunctionType represents an LLVM function type.
type FunctionType struct {
	typeBase
}

// NewFunctionType returns a new function type with no variadic argument.
func NewFunctionType(returnType Type, paramTypes ...Type) (ft FunctionType) {
	var paramArrPtr *C.LLVMTypeRef
	if len(paramTypes) > 0 {
		paramArr := make([]C.LLVMTypeRef, len(paramTypes))
		for i, paramType := range paramTypes {
			paramArr[i] = paramType.ptr()
		}

		paramArrPtr = byref(&paramArr[0])
	}

	ft.c = C.LLVMFunctionType(returnType.ptr(), paramArrPtr, C.uint(len(paramTypes)), C.LLVMBool(0))
	return
}

// IsVarArg returns whether or not the function is variadic.
func (ft FunctionType) IsVarArg() bool {
	return C.LLVMIsFunctionVarArg(ft.c) == 1
}

// ReturnType returns the return type of the function.
func (ft FunctionType) ReturnType() Type {
	return typeBase{c: C.LLVMGetReturnType(ft.c)}
}

// NumParams returns the number of parameters of the function.
func (ft FunctionType) NumParams() uint {
	return uint(C.LLVMCountParamTypes(ft.c))
}

// Params returns the parameter types of the function.
func (ft FunctionType) Params() []Type {
	numParams := ft.NumParams()

	if numParams == 0 {
		return nil
	}

	paramArr := make([]C.LLVMTypeRef, numParams)
	paramArrPtr := byref(&paramArr[0])

	C.LLVMGetParamTypes(ft.c, paramArrPtr)

	params := make([]Type, numParams)
	for i, paramPtr := range paramArr {
		params[i] = typeBase{c: paramPtr}
	}

	return params
}

// -----------------------------------------------------------------------------

// FloatType returns a new LLVM `float` type.
func (c Context) FloatType() Type {
	return typeBase{c: C.LLVMFloatTypeInContext(c.c)}
}

// FloatType returns a new LLVM `double` type.
func (c Context) DoubleType() Type {
	return typeBase{c: C.LLVMDoubleTypeInContext(c.c)}
}

// VoidType returns a new LLVM `void` type.
func (c Context) VoidType() Type {
	return typeBase{c: C.LLVMVoidTypeInContext(c.c)}
}

// LabelType returns a new LLVM `label` type.
func (c Context) LabelType() Type {
	return typeBase{c: C.LLVMLabelTypeInContext(c.c)}
}

// MetadataType returns a new LLVM `float` type.
func (c Context) MetadataType() Type {
	return typeBase{c: C.LLVMMetadataTypeInContext(c.c)}
}
