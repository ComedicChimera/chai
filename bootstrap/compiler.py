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
        
        if path := os.environ.get('CHAI_PATH'):
            self.chai_path = path
        else:
            print('missing environment variable `CHAI_PATH`')
            exit(1)

    def compile(self):
        '''Runs the compiler with the configuration provided in the constructor.'''

        # DEBUG Code
        root_dir = os.path.dirname(self.root_dir)
        pkg = Package('test', root_dir)
        srcfile = SourceFile(pkg, self.root_dir)
        pkg.files.append(srcfile)

        p = Parser(srcfile)
        p.parse()
        
        w = Walker(srcfile)
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

    def create_executable(self, link_flags: List[str], link_objects: List[str]):
        link_command = []

        match sys.platform:
            case 'win32':
                vswhere_path = os.path.join(self.chai_path, 'tools/bin/vswhere.exe')
                output, ok = exec_command([vswhere_path, '-latest', '-products', '*', '-property', 'installationPath'])

                if not ok:
                    print(output)
                    exit(1)

                vs_dev_cmd_path = os.path.join(output[:-2], 'VC/Auxiliary/Build/vcvars64.bat')

                link_command = [vs_dev_cmd_path, '&', 'link.exe', '/entry:_start', '/subsystem:console', '/nologo']
                link_objects = ['kernel32.lib', 'libcmt.lib'] + link_objects
            case _:
                print('unsupported platform')
                exit(1)

        link_command.extend(link_flags)
        link_command.extend(link_objects)

        output, ok = exec_command(link_command)
        if not ok:
            print('Link Error:')
            print(output)
            exit(1)

def exec_command(command: List[str]) -> Tuple[str, bool]:
    proc = subprocess.run(command, stdout=subprocess.PIPE)
    return proc.stdout.decode(), proc.returncode == 0
