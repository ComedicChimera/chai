from enum import auto
from ctypes import POINTER, c_char_p, c_uint, c_ulonglong, byref
from typing import Iterator, Optional

from . import *
from .types import Type

class ByteOrdering(LLVMEnum):
    BIG_ENDIAN = auto()
    LITTLE_ENDIAN = auto()

class CodeGenOptLevel(LLVMEnum):
    NONE = auto()
    LESS = auto()
    DEFAULT = auto()
    AGGRESSIVE = auto()

class RelocMode(LLVMEnum):
    DEFAULT = auto()
    STATIC = auto()
    PIC = auto()
    DYNAMIC_NO_PIC = auto()
    ROPI = auto()
    RWPI = auto()
    ROPI_RWPI = auto()

class CodeModel(LLVMEnum):
    DEFAULT = auto()
    JIT_DEFAULT = auto()
    TINY = auto()
    SMALL = auto()
    KERNEL = auto()
    MEDIUM = auto()
    LARGE = auto()

class CodeGenFileType(LLVMEnum):
    ASSEMBLY_FILE = auto()
    OBJECT_FILE = auto()

# ---------------------------------------------------------------------------- #

class TargetData(LLVMObject):
    def __init__(self, ptr: c_object_p):
        super().__init__(ptr)

        get_context().take_ownership(self)

    def dispose(self):
        LLVMDisposeTargetData(self)

    @property
    def byte_order(self) -> ByteOrdering:
        return ByteOrdering(LLVMByteOrder(self))

    @property
    def pointer_size(self) -> int:
        return LLVMPointerSize(self)

    @property
    def int_ptr_type(self) -> Type:
        return Type(LLVMIntPtrTypeInContext(get_context(), self))

    def store_size_of(self, typ: Type) -> int:
        return LLVMStoreSizeOfType(self, typ)

    def abi_size_of(self, typ: Type) -> int:
        return LLVMABISizeOfType(self, typ)

    def abi_align_of(self, typ: Type) -> int:
        return LLVMABIAlignmentOfType(self, typ)

    def call_frame_align_of(self, typ: Type) -> int:
        return LLVMCallFrameAlignmentOfType(self, typ)

    def preferred_align_of(self, typ: Type) -> int:
        return LLVMPreferredAlignmentOfType(typ)

# This import has to go down here to handle an import cycle.
from .module import Module

class Target(LLVMObject):
    def __init__(self, **kwargs):
        if ptr := kwargs.get('ptr'):
            assert isinstance(ptr, c_object_p)
            super().__init__(self, ptr)
        elif triple := kwargs.get('triple'):
            assert isinstance(triple, str)

            target_ptr = c_object_p()
            out_msg = c_char_p()
            if LLVMGetTargetFromTriple(triple.encode(), byref(target_ptr), byref(out_msg)):
                super().__init__(target_ptr)
            else:
                raise ValueError(str(out_msg, encoding='utf-8'))
        elif name := kwargs.get('name'):
            assert isinstance(name, str)
            super().__init__(self, LLVMGetTargetFromName(name.encode()))
        else:
            raise ValueError('expected kwarg of either `ptr`, `triple`, or `name`')   

    # @staticmethod
    # def initialize_all():
    #     LLVMInitializeAllTargetInfos()
    #     LLVMInitializeAllTargets()
    #     LLVMInitializeAllAsmPrinters()

    # @staticmethod
    # def initialize_native():
    #     LLVMInitializeNativeTarget()
    #     LLVMInitializeNativeAsmPrinter()

    class _Targets:
        _first: c_object_p

        def __init__(self):
            self._first = LLVMGetFirstTarget()

        def __iter__(self) -> Iterator['Target']:
            curr = self._first

            while curr:
                yield Target(ptr=curr)

                curr = LLVMGetNextTarget(curr)

        def first(self) -> Optional['Target']:
            if self._first:
               return Target(ptr=self._first)

            return None 

    @staticmethod
    def targets() -> _Targets:
        return Target._Targets()

    @property
    def has_jit(self) -> bool:
        return LLVMTargetHasJIT(self) == 1

    @property
    def has_machine(self) -> bool:
        return LLVMTargetHasTargetMachine(self) == 1

    @property
    def has_asm_backend(self) -> bool:
        return LLVMTargetHasAsmBackend(self) == 1

    def create_machine(self, **kwargs) -> 'TargetMachine':
        assert self.has_machine

        triple = kwargs.get('triple', str(LLVMGetDefaultTargetTriple(), encoding='utf-8'))
        cpu_name = kwargs.get('cpu_name', str(LLVMGetHostCPUName(), encoding='utf-8'))
        cpu_features = kwargs.get('cpu_features', str(LLVMGetHostCPUFeatures(), encoding='utf-8'))

        opt_level = kwargs.get('opt_level', CodeGenOptLevel.DEFAULT)
        reloc_mode = kwargs.get('reloc_mode', RelocMode.DEFAULT)
        code_model = kwargs.get('code_model', CodeModel.DEFAULT)

        return TargetMachine(LLVMCreateTargetMachine(
            self,
            triple.encode(),
            cpu_name.encode(),
            cpu_features.encode(),
            opt_level,
            reloc_mode,
            code_model
        ))

