package llvm

/*
#include <stdlib.h>

#include "llvm-c/Core.h"
*/
import "C"
import "unsafe"

// AttributeKind represents an LLVM attribute kind.
type AttributeKind C.uint

// Enumeration of the different attribute kinds.
const (
	AlwaysInlineAttrKind                    AttributeKind = 1
	ArgMemOnlyAttrKind                      AttributeKind = 2
	BuiltinAttrKind                         AttributeKind = 3
	ColdAttrKind                            AttributeKind = 4
	ConvergentAttrKind                      AttributeKind = 5
	DisableSanitizerInstrumentationAttrKind AttributeKind = 6
	HotAttrKind                             AttributeKind = 7
	ImmArgAttrKind                          AttributeKind = 8
	InRegAttrKind                           AttributeKind = 9
	InaccessibleMemOnlyAttrKind             AttributeKind = 10
	InaccessibleMemOrArgMemOnlyAttrKind     AttributeKind = 11
	InlineHintAttrKind                      AttributeKind = 12
	JumpTableAttrKind                       AttributeKind = 13
	MinSizeAttrKind                         AttributeKind = 14
	MustProgressAttrKind                    AttributeKind = 15
	NakedAttrKind                           AttributeKind = 16
	NestAttrKind                            AttributeKind = 17
	NoAliasAttrKind                         AttributeKind = 18
	NoBuiltinAttrKind                       AttributeKind = 19
	NoCallbackAttrKind                      AttributeKind = 20
	NoCaptureAttrKind                       AttributeKind = 21
	NoCfCheckAttrKind                       AttributeKind = 22
	NoDuplicateAttrKind                     AttributeKind = 23
	NoFreeAttrKind                          AttributeKind = 24
	NoImplicitFloatAttrKind                 AttributeKind = 25
	NoInlineAttrKind                        AttributeKind = 26
	NoMergeAttrKind                         AttributeKind = 27
	NoProfileAttrKind                       AttributeKind = 28
	NoRecurseAttrKind                       AttributeKind = 29
	NoRedZoneAttrKind                       AttributeKind = 30
	NoReturnAttrKind                        AttributeKind = 31
	NoSanitizeCoverageAttrKind              AttributeKind = 32
	NoSyncAttrKind                          AttributeKind = 33
	NoUndefAttrKind                         AttributeKind = 34
	NoUnwindAttrKind                        AttributeKind = 35
	NonLazyBindAttrKind                     AttributeKind = 36
	NonNullAttrKind                         AttributeKind = 37
	NullPointerIsValidAttrKind              AttributeKind = 38
	OptForFuzzingAttrKind                   AttributeKind = 39
	OptimizeForSizeAttrKind                 AttributeKind = 40
	OptimizeNoneAttrKind                    AttributeKind = 41
	ReadNoneAttrKind                        AttributeKind = 42
	ReadOnlyAttrKind                        AttributeKind = 43
	ReturnedAttrKind                        AttributeKind = 44
	ReturnsTwiceAttrKind                    AttributeKind = 45
	SExtAttrKind                            AttributeKind = 46
	SafeStackAttrKind                       AttributeKind = 47
	SanitizeAddressAttrKind                 AttributeKind = 48
	SanitizeHwAddressAttrKind               AttributeKind = 49
	SanitizeMemTagAttrKind                  AttributeKind = 50
	SanitizeMemoryAttrKind                  AttributeKind = 51
	SanitizeThreadAttrKind                  AttributeKind = 52
	ShadowCallStackAttrKind                 AttributeKind = 53
	SpeculatableAttrKind                    AttributeKind = 54
	SpeculativeLoadHardeningAttrKind        AttributeKind = 55
	StackProtectAttrKind                    AttributeKind = 56
	StackProtectReqAttrKind                 AttributeKind = 57
	StackProtectStrongAttrKind              AttributeKind = 58
	StrictFpAttrKind                        AttributeKind = 59
	SwiftAsyncAttrKind                      AttributeKind = 60
	SwiftErrorAttrKind                      AttributeKind = 61
	SwiftSelfAttrKind                       AttributeKind = 62
	UWTableAttrKind                         AttributeKind = 63
	WillReturnAttrKind                      AttributeKind = 64
	WriteOnlyAttrKind                       AttributeKind = 65
	ZExtAttrKind                            AttributeKind = 66
	LastEnumAttrAttrKind                    AttributeKind = 66
	FirstTypeAttrAttrKind                   AttributeKind = 67
	ByRefAttrKind                           AttributeKind = 67
	ByValAttrKind                           AttributeKind = 68
	ElementTypeAttrKind                     AttributeKind = 69
	InAllocaAttrKind                        AttributeKind = 70
	PreallocatedAttrKind                    AttributeKind = 71
	StructRetAttrKind                       AttributeKind = 72
	AlignmentAttrKind                       AttributeKind = 73
	AllocSizeAttrKind                       AttributeKind = 74
	DereferenceableAttrKind                 AttributeKind = 75
	DereferenceableOrNullAttrKind           AttributeKind = 76
	StackAlignmentAttrKind                  AttributeKind = 77
	VScaleRangeAttrKind                     AttributeKind = 78
)

