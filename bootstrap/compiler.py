'''Provides the main compiler driver.'''

__all__ = [
    'BuildOptions',
    'Compiler'
]

from dataclasses import dataclass
import os

from syntax.parser import Parser
from typecheck.walker import Walker
from depm.source import Package, SourceFile
from llvm import Context
from generate import Generator
from llvm.target import Target
from report.reporter import Reporter, LogLevel, CompileError
from linker import *

@dataclass
class BuildOptions:
    '''
    Represents the various build configuration options that can be passed to a
    particular compiler instance.

    Attributes
    ----------
    output_path: str
        The path to write the output binary to
    '''

    output_path: str

class Compiler:
    '''
    The high-level construct representing the Chai compiler.  It is responsible
    for storing and manipulating the Chai compiler's state.
    '''

    # The path to the root package directory.
    root_dir: str

    # The build options.
    build_options: BuildOptions

    # The value of the CHAI path environment variable
    chai_path: str

    # The reporter for the compiler.
    reporter: Reporter
    
    def __init__(self, root_dir: str, build_options: BuildOptions):
        '''
        Params
        ------
        root_dir: str
            The path to the root package directory.
        build_options
            The build options. 
        '''

        self.root_dir = os.path.abspath(root_dir)
        self.build_options = build_options
        self.reporter = Reporter(LogLevel.VERBOSE)
        
        if path := os.environ.get('CHAI_PATH'):
            self.chai_path = path
        else:
            self.reporter.report_error('missing required environment variable `CHAI_PATH`', 'env')
            exit(1)

    def compile(self) -> int:
        '''
        Runs the compiler with the configuration provided in the constructor.

        Returns
        -------
        return_code: int
            The return code for the program.
        
        '''

        # DEBUG Code
        root_dir = os.path.dirname(self.root_dir)
        pkg = Package('test', root_dir)
        src_file = SourceFile(pkg, self.root_dir)
        pkg.files.append(src_file)

        try:
            p = Parser(src_file)
            p.parse()
            
            w = Walker(src_file)
            for defin in src_file.definitions:
                w.walk_definition(defin)

            Target.initialize_all()

            with Context():
                g = Generator(pkg)
                m = g.generate()

                # m.dump()

                target = Target(triple=Target.default_triple())
                machine = target.create_machine()
                m.data_layout = machine.data_layout
                m.target_triple = machine.triple

                obj_output_path = os.path.join(root_dir, 'pkg0.o')
                machine.compile_module(m, obj_output_path)

            create_executable(self.build_options.output_path, [], [obj_output_path])
            os.remove(obj_output_path)
        except CompileError as cerr:
            self.reporter.report_compile_error(cerr)
        except LinkError as lerr:
            self.reporter.report_error('failed to link executable:\n' + lerr.link_output, 'link')
        except LinkConfigError as lcerr:
            err_msg = lcerr.message
            if lcerr.is_msvc_error:
                err_msg += '\nmake sure you have the Microsoft Visual C++ build tools installed'

            self.reporter.report_error(err_msg, 'config')
        except Exception as e:
            self.reporter.report_fatal_error(e)
