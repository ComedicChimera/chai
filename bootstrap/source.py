from dataclasses import dataclass

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
    root_pkg: Package
        The root package as defined in the package's package path, used to
        determine the package's position within the import resolution hierachy.
    pkg_id: int
        The unique ID of the package generated from its absolute path.
    '''

    pkg_name: str
    abs_path: str
    root_pkg: 'Package'

    pkg_id: int = 0

    def __post_init__(self):
        # Calculate the package ID based on its absolute path.
        self.pkg_id = hash(self.abs_path)

    