// Attribute is an interface representing all possible LLVM attributes.
type Attribute interface {
	// ptr returns the internal LLVM object pointer to the attribute.
	ptr() C.LLVMAttributeRef

	// Variant returns the variant of the attribute.
	Variant() AttributeVariant
}

// AttributeVariant indicates what kind of attribute we are dealing with. It
// must be one of the enumerated attribute variants below.
type AttributeVariant int

// Enumeration of the different attribute variants.
const (
	EnumAttr AttributeVariant = iota
	TypeAttr
	StringAttr
	InvalidAttr
)

// A base type for all attributes.
type attrBase struct {
	c C.LLVMAttributeRef
}

func (ab attrBase) ptr() C.LLVMAttributeRef {
	return ab.c
}

func (ab attrBase) Variant() AttributeVariant {
	if C.LLVMIsEnumAttribute(ab.c) == 1 {
		return EnumAttr
	} else if C.LLVMIsTypeAttribute(ab.c) == 1 {
		return TypeAttr
	} else if C.LLVMIsStringAttribute(ab.c) == 1 {
		return StringAttr
	} else {
		return InvalidAttr
	}
}

// -----------------------------------------------------------------------------

// EnumAttribute represents an LLVM enum attribute.
type EnumAttribute struct {
	attrBase
}

// NewEnumAttribute creates a new enum attribute in the current context.
func (c Context) NewEnumAttribute(kind AttributeKind, value int) (ea EnumAttribute) {
	ea.c = C.LLVMCreateEnumAttribute(c.c, C.uint(kind), C.ulonglong(value))
	return
}

// Kind returns the attribute kind of the enum attribute.
func (ea EnumAttribute) Kind() AttributeKind {
	return AttributeKind(C.LLVMGetEnumAttributeKind(ea.c))
}

// Value returns the integer value of the enum attribute.
func (ea EnumAttribute) Value() int {
	return int(C.LLVMGetEnumAttributeValue(ea.c))
}

// -----------------------------------------------------------------------------

// TypeAttribute represents an LLVM type attribute.
type TypeAttribute struct {
	attrBase
}

// NewTypeAttribute creates a new type attribute in the current context.
func (c Context) NewTypeAttribute(kind AttributeKind, value Type) (ta TypeAttribute) {
	ta.c = C.LLVMCreateTypeAttribute(c.c, C.uint(kind), value.ptr())
	return
}

// Kind returns the attribute kind of the type attribute.
func (ta TypeAttribute) Kind() AttributeKind {
	return AttributeKind(C.LLVMGetEnumAttributeKind(ta.c))
}

// Value returns the type value of the type attribute.
func (ta TypeAttribute) Value() Type {
	return typeBase{c: C.LLVMGetTypeAttributeValue(ta.c)}
}

// -----------------------------------------------------------------------------

// StringAttribute represents an LLVM string attribute.
type StringAttribute struct {
	attrBase
}

// NewStringAttribute creates a new string attribute in the current context.
func (c Context) NewStringAttribute(kind, value string) (sa StringAttribute) {
	ckind := C.CString(kind)
	defer C.free(unsafe.Pointer(ckind))

	cvalue := C.CString(value)
	defer C.free(unsafe.Pointer(cvalue))

	sa.c = C.LLVMCreateStringAttribute(c.c, ckind, (C.uint)(len(kind)), cvalue, (C.uint)(len(value)))
	return
}

// Kind returns the kind of the string attribute.
func (sa StringAttribute) Kind() string {
	var strlen C.uint
	ckind := C.LLVMGetStringAttributeKind(sa.c, byref(&strlen))
	return C.GoStringN(ckind, (C.int)(strlen))
}

// Value returns the value of the string attribute.
func (sa StringAttribute) Value() string {
	var strlen C.uint
	cvalue := C.LLVMGetStringAttributeValue(sa.c, byref(&strlen))
	return C.GoStringN(cvalue, (C.int)(strlen))
}

