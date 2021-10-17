import os
from typing import Optional

from . import CHAI_FILE_EXT, ChaiCompileError, ChaiFail
from .source import ChaiFile, ChaiModule, ChaiPackage
from .mod_loader import load_module, BuildProfile
from .parser import Parser
from .report import report
from .symbol import SymbolTable

class Compiler:
    root_abs_dir: str
    root_mod: ChaiModule
    base_prof: BuildProfile

    def __init__(self, root_dir: str):
        # convert the root directory to an absolute path
        self.root_abs_dir = os.path.abspath(root_dir)

    # analyze runs the analysis phase of compilation.
    def analyze(self) -> None:
        # load the root module
        self.root_mod, self.base_prof = load_module(self.root_abs_dir, None, "")

        # initialize the root package (which also initializes all sub-packages)
        self.init_pkg(self.root_mod, self.root_mod.abs_path)

        # TODO: report unresolved symbols (update to use depgraph)
        for pkg in self.root_mod.packages():
            pkg.global_table.report_unresolved()

        # TODO: check for recursive symbol usages that require references (eg. a
        # struct being used within itself)

        # NOTE: we continue with semantic analysis even if some files failed to
        # parse since if they did, they were never added as valid files so we
        # can still process what remains.

        # TODO: rest of semantic analysis

    # generate produces the target output for the current project.  It returns
    # the output location if it succeeds.
    def generate(self) -> Optional[str]:
        # TODO
        return self.base_prof.output_path

    # init_pkg initializes a package and all its dependencies.
    def init_pkg(self, parent_mod: ChaiModule, pkg_abs_path: str) -> ChaiPackage:
        # create a new package for the given parent module
        pkg_name = os.path.basename(pkg_abs_path)
        pkg_id = hash(pkg_name)
        pkg = ChaiPackage(pkg_id, pkg_name, parent_mod.id, [], SymbolTable(pkg_id))

        # add it to the parent package
        if os.path.samefile(parent_mod.abs_path, pkg_abs_path):
            parent_mod.root_package = pkg
        else:
            parent_mod.sub_packages[os.path.relpath(pkg_abs_path, parent_mod.abs_path)]

        # walk through the files in the package directory
        p = Parser(self.base_prof, pkg.global_table)
        for file in os.listdir(pkg_abs_path):
            _, ext = os.path.splitext(file)
            if not os.path.isdir(file) and ext == CHAI_FILE_EXT:
                # create the Chai file
                file_abs_path = os.path.join(pkg_abs_path, file)
                ch_file = ChaiFile(os.path.relpath(file_abs_path, parent_mod.abs_path), pkg.id, [])

                # catch parse errors so we can continue with analysis
                try:
                    # DEBUG: parse the file
                    with open(file_abs_path) as fp:
                        p.parse(ch_file, fp)
                        print(ch_file.defs)

                    # TODO: parse it and determine if it should be added
                except ChaiCompileError as cce:
                    report.report_compile_error(cce)
                except ChaiFail:
                    pass

# compile_module compiles a module and all of its sub-dependencies.  This is the
# main entry for compilation.  It returns the output location if it succeeds.
def compile_module(root_dir) -> Optional[str]:
    try:
        c = Compiler(root_dir)
        c.analyze()

        if report.should_proceed():
            return c.generate()
    except ChaiFail:
        pass