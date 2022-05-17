from typing import Dict, Iterator, Optional
from ctypes import c_uint, c_char_p, POINTER, byref, c_uint64
from enum import Enum, auto

from . import *
from .value import GlobalValue, Value
from .types import Type
from .ir import BasicBlock

class Attribute(LLVMObject):
    class Kind(LLVMEnum):
        ALWAYS_INLINE = 1
        ARG_MEM_ONLY = 2
        BUILTIN = 3
        COLD = 4
        CONVERGENT = 5
        DISABLE_SANITIZER_INSTRUMENTATION = 6
        HOT = 7
        IMM_ARG = 8
        IN_REG = 9
        INACCESSIBLE_MEM_ONLY = 10
        INACCESSIBLE_MEM_OR_ARG_MEM_ONLY = 11
        INLINE_HINT = 12
        JUMP_TABLE = 13
        MIN_SIZE = 14
        MUST_PROGRESS = 15
        NAKED = 16
        NEST = 17
        NO_ALIAS = 18
        NO_BUILTIN = 19
        NO_CALLBACK = 20
        NO_CAPTURE = 21
        NO_CF_CHECK = 22
        NO_DUPLICATE = 23
        NO_FREE = 24
        NO_IMPLICIT_FLOAT = 25
        NO_INLINE = 26
        NO_MERGE = 27
        NO_PROFILE = 28
        NO_RECURSE = 29
        NO_RED_ZONE = 30
        NO_RETURN = 31
        NO_SANITIZE_COVERAGE = 32
        NO_SYNC = 33
        NO_UNDEF = 34
        NO_UNWIND = 35
        NON_LAZY_BIND = 36
        NON_NULL = 37
        NULL_POINTER_IS_VALID = 38
        OPT_FOR_FUZZING = 39
        OPTIMIZE_FOR_SIZE = 40
        OPTIMIZE_NONE = 41
        READ_NONE = 42
        READ_ONLY = 43
        RETURNED = 44
        RETURNS_TWICE = 45
        S_EXT = 46
        SAFE_STACK = 47
        SANITIZE_ADDRESS = 48
        SANITIZE_HW_ADDRESS = 49
        SANITIZE_MEM_TAG = 50
        SANITIZE_MEMORY = 51
        SANITIZE_THREAD = 52
        SHADOW_CALL_STACK = 53
        SPECULATABLE = 54
        SPECULATIVE_LOAD_HARDENING = 55
        STACK_PROTECT = 56
        STACK_PROTECT_REQ = 57
        STACK_PROTECT_STRONG = 58
        STRICT_FP = 59
        SWIFT_ASYNC = 60
        SWIFT_ERROR = 61
        SWIFT_SELF = 62
        UW_TABLE = 63
        WILL_RETURN = 64
        WRITE_ONLY = 65
        Z_EXT = 66
        LAST_ENUM_ATTR = 66
        FIRST_TYPE_ATTR = 67
        BY_REF = 67
        BY_VAL = 68
        ELEMENT_TYPE = 69
        IN_ALLOCA = 70
        PREALLOCATED = 71
        STRUCT_RET = 72
        LAST_TYPE_ATTR = 72
        FIRST_INT_ATTR = 73
        ALIGNMENT = 73
        ALLOC_SIZE = 74
        DEREFERENCEABLE = 75
        DEREFERENCEABLE_OR_NULL = 76
        STACK_ALIGNMENT = 77
        V_SCALE_RANGE = 78

    class Variant(Enum):
        ENUM = auto()
        STRING = auto()
        TYPE = auto()

    _variant: Variant

    def __init__(self, **kwargs):
        if ptr := kwargs.get('ptr'):
            assert isinstance(ptr, c_object_p)

            super().__init__(ptr)

            if LLVMIsEnumAttribute(self) == 1:
                self._variant = Attribute.Variant.ENUM
            elif LLVMIsStringAttribute(self) == 1:
                self._variant = Attribute.Variant.STRING
            else:
                self._variant = Attribute.Variant.TYPE
        else:
            assert (kind := kwargs.get('kind'))

            ctx = kwargs.get('context', Context.global_ctx())

            if isinstance(kind, str):
                assert (value := kwargs.get('kind'))

                value_bytes = value.encode()
                super().__init__(LLVMCreateStringAttribute(ctx, kind, value_bytes, len(value_bytes)))
                self._variant = Attribute.Variant.STRING
            else:
                assert isinstance(kind, Attribute.Kind)

                if value := kwargs.get('value'):
                    if isinstance(value, Type):
                        super().__init__(LLVMCreateTypeAttribute(ctx, kind, value))
                        self._variant = Attribute.Variant.TYPE
                        return

                    super().__init__(LLVMCreateEnumAttribute(ctx, kind, value))
                else:
                    super().__init__(LLVMCreateEnumAttribute(ctx, kind, 0))

                self._variant = Attribute.Variant.ENUM

    @property
    def kind(self) -> Kind | str:
        match self._variant:
            case Attribute.Variant.ENUM | Attribute.Variant.TYPE:
                return Attribute.Kind(LLVMGetEnumAttributeKind(self))
            case Attribute.Variant.STRING:
                length = c_uint()
                return str(LLVMGetStringAttributeKind(self, byref(length)), encoding='utf-8')

    @property
    def value(self) -> int | str | Type:
        match self._variant:
            case Attribute.Variant.ENUM:
                return LLVMGetEnumAttributeValue(self)
            case Attribute.Variant.TYPE:
                ptr = LLVMGetTypeAttributeValue(self)
                return Type(ptr)
            case Attribute.Variant.STRING:
                length = c_uint()
                return str(LLVMGetStringAttributeValue(self, byref(length)), encoding='utf-8')

    @property
    def variant(self) -> Variant:
        return self._variant 

