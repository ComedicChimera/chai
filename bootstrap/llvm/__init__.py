'''Provides some utilities common to the LLVM API bindings.'''

import os
from ctypes import cdll, POINTER, c_void_p, c_int
from functools import wraps
from enum import Enum
from typing import Callable, get_type_hints, List

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
        '''Overriddes the default Enum class to start enum values at 0.'''
        return count

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

null_object_ptr = POINTER(c_object_p)

# ---------------------------------------------------------------------------- #

# Load the LLVM C library using the environment variable to locate it.
llvm_c_path = os.environ.get('LLVM_C_PATH')

if not llvm_c_path:
    print('missing environment variable `LLVM_C_PATH`')
    exit(-1)

lib = cdll.LoadLibrary(llvm_c_path)
