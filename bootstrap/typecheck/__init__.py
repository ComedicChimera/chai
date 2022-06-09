'''Represents Chai's type system.'''

from abc import ABC, abstractmethod
from dataclasses import dataclass
from enum import Enum
from typing import List

import util

__all__ = [
    'Type',
    'PrimitiveType',
    'PointerType',
    'FuncType',
    'typedataclass'
]

class Type(ABC):
    '''An abstract base class for all types.'''

    _compare_exact: bool = False

    @abstractmethod
    def _compare(self, other: 'Type') -> bool:
        '''
        Returns whether this type is exactly equal or equivalent to the other
        type depending on the value of `_compare_exact`.

        .. warning: This method should only be called by `Type`.

        Params
        ------
        other: Type
            The type to compare this type to.
        '''

    @abstractmethod
    def _cast_from(self, other: 'Type') -> bool:
        '''
        Returns whether you can cast the other type to this type.  This method
        need only return whether a cast is strictly possible: casting based on
        equivalency is tested by the casting operators.

        .. warning: This method should only be called by `Type`.

        Params
        ------
        other: Type
            The type to cast from.
        '''

    @property
    @abstractmethod
    def size(self) -> int:
        '''Returns the size of the type in bytes.'''

    # ---------------------------------------------------------------------------- #

    def inner_type(self) -> 'Type':
        '''
        Returns the "inner" type of self.  For most types, this is just an
        identity function; however, for types such as type variables and aliases
        which essentially just wrap other types, this method is useful.
        '''
        return self

    def exact_equals(self, other: 'Type') -> bool:
        '''
        Returns whether this type is exactly equal to the other type.

        Params
        ------
        other: Type
            The type to compare this type to.
        '''

        self._compare_exact = True
        result = self.inner_type()._compare(other.inner_type())
        self._compare_exact = False
        return result

    # ---------------------------------------------------------------------------- #

    @property
    def bit_size(self) -> int:
        '''Returns the size of the type in bits.'''

        return self.size * 8

    @property
    def align(self) -> int:
        '''Returns the alignment of the type in bytes.'''

        return self.size

    @property
    def bit_align(self) -> int:
        '''Returns the alignment of the type in bits.'''

        return self.bit_size

    # ---------------------------------------------------------------------------- #

    def __eq__(self, other: object) -> bool:
        '''
        Returns whether this type is equivalent to other.
        
        Params
        ------
        other: object
            The object to compare this type to.
        '''

        if isinstance(other, Type):
            return self.inner_type()._compare(other.inner_type())

        return False

    def __lt__(self, other: object) -> bool:
        '''
        Returns whether you can cast this type to the other type.

        Params
        ------
        other: Type
            The type to cast to.  
        '''

        if isinstance(other, Type):
            self_inner = self.inner_type()
            other_inner = other.inner_type()

            if self_inner._compare(other_inner):
                return True

            return other_inner._cast_from(self_inner)

        raise TypeError('can only cast between types')
    
    def __gt__(self, other: object) -> bool:
        '''
        Returns whether you can cast the other type to this type.

        Params
        ------
        other: Type
            The type to cast from.
        '''

        if isinstance(other, Type):
            self_inner = self.inner_type()
            other_inner = other.inner_type()

            if self_inner._compare(other_inner):
                return True

            return self_inner._cast_from(other_inner)

        raise TypeError('can only cast between types')

    def __str__(self) -> str:
        '''Returns the string representation of a type.'''

        return repr(self)

typedataclass = dataclass(eq=False)

class PrimitiveType(Type, Enum, metaclass=util.merge_metaclasses(Type, Enum)):
    '''
    Represents a primitive type.

    .. note: The values of all the integral types are their bit sizes.
    '''

    BOOL = 1
    I8 = 7
    U8 = 8
    I16 = 15
    U16 = 16
    I32 = 31
    U32 = 32
    I64 = 63
    U64 = 64

    F32 = 2
    F64 = 3
    NOTHING = 4

    def _compare(self, other: Type) -> bool:
        return super.__eq__(self, other)

    def _cast_from(self, other: Type) -> bool:
        # TODO
        return False

    def __repr__(self) -> str:
        return self.name.lower()

    @property
    def is_integral(self) -> bool:
        '''Returns whether this primitive is an integral type.'''

        return self.value > 4

    @property
    def is_floating(self) -> bool:
        '''Returns whether this primitive is a floating-point type.'''

        return self.value == 2 or self.value == 3

    @property
    def is_signed(self) -> bool:
        '''Returns whether this primitive type is signed.'''

        return self.value % 2 == 1

    @property
    def usable_width(self) -> int:
        '''
        If this type is integral, returns the usable width of the type in bits.
        '''

        return self.value

    @property
    def size(self) -> int:
        match self:
            case PrimitiveType.NOTHING | PrimitiveType.BOOL | PrimitiveType.I8 | PrimitiveType.U8:
                return 1
            case PrimitiveType.I16 | PrimitiveType.U16:
                return 2
            case PrimitiveType.I32 | PrimitiveType.U32 | PrimitiveType.F32:
                return 4
            case PrimitiveType.I64 | PrimitiveType.U64 | PrimitiveType.F64:
                return 8

@typedataclass
class PointerType(Type):
    '''
    Represents a pointer type.

    Attributes
    ----------
    elem_type: Type
        The element type of the pointer. 
    const: bool
        Whether the pointer points a immutable value. 
    '''

    elem_type: Type
    const: bool

    def _compare(self, other: Type) -> bool:
        if isinstance(other, PointerType):
            return self.elem_type == other.elem_type

        return False

    def _cast_from(self, other: Type) -> bool:
        # NOTE This is temporary -- allows for easy system bindings until the
        # unsafe package is properly implemented.
        if isinstance(other, PointerType):   
            return True

        return False

    def __repr__(self) -> str:
        if self.const:
            return f'*const {self.elem_type}'
        else:
            return f'*{self.elem_type}'

    @property
    def size(self) -> int:
        return util.POINTER_SIZE

@typedataclass
class FuncType(Type):
    '''
    Represents a function type.

    Attributes
    ----------
    param_types: List[Type]
        The parameter types of the function.
    rt_type: Type
        The return type of the function.
    '''

    param_types: List[Type]
    rt_type: Type

    def _compare(self, other: Type) -> bool:
        if isinstance(other, FuncType):
            return all(a == b for a, b in zip(self.param_types, other.param_types)) \
                and self.rt_type == other.rt_type

    def _cast_from(self, other: Type) -> bool:
        # No legal function casts.
        return False

    def __repr__(self) -> str:
        if len(self.param_types) == 1:
            param_str = repr(self.param_types[0])
        else:
            param_str = '(' + ', '.join(str(x) for x in self.param_types) + ')'

        return f'{param_str} -> {self.rt_type}'

    def size(self) -> int:
        return util.POINTER_SIZE
