package llvm

/*
#include <stdlib.h>

#include "llvm-c/Core.h"
#include "llvm-c/Target.h"
#include "llvm-c/TargetMachine.h"
*/
import "C"
import (
	"errors"
	"unsafe"
)

// ByteOrdering represents an LLVM byte ordering.
type ByteOrdering C.enum_LLVMByteOrdering

// // Enumeration of LLVM byte orderings.
const (
	BigEndian    ByteOrdering = C.LLVMBigEndian
	LittleEndian ByteOrdering = C.LLVMLittleEndian
)

// // CodeGenOptLevel represents an LLVm code generation optimization level.
type CodeGenOptLevel C.LLVMCodeGenOptLevel

// // Enumeration of LLVM codegen optimization levels.
const (
	CodeGenLevelNone       CodeGenOptLevel = C.LLVMCodeGenLevelNone
	CodeGenLevelLess       CodeGenOptLevel = C.LLVMCodeGenLevelLess
	CodeGenLevelDefault    CodeGenOptLevel = C.LLVMCodeGenLevelDefault
	CodeGenLevelAggressive CodeGenOptLevel = C.LLVMCodeGenLevelAggressive
)

// // CodeModel represents an LLVM code model.
type CodeModel C.LLVMCodeModel

// // Enumeration of LLVM code models.
const (
	CodeModelDefault    CodeModel = C.LLVMCodeModelDefault
	CodeModelJITDefault CodeModel = C.LLVMCodeModelJITDefault
	CodeModelTiny       CodeModel = C.LLVMCodeModelTiny
	CodeModelSmall      CodeModel = C.LLVMCodeModelSmall
	CodeModelKernel     CodeModel = C.LLVMCodeModelKernel
	CodeModelMedium     CodeModel = C.LLVMCodeModelMedium
	CodeModelLarge      CodeModel = C.LLVMCodeModelLarge
)

// // RelocMode represents an LLVM relocation mode.
type RelocMode C.LLVMRelocMode

// // Enumeration of LLVM relocation modes.
const (
	RelocDefault      RelocMode = C.LLVMRelocDefault
	RelocStatic       RelocMode = C.LLVMRelocStatic
	RelocPIC          RelocMode = C.LLVMRelocPIC
	RelocDynamicNoPic RelocMode = C.LLVMRelocDynamicNoPic
	RelocROPI         RelocMode = C.LLVMRelocROPI
	RelocRWPI         RelocMode = C.LLVMRelocRWPI
	RelocROPI_RWPI    RelocMode = C.LLVMRelocROPI_RWPI
)

// // CodeGenFileType represents a possible code generation output type.
type CodeGenFileType C.LLVMCodeGenFileType

// // Enumeration of LLVM codegen file types.
const (
	AssemblyFile CodeGenFileType = C.LLVMAssemblyFile
	ObjectFile   CodeGenFileType = C.LLVMObjectFile
)

// -----------------------------------------------------------------------------

// TargetData represents an LLVM target data layout.
type TargetData struct {
	c C.LLVMTargetDataRef
}

// NewTargetData creates a new target data from the data layout string layout.
func (c Context) NewTargetData(layout string) (td TargetData) {
	clayout := C.CString(layout)
	defer C.free(unsafe.Pointer(clayout))

	td.c = C.LLVMCreateTargetData(clayout)
	c.takeOwnership(td)
	return
}

// TargetData returns the target data layout of a module.
func (m Module) TargetData() (td TargetData) {
	td.c = C.LLVMGetModuleDataLayout(m.c)
	return
}

// SetTargetData sets the target data layout of a module.
func (m Module) SetTargetData(td TargetData) {
	C.LLVMSetModuleDataLayout(m.c, td.c)
}

// dispose disposes of the target data.
func (td TargetData) dispose() {
	C.LLVMDisposeTargetData(td.c)
}