class TargetMachine(LLVMObject):
    def __init__(self, ptr: c_object_p):
        super().__init__(ptr)

        get_context().take_ownership(self)

    def dispose(self):
        LLVMDisposeTargetMachine(self)

    @property
    def target(self) -> Target:
        return Target(ptr=LLVMGetTargetMachineTarget(self))

    @property
    def triple(self) -> str:
        return str(LLVMGetTargetMachineTriple(self), encoding='utf-8')

    @property
    def cpu_name(self) -> str:
        return str(LLVMGetTargetMachineCPU(self), encoding='utf-8')

    @property
    def feature_string(self) -> str:
        return str(LLVMGetTargetMachineFeatureString(self), encoding='utf-8')

    @property
    def data_layout(self) -> TargetData:
        return TargetData(LLVMCreateTargetDataLayout(self))

    def compile_module(
        self, 
        mod: Module, 
        out_file_path: str, 
        file_type: CodeGenFileType=CodeGenFileType.OBJECT_FILE,
    ):
        err_msg = c_char_p()

        if LLVMTargetMachineEmitToFile(self, mod, out_file_path.encode(), file_type, byref(err_msg)) == 0:
            raise Exception(str(err_msg, encoding='utf-8'))

# ---------------------------------------------------------------------------- #

# @llvm_api
# def LLVMInitializeAllTargets():
#     pass

# @llvm_api
# def LLVMInitializeAllTargetInfos():
#     pass

# @llvm_api
# def LLVMInitializeAllAsmPrinters():
#     pass

# @llvm_api
# def LLVMInitializeNativeTarget():
#     pass

# @llvm_api
# def LLVMInitializeNativeAsmPrinter():
#     pass

@llvm_api
def LLVMDisposeTargetData(td: TargetData):
    pass

@llvm_api
def LLVMByteOrder(td: TargetData) -> c_enum:
    pass

@llvm_api
def LLVMPointerSize(td: TargetData) -> c_uint:
    pass

@llvm_api
def LLVMIntPtrTypeInContext(td: TargetData) -> c_object_p:
    pass

@llvm_api
def LLVMStoreSizeOfType(td: TargetData, typ: Type) -> c_ulonglong:
    pass

@llvm_api
def LLVMABISizeOfType(td: TargetData, typ: Type) -> c_ulonglong:
    pass

@llvm_api
def LLVMABIAlignmentOfType(td: TargetData, typ: Type) -> c_uint:
    pass

@llvm_api
def LLVMCallFrameAlignmentOfType(td: TargetData, typ: Type) -> c_uint:
    pass

@llvm_api
def LLVMPreferredAlignmentOfType(td: TargetData, typ: Type) -> c_uint:
    pass

@llvm_api
def LLVMGetFirstTarget() -> c_object_p:
    pass

@llvm_api
def LLVMGetNextTarget(target: c_object_p) -> c_object_p:
    pass

@llvm_api
def LLVMGetTargetFromName(name: c_char_p) -> c_object_p:
    pass

@llvm_api
def LLVMGetTargetFromTriple(triple: c_char_p, target: POINTER(c_object_p), err_msg: POINTER(c_char_p)) -> c_enum:
    pass

@llvm_api
def LLVMTargetHasJIT(target: Target) -> c_enum:
    pass

@llvm_api
def LLVMTargetHasTargetMachine(target: Target) -> c_enum:
    pass

@llvm_api
def LLVMTargetHasAsmBackend(target: Target) -> c_enum:
    pass

@llvm_api
def LLVMCreateTargetMachine(
    target: Target, 
    triple: c_char_p, 
    cpu: c_char_p, 
    features: c_char_p,
    opt_level: CodeGenOptLevel,
    reloc_mode: RelocMode,
    code_model: CodeModel
) -> c_object_p:
    pass

@llvm_api
def LLVMDisposeTargetMachine(tm: TargetMachine):
    pass

@llvm_api
def LLVMGetDefaultTargetTriple() -> c_char_p:
    pass

@llvm_api
def LLVMGetHostCPUName() -> c_char_p:
    pass

@llvm_api
def LLVMGetHostCPUFeatures() -> c_char_p:
    pass

@llvm_api
def LLVMGetTargetMachineTarget(tm: TargetMachine) -> c_object_p:
    pass

@llvm_api
def LLVMGetTargetMachineTriple(tm: TargetMachine) -> c_char_p:
    pass

@llvm_api
def LLVMGetTargetMachineCPU(tm: TargetMachine) -> c_char_p:
    pass

@llvm_api
def LLVMGetTargetMachineFeatureString(tm: TargetMachine) -> c_char_p:
    pass

@llvm_api
def LLVMCreateTargetDataLayout(tm: TargetMachine) -> c_object_p:
    pass

@llvm_api
def LLVMTargetMachineEmitToFile(
    tm: TargetMachine,
    mod: Module,
    file_name: c_char_p,
    file_type: CodeGenFileType,
    error_msg: POINTER(c_char_p)
) -> c_enum:
    pass