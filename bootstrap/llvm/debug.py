from enum import auto
from ctypes import c_size_t, c_uint32, c_uint64, c_int64, c_char_p, c_uint, POINTER

from . import *
from .module import Module
from .metadata import *
from .value import Value
from .ir import BasicBlock

class DWARFSourceLanguage(LLVMEnum):
    C89 = auto()
    C = auto()
    ADA83 = auto()
    C_PLUS_PLUS = auto()
    COBOL74 = auto()
    COBOL85 = auto()
    FORTRAN77 = auto()
    FORTRAN90 = auto()
    PASCAL83 = auto()
    MODULA2 = auto()
    JAVA = auto()
    C99 = auto()
    ADA95 = auto()
    FORTRAN95 = auto()
    PLI = auto()
    OBJ_C = auto()
    OBJ_C_PLUS_PLUS = auto()
    UPC = auto()
    D = auto()
    PYTHON = auto()
    OPEN_CL = auto()
    GO = auto()
    MODULA3 = auto()
    HASKELL = auto()
    C_PLUS_PLUS_03 = auto()
    C_PLUS_PLUS_11 = auto()
    O_CAML = auto()
    RUST = auto()
    C11 = auto()
    SWIFT = auto()
    JULIA = auto()
    DYLAN = auto()
    C_PLUS_PLUS_14 = auto()
    FORTRAN03 = auto()
    FORTRAN08 = auto()
    RENDER_SCRIPT = auto()
    BLISS = auto()
    MIPS_ASSEMBLER = auto()
    GOOGLE_RENDER_SCRIPT = auto()
    BORLAND_DELPHI = auto()

class DWARFEmissionKind(LLVMEnum):
    NONE = auto()
    FULL = auto()
    LINE_TABLES_ONLY = auto()

class DWARFMacroInfoRecordType(LLVMEnum):
    DEFINE = 1
    MACRO = 2
    START_FILE = 3
    END_FILE = 4
    VENDOR_EXT = 0xff

class DWARFTypeEncoding(LLVMEnum):
    ADDRESS = auto()
    BOOLEAN = auto()
    COMPLEX_FLOAT = auto()
    FLOAT = auto()
    SIGNED = auto()
    SIGNED_CHAR = auto()
    UNSIGNED = auto()
    UNSIGNED_CHAR = auto()
    IMAGINARY_FLOAT = auto()
    PACKED_DECIMAL = auto()
    NUMERIC_STRING = auto()
    EDITED = auto()
    SIGNED_FIXED = auto()
    UNSIGNED_FIXED = auto()
    DECIMAL_FLOAT = auto()

# ---------------------------------------------------------------------------- #