// ByteOrder returns the byte ordering from the target data.
func (td TargetData) ByteOrder() ByteOrdering {
	return ByteOrdering(C.LLVMByteOrder(td.c))
}

// PointerSize returns the pointer size in bytes from the target data.
func (td TargetData) PointerSize() uint {
	return uint(C.LLVMPointerSize(td.c))
}

// IntPtr returns the integer pointer type for target in the context.
func (c Context) IntPtr(td TargetData) (it IntegerType) {
	it.c = C.LLVMIntPtrTypeInContext(c.c, td.c)
	return
}

// BitSizeOf returns the size of typ in bits on the target.
func (td TargetData) BitSizeOf(typ Type) uint {
	return uint(C.LLVMSizeOfTypeInBits(td.c, typ.ptr()))
}

// StorageSizeOf returns the storage size of typ in bytes on the target: the
// maximum number of bytes that may be overwritten by storing typ.
func (td TargetData) StorageSizeOf(typ Type) uint {
	return uint(C.LLVMStoreSizeOfType(td.c, typ.ptr()))
}

// ABISizeOf returns the ABI size of typ in bytes on the target: the offset
// in bytes between successive objects of the typ, including alignment padding.
func (td TargetData) ABISizeOf(typ Type) uint {
	return uint(C.LLVMABISizeOfType(td.c, typ.ptr()))
}

// ABIAlignOf returns the minimum ABI-required alignment of typ on the target.
func (td TargetData) ABIAlignOf(typ Type) uint {
	return uint(C.LLVMABIAlignmentOfType(td.c, typ.ptr()))
}

// PreferredAlignOf returns the preferred alignment of typ on the target.
func (td TargetData) PreferredAlignOf(typ Type) uint {
	return uint(C.LLVMPreferredAlignmentOfType(td.c, typ.ptr()))
}

// -----------------------------------------------------------------------------

// Target represents an LLVM output target.
type Target struct {
	c C.LLVMTargetRef
}

// HostTriple returns the target triple of the host system.
func HostTriple() string {
	ctriple := C.LLVMGetDefaultTargetTriple()
	defer C.LLVMDisposeMessage(ctriple)
	return C.GoString(ctriple)
}

// GetTargetFromName finds the target corresponding to name.
func GetTargetFromName(name string) (Target, bool) {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))

	targetPtr := C.LLVMGetTargetFromName(cname)
	if targetPtr == nil {
		return Target{}, false
	}

	return Target{c: targetPtr}, true
}

// GetTargetFromTriple finds the target corresponding to triple.
func GetTargetFromTriple(triple string) (Target, bool) {
	ctriple := C.CString(triple)
	defer C.free(unsafe.Pointer(ctriple))

	var targetPtr C.LLVMTargetRef
	if C.LLVMGetTargetFromTriple(ctriple, byref(&targetPtr), nil) == 0 {
		return Target{c: targetPtr}, true
	}

	return Target{}, false
}

// Name returns the name of the target.
func (t Target) Name() string {
	return C.GoString(C.LLVMGetTargetName(t.c))
}

// Description returns the description of the target.
func (t Target) Description() string {
	return C.GoString(C.LLVMGetTargetDescription(t.c))
}

// HasJIT returns if the target has a JIT.
func (t Target) HasJit() bool {
	return C.LLVMTargetHasJIT(t.c) == 1
}

// HasMachine returns if the target has a machine.
func (t Target) HasMachine() bool {
	return C.LLVMTargetHasTargetMachine(t.c) == 1
}

// HasASMBackend returns if the target has an ASM backend.
func (t Target) HasASMBackend() bool {
	return C.LLVMTargetHasAsmBackend(t.c) == 1
}

// -----------------------------------------------------------------------------

// TargetMachine represents an LLVM target machine: used to generate output.
type TargetMachine struct {
	c C.LLVMTargetMachineRef
}

