'''Provides the type solver and its associated constructs.'''

from dataclasses import dataclass, field
from typing import List, Dict, Optional, Set
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

    @property
    def size(self) -> int:
        assert self.value, 'unable to calculate size of an undetermined type variable'

        return self.value.size

    @property
    def align(self) -> int:
        assert self.value, 'unable to calculate alignment of an undetermined type variable'

        return self.value.align

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
    '''
    Represents a type variable within the solution graph.

    Attributes
    ----------
    type_var: TypeVariable
        The type variable represented by this node.
    sub_nodes: List[SubNode]
        The substitution nodes associated with this type variable.
    default: bool
        Whether or not to default to the first substitution in the list of
        substitution nodes if there is more than one possible substitution when
        finalizing type deduction for this type variable.
    id: int
        The unique ID of the type variable node (which is the same as the ID of
        the type variable it represents).
    '''

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
    '''
    Represents a substitution within the solution graph.

    Attributes
    ----------
    id: int
        The unique ID of this substitution node.
    sub: Substitution
        The substitution represented by this substitution node.
    parent: TypeVarNode
        The parent type variable node to this substitution: the type variable
        for which this node represents a possible substitution.
    edges: List[SubNode]
        The list of substitution nodes this node shares an edge with.
    type: Type 
        The resultant type of the substitution associated with this node.
    is_overload: bool
        Whether or not the substitution node is an overload of the type variable
        it is associated with: ie. whether its parent type variable has more
        than one possible substitution.
    '''

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
    '''
    Represents the complex result of unification.

    Attributes
    ----------
    unified: bool
        Whether or not unification was successful.
    visited: Dict[int, bool]
        The map of substitution node IDs visited during substitution associated
        with whether or not they should be pruned once the equivalency assertion
        which caused this unification completes (true => prune).
    completes: Set[int]
        The set of nodes that were completed (ie. given substitutions) by this
        unification.
    '''

    unified: bool
    visited: Dict[int, bool] = field(default_factory=dict)
    completes: Set[int] = field(default_factory=set)

    def __and__(self, other: object) -> 'UnifyResult':
        '''
        Combines two unification results in such a way that both unification
        must hold in order for the yielded unification result to hold.
        '''

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
        '''
        Combines two unification results in such a way that either unification
        can hold in order for the yielded unification result to hold.
        '''

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
    performing type deduction.  

    Chai's type deduction algorithm is loosely based on the Algorithm J for the
    Hindley-Milner type system.  However, it has been extended and redesigned to
    support generalized type overloading.

    The algorithm works by considering a *solution graph* comprised of
    *type variable nodes* and *substitution nodes*.  The type variable nodes
    represent the undetermined types of the given solution context (eg. inside a
    given function).  The substitution nodes represent the possible values for
    those undetermined type variables.  Each type variable node has a number
    of substitution nodes associated with it that represent the possible types
    this node can have.  
    
    These substitution nodes can be determined either before unification
    involving that type variable begins or during the unifications involving the
    type variable: type variable nodes which are *complete* can no longer has
    substitutions added to them.

    All substitution nodes have *edges* which represent relationships between
    substitution nodes.  More precisely, if two substitution nodes, A and B,
    share an edge, then substitution A is valid if and only if substitution B is
    valid.

    The principle mechanism of Chai's type deduction algorithm is *unification*:
    the process by which the solver attempts to make two types equal by
    unification. This process fails if the two types cannot be made equal (eg.
    are of a different shape or represent different primitive types).  The
    wrinkle to this algorithm that Chai's type solver adds (extending from
    Algorithm J) is that the unification algorithm also takes into account the
    substitution that led to the unification: eg. if we are unifying a type
    against a substitution of a type variable, different behavior may occur.
    This substitution node is called the *unification root* and may not be
    present if the unification is occurring at the top level (ie. between two
    types which are not currently substitutions of any type variable).

    The exact formulation of the unification algorithm can be inferred from the
    implementation below using the context and terminology established above.

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

        # Get the type variable node associated with the given type variable.
        tv_node = self.type_var_nodes[tv.id]

        # Set it to default (since literal overloads always default).
        tv_node.default = True

        # Add all the substitutions to the type variable node.
        for overload in overloads:
            self.add_substitution(tv_node, BasicSubstitution(overload))

        # Mark the type variable node as complete.
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

        # Get the type variable node associated with the given type variable.
        tv_node = self.type_var_nodes[tv.id]

        # Add all the substitutions to the type variable node.
        for overload in overloads:
            self.add_substitution(tv_node, OperatorSubstitution(op, overload))   

        # Mark the type variable node as complete.
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

        # Unify the given LHS type with the RHS type with no unification root,
        result = self.unify(None, lhs, rhs)

        # Raise an error if unification fails.
        if not result:
            self.error(f'type mismatch: {lhs} v. {rhs}', span)

        # Prune all nodes which the unification algorithm marked for pruning.
        for sid, prune in result.visited.items():
            # Note that we need to check to make sure the `sid` has not already
            # been pruned through its connection to another pruned node.
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

        # Unify the first type substitution for any type variable nodes which
        # should default and have more than one remaining possible substitution.
        for tv_node in self.type_var_nodes:
            if tv_node.default and len(tv_node.sub_nodes) > 1:
                # We use `assert_equiv` to perform the unification so we can
                # avoid rewriting all the boiler-plate code written inside
                # `assert_equiv` for top level unification, but we pass in a
                # position of `None` since this operation *should* never fail.
                self.assert_equiv(tv_node.type_var, next(iter(tv_node.sub_nodes)).type, None)

        # Go through each type variable and make final deductions based on
        # remaining nodes in the solution graph.  
        for tv_node in self.type_var_nodes:
            # Any remaining type variable which has exactly one substitution
            # associated with it is considered solved.
            if len(tv_node.sub_nodes) == 1:
                sub = next(iter(tv_node.sub_nodes)).sub
                tv_node.type_var.value = sub.type
                sub.finalize()
            else:
                # Otherwise, report an appropriate error.
                self.error(f'unable to infer type for {tv_node.type_var}', tv_node.type_var.span)

        # Apply all cast assertions.
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
        '''
        Makes two types equal to each other by substitution with respect to the
        given root substitution.  

        Params
        ------
        root: Optional[SubNode]
            The unification root substitution.  This can be `None` if there is
            no root.
        lhs: Type
            The left-hand type.
        rhs: Type
            The right-hand type.
        '''

        # NOTE For all composite types, we use `&` to combine results of
        # sub-unifications since all sub-results must be true in order for
        # unification to succeed.

        # Match the shape of the types first.
        match (lhs, rhs):
            case (TypeVariable(lhs_id), TypeVariable(rhs_id)):
                # Make sure we don't unify a type variable with itself.
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
        '''
        Attempts to make a type variable equal to a given type by substitution
        with respect to a given root substitution.

        Params
        ------
        root: Optional[SubNode]
            The unification root substitution.  This can be `None` if there is
            no root.
        tv_id: int
            The type variable ID.
        typ: Type
            The type to unify the type variable with.
        '''

        tv_node = self.type_var_nodes[tv_id]

        result = UnifyResult(True)

        # Handle incomplete type variables.
        if tv_id not in self.completes:
            # Mark the type variable as completed by this unification.
            result.completes.add(tv_id)

            # Add a substitution of the given type to the type variable if it
            # does not already have a substitution which is equivalent to the
            # given type.
            for sub_node in tv_node.sub_nodes:
                if self.unify(sub_node, sub_node.type, typ):
                    break
            else:
                sub_node = self.add_substitution(tv_node, BasicSubstitution(typ))

             # Add an edge between the root and matching/new substitution.
            if root:
                self.add_edge(root, sub_node) 
        # Handle complete type variables.
        else:
            # Go through each substitution of the given type variable.
            for i, sub_node in enumerate(tv_node.sub_nodes):
                # Unify the substitution with the inputted type making the
                # substitution the root of unification.
                sub_result = self.unify(sub_node, sub_node.type, typ)

                # Merge the substitution unification result with the global
                # unification result.
                if i == 0:
                    result = sub_result
                else:
                    result |= sub_result

                # If the local unification succeeded, ...
                if sub_result:
                    # Indicate that this substitution node should not be pruned.
                    result.visited[sub_node.id] = False

                    # If there is a root, add an edge between this substitution
                    # and that unification root.
                    if root:
                        self.add_edge(root, sub_node)
                else:
                    # Otherwise, indicate that this substitution should be pruned.
                    result.visited[sub_node.id] = True

        # If we are not dealing with an overload substitution, update the global
        # list of completed type variables with the completed type variables of
        # the overall result.
        if not root or not root.is_overload:
            self.completes |= result.completes

        # Return the generated unification result.
        return result

    # ---------------------------------------------------------------------------- #

    def add_substitution(self, parent: TypeVarNode, sub: Substitution) -> SubNode:
        '''
        Adds a substitution node to a type variable node.

        Params
        ------
        parent: TypeVarNode
            The type variable node to add the substitution to.
        sub: Substitution
            The substitution to add as a substitution node.

        Returns
        -------
        SubNode
            The created substitution node.
        '''

        # Create the new substitution node.
        sub_node = SubNode(self.sub_id_counter, sub, parent)
        self.sub_id_counter += 1

        # Add it to the graph and to the given parent.
        self.sub_nodes[sub_node.id] = sub_node
        parent.sub_nodes.append(sub_node)

        # Return the newly created substitution node.
        return sub_node

    def prune_substitution(self, sub_node: SubNode, pruning: Set[int]):
        '''
        Removes a substitution node from the solution graph as well as all
        substitutions which are related to it by an edge.

        Params
        ------
        sub_node: SubNode
            The substitution node to prune.
        pruning: Set[int]
            The set of IDs of substitution nodes already being pruned: to
            prevent cyclic pruning.  This should be an empty set when called
            externally.
        '''

        # First, mark the current node as already being pruned so that we don't
        # try to prune it multiple times and get caught in a cycle while pruning
        # the graph.
        pruning.add(sub_node.id)

        # Go through each edge of the current substitution node and prune them
        # if they are not already being pruned.
        for edge in sub_node.edges:
            if edge.id not in pruning:
                self.prune_substitution(edge, pruning)

        # Remove the substitution from its parent type variable node.
        sub_node.parent.sub_nodes.remove(sub_node)

        # Remove the substitution from the substitution graph.
        del self.sub_nodes[sub_node.id]

    def add_edge(self, a: SubNode, b: SubNode):
        '''
        Relates two substitution nodes by an edge.

        Params
        ------
        a: SubNode
            The first node in the pair.
        b: SubNode
            The second node in the pair.
        '''

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
