from dataclasses import dataclass
from typing import List, Dict

from .ast import ASTDef

# ChaiFile represents a Chai source file.
@dataclass
class ChaiFile:
    rel_path: str
    parent_id: int

    # defs is the list of top level AST definitions of this file
    defs: List[ASTDef]

# ChaiPackage represents a Chai package.
@dataclass
class ChaiPackage:
    id: int
    name: str
    parent_id: int
    files: List[ChaiFile]

    # TODO: rest

# ChaiModule represents a Chai module.
@dataclass
class ChaiModule:
    id: int
    name: str
    abs_path: str
    root_package: ChaiPackage

    # sub_packages is a dictionary of the sub-packages of the module organized
    # by sub-path.  For example, the package `io.fs.path` would have a sub-path
    # of `fs/path`.
    sub_packages: Dict[str, ChaiPackage]