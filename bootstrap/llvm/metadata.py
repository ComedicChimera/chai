from typing import List, Optional
from ctypes import c_uint, c_char_p, c_size_t, POINTER, byref, c_uint32, c_uint64
from enum import Flag, auto

from . import *
from .value import Value

__all__ = [
    'Metadata',
    'MDString',
    'MDNode',
    'DIFlags',
    'DIFile',
    'DIScope',
    'DILocation',
    'DIVariable',
    'DIGlobalVarExpr',
    'DISubprogram',
    'DIType'
]

class Metadata(LLVMObject):
    class Kind(LLVMEnum):
        MD_STRING = auto()
        CONSTANT_AS_METADATA = auto()
        LOCAL_AS_METADATA = auto()
        DISTINCT_MD_OPERAND_PLACEHOLDER = auto()
        MD_TUPLE = auto()
        DI_LOCATION = auto()
        DI_EXPRESSION = auto()
        DI_GLOBAL_VARIABLE_EXPRESSION = auto()
        GENERIC_DI_NODE = auto()
        DI_SUBRANGE = auto()
        DI_ENUMERATOR = auto()
        DI_BASIC_TYPE = auto()
        DI_DERIVED_TYPE = auto()
        DI_COMPOSITE_TYPE = auto()
        DI_SUBROUTINE_TYPE = auto()
        DI_FILE = auto()
        DI_COMPILE_UNIT = auto()
        DI_SUBPROGRAM = auto()
        DI_LEXICAL_BLOCK = auto()
        DI_LEXICAL_BLOCK_FILE = auto()
        DI_NAMESPACE = auto()
        DI_MODULE = auto()
        DI_TEMPLATE_TYPE_PARAMETER = auto()
        DI_TEMPLATE_VALUE_PARAMETER = auto()
        DI_GLOBAL_VARIABLE = auto()
        DI_LOCAL_VARIABLE = auto()
        DI_LABEL = auto()
        DI_OBJ_C_PROPERTY = auto()
        DI_IMPORTED_ENTITY = auto()
        DI_MACRO = auto()
        DI_MACRO_FILE = auto()
        DI_COMMON_BLOCK = auto()
        DI_STRING_TYPE = auto()
        DI_GENERIC_SUBRANGE = auto()
        DI_ARG_LIST = auto()

    def __init__(self, ptr: c_object_p):
        super().__init__(ptr)

    @staticmethod
    def from_value(value: Value) -> 'Metadata':
        return Metadata(LLVMValueAsMetadata(value))

    def to_value(self) -> Value:
        return Value(LLVMMetadataAsValue(self))

    @property
    def is_md_string(self) -> bool:
        return bool(LLVMIsAMDString(self.to_value()))

    @property
    def is_md_node(self) -> bool:
        return bool(LLVMIsAMDNode(self.to_value()))

    @property
    def kind(self) -> Kind:
        return Metadata.Kind(LLVMGetMetadataKind(self))

class MDString(Metadata):
    def __init__(self, string: str, ptr: Optional[c_object_p] = None):
        if ptr:
            super().__init__(ptr)
        else:
            str_bytes = string.encode()
            super().__init__(LLVMMDStringInContext2(get_context(), str_bytes, len(str_bytes)))

    @staticmethod
    def from_metadata(md: Metadata) -> 'MDString':
        assert md.is_md_string

        return MDString("", md.ptr)

    @property
    def string(self) -> str:
        return str(LLVMGetMDString(self, byref(c_uint())), encoding='utf-8')

class MDNode(Metadata):
    def __init__(self, *mds: Metadata | Value, ptr: Optional[c_object_p] = None):
        if ptr:
            super().__init__(ptr)
        else:
            md_arr = (c_object_p * len(mds))(*(
                x.ptr if isinstance(x, Metadata) else Metadata.from_value(x).ptr
                for x in mds
            ))
            super().__init__(LLVMMDNodeInContext2(get_context(), md_arr, len(mds)))

    @staticmethod
    def from_metadata(md: Metadata) -> 'MDNode':
        assert md.is_md_node
        
        return MDNode(ptr=md.ptr)

    @property
    def operands(self) -> List[Value]:
        md_arr_len = LLVMGetMDNodeNumOperands(self)
        md_arr = (c_object_p * md_arr_len)()
        LLVMGetMDNodeNumOperands(self, md_arr)
        return [Value(x) for x in md_arr]

# ---------------------------------------------------------------------------- #

