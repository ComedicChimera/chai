from dataclasses import dataclass
import os

from syntax.parser import Parser
from typecheck.walker import Walker
from depm.source import Package, SourceFile
from llvm.context import Context
from llvm.module import Module
import llvm.types as lltypes

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

    def compile(self):
        '''Runs the compiler with the configuration provided in the constructor.'''

        # DEBUG Code
        pkg = Package('test', os.path.dirname(self.root_dir))
        srcfile = SourceFile(pkg, self.root_dir)
        pkg.files.append(srcfile)

        p = Parser(srcfile)
        p.parse()
        
        w = Walker(srcfile)
        w.walk_file()

        with Context() as ctx:
            m = Module('test', ctx)
            
            func = m.add_function('add', lltypes.FunctionType(
                [lltypes.PointerType(lltypes.Int32Type), lltypes.Int32Type],
                lltypes.Int32Type,
            ))

            first_param = func.params[0]
            first_param.name = "a"

            second_param = func.params[1]
            second_param.name = "b"

            m.dump()
