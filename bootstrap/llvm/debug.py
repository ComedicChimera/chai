from enum import auto
from ctypes import c_size_t, c_uint32, c_uint64, c_int64, c_char_p, c_uint, POINTER
from typing import List, Optional 

from . import *
from .module import Module
from .metadata import *
from .value import Value
from .ir import BasicBlock, Instruction, CallInstruction

class DWARFSourceLanguage(LLVMEnum):
    C89 = 1
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
    UNDEF = 2
    START_FILE = 3
    END_FILE = 4
    VENDOR_EXT = 0xff

class DWARFTypeEncoding(LLVMEnum):
    ADDRESS = 1
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
    UTF = auto()
    UCS = auto()
    ASCII = auto()

class DWARFTypeQualifier(LLVMEnum):
    ATOMIC = 0x47
    CONST = 0x26
    IMMUTABLE = 0x4b
    PACKED = 0x2d
    RESTRICT = 0x37
    RVALUE_REFERENCE = 0x42
    SHARED = 0x40
    VOLATILE = 0x35

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
        producer: str = "",
        is_optimized: bool = False,
        flags: str = "",
        debug_file_output_path: str = "",
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
        file: DIFile,
        name: str,
        mangled_name: str,
        line: int,
        func_type: DIType,
        internal: bool,
        is_definition: bool,
        scope_line: int,
        flags: DIFlags = DIFlags.ZERO,
        is_optimized: bool = False
    ) -> DISubprogram:
        name_bytes = name.encode()
        mangled_bytes = mangled_name.encode()

        return DISubprogram(LLVMDIBuilderCreateFunction(
            self,
            func_scope,
            name_bytes,
            len(name_bytes),
            mangled_bytes,
            len(mangled_bytes),
            file,
            line,
            func_type,
            int(internal),
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

    def create_namespace(self, parent_scope: DIScope, name: str, exports_symbols: bool = False) -> MDNode:
        name_bytes = name.encode()

        return MDNode(ptr=LLVMDIBuilderCreateNameSpace(
            self, 
            parent_scope, 
            name_bytes,
            len(name_bytes),
            int(exports_symbols)
        ))

    def create_imported_namespace(self, parent_scope: DIScope, file: DIFile, line: int) -> MDNode:
        return MDNode(ptr=LLVMDIBuilderCreateImportedModuleFromNamespace(
            self,
            parent_scope,
            file,
            line
        ))

    def create_aliased_imported_module(
        self, 
        parent_scope: DIScope, 
        alias_entity: MDNode, 
        file: DIFile, 
        line: int, 
        renamed: List[Metadata] = []
    ) -> MDNode:
        return MDNode(ptr=LLVMDIBuilderCreateImportedModuleFromAlias(
            self,
            parent_scope,
            alias_entity,
            file,
            line,
            create_object_array(renamed),
            len(renamed)
        ))

    def create_imported_module(
        self,
        parent_scope: DIScope,
        mod: Module,
        file: DIFile,
        line: int,
        renamed: List[Metadata] = []
    ) -> MDNode:
        return MDNode(ptr=LLVMDIBuilderCreateImportedModuleFromModule(
            self,
            parent_scope,
            mod,
            file,
            line,
            create_object_array(renamed),
            len(renamed)
        ))

    def create_imported_decl(
        self,
        parent_scope: DIScope,
        decl: MDNode,
        file: DIFile,
        line: int,
        name: str,
        renamed: List[Metadata] = []
    ) -> MDNode:
        name_bytes = name.encode()

        return MDNode(ptr=LLVMDIBuilderCreateImportedDeclaration(
            self,
            parent_scope,
            decl,
            file,
            line,
            name_bytes,
            len(name_bytes),
            create_object_array(renamed),
            len(renamed)
        ))

    def create_type_array(self, *types: DIType) -> MDNode:
        return MDNode(ptr=LLVMDIBuilderGetOrCreateTypeArray(self, create_object_array(types), len(types)))

    def create_subroutine_type(self, file: DIFile, *param_types: DIType, flags: DIFlags = DIFlags.ZERO) -> DIType:
        return DIType(LLVMDIBuilderCreateSubroutineType(self, file, create_object_array(param_types), len(param_types), flags))

    def create_enum_value(self, name: str, value: int, unsigned: bool = False) -> MDNode:
        name_bytes = name.encode()

        return MDNode(ptr=LLVMDIBuilderCreateEnumerator(self, name_bytes, len(name_bytes), value, int(unsigned)))

    def create_enum_type(
        self, 
        scope: DIScope, 
        file: DIFile, 
        name: str, 
        line: int,
        bit_size: int,
        bit_align: int,
        *elements: MDNode,
        underlying_type: Optional[DIType] = None
    ) -> DIType:
        name_bytes = name.encode()

        return DIType(LLVMDIBuilderCreateEnumerationType(
            self,
            scope,
            name_bytes,
            len(name_bytes),
            file, 
            line,
            bit_size,
            bit_align,
            create_object_array(elements),
            len(elements),
            underlying_type
        ))

    def create_union_type(
        self,
        scope: DIScope,
        file: DIFile,
        name: str,
        line: int,
        bit_size: int,
        bit_align: int,
        *elements: DIType,
        flags: DIFlags = DIFlags.ZERO,
        unique_id: str = "",
        runtime_ver: int = 0
    ) -> DIType:
        name_bytes = name.encode()
        unique_id_bytes = unique_id.encode() if unique_id else name_bytes

        return DIType(LLVMDIBuilderCreateUnionType(
            self,
            scope,
            name_bytes,
            len(name_bytes),
            file,
            line,
            bit_size,
            bit_align,
            flags,
            create_object_array(elements),
            len(elements),
            runtime_ver,
            unique_id_bytes,
            len(unique_id_bytes)
        ))

    def create_array_type(self, elem_type: DIType, bit_size: int, bit_align: int, *subranges: MDNode) -> DIType:
        subranges_arr = (c_object_p * len(subranges))([x.ptr for x in subranges])

        return DIType(LLVMDIBuilderCreateArrayType(self, bit_size, bit_align, elem_type, subranges_arr, len(subranges)))

    def create_vector_type(self, elem_type: DIType, bit_size: int, bit_align: int, *subranges: MDNode) -> DIType:
        subranges_arr = (c_object_p * len(subranges))([x.ptr for x in subranges])

        return DIType(LLVMDIBuilderCreateVectorType(self, bit_size, bit_align, elem_type, subranges_arr, len(subranges)))

    def create_basic_type(self, name: str, bit_size: int, encoding: DWARFTypeEncoding, flags: DIFlags = DIFlags.ZERO) -> DIType:
        name_bytes = name.encode()

        return DIType(LLVMDIBuilderCreateBasicType(self, name_bytes, len(name_bytes), bit_size, encoding, flags))

    def create_pointer_type(self, elem_type: DIType, bit_size: int, bit_align: int, addr_space: int = 0, name: str = "") -> DIType:
        name_bytes = name.encode()

        return DIType(LLVMDIBuilderCreatePointerType(
            self, 
            elem_type, 
            bit_size, 
            bit_align, 
            addr_space, 
            name_bytes, 
            len(name_bytes)
        ))

    def create_struct_type(
        self,
        scope: DIScope,
        file: DIFile,
        name: str,
        line: int,
        bit_size: int,
        bit_align: int,
        *elements: DIType,
        derived_from: Optional[DIType] = None,
        vtable_holder: Optional[DIType] = None,
        runtime_ver: int = 0,
        unique_id: str = "",
        flags: DIFlags = DIFlags.ZERO 
    ) -> DIType:
        name_bytes = name.encode()
        unique_id_bytes = unique_id.encode() if unique_id else name_bytes

        return DIType(LLVMDIBuilderCreateStructType(
            self,
            scope,
            name_bytes,
            len(name_bytes),
            file,
            line,
            bit_size,
            bit_align,
            flags,
            derived_from,
            create_object_array(elements),
            len(elements),
            runtime_ver,
            vtable_holder,
            unique_id_bytes,
            len(unique_id_bytes)
        ))

    def create_member_type(
        self,
        scope: DIScope,
        file: DIFile,
        parent_type: DIType,
        name: str,
        line: int,
        bit_size: int,
        bit_align: int,
        bit_offset: int,
        flags: DIFlags = DIFlags.ZERO
    ) -> DIType:
        name_bytes = name.encode()

        return DIType(LLVMDIBuilderCreateMemberType(
            self,
            scope,
            name_bytes,
            len(name_bytes),
            file,
            line,
            bit_size,
            bit_align,
            bit_offset,
            flags,
            parent_type
        ))

    def create_static_member_type(
        self,
        scope: DIScope,
        file: DIFile,
        name: str,
        line: int,
        member_type: DIType,
        bit_align: int,
        const_init: Optional[MDNode] = None,
        flags: DIFlags = DIFlags.ZERO
    ) -> DIType:
        name_bytes = name.encode()

        return DIType(LLVMDIBuilderCreateStaticMemberType(
            self,
            scope,
            name_bytes,
            len(name_bytes),
            file,
            line,
            member_type,
            flags,
            const_init,
            bit_align
        ))

    def create_pointer_to_member_type(
        self,
        class_type: DIType,
        elem_type: DIType,
        bit_size: int,
        bit_align: int,
        flags: DIFlags = DIFlags.ZERO,
    ) -> DIType:
        return DIType(LLVMDIBuilderCreateMemberPointerType(self, elem_type, class_type, bit_size, bit_align, flags))

    def create_object_pointer_type(self, obj_type: DIType) -> DIType:
        return DIType(LLVMDIBuilderCreateObjectPointerType(self, obj_type))

    def create_qualified_type(self, qualifier: DWARFTypeQualifier, typ: DIType) -> DIType:
        return DIType(LLVMDIBuilderCreateQualifiedType(self, qualifier, typ))

    def create_typedef(self, scope: DIScope, file: DIFile, name: str, line: int, typ: DIType, bit_align: int) -> DIType:
        name_bytes = name.encode()

        return DIType(LLVMDIBuilderCreateTypedef(
            self, 
            typ, 
            name_bytes,
            len(name_bytes),
            file,
            line,
            scope,
            bit_align
        ))

    def create_subrange(self, lower_bound: int, upper_bound: int) -> MDNode:
        return MDNode(ptr=LLVMDIBuilderGetOrCreateSubrange(self, lower_bound, upper_bound))

    def create_array(self, *data: MDNode) -> MDNode:
        return MDNode(ptr=LLVMDIBuilderGetOrCreateArray(self, create_object_array(data), len(data)))

    def create_expression(self, *addr_ops: int) -> MDNode:
        return MDNode(ptr=LLVMDIBuilderCreateExpression(
            self, 
            (c_uint64 * len(addr_ops))([x for x in addr_ops]), 
            len(addr_ops))
        )

    def create_const_value(self, value: int) -> MDNode:
        return MDNode(ptr=LLVMDIBuilderCreateConstantValueExpression(self, value))

    def create_global_var_expr(
        self,
        scope: DIScope,
        file: DIFile,
        name: str,
        mangled_name: str,
        line: int,
        var_type: DIType,
        var_expr: MDNode,
        global_var_decl: MDNode,
        bit_align: int,
        internal: bool
    ) -> DIGlobalVarExpr:
        name_bytes = name.encode()
        mangled_bytes = mangled_name.encode()

        return DIGlobalVarExpr(
            self,
            scope,
            name_bytes,
            len(name_bytes),
            mangled_bytes,
            len(mangled_bytes),
            file,
            line,
            var_type,
            int(internal),
            var_expr,
            global_var_decl,
            bit_align
        )

    def create_param_variable(
        self, 
        scope: DIScope, 
        file: DIFile, 
        name: str, 
        arg_n: int,
        line: int, 
        param_type: DIType,
        survives_optimizations: bool = True,
        flags: DIFlags = DIFlags.ZERO
    ) -> DIVariable:
        name_bytes = name.encode()

        return DIVariable(
            self,
            scope,
            name_bytes,
            len(name_bytes),
            arg_n,
            file,
            line,
            param_type,
            int(survives_optimizations),
            flags,
        )
    
    def insert_debug_declare(
        self,
        var: Value,
        var_info: DIVariable,
        addr_expr: MDNode,
        debug_loc: DILocation,
        on: Instruction | BasicBlock
    ) -> CallInstruction:
        if isinstance(on, Instruction):
            return CallInstruction(LLVMDIBuilderInsertDeclareBefore(
                self,
                var,
                var_info,
                addr_expr,
                debug_loc,
                on
            ))
        else:
            return CallInstruction(LLVMDIBuilderInsertDeclareAtEnd(
                self,
                var,
                var_info,
                addr_expr,
                debug_loc,
                on
            ))

    def insert_debug_value(
        self,
        val: Value,
        var_info: DIVariable,
        addr_expr: MDNode,
        debug_loc: DILocation,
        on: Instruction | BasicBlock
    ) -> CallInstruction:
        if isinstance(on, Instruction):
            return CallInstruction(LLVMDIBuilderInsertDbgValueBefore(
                self,
                val,
                var_info,
                addr_expr,
                debug_loc,
                on
            ))
        else:
            return CallInstruction(LLVMDIBuilderInsertDbgValueAtEnd(
                self,
                val,
                var_info,
                addr_expr,
                debug_loc,
                on
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
def LLVMDIBuilderInsertDeclareBefore(builder: DIBuilder, storage: Value, var_info: DIVariable, expr: Metadata, debug_loc: DILocation, instr: Instruction) -> c_object_p:
    pass

@llvm_api
def LLVMDIBuilderInsertDeclareAtEnd(builder: DIBuilder, storage: Value, var_info: DIVariable, expr: Metadata, debug_loc: DILocation, block: BasicBlock) -> c_object_p:
    pass

@llvm_api
def LLVMDIBuilderInsertDbgValueBefore(builder: DIBuilder, val: Value, var_info: DIVariable, expr: Metadata, debug_loc: DILocation, instr: Instruction) -> c_object_p:
    pass

@llvm_api
def LLVMDIBuilderInsertDbgValueAtEnd(builder: DIBuilder, val: Value, var_info: DIVariable, expr: Metadata, debug_loc: DILocation, block: BasicBlock) -> c_object_p:
    pass

@llvm_api
def LLVMDIBuilderCreateParameterVariable(builder: DIBuilder, scope: DIScope, name: c_char_p, name_len: c_size_t, arg_no: c_uint, file: DIFile, line_no: c_uint, ty: DIType, always_preserve: c_enum, flags: DIFlags) -> c_object_p:
    pass

@llvm_api
def LLVMDIBuilderCreateNameSpace(builder: DIBuilder, parent_scope: DIScope, name: c_char_p, name_len: c_size_t, export_symbols: c_enum) -> c_object_p:
    pass

@llvm_api
def LLVMDIBuilderCreateImportedModuleFromNamespace(builder: DIBuilder, scope: DIScope, n_s: Metadata, file: DIFile, line: c_uint) -> c_object_p:
    pass

@llvm_api
def LLVMDIBuilderCreateImportedModuleFromAlias(builder: DIBuilder, scope: DIScope, imported_entity: Metadata, file: DIFile, line: c_uint, elements: POINTER(c_object_p), num_elements: c_uint) -> c_object_p:
    pass

@llvm_api
def LLVMDIBuilderCreateImportedModuleFromModule(builder: DIBuilder, scope: DIScope, m: Metadata, file: DIFile, line: c_uint, elements: POINTER(c_object_p), num_elements: c_uint) -> c_object_p:
    pass

@llvm_api
def LLVMDIBuilderCreateArrayType(builder: DIBuilder, size: c_uint64, align_in_bits: c_uint32, ty: DIType, subscripts: POINTER(c_object_p), num_subscripts: c_uint) -> c_object_p:
    pass

@llvm_api
def LLVMDIBuilderCreateVectorType(builder: DIBuilder, size: c_uint64, align_in_bits: c_uint32, ty: DIType, subscripts: POINTER(c_object_p), num_subscripts: c_uint) -> c_object_p:
    pass

@llvm_api
def LLVMDIBuilderCreateMemberPointerType(builder: DIBuilder, pointee_type: DIType, class_type: DIType, size_in_bits: c_uint64, align_in_bits: c_uint32, flags: DIFlags) -> c_object_p:
    pass

@llvm_api
def LLVMDIBuilderCreateObjectPointerType(builder: DIBuilder, type: DIType) -> c_object_p:
    pass

@llvm_api
def LLVMDIBuilderCreateSubroutineType(builder: DIBuilder, file: DIFile, param_types: POINTER(c_object_p), num_param_types: c_uint, flags: DIFlags) -> c_object_p:
    pass

@llvm_api
def LLVMDIBuilderCreateQualifiedType(builder: DIBuilder, qualifier: DWARFTypeQualifier, base_type: DIType) -> c_object_p:
    pass

@llvm_api
def LLVMDIBuilderCreateConstantValueExpression(builder: DIBuilder, value: c_uint64) -> c_object_p:
    pass