from ctypes import c_uint, POINTER
from enum import auto
from typing import List, Optional, Callable

from . import *

class Type(LLVMObject):
    class Kind(LLVMEnum):
        VOID = auto()
        HALF = auto()
        FLOAT = auto()
        DOUBLE = auto()
        X86_FP80 = auto()
        FP128 = auto()
        PPC_FP128 = auto()
        LABEL = auto()
        INTEGER = auto()
        FUNCTION = auto()
        STRUCT  = auto()
        ARRAY = auto()
        POINTER = auto()
        VECTOR = auto()
        METADATA = auto()
        X86_MMX = auto()
        TOKEN = auto()
        SCALABLE_VECTOR = auto()
        B_FLOAT = auto()
        X86_AMX  = auto()

    def __init__(self, ptr: c_object_p):
        super().__init__(ptr)

    @property
    def kind(self) -> Kind:
        return Type.Kind(LLVMGetTypeKind(self))

    @property
    def sized(self) -> bool:
        return bool(LLVMTypeIsSized(self))

    def dump(self):
        LLVMDumpType(self)

class IntegerType(Type):
    def __init__(self, ptr: c_object_p):
        super().__init__(ptr)

    @staticmethod
    def from_type(typ: Type) -> 'IntegerType':
        assert typ.kind == Type.Kind.INTEGER

        return IntegerType(typ.ptr)

    @property
    def width(self) -> int:
        return LLVMGetIntTypeWidth(self)

class FunctionType(Type):
    def __init__(self, param_types: List[Type], rt_type: Type, is_var_arg: bool = False, ptr: Optional[c_object_p] = None):
        if ptr:
            super().__init__(ptr)
            return

        if len(param_types) == 0:
            param_arr = None
        else:
            param_arr = create_object_array(param_types)

        ptr = LLVMFunctionType(rt_type, param_arr, len(param_types), int(is_var_arg))
        super().__init__(ptr)

    @staticmethod
    def from_type(typ: Type) -> 'FunctionType':
        assert typ.kind == Type.Kind.FUNCTION

        return FunctionType(None, None, None, typ.ptr)
    
    @property
    def var_arg(self) -> bool:
        return LLVMIsFunctionVarArg(self)

    @property
    def rt_type(self) -> Type:
        return Type(LLVMGetReturnType(self))

    @property
    def param_types(self) -> List[Type]:
        num_params = LLVMCountParamTypes(self)

        if num_params == 0:
            return []

        param_arr_type = c_object_p * num_params
        param_arr = param_arr_type()
        LLVMGetParamTypes(self, param_arr)

        return [Type(x) for x in param_arr]

class PointerType(Type):
    def __init__(self, elem_type: Type, addr_space: int = 0, ptr: Optional[c_object_p] = None):
        if ptr:
            super().__init__(ptr)
            return

        super().__init__(LLVMPointerType(elem_type, addr_space))

    @staticmethod
    def from_type(typ: Type) -> 'PointerType':
        assert typ.kind == Type.Kind.POINTER

        return PointerType(None, ptr=typ.ptr) 

    @property
    def elem_type(self) -> Type:
        return Type(LLVMGetElementType(self))

    @property
    def addr_space(self) -> int:
        return LLVMGetPointerAddressSpace(self)

# ---------------------------------------------------------------------------- #

@llvm_api
def LLVMGetTypeKind(typ: Type) -> c_enum:
    pass

@llvm_api
def LLVMTypeIsSized(typ: Type) -> c_enum:
    pass

@llvm_api
def LLVMDumpType(typ: Type):
    pass

@llvm_api
def LLVMGetIntTypeWidth(int_type: IntegerType) -> c_uint:
    pass

@llvm_api
def LLVMInt1TypeInContext() -> c_object_p:
    pass

@llvm_api
def LLVMInt8TypeInContext() -> c_object_p:
    pass

@llvm_api
def LLVMInt16TypeInContext() -> c_object_p:
    pass

@llvm_api
def LLVMInt32TypeInContext() -> c_object_p:
    pass

@llvm_api
def LLVMInt64TypeInContext() -> c_object_p:
    pass

@llvm_api
def LLVMFloatTypeInContext() -> c_object_p:
    pass

@llvm_api
def LLVMDoubleTypeInContext() -> c_object_p:
    pass

@llvm_api
def LLVMFunctionType(
    rt_type: Type, 
    param_types: POINTER(c_object_p), 
    param_count: c_uint, 
    is_arg: c_enum
) -> c_object_p:
    pass

@llvm_api
def LLVMIsFunctionVarArg(func_typ: FunctionType) -> c_enum:
    pass

@llvm_api
def LLVMGetReturnType(func_typ: FunctionType) -> c_object_p:
    pass

@llvm_api
def LLVMCountParamTypes(func_typ: FunctionType) -> c_uint:
    pass

@llvm_api
def LLVMGetParamTypes(func_typ: FunctionType, dest: POINTER(c_object_p)):
    pass

@llvm_api
def LLVMPointerType(elem_typ: Type, addr_space: c_uint) -> c_object_p:
    pass

@llvm_api
def LLVMGetElementType(typ: Type) -> c_object_p:
    pass

@llvm_api
def LLVMGetPointerAddressSpace(ptr_typ: PointerType) -> c_uint:
    pass

@llvm_api
def LLVMVoidTypeInContext(ctx: Context) -> c_object_p:
    pass

@llvm_api
def LLVMLabelTypeInContext(ctx: Context) -> c_object_p:
    pass

@llvm_api
def LLVMMetadataTypeInContext(ctx: Context) -> c_object_p:
    pass

# ---------------------------------------------------------------------------- #

# Utility Type Constructors

def Int1Type() -> IntegerType:
    return IntegerType(LLVMInt1TypeInContext(get_context()))

def Int8Type() -> IntegerType:
    return IntegerType(LLVMInt8TypeInContext(get_context()))

def Int16Type() -> IntegerType:
    return IntegerType(LLVMInt16TypeInContext(get_context()))

def Int32Type() -> IntegerType:
    return IntegerType(LLVMInt32TypeInContext(get_context()))

def Int64Type() -> IntegerType:
    return IntegerType(LLVMInt64TypeInContext(get_context()))

def FloatType() -> Type:
    return Type(LLVMFloatTypeInContext(get_context()))

def DoubleType() -> Type:
    return Type(LLVMDoubleTypeInContext(get_context()))

def VoidType() -> Type:
    return Type(LLVMVoidTypeInContext(get_context()))

def LabelType() -> Type:
    return Type(LLVMLabelTypeInContext(get_context()))

def MetadataType() -> Type:
    return Type(LLVMMetadataTypeInContext(get_context()))