// NewMachine creates a new target machine for target.
func (c Context) NewMachine(
	target Target,
	triple, cpu, features string,
	level CodeGenOptLevel,
	reloc RelocMode,
	model CodeModel,
) (tm TargetMachine) {
	ctriple := C.CString(triple)
	defer C.free(unsafe.Pointer(ctriple))

	ccpu := C.CString(cpu)
	defer C.free(unsafe.Pointer(ccpu))

	cfeatures := C.CString(features)
	defer C.free(unsafe.Pointer(cfeatures))

	tm.c = C.LLVMCreateTargetMachine(
		target.c,
		ctriple,
		ccpu,
		cfeatures,
		(C.LLVMCodeGenOptLevel)(level),
		(C.LLVMRelocMode)(reloc),
		(C.LLVMCodeModel)(model),
	)
	c.takeOwnership(tm)
	return
}

// NewHostMachine creates a new target machine for the host system for target.
func (c Context) NewHostMachine(target Target, level CodeGenOptLevel, reloc RelocMode, model CodeModel) (tm TargetMachine) {
	ctriple := C.LLVMGetDefaultTargetTriple()
	defer C.LLVMDisposeMessage(ctriple)

	ccpu := C.LLVMGetHostCPUName()
	defer C.LLVMDisposeMessage(ccpu)

	cfeatures := C.LLVMGetHostCPUFeatures()
	defer C.LLVMDisposeMessage(cfeatures)

	tm.c = C.LLVMCreateTargetMachine(
		target.c,
		ctriple,
		ccpu,
		cfeatures,
		(C.LLVMCodeGenOptLevel)(level),
		(C.LLVMRelocMode)(reloc),
		(C.LLVMCodeModel)(model),
	)
	c.takeOwnership(tm)
	return
}

// dispose disposes of target machine.
func (tm TargetMachine) dispose() {
	C.LLVMDisposeTargetMachine(tm.c)
}

// Target returns the target associated with the target machine.
func (tm TargetMachine) Target() (t Target) {
	t.c = C.LLVMGetTargetMachineTarget(tm.c)
	return
}

// Triple returns the target triple of the target machine.
func (tm TargetMachine) Triple() string {
	ctriple := C.LLVMGetTargetMachineTriple(tm.c)
	defer C.LLVMDisposeMessage(ctriple)
	return C.GoString(ctriple)
}

// CPU returns the CPU of the target machine.
func (tm TargetMachine) CPU() string {
	ccpu := C.LLVMGetTargetMachineCPU(tm.c)
	defer C.LLVMDisposeMessage(ccpu)
	return C.GoString(ccpu)
}

// Features returns the feature string of the target machine.
func (tm TargetMachine) Features() string {
	cfeatures := C.LLVMGetTargetMachineFeatureString(tm.c)
	defer C.LLVMDisposeMessage(cfeatures)
	return C.GoString(cfeatures)
}

// DataLayout returns the target data layout of the target machine.
func (tm TargetMachine) DataLayout() (td TargetData) {
	td.c = C.LLVMCreateTargetDataLayout(tm.c)
	return
}

// SetASMVerbosity sets the ASM verbosity of the target machine.
func (tm TargetMachine) SetASMVerbosity(verbose bool) {
	C.LLVMSetTargetMachineAsmVerbosity(tm.c, llvmBool(verbose))
}

// CompileModule compiles mod to fileType and outputs it to path.
func (tm TargetMachine) CompileModule(mod Module, path string, fileType CodeGenFileType) error {
	var cerr *C.char

	cpath := C.CString(path)
	defer C.free(unsafe.Pointer(cpath))

	if C.LLVMTargetMachineEmitToFile(tm.c, mod.c, cpath, (C.LLVMCodeGenFileType)(fileType), byref(&cerr)) == 0 {
		return nil
	}

	err := errors.New(C.GoString(cerr))
	defer C.LLVMDisposeMessage(cerr)

	return err
}

// -----------------------------------------------------------------------------