class DIBuilder(LLVMObject):
    def __init__(self, mod: Module):
        super().__init__(LLVMCreateDIBuilder(mod))
        get_context().take_ownership(self)

    def dispose(self):
        LLVMDisposeDIBuilder(self)

    def finalize(self):
        LLVMDIBuilderFinalize(self)

    # ---------------------------------------------------------------------------- #

    def create_compile_unit(
        self,
        file: DIFile,
        source_lang: DWARFSourceLanguage,
        debug_file_output_path: str,
        producer: str = "",
        is_optimized: bool = False,
        flags: str = "",
        runtime_version: int = 0,
        kind: DWARFEmissionKind = DWARFEmissionKind.FULL,
        dwo_id: int = 0,
        inline_debug_info: bool = False,
        profile_debug_info: bool = False,
        sys_root: str = "",
        sdk: str = "",
    ) -> MDNode:
        out_path_bytes = debug_file_output_path.encode()
        producer_bytes = producer.encode()
        flags_bytes = flags.encode()
        sys_root_bytes = sys_root.encode()
        sdk_bytes = sdk.encode()

        return MDNode(ptr=LLVMDIBuilderCreateCompileUnit(
            self,
            source_lang,
            file,
            producer_bytes,
            len(producer_bytes),
            int(is_optimized),
            flags_bytes,
            len(flags_bytes),
            runtime_version,
            out_path_bytes,
            len(out_path_bytes),
            kind,
            dwo_id,
            int(inline_debug_info),
            int(profile_debug_info),
            sys_root_bytes,
            len(sys_root_bytes),
            sdk_bytes,
            len(sdk_bytes)
        ))

    def create_file(self, file_name: str, dir_name: str) -> DIFile:
        file_name_bytes = file_name.encode()
        dir_name_bytes = dir_name.encode()

        return DIFile(LLVMDIBuilderCreateFile(self, file_name_bytes, len(file_name_bytes), dir_name_bytes, len(dir_name_bytes)))

    def create_module(
        self,
        parent_scope: DIScope,
        name: str,
        config_macros: str = "",
        include_paths: str = "",
        api_notes_path: str = ""
    ) -> MDNode:
        name_bytes = name.encode()
        cm_bytes = config_macros.encode()
        inc_path_bytes = include_paths.encode()
        api_notes_path_bytes = api_notes_path.encode()

        return MDNode(LLVMDIBuilderCreateModule(
            self,
            parent_scope,
            name_bytes,
            len(name_bytes),
            cm_bytes,
            len(cm_bytes),
            inc_path_bytes,
            len(inc_path_bytes),
            api_notes_path_bytes,
            len(api_notes_path_bytes)
        ))

    def create_function(
        self,
        func_scope: DIScope,
        name: str,
        mangled_name: str,
        file: DIFile,
        line: int,
        func_type: DIType,
        externally_visible: bool,
        is_definition: bool,
        scope_line: int,
        flags: DIFlags = DIFlags.ZERO,
        is_optimized: bool = False
    ) -> MDNode:
        name_bytes = name.encode()
        mangled_bytes = mangled_name.encode()

        return MDNode(ptr=LLVMDIBuilderCreateFunction(
            self,
            func_scope,
            name_bytes,
            len(name_bytes),
            mangled_bytes,
            len(mangled_bytes),
            file,
            line,
            func_type,
            int(externally_visible),
            int(is_definition),
            scope_line,
            flags,
            int(is_optimized)
        ))

    def create_lexical_block(self, scope: DIScope, file: DIFile, line: int, col: int) -> MDNode:
        return MDNode(ptr=LLVMDIBuilderCreateLexicalBlock(
            scope,
            file,
            line,
            col,
        ))

    
    

# ---------------------------------------------------------------------------- #

@llvm_api
def LLVMCreateDIBuilder(m: Module) -> c_object_p:
    pass

@llvm_api
def LLVMDisposeDIBuilder(builder: DIBuilder):
    pass

@llvm_api
def LLVMDIBuilderFinalize(builder: DIBuilder):
    pass

@llvm_api
def LLVMDIBuilderCreateCompileUnit(builder: DIBuilder, lang: DWARFSourceLanguage, file_ref: DIFile, producer: c_char_p, producer_len: c_size_t, is_optimized: c_enum, flags: c_char_p, flags_len: c_size_t, runtime_ver: c_uint, split_name: c_char_p, split_name_len: c_size_t, kind: DWARFEmissionKind, dwo_id: c_uint, split_debug_inlining: c_enum, debug_info_for_profiling: c_enum, sys_root: c_char_p, sys_root_len: c_size_t, sdk: c_char_p, sdk_len: c_size_t) -> c_object_p:
    pass

@llvm_api
def LLVMDIBuilderCreateFile(builder: DIBuilder, filename: c_char_p, filename_len: c_size_t, directory: c_char_p, directory_len: c_size_t) -> c_object_p:
    pass

@llvm_api
def LLVMDIBuilderCreateModule(builder: DIBuilder, parent_scope: DIScope, name: c_char_p, name_len: c_size_t, config_macros: c_char_p, config_macros_len: c_size_t, include_path: c_char_p, include_path_len: c_size_t, api_notes_file: c_char_p, api_notes_file_len: c_size_t) -> c_object_p:
    pass

@llvm_api
def LLVMDIBuilderCreateFunction(builder: DIBuilder, scope: DIScope, name: c_char_p, name_len: c_size_t, linkage_name: c_char_p, linkage_name_len: c_size_t, file: DIFile, line_no: c_uint, ty: DIType, is_local_to_unit: c_enum, is_definition: c_enum, scope_line: c_uint, flags: DIFlags, is_optimized: c_enum) -> c_object_p:
    pass

