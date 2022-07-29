package llc

/*
#include "llvm-c/Core.h"
#include "llvm-c/Initialization.h"
*/
import "C"
import "unsafe"

// OwnedObject represents an LLVM object that can be disposed.
type OwnedObject interface {
	// Dispose frees all the resources associated with the LLVM object.
	dispose()
}

// Context represents an LLVM context.
type Context struct {
	c C.LLVMContextRef

	// The list of LLVM objects owned by this context.
	ownedObjects []OwnedObject
}

// NewContext creates a new LLVM context.
func NewContext() *Context {
	return &Context{c: C.LLVMContextCreate()}
}

// The global context reference.
var globalCtx *Context

func GetGlobalContext() *Context {
	if globalCtx == nil {
		globalCtx = &Context{c: C.LLVMGetGlobalContext()}
	}

	return globalCtx
}

// takeOwnership marks the given disposable LLVM object as being owned by this
// context: this context is responsible for its disposal.
func (c *Context) takeOwnership(obj OwnedObject) {
	c.ownedObjects = append(c.ownedObjects, obj)
}

// Dispose frees all the resources associated with this context: the context
// itself and all the owned resources of this context.
func (c Context) Dispose() {
	for _, obj := range c.ownedObjects {
		obj.dispose()
	}

	C.LLVMContextDispose(c.c)
}

// -----------------------------------------------------------------------------

// Iterator represents an iterator of LLVM objects.  This is needed because many
// LLVM C API's don't expose a way to access elements by index but do allow you
// to iterate over them.  The pattern for using iterators is as follows:
//
//  for it := v.Items(); it.Next(); {
//  	item := it.Item()
//  	..
//  }
//
type Iterator[T any] interface {
	// Item returns the current item the iterator is positioned over if it
	// exists.  If the item does not exist, the return value is invalid.
	Item() T

	// Next moves the iterator forward one element if an element exists. It
	// returns whether or not it was able to move the iterator forward. Next
	// should be called to get the first element.
	Next() bool
}

// -----------------------------------------------------------------------------

// byref passes a Go value by reference to C.
func byref[T any](v *T) *T {
	return (*T)(unsafe.Pointer(v))
}

// llvmBool converts a boolean value to an LLVMBool.
func llvmBool(v bool) C.int {
	if v {
		return 1
	}

	return 0
}

// -----------------------------------------------------------------------------

func init() {
	// Initalize the various LLVM components that we need.
	pr := C.LLVMGetGlobalPassRegistry()
	C.LLVMInitializeCore(pr)
	C.LLVMInitializeAnalysis(pr)
	C.LLVMInitializeCodeGen(pr)
	C.LLVMInitializeTarget(pr)

	// Initialize all output targets.
	initializeAllTargets()
}
