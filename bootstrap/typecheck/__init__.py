'''Represents Chai's type system.'''

from abc import ABC, abstractmethod
from dataclasses import dataclass
from enum import Enum, auto

class Type(ABC):
    '''An abstract base class for all types.'''

    @abstractmethod
    def equals(self, other: 'Type') -> bool:
        '''
        Returns whether this type is equal to the other type.

        .. warning: This method should only be called by `Type`.

        Params
        ------
        other: Type
            The type to compare this type to.
        '''

    @abstractmethod
    def cast_from(self, other: 'Type') -> bool:
        '''
        Returns whether you can cast the other type to this type.

        .. warning: This method should only be called by `Type`.

        Params
        ------
        other: Type
            The type to cast from.
        '''

    def inner_type(self) -> 'Type':
        '''
        Returns the "inner" type of self.  For most types, this is just an
        identity function; however, for types such as type variables and aliases
        which essentially just wrap other types, this method is useful.
        '''
        return self

    def __eq__(self, other: object) -> bool:
        '''
        Returns whether this type is equal to other.
        
        Params
        ------
        other: object
            The object to compare this type to.
        '''

        if isinstance(other, Type):
            return self.inner_type().equals(other.inner_type())

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

            if self_inner.equals(other_inner):
                return True

            return other_inner.cast_from(self_inner)

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

            if self_inner.equals(other_inner):
                return True

            return self_inner.cast_from(other_inner)

        raise TypeError('can only cast between types')

typedataclass = dataclass(eq=False)

class PrimitiveType(Enum, Type):
    '''Represents a primitive type.'''

    BOOL = auto()
    I8 = auto()
    U8 = auto()
    I16 = auto()
    U16 = auto()
    I32 = auto()
    U32 = auto()
    I64 = auto()
    U64 = auto()
    F32 = auto()
    F64 = auto()
    NOTHING = auto()

    def equals(self, other: Type) -> bool:
        return super.__eq__(self, other)

    def cast_from(self, other: Type) -> bool:
        # TODO
        return False

@typedataclass
class PointerType(Type):
    '''
    Represents a pointer type.

    Attributes
    ----------
    elem_type: Type
        The element type of the pointer.
    indirection: int
        The level of indirection of the pointer.  
    '''

    elem_type: Type
    indirection: int

    def equals(self, other: Type) -> bool:
        if isinstance(other, PointerType):
            return self.elem_type == other.elem_type and \
                self.indirection == other.indirection

        return False

    def cast_from(self, other: Type) -> bool:
        # NOTE This is temporary -- allows for easy system bindings until the
        # unsafe package is properly implemented.
        if isinstance(other, PointerType):   
            return True

        return False
