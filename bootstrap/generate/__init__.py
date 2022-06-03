from typing import List

import llvm.value as llvalue
import llvm.types as lltypes
from llvm.module import Module
from depm.source import Package
from syntax.ast import *
from .type_util import conv_type, is_nothing
from .predicate import Predicate, PredicateGenerator

class Generator:
    pkg: Package
    pkg_prefix: str
    mod: Module

    pred_gen: PredicateGenerator
    predicates: List[Predicate]

    def __init__(self, pkg: Package):
        self.pkg = pkg
        self.pkg_prefix = f'p{pkg.id}.'

        self.mod = Module(pkg.name)

        self.pred_gen = PredicateGenerator()
        self.predicates = []

    def generate(self) -> Module:
        for file in self.pkg.files:
            for defin in file.definitions:
                match defin:
                    case FuncDef():
                        self.generate_func_def(defin)
                    case OperDef():
                        self.generate_oper_def(defin)

        for predicate in self.predicates:
            self.pred_gen.generate(predicate)

        return self.mod

    # ---------------------------------------------------------------------------- #

    def generate_func_def(self, fd: FuncDef):
        if 'intrinsic' in fd.annots:
            return

        mangle = True
        public = False

        for annot in fd.annots:
            match annot:
                case 'extern' | 'entry':
                    mangle = False
                    public = True

        if mangle:
            ll_name = self.pkg_prefix + fd.symbol.name
        else:
            ll_name = fd.symbol.name

        ll_func_type = lltypes.FunctionType(
            [conv_type(x.type) for x in fd.params],
            conv_type(fd.type.rt_type, rt_type=True)
        )

        ll_func = self.mod.add_function(ll_name, ll_func_type)

        ll_func.linkage = llvalue.Linkage.EXTERNAL if public else llvalue.Linkage.INTERNAL

        for param, ll_param in zip(fd.params, ll_func.params):
            ll_param.name = param.name
            param.ll_value = ll_param

        fd.symbol.ll_value = ll_func

        if fd.body:
            self.predicates.append(Predicate(ll_func, fd.params, not is_nothing(fd.type.rt_type), fd.body))

    def generate_oper_def(self, od: OperDef):
        if 'intrinsic' in od.annots:
            return

        ll_name = f'{self.pkg_prefix}.oper.overload.{od.overload.id}'

        ll_func_type = lltypes.FunctionType(
            [conv_type(x.type) for x in od.params],
            conv_type(od.type.rt_type, rt_type=True)
        )

        ll_func = self.mod.add_function(ll_name, ll_func_type)
        
        ll_func.linkage = llvalue.Linkage.INTERNAL

        for param, ll_param in zip(od.params, ll_func.params):
            ll_param.name = param.name
            param.ll_value = ll_param

        od.overload.ll_value = ll_func

        if od.body:
            self.predicates.append(Predicate(ll_func, od.params, not is_nothing(od.overload.signature.rt_type), od.body))


            





