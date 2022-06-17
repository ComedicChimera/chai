'''Provides Chai's definitions for symbols and scopes.'''

__all__ = [
    'Symbol',
    'Operator',
    'OperatorOverload'
]

from dataclasses import dataclass, field
from enum import Enum, auto
from typing import Optional, List, ClassVar

from report import TextSpan
from typecheck import Type, FuncType
from llvm.value import Value
from syntax.token import Token

@dataclass
class Symbol:
    '''
    Represents a semantic symbol: a named value or definition.

    Attributes
    ----------
    id: int
        The unique ID of the symbol.
    name: str
        The symbol's name.
    parent_id: int
        The ID of the package the symbol is defined in.
    file_number: int
        The number identifying the file this symbol is defined in.
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

    _id_counter: ClassVar[int] = 0

    class Kind(Enum):
        '''Enumerates the different kinds of symbols.'''

        VALUE = auto()
        TYPE = auto()

        def __str__(self) -> str:
            return self.name.lower()

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
    file_number: int
    type: Type
    kind: Kind
    mutability: Mutability
    def_span: TextSpan
    used: bool = False
    id: int = 0
    ll_value: Optional[Value] = None

    def __post_init__(self):
        '''Determine the unique ID of the symbol.'''

        self.id = Symbol._id_counter
        Symbol._id_counter += 1

@dataclass
class OperatorOverload:
    '''
    Represents a single overload of a particular operator.

    Attributes
    ----------
    parent_id: int
        The ID of the parent package to this overload.
    file_number: int
        The number identifying the file this overload is defined in.
    signature: Type
        The signature of the operator overload.
    def_span: TextSpan
        The span containing the operator symbol.
    id: int
        The unique ID of the overload: used to generate its name on the backend.
    intrinsic_name: str
        The instrinic generator associated with this overload if any.
    ll_value: Optional[Value]
        The LLVM value that this overload refers to.  This value is `None` until
        generation begins.
    '''

    parent_id: int
    file_number: int
    signature: Type
    def_span: TextSpan
    id: int = field(init=False)
    intrinsic_name: str = ''
    ll_value: Optional[Value] = None

    _id_counter: ClassVar[int] = 0

    def __post_init__(self):
        '''Determine the unique ID of the operator overload.'''

        self.id = OperatorOverload._id_counter
        OperatorOverload._id_counter += 1

    def conflicts(self, other: 'OperatorOverload') -> bool:
        '''
        Returns whether the signature of another operator overload conflicts
        with this one.
        '''

        match (self.signature, other.signature):
            case (FuncType(param_types=self_params), FuncType(param_types=other_params)):
                return self_params == other_params
            case _:
                raise TypeError(other)

@dataclass
class Operator:
    '''
    Represents all the visible definitions for an operator of a given arity.

    Attributes
    ----------
    kind: Token.Kind
        The token kind of the operator.
    op_sym: str
        The string value of the operator used when displaying it.
    arity: int
        The arity of the operator overloads associated with this operator.
    overloads: List[OperatorOverload]
        The overloads defined for this operator.
    '''

    kind: Token.Kind
    op_sym: str
    arity: int
    overloads: List[OperatorOverload]
