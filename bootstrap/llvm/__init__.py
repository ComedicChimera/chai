'''Provides some utilities common to the LLVM API bindings.'''

import os
from ctypes import cdll, POINTER, c_void_p, c_int
from functools import wraps
from enum import Enum
import threading
from typing import Callable, get_type_hints, List, Optional
import atexit

__all__ = [
    'c_object_p',
    'c_enum',
    'LLVMObject',
    'LLVMEnum',
    'llvm_api',
    'Context',
    'get_context',
]

# This is the type to be used whenever an LLVM API function accepts or returns a
# pointer to an object of some form.
c_object_p = POINTER(c_void_p)

# This type is used to represent an LLVM enum.
c_enum = c_int

class LLVMObject:
    '''
    Represents an object in the LLVM API wrapper.  This is essentially
    just a base class and is not meant to be used on its own.

    Attributes
    ----------
    ptr: c_object_p
        The internal object pointer (provided by libLLVM).

    Methods
    -------
    take_ownership(obj: LLVMObject)
    '''

    ptr: c_object_p
    _as_parameter_: c_object_p

    def __init__(self, ptr: c_object_p):
        '''
        Params
        ------
        ptr: c_object_p
            The internal object pointer.
        '''

        self.ptr = ptr
        self._as_parameter_ = ptr

    def dispose(self):
        '''
        Disposes of this object's associated resources.  This method is intended
        to be overridden by deriving classes.  This method does not need to
        dispose of the resources of objects owned by this object.
        '''

    @classmethod
    def from_param(cls: 'LLVMObject', self: object) -> c_object_p:
        '''ctypes function to use this class as an argument type.'''

        if not isinstance(self, cls):
            raise TypeError()

        return self._as_parameter_

class LLVMEnum(Enum):
    '''Represents an LLVM enumeration.'''

    def _generate_next_value_(name: str, start: int, count: int, last_values: List[int]) -> int:
        '''
        Overriddes the default Enum class to start enum values at 0 if no first
        value is given.  Otherwise, it starts at whatever the first value is.
        '''

        if len(last_values) > 0:
            return last_values[-1] + 1
            
        return 0

    @classmethod
    def from_param(cls: 'LLVMEnum', self: object) -> int:
        '''ctypes function to use this class as an argument type.'''

        if not isinstance(self, cls):
            raise TypeError()

        return self.value

# ---------------------------------------------------------------------------- #     

def llvm_api(f: Callable) -> Callable:
    '''
    A decorator for converting a Python function into an LLVM API entry using
    its type annotations to fill in in `ctypes` properties.  The Python function
    can then be called as an API function.
    '''

    # Get the function by the same name as the input function from the library.
    lib_func = getattr(lib, f.__name__)

    # Update its argument and return types using the type hints of the passed in
    # function.
    type_hints = get_type_hints(f)
    lib_func.argtypes = [v for k, v in type_hints.items() if k != 'return']
    
    if 'return' in type_hints:
        lib_func.restype = type_hints['return'] 

    # Create and return the wrapper: using `wraps` to preserve the name and any
    # associated docstrings.
    @wraps(f)
    def wrapper(*args, **kwargs):
        return lib_func(*args)

    return wrapper

# ---------------------------------------------------------------------------- #

# Load the LLVM C library using the environment variable to locate it.
llvm_c_path = os.environ.get('LLVM_C_PATH')

if not llvm_c_path:
    print('missing environment variable `LLVM_C_PATH`')
    exit(-1)

lib = cdll.LoadLibrary(llvm_c_path)

# ---------------------------------------------------------------------------- #

class Context(LLVMObject):
    '''
    Represents an LLVM Context object.  The wrapper uses this object to help
    manage resources and ensure partial thread-safety when using LLVM.
    '''

    # The shared Python object associated with LLVM's global context.
    _global_ctx: Optional['Context'] = None

    # This list of objects owned by this object: ie. the list of objects it is
    # responsible for deleting.
    _owned_objects: List['LLVMObject']

    def __init__(self, ptr: Optional[c_object_p] = None):
        '''
        Params
        ------
        ptr: Optional[c_object_p]
            The C object pointer to the context this object wraps.  A new
            context is created if it is not provided.
        '''

        super().__init__(ptr if ptr else LLVMContextCreate())
        
        self._owned_objects = []

    @staticmethod
    def global_ctx() -> 'Context':
        '''Returns LLVM's global context.'''

        if not Context._global_ctx:
            Context._global_ctx = Context(LLVMGetGlobalContext())

        return Context._global_ctx

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
        '''Disposes of this object's resources.'''

        LLVMContextDispose(self)

    def __enter__(self) -> 'Context':
        '''Makes this context the active context for the current thread.'''

        thread_id = curr_thread_id()

        assert thread_id not in ctx_table, "only one context per thread at a time"

        ctx_table[thread_id] = self

        return self

    def __exit__(self, exc_type, exc_val, exc_tb):
        '''
        Removes this context as the active context for the current thread and
        disposes of all the resources associated with it.
        '''

        del ctx_table[curr_thread_id()]

        for obj in self._owned_objects:
            obj.dispose()

        self.dispose()

# The table mapping thread IDs to their active contexts.
ctx_table = {}

def curr_thread_id() -> Optional[int]:
    '''Returns the current thread identifier.'''

    return threading.current_thread().ident

def get_context() -> Context:
    '''Returns the current context.'''

    if ctx := ctx_table.get(curr_thread_id()):
        return ctx

    return Context.global_ctx()

class PassRegistry(LLVMObject):
    '''Represents an LLVM Pass Registry object.'''

    def __init__(self):
        super().__init__(LLVMGetGlobalPassRegistry())  

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
def LLVMInitializeAnalysis(p: PassRegistry):
    pass

@llvm_api
def LLVMInitializeCodeGen(p: PassRegistry):
    pass

@llvm_api
def LLVMInitializeTarget(p: PassRegistry):
    pass

@llvm_api
def LLVMShutdown():
    pass

# ---------------------------------------------------------------------------- #

# LLVM Initialization

# Force LLVM to initialize the global context.
Context.global_ctx()

# Get the global pass registry.
p = PassRegistry()

# Initialize all of the LLVM libraries we use.
LLVMInitializeCore(p)
LLVMInitializeAnalysis(p)
LLVMInitializeCodeGen(p)
LLVMInitializeTarget(p)

# LLVM Shutdown Logic
def llvm_shutdown():
    '''Cleans up and disposes of LLVM resources.'''

    LLVMShutdown()

atexit.register(llvm_shutdown)
