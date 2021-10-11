from dataclasses import dataclass
from typing import List, Optional, Tuple
from enum import Enum, auto
import os

import toml

from . import CHAI_VERSION, ChaiModuleError
from .source import ChaiModule

# The expected name of the module file.
MOD_FILE_NAME = "chai-mod.toml"

# Supported configurations.
SUPPORTED_OS = ['windows']
SUPPORTED_ARCH = ['x64']

# Supported output formats.
class TargetFormat(Enum):
    BIN = auto()
    LLVM = auto()
    ASM = auto()
    OBJ = auto()
    LIB = auto()

# BuildProfile is the global build configuration for the current compilation.
@dataclass
class BuildProfile:
    debug: bool
    target_os: str
    target_arch: str
    target_format: TargetFormat

    # link_objects is the list of paths to additional object files and static
    # libraries to link with the final program.
    link_objects: List[str]

# load_module loads a module from a TOML file.  The profile argument should be a
# reference the global build profile which will be updated when this module is
# loaded.  If the profile passed is `None`, then a new profile will be
# generated. A module and the updated/new profile are returned.
def load_module(mod_abs_path: str, profile: Optional[BuildProfile]) -> Tuple[ChaiModule, BuildProfile]:
    # create the path to the module file
    mod_file_abs_path = os.path.join(mod_abs_path, MOD_FILE_NAME)

    # check if the module file exists
    if not os.path.exists(mod_file_abs_path):
        raise ChaiModuleError(f'<module at {mod_abs_path}>', '`chai-mod.toml` does not exist in module directory')

    # load the module file with TOML
    with open(mod_file_abs_path, 'r') as toml_file:
        try:
            toml_data = toml.load(toml_file)
        except toml.TomlDecodeError as te:
            raise ChaiModuleError(f'<module at {mod_abs_path}>', f'error parsing module file: {te.args[0]}')

    # get the module name
    if 'name' in toml_data:
        mod_name = toml_data['name']
    else:
        raise ChaiModuleError(f'<module at {mod_abs_path}>', 'module missing required field: `name`')

    # read in the rest of the top level data
    try:
        # check Chai version compatibility
        mod_chai_version = toml_data['chai-version']

        # version is older than current version => warn, but still compile
        if mod_chai_version < CHAI_VERSION:
            print(f'warning in module {mod_name}: current Chai version ({CHAI_VERSION}) is ahead of version specified by module ({mod_chai_version})')
        # version is newer than current version => fail
        elif mod_chai_version > CHAI_VERSION:
            raise ChaiModuleError(mod_name, f'current Chai version ({CHAI_VERSION}) is behind version specified by module ({mod_chai_version})')

        # TODO: other top level options (necessary for Alpha Chai)

        # TODO: build profile handling

        return ChaiModule(hash(mod_abs_path), mod_name, mod_abs_path, None, {}), None
    except KeyError as e:
        raise ChaiModuleError(mod_name, f'module missing required field: `{e.args[0]}`')