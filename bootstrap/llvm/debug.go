package llvm

/*
#include <stdlib.h>

#include "llvm-c/Core.h"
#include "llvm-c/DebugInfo.h"
*/
import "C"
import "unsafe"

// DIFlags represents a set of LLVM debug info flags.
type DIFlags C.LLVMDIFlags

// Enumeration of LLVM debug info flags.
const (
	DIFlagZero                DIFlags = C.LLVMDIFlagZero
	DIFlagPrivate             DIFlags = C.LLVMDIFlagPrivate
	DIFlagProtected           DIFlags = C.LLVMDIFlagProtected
	DIFlagPublic              DIFlags = C.LLVMDIFlagPublic
	DIFlagFwdDecl             DIFlags = C.LLVMDIFlagFwdDecl
	DIFlagAppleBlock          DIFlags = C.LLVMDIFlagAppleBlock
	DIFlagReservedBit4        DIFlags = C.LLVMDIFlagReservedBit4
	DIFlagVirtual             DIFlags = C.LLVMDIFlagVirtual
	DIFlagArtificial          DIFlags = C.LLVMDIFlagArtificial
	DIFlagExplicit            DIFlags = C.LLVMDIFlagExplicit
	DIFlagPrototyped          DIFlags = C.LLVMDIFlagPrototyped
	DIFlagObjcClassComplete   DIFlags = C.LLVMDIFlagObjcClassComplete
	DIFlagObjectPointer       DIFlags = C.LLVMDIFlagObjectPointer
	DIFlagVector              DIFlags = C.LLVMDIFlagVector
	DIFlagStaticMember        DIFlags = C.LLVMDIFlagStaticMember
	DIFlagLValueReference     DIFlags = C.LLVMDIFlagLValueReference
	DIFlagRValueReference     DIFlags = C.LLVMDIFlagRValueReference
	DIFlagReserved            DIFlags = C.LLVMDIFlagReserved
	DIFlagSingleInheritance   DIFlags = C.LLVMDIFlagSingleInheritance
	DIFlagMultipleInheritance DIFlags = C.LLVMDIFlagMultipleInheritance
	DIFlagVirtualInheritance  DIFlags = C.LLVMDIFlagVirtualInheritance
	DIFlagIntroducedVirtual   DIFlags = C.LLVMDIFlagIntroducedVirtual
	DIFlagBitField            DIFlags = C.LLVMDIFlagBitField
	DIFlagNoReturn            DIFlags = C.LLVMDIFlagNoReturn
	DIFlagTypePassByValue     DIFlags = C.LLVMDIFlagTypePassByValue
	DIFlagTypePassByReference DIFlags = C.LLVMDIFlagTypePassByReference
	DIFlagEnumClass           DIFlags = C.LLVMDIFlagEnumClass
	DIFlagFixedEnum           DIFlags = C.LLVMDIFlagFixedEnum
	DIFlagThunk               DIFlags = C.LLVMDIFlagThunk
	DIFlagNonTrivial          DIFlags = C.LLVMDIFlagNonTrivial
	DIFlagBigEndian           DIFlags = C.LLVMDIFlagBigEndian
	DIFlagLittleEndian        DIFlags = C.LLVMDIFlagLittleEndian
	DIFlagIndirectVirtualBase DIFlags = C.LLVMDIFlagIndirectVirtualBase
	DIFlagAccessibility       DIFlags = C.LLVMDIFlagAccessibility
	DIFlagPtrToMemberRep      DIFlags = C.LLVMDIFlagPtrToMemberRep
)

// -----------------------------------------------------------------------------

// DIScope represents an LLVM debug entry for a scope.
type DIScope struct {
	MDNode
}

// File returns the DI file associated with the DI scope.
func (dis DIScope) File() (dif DIFile) {
	dif.c = C.LLVMDIScopeGetFile(dis.c)
	return
}

// -----------------------------------------------------------------------------

// DIFile represents an LLVM debug info entry for a file.
type DIFile struct {
	MDNode
}