// initializeAllTargets initializes all targets LLVM is configured to support.
func initializeAllTargets() {
	// Initialize targets.
	C.LLVMInitializeAArch64Target()
	C.LLVMInitializeAMDGPUTarget()
	C.LLVMInitializeARMTarget()
	C.LLVMInitializeAVRTarget()
	C.LLVMInitializeBPFTarget()
	C.LLVMInitializeHexagonTarget()
	C.LLVMInitializeLanaiTarget()
	C.LLVMInitializeMSP430Target()
	C.LLVMInitializeMipsTarget()
	C.LLVMInitializeNVPTXTarget()
	C.LLVMInitializePowerPCTarget()
	C.LLVMInitializeRISCVTarget()
	C.LLVMInitializeSparcTarget()
	C.LLVMInitializeSystemZTarget()
	C.LLVMInitializeWebAssemblyTarget()
	C.LLVMInitializeX86Target()
	C.LLVMInitializeXCoreTarget()

	// Initialize target infos.
	C.LLVMInitializeAArch64TargetInfo()
	C.LLVMInitializeAMDGPUTargetInfo()
	C.LLVMInitializeARMTargetInfo()
	C.LLVMInitializeAVRTargetInfo()
	C.LLVMInitializeBPFTargetInfo()
	C.LLVMInitializeHexagonTargetInfo()
	C.LLVMInitializeLanaiTargetInfo()
	C.LLVMInitializeMSP430TargetInfo()
	C.LLVMInitializeMipsTargetInfo()
	C.LLVMInitializeNVPTXTargetInfo()
	C.LLVMInitializePowerPCTargetInfo()
	C.LLVMInitializeRISCVTargetInfo()
	C.LLVMInitializeSparcTargetInfo()
	C.LLVMInitializeSystemZTargetInfo()
	C.LLVMInitializeWebAssemblyTargetInfo()
	C.LLVMInitializeX86TargetInfo()
	C.LLVMInitializeXCoreTargetInfo()

	// Initialize target MCs.
	C.LLVMInitializeAArch64TargetMC()
	C.LLVMInitializeAMDGPUTargetMC()
	C.LLVMInitializeARMTargetMC()
	C.LLVMInitializeAVRTargetMC()
	C.LLVMInitializeBPFTargetMC()
	C.LLVMInitializeHexagonTargetMC()
	C.LLVMInitializeLanaiTargetMC()
	C.LLVMInitializeMSP430TargetMC()
	C.LLVMInitializeMipsTargetMC()
	C.LLVMInitializeNVPTXTargetMC()
	C.LLVMInitializePowerPCTargetMC()
	C.LLVMInitializeRISCVTargetMC()
	C.LLVMInitializeSparcTargetMC()
	C.LLVMInitializeSystemZTargetMC()
	C.LLVMInitializeWebAssemblyTargetMC()
	C.LLVMInitializeX86TargetMC()
	C.LLVMInitializeXCoreTargetMC()

	// Initiailize ASM printers.
	C.LLVMInitializeAArch64AsmPrinter()
	C.LLVMInitializeAMDGPUAsmPrinter()
	C.LLVMInitializeARMAsmPrinter()
	C.LLVMInitializeAVRAsmPrinter()
	C.LLVMInitializeBPFAsmPrinter()
	C.LLVMInitializeHexagonAsmPrinter()
	C.LLVMInitializeLanaiAsmPrinter()
	C.LLVMInitializeMSP430AsmPrinter()
	C.LLVMInitializeMipsAsmPrinter()
	C.LLVMInitializeNVPTXAsmPrinter()
	C.LLVMInitializePowerPCAsmPrinter()
	C.LLVMInitializeRISCVAsmPrinter()
	C.LLVMInitializeSparcAsmPrinter()
	C.LLVMInitializeSystemZAsmPrinter()
	C.LLVMInitializeWebAssemblyAsmPrinter()
	C.LLVMInitializeX86AsmPrinter()
	C.LLVMInitializeXCoreAsmPrinter()
}
