'''Provides the type solver and its associated constructs.'''

from dataclasses import dataclass, field
from typing import List, Dict, Optional
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
    substitutions: Dict[int, 'SubNode'] = field(default_factory=dict)
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
    edges: Dict[int, 'SubNode'] = field(default_factory=dict)

    @property
    def type(self) -> Type:
        return self.sub.type

    @property
    def is_overload(self) -> bool:
        return len(self.parent.substitutions) > 1

    def __eq__(self, other: object) -> bool:
        if isinstance(other, SubNode):
            return self.id == other.id

        return False

# ---------------------------------------------------------------------------- #

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

    # The list of type variables which are still unknown: the boolean flag
    # indicates whether or not the variable is still an unknown at the end of a
    # type unification.  
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

        tv = TypeVariable(len(self.type_var_nodes), span, name, self)
        self.type_var_nodes.append(TypeVarNode(tv))
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

        tv_node = self.type_var_nodes[tv.id]

        tv_node.default = True

        for overload in overloads:
            self.add_substitution(tv_node, BasicSubstitution(overload))

        del self.unknowns[tv.id]
        
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

        del self.unknowns[tv.id]    

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
            if tv_node.default and len(tv_node.substitutions) > 1:
                self.unify(None, tv_node.type_var, next(iter(tv_node.substitutions.values())).type)

        for tv_node in self.type_var_nodes:
            if len(tv_node.substitutions) == 1:
                tv_node.type_var.value = next(iter(tv_node.substitutions.values())).type
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
        self.unknowns = {}
        self.cast_asserts = []

    # ---------------------------------------------------------------------------- #

    def unify(self, root: Optional[SubNode], lhs: Type, rhs: Type) -> bool:
        match (lhs, rhs):
            case (TypeVariable(lhs_id), TypeVariable(rhs_id)):
                if lhs_id == rhs_id:
                    return True
                    
                return self.link_type_vars(root, lhs_id, rhs_id)
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

    def link_type_vars(self, root: Optional[SubNode], lhs_id: int, rhs_id: int) -> bool:
        lhs_node, rhs_node = self.type_var_nodes[lhs_id], self.type_var_nodes[rhs_id]

        match (len(lhs_node.substitutions), len(rhs_node.substitutions)):
            case (0, 0):
                sub_node = self.add_substitution(lhs_node, rhs_node.type_var)

                if root:
                    self.add_edge(root, sub_node)

                self.unknowns[lhs_node.id] = False
            case (0, _):
                for sub_node in rhs_node.substitutions.values():
                    sub_node = self.add_substitution(lhs_node, sub_node.sub)

                    if root:
                        self.add_edge(root, sub_node)

                self.unknowns[lhs_node.id] = False
            case (_, 0):
                for sub_node in lhs_node.substitutions.values():
                    sub_node = self.add_substitution(rhs_node, sub_node.sub)

                    if root:
                        self.add_edge(root, sub_node)

                self.unknowns[rhs_node.id] = False
            case _:
                if root and root.is_overload:
                    return self.union_substitutions(root, lhs_node, rhs_node)
                else:
                    return self.intersect_substitutions(root, lhs_node, rhs_node)

        self.update_unknowns(root)
        return True

    def union_substitutions(self, root: SubNode, lhs_node: TypeVarNode, rhs_node: TypeVarNode) -> bool:
        matches = []

        for lhs_sub_node in lhs_node.substitutions.values():
            for rhs_sub_node in rhs_node.substitutions.values():
                if self.unify(root, lhs_sub_node.type, rhs_sub_node.type):
                    matches.append((lhs_sub_node, rhs_sub_node))

        for a, b in matches:
            self.add_edge(root, a)
            self.add_edge(root, b)

        self.update_unknowns(root)
        return len(matches) > 0         
    
    def intersect_substitutions(self, root: Optional[SubNode], lhs_node: TypeVarNode, rhs_node: TypeVarNode) -> bool:
        matches = []

        for lhs_sub_node in list(lhs_node.substitutions.values()):
            no_match = True

            for rhs_sub_node in list(rhs_node.substitutions.values()):
                if self.unify(root, lhs_sub_node.type, rhs_sub_node.type):
                    matches.append((lhs_sub_node, rhs_sub_node))
                    no_match = False
                else:
                    self.prune_substitution(rhs_sub_node)

            if no_match:
                self.prune_substitution(lhs_sub_node)

        if root:
            for a, b in matches:
                self.add_edge(root, a)
                self.add_edge(root, b)
        else:
            for a, b in matches:
                self.add_edge(a, b)   

        self.update_unknowns(root)
        return len(matches) > 0             

    def unify_type_var(self, root: Optional[SubNode], tv_id: int, typ: Type) -> bool:
        tv_node = self.type_var_nodes[tv_id]

        if len(tv_node.substitutions) == 0:
            sub_node = self.add_substitution(tv_node, BasicSubstitution(typ))
            self.unknowns[tv_id] = False

            if root:
                self.add_edge(root, sub_node)

            self.update_unknowns(root)
            return True
        elif root:
            for sub_node in list(tv_node.substitutions.values()):
                if sub_node == root:
                    continue

                ok = self.unify(sub_node, sub_node.type, typ)

                if ok:
                    self.add_edge(root, sub_node)
                    break
                elif root.is_overload:
                    continue
                else:
                    self.prune_substitution(sub_node)
            else:
                return False

            self.update_unknowns(root)
            return True
        else:
            for sub_node in list(tv_node.substitutions.values()):
                if not self.unify(sub_node, sub_node.type, typ):
                    self.prune_substitution(sub_node)

            # TODO prune all substitutions that aren't matched by any overload:
            # eg. (i64, i64) -> bool v. (i64, {i64 | i32, ...}) -> {_} should
            # prune all `i32`, ... but right now it doesn't

            self.update_unknowns(root)
            return len(tv_node.substitutions) > 0

    def update_unknowns(self, root: Optional[SubNode]):
        if not root or not root.is_overload:
            self.unknowns = {uid: remain for uid, remain in self.unknowns.items() if remain}

    # ---------------------------------------------------------------------------- #

    def add_substitution(self, parent: TypeVarNode, sub: Substitution) -> SubNode:
        sub_node = SubNode(self.sub_id_counter, sub, parent)
        self.sub_id_counter += 1

        self.sub_nodes[sub_node.id] = sub_node
        parent.substitutions[sub_node.id] = sub_node

        return sub_node

    def prune_substitution(self, sub_node: SubNode):
        for edge in list(sub_node.edges.values()):
            del edge.edges[sub_node.id]

        for eid, edge in list(sub_node.edges.items()):
            del sub_node.edges[eid]

            self.prune_substitution(edge)

        del sub_node.parent.substitutions[sub_node.id]

        del self.sub_nodes[sub_node.id]

    def add_edge(self, a: SubNode, b: SubNode):
        a.edges[b.id] = b
        b.edges[a.id] = a

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

        type_str = ' | '.join(repr(sub_node.type) for sub_node in self.type_var_nodes[tv.id].substitutions.values())

        if len(type_str) == 0:
            type_str = '_'
                
        return '{' + type_str + '}'
