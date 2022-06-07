from . import *
from .module import Module

class DIBuilder(LLVMObject):
    def __init__(self, mod: Module):
        super().__init__(LLVMCreateDIBuilder(mod))
        get_context().take_ownership(self)

    def dispose(self):
        LLVMDisposeDIBuilder(self)

    def finalize(self):
        LLVMDIBuilderFinalize(self)

    # ---------------------------------------------------------------------------- #

# ---------------------------------------------------------------------------- #

@llvm_api
def LLVMCreateDIBuilder(m: Module) -> c_object_p:
    pass

@llvm_api
def LLVMDisposeDIBuilder(builder: DIBuilder):
    pass

@llvm_api
def LLVMDIBuilderFinalize(builder: DIBuilder):
    pass