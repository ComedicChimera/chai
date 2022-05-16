from typing import Optional, List
import atexit
from ctypes import c_char_p, c_size_t

from . import LLVMObject, llvm_api, c_object_p

class Context(LLVMObject):
    # This list of objects owned by this object: ie. the list of objects it is
    # responsible for deleting.
    _owned_objects: List['LLVMObject']

    def __init__(self, ptr: Optional[c_object_p] = None):
        super().__init__(ptr if ptr else LLVMContextCreate())
        
        self._owned_objects = []

    @staticmethod
    def global_ctx() -> 'Context':
        return Context(LLVMGetGlobalContext())

    def take_ownership(self, obj: 'LLVMObject'):
        '''
        Prompts this context to take ownership of an LLVM object, making this
        context responsible for the deletion of `obj`.

        Params
        ------
        obj: LLVMObject
            The object to take ownership of.       
        '''

        self._owned_objects.append(obj)

    def dispose(self):
        LLVMContextDispose(self)

    def __enter__(self) -> 'Context':
        return self

    def __exit__(self):
        for obj in self._owned_objects:
            obj.dispose()

        self.dispose()

class PassRegistry(LLVMObject):
    def __init__(self):
        super().__init__(LLVMGetGlobalPassRegistry())  

# ---------------------------------------------------------------------------- #

class Module(LLVMObject):
    def __init__(self, name: str, ctx: Context):
        super().__init__(LLVMModuleCreateWithNameInContext(c_char_p(name), ctx))  

        ctx.take_ownership(self)

    def dispose(self):
        LLVMDisposeModule(self)

    @property
    def name(self) -> str:
        return str(LLVMGetModuleIdentifier(self), encoding='utf-8')

    @name.setter
    def name(self, new_name: str):
        LLVMSetModuleIdentifier(self, new_name, len(new_name))

    @property
    def data_layout(self) -> str:
        return str(LLVMGetDataLayoutStr(self), encoding='utf-8')

    @data_layout.setter
    def data_layout(self, data_layout: str):
        LLVMSetDataLayout(self, data_layout)

    @property
    def target(self) -> str:
        return str(LLVMGetTarget(self), encoding='utf-8')

    @target.setter
    def target(self, target: str):
        LLVMSetTarget(self, target)

    @property
    def context(self) -> Context:
        return Context(LLVMGetModuleContext(self))

    def dump(self):
        LLVMDumpModule(self)

# ---------------------------------------------------------------------------- #

@llvm_api
def LLVMContextCreate() -> c_object_p:
    pass

@llvm_api
def LLVMContextDispose(ctx: Context):
    pass

@llvm_api
def LLVMGetGlobalContext() -> c_object_p:
    pass

@llvm_api
def LLVMGetGlobalPassRegistry() -> c_object_p:
    pass

@llvm_api
def LLVMModuleCreateWithNameInContext(name: c_char_p, ctx: Context) -> c_object_p:
    pass

@llvm_api
def LLVMDisposeModule(m: Module):
    pass

@llvm_api
def LLVMGetModuleIdentifier(m: Module) -> c_char_p:
    pass

@llvm_api
def LLVMSetModuleIdentifier(m: Module, mod_id: c_char_p, length: c_size_t):
    pass

@llvm_api
def LLVMGetDataLayoutStr(m: Module) -> c_char_p:
    pass

@llvm_api
def LLVMSetDataLayout(m: Module, data_layout: c_char_p):
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
def LLVMInitializeCore(p: PassRegistry):
    pass

@llvm_api
def LLVMShutdown():
    pass

# ---------------------------------------------------------------------------- #

# LLVM Initialization
def llvm_init():
    '''Initializes LLVM.'''

    # Force LLVM to initialize the global context.
    Context.global_ctx()

    # Get the global pass registry.
    p = PassRegistry()

    # Initialize all of the LLVM libraries we use.
    LLVMInitializeCore(p)

llvm_init()

# LLVM Shutdown Logic
def llvm_shutdown():
    '''Cleans up and disposes of LLVM resources.'''

    LLVMShutdown()

atexit.register(llvm_shutdown)
