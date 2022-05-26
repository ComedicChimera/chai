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

    id: int
    parent: TypeVariable
    edges: Dict[int, 'Substitution']

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

    # The list of type variable in the solution graph.
    type_vars: List[TypeVariable]

    # The dictionary of substitutions organized by type variable.
    graph: Dict[int, Dict[int, Substitution]]

    # The counter used to generate new substitution IDs.
    id_counter: int

    # The table of type variables which are currently have no substitutions.
    unknowns: Dict[int, bool]

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

        tv = TypeVariable(len(self.type_vars), span, name)
        self.type_vars.append(tv)
        self.graph[tv.id] = {}
        self.unknowns[tv.id] = True
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
        del self.unknowns[tv.id]

        for overload in overloads:
            sub = BasicSubstitution(self.id_counter, tv, {}, overload)
            self.id_counter += 1
            self.graph[tv.id][sub.id] = sub

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

        del self.unknowns[tv.id]

        for overload in overloads:
            sub = OperatorSubstitution(self.id_counter, tv, {}, op, overload)
            self.id_counter += 1
            self.graph[tv.id][sub.id] = sub

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

        if not self.unify(None, lhs, rhs):
            self.error(f'type mismatch: {lhs} v. {rhs}', span)

        self.unknowns = {tv_id: remain for tv_id, remain in self.unknowns.items() if remain}

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

        for node in self.type_var_nodes:
            if len(node.edges) > 1 and node.default:
                for edge in node.edges[1:]:
                    self.prune_substitution(node, edge)

        for node in self.type_var_nodes:
            if len(node.edges) == 1:
                sub = node.edges[0].value
                node.value.value = sub.type
                sub.finalize()
            else:
                self.error(f'unable to infer type of `{node.value}`', node.value.span)

        for ca in self.cast_asserts:
            if not ca.src < ca.dest:
                self.error(f'cannot cast {ca.src} to {ca.dest}', ca.span)

    def reset(self):
        '''Resets the solver to its default state.'''

        self.type_vars = []
        self.graph = {}
        self.unknowns = {}
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
        if root:
            self.join_edges(root, tv_id, typ)
            return True    
        else:
            return self.prune_edges(tv_id, typ)

    def join_edges(self, root: Substitution, tv_id: int, sub_type: Type):
        if tv_id in self.unknowns:
            self.unknowns[tv_id] = False

            sub = BasicSubstitution(self.id_counter, self.type_vars[tv_id], {}, sub_type)
            self.id_counter += 1
            self.graph[tv_id][sub.id] = sub
            self.add_edge(root, sub)
        else:
            for edge_sub in self.graph[tv_id].values():
                if edge_sub.type == sub_type:
                    self.add_edge(root, edge_sub)

    def prune_edges(self, tv_id: int, typ: Type) -> bool:
        if tv_id in self.unknowns:
            self.unknowns[tv_id] = False

            # TODO
        else:
            matched_any = False
            for sub in self.graph[tv_id].values():
                if self.unify(sub, sub.type, typ):
                    matched_any = True
                else:
                    self.prune_substitution(sub)

            return matched_any

    def add_edge(self, a: Substitution, b: Substitution):
        a.edges[b.id] = b
        b.edges[a.id] = a

    def prune_substitution(self, sub: Substitution):
        for edge_sub in sub.edges.values():
            del edge_sub.edges[sub.id]

            self.prune_substitution(edge_sub)

        del self.graph[sub.parent.id][sub.id]

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