// Directory returns the directory of the DI file.
func (dif DIFile) Directory() string {
	var strlen C.uint
	cdir := C.LLVMDIFileGetDirectory(dif.c, byref(&strlen))

	return C.GoStringN(cdir, (C.int)(strlen))
}

// FileName returns the file name of the DI file.
func (dif DIFile) FileName() string {
	var strlen C.uint
	cfname := C.LLVMDIFileGetFilename(dif.c, byref(&strlen))

	return C.GoStringN(cfname, (C.int)(strlen))
}

// Source returns the source of the given DI file.
func (dif DIFile) Source() string {
	var strlen C.uint
	csrc := C.LLVMDIFileGetSource(dif.c, byref(&strlen))

	return C.GoStringN(csrc, (C.int)(strlen))
}

// AsScope converts the given DI file into a DI scope.
func (dif DIFile) AsScope() (dis DIScope) {
	dis.c = dif.c
	return
}

// -----------------------------------------------------------------------------

// DILocation represents an LLVM debug entry for a location.
type DILocation struct {
	MDNode
}

// NewDILocation creates a new a DI location in the context.
func (c Context) NewDILocation(scope DIScope, line, col int) (dil DILocation) {
	dil.c = C.LLVMDIBuilderCreateDebugLocation(
		c.c,
		(C.uint)(line),
		(C.uint)(col),
		scope.c,
		nil,
	)
	return
}

// Line returns the line number of the DI location.
func (dil DILocation) Line() int {
	return int(C.LLVMDILocationGetLine(dil.c))
}

// Column returns the column number of the DI location.
func (dil DILocation) Column() int {
	return int(C.LLVMDILocationGetColumn(dil.c))
}

// Scope returns the DI scope associated with the DI location.
func (dil DILocation) Scope() (dis DIScope) {
	dis.c = C.LLVMDILocationGetScope(dil.c)
	return
}

// InlinedAt returns the inlined location of this debug location.
func (dil DILocation) InlinedAt() (inlinedAt DILocation) {
	inlinedAt.c = C.LLVMDILocationGetInlinedAt(dil.c)
	return
}

// -----------------------------------------------------------------------------

// DIType represents an LLVM debug entry for a type.
type DIType struct {
	MDNode
}

// Name returns the name of the DI type.
func (dit DIType) Name() string {
	var strlen C.size_t
	cname := C.LLVMDITypeGetName(dit.c, byref(&strlen))

	return C.GoStringN(cname, (C.int)(strlen))
}

// BitSize returns the size of the DI type in bits.
func (dit DIType) BitSize() int {
	return int(C.LLVMDITypeGetSizeInBits(dit.c))
}

// BitAlign returns the alignment of the DI type in bits.
func (dit DIType) BitAlign() int {
	return int(C.LLVMDITypeGetAlignInBits(dit.c))
}

// BitOffset returns the offset of the DI type in bits.
func (dit DIType) BitOffset() int {
	return int(C.LLVMDITypeGetOffsetInBits(dit.c))
}

// Line returns the line number the DI type is defined on.
func (dit DIType) Line() int {
	return int(C.LLVMDITypeGetLine(dit.c))
}

// Flags returns the DI flags assigned to the DI type.
func (dit DIType) Flags() DIFlags {
	return DIFlags(C.LLVMDITypeGetFlags(dit.c))
}

// -----------------------------------------------------------------------------

// DIVariable represents an LLVM debug entry for a variable.
type DIVariable struct {
	MDNode
}

// File returns the file the DI variable is defined in.
func (div DIVariable) File() (dif DIFile) {
	dif.c = C.LLVMDIVariableGetFile(div.c)
	return
}

// Scope returns the scope the DI variable is defined in.
func (div DIVariable) Scope() (dis DIScope) {
	dis.c = C.LLVMDIVariableGetScope(div.c)
	return
}

// Line returns the line the DI variable is defined on.
func (div DIVariable) Line() int {
	return int(C.LLVMDIVariableGetLine(div.c))
}

// -----------------------------------------------------------------------------

// DIGlobalVarExpr represents an LLVM debug entry for a global var declaration.
type DIGlobalVarExpr struct {
	MDNode
}