@llvm_api
def LLVMDIBuilderCreateLexicalBlock(builder: DIBuilder, scope: DIScope, file: DIFile, line: c_uint, column: c_uint) -> c_object_p:
    pass

@llvm_api
def LLVMDIBuilderCreateImportedDeclaration(builder: DIBuilder, scope: DIScope, decl: MDNode, file: DIFile, line: c_uint, name: c_char_p, name_len: c_size_t, elements: POINTER(c_object_p), num_elements: c_uint) -> c_object_p:
    pass

@llvm_api
def LLVMDIBuilderCreateDebugLocation(ctx: Context, line: c_uint, column: c_uint, scope: DIScope, inlined_at: DILocation) -> c_object_p:
    pass

@llvm_api
def LLVMDIBuilderGetOrCreateTypeArray(builder: DIBuilder, data: POINTER(c_object_p), num_elements: c_size_t) -> c_object_p:
    pass

@llvm_api
def LLVMDIBuilderCreateEnumerator(builder: DIBuilder, name: c_char_p, name_len: c_size_t, value: c_int64, is_unsigned: c_enum) -> c_object_p:
    pass

@llvm_api
def LLVMDIBuilderCreateEnumerationType(builder: DIBuilder, scope: DIScope, name: c_char_p, name_len: c_size_t, file: DIFile, line_number: c_uint, size_in_bits: c_uint64, align_in_bits: c_uint32, elements: POINTER(c_object_p), num_elements: c_uint, class_ty: DIType) -> c_object_p:
    pass

@llvm_api
def LLVMDIBuilderCreateUnionType(builder: DIBuilder, scope: DIScope, name: c_char_p, name_len: c_size_t, file: DIFile, line_number: c_uint, size_in_bits: c_uint64, align_in_bits: c_uint32, flags: DIFlags, elements: POINTER(c_object_p), num_elements: c_uint, run_time_lang: c_uint, unique_id: c_char_p, unique_id_len: c_size_t) -> c_object_p:
    pass

@llvm_api
def LLVMDIBuilderCreateBasicType(builder: DIBuilder, name: c_char_p, name_len: c_size_t, size_in_bits: c_uint64, encoding: DWARFTypeEncoding, flags: DIFlags) -> c_object_p:
    pass

@llvm_api
def LLVMDIBuilderCreatePointerType(builder: DIBuilder, pointee_ty: Metadata, size_in_bits: c_uint64, align_in_bits: c_uint32, address_space: c_uint, name: c_char_p, name_len: c_size_t) -> c_object_p:
    pass

@llvm_api
def LLVMDIBuilderCreateStructType(builder: DIBuilder, scope: DIScope, name: c_char_p, name_len: c_size_t, file: DIFile, line_number: c_uint, size_in_bits: c_uint64, align_in_bits: c_uint32, flags: DIFlags, derived_from: Metadata, elements: POINTER(c_object_p), num_elements: c_uint, run_time_lang: c_uint, v_table_holder: Metadata, unique_id: c_char_p, unique_id_len: c_size_t) -> c_object_p:
    pass

@llvm_api
def LLVMDIBuilderCreateMemberType(builder: DIBuilder, scope: DIScope, name: c_char_p, name_len: c_size_t, file: DIFile, line_no: c_uint, size_in_bits: c_uint64, align_in_bits: c_uint32, offset_in_bits: c_uint64, flags: DIFlags, ty: DIType) -> c_object_p:
    pass

@llvm_api
def LLVMDIBuilderCreateStaticMemberType(builder: DIBuilder, scope: DIScope, name: c_char_p, name_len: c_size_t, file: DIFile, line_number: c_uint, type: DIType, flags: DIFlags, constant_val: Value, align_in_bits: c_uint32) -> c_object_p:
    pass

@llvm_api
def LLVMDIBuilderCreateTypedef(builder: DIBuilder, type: Metadata, name: c_char_p, name_len: c_size_t, file: DIFile, line_no: c_uint, scope: DIScope, align_in_bits: c_uint32) -> c_object_p:
    pass

