from ctypes import c_char_p, c_double, c_size_t, c_int, POINTER, byref, c_uint, c_ulonglong
from enum import auto
from typing import Dict, Iterator, Set

from . import LLVMObject, LLVMEnum, llvm_api, c_object_p, c_enum
from .types import Int1Type, IntegerType, Type
from .context import Attribute

class Value(LLVMObject):
    class Kind(LLVMEnum):
        ARGUMENT = auto()
        BASIC_BLOCK = auto()
        MEMORY_USE = auto()
        MEMORY_DEF = auto()
        MEMORY_PHI = auto()
        FUNCTION = auto()
        GLOBAL_ALIAS = auto()
        GLOBAL_IFUNC = auto()
        GLOBAL_VARIABLE_VAlUE = auto()
        BLOCK_ADDRESS = auto()
        CONSTANT_EXPR = auto()
        CONSTANT_ARRAY = auto()
        CONSTANT_STRUCT = auto()
        CONSTANT_VECTOR = auto()
        UNDEF_VALUE = auto()
        CONSTANT_AGGREGATE_ZERO = auto()
        CONSTANT_DATA_ARRAY = auto()
        CONSTANT_DATA_VECTOR = auto()
        CONSTANT_INT = auto()
        CONSTANT_FP = auto()
        CONSTANT_POINTER_NULL = auto()
        CONSTANT_TOKEN_NONE = auto()
        METADATA_AS_VALUE = auto()
        INLINE_ASM = auto()
        INSTRUCTION = auto()
        POISON_VALUE = auto()

    def __init__(self, ptr: c_object_p):
        super().__init__(ptr)

    @property
    def type(self) -> Type:
        return Type(LLVMTypeOf(self))

    @property
    def kind(self) -> Kind:
        return Value.Kind(LLVMGetValueKind(self))

    @property
    def name(self) -> str:
        length = c_size_t()
        return str(LLVMGetValueName2(self, byref(length)), encoding='utf-8')

    @name.setter
    def name(self, new_name: str):
        LLVMSetValueName2(self, new_name.encode(), len(new_name))

    @property
    def constant(self) -> bool:
        return bool(LLVMIsConstant(self))

    @property
    def undef(self) -> bool:
        return bool(LLVMIsUndef(self))

    @property
    def poison(self) -> bool:
        return bool(LLVMIsPoison(self))

    def replace(self, new_value: 'Value'):
        LLVMReplaceAllUsesWith(self, new_value)

    def dump(self):
        LLVMDumpValue(self)

class UserValue(Value):
    class OperandArray:
        val: 'UserValue'

        def __init__(self, val: 'UserValue'):
            self.val = val

        def __len__(self) -> int:
            return LLVMGetNumOperands(self.val)

        def __getitem__(self, ndx: int) -> Value:
            if 0 <= ndx < len(self):
                return Value(LLVMGetOperand(self.val, ndx))
            else:
                raise IndexError()

        def __setitem__(self, ndx: int, operand: Value):
            if 0 <= ndx < len(self):
                LLVMSetOperand(self.val, ndx, operand)
            else:
                raise IndexError()

        def __iter__(self) -> Iterator[Value]:
            for i in range(len(self)):
                yield Value(LLVMGetOperand(self.val, i))            

    def __init__(self, ptr: c_object_p):
        super().__init__(ptr)

    @property
    def operands(self) -> OperandArray:
        return UserValue.OperandArray(self)

# ---------------------------------------------------------------------------- #

class Constant(Value):
    def __init__(self, ptr: c_object_p):
        super().__init__(ptr)

    @staticmethod
    def Null(typ: Type) -> 'Constant':
        return Constant(LLVMConstNull(typ))

    @staticmethod
    def Undef(typ: Type) -> 'Constant':
        return Constant(LLVMGetUndef(typ))

    @staticmethod
    def Poison(typ: Type) -> 'Constant':
        return Constant(LLVMGetPoison(typ))

    @staticmethod
    def Nullptr(typ: Type) -> 'Constant':
        return Constant(LLVMConstPointerNull(typ))

    @staticmethod
    def Int(int_typ: IntegerType, n: int, signed: bool = False) -> 'Constant':
        return Constant(LLVMConstInt(int_typ, n, int(signed)))

    @staticmethod
    def Real(float_typ: Type, n: float) -> 'Constant':
        return Constant(LLVMConstReal(float_typ, n))

    @staticmethod
    def Bool(b: bool) -> 'Constant':
        return Constant.Int(Int1Type, int(b))

    @property
    def null(self) -> bool:
        return bool(LLVMIsNull(self))

class Linkage(LLVMEnum):
    EXTERNAL = auto()
    AVAILABLE_EXTERNALLY = auto()
    LINK_ONCE_ANY = auto()
    LINK_ONCE_ODR = auto()
    _LINK_ONCE_ODE_AUTO_HIDE = auto()
    WEAK_ANY = auto()
    WEAK_ODR = auto()
    APPENDING = auto()
    INTERNAL = auto()
    PRIVATE = auto()
    _DLL_IMPORT = auto()
    _DLL_EXPORT = auto()
    EXTERNAL_WEAK = auto()
    _GHOST = auto()
    COMMON = auto()
    LINKER_PRIVATE = auto()
    LINKER_PRIVATE_WEAK = auto()