class DIFlags(Flag):
    ZERO = 0
    PRIVATE = auto()
    PROTECTED = auto()
    PUBLIC = 3
    FWD_DECL = auto()
    APPLE_BLOCK = auto()
    RESERVED_BIT4 = auto()
    VIRTUAL = auto()
    ARTIFICIAL = auto()
    EXPLICIT = auto()
    PROTOTYPED = auto()
    OBJC_CLASS_COMPLETE = auto()
    OBJECT_POINTER = auto()
    VECTOR = auto()
    STATIC_MEMBER = auto()
    LVALUE_REFERENCE = auto()
    RVALUE_REFERENCE = auto()
    RESERVED = auto()
    SINGLE_INHERITANCE = auto()
    MULTIPLE_INHERITANCE = auto()
    VIRTUAL_INHERITANCE = auto()
    INTRODUCED_VIRTUAL = auto()
    BIT_FIELD = auto()
    NO_RETURN = auto()
    TYPE_PASS_BY_VALUE = auto()
    TYPE_PASS_BY_REFERENCE = auto()
    ENUM_CLASS = auto()
    THUNK = auto()
    NON_TRIVIAL = auto()
    BIG_ENDIAN = auto()
    LITTLE_ENDIAN = auto()
    INDIRECT_VIRTUAL_BASE = (1 << 2) | (1 << 5)
    ACCESSIBILITY = PRIVATE | PROTECTED | PUBLIC
    PTR_TO_MEMBER_REP = SINGLE_INHERITANCE | MULTIPLE_INHERITANCE | VIRTUAL_INHERITANCE

    def from_param(self) -> int:
        return self.value

class DIFile(MDNode):
    def __init__(self, ptr: c_object_p):
        super().__init__(ptr=ptr)

    @property
    def directory(self) -> str:
        return str(LLVMDIFileGetDirectory(self, byref(c_uint())), encoding='utf-8')

    @property
    def file_name(self) -> str:
        return str(LLVMDIFileGetFilename(self, byref(c_uint())), encoding='utf-8')

    @property
    def source(self) -> str:
        return str(LLVMDIFileGetSource(self, byref(c_uint())), encoding='utf-8')

    def as_scope(self) -> 'DIScope':
        return DIScope(self.ptr)

class DIScope(MDNode):
    def __init__(self, ptr: c_object_p):
        super().__init__(ptr=ptr)

    @property
    def file(self) -> DIFile:
        return DIFile(LLVMDIScopeGetFile(self))

class DILocation(MDNode):
    def __init__(self, 
        scope: DIScope, 
        line: int, 
        col: int, 
        inlined_at: Optional['DILocation'] = None,
        ptr: Optional[c_object_p] = None
    ):
        if ptr:
            super().__init__(ptr=ptr)
        else:
            loc_ptr = LLVMDIBuilderCreateDebugLocation(
                get_context(),
                line,
                col,
                scope,
                inlined_at.ptr if inlined_at else None,
            )
            
            super().__init__(ptr=loc_ptr)

    @property
    def scope(self) -> DIScope:
        return DIScope(LLVMDILocationGetScope(self))

    @property
    def line(self) -> int:
        return LLVMDILocationGetLine(self)

    @property
    def col(self) -> int:
        return LLVMDILocationGetColumn(self)

    @property
    def inlined_at(self) -> Optional['DILocation']:
        loc_ptr = LLVMDILocationGetInlinedAt(self)
        if loc_ptr:
            return DILocation(None, 0, 0, ptr=loc_ptr)

        return None

class DIType(MDNode):
    def __init__(self, ptr: c_object_p):
        super().__init__(ptr=ptr)

    @property
    def name(self) -> str:
        return str(LLVMDITypeGetName(self, byref(c_uint())), encoding='utf-8')

    @property
    def bit_size(self) -> int:
        return LLVMDITypeGetSizeInBits(self)

    @property
    def bit_align(self) -> int:
        return LLVMDITypeGetAlignInBits(self)

    @property
    def bit_offset(self) -> int:
        return LLVMDITypeGetOffsetInBits(self)

    @property
    def line(self) -> int:
        return LLVMDITypeGetLine(self)

    @property
    def flags(self) -> DIFlags:
        return DIFlags(LLVMDITypeGetFlags(self))

class DIVariable(MDNode):
    def __init__(self, ptr: c_object_p):
        super().__init__(ptr=ptr)

    @property
    def file(self) -> DIFile:
        return DIFile(LLVMDIVariableGetFile(self))

    @property
    def scope(self) -> DIScope:
        return DIScope(LLVMDIVariableGetScope(self))

    @property
    def line(self) -> int:
        return LLVMDIVariableGetLine(self)

