from dataclasses import dataclass
import os
import sys
import subprocess
from typing import List, Tuple

from syntax.parser import Parser
from typecheck.walker import Walker
from depm.source import Package, SourceFile
from llvm import Context
from generate import Generator
from llvm.target import Target
from report.reporter import Reporter, LogLevel, CompileError

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
            w.walk_file()

            Target.initialize_all()

            with Context():
                g = Generator(pkg)
                m = g.generate()

                # m.dump()

                target = Target(triple=Target.default_triple())
                machine = target.create_machine()
                m.data_layout = machine.data_layout
                m.target_triple = machine.triple

                output_path = os.path.join(root_dir, 'pkg0.o')
                machine.compile_module(m, output_path)

            self.create_executable(['/out:' + self.build_options.output_path], [output_path])
            os.remove(output_path)
        except CompileError as cerr:
            self.reporter.report_compile_error(cerr)
        except Exception as e:
            self.reporter.report_fatal_error(e)

    def create_executable(self, link_flags: List[str], link_objects: List[str]):
        link_command = []

        match sys.platform:
            case 'win32':
                vswhere_path = os.path.join(self.chai_path, 'tools/bin/vswhere.exe')
                output, ok = exec_command([vswhere_path, '-latest', '-products', '*', '-property', 'installationPath'])

                if not ok:
                    self.reporter.report_error('failed to find linker:\n' + output, 'env')
                    return

                vs_dev_cmd_path = os.path.join(output[:-2], 'VC/Auxiliary/Build/vcvars64.bat')

                link_command = [vs_dev_cmd_path, '&', 'link.exe', '/entry:_start', '/subsystem:console', '/nologo']
                link_objects = ['kernel32.lib', 'libcmt.lib'] + link_objects
            case _:
                self.reporter.report_error('unsupported platform', 'env')
                return

        link_command.extend(link_flags)
        link_command.extend(link_objects)

        output, ok = exec_command(link_command)
        if not ok:
            self.reporter.report_error('failed to link program:\n' + output, 'link')
            return

def exec_command(command: List[str]) -> Tuple[str, bool]:
    proc = subprocess.run(command, stdout=subprocess.PIPE)
    return proc.stdout.decode(), proc.returncode == 0
