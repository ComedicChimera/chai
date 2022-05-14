'''Provides the type solver and its associated constructs.'''

from dataclasses import dataclass, field
from typing import List, Dict, Optional

from . import *
from report import TextSpan, CompileError
from depm.source import SourceFile

@typedataclass
class TypeVariable(Type):
    __match_args__ = ('id')

    id: int
    span: TextSpan
    display_name: Optional[str]
    
    parent: 'Solver'

    value: Optional[Type] = None

    def _compare(self, other: Type) -> bool:
        raise TypeError('unable to compare to an undetermined type variable')

    def _cast_from(self, other: Type) -> bool:
        raise TypeError('unable to check cast involving undetermined type variable')

    def inner_type(self) -> Type:
        assert self.value, 'unable to directly operate on undetermined type variable'

        return self.value

    def __repr__(self) -> str:
        if self.value:
            return repr(self.value)

        if not self.display_name:
            return self.parent.repr_overloaded_type_var(self)

        return self.display_name

@dataclass
class Substitution:
    _type: Type

    @property
    def type(self) -> Type:
        return self._type

    def finalize(self):
        pass

@dataclass
class OverloadSet:
    overloads: List[Substitution]
    default: bool

@dataclass
class CastAssert:
    src: Type
    dest: Type

# ---------------------------------------------------------------------------- #    

@dataclass
class SolutionContext:
    substitutions: Dict[int, Substitution] = field(default_factory=dict)
    overload_sets: Dict[int, OverloadSet] = field(default_factory=dict)

    def copy(self) -> 'SolutionContext':
        return SolutionContext(self.substitutions.copy(), self.overload_sets.copy())

    def merge(self, sub_ctx: 'SolutionContext'):
        self.substitutions |= sub_ctx.substitutions
        self.overload_sets |= sub_ctx.overload_sets

class Solver:
    srcfile: SourceFile

    type_vars: List[TypeVariable]

    global_ctx: SolutionContext
    local_ctx: SolutionContext

    cast_asserts: List[CastAssert]

    def __init__(self, srcfile: SourceFile):
        self.srcfile = srcfile
        self.type_vars = []
        self.global_ctx = SolutionContext()
        self.local_ctx = self.global_ctx
        self.cast_asserts = []

    def new_type_var(self, span: TextSpan, name: Optional[str] = None):
        tv = TypeVariable(len(self.type_vars), span, name, self)
        self.type_vars.append(tv)
        return tv

    def add_literal_overloads(self, tv: TypeVariable, *overloads: Type):
        self.global_ctx.overload_sets[tv.id] = OverloadSet([Substitution(typ) for typ in overloads], True)

    def assert_equiv(self, lhs: Type, rhs: Type, span: TextSpan):
        if not self.unify(lhs, rhs):
            self.error(f'type mismatch: {lhs} v. {rhs}', span)

    def assert_cast(self, src: Type, dest: Type):
        self.cast_asserts.append(CastAssert(src, dest))

    def solve(self):
        for tv in self.type_vars:
            if sub := self.global_ctx.substitutions.get(tv.id):
                tv.value = sub.type
                sub.finalize()
            elif (oset := self.global_ctx.overload_sets.get(tv.id)) and oset.default:
                tv.value = oset.overloads[0].type
                oset.overloads[0].finalize()
            else:
                self.error(f'unable to infer type of `{tv.display_name}`', tv.span)

        for ca in self.cast_asserts:
            if not ca.src < ca.dest:
                self.error(f'cannot cast {ca.src} to {ca.dest}')

    # ---------------------------------------------------------------------------- #

    def unify(self, lhs: Type, rhs: Type) -> bool:
        match (lhs, rhs):
            case (TypeVariable(lhs_id), TypeVariable(rhs_id)):
                return lhs_id == rhs_id
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
        if sub := self.get_substitution(id):
            return self.unify(sub.type, typ)
        elif oset := self.get_overload_set(id):
            return self.reduce_overloads(id, oset, typ)
        else:
            self.local_ctx.substitutions[id] = Substitution(typ)
            return True

    def get_substitution(self, id: int) -> Optional[Substitution]:
        if sub := self.local_ctx.substitutions.get(id):
            return sub
        elif sub := self.global_ctx.substitutions.get(id):
            return sub

    def get_overload_set(self, id: int) -> Optional[OverloadSet]:
        if oset := self.local_ctx.overload_sets.get(id):
            return oset
        elif oset := self.global_ctx.overload_sets.get(id):
            return oset

    def reduce_overloads(self, id: int, oset: OverloadSet, typ: Type) -> bool:
        outer_ctx = self.local_ctx
        
        new_overloads = []
        for overload in oset.overloads:
            self.local_ctx = outer_ctx.copy()

            if self.unify(overload.type, typ):
                new_overloads.append(overload)
                updated_ctx = self.local_ctx

        self.local_ctx = outer_ctx

        match len(new_overloads):
            case 0:
                return False
            case 1:
                self.local_ctx.merge(updated_ctx)
                self.local_ctx.substitutions[id] = new_overloads[0]
            case _:
                self.local_ctx.overload_sets[id] = OverloadSet(new_overloads, oset.default)

        return True

    def error(self, msg: str, span: TextSpan):
        raise CompileError(msg, self.srcfile.rel_path, span)

    def repr_overloaded_type_var(self, tv: TypeVariable) -> str:
        return '{' + ' | '.join(x.type for x in self.global_ctx.overload_sets[tv.id].overloads) + '}'