// Variable returns the DI variable of the DI global variable expression.
func (dig DIGlobalVarExpr) Variable() (div DIVariable) {
	div.c = C.LLVMDIGlobalVariableExpressionGetVariable(dig.c)
	return
}

// Expr returns the DI expression of the DI global variable expression.
func (dig DIGlobalVarExpr) Expr() Metadata {
	return metaBase{c: C.LLVMDIGlobalVariableExpressionGetExpression(dig.c)}
}

// -----------------------------------------------------------------------------

// DISubprogram represents an LLVM debug entry for a sub-program.
type DISubprogram struct {
	MDNode
}

// Line returns the line the DI sub-program begins on.
func (dis DISubprogram) Line() int {
	return int(C.LLVMDISubprogramGetLine(dis.c))
}

// AsScope returns the scope of the DI sub-program.
func (dis DISubprogram) AsScope() (adis DIScope) {
	adis.c = dis.c
	return
}

// Subprogram returns the sub-program associated with the function.
func (fn Function) Subprogram() (dis DISubprogram) {
	dis.c = C.LLVMGetSubprogram(fn.c)
	return
}

// SetSubprogram sets the sub-program associated with the function to dis.
func (fn Function) SetSubprogram(dis DISubprogram) {
	C.LLVMSetSubprogram(fn.c, dis.c)
}

// -----------------------------------------------------------------------------

// DWARFSourceLanguage represents a DWARF source language.
type DWARFSourceLanguage C.LLVMDWARFSourceLanguage

// Enumeration of DWARF source languages.
const (
	DWARFSourceLanguageC89                 DWARFSourceLanguage = C.LLVMDWARFSourceLanguageC89
	DWARFSourceLanguageC                   DWARFSourceLanguage = C.LLVMDWARFSourceLanguageC
	DWARFSourceLanguageAda83               DWARFSourceLanguage = C.LLVMDWARFSourceLanguageAda83
	DWARFSourceLanguageC_plus_plus         DWARFSourceLanguage = C.LLVMDWARFSourceLanguageC_plus_plus
	DWARFSourceLanguageCobol74             DWARFSourceLanguage = C.LLVMDWARFSourceLanguageCobol74
	DWARFSourceLanguageCobol85             DWARFSourceLanguage = C.LLVMDWARFSourceLanguageCobol85
	DWARFSourceLanguageFortran77           DWARFSourceLanguage = C.LLVMDWARFSourceLanguageFortran77
	DWARFSourceLanguageFortran90           DWARFSourceLanguage = C.LLVMDWARFSourceLanguageFortran90
	DWARFSourceLanguagePascal83            DWARFSourceLanguage = C.LLVMDWARFSourceLanguagePascal83
	DWARFSourceLanguageModula2             DWARFSourceLanguage = C.LLVMDWARFSourceLanguageModula2
	DWARFSourceLanguageJava                DWARFSourceLanguage = C.LLVMDWARFSourceLanguageJava
	DWARFSourceLanguageC99                 DWARFSourceLanguage = C.LLVMDWARFSourceLanguageC99
	DWARFSourceLanguageAda95               DWARFSourceLanguage = C.LLVMDWARFSourceLanguageAda95
	DWARFSourceLanguageFortran95           DWARFSourceLanguage = C.LLVMDWARFSourceLanguageFortran95
	DWARFSourceLanguagePLI                 DWARFSourceLanguage = C.LLVMDWARFSourceLanguagePLI
	DWARFSourceLanguageObjC                DWARFSourceLanguage = C.LLVMDWARFSourceLanguageObjC
	DWARFSourceLanguageObjC_plus_plus      DWARFSourceLanguage = C.LLVMDWARFSourceLanguageObjC_plus_plus
	DWARFSourceLanguageUPC                 DWARFSourceLanguage = C.LLVMDWARFSourceLanguageUPC
	DWARFSourceLanguageD                   DWARFSourceLanguage = C.LLVMDWARFSourceLanguageD
	DWARFSourceLanguagePython              DWARFSourceLanguage = C.LLVMDWARFSourceLanguagePython
	DWARFSourceLanguageOpenCL              DWARFSourceLanguage = C.LLVMDWARFSourceLanguageOpenCL
	DWARFSourceLanguageGo                  DWARFSourceLanguage = C.LLVMDWARFSourceLanguageGo
	DWARFSourceLanguageModula3             DWARFSourceLanguage = C.LLVMDWARFSourceLanguageModula3
	DWARFSourceLanguageHaskell             DWARFSourceLanguage = C.LLVMDWARFSourceLanguageHaskell
	DWARFSourceLanguageC_plus_plus_03      DWARFSourceLanguage = C.LLVMDWARFSourceLanguageC_plus_plus_03
	DWARFSourceLanguageC_plus_plus_11      DWARFSourceLanguage = C.LLVMDWARFSourceLanguageC_plus_plus_11
	DWARFSourceLanguageOCaml               DWARFSourceLanguage = C.LLVMDWARFSourceLanguageOCaml
	DWARFSourceLanguageRust                DWARFSourceLanguage = C.LLVMDWARFSourceLanguageRust
	DWARFSourceLanguageC11                 DWARFSourceLanguage = C.LLVMDWARFSourceLanguageC11
	DWARFSourceLanguageSwift               DWARFSourceLanguage = C.LLVMDWARFSourceLanguageSwift
	DWARFSourceLanguageJulia               DWARFSourceLanguage = C.LLVMDWARFSourceLanguageJulia
	DWARFSourceLanguageDylan               DWARFSourceLanguage = C.LLVMDWARFSourceLanguageDylan
	DWARFSourceLanguageC_plus_plus_14      DWARFSourceLanguage = C.LLVMDWARFSourceLanguageC_plus_plus_14
	DWARFSourceLanguageFortran03           DWARFSourceLanguage = C.LLVMDWARFSourceLanguageFortran03
	DWARFSourceLanguageFortran08           DWARFSourceLanguage = C.LLVMDWARFSourceLanguageFortran08
	DWARFSourceLanguageRenderScript        DWARFSourceLanguage = C.LLVMDWARFSourceLanguageRenderScript
	DWARFSourceLanguageBLISS               DWARFSourceLanguage = C.LLVMDWARFSourceLanguageBLISS
	DWARFSourceLanguageMips_Assembler      DWARFSourceLanguage = C.LLVMDWARFSourceLanguageMips_Assembler
	DWARFSourceLanguageGOOGLE_RenderScript DWARFSourceLanguage = C.LLVMDWARFSourceLanguageGOOGLE_RenderScript
	DWARFSourceLanguageBORLAND_Delphi      DWARFSourceLanguage = C.LLVMDWARFSourceLanguageBORLAND_Delphi
)

