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
SUPPORTED_ARCH = ['amd64']

# Supported output formats.
class TargetFormat(Enum):
    BIN = auto()
    LLVM = auto()
    ASM = auto()
    OBJ = auto()
    LIB = auto()

SUPPORTED_FORMATS = {
    'bin': TargetFormat.BIN,
    'llvm': TargetFormat.LLVM,
    'asm': TargetFormat.ASM,
    'obj': TargetFormat.OBJ,
    # NOTE: this compiler will not support `lib` or `mir`
}

# BuildProfile is the global build configuration for the current compilation.
@dataclass
class BuildProfile:
    debug: bool
    target_os: str
    target_arch: str
    target_format: TargetFormat

    output_path: str

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
        self.mod_name = ""
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

        # load the appropriate profile/update the base profile
        allow_prof_el = self._get_field(toml_data, 'allow-profile-elision', bool, False)
        profiles = self._get_field(toml_data, 'profiles', list, [])
        prof = self._update_profile(profiles, allow_prof_el)

        # return the completed module
        return ChaiModule(hash(self.mod_abs_path), self.mod_name, self.mod_abs_path, None, {}), prof

    # ---------------------------------------------------------------------------- #

    # _update_profile updates (or determined if it doesn't already exist) the
    # base profile of a project based on the profiles provided in a given
    # module. 
    def _update_profile(self, profiles: List[Dict[str, Any]], allow_prof_el: bool) -> BuildProfile:
        if not all(isinstance(x, dict) for x in profiles):
            self._error("field `profiles` must be an array of tables")

        def get_profile_field(prof_name: str, prof: Dict[str, Any], name: str, typ, default = None):
            if name in prof:
                field_val = prof[name]
                if isinstance(field_val, typ):
                    return field_val
                else:
                    self._error(f'field `{name}` of profile `{prof_name}` must be a {typ}')
            elif default != None:
                return default
            else:
                self._error(f'profile {prof_name} missing required field: `{name}`')

        def select_profile(target_os: str, target_arch: str, debug: bool) -> Optional[Dict[str, Any]]:
            matching_profiles = []

            # isolate matches using target characteristics
            for i, prof in enumerate(profiles):
                prof_name = get_profile_field(f'#{i}', prof, 'name', str)

                prof_target_os = get_profile_field(prof_name, prof, 'target-os', str)
                prof_target_arch = get_profile_field(prof_name, prof, 'target-arch', str)
                prof_debug = get_profile_field(prof_name, prof, 'debug', bool)

                if prof_target_os != target_os or prof_target_arch != target_arch or prof_debug != debug:
                    continue

                # handle base only profiles
                if self.base_profile and get_profile_field(prof_name, prof, 'base-only', bool, False):
                    continue

                matching_profiles.append(prof)

            # no matches => None
            if not matching_profiles:
                return None
            # one match => just return
            elif len(matching_profiles) == 1:
                return matching_profiles[0]
            # multiple matches => return default or first if no default exists
            else:
                for match in matching_profiles:
                    if get_profile_field(match['name'], match, 'default', bool, False):
                        return match

                return matching_profiles[0]

        # base profile already exists
        if self.base_profile:
            match = select_profile(self.base_profile.target_os, self.base_profile.target_arch, self.base_profile.debug)

            # handle no match
            if not match:
                if allow_prof_el:
                    return self.base_profile
                else:
                    self._error('no matching profiles')

            # update base profile
            self.base_profile.link_objects += get_profile_field(match['name'], match, 'link-objects', list, [])

            # return updated base profile
            return self.base_profile
        elif self.selected_profile_name:
            # only characteristic is profile name
            for i, prof in enumerate(profiles):
                prof_name = get_profile_field(f'#{i}', prof, 'name', str)
                if prof_name == self.selected_profile_name:
                    match = prof
                    break
            else:
                self._error(f'no profile found matching name {self.selected_profile_name}')
            
            # get the target output format
            target_format = get_profile_field(match['name'], match, 'format', str)
            if target_format in SUPPORTED_FORMATS:
                target_format = SUPPORTED_FORMATS[target_format]
            else:
                self._error(f'unsupported target format: {target_format}')

            # build the profile object assuming that none of the relevant fields
            # have already been checked
            return BuildProfile(
                debug=get_profile_field(match['name'], match, 'debug', bool),
                target_os=get_profile_field(match['name'], match, 'target-os', str),
                target_arch=get_profile_field(match['name'], match, 'target-arch', str),
                target_format=target_format,
                link_objects=get_profile_field(match['name'], match, 'link-objects', list, []),
                output_path=get_profile_field(match['name'], match, 'output-path', str)
                )
        else:
            # no extra info provided => use system config + debug enabled
            target_os, target_arch = self._get_config()
            match = select_profile(target_os, target_arch, True)

            if not match:
                self._error('unable to determine base profile')

            # get the target output format
            target_format = get_profile_field(match['name'], match, 'format', str)
            if target_format in SUPPORTED_FORMATS:
                target_format = SUPPORTED_FORMATS[target_format]
            else:
                self._error(f'unsupported target format: {target_format}')

            # construct build profile object assuming that all relevant fields
            # exist (except `link-objects`)
            return BuildProfile(
                debug=match['debug'],
                target_os=match['target-os'],
                target_arch=match['target-arch'],
                target_format=target_format,
                link_objects=get_profile_field(match['name'], match, 'link-objects', list, []),
                output_path=get_profile_field(match['name'], match, 'output-path', str)
            )

    # _get_config gets the current system configuration as Chai strings
    def _get_config(self) -> Tuple[str, str]:
        # NOTE: this bootstrapped compiler is only ever intented to work on
        # Windows (at least at the moment), so it just defaults to returning the
        # Windows target information:
        return ('windows', 'amd64')

    # ---------------------------------------------------------------------------- #

    # _error reports an error loading the module
    def _error(self, msg: str) -> None:
        if self.mod_name == "":
            raise ChaiModuleError(f'<module at {self.mod_abs_path}>', msg)
        else:
            raise ChaiModuleError(self.mod_name, msg)

    # _get_field gets a field from the module.  It throws an error if no default
    # is provided: -- indicating that the field is optional.
    def _get_field(self, toml_data: Dict[str, Any], field_name: str, field_type, default = None) -> Any:
        if field_name in toml_data:
            field_val = toml_data[field_name]

            if isinstance(field_val, field_type):
                return field_val
            else:
                self._error(f'field `{field_name}`` must be of type {field_type}')

        if default != None:
            return default
        else:
            self._error(f'missing required field: `{field_name}`')

# load_module loads a module from a TOML file.  The profile argument should be a
# reference the global build profile which will be updated when this module is
# loaded.  If the profile passed is `None`, then a new profile will be
# generated. A module and the updated/new profile are returned.
def load_module(mod_abs_path: str, base_profile: Optional[BuildProfile], profile_name: str) -> Tuple[ChaiModule, BuildProfile]:
    ml = ModuleLoader(mod_abs_path, base_profile, profile_name)
    return ml.load()