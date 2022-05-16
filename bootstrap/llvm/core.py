from typing import Optional
import atexit
from ctypes import c_char_p, c_size_t

from . import LLVMObject, LLVMStringMessage, llvm_api, c_object_p

class Context(LLVMObject):
    def __init__(self, ptr: Optional[c_object_p] = None):
        super().__init__(ptr if ptr else LLVMContextCreate())

    def dispose(self):
        LLVMContextDispose(self)

    @staticmethod
    def global_ctx() -> 'Context':
        return Context(LLVMGetGlobalContext())

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
        return LLVMGetModuleIdentifier(self).value

    @name.setter
    def name(self, new_name: str):
        LLVMSetModuleIdentifier(self, c_char_p(new_name), c_size_t(len(new_name)))

    @property
    def data_layout(self):
        return LLVMGetDataLayoutStr(self).value

    @data_layout.setter
    def data_layout(self, data_layout: str):
        LLVMSetDataLayout(self, c_char_p(data_layout))

    @property
    def target(self):
        return LLVMGetTarget(self).value

    @target.setter
    def target(self, target: str):
        LLVMSetTarget(self, c_char_p(target))

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
