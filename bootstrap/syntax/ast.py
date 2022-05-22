'''Provides the definitions of Chai's abstract syntax tree.'''

from abc import ABC, abstractmethod
from dataclasses import dataclass, field
from enum import Enum, auto
from typing import List, Optional, Dict, Tuple

from typecheck import Type, PrimitiveType, PointerType
from report import TextSpan
from depm import Symbol, OperatorOverload
from .token import Token

__all__ = [
    'ValueCategory',
    'ASTNode',
    'Annotations',
    'FuncDef',
    'OperDef',
    'Block',
    'VarList',
    'VarDecl',
    'Assignment',
    'IncDecStmt',
    'TypeCast',
    'BinaryOpApp',
    'UnaryOpApp',
    'Indirect',
    'Dereference',
    'FuncCall',
    'Identifier',
    'Literal',
    'Null'
]

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
class FuncDef(ASTNode):
    '''
    The AST node representing a function definition.
    
    Attributes
    ----------
    symbol: Symbol
        The symbol corresponding to the function.
    params: List[Symbol]
        The parameters to the function.
    body: Optional[ASTNode]
        The function's optional body.
    annots: Annotations
        The function's annotations.
    '''

    symbol: Symbol
    params: List[Symbol] 
    body: Optional[ASTNode]
    annots: Annotations
    _span: TextSpan

    @property
    def type(self) -> Type:
        return self.symbol.type

    @property
    def span(self) -> TextSpan:
        return self._span

@dataclass
class OperDef(ASTNode):
    '''
    The AST node representing an operator definition.

    Attributes
    ----------
    overload: OperatorOverload
        The operator overload corresponding to this operator definition. 
    params: List[Symbol]
        The parameters to the operator.
    body: Optional[ASTNode]
        The operator's optional body.
    annots: Annotations
        The operator's annotations.
    '''

    overload: OperatorOverload
    params: List[Symbol]
    body: Optional[ASTNode]
    annots: Annotations
    _span: TextSpan

    @property
    def type(self) -> Type:
        return self.oper.signature

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

@dataclass
class Assignment(ASTNode):
    '''
    The AST node representing an assignment operation.

    Attributes
    ----------
    lhs_exprs: List[ASTNode]
        The expressions being assigned to.
    rhs_exprs: List[ASTNode]
        The values being assigned.
    compound_op_token: Optional[Token]
        The compound operator token (if it exists).
    compound_op_overload: List[OperatorOverload]
        The list of compound operator overloads (for each LHS-RHS pair).
    '''

    lhs_exprs: List[ASTNode]
    rhs_exprs: List[ASTNode]

    compound_op_token: Optional[Token] = None
    compound_op_overloads: List[OperatorOverload] = field(default_factory=list)

    @property
    def type(self) -> Type:
        return PrimitiveType.NOTHING

    @property
    def span(self) -> TextSpan:
        return TextSpan.over(self.lhs_exprs[0].span, self.rhs_exprs[-1].span)

@dataclass
class IncDecStmt:
    '''
    The AST node representing an increment or a decrement statement.

    Attributes
    ----------
    lhs_operand: ASTNode
        The value being incremented or decremented.
    op_token: Token
        The increment/decrement token.
    overload: Optional[OperatorOverload] = None
        The corresponding `+`/`-` operator overload.
    '''

    lhs_operand: ASTNode
    op_token: Token
    overload: Optional[OperatorOverload] = None

    @property
    def type(self) -> Type:
        return PrimitiveType.NOTHING

    @property
    def span(self) -> TextSpan:
        return TextSpan.over(self.lhs_operand.span, self.op_token.span)

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
class BinaryOpApp(ASTNode):
    '''
    The AST node representing a binary operator application.

    Attributes
    ----------
    op_token: Token
        The operator token.
    lhs: ASTNode
        The LHS operand.
    rhs: ASTNode
        The RHS operand.
    rt_type: Type
        The yielded type of the operation.
    overload: Optional[OperatorOverload]
        The binary operator overload corresponding to the application.  This is
        `None` until it is determined by the Walker.
    '''

    __match_args__ = ('op_token', 'overload', 'lhs', 'rhs', 'rt_type')

    op_token: Token
    lhs: ASTNode
    rhs: ASTNode

    rt_type: Type = PrimitiveType.NOTHING
    overload: Optional[OperatorOverload] = None

    def type(self) -> Type:
        return self.rt_type

    def span(self) -> TextSpan:
        return TextSpan.over(self.lhs.span, self.rhs.span)

@dataclass
class UnaryOpApp(ASTNode):
    '''
    The AST node representing a unary operator application.

    Attributes
    ----------
    op_token: Token
        The operator token.
    operand: ASTNode
        The operand of the application.
    rt_type: Type
        The yielded type of the operation.
    overload: Optional[OperatorOverload]
        The unary operator overload corresponding to the application.  This is
        `None` until it is determined by the Walker.
    '''

    __match_args__ = ('op_token', 'overload', 'operand', 'rt_type')

    op_token: Token
    operand: ASTNode
    _span: TextSpan
    
    rt_type: Type = PrimitiveType.NOTHING
    overload: Optional[OperatorOverload] = None

    def type(self) -> Type:
        return self.rt_type

    def span(self) -> TextSpan:
        return self._span

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
    elem_type: Type
        The element type of the pointer.
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

    Attributes
    ----------
    span: TextSpan
        The span over which the constant null occurs.
    type: Type
        The type of the null value.
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
