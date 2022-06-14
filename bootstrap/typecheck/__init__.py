'''Represents Chai's type system.'''

from abc import ABC, abstractmethod
from dataclasses import dataclass, field
from enum import Enum
from typing import List, Dict, Optional, Iterator

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

    @abstractmethod
    def _equals(self, other: 'Type') -> bool:
        '''
        Returns whether this type is exactly equal to the other type.

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
            return self.inner_type()._equals(other.inner_type())

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

            if self_inner._equals(other_inner):
                return True

            # Handle the special case of synthetic type casting.
            if isinstance(self_inner, SynthType):
                return other_inner._cast_from(self_inner.base_type.inner_type())

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

            if self_inner._equals(other_inner):
                return True

            # Handle the special case of synthetic type casting.
            if isinstance(other_inner, SynthType):
                return self_inner._cast_from(other_inner.base_type.inner_type())

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
    UNIT = 4

    def _equals(self, other: Type) -> bool:
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
            case PrimitiveType.UNIT | PrimitiveType.BOOL | PrimitiveType.I8 | PrimitiveType.U8:
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

    def _equals(self, other: Type) -> bool:
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

    def _equals(self, other: Type) -> bool:
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

# ---------------------------------------------------------------------------- #

@typedataclass
class DefinedType(Type):
    '''
    A parent class for defined types.

    Attributes
    ----------
    name: str
        The name of the defined type.
    parent_id: int
        The ID of the parent package to this type.
    parent_name: str
        The name of the parent package to this type.
    '''

    name: str
    parent_id: int
    parent_name: str

    def _equals(self, other: Type) -> bool:
        if isinstance(other, DefinedType):
            return self.name == other.name and self.parent_id == other.parent_id

        return False

    def __repr__(self) -> str:
        return self.parent_name + '.' + self.name

@dataclass
class RecordField:
    '''
    Represents a field in a record type.

    Attributes
    ----------
    name: str
        The name of the field.
    type: Type
        The type of the field.
    const: bool
        Whether the field is constant.
    requires_init: bool
        Whether the field must be initialized.
    '''

    name: str
    type: Type
    const: bool 
    requires_init: bool

@typedataclass
class RecordType(DefinedType):
    '''
    Represents a record type.

    Attributes
    ----------
    fields: Dict[str, RecordField]
        The ordered dictionary of fields of the record.
    extends: List[RecordType]
        The list of records this record extends.
    packed: bool
        Whether or not the struct is packed.
    '''

    fields: Dict[str, RecordField]
    extends: List['RecordType'] = field(default_factory=list)
    packed: bool = False

    def _cast_from(self, other: 'Type') -> bool:
        # Records can't be cast to any other type.
        return False

    @property
    def size(self) -> int:
        r_size = 0

        for field in self.all_fields:
            # Account for alignment padding when computing the size.
            if (off := r_size % field.type.align) != 0:
                r_size += field.type.align - off

            r_size += field.type.size

        # Account for any end padding for alignment.
        return r_size + self.align - r_size % self.align

    @property
    def align(self) -> int:
        return max(field.align for field in self.all_fields)

    def lookup_field(self, name: str) -> Optional[RecordField]:
        '''
        Finds and returns the field by the given name if it exists.

        Params
        ------
        name: str
            The field name to lookup.
        '''

        if name in self.fields:
            return self.fields[name]

        for extend in self.extends:
            if efield := extend.lookup_field(name):
                return efield

        return None

    @property
    def all_fields(self) -> Iterator[RecordField]:
        '''
        Returns an iterator over the fields of the record and the fields of the
        various records it extends.
        '''

        for field in self.fields:
            yield field

        for extend in self.extends:
            for efield in extend.fields:
                yield efield

class SynthType(DefinedType):
    '''
    Represents a synthetic type declared via `newtype`.

    Attributes
    ----------
    base_type: Type
        The type used as the base for this synthetic type.
    '''

    base_type: Type

    def _cast_from(self, other: 'Type') -> bool:
        return self.base_type.inner_type()._cast_from(other)

    @property
    def size(self) -> int:
        return self.base_type.size

    @property
    def align(self) -> int:
        return self.base_type.align