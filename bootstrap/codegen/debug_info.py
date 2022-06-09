import os

import llvm.debug as lldbg
import llvm.metadata as llmeta
from depm.source import *
from llvm.module import Module as LLModule

class DebugInfoEmitter:
    '''Responsible for emitting LLVM debug info.'''

    # The LLVM debug info builder.
    dib: lldbg.DIBuilder

    # The debug info entry for the package's global scope.
    di_pkg_scope: llmeta.DIScope

    # The debug info entry for the current file.
    di_file: llmeta.DIFile

    # The debug info entry for the file's local scop0e.
    di_file_scope: llmeta.DIScope

    def __init__(self, pkg: Package, mod: LLModule):
        '''
        Params
        ------
        pkg: Package
            The Chai package debug info is being emitted for.
        mod: LLModule
            The LLVM module being generated.
        '''

        self.dib = lldbg.DIBuilder(mod)

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

    def emit_file_header(self, src_file: SourceFile):
        '''
        Emits the debug information header for a source file.

        Params
        ------
        src_file: SourceFile
            The source file whose debug header to emit.
        '''

        # Create the debug info file and file local scope.
        self.di_file = self.dib.create_file(src_file.abs_path, os.getcwd())
        self.di_file_scope = self.di_file.as_scope()
