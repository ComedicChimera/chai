from typing import Optional

import llvm.value as llvalue
import llvm.types as lltypes
from llvm.module import Module
from depm.source import Package
from syntax.ast import *

from .type_util import conv_type
from .predicate import BodyPredicate, PredicateGenerator
from .debug_info import DebugInfoEmitter

class Generator:
    pkg: Package
    pkg_prefix: str
    mod: Module

    pred_gen: PredicateGenerator
    die: Optional[DebugInfoEmitter]

    def __init__(self, pkg: Package, debug: bool):
        self.pkg = pkg
        self.pkg_prefix = f'p{pkg.id}.'

        self.mod = Module(pkg.name)

        self.pred_gen = PredicateGenerator()

        if debug:
            self.die = DebugInfoEmitter(pkg, self.mod)
        else:
            self.die = None

    def generate(self) -> Module:
        for src_file in self.pkg.files:
            if self.die:
                self.die.emit_file_header(src_file)

            for defin in src_file.definitions:
                match defin:
                    case FuncDef():
                        self.generate_func_def(defin)
                    case OperDef():
                        self.generate_oper_def(defin)

        self.pred_gen.generate()

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
            self.pred_gen.add_predicate(BodyPredicate(ll_func, fd.params, fd.body))

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
            self.pred_gen.add_predicate(BodyPredicate(ll_func, od.params, od.body))


            