class DIGlobalVarExpr(MDNode):
    def __init__(self, ptr: c_object_p):
        super().__init__(ptr=ptr)

    @property
    def variable(self) -> DIVariable:
        return DIVariable(LLVMDIGlobalVariableExpressionGetVariable(self))

    @property
    def expr(self) -> Metadata:
        return Metadata(LLVMDIGlobalVariableExpressionGetExpression(self))

class DISubprogram(MDNode):
    def __init__(self, ptr: c_object_p):
        super().__init__(ptr=ptr)

    @property
    def line(self) -> int:
        return LLVMDISubprogramGetLine(self)

    def as_scope(self) -> DIScope:
        return DIScope(self.ptr)

# ---------------------------------------------------------------------------- #

@llvm_api
def LLVMMDStringInContext2(c: Context, str: c_char_p, s_len: c_size_t) -> c_object_p:
    pass

@llvm_api
def LLVMMDNodeInContext2(c: Context, mds: POINTER(c_object_p), count: c_size_t) -> c_object_p:
    pass

@llvm_api
def LLVMMetadataAsValue(c: Context, md: Metadata) -> c_object_p:
    pass

@llvm_api
def LLVMValueAsMetadata(val: Value) -> c_object_p:
    pass

@llvm_api
def LLVMGetMDString(v: Value, length: POINTER(c_uint)) -> c_char_p:
    pass

@llvm_api
def LLVMGetMDNodeNumOperands(v: Value) -> c_uint:
    pass

@llvm_api
def LLVMGetMDNodeOperands(v: Value, dest: POINTER(c_object_p)):
    pass

@llvm_api
def LLVMIsAMDString(val: Value) -> c_object_p:
    pass

@llvm_api
def LLVMIsAMDNode(val: Value) -> c_object_p:
    pass

@llvm_api
def LLVMGetMetadataKind(metadata: Metadata) -> c_enum:
    pass 

@llvm_api
def LLVMDISubprogramGetLine(subprogram: Metadata) -> c_uint:
    pass

@llvm_api
def LLVMDIVariableGetFile(var: Metadata) -> c_object_p:
    pass

@llvm_api
def LLVMDIVariableGetScope(var: Metadata) -> c_object_p:
    pass

@llvm_api
def LLVMDIVariableGetLine(var: Metadata) -> c_uint:
    pass

@llvm_api
def LLVMDIGlobalVariableExpressionGetVariable(g_v_e: Metadata) -> c_object_p:
    pass

@llvm_api
def LLVMDIGlobalVariableExpressionGetExpression(g_v_e: Metadata) -> c_object_p:
    pass

@llvm_api
def LLVMDITypeGetName(d_type: Metadata, length: POINTER(c_size_t)) -> c_char_p:
    pass

@llvm_api
def LLVMDITypeGetSizeInBits(d_type: Metadata) -> c_uint64:
    pass

@llvm_api
def LLVMDITypeGetOffsetInBits(d_type: Metadata) -> c_uint64:
    pass

@llvm_api
def LLVMDITypeGetAlignInBits(d_type: Metadata) -> c_uint32:
    pass

@llvm_api
def LLVMDITypeGetLine(d_type: Metadata) -> c_uint:
    pass

@llvm_api
def LLVMDITypeGetFlags(d_type: Metadata) -> c_enum:
    pass

@llvm_api
def LLVMDIBuilderCreateDebugLocation(ctx: Context, line: c_uint, column: c_uint, scope: DIScope, inlined_at: c_object_p) -> c_object_p:
    pass

@llvm_api
def LLVMDILocationGetLine(location: Metadata) -> c_uint:
    pass

@llvm_api
def LLVMDILocationGetColumn(location: Metadata) -> c_uint:
    pass

@llvm_api
def LLVMDILocationGetScope(location: Metadata) -> c_object_p:
    pass

@llvm_api
def LLVMDILocationGetInlinedAt(location: Metadata) -> c_object_p:
    pass

@llvm_api
def LLVMDIScopeGetFile(scope: Metadata) -> c_object_p:
    pass

@llvm_api
def LLVMDIFileGetDirectory(file: Metadata, len: POINTER(c_uint)) -> c_char_p:
    pass

@llvm_api
def LLVMDIFileGetFilename(file: Metadata, len: POINTER(c_uint)) -> c_char_p:
    pass

@llvm_api
def LLVMDIFileGetSource(file: Metadata, len: POINTER(c_uint)) -> c_char_p:
    pass