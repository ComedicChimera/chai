'''Provides the relevant definitions for Chai's package system.'''

from dataclasses import dataclass, field
from typing import Dict, List, Optional
import os

from . import Symbol, Operator
from syntax.token import Token
from syntax.ast import ASTNode

@dataclass
class SourceFile:
    '''
    Represents a Chai source file.

    Attributes
    ----------
    parent: Package
        The parent package to this file.
    file_number: int
        The number uniquely identifying this file within its own package.
    abs_path: str
        The absolute path to this file.
    definitions: List[ASTNode]
        The AST definitions that comprise this file.
    rel_path: str
        The package-relative path to the file (usually the file name).
    '''

    parent: 'Package'
    file_number: int
    abs_path: str
    definitions: List[ASTNode] = field(default_factory=list)

    @property
    def rel_path(self):
        return os.path.relpath(self.parent.abs_path, self.abs_path)

    def lookup_symbol(self, name: str) -> Optional[Symbol]:
        '''
        Looks up a symbol by name defined in the top scope of the file or in the
        global scope of the package.

        Params
        ------
        name: str
            The name of the symbol to look up.
        '''

        # TODO imported symbols

        if sym := self.parent.symbol_table.get(name):
            return sym

        return None

    def lookup_operator(self, op_kind: Token.Kind, arity: int) -> Optional[Operator]:
        '''
        Looks up an operator definition by kind and arity defined in the top scope
        of the file or global scope of the parent package.

        Params
        ------
        op_kind: Token.Kind
            The operator kind.
        arity: int
            The operator arity.
        '''

        # TODO imported operators

        if ops := self.parent.operator_table.get(op_kind):
            for op in ops:
                if op.arity == arity:
                    return op

        return None

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
    operator_table: Dict[Token.Kind, Operator]
        The table mapping operator kinds to the list of operators defined for
        that operator kind in the given package.
    '''

    name: str
    abs_path: str

    id: int = field(init=False)
    pkg_path: str = ""
    root_pkg: 'Package' = field(init=False)
    files: List[SourceFile] = field(default_factory=list)

    symbol_table: Dict[str, Symbol] = field(default_factory=dict)
    operator_table: Dict[Token.Kind, List[Operator]] = field(default_factory=dict)

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

    