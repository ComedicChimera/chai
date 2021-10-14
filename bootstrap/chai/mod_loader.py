from dataclasses import dataclass
from typing import List, Optional, Tuple, Dict, Any
from enum import Enum, auto
import os

import toml

from . import CHAI_VERSION, ChaiModuleError
from .source import ChaiModule
from .report import report

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

# ModuleLoader is the class responsible for loading a module
class ModuleLoader:
    mod_abs_path: str
    mod_name: str
    base_profile: Optional[BuildProfile]

    # if this field is "", there is no selected profile by name
    selected_profile_name: str

    def __init__(self, mod_abs_path: str, base_profile: Optional[BuildProfile], selected_profile_name: str) -> None:
        self.mod_abs_path = mod_abs_path
        self.base_profile = base_profile
        self.selected_profile_name = selected_profile_name

    # load is the main entry point for the module loader
    def load(self) -> Tuple[ChaiModule, BuildProfile]:
        # create the path to the module file
        mod_file_abs_path = os.path.join(self.mod_abs_path, MOD_FILE_NAME)

        # check if the module file exists
        if not os.path.exists(mod_file_abs_path):
            self._error('`chai-mod.toml` does not exist in module directory')

        # load the module file with TOML
        with open(mod_file_abs_path, 'r') as toml_file:
            try:
                toml_data = toml.load(toml_file)
            except toml.TomlDecodeError as te:
                self._error(f'error parsing module file: {te.args[0]}')

        # get the module name
        self.mod_name = self._get_field(toml_data, 'name', str)

        # check Chai version compatibility
        mod_chai_version = self._get_field(toml_data, 'chai-version', str)

        # version is older than current version => warn, but still compile
        if mod_chai_version < CHAI_VERSION:
            report.report_module_warning(self.mod_name, 'current Chai version ({CHAI_VERSION}) is ahead of version specified by module ({mod_chai_version})')
        # version is newer than current version => fail
        elif mod_chai_version > CHAI_VERSION:
            self._error(f'current Chai version ({CHAI_VERSION}) is behind version specified by module ({mod_chai_version})')

        # TODO: load the profile

        # return the completed module
        return ChaiModule(hash(self.mod_abs_path), self.mod_name, self.mod_abs_path, None, {}), None

    # _error reports an error loading the module
    def _error(self, msg: str) -> None:
        if self.mod_name == "":
            raise ChaiModuleError(f'<module at {self.mod_abs_path}', msg)
        else:
            raise ChaiModuleError(self.mod_name, msg)

    # _get_field gets a field from the module.  It throws an error if no default
    # is provided: -- indicating that the field is optional.
    def _get_field(self, toml_data: Dict[str, Any], field_name: str, field_type, default = None) -> Any:
        pass

# load_module loads a module from a TOML file.  The profile argument should be a
# reference the global build profile which will be updated when this module is
# loaded.  If the profile passed is `None`, then a new profile will be
# generated. A module and the updated/new profile are returned.
def load_module(mod_abs_path: str, base_profile: Optional[BuildProfile], profile_name: str) -> Tuple[ChaiModule, BuildProfile]:
    ml = ModuleLoader(mod_abs_path, base_profile, profile_name)
    return ml.load()