// DWARFEmissionKind represents a DWARF emission kind: the amount of debug
// information to emit.
type DWARFEmissionKind C.LLVMDWARFEmissionKind

// Enumeration of DWARF emission kinds.
const (
	DWARFEmissionNone           DWARFEmissionKind = C.LLVMDWARFEmissionNone
	DWARFEmissionFull           DWARFEmissionKind = C.LLVMDWARFEmissionFull
	DWARFEmissionLineTablesOnly DWARFEmissionKind = C.LLVMDWARFEmissionLineTablesOnly
)

// DWARFTypeEncoding represents an LLVM DWARF type encoding.
type DWARFTypeEncoding C.LLVMDWARFTypeEncoding

// Enumeration of DWARF type encodings.
const (
	AddressTypeEncoding DWARFTypeEncoding = iota + 1
	BooleanTypeEncoding
	ComplexFloatTypeEncoding
	FloatTypeEncoding
	SignedTypeEncoding
	SignedCharTypeEncoding
	UnsignedTypeEncoding
	UnsignedCharTypeEncoding
	ImaginaryFloatTypeEncoding
	PackedDecimalTypeEncoding
	NumericStringTypeEncoding
	EditedTypeEncoding
	SignedFixedTypeEncoding
	UnsignedFixedTypeEncoding
	DecimalFloatTypeEncoding
	UTFTypeEncoding
	UCSTypeEncoding
	ASCIITypeEncoding
)

// DWARFTypeQualifier represents a DWARF type qualifier.
type DWARFTypeQualifier C.uint