class CallConv(LLVMEnum):
    C = 0
    FAST = 8
    COLD = 9
    GHC = 10
    HIPE = 11
    WEBKIT_JS = 12
    ANY_REG = 13
    PRESERVE_MOST = 14
    PRESERVEALL = 15
    SWIFT = 16
    CXX_FAST_TLS = 17
    TAIL = 18,
    CFGUARD_CHECK = 19
    SWIFTTAIL = 20
    FIRSTTARGETCC = 64
    X86_STDCALL = 64
    X86_FASTCALL = 65
    ARM_APCS = 66
    ARM_AAPCS = 67
    ARM_AAPCS_VFP = 68,
    MSP430_INTR = 69
    X86_THISCALL = 70
    PTX_KERNEL = 71
    PTX_DEVICE = 72
    SPIR_FUNC = 75
    SPIR_KERNEL = 76
    INTEL_OCL_BI = 77
    X86_64_SYSV = 78
    WIN64 = 79
    X86_VECTORCALL = 80
    HHVM = 81
    HHVM_C = 82
    X86_INTR = 83
    AVR_INTR = 84
    AVR_SIGNAL = 85
    AVR_BUILTIN = 86
    AMDGPU_VS = 87
    AMDGPU_GS = 88
    AMDGPU_PS = 89
    AMDGPU_CS = 90,
    AMDGPU_KERNEL = 91
    X86_REGCALL = 92
    AMDGPU_HS = 93
    MSP430_BUILTIN = 94
    AMDGPU_LS = 95
    AMDGPU_ES = 96
    AARCH64_VECTORCALL = 97
    AARCH64_SVE_VECTORCALL = 98
    WASM_EMSCRIPTEN_INVOKE = 99
    AMDGPU_GFX = 100
    M68K_INTR = 101
    MAXID = 1023

