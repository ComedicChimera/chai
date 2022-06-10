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
    'AppliedOperator',
    'Annotations',
    'FuncDef',
    'OperDef',
    'CondBranch',
    'IfTree',
    'WhileLoop',
    'Block',
    'VarList',
    'VarDecl',
    'Assignment',
    'IncDecStmt',
    'KeywordStmt',
    'ReturnStmt',
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

@dataclass
class AppliedOperator:
    '''
    A utility structure representing a specific application of an operator.

    Attributes
    ----------
    token: Token
        The operator token corresponding to this application.
    overload: Optional[OperatorOverload]
        The determined operator overload corresponding to this application: this
        is `None` until it is deduced by the solver.
    '''

    token: Token
    overload: Optional[OperatorOverload] = None

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
    op_sym: str
        The string value of the operator used when displaying it.
    overload: OperatorOverload
        The operator overload corresponding to this operator definition. 
    params: List[Symbol]
        The parameters to the operator.
    body: Optional[ASTNode]
        The operator's optional body.
    annots: Annotations
        The operator's annotations.
    '''

    op_sym: str
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
class CondBranch:
    '''
    Represents a single conditional branch in an if/elif/else tree.

    Attributes
    ----------
    header_var: Optional['VarDecl']
        The header variable of the if/elif/else statement if it exists.
    condition: ASTNode
        The branch condition.
    body: ASTNode
        The body to execute if the condition is true.
    '''

    header_var: Optional['VarDecl']
    condition: ASTNode
    body: ASTNode

@dataclass
class IfTree(ASTNode):
    '''
    The AST node representing an if/elif/else tree.

    Attributes
    ----------
    cond_branches: List[CondBranch]
        The list of condition (if/elif) branches of the if tree.
    else_branch: Optional[ASTNode]
        The default (else) branch of the if tree.
    rt_type: Type
        The returned type of the if tree.
    '''

    cond_branches: List[CondBranch]
    else_branch: Optional[ASTNode]
    _span: TextSpan

    rt_type: Type = PrimitiveType.NOTHING

    @property
    def type(self) -> Type:
        return self.rt_type

    @property
    def span(self) -> TextSpan:
        return self._span

@dataclass
class WhileLoop(ASTNode):
    '''
    The AST node representing a while loop.

    Attributes
    ----------
    header_var: Optional['VarDecl']
        The header variable of the loop if it exists.
    condition: ASTNode
        The loop condition.
    update_stmt: Optional[ASTNode]
        The statement to be run at the end of each loop iteration if it exists.
    body: ASTNode
        The code to execute while the condition is true.
    rt_type: Type = PrimitiveType.NOTHING
        The returned type of the loop.
    '''

    header_var: Optional['VarDecl']
    condition: ASTNode
    update_stmt: Optional[ASTNode]
    body: ASTNode
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
    compound_ops: List[AppliedOperator]
        The list of applied compound operators corresponding to each LHS-RHS
        pair.  This list is empty if there are no compound operators.
    '''

    lhs_exprs: List[ASTNode]
    rhs_exprs: List[ASTNode]
    compound_ops: List[AppliedOperator]

    @property
    def type(self) -> Type:
        return PrimitiveType.NOTHING

    @property
    def span(self) -> TextSpan:
        return TextSpan.over(self.lhs_exprs[0].span, self.rhs_exprs[-1].span)

@dataclass
class IncDecStmt(ASTNode):
    '''
    The AST node representing an increment or a decrement statement.

    Attributes
    ----------
    lhs_operand: ASTNode
        The value being incremented or decremented.
    op: AppliedOperator
        The applied `+`/`-` operator.
    '''

    lhs_operand: ASTNode
    op: AppliedOperator

    @property
    def type(self) -> Type:
        return PrimitiveType.NOTHING

    @property
    def span(self) -> TextSpan:
        return TextSpan.over(self.lhs_operand.span, self.op.token.span)

@dataclass
class KeywordStmt(ASTNode):
    '''
    The AST node representing a statement denoted by a single keyword such as
    `break` or `continue`.

    Attributes
    ----------
    keyword: Token
        The keyword comprising the statement.
    '''

    keyword: Token

    @property
    def type(self) -> Type:
        return PrimitiveType.NOTHING

    @property
    def span(self) -> TextSpan:
        return self.keyword.span

@dataclass
class ReturnStmt(ASTNode):
    '''
    The AST node represent a return statement.

    Attributes
    ----------
    exprs: List[ASTNode]
        The list of expression values being returned.
    '''

    exprs: List[ASTNode]
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
class BinaryOpApp(ASTNode):
    '''
    The AST node representing a binary operator application.

    Attributes
    ----------
    op: AppliedOperator
        The applied binary operator.
    lhs: ASTNode
        The LHS operand.
    rhs: ASTNode
        The RHS operand.
    rt_type: Type
        The yielded type of the operation.
    '''

    op: AppliedOperator
    lhs: ASTNode
    rhs: ASTNode

    rt_type: Type = PrimitiveType.NOTHING

    @property
    def type(self) -> Type:
        return self.rt_type

    @property
    def span(self) -> TextSpan:
        return TextSpan.over(self.lhs.span, self.rhs.span)

@dataclass
class UnaryOpApp(ASTNode):
    '''
    The AST node representing a unary operator application.

    Attributes
    ----------
    op: AppliedOperator
        The applied unary operator.
    operand: ASTNode
        The operand of the application.
    rt_type: Type
        The yielded type of the operation.
    overload: Optional[OperatorOverload]
        The unary operator overload corresponding to the application.  This is
        `None` until it is determined by the Walker.
    '''

    op: AppliedOperator
    operand: ASTNode
    _span: TextSpan
    
    rt_type: Type = PrimitiveType.NOTHING

    @property
    def type(self) -> Type:
        return self.rt_type

    @property
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
    const: bool
        Whether the indirection is explicitly constant.
    ptr_type: Type
        The pointer type produced by the indirection.
    is_alloc: bool
        Whether this indirection acts as an allocation.
    '''

    __match_args__ = ('elem', 'const', '_ptr_type', '_span')

    elem: ASTNode
    const: bool
    _span: TextSpan

    ptr_type: Type = PrimitiveType.NOTHING

    @property
    def type(self) -> Type:
        return self.ptr_type

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
