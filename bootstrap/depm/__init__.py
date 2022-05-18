'''Provides Chai's definitions for symbols and scopes.'''

from dataclasses import dataclass
from enum import Enum, auto
from typing import Optional

from report import TextSpan
from typecheck import Type
from llvm.value import Value

@dataclass
class Symbol:
    '''
    Represents a semantic symbol: a named value or definition.

    Attributes
    ----------
    name: str
        The symbol's name.
    parent_id: int
        The ID of the package the symbol is defined in.
    type: Type
        The symbol's data type.
    kind: Kind
        The symbol's kind: what kind of thing does this symbol represent.
    mutability: Mutability
        The symbol's mutability status.
    def_span: TextSpan
        Where the symbol was defined.
    used: bool
        Whether or not the symbol is used in source code.
    ll_value: Optional[Value]
        The LLVM value that this symbol refers to.  This value is `None` until
        generation begins.
    '''

    class Kind(Enum):
        '''Enumerates the different kinds of symbols.'''

        VALUE = auto()
        FUNC = auto()

    class Mutability(Enum):
        '''
        Enumerates the different possible mutability statuses of symbols.  This
        status indicates both the symbol's defined mutability and it's inferred
        mutability from use.
        '''

        MUTABLE = auto()
        IMMUTABLE = auto()
        NEVER_MUTATED = auto()

    name: str
    parent_id: int
    type: Type
    kind: Kind
    mutability: Mutability
    def_span: TextSpan
    intrinsic: bool = False
    used: bool = False
    ll_value: Optional[Value] = None