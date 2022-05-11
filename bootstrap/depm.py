from dataclasses import dataclass, field
from typing import List

@dataclass
class SourceFile:
    '''
    Represents a Chai source file.

    Attributes
    ----------
    parent: Package
        The parent package to this file.
    rel_path: str
        The package-relative path to this file.
    '''

    parent: 'Package'
    rel_path: str

@dataclass
class Package:
    '''
    Represents a Chai package: a collection of source files in the same
    directory sharing a common namespace and position in the import resolution
    hierachy.

    Attributes
    ----------
    pkg_name: str
        The name of the package.
    abs_path: str
        The absolute path to the package directory.
    pkg_id: int
        The unique ID of the package generated from its absolute path.
    root_pkg: Package
        The root package as defined in the package's package path, used to
        determine the package's position within the import resolution hierachy.
    files: List[SourceFile]
        The list of Chai files that make up this package.
    '''

    pkg_name: str
    abs_path: str

    pkg_id: int = field(init=False)
    root_pkg: 'Package' = field(init=False)
    files: List[SourceFile] = field(default_factory=list)

    def __post_init__(self):
        '''Calculates the package ID based on its absolute path.'''

        self.pkg_id = hash(self.abs_path)

    def __eq__(self, other: object) -> bool:
        '''
        Returns whether this package is equal to other.
        
        Params
        ------
        other: object
            The object to compare this package to.
        '''

        if isinstance(other, Package):
            return self.pkg_id == other.pkg_id

        return False


    