from ctypes import c_size_t, byref, c_char_p, POINTER
from typing import Optional, Iterator, Tuple

from . import *
from .types import FunctionType
from .ir import Function

class Module(LLVMObject):
    def __init__(self, name: str):
        ctx = get_context()

        super().__init__(LLVMModuleCreateWithNameInContext(name.encode(), ctx))  

        ctx.take_ownership(self)

    def dispose(self):
        LLVMDisposeModule(self)

    @property
    def name(self) -> str:
        length = c_size_t()
        return str(LLVMGetModuleIdentifier(self, byref(length)), encoding='utf-8')

    @name.setter
    def name(self, new_name: str):  
        name_bytes = new_name.encode()  
        LLVMSetModuleIdentifier(self, name_bytes, len(name_bytes))

    @property
    def data_layout(self) -> TargetData:
        return TargetData(LLVMGetModuleDataLayout(self))

    @data_layout.setter
    def data_layout(self, data_layout: TargetData):
        LLVMSetModuleDataLayout(self, data_layout)

    @property
    def target_triple(self) -> str:
        return str(LLVMGetTarget(self), encoding='utf-8')

    @target_triple.setter
    def target_triple(self, target: str):
        LLVMSetTarget(self, target.encode())

    @property
    def context(self) -> Context:
        return Context(LLVMGetModuleContext(self))

    def verify(self) -> Tuple[str, bool]:
        p = c_char_p()
        ok = LLVMVerifyModule(self, 2, byref(p))
        
        if ok:
            return "", True

        return str(p.contents, encoding='utf-8'), False

    def dump(self):
        LLVMDumpModule(self)

    def add_function(self, name: str, func_type: FunctionType) -> Function:
        return Function(LLVMAddFunction(self, name.encode(), func_type))

    def get_function_by_name(self, name: str) -> Optional[Function]:
        func_ptr = LLVMGetNamedFunction(self, name.encode())
        
        if func_ptr:
            return Function(func_ptr)
        
        return None

    def delete_function(self, func: Function):
        LLVMDeleteFunction(func)

    class _Functions:
        _mod: 'Module'

        def __init__(self, mod: 'Module'):
            self._mod = mod

        def __iter__(self) -> Iterator[Function]:
            func_ptr = LLVMGetFirstFunction(self._mod)

            while func_ptr:
                yield Function(func_ptr)
                func_ptr = LLVMGetNextFunction(func_ptr)

        def __reversed__(self) -> Iterator[Function]:
            func_ptr = LLVMGetLastFunction(self._mod)

            while func_ptr:
                yield Function(func_ptr)

                func_ptr = LLVMGetPreviousFunction(self)

        def first(self) -> Optional[Function]:
            func_ptr = LLVMGetFirstFunction(self._mod)
            
            if func_ptr:
                return Function(func_ptr)
            
            return None

        def last(self) -> Optional[Function]:
            func_ptr = LLVMGetLastFunction(self._mod)

            if func_ptr:
                return Function(func_ptr)    

            return None

    @property
    def functions(self) -> _Functions:
        return Module._Functions(self) 

# This import has to go down here to handle an import cycle.
from .target import TargetData

# ---------------------------------------------------------------------------- #

@llvm_api
def LLVMModuleCreateWithNameInContext(name: c_char_p, ctx: Context) -> c_object_p:
    pass

@llvm_api
def LLVMDisposeModule(m: Module):
    pass

@llvm_api
def LLVMGetModuleIdentifier(m: Module, length: POINTER(c_size_t)) -> c_char_p:
    pass

@llvm_api
def LLVMSetModuleIdentifier(m: Module, mod_id: c_char_p, length: c_size_t):
    pass

@llvm_api
def LLVMGetModuleDataLayout(m: Module) -> TargetData:
    pass

@llvm_api
def LLVMSetModuleDataLayout(m: Module, data_layout: TargetData):
    pass

@llvm_api
def LLVMGetTarget(m: Module) -> c_char_p:
    pass

@llvm_api
def LLVMSetTarget(m: Module, target: c_char_p):
    pass

@llvm_api
def LLVMGetModuleContext(m: Module) -> c_object_p:
    pass

@llvm_api
def LLVMDumpModule(m: Module):
    pass

@llvm_api
def LLVMAddFunction(m: Module, name: c_char_p, func_type: FunctionType) -> c_object_p:
    pass

@llvm_api
def LLVMGetNamedFunction(m: Module, name: c_char_p) -> c_object_p:
    pass

@llvm_api
def LLVMDeleteFunction(func: Function):
    pass

@llvm_api
def LLVMGetFirstFunction(m: Module) -> c_object_p:
    pass

@llvm_api
def LLVMGetNextFunction(func_iter: c_object_p) -> c_object_p:
    pass

@llvm_api
def LLVMGetLastFunction(m: Module) -> c_object_p:
    pass

@llvm_api
def LLVMGetPreviousFunction(func_iter: c_object_p) -> c_object_p:
    pass

@llvm_api
def LLVMVerifyModule(m: Module, verifier_action: c_enum, out_message: POINTER(c_char_p)) -> c_enum:
    pass