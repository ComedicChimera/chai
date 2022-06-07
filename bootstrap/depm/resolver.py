'''Provides Chai's symbol resolver.'''

__all__ = ['Resolver']

from typing import Dict, List

from . import *
from .source import Package, SourceFile
from report import TextSpan
from report.reporter import Reporter, CompileError

class Resolver:
    '''
    The resolver is responsible for resolving global symbols, checking for
    infinite definitions, and checking for operator collisions: basically, it is
    responsible for validating the global namespace of the project. It operates
    on all packages at once since often times resolution, particularly for
    cyclically dependent packages, relies on being able to resolve symbols
    across multiple packages not in package-order.

    Methods
    -------
    resolve()
    '''

    reporter: Reporter
    dep_graph: Dict[int, Package]

    def __init__(self, reporter: Reporter, dep_graph: Dict[int, Package]):
        '''
        Params
        ------
        reporter: Reporter
            The compiler's global error reporter.
        dep_graph: Dict[int, Package]
            The package dependency graph.
        '''
        
        self.reporter = reporter
        self.dep_graph = dep_graph

    def resolve(self) -> bool:
        '''
        Runs the full symbol resolution pass.

        Returns
        -------
        bool: whether resolution succeeded  
        '''

        # TODO resolve global symbols and check for infinite definitions

        self.check_operator_conflicts()

        return self.reporter.return_code == 0

    # ---------------------------------------------------------------------------- #

    def check_operator_conflicts(self):
        '''Checks for operator conflicts in global and file namespaces.'''

        for pkg in self.dep_graph.values():
            for ops in pkg.operator_table.values():
                # TODO update to call with imported operators as well
                for op in ops:
                    self.find_colliding_overloads(op.op_sym, op.overloads)

    def find_colliding_overloads(self, op_sym: str, overloads: List[OperatorOverload]):
        '''
        Finds and reports any overloads of an operator which conflict.

        Params
        ------
        op_sym: str
            The operator symbol of the operator whose overloads are being
            searched (for error reporting).
        overloads: List[OperatorOverload]
            The overloads to search for conflicts.
        '''

        # This weird for loop pattern is designed to give us all possible unique
        # combinations of overloads without duplication or pairs of the same
        # overload.  Basically, the idea is that we start with the first
        # overload and compare it to every overload after it.  Then, we move to
        # the second overload and compare it to every overload after it -- we
        # can skip the first overload since the first iteration already did that
        # comparison. This process can be carried out to the end of the overload
        # list. We use the `-1` on the outer for loop just to avoid an
        # unnecessary iteration (the i+1 means the inner loop doesn't run when i
        # = len(overloads-1).  This algorithm also deterministically reports
        # overloads which occur earlier in the list as conflicting with
        # overloads later in the list: this property allows us to choose the
        # order of errors.
        for i, a in enumerate(overloads[:-1]):
            for b in overloads[i+1:]:
                if a.conflicts(b):
                    self.error(
                        self.get_src_file(a.id, b.id),
                        f'operator definition for {op_sym} conflicts with another visible definition ' + 
                        f'for the same operator: {a.signature} v {b.signature}',
                        a.def_span
                    )

    # ---------------------------------------------------------------------------- #

    def get_src_file(self, pkg_id: int, file_number: int) -> SourceFile:
        '''
        Get the source file object corresponding to a package ID and file number.

        Params
        ------
        pkg_id: int
            The parent package ID of the source file.
        file_number: int
            The file number within the parent package of the source file.
        '''

        return self.dep_graph[pkg_id].files[file_number]

    def error(self, src_file: SourceFile, msg: str, span: TextSpan):
        '''
        Reports an error in the given source file.  This does NOT raise an
        exception.

        Params
        ------
        src_file: SourceFile
            The file the error occurred in.
        msg: str
            The error message.
        span: TextSpan
            The text span of erroneous source text.
        '''

        self.reporter.report_compile_error(CompileError(msg, src_file, span))