@llvm_api
def LLVMDIBuilderGetOrCreateSubrange(builder: DIBuilder, lower_bound: c_int64, count: c_int64) -> c_object_p:
    pass

@llvm_api
def LLVMDIBuilderGetOrCreateArray(builder: DIBuilder, data: POINTER(c_object_p), num_elements: c_size_t) -> c_object_p:
    pass

@llvm_api
def LLVMDIBuilderCreateExpression(builder: DIBuilder, addr: POINTER(c_uint64), length: c_size_t) -> c_object_p:
    pass

@llvm_api
def LLVMDIBuilderCreateGlobalVariableExpression(builder: DIBuilder, scope: DIScope, name: c_char_p, name_len: c_size_t, linkage: c_char_p, link_len: c_size_t, file: DIFile, line_no: c_uint, ty: DIType, local_to_unit: c_enum, expr: Metadata, decl: Metadata, align_in_bits: c_uint32) -> c_object_p:
    pass

@llvm_api
def LLVMDIBuilderInsertDeclareBefore(builder: DIBuilder, storage: Value, var_info: DIVariable, expr: Metadata, debug_loc: DILocation, instr: Value) -> c_object_p:
    pass

@llvm_api
def LLVMDIBuilderInsertDeclareAtEnd(builder: DIBuilder, storage: Value, var_info: DIVariable, expr: Metadata, debug_loc: DILocation, block: BasicBlock) -> c_object_p:
    pass

@llvm_api
def LLVMDIBuilderInsertDbgValueBefore(builder: DIBuilder, val: Value, var_info: DIVariable, expr: Metadata, debug_loc: DILocation, instr: Value) -> c_object_p:
    pass

@llvm_api
def LLVMDIBuilderInsertDbgValueAtEnd(builder: DIBuilder, val: Value, var_info: DIVariable, expr: Metadata, debug_loc: DILocation, block: BasicBlock) -> c_object_p:
    pass

@llvm_api
def LLVMDIBuilderCreateParameterVariable(builder: DIBuilder, scope: DIScope, name: c_char_p, name_len: c_size_t, arg_no: c_uint, file: DIFile, line_no: c_uint, ty: DIType, always_preserve: c_enum, flags: DIFlags) -> c_object_p:
    pass

@llvm_api
def LLVMDIBuilderCreateNameSpace(builder: DIBuilder, parent_scope: Metadata, name: c_char_p, name_len: c_size_t, export_symbols: c_enum) -> c_object_p:
    pass

@llvm_api
def LLVMDIBuilderCreateImportedModuleFromNamespace(builder: DIBuilder, scope: Metadata, n_s: Metadata, file: Metadata, line: c_uint) -> c_object_p:
    pass

@llvm_api
def LLVMDIBuilderCreateImportedModuleFromAlias(builder: DIBuilder, scope: Metadata, imported_entity: Metadata, file: Metadata, line: c_uint, *_elements: Metadata, num_elements: c_uint) -> c_object_p:
    pass

@llvm_api
def LLVMDIBuilderCreateImportedModuleFromModule(builder: DIBuilder, scope: Metadata, m: Metadata, file: Metadata, line: c_uint, *_elements: Metadata, num_elements: c_uint) -> c_object_p:
    pass

@llvm_api
def LLVMDIBuilderCreateArrayType(builder: DIBuilder, size: c_uint64, align_in_bits: c_uint32, ty: Metadata, *_subscripts: Metadata, num_subscripts: c_uint) -> c_object_p:
    pass

@llvm_api
def LLVMDIBuilderCreateVectorType(builder: DIBuilder, size: c_uint64, align_in_bits: c_uint32, ty: Metadata, *_subscripts: Metadata, num_subscripts: c_uint) -> c_object_p:
    pass

@llvm_api
def LLVMDIBuilderCreateMemberPointerType(builder: DIBuilder, pointee_type: Metadata, class_type: Metadata, size_in_bits: c_uint64, align_in_bits: c_uint32, flags: DIFlags) -> c_object_p:
    pass

@llvm_api
def LLVMDIBuilderCreateObjectPointerType(builder: DIBuilder, type: Metadata) -> c_object_p:
    pass