// Enumeration of DWARF type qualifiers.
const (
	AtomicTypeQualifier          DWARFTypeQualifier = 0x47
	ConstTypeQualifier           DWARFTypeQualifier = 0x26
	ImmutableTypeQualifier       DWARFTypeQualifier = 0x4b
	PackedTypeQualifier          DWARFTypeQualifier = 0x2d
	RestrictTypeQualifier        DWARFTypeQualifier = 0x37
	RvalueReferenceTypeQualifier DWARFTypeQualifier = 0x42
	SharedTypeQualifier          DWARFTypeQualifier = 0x40
	VolatileTypeQualifier        DWARFTypeQualifier = 0x35
)

// DWARFExprOpCode represents a DWARF expression op code.
type DWARFExprOpCode C.uint

// Enumeration of DWARF expression op codes supported by LLVM.
const (
	DerefExprOpCode           DWARFExprOpCode = 0x06
	PlusExprOpCode            DWARFExprOpCode = 0x22
	MinusExprOpCode           DWARFExprOpCode = 0x1c
	PlusUConstExprOpCode      DWARFExprOpCode = 0x23
	SwapExprOpCode            DWARFExprOpCode = 0x16
	XDerefExprOpCode          DWARFExprOpCode = 0x18
	StackValueExprOpCode      DWARFExprOpCode = 0x9f
	BRegExprOpCode            DWARFExprOpCode = 0x92
	PushObjectAddrExprOpCode  DWARFExprOpCode = 0x97
	OverExprOpCode            DWARFExprOpCode = 0x14
	LLVMFragmentExprOpCode    DWARFExprOpCode = 0x1000
	LLVMConvertExprOpCode     DWARFExprOpCode = 0x1001
	LLVMTagOffsetExprOpCode   DWARFExprOpCode = 0x1002
	LLVMEntryValueExprOpCode  DWARFExprOpCode = 0x1003
	LLVMImplicitPtrExprOpCode DWARFExprOpCode = 0x1004
	LLVMArgExprOpCode         DWARFExprOpCode = 0x1005
)

// -----------------------------------------------------------------------------

// DIBuilder represents an LLVM debug info builder.
type DIBuilder struct {
	c C.LLVMDIBuilderRef
}

// NewDIBuilder creates a new DI builder for mod in the context.
func (c Context) NewDIBuilder(mod Module) (dib DIBuilder) {
	dib.c = C.LLVMCreateDIBuilder(mod.c)
	c.takeOwnership(dib)
	return
}

// Finalize finalizes the generation of debug information.
func (dib DIBuilder) Finalize() {
	C.LLVMDIBuilderFinalize(dib.c)
}

// dispose disposes of the DI builder.
func (dib DIBuilder) dispose() {
	C.LLVMDisposeDIBuilder(dib.c)
}

// -----------------------------------------------------------------------------

// Location returns the builder's current debug location.
func (irb IRBuilder) Location() (dil DILocation, exists bool) {
	dilPtr := C.LLVMGetCurrentDebugLocation2(irb.c)

	if dilPtr == nil {
		exists = false
	} else {
		dil.c = dilPtr
		exists = true
	}

	return
}

// SetLocation sets the builder's current debug location to loc.
func (irb IRBuilder) SetLocation(loc DILocation) {
	C.LLVMSetCurrentDebugLocation2(irb.c, loc.c)
}

// ClearLocation clears the builder's current debug location (sets it to NULL).
func (irb IRBuilder) ClearLocation() {
	C.LLVMSetCurrentDebugLocation2(irb.c, nil)
}

// -----------------------------------------------------------------------------

// NewFile creates a new debug descriptor for a file.
func (dib DIBuilder) NewFile(fileName, dirName string) (dif DIFile) {
	cfname := C.CString(fileName)
	defer C.free(unsafe.Pointer(cfname))

	cdname := C.CString(dirName)
	defer C.free(unsafe.Pointer(cdname))

	dif.c = C.LLVMDIBuilderCreateFile(
		dib.c,
		cfname,
		(C.size_t)(len(fileName)),
		cdname,
		(C.size_t)(len(dirName)),
	)
	return
}