// -----------------------------------------------------------------------------

// AttributeSet is used to access the attributes of an LLVM value with attributes.
type AttributeSet interface {
	// NumAttrs returns the number of attributes in the attribute set.
	NumAttrs() int

	// GetAttrByEnum gets an enum or type attribute by its enum kind.
	GetAttrByEnum(kind AttributeKind) (Attribute, bool)

	// GetAttrByString gets a string attribute by its string kind.
	GetAttrByString(kind string) (Attribute, bool)

	// SetAttr sets an attribute to `attr` if such an attribute exists.
	// Otherwise, `attr` is added to the attribute set as a new attribute.
	SetAttr(attr Attribute)

	// DelAttrByEnum deletes an enum or type attribute by its enum kind.
	DelAttrByEnum(kind AttributeKind)

	// DelAttrByString deletes a string attribute by its string kind.
	DelAttrByString(kind string)
}

// funcAttrSet is an attribute set for a function, parameter, or return value.
type funcAttrSet struct {
	fn  C.LLVMValueRef
	ndx C.uint
}

func (fas funcAttrSet) NumAttrs() int {
	return int(C.LLVMGetAttributeCountAtIndex(fas.fn, fas.ndx))
}

func (fas funcAttrSet) GetAttrByEnum(kind AttributeKind) (Attribute, bool) {
	attr := attrBase{c: C.LLVMGetEnumAttributeAtIndex(fas.fn, fas.ndx, (C.uint)(kind))}

	if attr.Variant() == InvalidAttr {
		return nil, false
	}

	return attr, true
}

func (fas funcAttrSet) GetAttrByString(kind string) (Attribute, bool) {
	ckind := C.CString(kind)
	defer C.free(unsafe.Pointer(ckind))

	attr := attrBase{c: C.LLVMGetStringAttributeAtIndex(fas.fn, fas.ndx, ckind, (C.uint)(len(kind)))}

	if attr.Variant() == InvalidAttr {
		return nil, false
	}

	return attr, true
}

func (fas funcAttrSet) SetAttr(attr Attribute) {
	C.LLVMAddAttributeAtIndex(fas.fn, fas.ndx, attr.ptr())
}

func (fas funcAttrSet) DelAttrByEnum(kind AttributeKind) {
	C.LLVMRemoveEnumAttributeAtIndex(fas.fn, fas.ndx, (C.uint)(kind))
}

func (fas funcAttrSet) DelAttrByString(kind string) {
	ckind := C.CString(kind)
	defer C.free(unsafe.Pointer(ckind))

	C.LLVMRemoveStringAttributeAtIndex(fas.fn, fas.ndx, ckind, (C.uint)(len(kind)))
}

// callSiteAttrSet is the set of attributes applied at a callsite.
type callSiteAttrSet struct {
	call C.LLVMValueRef
	ndx  C.uint
}

func (csas callSiteAttrSet) NumAttrs() int {
	return int(C.LLVMGetCallSiteAttributeCount(csas.call, csas.ndx))
}

func (csas callSiteAttrSet) GetAttrByEnum(kind AttributeKind) (Attribute, bool) {
	attrBase := attrBase{c: C.LLVMGetCallSiteEnumAttribute(csas.call, csas.ndx, (C.uint)(kind))}

	if attrBase.Variant() == InvalidAttr {
		return nil, false
	}

	return attrBase, true
}

func (csas callSiteAttrSet) GetAttrByString(kind string) (Attribute, bool) {
	ckind := C.CString(kind)
	defer C.free(unsafe.Pointer(ckind))

	attrBase := attrBase{c: C.LLVMGetCallSiteStringAttribute(csas.call, csas.ndx, ckind, (C.uint)(len(kind)))}

	if attrBase.Variant() == InvalidAttr {
		return nil, false
	}

	return attrBase, true
}

func (csas callSiteAttrSet) SetAttr(attr Attribute) {
	C.LLVMAddCallSiteAttribute(csas.call, csas.ndx, attr.ptr())
}

func (csas callSiteAttrSet) DelAttrByEnum(kind AttributeKind) {
	C.LLVMRemoveCallSiteEnumAttribute(csas.call, csas.ndx, (C.uint)(kind))
}

func (csas callSiteAttrSet) DelAttrByString(kind string) {
	ckind := C.CString(kind)
	defer C.free(unsafe.Pointer(ckind))

	C.LLVMRemoveCallSiteStringAttribute(csas.call, csas.ndx, ckind, (C.uint)(len(kind)))
}
