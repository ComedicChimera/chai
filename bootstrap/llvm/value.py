from ctypes import c_char_p, c_double, c_size_t, c_int, POINTER, byref, c_uint, c_ulonglong
from enum import auto
from typing import Iterator

from . import LLVMObject, LLVMEnum, llvm_api, c_object_p, c_enum
from .types import Int1Type, IntegerType, Type

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
    def is_constant(self) -> bool:
        return bool(LLVMIsConstant(self))

    @property
    def is_undef(self) -> bool:
        return bool(LLVMIsUndef(self))

    @property
    def is_poison(self) -> bool:
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

# ---------------------------------------------------------------------------- #

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
