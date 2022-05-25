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
        The unique ID of the type variable.
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
class Overload:
    '''
    Represents a single possible overload in an overload set.

    Attributes
    ----------
    substitution: Substitution
        The substitution associated with this overload.
    context: 'SolutionContext'
        The persistent solution context that is associated with this overload:
        it is accumulated through many test unifications.
    '''

    substitution: Substitution
    context: 'SolutionContext' = None

    def __post_init__(self):
        if not self.context:
            self.context = SolutionContext()

@dataclass
class OverloadSet:
    '''
    Represents a set of possible substitutions for a type variable.

    Attributes
    ----------
    overloads: List[Substitution]
        The possible overloads for the type variable.
    default: bool
        Whether to select the first element of substitution's list as the final
        type value for the associated type variable if no other type value can
        be inferred.  This mechanism is used by literals.
    '''

    overloads: List[Overload]
    default: bool

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
class SolutionContext:
    '''
    Represents a set of possible substitutions and overload sets that result
    from a particular unification or set of unifications.  Solution contexts
    effectively contain (part of) the state of the solver at a particular time
    thereby allowing the solver to consider multiple possible sequences of
    deductions ie. perform "test unifications."  This construct is particularly
    useful when deducing the types of overloaded type variables as it allows the
    solver to "prune" possibilities based on what it can infer about the program
    so far.

    Attributes
    ----------
    substitutions: Dict[int, Substitution]
        The applied substitutions in this solution context.
    overload_sets: Dict[int, OverloadSet]
        The applied overload sets in this solution context.

    Methods
    -------
    update(sub_ctx: SolutionContext)  
    copy() -> SolutionContext
    '''

    substitutions: Dict[int, Substitution] = field(default_factory=dict)
    overload_sets: Dict[int, OverloadSet] = field(default_factory=dict)
    
    def update(self, sub_ctx: 'SolutionContext'):
        '''
        Updates this context with the deductions made in another context.

        Params
        ------
        sub_ctx: SolutionContext
            The context to update from.
        '''

        self.substitutions |= sub_ctx.substitutions
        self.overload_sets |= sub_ctx.overload_sets

    def copy(self) -> 'SolutionContext':
        '''Returns a copy of this solution context.'''

        return SolutionContext(self.substitutions.copy(), self.overload_sets.copy())

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

    # The list of type variables defined in the global solution context.
    type_vars: List[TypeVariable]

    # The global solution context: the finalized result of all unifications.
    global_ctx: SolutionContext

    # The local solution context: used to temporarily store results of
    # unifications until they are either discarded or used to update a parent
    # context.
    local_ctx: SolutionContext

    # The list of cast assertions applied in the global solution context.
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

        tv = TypeVariable(len(self.type_vars), span, name, self)
        self.type_vars.append(tv)
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

        self.global_ctx.overload_sets[tv.id] = OverloadSet(
            [Overload(BasicSubstitution(typ)) for typ in overloads],
            True
        )

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

        self.global_ctx.overload_sets[tv.id] = OverloadSet(
            [Overload(OperatorSubstitution(op, overload)) for overload in overloads],
            False
        )

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

        # NOTE The unification for an equivalency assertion takes place in a
        # "blank" local context (instead of starting in the global context) to
        # minimize copying during overload pruning.  If we were to use the
        # global context here, any overload pruning that happened in that base
        # context would have to copy the entire global context for each overload
        # which far exceeds the slight performance cost of using a blank local
        # context for every assertion.

        # Unify to test equivalency: report an error if it fails.
        if not self.unify(lhs, rhs):
            self.error(f'type mismatch: {lhs} v. {rhs}', span)

        # If unification succeeds, update the global context with the results of
        # the unification and clear the local context for the next equivalency
        # assertion.
        self.global_ctx.update(self.local_ctx)
        self.local_ctx = SolutionContext()

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

        for tv in self.type_vars:
            if (oset := self.global_ctx.overload_sets.get(tv.id)) and oset.default:
                self.substitute_overload(tv.id, oset.overloads[0])
                self.global_ctx.update(self.local_ctx)
                self.local_ctx = SolutionContext()

        for tv in self.type_vars:
            if sub := self.global_ctx.substitutions.get(tv.id):
                tv.value = sub.type
                sub.finalize()
            else:
                self.error(f'unable to infer type of `{tv}`', tv.span)

        for ca in self.cast_asserts:
            if not ca.src < ca.dest:
                self.error(f'cannot cast {ca.src} to {ca.dest}', ca.span)

    def reset(self):
        '''Resets the solver to its default state.'''

        self.type_vars = []
        self.global_ctx = SolutionContext()
        self.local_ctx = SolutionContext()
        self.cast_asserts = []

    # ---------------------------------------------------------------------------- #

    def unify(self, lhs: Type, rhs: Type) -> bool:
        '''
        Attempts to make two types equal by substitution and overload pruning.
        The results of its deductions are stored in the local solution context.

        Params
        ------
        lhs: Type
            The LHS type.
        rhs: Type
            The RHS type.

        Returns
        -------
        ok: bool
            Whether unification succeeded.
        '''

        match (lhs, rhs):
            case (TypeVariable(lhs_id), TypeVariable(rhs_id)):
                if lhs_id == rhs_id:
                    return True

                return self.unify_type_var(lhs_id, rhs)
            case (TypeVariable(lhs_id), _):
                return self.unify_type_var(lhs_id, rhs)
            case (_, TypeVariable(rhs_id)):
                return self.unify_type_var(rhs_id, lhs)
            case (PointerType(lhs_elem), PointerType(rhs_elem)):
                return self.unify(lhs_elem, rhs_elem)
            case (FuncType(lhs_params, lhs_rt_type), FuncType(rhs_params, rhs_rt_type)):
                if len(lhs_params) != len(rhs_params):
                    return False

                for lparam, rparam in zip(lhs_params, rhs_params):
                    if not self.unify(lparam, rparam):
                        return False

                return self.unify(lhs_rt_type, rhs_rt_type)
            case (PrimitiveType(), PrimitiveType()):
                return lhs == rhs
            case _:
                return False
        
    def unify_type_var(self, id: int, typ: Type) -> bool:
        '''
        Performs unification involving a type variable.

        Params
        ------
        id: int
            The ID of the type variable to unify against.
        typ: Type
            The type to unify the type variable with.

        Returns
        -------
        ok: bool
            Whether type variable unification succeeded.
        '''

        if sub := self.get_substitution(id):
            return self.unify(sub.type, typ)
        elif oset := self.get_overload_set(id):
            return self.prune_overloads(id, oset, typ)
        else:
            self.local_ctx.substitutions[id] = BasicSubstitution(typ)
            return True

    def get_substitution(self, id: int) -> Optional[Substitution]:
        '''
        Returns the substitution corresponding to the given type variable if it
        exists in any context.

        Params
        ------
        id: int
            The ID of the type variable whose substitution to look up.
        '''

        if sub := self.local_ctx.substitutions.get(id):
            return sub
        elif sub := self.global_ctx.substitutions.get(id):
            return sub

    def get_overload_set(self, id: int) -> Optional[OverloadSet]:
        '''
        Returns the overload set corresponding to the given type variable if it
        exists in any context.

        Params
        ------
        id: int
            The ID of the type variable whose overload set to look up.
        '''

        if oset := self.local_ctx.overload_sets.get(id):
            return oset
        elif oset := self.global_ctx.overload_sets.get(id):
            return oset

    def prune_overloads(self, id: int, oset: OverloadSet, typ: Type) -> bool:
        '''
        Attempts to unify an overload set with a type by pruning all types in
        the overload set which cannot be individually unified with the type.  If
        a single type remains after pruning, then that type becomes the
        substitution for the type variable associated with the given overload
        set. If no substitutions remain, the pruning fails.  Otherwise, the
        local overload set is updated to contain only the overloads which
        survived pruning.

        Params
        ------
        id: int
            The ID of the type variable whose overload set is being pruned.
        oset: OverloadSet
            The overload set to prune.
        typ: Type
            The type to unify with the overload set.
        
        Returns
        -------
        ok: bool
            Whether pruning succeeded.
        '''

        outer_ctx = self.local_ctx
        
        new_overloads = []
        for overload in oset.overloads:
            self.local_ctx = outer_ctx.copy()
            self.local_ctx.update(overload.context)

            if self.unify(overload.substitution.type, typ):
                new_overloads.append(Overload(overload.substitution, self.local_ctx))

        self.local_ctx = outer_ctx

        match len(new_overloads):
            case 0:
                return False
            case 1:
                self.substitute_overload(id, new_overloads[0])
            case _:
                # TODO figure out how to make the relationship between the context elements
                # and the operator bidirectional
                self.local_ctx.overload_sets[id] = OverloadSet(new_overloads, oset.default)

        return True

    def substitute_overload(self, id: int, overload: Overload):
        '''
        Unifies an overload as the substitution for a type variable.  This
        function assumes the overload is valid as a substitution for the type
        variable.

        Params
        ------
        id: int
            The type variable whose substitution to update.
        overload: Overload
            The overload to apply as a substitution.
        '''

        self.local_ctx.update(overload.context)
        self.local_ctx.substitutions[id] = overload.substitution

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

        if sub := self.get_substitution(tv.id):
            return repr(sub.type)
        elif oset := self.get_overload_set(tv.id):
            return '{' + ' | '.join(str(x.type) for x in oset.overloads) + '}'
        else:
            return '{_}'