// CompileUnitOptions represents the additional options used to create a DWARF
// compile unit.  This helps manage the excessive number of extraneous options
// LLVM expects to create a DICompileUnit.
type CompileUnitOptions struct {
	// The identifying string of the compiler which produced this compile unit.
	Producer string

	// Whether the compile unit is optimized.
	IsOptimized bool

	// The compile flags used to generate this compile unit.
	CompileFlags string

	// The path to output the debug file to if a separate file is to be produced.
	DebugFileOutputPath string

	// The runtime version (if relevant).
	RuntimeVersion int

	// The DWO ID if this is a split skeleton compile unit.
	DWOID int

	// Whether to emit inline debug info.
	EmitInlineDebugInfo bool

	// Whether to emit extra info for profile collection.
	EmitDebugInfoForProfiling bool

	// The `clang` system root.
	SystemRoot string

	// The SDK string.
	SDK string
}

// NewCompileUnit creates a new debug descriptor for a compile unit.
func (dib DIBuilder) NewCompileUnit(
	file DIFile,
	sourceLang DWARFSourceLanguage,
	emissionKind DWARFEmissionKind,
	options CompileUnitOptions,
) (cu MDNode) {
	cproducer := C.CString(options.Producer)
	defer C.free(unsafe.Pointer(cproducer))

	cflags := C.CString(options.CompileFlags)
	defer C.free(unsafe.Pointer(cflags))

	cdebugfpath := C.CString(options.DebugFileOutputPath)
	defer C.free(unsafe.Pointer(cdebugfpath))

	csysroot := C.CString(options.SystemRoot)
	defer C.free(unsafe.Pointer(csysroot))

	csdk := C.CString(options.SDK)
	defer C.free(unsafe.Pointer(csdk))

	cu.c = C.LLVMDIBuilderCreateCompileUnit(
		dib.c,
		(C.LLVMDWARFSourceLanguage)(sourceLang),
		file.c,
		cproducer,
		(C.size_t)(len(options.Producer)),
		llvmBool(options.IsOptimized),
		cflags,
		(C.size_t)(len(options.CompileFlags)),
		(C.uint)(options.RuntimeVersion),
		cdebugfpath,
		(C.size_t)(len(options.DebugFileOutputPath)),
		(C.LLVMDWARFEmissionKind)(emissionKind),
		(C.uint)(options.DWOID),
		llvmBool(options.EmitInlineDebugInfo),
		llvmBool(options.EmitDebugInfoForProfiling),
		csysroot,
		(C.size_t)(len(options.SystemRoot)),
		csdk,
		(C.size_t)(len(options.SDK)),
	)
	return
}

// ModuleOptions represents the additional options used to create a DWARF module.
type ModuleOptions struct {
	// A space-separated shell-quoted list of -D macro definitions as they would
	// appear on a command line.
	ConfigMacros string

	// The path to the module map file.
	IncludePath string

	// The path to an API notes file for the module.
	APINotesPath string
}

// NewModule creates a new debug descriptor for a module.
func (dib DIBuilder) NewModule(parentScope DIScope, name string, options ModuleOptions) (mod MDNode) {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))

	cmacros := C.CString(options.ConfigMacros)
	defer C.free(unsafe.Pointer(cmacros))

	cip := C.CString(options.IncludePath)
	defer C.free(unsafe.Pointer(cip))

	cnotes := C.CString(options.APINotesPath)
	defer C.free(unsafe.Pointer(cnotes))

	mod.c = C.LLVMDIBuilderCreateModule(
		dib.c,
		parentScope.c,
		cname,
		(C.size_t)(len(name)),
		cmacros,
		(C.size_t)(len(options.ConfigMacros)),
		cip,
		(C.size_t)(len(options.IncludePath)),
		cnotes,
		(C.size_t)(len(options.APINotesPath)),
	)
	return
}

// NewNamespace creates a new debug descriptor for a namespace.
func (dib DIBuilder) NewNamespace(parentScope DIScope, name string, exportsSymbols bool) (ns MDNode) {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))

	ns.c = C.LLVMDIBuilderCreateNameSpace(dib.c, parentScope.c, cname, (C.size_t)(len(name)), llvmBool(exportsSymbols))
	return
}

