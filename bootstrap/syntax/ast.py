'''Provides the definitions of Chai's abstract syntax tree.'''

from abc import ABC, abstractmethod
from dataclasses import dataclass, field
from enum import Enum, auto
from typing import List, Optional, Dict, Tuple

from typecheck import Type, PrimitiveType, PointerType
from report import TextSpan
from depm import Symbol
from .token import Token

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

# ---------------------------------------------------------------------------- #

Annotations = Dict[str, Tuple[str, TextSpan]]

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
    mutated: bool = False

@dataclass
class FuncDef(ASTNode):
    '''
    The AST node representing a function definition.
    
    Attributes
    ----------
    func_id: 'Identifier'
        The identifier representing the function's name.
    body: Optional[ASTNode]
        The function's optional body.
    func_params: List[FuncParam]
        The parameters to the function.
    '''

    func_id: 'Identifier'
    func_params: List[FuncParam] 
    body: Optional[ASTNode]
    annots: Annotations
    _span: TextSpan

    @property
    def type(self) -> Type:
        return self.func_id.type.rt_type

    @property
    def span(self) -> TextSpan:
        return self._span

# ---------------------------------------------------------------------------- #

@dataclass
class Block(ASTNode):
    '''
    The AST node representing a procedural block of statements.

    Attributes
    ----------
    stmts: List[ASTNode]
        The statements that comprise this block.
    '''

    stmts: List[ASTNode]

    @property
    def type(self) -> Type:
        if len(self.stmts) > 0:
            return self.stmts[-1].type
        else:
            return PrimitiveType.NOTHING

    @property
    def span(self) -> TextSpan:
        return TextSpan.over(self.stmts[0].span, self.stmts[-1].span)

@dataclass
class VarList:
    '''
    The AST node representing a list of variables sharing the same initializer.

    Attributes
    ----------
    symbols: List[Symbol]
        The list of variable symbols declared by this variable list.
    initializer: Optional[ASTNode]
        The initializer for the variable list: may be `None` if there is no
        initializer.
    '''

    symbols: List[Symbol]
    initializer: Optional[ASTNode]

@dataclass
class VarDecl(ASTNode):
    '''
    The AST node representing a variable declaration.

    Attributes
    ----------
    var_lists: List[VarList]
        The list of variable lists declared by this variable declaration.
    '''

    var_lists: List[VarList]
    _span: TextSpan

    @property
    def type(self) -> Type:
        return PrimitiveType.NOTHING

    @property
    def span(self) -> TextSpan:
        return self._span

# ---------------------------------------------------------------------------- #

@dataclass
class TypeCast(ASTNode):
    '''
    The AST node representing a type cast.

    Attributes
    ----------
    src_expr: ASTNode
        The source expression being cast to another type.
    dest_type: Type
        The type to cast to.
    '''

    src_expr: ASTNode
    dest_type: Type
    _span: TextSpan

    @property
    def type(self) -> Type:
        return self.dest_type

    @property
    def span(self) -> TextSpan:
        return self._span

    @property
    def category(self) -> ValueCategory:
        return self.src_expr.category

@dataclass
class Indirect(ASTNode):
    '''
    The AST node representing an indirection operation.

    Attributes
    ----------
    elem: ASTNode
        The element being indirected.
    is_alloc: bool
        Whether this indirection acts as an allocation.
    '''

    elem: ASTNode
    _span: TextSpan

    @property
    def type(self) -> Type:
        return PointerType(self.elem.type)

    @property
    def span(self) -> TextSpan:
        return self._span

    @property
    def is_alloc(self) -> bool:
        return self.elem.category == ValueCategory.RVALUE

@dataclass
class Dereference(ASTNode):
    '''
    The AST node representing a pointer dereference.

    Attributes
    ----------
    ptr: ASTNode
        The pointer being dereferenced.
    '''

    __match_args__ = ('ptr', '_span')

    ptr: ASTNode
    _span: TextSpan
    elem_type: Type = PrimitiveType.NOTHING

    @property
    def type(self) -> Type:
        return self.elem_type

    @property
    def span(self) -> TextSpan:
        return self._span

    @property
    def category(self) -> ValueCategory:
        return ValueCategory.LVALUE

@dataclass
class FuncCall(ASTNode):
    '''
    The AST node representing a function call.

    Attributes
    ---------
    func: ASTNode
        The function being called.
    args: List[ASTNode]
        The arguments to pass to the function.
    '''

    __match_args__ = ('func', 'args', '_span')

    func: ASTNode
    args: List[ASTNode]
    _span: TextSpan
    rt_type: Type = PrimitiveType.NOTHING

    @property
    def type(self) -> Type:
        return self.rt_type

    @property
    def span(self) -> TextSpan:
        return self._span

# ---------------------------------------------------------------------------- #

@dataclass
class Identifier(ASTNode):
    '''
    The AST node representing an identifier.

    Attributes
    ----------
    name: str
        The name of the identifier.
    local: bool
        Whether the identifier is defined locally.
    symbol: Symbol
        The symbol this identifier corresponds to.
    '''

    __match_args__ = ('name', '_span')

    name: str
    _span: TextSpan

    local: bool = False
    symbol: Optional[Symbol] = None

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

    kind: Token.Kind
    value: str
    _span: TextSpan
    _type: Type = PrimitiveType.NOTHING

    @property
    def type(self) -> Type:
        return self._type

    @type.setter
    def type(self, typ: Type):
        self._type = typ

    @property
    def span(self) -> TextSpan:
        return self._span

@dataclass
class Null(ASTNode):
    '''
    The AST node representing the language constant `null`.
    '''

    _span: TextSpan
    _type: Type = PrimitiveType.NOTHING

    @property
    def type(self) -> Type:
        return self._type

    @type.setter
    def type(self, typ: Type):
        self._type = typ

    @property
    def span(self) -> TextSpan:
        return self._span
