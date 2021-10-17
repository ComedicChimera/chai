from dataclasses import dataclass
from typing import List
from enum import IntEnum, auto

# DataType is the base class for all data types in Chai
class DataType:
    # equals returns if two types are exactly identical
    def equals(self, other):
        return self == other

    # equiv returns if two types are semantically equivalent: eg. a type and an
    # alias to that type are equivalent but not equal.
    def equiv(self, other):
        return self.equals(other)

# ---------------------------------------------------------------------------- #

# PrimType represents a primtive type
class PrimType(DataType, IntEnum):
    U8 = auto()
    U16 = auto()
    U32 = auto()
    U64 = auto()
    I8 = auto()
    I16 = auto()
    I32 = auto()
    I64 = auto()
    F32 = auto()
    F64 = auto()
    STRING = auto()
    BOOL = auto()
    RUNE = auto()
    NOTHING = auto()

    def __repr__(self) -> str:
        return PRIM_STRINGS[self.value]

    def __str__(self) -> str:
        return self.__repr__()

# PRIM_STRINGS is a listing of the primitive values as strings
PRIM_STRINGS = {
    PrimType.U8: 'u8',
    PrimType.U16: 'u16',
    PrimType.U32: 'u32',
    PrimType.U64: 'u64',
    PrimType.I8: 'i8',
    PrimType.I16: 'i16',
    PrimType.I32: 'i32',
    PrimType.I64: 'i64',
    PrimType.F32: 'f32',
    PrimType.F64: 'f64',
    PrimType.STRING: 'string',
    PrimType.BOOL: 'bool',
    PrimType.RUNE: 'rune',
    PrimType.NOTHING: 'nothing',
}

# ---------------------------------------------------------------------------- #

# FuncArg represents a function argument
@dataclass(repr=True)
class FuncArg:
    name: str
    typ: DataType
    by_ref: bool

    def __repr__(self) -> str:
        return f'&{self.typ}' if self.by_ref else repr(self.typ)

# FuncType represents a function type
@dataclass(repr=False)
class FuncType(DataType):
    args: List[FuncArg]
    rt_type: DataType

    def __repr__(self) -> str:
        return f'({", ".join(map(repr, self.args))}) -> {self.rt_type}'
