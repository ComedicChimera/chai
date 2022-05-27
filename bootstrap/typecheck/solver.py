'''Provides the type solver and its associated constructs.'''

from dataclasses import dataclass, field
from typing import List, Dict, Optional, Tuple, TypeVar
from abc import ABC, abstractmethod

from . import *
from report import TextSpan
from report.reporter import CompileError
from depm import OperatorOverload
from depm.source import SourceFile
from syntax.ast import AppliedOperator

@typedataclass
class TypeVariable(Type):
    '''
    Represents a type variable -- an undetermined type.

    Attributes
    ----------
    id: int
        The unique node ID of the type variable.
    span: TextSpan
        The span to error on if the type variable cannot be inferred.
    display_name: Optional[str]
        The representative string to display in error message when this type
        variable is encountered.
    value: Optional[Type]
        The final, inferred type for the type variable.
    '''

    __match_args__ = ('id',)

    id: int
    span: TextSpan
    display_name: Optional[str]
    _parent: 'Solver'

    default: bool = False
    value: Optional[Type] = None

    def _compare(self, other: Type) -> bool:
        raise TypeError('unable to compare to an undetermined type variable')

    def _cast_from(self, other: Type) -> bool:
        raise TypeError('unable to check cast involving undetermined type variable')

    def inner_type(self) -> Type:
        assert self.value, 'unable to directly operate on undetermined type variable'

        return self.value.inner_type()

    def __repr__(self) -> str:
        if self.value:
            return repr(self.value)

        if not self.display_name:
            return self._parent.repr_unnamed_type_var(self)

        return self.display_name

# ---------------------------------------------------------------------------- #

@dataclass
class Substitution(ABC):
    '''
    Represents the solver's "guess" for what type should be inferred for a given
    type variable.

    Attributes
    ----------
    type: Type
        The guessed type.
    
    Methods
    -------
    finalize()
        Called when this substitution's value is inferred as the final value for
        a type variable.  This method should be used to handle any side-effects
        of type inference (eg. setting the package ID of the operator that was
        selected).
    '''

    @property
    @abstractmethod
    def type(self) -> Type:
        pass

    def finalize(self):
        pass

@dataclass
class BasicSubstitution(Substitution):
    '''
    A general purpose substitution: used for all types which don't require
    special substitution logic.
    '''

    _type: Type

    @property
    def type(self) -> Type:
        return self._type

    def finalize(self):
        pass

@dataclass
class OperatorSubstitution(Substitution):
    '''
    A specialized kind of substitution for operator overloads.
    
    Attributes
    ----------
    op: AppliedOperator
        The applied operator generating this substitution. This is where the
        determined overload is stored.
    overload: OperatorOverload
        The operator overload associated with this substitution.
    '''
    
    op: AppliedOperator
    overload: OperatorOverload

    @property
    def type(self) -> Type:
        return self.overload.signature

    def finalize(self):
        self.op.overload = self.overload

# ---------------------------------------------------------------------------- #

@dataclass
class CastAssert:
    '''
    Represents an assertion that a given cast is possible between two types.

    Attributes
    ----------
    src: Type
        The type being casted from.
    dest: Type
        The type being casted to.
    span: TextSpan
        The text span where the cast occurs in source text.
    '''

    src: Type
    dest: Type
    span: TextSpan

# ---------------------------------------------------------------------------- #

@dataclass
class Node:
    id: int
    value_edges: List['ValueEdge'] = field(default_factory=list, init=False)
    tv_edges: List['TypeVarEdge'] = field(default_factory=list, init=False)

@typedataclass
class NodeVar(Type):
    node_id: int = -1
    value: Type = PrimitiveType.NOTHING

    # TODO: NodeVar type definition

@dataclass
class Edge:
    lhs_pattern: Type
    rhs_pattern: Type

@dataclass
class ValueNode(Node):
    value: Substitution

@dataclass
class ValueEdge(Edge):
    node: ValueNode

@dataclass
class TypeVarNode(Node):
    type_var: TypeVariable

@dataclass
class TypeVarEdge(Edge):
    node: TypeVarNode 

# ---------------------------------------------------------------------------- #

@dataclass
class TypeVarEdge:
    edge_pattern: Type
    node: TypeVarNode

