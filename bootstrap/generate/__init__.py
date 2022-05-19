from lib2to3.pytree import generate_matches
from typing import List

import llvm.ir as ir
import llvm.value as llvalue
import llvm.types as lltypes
from llvm.module import Module
from depm.source import Package
from syntax.ast import *
from .type_util import conv_type
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

        for predicate in self.predicates:
            self.pred_gen.generate(predicate)

        return self.mod

    # ---------------------------------------------------------------------------- #

    def generate_func_def(self, fd: FuncDef):
        mangle = True

        for annot in fd.annots:
            match annot:
                case 'intrinsic':
                    return
                case 'extern' | 'entry':
                    mangle = False

        if mangle:
            ll_name = self.pkg_prefix + fd.ident.name
        else:
            ll_name = fd.ident.name

        ll_func_type = lltypes.FunctionType(
            [conv_type(x.type) for x in fd.params],
            conv_type(fd.type, rt_type=True)
        )

        ll_func = self.mod.add_function(ll_name, ll_func_type)

        ll_func.linkage = llvalue.Linkage.INTERNAL
        ll_func.attr_set.add(ir.Attribute(kind=ir.Attribute.Kind.NO_UNWIND))

        for param, ll_param in zip(fd.params, ll_func.params):
            ll_param.name = param.name
            param.ll_value = ll_param

        fd.ident.symbol.ll_value = ll_func

        if fd.body:
            self.predicates.append(Predicate(ll_func, fd.params, fd.body))



