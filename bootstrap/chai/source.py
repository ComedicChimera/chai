from dataclasses import dataclass
from typing import List, Dict, MutableSet

from .ast import ASTDef
from .symbol import SymbolTable

# ChaiFile represents a Chai source file.
@dataclass
class ChaiFile:
    rel_path: str
    parent: "ChaiPackage"
    metadata: Dict[str, str]

    # defs is the list of top level AST definitions of this file
    defs: List[ASTDef]

    # visible_packages is a dict of all the packages that are directly by as a
    # named package (eg. `import pkg`) by this file stored by package organized
    # by the name they are defined with
    visible_packages: Dict[str, "ChaiPackage"]

@dataclass
class ChaiPackageImport:
    pkg: "ChaiPackage"
    used_names: MutableSet[str] = set()

# ChaiPackage represents a Chai package.
@dataclass
class ChaiPackage:
    id: int
    name: str
    parent: "ChaiModule"
    files: List[ChaiFile]
    global_table: SymbolTable
    import_table: Dict[int, ChaiPackageImport]

# ChaiModule represents a Chai module.
@dataclass
class ChaiModule:
    id: int
    name: str
    abs_path: str
    root_package: ChaiPackage

    # sub_packages is a dictionary of the sub-packages of the module organized
    # by sub-path.  For example, the package `io.fs.path` would have a sub-path
    # of `fs.path`.
    sub_packages: Dict[str, ChaiPackage]

    # packages returns the full list of packages
    def packages(self) -> List[ChaiPackage]:
        return [self.root_package] + list(self.sub_packages.values())