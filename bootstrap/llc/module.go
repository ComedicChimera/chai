package llc

/*
#include <stdlib.h>
#include "llvm-c/Core.h"
#include "llvm-c/Analysis.h"
#include "llvm-c/IRReader.h"
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

// NewModuleFromIR creates a new module using from the given string of LLVM IR
// in the given context.
func (ctx *Context) NewModuleFromIR(irString string) (*Module, error) {
	// Convert the IR string to a C string.
	cir := C.CString(irString)

	// Create a new LLVM memory buffer from the IR string.
	memBuff := C.LLVMCreateMemoryBufferWithMemoryRange(
		cir,
		(C.size_t)(len(irString)),
		nil,
		0,
	)
	defer C.LLVMDisposeMemoryBuffer(memBuff)

	// Try to parse the LLVM IR if possible.
	var modPtr C.LLVMModuleRef
	var msg *C.char
	if C.LLVMParseIRInContext(ctx.c, memBuff, byref(&modPtr), byref(&msg)) == 0 {
		return &Module{c: modPtr, mctx: ctx.c}, nil
	}

	defer C.LLVMDisposeMessage(msg)
	return nil, errors.New(C.GoString(msg))
}

/* -------------------------------------------------------------------------- */

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

	if C.LLVMPrintModuleToFile(m.c, cfpath, byref(&errMsg)) != 0 {
		defer C.LLVMDisposeMessage(errMsg)
		return errors.New(C.GoString(errMsg))
	}

	return nil
}

/* -------------------------------------------------------------------------- */

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

/* -------------------------------------------------------------------------- */

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
