'''The module responsible for emitting debug information.'''

__all__ = ['DebugInfoEmitter']

import os
from contextlib import contextmanager
from typing import Optional

import llvm.debug as lldbg
import llvm.metadata as llmeta
from report import TextSpan
from depm import *
from depm.source import *
from syntax.ast import *
from typecheck import *
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

    # The current local scope of the debug builder.
    local_scope: Optional[llmeta.DIScope]

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

        # Initialize the local lexical scope.
        self.local_scope = None

    def finalize(self):
        '''Finalizes debug info.'''

        self.dib.finalize()

    # ---------------------------------------------------------------------------- #

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

    def emit_function_info(self, fd: FuncDef, mangled_name: str):
        '''
        Emits the debug information for a function definition.
        
        Params
        ------
        fd: FuncDef
            The function definition whose debug information to emit.
        mangled_name: str
            The mangled name of the function being defined.
        '''

        self.dib.create_function(
            self.di_pkg_scope,
            self.di_file,
            fd.symbol.name,
            mangled_name,
            fd.span.start_line,
            self.as_di_type(fd.type),
            'extern' not in fd.annots,
            bool(fd.body),
            fd.body.span.start_line if fd.body else fd.span.start_line,
        )

    def emit_oper_info(self, od: OperDef, mangled_name: str):
        '''
        Emits the debug information for an operator definition.

        Params
        ------
        od: OperDef
            The operator definition whose debug information to emit.
        mangled_name: str
            The mangled name of the operator function being defined.
        '''

        self.dib.create_function(
            self.di_pkg_scope,
            self.di_file,
            od.op_sym,
            mangled_name,
            od.span.start_line,
            self.as_di_type(od.overload.signature),
            # TODO update to be external when necessary
            True,
            bool(od.body),
            od.body.span.start_line if od.body else od.span.start_line,
        )

    # ---------------------------------------------------------------------------- #

    def emit_param_info(self, ndx: int, sym: Symbol):
        '''
        Emits the debug information for a variable declaration.

        Params
        ------
        ndx: int
            The parameter's index in the argument list.
        sym: Symbol
            The parameter symbol.
        '''

        self.dib.create_param_variable(
            self.scope, 
            self.di_file, 
            sym.name,
            ndx + 1,
            sym.def_span.start_line,
            self.as_di_type(sym.type),
        )

    @contextmanager
    def emit_scope(self, span: TextSpan):
        '''
        Emits a new debug scope.

        Params
        ------
        span: TextSpan
            The starting text span of the debug scope.

        Returns
        -------
        ContextManager[None]
            A context manager used to manage the scope.
        '''

        scope = self.dib.create_lexical_block(self.scope, self.di_file, span.start_line, span.start_col)

        outer_scope = self.local_scope
        self.local_scope = scope

        yield

        self.local_scope = outer_scope

    # ---------------------------------------------------------------------------- #

    def as_di_type(self, typ: Type) -> llmeta.DIType:
        '''
        Converts a Chai data type into a DI equivalent.

        Params
        ------
        typ: Type
            The type to convert.
        '''

        match typ:
            case PrimitiveType():
                pass
            case PointerType(elem_type):
                return self.dib.create_pointer_type(elem_type, typ.bit_size, typ.bit_align)
            case FuncType(param_types):
                return self.dib.create_subroutine_type(self.di_file, *map(self.as_di_type, param_types))

    def as_di_location(self, span: Optional[TextSpan]) -> Optional[llmeta.DILocation]:
        '''Returns the given text span as a debug location.'''
        
        if span:
            return llmeta.DILocation(self.scope, span.start_line, span.start_col)

        return None

    # ---------------------------------------------------------------------------- #

    @property
    def scope(self) -> llmeta.DIScope:
        '''Returns the current enclosing scope.'''
        
        return self.local_scope or self.di_pkg_scope
