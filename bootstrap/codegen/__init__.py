'''
The main backend of the Chai compiler: responsible for converting Chai into LLVM.
'''

__all__ = ['Generator']

from typing import Optional

import llvm.value as llvalue
import llvm.types as lltypes
from llvm.module import Module as LLModule
from depm.source import Package
from syntax.ast import *

from .type_util import conv_type
from .predicate import BodyPredicate, PredicateGenerator
from .debug_info import DebugInfoEmitter

class Generator:
    '''
    Responsible for converting a Chai package into an LLVM module.

    Methods
    -------
    generate() -> LLModule
    '''

    # The package being converted to LLVM.
    pkg: Package

    # The global package prefix used for name mangling: by adding this prefix,
    # we prevent link collisions between identically named functions defined in
    # different packages.
    pkg_prefix: str

    # The LLVM module being generated from the package.
    mod: LLModule

    # The debug info emitter: this may be None if no debug info is to be emitted.
    die: Optional[DebugInfoEmitter]

    # The predicate generator: used to convert expressions into LLVM IR.
    pred_gen: PredicateGenerator

    def __init__(self, pkg: Package, debug: bool):
        '''
        Params
        ------
        pkg: Package
            The package to generate.
        debug: bool
            Whether to emit debug information.
        '''

        self.pkg = pkg
        self.pkg_prefix = f'p{pkg.id}.'

        self.mod = LLModule(pkg.name)

        if debug:
            self.die = DebugInfoEmitter(pkg, self.mod)
        else:
            self.die = None

        self.pred_gen = PredicateGenerator(self.die)

    def generate(self) -> LLModule:
        '''
        Generates the package and returns the produced LLVM module.
        '''

        # Start by generating all the global declarations.
        for src_file in self.pkg.files:
            if self.die:
                self.die.emit_file_header(src_file)

            for defin in src_file.definitions:
                match defin:
                    case FuncDef():
                        self.generate_func_def(defin)
                    case OperDef():
                        self.generate_oper_def(defin)

        # Then, generate all the predicates of those global declarations: since
        # predicate bodies can refer to any global declarations, we have to
        # generate them last so that all the declarations are visible.
        self.pred_gen.generate()

        # Finalize debug info if it exists.
        if self.die:
            self.die.finalize()

        # Return the completed module.
        return self.mod

    # ---------------------------------------------------------------------------- #

    def generate_func_def(self, fd: FuncDef):
        '''
        Generate a function definition.

        Params
        ------
        fd: FuncDef
            The function definition to generate.
        '''

        # TODO implementation of intrinsic functions
        if 'intrinsic' in fd.annots:
            return

        # Whether to apply standard name mangling.
        mangle = True

        # Whether this function definition should be marked external.
        public = False

        for annot in fd.annots:
            match annot:
                case 'extern' | 'entry':
                    # Both @extern and @entry stop name mangling and make the
                    # symbol public.  @entry because the name of the entry point
                    # is hard-coded in the linker flags and @extern because the
                    # name of the function has to match the name of the external
                    # symbol.
                    mangle = False
                    public = True

        # Determine the LLVM name based on whether or not it should be mangled
        # from the Chai name.
        if mangle:
            ll_name = self.pkg_prefix + fd.symbol.name
        else:
            ll_name = fd.symbol.name

        # Create the LLVM function type.
        ll_func_type = lltypes.FunctionType(
            [conv_type(x.type) for x in fd.params],
            conv_type(fd.type.rt_type, rt_type=True)
        )

        # Create the LLVM function.
        ll_func = self.mod.add_function(ll_name, ll_func_type)

        # Mark it as external if necessary.
        ll_func.linkage = llvalue.Linkage.EXTERNAL if public else llvalue.Linkage.INTERNAL

        # Assign all the function parameters their corresponding LLVM values and
        # update the LLVM parameters with their appropriate names.
        for param, ll_param in zip(fd.params, ll_func.params):
            ll_param.name = param.name
            param.ll_value = ll_param

        # Assign the function its corresponding LLVM function value.
        fd.symbol.ll_value = ll_func

        # Add the function body as a predicate to generate if it exists.
        if fd.body:
            self.pred_gen.add_predicate(BodyPredicate(ll_func, fd.params, fd.body))

        # Emit function debug info as necessary.
        if self.die:
            self.die.emit_function_info(fd, ll_name)

    def generate_oper_def(self, od: OperDef):
        '''
        Generate an operator definition.

        Params
        ------
        od: OperDef
            The operator definition to generate.
        '''

        # Intrinsic operators are not generated: their applications are replaced
        # with LLVM code snippets.
        if 'intrinsic' in od.annots:
            return

        # All operators have the LLVM name `oper.overload.[overload_id]`.
        ll_name = f'{self.pkg_prefix}.oper.overload.{od.overload.id}'

        # Generate the function type for the operator definition.
        ll_func_type = lltypes.FunctionType(
            [conv_type(x.type) for x in od.params],
            conv_type(od.type.rt_type, rt_type=True)
        )

        # Generate the LLVM function for the operator definition.
        ll_func = self.mod.add_function(ll_name, ll_func_type)
        
        # TODO add support for exported operators.
        ll_func.linkage = llvalue.Linkage.INTERNAL

        # Assign all the operator parameters their corresponding LLVM values and
        # update the LLVM parameters with their appropriate names.
        for param, ll_param in zip(od.params, ll_func.params):
            ll_param.name = param.name
            param.ll_value = ll_param

        # Assign the overload its corresponding LLVM function value.
        od.overload.ll_value = ll_func

        # Add the overload body as a predicate to generate if it exists.
        if od.body:
            self.pred_gen.add_predicate(BodyPredicate(ll_func, od.params, od.body))

        # Emit operator debug info if necessary.
        if self.die:
            self.die.emit_oper_info(od, ll_name)


            





