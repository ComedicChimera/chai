'''Provides the definitions of Chai's abstract syntax tree.'''

from abc import ABC, abstractmethod
from dataclasses import dataclass, field
from enum import Enum, auto
from typing import List, Optional

from typecheck import Type
from report import TextSpan
from depm import Symbol

class ValueCategory(Enum):
    '''Enumeration of possible value categories of AST nodes.'''
    LVALUE = auto()
    RVALUE = auto()

class ASTNode(ABC):
    '''The abstract base class for all AST nodes.'''

    @property
    @abstractmethod
    def type(self) -> Type:
        '''Returns the yielded type of this AST node.'''

    @property
    @abstractmethod
    def span(self) -> TextSpan:
        '''Returns the text span over which this AST node extends.'''

    @property
    def category(self) -> ValueCategory:
        return ValueCategory.RVALUE

@dataclass
class FuncParam:
    '''
    Represents a parameter in a function or operator definition.

    Attributes
    ----------
    name: str
        The name of the function parameter.
    type: Type
        The type of the function parameter.
    mutated: bool
        Whether the parameter is ever mutated.
    '''

    name: str
    type: Type
    mutated: bool

@dataclass
class FuncDef(ASTNode):
    '''
    The AST node representing a function definition.
    
    Attributes
    ----------
    func_id: 'Identifier'
        The identifier representing the function's name.
    rt_type: Type
        The return type of the function.
    body: Optional[ASTNode]
        The function's optional body.
    func_params: List[FuncParam]
        The parameters to the function.
    '''

    func_id: 'Identifier'
    rt_type: Type
    body: Optional[ASTNode]
    _span: TextSpan
    func_params: List[FuncParam] = field(default_factory=list)

    @property
    def type(self) -> Type:
        return self.rt_type

    @property
    def span(self) -> TextSpan:
        return self._span

class Identifier(ASTNode):
    '''
    The AST node representing an identifier.

    Attributes
    ----------
    symbol: Symbol
        The symbol this identifier corresponds to.
    local: bool
        Whether the identifier is defined locally.
    '''

    symbol: Symbol
    local: bool
    _span: TextSpan

    @property
    def type(self) -> Type:
        return self.symbol.type

    @property
    def span(self) -> TextSpan:
        return self._span

    @property
    def category(self) -> ValueCategory:
        return ValueCategory.LVALUE

@dataclass
class Literal(ASTNode):
    '''
    The AST node representing a literal.

    Attributes
    ----------
    value: str
        The literal value stored in the AST node.
    '''

    value: str
    _type: Type
    _span: TextSpan

    @property
    def type(self) -> Type:
        return self._type

    @property
    def span(self) -> TextSpan:
        return self._span
