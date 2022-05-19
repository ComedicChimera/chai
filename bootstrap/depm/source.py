'''Provides the relevant definitions for Chai's package system.'''

from dataclasses import dataclass, field
from typing import Dict, List
import os

from . import Symbol
from syntax.ast import ASTNode

@dataclass
class SourceFile:
    '''
    Represents a Chai source file.

    Attributes
    ----------
    parent: Package
        The parent package to this file.
    abs_path: str
        The absolute path to this file.
    definitions: List[ASTNode]
        The AST definitions that comprise this file.
    '''

    parent: 'Package'
    abs_path: str
    definitions: List[ASTNode] = field(default_factory=list)

    @property
    def rel_path(self):
        return os.path.relpath(self.parent.abs_path, self.abs_path)

@dataclass
class Package:
    '''
    Represents a Chai package: a collection of source files in the same
    directory sharing a common namespace and position in the import resolution
    hierachy.

    Attributes
    ----------
    name: str
        The name of the package.
    abs_path: str
        The absolute path to the package directory.
    id: int
        The unique ID of the package generated from its absolute path.
    root_pkg: Package
        The root package as defined in the package's package path, used to
        determine the package's position within the import resolution hierachy.
    files: List[SourceFile]
        The list of Chai files that make up this package.
    symbol_table: Dict[str, Symbol]
        The table mapping symbol names to symbol objects defined in the
        package's global, shared namespace.
    '''

    name: str
    abs_path: str

    id: int = field(init=False)
    pkg_path: str = ""
    root_pkg: 'Package' = field(init=False)
    files: List[SourceFile] = field(default_factory=list)

    symbol_table: Dict[str, Symbol] = field(default_factory=dict)

    def __post_init__(self):
        '''Calculates the package ID based on its absolute path.'''

        self.id = abs(hash(self.abs_path))

    def __eq__(self, other: object) -> bool:
        '''
        Returns whether this package is equal to other.
        
        Params
        ------
        other: object
            The object to compare this package to.
        '''

        if isinstance(other, Package):
            return self.id == other.id

        return False

    @property
    def display_path(self) -> str:
        '''
        Returns the path that should be used to identify the package. Most
        often, this will be the package path, but if that cannot be determined,
        then the name is given instead.
        '''

        if self.pkg_path:
            return self.pkg_path
        else:
            return self.name

    