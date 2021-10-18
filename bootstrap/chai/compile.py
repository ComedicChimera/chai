import os
from typing import Optional, Dict

from . import CHAI_FILE_EXT, ChaiCompileError, ChaiFail
from .source import ChaiFile, ChaiModule, ChaiPackage
from .mod_loader import load_module, BuildProfile, MOD_FILE_NAME
from .parser import Parser
from .report import report
from .symbol import SymbolTable

class Compiler:
    root_abs_dir: str
    root_mod: ChaiModule
    base_prof: BuildProfile
    dep_graph: Dict[int, ChaiModule] = {}
    chai_path: str

    def __init__(self, root_dir: str):
        # convert the root directory to an absolute path
        self.root_abs_dir = os.path.abspath(root_dir)

        # assert the existence of Chai path
        assert 'CHAI_PATH' in os.environ, 'missing `CHAI_PATH` environment variable'
        self.chai_path = os.environ['CHAI_PATH']    

    # analyze runs the analysis phase of compilation.
    def analyze(self) -> None:
        # load the root module
        self.root_mod, self.base_prof = load_module(self.root_abs_dir, None, "")
        self.dep_graph[self.root_mod.id] = self.root_mod

        # initialize the root package (which also initializes all sub-packages)
        self.init_pkg(self.root_mod, self.root_mod.abs_path)

        # report all unresolved symbols
        for mod in self.dep_graph.values():
            for pkg in mod.packages():
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

    # ---------------------------------------------------------------------------- #

    # init_pkg initializes a package and all its dependencies.
    def init_pkg(self, parent_mod: ChaiModule, pkg_abs_path: str) -> ChaiPackage:
        # create a new package for the given parent module
        pkg_name = os.path.basename(pkg_abs_path)
        pkg_id = hash(pkg_abs_path)
        pkg = ChaiPackage(pkg_id, pkg_name, parent_mod.id, [], SymbolTable(pkg_id))

        # add it to the parent package
        if os.path.samefile(parent_mod.abs_path, pkg_abs_path):
            parent_mod.root_package = pkg
        else:
            sub_path = os.path.relpath(pkg_abs_path, parent_mod.abs_path)
            if '\\' in sub_path:
                sub_path = sub_path.replace('\\', '.')
            else:
                sub_path = sub_path.replace('/', '.')
                
            parent_mod.sub_packages[sub_path] = pkg

        # walk through the files in the package directory
        p = Parser(self.base_prof, pkg.global_table)
        for file in os.listdir(pkg_abs_path):
            _, ext = os.path.splitext(file)
            if not os.path.isdir(file) and ext == CHAI_FILE_EXT:
                # create the Chai file
                file_abs_path = os.path.join(pkg_abs_path, file)
                ch_file = ChaiFile(os.path.relpath(file_abs_path, parent_mod.abs_path), pkg.id, {}, [])

                # catch parse errors so we can continue with analysis
                try:
                    # parse the file and/or determine whether it should be
                    # parsed (based on the metadata)
                    with open(file_abs_path) as fp:
                        if p.parse(ch_file, fp):
                            pkg.files.append(ch_file)

                            # DEBUG
                            print(ch_file.metadata)
                            print(ch_file.defs)
                except ChaiCompileError as cce:
                    report.report_compile_error(cce)
                except ChaiFail:
                    pass

        # test to make sure package isn't empty (but only if there aren't other
        # errors -- we don't want to create more errors than necessary)
        if report.should_proceed() and not pkg.files:
            report.report_fatal_error(f'package `{pkg.name}` contains no compileable source files')

    # import_package imports a package based on a module name and package sub
    # path (eg. a.b.c would have a module name of `a` and a sub path `b.c`).
    # If the package can't be loaded, then none is returned.
    def import_package(self, parent_mod: ChaiModule, mod_name: str, pkg_sub_path: str) -> Optional[ChaiPackage]:
        # get the path to the module
        mod_abs_path = self.locate_module(parent_mod, mod_name)
        if not mod_abs_path:
            return None

        # get the module: fetch it from the dependency graph if it has already
        # been loaded or load it and add it if not
        mod_id = hash(mod_abs_path)
        if mod_id in self.dep_graph:
            mod = self.dep_graph[mod_id]
        else:
            mod, self.base_prof = load_module(mod_abs_path, self.base_prof, "")
            self.dep_graph[mod.id] = mod

        # get the package: retrieve it from the module if it already exists.
        # Otherwise, initialize it manually.
        if pkg_sub_path == "":
            if mod.root_package:
                return mod.root_package
            else:
                return self.init_pkg(mod, mod.abs_path)
        elif pkg_sub_path in mod.sub_packages:
            return mod.sub_packages[pkg_sub_path]
        elif os.path.exists(pkg_abs_path := os.path.join(mod.abs_path, pkg_sub_path)):
            return self.init_pkg(mod, pkg_abs_path)
        else:
            return None

    # locate_module takes a module name and finds its absolute path if it can be
    # found.  It returns "" if not.
    def locate_module(self, parent_mod: ChaiModule, mod_name: str) -> Optional[str]:
        # search the directory the module is contained in
        local_path = os.path.join(os.path.basename(parent_mod.abs_path), mod_name)
        if os.path.exists(os.path.join(local_path, MOD_FILE_NAME)):
            return local_path

        # search the public directory
        pub_path = os.path.join(self.chai_path, 'pub', mod_name)
        if os.path.exists(os.path.join(pub_path, MOD_FILE_NAME)):
            return pub_path

        # search the standard directory
        std_path = os.path.join(self.chai_path, 'std', mod_name)
        if os.path.exists(os.path.join(std_path, MOD_FILE_NAME)):
            return std_path

        # no module path located
        return ""

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