// NewFunction creates a new debug descriptor for a function.
func (dib DIBuilder) NewFunction(
	parent_scope DIScope,
	file DIFile,
	name, mangledName string,
	line int,
	funcType DIType,
	internal bool,
	isDefinition bool,
	scopeLine int,
	flags DIFlags,
	isOptimized bool,
) (dis DISubprogram) {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))

	cmangled := C.CString(name)
	defer C.free(unsafe.Pointer(cmangled))

	dis.c = C.LLVMDIBuilderCreateFunction(
		dib.c,
		parent_scope.c,
		cname,
		(C.size_t)(len(name)),
		cmangled,
		(C.size_t)(len(mangledName)),
		file.c,
		(C.uint)(line),
		funcType.c,
		llvmBool(internal),
		llvmBool(isDefinition),
		(C.uint)(scopeLine),
		(C.LLVMDIFlags)(flags),
		llvmBool(isOptimized),
	)
	return
}

// NewLexicalBlock creates a new debug descriptor for a lexical block.
func (dib DIBuilder) NewLexicalBlock(scope DIScope, file DIFile, line, col int) (dis DIScope) {
	dis.c = C.LLVMDIBuilderCreateLexicalBlock(dib.c, scope.c, file.c, (C.uint)(line), (C.uint)(col))
	return
}

// NewParameterVariable creates a new debug descriptor for a parameter variable.
func (dib DIBuilder) NewParameterVariable(
	scope DIScope,
	file DIFile,
	name string,
	argn int,
	line int,
	paramType DIType,
	surivesOptimizations bool,
	flags DIFlags,
) (div DIVariable) {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))

	div.c = C.LLVMDIBuilderCreateParameterVariable(
		dib.c,
		scope.c,
		cname,
		(C.size_t)(len(name)),
		(C.uint)(argn),
		file.c,
		(C.uint)(line),
		paramType.c,
		llvmBool(surivesOptimizations),
		(C.LLVMDIFlags)(flags),
	)
	return
}

// NewLocalVariable creates a new debug descriptor for a local variable.
func (dib DIBuilder) NewLocalVariable(
	scope DIScope,
	file DIFile,
	name string,
	line int,
	typ DIType,
	bitAlign int,
	survivesOptimizations bool,
	flags DIFlags,
) (div DIVariable) {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))

	div.c = C.LLVMDIBuilderCreateAutoVariable(
		dib.c,
		scope.c,
		cname,
		(C.size_t)(len(name)),
		file.c,
		(C.uint)(line),
		typ.c,
		llvmBool(survivesOptimizations),
		(C.LLVMDIFlags)(flags),
		(C.uint)(bitAlign),
	)
	return
}

// NewConstant creates a new debug descriptor for a variable which does not
// have an address but does have a constant value.
func (dib DIBuilder) NewConstant(value int) (cv MDNode) {
	cv.c = C.LLVMDIBuilderCreateConstantValueExpression(dib.c, (C.int64_t)(value))
	return
}

// NewAddrExpression creates a new debug descriptor for a variable which has a
// complex address expression for its address: eg. a complex LHS in assignment.
func (dib DIBuilder) NewAddrExpression(ops ...DWARFExprOpCode) (ae MDNode) {
	opsArr := make([]C.int64_t, len(ops))
	for i, op := range ops {
		opsArr[i] = (C.int64_t)(op)
	}

	ae.c = C.LLVMDIBuilderCreateExpression(dib.c, byref(&opsArr[0]), (C.size_t)(len(ops)))
	return
}

// -----------------------------------------------------------------------------

// NewSubrange creates a debug descriptor for a value range.
func (dib DIBuilder) NewSubrange(lowerBound, count int) (sr MDNode) {
	sr.c = C.LLVMDIBuilderGetOrCreateSubrange(dib.c, (C.int64_t)(lowerBound), (C.int64_t)(count))
	return
}