class AttributeSet:
    KeyType = Attribute.Kind | str 

    _func: 'Function'
    _ndx: int
    _attr_dict: Dict[KeyType, Attribute]

    def __init__(self, func: 'Function', ndx: int):
        self._func = func
        self._ndx = ndx

        attr_arr_len = LLVMGetAttributeCountAtIndex(func, ndx)
        attr_arr = (c_object_p * attr_arr_len)()
        LLVMGetAttributesAtIndex(func, ndx, attr_arr)

        self._attr_dict = {attr.kind: attr for attr in map(lambda p: Attribute(ptr=p), attr_arr)}

    def __len__(self) -> int:
        return len(self._attr_dict)

    def __getitem__(self, key: KeyType) -> Attribute:
        return self._attr_dict(key)

    def __iter__(self) -> Iterator[Attribute]:
        return self._attr_dict.values()

    def add(self, attr: Attribute):
        self._attr_dict[attr.kind] = attr
        LLVMAddAttributeAtIndex(self._func, self._ndx, attr)

    def remove(self, key: KeyType):
        del self._attr_dict[key]

        if isinstance(key, str):
            key_bytes = key.encode()
            LLVMRemoveStringAttributeAtIndex(self._func, self._ndx, key_bytes, len(key_bytes))
        else:
            LLVMRemoveEnumAttributeAtIndex(self._func, self._ndx, key)        

class FuncParam(Value):
    _attr_set: AttributeSet

    def __init__(self, ptr: c_object_p, func: 'Function', attr_ndx: int):
        super().__init__(ptr)

        self._attr_set = AttributeSet(func, attr_ndx)

    @property
    def attr_set(self):
        return self._attr_set

class BasicBlockList:
    _func: 'Function'

    def __init__(self, func: 'Function'):
        self._func = func

    def __iter__(self) -> Iterator[BasicBlock]:
        bb_ptr = LLVMGetFirstBasicBlock(self._func)

        while bb_ptr:
            yield BasicBlock(bb_ptr)

            bb_ptr = LLVMGetNextBasicBlock(bb_ptr)

    def __reversed__(self) -> Iterator[BasicBlock]:
        bb_ptr = LLVMGetLastBasicBlock(self._func)

        while bb_ptr:
            yield BasicBlock(bb_ptr)

            bb_ptr = LLVMGetPreviousBasicBlock(bb_ptr)

    def __len__(self) -> int:
        return LLVMCountBasicBlocks(self._func)

    def first(self) -> Optional[BasicBlock]:
        bb_ptr = LLVMGetFirstBasicBlock(self._func)

        if bb_ptr:
            return BasicBlock(bb_ptr)

        return None

    def last(self) -> Optional[BasicBlock]:
        bb_ptr = LLVMGetLastBasicBlock(self._func)

        if bb_ptr:
            return BasicBlock(bb_ptr)

        return None

    def entry(self) -> Optional[BasicBlock]:
        bb_ptr = LLVMGetEntryBasicBlock(self._func)

        if bb_ptr:
            return BasicBlock(bb_ptr)

        return None

    def append(self, name: str) -> BasicBlock:
        return BasicBlock(LLVMAppendBasicBlockInContext(
            get_context(),
            self._func,
            name.encode(),
        ))

    def insert(self, before: BasicBlock, name: str) -> BasicBlock:
        return BasicBlock(LLVMInsertBasicBlockInContext(
            get_context(),
            before,
            name.encode()
        ))

    def remove(self, bb: BasicBlock):
        LLVMDeleteBasicBlock(bb)

class Function(GlobalValue):
    _attr_set: AttributeSet
    _rt_attr_set: AttributeSet

    class FuncParamsView:
        _func: 'Function'

        def __init__(self, func: 'Function'):
            self._func = func

        def __len__(self) -> int:
            return LLVMCountParams(self._func)

        def __getitem__(self, ndx: int) -> FuncParam:
            if 0 <= ndx < len(self):
                return FuncParam(LLVMGetParam(self._func, ndx), self._func, ndx + 1)
            else:
                raise IndexError()

        def __iter__(self) -> Iterator[FuncParam]:
            for i in range(len(self)):
                yield FuncParam(LLVMGetParam(self._func, i), self._func, i + 1)

    def __init__(self, ptr: c_object_p):
        super().__init__(ptr)

        self._attr_set = AttributeSet(self, 0)
        self._rt_attr_set = AttributeSet(self, -1)

    @property
    def call_conv(self) -> CallConv:
        return CallConv(LLVMGetFunctionCallConv(self))

    @call_conv.setter
    def call_conv(self, cc: CallConv):
        LLVMSetFunctionCallConv(self, cc)

    @property
    def params(self) -> FuncParamsView:
        return Function.FuncParamsView(self)

    @property
    def attr_set(self) -> AttributeSet:
        return self._attr_set

    @property
    def rt_attr_set(self) -> AttributeSet:
        return self._rt_attr_set

    @property
    def basic_blocks(self) -> BasicBlockList:
        return BasicBlockList(self)