class Solver:
    '''
    Represents Chai's type solver.  The type solver is the primary mechanism for
    performing type deduction: it is based on the Hindley-Milner type
    inferencing algorithm -- the algorithm used by the solver has been extended
    to support overloading as well as several other complexities of Chai's type
    system.

    Methods
    -------
    new_type_var(span: TextSpan, name: Optional[str] = None) -> TypeVariable
    add_literal_overloads(tv: TypeVariable, *overloads: Type)     
    assert_equiv(lhs: Type, rhs: Type, span: TextSpan)
    assert_cast(src: Type, dest: Type, span: TextSpan)
    solve()
    reset()
    '''

    # The source file in which this solver is operating.
    src_file: SourceFile

    # The type solution graph.
    graph: List[Node]

    # The list of type variable node IDs.
    type_var_ids: List[int]

    # The counter used to generate new node IDs.
    id_counter: int

    # The list of applied cast assertions.
    cast_asserts: List[CastAssert]

    def __init__(self, src_file: SourceFile):
        '''
        Params
        ------
        srcfile: SourceFile
            The source file this solver is operating in.
        '''

        self.src_file = src_file

        # Prime the solver to begin accepting constraints.
        self.reset()

    def new_type_var(self, span: TextSpan, name: Optional[str] = None) -> TypeVariable:
        '''
        Creates a new type variable in global context.

        Params
        ------
        span: TextSpan
            The span over which to error if the type value of the type variable
            cannot be inferred.
        name: Optional[str]
            The (optional) display name for the type variable.
        '''

        tv = TypeVariable(self.id_counter, span, name, self)
        self.graph.append(TypeVarNode(tv.id, tv))
        self.id_counter += 1
        return tv

    def add_literal_overloads(self, tv: TypeVariable, overloads: List[Type]):
        '''
        Binds an overload set for a literal (ie. defaulting overload set)
        comprised of the given type overloads to the given type variable.

        Params
        ------
        tv: TypeVariable
            The type variable to bind to.
        *overloads: Type
            The list of type overloads for the literal.
        '''

        tv.default = True
        


    def add_operator_overloads(self, tv: TypeVariable, op: AppliedOperator, overloads: List[OperatorOverload]):
        '''
        Binds an overload set for an operator application comprised of the given
        operator overloads to the given type variable.

        Params
        ------
        tv: TypeVariable
            The type variable to bind to.
        op_ast: OperatorAST
            The AST node of the operator application generating this overload
            constraint.
        *overloads: OperatorOverload
            The list of operator overloads for the operator.
        '''        

    def assert_equiv(self, lhs: Type, rhs: Type, span: TextSpan):
        '''
        Asserts that two types are equivalent.

        Params
        ------
        lhs: Type
            The LHS type.
        rhs: Type
            The RHS type.
        span: TextSpan
            The text span to error over if the assertion fails.
        '''

    def assert_cast(self, src: Type, dest: Type, span: TextSpan):
        '''
        Asserts that one type can be cast to another.

        Params
        ------
        src: Type
            The type to cast from.
        dest: Type
            The type to cast to.
        span: TextSpan
            The text span to error over if the assertion fails.
        '''

        self.cast_asserts.append(CastAssert(src, dest, span))

    def solve(self):
        '''
        Prompts the solver to make its final type deductions based on all the
        constraints it has been given -- this assumes that no more relevant
        constraints will be provided.  This does NOT reset the solver.
        '''

        # TODO

        for ca in self.cast_asserts:
            if not ca.src < ca.dest:
                self.error(f'cannot cast {ca.src} to {ca.dest}', ca.span)

    def reset(self):
        '''Resets the solver to its default state.'''

        self.graph = {}
        self.type_var_ids = []
        self.id_counter = 0
        self.cast_asserts = []

    # ---------------------------------------------------------------------------- #

    def unify(self, root: Optional[Substitution], lhs: Type, rhs: Type) -> bool:
        match (lhs, rhs):
            case (TypeVariable(lhs_id), TypeVariable(rhs_id)):
                if lhs_id == rhs_id:
                    return True
                    
                # TODO
            case (TypeVariable(lhs_id), _):
                return self.unify_type_var(root, lhs_id, rhs)
            case (_, TypeVariable(rhs_id)):
                return self.unify_type_var(root, rhs_id, lhs)
            case (PointerType(lhs_elem), PointerType(rhs_elem)):
                return self.unify(root, lhs_elem, rhs_elem)
            case (FuncType(lhs_params, lhs_rt_type), FuncType(rhs_params, rhs_rt_type)):
                if len(lhs_params) != len(rhs_params):
                    return False

                for lparam, rparam in zip(lhs_params, rhs_params):
                    if not self.unify(root, lparam, rparam):
                        return False

                return self.unify(root, lhs_rt_type, rhs_rt_type)
            case (PrimitiveType(), PrimitiveType()):
                return lhs == rhs
            case _:
                return False

    def unify_type_var(self, root: Optional[Substitution], tv_id: int, typ: Type) -> bool:
        pass

    # ---------------------------------------------------------------------------- #

    def error(self, msg: str, span: TextSpan):
        '''
        Raise a compile error indicating a type solution failure.

        Params
        ------
        msg: str
            The error message.
        span: TextSpan
            The text span of the erroneous source text.
        '''

        raise CompileError(msg, self.src_file, span)

    def repr_unnamed_type_var(self, tv: TypeVariable) -> str:
        '''
        Returns the string representation of an unnamed type variable.

        Params
        ------
        tv: TypeVariable
            The type variable whose string representation to return.
        '''

        # TODO
