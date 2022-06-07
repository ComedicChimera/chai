from ctypes import c_size_t, byref, c_char_p, POINTER, c_uint
from ctypes import Array as c_array
from typing import Optional, Iterator, Tuple

from . import *
from .types import FunctionType
from .ir import Function
from .value import Value

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
    def data_layout(self) -> 'TargetData':
        return TargetData(LLVMGetModuleDataLayout(self))

    @data_layout.setter
    def data_layout(self, data_layout: 'TargetData'):
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
        
        if LLVMVerifyModule(self, 2, byref(p)) == 0:
            return "", True

        return str(p.value, encoding='utf-8'), False

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

    class _NamedMetadataNodes:
        mod: 'Module'

        def __init__(self, mod: 'Module'):
            self.mod = mod
        
        def __iter__(self) -> Iterator['NamedMetadataNode']:
            named_md_ptr = LLVMGetFirstNamedMetadata(self.mod)

            while named_md_ptr:
                yield NamedMetadataNode(named_md_ptr)

                named_md_ptr = LLVMGetNextNamedMetadata(self.mod)

        def __reversed__(self) -> Iterator['NamedMetadataNode']:
            named_md_ptr = LLVMGetLastNamedMetadata(self.mod)

            while named_md_ptr:
                yield NamedMetadataNode(named_md_ptr)

                named_md_ptr = LLVMGetPreviousNamedMetadata(self.mod)

        def __getitem__(self, key: str) -> 'NamedMetadataNode':
            key_bytes = key.encode()
            named_md_ptr = LLVMGetNamedMetadata(self.mod, key_bytes, len(key_bytes))

            if named_md_ptr:
                return named_md_ptr

            raise KeyError(key)

        def add(self, key: str) -> 'NamedMetadataNode':
            key_bytes = key.encode()
            named_md_ptr = LLVMGetOrInsertNamedMetadata(self.mod, key_bytes, len(key_bytes))
            return NamedMetadataNode(named_md_ptr)
            
    @property
    def named_md_nodes(self) -> _NamedMetadataNodes:
        return Module._NamedMetadataNodes(self)

# This import has to go down here to handle an import cycle.
from .target import TargetData

class NamedMetadataNode(LLVMObject):
    parent: Module

    def __init__(self, ptr: c_object_p, parent: Module):
        super().__init__(ptr)
        self.parent = parent

    @property
    def name(self) -> str:
        return str(LLVMGetNamedMetadataName(self, byref(c_size_t())), encoding='utf-8')

    class _NamedMDOperands:
        named_md: 'NamedMetadataNode'

        def __init__(self, named_md: 'NamedMetadataNode'):
            self.named_md = named_md

        def __len__(self) -> int:
            return LLVMGetNamedMetadataNumOperands(self.named_md.parent, self.named_md.name.encode('utf-8'))

        def __iter__(self) -> Iterator[Value]:
            for item in self._get_operands():
                yield Value(item)

        def __getitem__(self, ndx: int) -> Value:
            if 0 <= ndx < len(self):
                return self._get_operands()[ndx]

            raise IndexError(ndx)

        def add(self, item: Value):
            LLVMAddNamedMetadataOperand(self.named_md.parent, self.named_md.name, item)

        def _get_operands(self) -> c_array:
            op_arr = (c_object_p * len(self))()
            LLVMGetNamedMetadataOperands(self.named_md.parent, self.named_md.name, op_arr)
            return op_arr

    @property
    def operands(self) -> _NamedMDOperands:
        return self._NamedMDOperands(self)

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

@llvm_api
def LLVMGetNamedMetadataName(named_md: NamedMetadataNode, name_len: POINTER(c_size_t)) -> c_char_p:
    pass

@llvm_api
def LLVMGetNamedMetadataNumOperands(m: Module, name: c_char_p) -> c_uint:
    pass

@llvm_api
def LLVMGetNamedMetadataOperands(m: Module, name: c_char_p, dest: POINTER(c_object_p)):
    pass

@llvm_api
def LLVMAddNamedMetadataOperand(m: Module, name: c_char_p, val: Value):
    pass

@llvm_api
def LLVMGetFirstNamedMetadata(m: Module) -> c_object_p:
    pass

@llvm_api
def LLVMGetLastNamedMetadata(m: Module) -> c_object_p:
    pass

@llvm_api
def LLVMGetNextNamedMetadata(named_m_d_node: NamedMetadataNode) -> c_object_p:
    pass

@llvm_api
def LLVMGetPreviousNamedMetadata(named_m_d_node: NamedMetadataNode) -> c_object_p:
    pass

@llvm_api
def LLVMGetNamedMetadata(m: Module, name: c_char_p, name_len: c_size_t) -> c_object_p:
    pass

@llvm_api
def LLVMGetOrInsertNamedMetadata(m: Module, name: c_char_p, name_len: c_size_t) -> c_object_p:
    pass