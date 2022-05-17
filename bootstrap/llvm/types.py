from ctypes import c_int, c_uint, POINTER
from enum import auto
from typing import List, Optional

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
            param_arr_type = c_object_p * len(param_types)
            param_arr = param_arr_type(*(x.ptr for x in param_types)) 

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
def LLVMInt1Type() -> c_object_p:
    pass

@llvm_api
def LLVMInt8Type() -> c_object_p:
    pass

@llvm_api
def LLVMInt16Type() -> c_object_p:
    pass

@llvm_api
def LLVMInt32Type() -> c_object_p:
    pass

@llvm_api
def LLVMInt64Type() -> c_object_p:
    pass

@llvm_api
def LLVMFloatType() -> c_object_p:
    pass

@llvm_api
def LLVMDoubleType() -> c_object_p:
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
def LLVMVoidType() -> c_object_p:
    pass

@llvm_api
def LLVMLabelType() -> c_object_p:
    pass

# ---------------------------------------------------------------------------- #

# Utility Type Constants

Int1Type = IntegerType(LLVMInt1Type())
Int8Type = IntegerType(LLVMInt8Type())
Int16Type = IntegerType(LLVMInt16Type())
Int32Type = IntegerType(LLVMInt32Type())
Int64Type = IntegerType(LLVMInt64Type())

FloatType = Type(LLVMFloatType())
DoubleType = Type(LLVMDoubleType())

VoidType = Type(LLVMVoidType())
LabelType = Type(LLVMLabelType())
