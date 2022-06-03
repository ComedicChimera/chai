'''Provides the type solver and its associated constructs.'''

from dataclasses import dataclass, field
from typing import List, Dict, Optional, Set, Tuple
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
class TypeVarNode:
    type_var: TypeVariable
    sub_nodes: List['SubNode'] = field(default_factory=list)
    default: bool = False

    @property
    def id(self) -> int:
        return self.type_var.id

    def __eq__(self, other) -> bool:
        if isinstance(other, TypeVarNode):
            return self.id == other.id

        return False

@dataclass
class SubNode:
    id: int
    sub: Substitution
    parent: TypeVarNode
    edges: List['SubNode'] = field(default_factory=list)

    @property
    def type(self) -> Type:
        return self.sub.type

    @property
    def is_overload(self) -> bool:
        return len(self.parent.sub_nodes) > 1

    def __eq__(self, other: object) -> bool:
        if isinstance(other, SubNode):
            return self.id == other.id

        return False

    def __hash__(self) -> int:
        return self.id

# ---------------------------------------------------------------------------- #

@dataclass
class UnifyResult:
    unified: bool
    visited: Dict[int, bool] = field(default_factory=dict)
    completes: Set[int] = field(default_factory=set)

    def __and__(self, other: object) -> 'UnifyResult':
        if isinstance(other, UnifyResult):
            return UnifyResult(
                self.unified and other.unified, 
                {
                    k: self.visited.get(k, False) or other.visited.get(k, False)
                    for k in self.visited | other.visited
                },
                self.completes | other.completes,
            )

        raise TypeError()

    def __or__(self, other: object) -> 'UnifyResult':
        if isinstance(other, UnifyResult):
            return UnifyResult(
                self.unified or other.unified,
                {
                    k: self.visited.get(k, True) and other.visited.get(k, True)
                    for k in self.visited | other.visited
                },
                self.completes & other.completes
            )

        raise TypeError()

    def __bool__(self) -> bool:
        return self.unified

