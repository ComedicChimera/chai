package llvm

/*
#include <stdlib.h>

#include "llvm-c/Core.h"
#include "llvm-c/Analysis.h"
*/
import "C"

import (
	"errors"
	"unsafe"
)

// Module represents an LLVM module.
type Module struct {
	c    C.LLVMModuleRef
	mctx C.LLVMContextRef
}

// NewModule creates a new module with the given name in the current context.
func (c *Context) NewModule(name string) (m Module) {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))
	m.c = C.LLVMModuleCreateWithNameInContext(cname, c.c)
	c.takeOwnership(m)
	m.mctx = c.c
	return
}

// dispose disposes of the current module.
func (m Module) dispose() {
	C.LLVMDisposeModule(m.c)
}

// Dump prints the LLVM IR of the module to standard out.
func (m Module) Dump() {
	C.LLVMDumpModule(m.c)
}

// WriteToFile writes the LLVM IR of the module to a file.
func (m Module) WriteToFile(filepath string) error {
	var errMsg *C.char

	cfpath := C.CString(filepath)
	defer C.free(unsafe.Pointer(cfpath))

	if C.LLVMPrintModuleToFile(m.c, cfpath, byref(&errMsg)) == 1 {
		defer C.LLVMDisposeMessage(errMsg)
		return errors.New(C.GoString(errMsg))
	}

	return nil
}

// -----------------------------------------------------------------------------

// Name returns the name of the module.
func (m Module) Name() string {
	var strlen C.size_t
	str := C.LLVMGetModuleIdentifier(m.c, byref(&strlen))
	return C.GoStringN(str, (C.int)(strlen))
}

// SetName sets the name of the module.
func (m Module) SetName(name string) {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))
	C.LLVMSetModuleIdentifier(m.c, cname, (C.size_t)(len(name)))
}

// SourceFileName returns the source file name of the module.
func (m Module) SourceFileName() string {
	var strlen C.size_t
	cname := C.LLVMGetSourceFileName(m.c, byref(&strlen))
	return C.GoStringN(cname, (C.int)(strlen))
}

// SetSourceFileName sets the source file name of the module to fname.
func (m Module) SetSourceFileName(name string) {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))
	C.LLVMSetSourceFileName(m.c, cname, (C.size_t)(len(name)))
}

// DataLayout returns the data layout string of the module.
func (m Module) DataLayout() string {
	return C.GoString(C.LLVMGetDataLayout(m.c))
}

// SetDataLayout sets the data layout string of the module.
func (m Module) SetDataLayout(layout string) {
	clayout := C.CString(layout)
	defer C.free(unsafe.Pointer(clayout))
	C.LLVMSetDataLayout(m.c, clayout)
}

// TargetTriple returns the target triple string of the module.
func (m Module) TargetTriple() string {
	return C.GoString(C.LLVMGetTarget(m.c))
}

// SetTargetTriple sets the target triple string of the module.
func (m Module) SetTargetTriple(triple string) {
	ctriple := C.CString(triple)
	defer C.free(unsafe.Pointer(ctriple))
	C.LLVMSetTarget(m.c, ctriple)
}

// -----------------------------------------------------------------------------

// AddFunction adds a new function the module.
func (m Module) AddFunction(name string, funcType FunctionType) (fn Function) {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))

	fn.c = C.LLVMAddFunction(m.c, cname, funcType.c)
	fn.mctx = m.mctx
	return
}

// DeleteFunction deletes a function from the module.
func (m Module) DeleteFunction(fn Function) {
	C.LLVMDeleteFunction(fn.c)
}

// GetFunction returns the declared function corresponding to name.
func (m Module) GetFunction(name string) (fn Function, exists bool) {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))

	fnPtr := C.LLVMGetNamedFunction(m.c, cname)
	if fnPtr != nil {
		fn.c = fnPtr
		exists = true
	} else {
		exists = false
	}

	fn.mctx = m.mctx
	return
}

// GetIntrinsic returns the intrinsic function declaration corresponding to name.
func (m Module) GetIntrinsic(name string, overloadTypes ...Type) (fn Function, exists bool) {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))

	id := C.LLVMLookupIntrinsicID(cname, (C.size_t)(len(name)))

	if id == 0 {
		exists = false
		return
	}

	overloadTypeArr := make([]C.LLVMTypeRef, len(overloadTypes))
	for i, overloadType := range overloadTypes {
		overloadTypeArr[i] = overloadType.ptr()
	}

	fn.c = C.LLVMGetIntrinsicDeclaration(m.c, id, byref(&overloadTypeArr[0]), (C.size_t)(len(overloadTypes)))
	fn.mctx = m.mctx
	return
}

// funcIter is an iterator over the functions of a module.
type funcIter struct {
	mctx       C.LLVMContextRef
	curr, next C.LLVMValueRef
}

func (it *funcIter) Item() (fn Function) {
	fn.c = it.curr
	fn.mctx = it.mctx
	return
}

func (it *funcIter) Next() bool {
	it.curr = it.next
	it.next = C.LLVMGetNextFunction(it.curr)
	return it.curr != nil
}

// Functions returns an iterator of the functions of the module.
func (m Module) Functions() Iterator[Function] {
	return &funcIter{mctx: m.mctx, next: C.LLVMGetFirstFunction(m.c)}
}

// -----------------------------------------------------------------------------

// Verify verifies that the module is correct/well-formed.
func (m Module) Verify() error {
	var cmsg *C.char

	if C.LLVMVerifyModule(m.c, C.LLVMReturnStatusAction, byref(&cmsg)) == 1 {
		msg := C.GoString(cmsg)
		C.LLVMDisposeMessage(cmsg)

		return errors.New(msg)
	}

	return nil
}
