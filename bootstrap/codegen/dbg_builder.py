import os
from typing import Callable, List
from functools import wraps

import llvm.debug as lldbg
import llvm.metadata as llmeta

from . import *
from .base_builder import BaseLLBuilder

@staticmethod
def debug_ext(f: Callable) -> Callable:
    '''Extends a method defined by BaseLLBuilder to have additional logic.'''

    @wraps(f)
    def wrapper(self, *args, **kwargs):
        r = getattr(BaseLLBuilder, f.__name__)(self, *args, **kwargs)
        f(self, *args, **kwargs)
        return r

    return wrapper

class DebugLLBuilder(BaseLLBuilder):
    '''
    An extension of the base LLBuilder that builds debug information in addition
    to performing basic code generation.
    '''

    # The LLVM debug info builder.
    dib: lldbg.DIBuilder


    # The debug info entry for the package's global scope.
    di_pkg_scope: llmeta.DIScope

    # The debug info entry for the current file.
    di_file: llmeta.DIFile

    # The debug info entry for the file's local scop0e.
    di_file_scope: llmeta.DIScope

    def __init__(self, pkg: Package):
        '''
        Params
        ------
        pkg: Package
            The package being used to generate LLVM IR.
        '''

        super().__init__(pkg)

        self.dib = lldbg.DIBuilder(self.mod)

        # Create the psuedo DIFile used to represent the entire Chai package in
        # DWARF since Chai treats package as single compile units. This file
        # also corresponds to the global scope of the Chai package.
        di_pkg_file = self.dib.create_file(pkg.abs_path, os.getcwd())
        self.di_pkg_scope = di_pkg_file.as_scope()

        # Create the compile unit entry for the Chai package.
        self.dib.create_compile_unit(
            di_pkg_file,
            # Chai doesn't have its own DWARF source language tag yet.
            lldbg.DWARFSourceLanguage.C99,
            "pybs-chaic v0.3.0"
        )

    # ---------------------------------------------------------------------------- #

    @debug_ext
    def begin_file(self, src_file: SourceFile):
        # Create the debug info file and file local scope.
        self.di_file = self.dib.create_file(src_file.abs_path, os.getcwd())
        self.di_file_scope = self.di_file.as_scope()

    @debug_ext
    def build_func_decl(
        self, 
        name: str, 
        params: List[Symbol], 
        rt_type: Type,
        external: bool = False,
        call_conv = ''
    ) -> ir.Function:
        # TODO generate function debug info
        pass