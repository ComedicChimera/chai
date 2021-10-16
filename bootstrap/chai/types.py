from dataclasses import dataclass
from typing import List
from enum import Enum, auto

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
class PrimType(Enum):
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
        # TODO
        return ''

# ---------------------------------------------------------------------------- #

# FuncArg represents a function argument
@dataclass(repr=True)
class FuncArg:
    by_ref: bool
    name: str
    typ: DataType

    def __repr__(self) -> str:
        return f'&{self.typ}' if self.by_ref else repr(self.typ)

# FuncType represents a function type
@dataclass(repr=False)
class FuncType(DataType):
    args: List[FuncArg]
    rt_type: DataType

    def __repr__(self) -> str:
        return f'({", ".join(map(repr, self.args))}) -> {self.rt_type}'