class Visibility(LLVMEnum):
    DEFAULT = auto()
    HIDDEN = auto()
    PROTECTED = auto()

class DLLStorageClass(LLVMEnum):
    DEFAULT = auto()
    DLL_IMPORT = auto()
    DLL_EXPORT = auto()

class UnnamedAddr(LLVMEnum):
    NONE = auto()
    LOCAL = auto()
    GLOBAL = auto()

class GlobalValue(Value):
    def __init__(self, ptr: c_object_p):
        super().__init__(ptr)  

    @property
    def declaration(self) -> bool:
        return bool(LLVMIsDeclaration(self))

    @property
    def linkage(self) -> Linkage:
        return Linkage(LLVMGetLinkage(self))

    @linkage.setter
    def linkage(self, new_linkage: Linkage):
        LLVMSetLinkage(self, new_linkage)

    @property
    def visibility(self) -> Visibility:
        return Visibility(LLVMGetVisibility(self))

    @visibility.setter
    def visibility(self, new_vis: Visibility):
        LLVMSetVisibility(self, new_vis)

    @property
    def dll_storage_class(self) -> DLLStorageClass:
        return DLLStorageClass(LLVMGetDLLStorageClass(self))

    @dll_storage_class.setter
    def dll_storage_class(self, new_dsc: DLLStorageClass):
        LLVMSetDLLStorageClass(self, new_dsc)

    @property
    def unnamed_addr(self) -> UnnamedAddr:
        return UnnamedAddr(LLVMGetUnnamedAddress(self))
        
    @unnamed_addr.setter
    def unnamed_addr(self, ua: UnnamedAddr):
        LLVMSetUnnamedAddress(self, ua)

    @property
    def type(self) -> Type:
        return Type(LLVMGlobalGetValueType(self))

class Aligned(LLVMObject):
    @property
    def align(self) -> int:
        return LLVMGetAlignment(self)

    @align.setter
    def align(self, new_align: int):
        LLVMSetAlignment(self, new_align)

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

    def delete(self):
        LLVMDeleteFunction(self)

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

# ---------------------------------------------------------------------------- #

@llvm_api
def LLVMTypeOf(val: Value) -> c_object_p:
    pass

@llvm_api
def LLVMGetValueKind(val: Value) -> c_enum:
    pass

@llvm_api
def LLVMGetValueName2(val: Value, length: POINTER(c_size_t)) -> c_char_p:
    pass

@llvm_api
def LLVMSetValueName2(val: Value, name: c_char_p, length: c_size_t):
    pass

@llvm_api
def LLVMDumpValue(val: Value):
    pass

@llvm_api
def LLVMReplaceAllUsesWith(old: Value, new: Value):
    pass

@llvm_api
def LLVMIsConstant(val: Value) -> c_enum:
    pass

@llvm_api
def LLVMIsUndef(val: Value) -> c_enum:
    pass

@llvm_api
def LLVMIsPoison(val: Value) -> c_enum:
    pass

@llvm_api
def LLVMGetOperand(val: UserValue, ndx: c_uint) -> c_object_p:
    pass

@llvm_api
def LLVMSetOperand(val: UserValue, ndx: c_uint, operand: Value):
    pass

@llvm_api
def LLVMGetNumOperands(val: UserValue) -> c_int:
    pass

@llvm_api
def LLVMConstNull(typ: Type) -> c_object_p:
    pass

@llvm_api
def LLVMGetUndef(typ: Type) -> c_object_p:
    pass

@llvm_api
def LLVMGetPoison(typ: Type) -> c_object_p:
    pass

@llvm_api
def LLVMIsNull(const: Constant) -> c_enum:
    pass

@llvm_api
def LLVMConstPointerNull(typ: Type) -> c_object_p:
    pass

@llvm_api
def LLVMConstInt(typ: Type, n: c_ulonglong, signed: c_enum) -> c_object_p:
    pass

@llvm_api
def LLVMConstReal(typ: Type, n: c_double) -> c_object_p:
    pass

@llvm_api
def LLVMIsDeclaration(val: GlobalValue) -> c_enum:
    pass

@llvm_api
def LLVMGetLinkage(val: GlobalValue) -> c_enum:
    pass

@llvm_api
def LLVMSetLinkage(val: GlobalValue, linkage: Linkage):
    pass

@llvm_api
def LLVMGetVisibility(val: GlobalValue) -> c_enum:
    pass

@llvm_api
def LLVMSetVisibility(val: GlobalValue, vis: Visibility):
    pass

@llvm_api
def LLVMGetDLLStorageClass(val: GlobalValue) -> c_enum:
    pass

@llvm_api
def LLVMSetDLLStorageClass(val: GlobalValue, dsc: DLLStorageClass):
    pass

@llvm_api
def LLVMGetUnnamedAddress(val: GlobalValue) -> c_enum:
    pass

@llvm_api
def LLVMSetUnnamedAddress(val: GlobalValue, ua: UnnamedAddr):
    pass

@llvm_api
def LLVMGlobalGetValueType(val: GlobalValue) -> c_object_p:
    pass

@llvm_api
def LLVMGetAlignment(val: Value) -> c_uint:
    pass

@llvm_api
def LLVMSetAlignment(val: Value, align: c_uint):
    pass

@llvm_api
def LLVMDeleteFunction(func: Function):
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