# ---------------------------------------------------------------------------- #

@llvm_api
def LLVMIsEnumAttribute(attr: Attribute) -> c_enum:
    pass

@llvm_api
def LLVMCreateEnumAttribute(ctx: Context, kind: Attribute.Kind, val: c_uint64) -> c_object_p:
    pass

@llvm_api
def LLVMGetEnumAttributeKind(attr: Attribute) -> c_enum:
    pass

@llvm_api
def LLVMGetEnumAttributeValue(attr: Attribute) -> c_uint64:
    pass

@llvm_api
def LLVMIsStringAttribute(attr: Attribute) -> c_enum:
    pass

@llvm_api
def LLVMCreateStringAttribute(ctx: Context, kind: c_char_p, klen: c_uint, val: c_char_p, vlen: c_uint) -> c_object_p:
    pass

@llvm_api
def LLVMGetStringAttributeKind(attr: Attribute, length: POINTER(c_uint)) -> c_char_p:
    pass

@llvm_api
def LLVMGetStringAttributeValue(attr: Attribute, length: POINTER(c_uint)) -> c_char_p:
    pass

@llvm_api
def LLVMIsTypeAttribute(attr: Attribute) -> c_enum:
    pass

@llvm_api
def LLVMCreateTypeAttribute(ctx: Context, kind: Attribute.Kind, val: Type) -> c_object_p:
    pass

@llvm_api
def LLVMGetTypeAttributeValue(attr: Attribute) -> c_object_p:
    pass

@llvm_api
def LLVMGetFunctionCallConv(func: Function) -> c_uint:
    pass

@llvm_api
def LLVMSetFunctionCallConv(func: Function, cc: CallConv):
    pass

@llvm_api
def LLVMCountParams(func: Function) -> c_uint:
    pass

@llvm_api
def LLVMGetParam(func: Function, ndx: c_uint) -> c_object_p:
    pass

@llvm_api
def LLVMGetAttributeCountAtIndex(func: Function, ndx: c_uint) -> c_uint:
    pass

@llvm_api
def LLVMGetAttributesAtIndex(func: Function, ndx: c_uint, attrs: POINTER(c_object_p)):
    pass

@llvm_api
def LLVMAddAttributeAtIndex(func: Function, ndx: c_uint, attr: Attribute):
    pass

@llvm_api
def LLVMRemoveEnumAttributeAtIndex(func: Function, ndx: c_uint, kind: Attribute.Kind):
    pass

@llvm_api
def LLVMRemoveStringAttributeAtIndex(func: Function, ndx: c_uint, kind: c_char_p, length: c_uint):
    pass

@llvm_api
def LLVMGetFirstBasicBlock(func: Function) -> c_object_p:
    pass

@llvm_api
def LLVMGetNextBasicBlock(bb: BasicBlock) -> c_object_p:
    pass

@llvm_api
def LLVMGetLastBasicBlock(func: Function) -> c_object_p:
    pass

@llvm_api
def LLVMGetPreviousBasicBlock(bb: BasicBlock) -> c_object_p:
    pass

@llvm_api
def LLVMGetEntryBasicBlock(func: Function) -> c_object_p:
    pass

@llvm_api
def LLVMCountBasicBlocks(func: Function) -> c_uint:
    pass

@llvm_api
def LLVMAppendBasicBlockInContext(ctx: Context, func: Function, name: c_char_p) -> c_object_p:
    pass

@llvm_api
def LLVMInsertBasicBlockInContext(ctx: Context, bb: BasicBlock, name: c_char_p) -> c_object_p:
    pass

@llvm_api
def LLVMDeleteBasicBlock(bb: BasicBlock) -> c_object_p:
    pass