class Solver:
    '''
    Represents Chai's type solver.  The type solver is the primary mechanism for
    performing type deduction.  TODO describe type solving algorithm

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

    # The list of type variable nodes in the solution graph.
    type_var_nodes: List[TypeVarNode]

    # The dictionary of type substitution nodes in the solution graph.
    sub_nodes: Dict[int, SubNode]

    # The counter used to generate substitution IDs.
    sub_id_counter: int

    # The set of type variables that are complete: they cannot have more
    # substitutions added to them.
    completes: Set[int]

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

        tv = TypeVariable(len(self.type_var_nodes), span, name, self)
        self.type_var_nodes.append(TypeVarNode(tv))
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

        tv_node = self.type_var_nodes[tv.id]

        tv_node.default = True

        for overload in overloads:
            self.add_substitution(tv_node, BasicSubstitution(overload))

        self.completes.add(tv.id)
        
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

        tv_node = self.type_var_nodes[tv.id]

        for overload in overloads:
            self.add_substitution(tv_node, OperatorSubstitution(op, overload))   

        self.completes.add(tv.id)    

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

        result = self.unify(None, lhs, rhs)
        if not result:
            self.error(f'type mismatch: {lhs} v. {rhs}', span)

        for sid, prune in result.visited.items():
            if prune and sid in self.sub_nodes:
                self.prune_substitution(self.sub_nodes[sid], set())

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

        for tv_node in self.type_var_nodes:
            if tv_node.default and len(tv_node.sub_nodes) > 1:
                self.assert_equiv(tv_node.type_var, next(iter(tv_node.sub_nodes)).type, None)

        for tv_node in self.type_var_nodes:
            if len(tv_node.sub_nodes) == 1:
                sub = next(iter(tv_node.sub_nodes)).sub
                tv_node.type_var.value = sub.type
                sub.finalize()
            else:
                self.error(f'unable to infer type for {tv_node.type_var}', tv_node.type_var.span)

        for ca in self.cast_asserts:
            if not ca.src < ca.dest:
                self.error(f'cannot cast {ca.src} to {ca.dest}', ca.span)

    def reset(self):
        '''Resets the solver to its default state.'''

        self.type_var_nodes = []
        self.sub_nodes = {}
        self.sub_id_counter = 0
        self.completes = set()
        self.cast_asserts = []

    # ---------------------------------------------------------------------------- #

    def unify(self, root: Optional[SubNode], lhs: Type, rhs: Type) -> UnifyResult:
        match (lhs, rhs):
            case (TypeVariable(lhs_id), TypeVariable(rhs_id)):
                if lhs_id == rhs_id:
                    return UnifyResult(True)
                    
                return self.unify_type_var(root, lhs_id, rhs)
            case (TypeVariable(lhs_id), _):
                return self.unify_type_var(root, lhs_id, rhs)
            case (_, TypeVariable(rhs_id)):
                return self.unify_type_var(root, rhs_id, lhs)
            case (PointerType(lhs_elem), PointerType(rhs_elem)):
                return self.unify(root, lhs_elem, rhs_elem)
            case (FuncType(lhs_params, lhs_rt_type), FuncType(rhs_params, rhs_rt_type)):
                if len(lhs_params) != len(rhs_params):
                    return UnifyResult(False)

                result = UnifyResult(True)
                for lparam, rparam in zip(lhs_params, rhs_params):
                    result &= self.unify(root, lparam, rparam)
                    if not result:
                        return result

                return result & self.unify(root, lhs_rt_type, rhs_rt_type)
            case (PrimitiveType(), PrimitiveType()):
                return UnifyResult(lhs == rhs)
            case _:
                return UnifyResult(False)         

    def unify_type_var(self, root: Optional[SubNode], tv_id: int, typ: Type) -> UnifyResult:
        tv_node = self.type_var_nodes[tv_id]

        result = UnifyResult(True)

        if tv_id not in self.completes:
            result.completes.add(tv_id)

            for sub_node in tv_node.sub_nodes:
                if self.unify(sub_node, sub_node.type, typ):
                    break
            else:
                sub_node = self.add_substitution(tv_node, BasicSubstitution(typ))
                
                if root:
                    self.add_edge(root, sub_node)
        else:
            result.visited = {x.id: True for x in tv_node.sub_nodes}

            for i, sub_node in enumerate(tv_node.sub_nodes):
                uresult = self.unify(sub_node, sub_node.type, typ)
                if i == 0:
                    result &= uresult
                else:
                    result |= uresult

                if uresult:
                    result.visited[sub_node.id] = False

                    if root:
                        self.add_edge(root, sub_node)

        if not root or not root.is_overload:
            self.completes |= result.completes

        return result

    # ---------------------------------------------------------------------------- #

    def add_substitution(self, parent: TypeVarNode, sub: Substitution) -> SubNode:
        sub_node = SubNode(self.sub_id_counter, sub, parent)
        self.sub_id_counter += 1

        self.sub_nodes[sub_node.id] = sub_node
        parent.sub_nodes.append(sub_node)

        return sub_node

    def prune_substitution(self, sub_node: SubNode, pruning: Set[int]):
        pruning.add(sub_node.id)

        for edge in sub_node.edges:
            if edge.id not in pruning:
                self.prune_substitution(edge, pruning)

        sub_node.edges.clear()

        sub_node.parent.sub_nodes.remove(sub_node)

        del self.sub_nodes[sub_node.id]

    def add_edge(self, a: SubNode, b: SubNode):
        a.edges.append(b)
        b.edges.append(a)

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

        type_str = ' | '.join(repr(sub_node.type) for sub_node in self.type_var_nodes[tv.id].sub_nodes)

        if len(type_str) == 0:
            type_str = '_'
                
        return '{' + type_str + '}'
