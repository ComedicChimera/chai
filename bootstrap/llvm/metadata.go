package llvm

/*
#include <stdlib.h>

#include "llvm-c/Core.h"
#include "llvm-c/DebugInfo.h"
*/
import "C"
import "unsafe"

// MetadataKind represents the LLVM metadata kind.
type MetadataKind C.uint

// Enumeration of metadata kinds.
const (
	MDStringMetadataKind                     MetadataKind = C.LLVMMDStringMetadataKind
	ConstantAsMetadataMetadataKind           MetadataKind = C.LLVMConstantAsMetadataMetadataKind
	LocalAsMetadataMetadataKind              MetadataKind = C.LLVMLocalAsMetadataMetadataKind
	DistinctMDOperandPlaceholderMetadataKind MetadataKind = C.LLVMDistinctMDOperandPlaceholderMetadataKind
	MDTupleMetadataKind                      MetadataKind = C.LLVMMDTupleMetadataKind
	DILocationMetadataKind                   MetadataKind = C.LLVMDILocationMetadataKind
	DIExpressionMetadataKind                 MetadataKind = C.LLVMDIExpressionMetadataKind
	DIGlobalVariableExpressionMetadataKind   MetadataKind = C.LLVMDIGlobalVariableExpressionMetadataKind
	GenericDINodeMetadataKind                MetadataKind = C.LLVMGenericDINodeMetadataKind
	DISubrangeMetadataKind                   MetadataKind = C.LLVMDISubrangeMetadataKind
	DIEnumeratorMetadataKind                 MetadataKind = C.LLVMDIEnumeratorMetadataKind
	DIBasicTypeMetadataKind                  MetadataKind = C.LLVMDIBasicTypeMetadataKind
	DIDerivedTypeMetadataKind                MetadataKind = C.LLVMDIDerivedTypeMetadataKind
	DICompositeTypeMetadataKind              MetadataKind = C.LLVMDICompositeTypeMetadataKind
	DISubroutineTypeMetadataKind             MetadataKind = C.LLVMDISubroutineTypeMetadataKind
	DIFileMetadataKind                       MetadataKind = C.LLVMDIFileMetadataKind
	DICompileUnitMetadataKind                MetadataKind = C.LLVMDICompileUnitMetadataKind
	DISubprogramMetadataKind                 MetadataKind = C.LLVMDISubprogramMetadataKind
	DILexicalBlockMetadataKind               MetadataKind = C.LLVMDILexicalBlockMetadataKind
	DILexicalBlockFileMetadataKind           MetadataKind = C.LLVMDILexicalBlockFileMetadataKind
	DINamespaceMetadataKind                  MetadataKind = C.LLVMDINamespaceMetadataKind
	DIModuleMetadataKind                     MetadataKind = C.LLVMDIModuleMetadataKind
	DITemplateTypeParameterMetadataKind      MetadataKind = C.LLVMDITemplateTypeParameterMetadataKind
	DITemplateValueParameterMetadataKind     MetadataKind = C.LLVMDITemplateValueParameterMetadataKind
	DIGlobalVariableMetadataKind             MetadataKind = C.LLVMDIGlobalVariableMetadataKind
	DILocalVariableMetadataKind              MetadataKind = C.LLVMDILocalVariableMetadataKind
	DILabelMetadataKind                      MetadataKind = C.LLVMDILabelMetadataKind
	DIObjCPropertyMetadataKind               MetadataKind = C.LLVMDIObjCPropertyMetadataKind
	DIImportedEntityMetadataKind             MetadataKind = C.LLVMDIImportedEntityMetadataKind
	DIMacroMetadataKind                      MetadataKind = C.LLVMDIMacroMetadataKind
	DIMacroFileMetadataKind                  MetadataKind = C.LLVMDIMacroFileMetadataKind
	DICommonBlockMetadataKind                MetadataKind = C.LLVMDICommonBlockMetadataKind
	DIStringTypeMetadataKind                 MetadataKind = C.LLVMDIStringTypeMetadataKind
	DIGenericSubrangeMetadataKind            MetadataKind = C.LLVMDIGenericSubrangeMetadataKind
)

// Metadata represents LLVM metadata.
type Metadata interface {
	// ptr returns the internal LLVM object pointer.
	ptr() C.LLVMMetadataRef

	// Kind returns the kind of the metadata.
	Kind() MetadataKind

	// IsNode returns whether the metadata is an MDNode.
	IsNode() bool

	// IsString returns whether the metadata is an MDString.
	IsString() bool

	// AsValue converts the metadata into a value.
	AsValue() Value
}

// metaBase the base struct for all metadata classes.
type metaBase struct {
	c    C.LLVMMetadataRef
	vptr C.LLVMValueRef
}

func (mb metaBase) ptr() C.LLVMMetadataRef {
	return mb.c
}

func (mb metaBase) Kind() MetadataKind {
	return MetadataKind(C.LLVMGetMetadataKind(mb.c))
}

func (mb metaBase) IsNode() bool {
	return C.LLVMIsAMDNode(mb.vptr) != nil
}

func (mb metaBase) IsString() bool {
	return C.LLVMIsAMDString(mb.vptr) != nil
}

func (mb metaBase) AsValue() Value {
	return valueBase{c: mb.vptr}
}

func (vb valueBase) AsMetadata() Metadata {
	return metaBase{c: C.LLVMValueAsMetadata(vb.c), vptr: vb.c}
}

// -----------------------------------------------------------------------------

// MDString represents a LLVM MD string.
type MDString struct {
	metaBase
}

// NewMDString creates a new MD string of str in the context.
func (c Context) NewMDString(str string) (mds MDString) {
	cstr := C.CString(str)
	defer C.free(unsafe.Pointer(cstr))

	mds.c = C.LLVMMDStringInContext2(c.c, cstr, (C.size_t)(len(str)))
	mds.vptr = C.LLVMMetadataAsValue(c.c, mds.c)
	return
}

// String returns the string of the MD string.
func (mds MDString) String() string {
	var strlen C.uint
	cstr := C.LLVMGetMDString(mds.vptr, byref(&strlen))
	return C.GoStringN(cstr, (C.int)(strlen))
}

// -----------------------------------------------------------------------------

// MDNode represents an LLVM MD node.
type MDNode struct {
	metaBase
}

// NewMDNode creates a new MD node containing elements in the context.
func (c Context) NewMDNode(elems ...Metadata) (mdn MDNode) {
	elemArr := make([]C.LLVMMetadataRef, len(elems))
	for i, elem := range elems {
		elemArr[i] = elem.ptr()
	}

	mdn.c = C.LLVMMDNodeInContext2(c.c, byref(&elemArr[0]), (C.size_t)(len(elems)))
	mdn.vptr = C.LLVMMetadataAsValue(c.c, mdn.c)
	return
}

// Elements returns a slice containing the elements of the MD node.
func (mdn MDNode) Elements() []Metadata {
	metaArr := make([]C.LLVMValueRef, C.LLVMGetMDNodeNumOperands(mdn.vptr))

	C.LLVMGetMDNodeOperands(mdn.vptr, byref(&metaArr[0]))

	metas := make([]Metadata, len(metaArr))
	for i, metaValuePtr := range metaArr {
		metas[i] = metaBase{c: C.LLVMValueAsMetadata(metaValuePtr), vptr: metaValuePtr}
	}

	return metas
}
