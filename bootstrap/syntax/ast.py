'''Provides the definitions of Chai's abstract syntax tree.'''

from abc import ABC, abstractmethod
from dataclasses import dataclass, field
from enum import Enum, auto
from typing import List

from typecheck import Type
from report import TextSpan

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
