from enum import Enum, auto
from typing import Optional, List
import atexit
from ctypes import POINTER, byref, c_char_p, c_uint64, c_uint

from . import LLVMEnum, LLVMObject, llvm_api, c_object_p, c_enum
from .types import Type

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

    def __exit__(self, exc_type, exc_val, exc_tb):
        for obj in self._owned_objects:
            obj.dispose()

        self.dispose()

class PassRegistry(LLVMObject):
    def __init__(self):
        super().__init__(LLVMGetGlobalPassRegistry())  

# ---------------------------------------------------------------------------- #

class Attribute(LLVMObject):
    class Kind(LLVMEnum):
        ALWAYS_INLINE = 1
        ARG_MEM_ONLY = 2
        BUILTIN = 3
        COLD = 4
        CONVERGENT = 5
        DISABLE_SANITIZER_INSTRUMENTATION = 6
        HOT = 7
        IMM_ARG = 8
        IN_REG = 9
        INACCESSIBLE_MEM_ONLY = 10
        INACCESSIBLE_MEM_OR_ARG_MEM_ONLY = 11
        INLINE_HINT = 12
        JUMP_TABLE = 13
        MIN_SIZE = 14
        MUST_PROGRESS = 15
        NAKED = 16
        NEST = 17
        NO_ALIAS = 18
        NO_BUILTIN = 19
        NO_CALLBACK = 20
        NO_CAPTURE = 21
        NO_CF_CHECK = 22
        NO_DUPLICATE = 23
        NO_FREE = 24
        NO_IMPLICIT_FLOAT = 25
        NO_INLINE = 26
        NO_MERGE = 27
        NO_PROFILE = 28
        NO_RECURSE = 29
        NO_RED_ZONE = 30
        NO_RETURN = 31
        NO_SANITIZE_COVERAGE = 32
        NO_SYNC = 33
        NO_UNDEF = 34
        NO_UNWIND = 35
        NON_LAZY_BIND = 36
        NON_NULL = 37
        NULL_POINTER_IS_VALID = 38
        OPT_FOR_FUZZING = 39
        OPTIMIZE_FOR_SIZE = 40
        OPTIMIZE_NONE = 41
        READ_NONE = 42
        READ_ONLY = 43
        RETURNED = 44
        RETURNS_TWICE = 45
        S_EXT = 46
        SAFE_STACK = 47
        SANITIZE_ADDRESS = 48
        SANITIZE_HW_ADDRESS = 49
        SANITIZE_MEM_TAG = 50
        SANITIZE_MEMORY = 51
        SANITIZE_THREAD = 52
        SHADOW_CALL_STACK = 53
        SPECULATABLE = 54
        SPECULATIVE_LOAD_HARDENING = 55
        STACK_PROTECT = 56
        STACK_PROTECT_REQ = 57
        STACK_PROTECT_STRONG = 58
        STRICT_FP = 59
        SWIFT_ASYNC = 60
        SWIFT_ERROR = 61
        SWIFT_SELF = 62
        UW_TABLE = 63
        WILL_RETURN = 64
        WRITE_ONLY = 65
        Z_EXT = 66
        LAST_ENUM_ATTR = 66
        FIRST_TYPE_ATTR = 67
        BY_REF = 67
        BY_VAL = 68
        ELEMENT_TYPE = 69
        IN_ALLOCA = 70
        PREALLOCATED = 71
        STRUCT_RET = 72
        LAST_TYPE_ATTR = 72
        FIRST_INT_ATTR = 73
        ALIGNMENT = 73
        ALLOC_SIZE = 74
        DEREFERENCEABLE = 75
        DEREFERENCEABLE_OR_NULL = 76
        STACK_ALIGNMENT = 77
        V_SCALE_RANGE = 78

    class Variant(Enum):
        ENUM = auto()
        STRING = auto()
        TYPE = auto()

    _variant: Variant

    def __init__(self, **kwargs):
        if ptr := kwargs.get('ptr'):
            assert isinstance(ptr, c_object_p)

            super().__init__(ptr)

            if LLVMIsEnumAttribute(self) == 1:
                self._variant = Attribute.Variant.ENUM
            elif LLVMIsStringAttribute(self) == 1:
                self._variant = Attribute.Variant.STRING
            else:
                self._variant = Attribute.Variant.TYPE
        else:
            assert (kind := kwargs.get('kind'))

            ctx = kwargs.get('context', Context.global_ctx())

            if isinstance(kind, str):
                assert (value := kwargs.get('kind'))

                value_bytes = value.encode()
                super().__init__(LLVMCreateStringAttribute(ctx, kind, value_bytes, len(value_bytes)))
                self._variant = Attribute.Variant.STRING
            else:
                assert isinstance(kind, Attribute.Kind)

                if value := kwargs.get('value'):
                    if isinstance(value, Type):
                        super().__init__(LLVMCreateTypeAttribute(ctx, kind, value))
                        self._variant = Attribute.Variant.TYPE
                        return

                    super().__init__(LLVMCreateEnumAttribute(ctx, kind, value))
                else:
                    super().__init__(LLVMCreateEnumAttribute(ctx, kind, 0))

                self._variant = Attribute.Variant.ENUM

    @property
    def kind(self) -> Kind | str:
        match self._variant:
            case Attribute.Variant.ENUM | Attribute.Variant.TYPE:
                return Attribute.Kind(LLVMGetEnumAttributeKind(self))
            case Attribute.Variant.STRING:
                length = c_uint()
                return str(LLVMGetStringAttributeKind(self, byref(length)), encoding='utf-8')

    @property
    def value(self) -> int | str | Type:
        match self._variant:
            case Attribute.Variant.ENUM:
                return LLVMGetEnumAttributeValue(self)
            case Attribute.Variant.TYPE:
                ptr = LLVMGetTypeAttributeValue(self)
                return Type(ptr)
            case Attribute.Variant.STRING:
                length = c_uint()
                return str(LLVMGetStringAttributeValue(self, byref(length)), encoding='utf-8')

    @property
    def variant(self) -> Variant:
        return self._variant 

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
def LLVMInitializeCore(p: PassRegistry):
    pass

@llvm_api
def LLVMShutdown():
    pass

@llvm_api
def LLVMIsEnumAttribute(attr: Attribute) -> c_enum:
    pass

@llvm_api
def LLVMCreateEnumAttribute(ctx: Context, kind: Attribute.Kind, val: c_uint64) -> c_object_p:
    pass

@llvm_api
def LLVMGetEnumAttributeKind(attr: Attribute) -> c_enum:
    pass

@llvm_api
def LLVMGetEnumAttributeValue(attr: Attribute) -> c_uint64:
    pass

@llvm_api
def LLVMIsStringAttribute(attr: Attribute) -> c_enum:
    pass

@llvm_api
def LLVMCreateStringAttribute(ctx: Context, kind: c_char_p, klen: c_uint, val: c_char_p, vlen: c_uint) -> c_object_p:
    pass

@llvm_api
def LLVMGetStringAttributeKind(attr: Attribute, length: POINTER(c_uint)) -> c_char_p:
    pass

@llvm_api
def LLVMGetStringAttributeValue(attr: Attribute, length: POINTER(c_uint)) -> c_char_p:
    pass

@llvm_api
def LLVMIsTypeAttribute(attr: Attribute) -> c_enum:
    pass

@llvm_api
def LLVMCreateTypeAttribute(ctx: Context, kind: Attribute.Kind, val: Type) -> c_object_p:
    pass

@llvm_api
def LLVMGetTypeAttributeValue(attr: Attribute) -> c_object_p:
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
