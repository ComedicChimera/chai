import os

from . import CHAI_FILE_EXT
from .source import ChaiFile, ChaiModule, ChaiPackage
from .mod_loader import load_module, BuildProfile
from .lexer import Lexer, TokenKind

class Compiler:
    root_abs_dir: str
    root_mod: ChaiModule
    global_prof: BuildProfile

    def __init__(self, root_dir):
        # convert the root directory to an absolute path
        self.root_abs_dir = os.path.abspath(root_dir)

    # analyze runs the analysis phase of compilation
    def analyze(self) -> None:
        # load the root module
        self.root_mod, self.global_prof = load_module(self.root_abs_dir, None)

        # initialize the root package (which also initializes all sub-packages)
        self.init_pkg(self.root_mod, self.root_mod.abs_path)

        # TODO: semantic analysis

    # generate produces the target output for the current project
    def generate(self) -> None:
        # TODO
        pass

    # init_pkg initializes a package and all its dependencies
    def init_pkg(self, parent_mod: ChaiModule, pkg_abs_path: str) -> ChaiPackage:
        # create a new package for the given parent module
        pkg_name = os.path.basename(pkg_abs_path)
        pkg = ChaiPackage(hash(pkg_name), pkg_name, parent_mod.id, [])

        # add it to the parent package
        if os.path.samefile(parent_mod.abs_path, pkg_abs_path):
            parent_mod.root_package = pkg
        else:
            parent_mod.sub_packages[os.path.relpath(pkg_abs_path, parent_mod.abs_path)]

        # walk through the files in the package directory
        for file in os.listdir(pkg_abs_path):
            _, ext = os.path.splitext(file)
            if not os.path.isdir(file) and ext == CHAI_FILE_EXT:
                # create the Chai file
                file_abs_path = os.path.join(pkg_abs_path, file)
                ch_file = ChaiFile(os.path.relpath(file_abs_path, parent_mod.abs_path), pkg.id)

                print(file_abs_path, ch_file)

                # DEBUG: run the lexer on the file
                with open(file_abs_path) as fp:
                    l = Lexer(ch_file, fp)
                    
                    while (tok := l.next_token()).kind != TokenKind.EndOfFile:
                        print(tok)

                # TODO: parse it and determine if it should be added



# compile_module compiles a module and all of its sub-dependencies.  This is the
# main entry for compilation.
def compile_module(root_dir):
    c = Compiler(root_dir)
    c.analyze()
    c.generate()