// NewArray creates a new array of DI nodes.
func (dib DIBuilder) NewArray(data ...Metadata) (diArr MDNode) {
	dataArr := make([]C.LLVMMetadataRef, len(data))
	for i, item := range data {
		dataArr[i] = item.ptr()
	}

	diArr.c = C.LLVMDIBuilderGetOrCreateArray(dib.c, byref(&dataArr[0]), (C.size_t)(len(data)))
	return
}

// -----------------------------------------------------------------------------

// NewBasicType creates a new DWARF basic type.
func (dib DIBuilder) NewBasicType(name string, bitSize int, encoding DWARFTypeEncoding, flags DIFlags) (dit DIType) {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))

	dit.c = C.LLVMDIBuilderCreateBasicType(dib.c, cname, (C.size_t)(len(name)), (C.size_t)(bitSize), (C.uint)(encoding), (C.LLVMDIFlags)(flags))
	return
}

// NewPointerType creates a new DWARF pointer type.
func (dib DIBuilder) NewPointerType(elemType DIType, bitSize, bitAlign, addrSpace int) (dit DIType) {
	dit.c = C.LLVMDIBuilderCreatePointerType(
		dib.c,
		elemType.c,
		(C.size_t)(bitSize),
		(C.uint)(bitAlign),
		(C.uint)(addrSpace),
		nil,
		0,
	)
	return
}

// NewSubroutineType creates a new DWARF subroutine type.
func (dib DIBuilder) NewSubroutineType(file DIFile, flags DIFlags, paramTypes ...DIType) (dit DIType) {
	paramTypeArr := make([]C.LLVMMetadataRef, len(paramTypes))
	for i, paramType := range paramTypes {
		paramTypeArr[i] = paramType.c
	}

	dit.c = C.LLVMDIBuilderCreateSubroutineType(
		dib.c,
		file.c,
		byref(&paramTypeArr[0]),
		(C.uint)(len(paramTypes)),
		(C.LLVMDIFlags)(flags),
	)
	return
}

// -----------------------------------------------------------------------------

// InsertDeclareBefore inserts an `llvm.dbg.declare` intrinsic call before instr.
func (dib DIBuilder) InsertDeclareBefore(
	variable Value,
	varInfo DIVariable,
	addrExpr MDNode,
	debugLoc DILocation,
	instr Instruction,
) (ci CallInstruction) {
	ci.c = C.LLVMDIBuilderInsertDeclareBefore(
		dib.c,
		variable.ptr(),
		varInfo.c,
		addrExpr.c,
		debugLoc.c,
		instr.c,
	)
	return
}

// InsertDeclareAtEnd inserts an `llvm.dbg.declare` intrinsic call at the end of bb.
func (dib DIBuilder) InsertDeclareAtEnd(
	variable Value,
	varInfo DIVariable,
	addrExpr MDNode,
	debugLoc DILocation,
	bb BasicBlock,
) (ci CallInstruction) {
	ci.c = C.LLVMDIBuilderInsertDeclareAtEnd(
		dib.c,
		variable.ptr(),
		varInfo.c,
		addrExpr.c,
		debugLoc.c,
		bb.c,
	)
	return
}

// InsertValueBefore inserts an `llvm.dbg.value` intrinsic call before instr.
func (dib DIBuilder) InsertValueBefore(
	val Value,
	varInfo DIVariable,
	addrExpr MDNode,
	debugLoc DILocation,
	instr Instruction,
) (ci CallInstruction) {
	ci.c = C.LLVMDIBuilderInsertDbgValueBefore(
		dib.c,
		val.ptr(),
		varInfo.c,
		addrExpr.c,
		debugLoc.c,
		instr.c,
	)
	return
}

// InsertValueAtEnd inserts an `llvm.dbg.value` intrinsic call at the end of bb.
func (dib DIBuilder) InsertValueAtEnd(
	val Value,
	varInfo DIVariable,
	addrExpr MDNode,
	debugLoc DILocation,
	bb BasicBlock,
) (ci CallInstruction) {
	ci.c = C.LLVMDIBuilderInsertDbgValueAtEnd(
		dib.c,
		val.ptr(),
		varInfo.c,
		addrExpr.c,
		debugLoc.c,
		bb.c,
	)
	return
}
