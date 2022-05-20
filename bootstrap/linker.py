'''Provides all the functionality needed to link binraries.'''

__all__ = [
    'LinkError',
    'LinkConfigError',
    'create_executable'
]

from dataclasses import dataclass
from typing import List, Tuple
import subprocess
import os
import sys
import winreg

@dataclass
class LinkError(Exception):
    '''
    Represents an error that occurred while linking an executable.

    Attributes
    ----------
    link_output: str
        The output from the linker to display.
    '''

    link_output: str

@dataclass
class LinkConfigError(Exception):
    '''
    Represents an error in the environment configuration of the linker.

    Attributes
    ----------
    message: str
        The error message to display to the user.
    is_msvc_error: bool
        Whether to error occurred while trying to locate an MSVC tool. If so, we
        want to display a message prompting the user to make sure they have the
        VC++ dev tools installed.
    '''

    message: str
    is_msvc_error: bool = False

def create_executable(output_path: str, link_flags: List[str], link_objects: List[str]):
    '''
    Links a collection of objects into an executable.

    Params
    ------
    output_path: str
        The path to output the executable to.
    link_flags: List[str]
        The list of additional flags to pass to the linker.
    link_objects: List[str]
        The list of paths to object files, libraries, etc. to give as input to
        the linker.     
    '''

    # Determine the linker information and default configuration based on the
    # platform.
    match sys.platform:
        case 'win32':
            linker_path = get_windows_linker_path()
            default_link_flags, default_link_libs = get_windows_default_link_args(output_path)
        case _:
            raise LinkConfigError('unsupported platform')

    # Run the linker based on the determined linker information and default
    # configuration.
    output, ok = exec_command([
        linker_path, 
        *default_link_flags, 
        *link_flags, 
        *default_link_libs, 
        *link_objects
    ])

    # Handle any errors that occur while linking.
    if not ok:
        raise LinkError(output)

# ---------------------------------------------------------------------------- #

def get_windows_linker_path() -> str:
    '''Calculates ands returns the path to the Windows native linker.'''

    # Calculate the expected path to `vswhere.exe`.
    vswhere_path = os.path.join(
        os.environ.get('ProgramFiles(x86)'),
        'Microsoft Visual Studio/Installer/vswhere.exe'
    )

    # Validate that `vswhere.exe` exists.
    if not os.path.exists(vswhere_path):
        raise LinkConfigError('cannot find `vswhere.exe`', True)

    # Call `vswhere.exe` with the appropriate arguments to find the location of
    # the installed MSVC build tools.
    output, ok = exec_command([
        vswhere_path, 
        '-latest', 
        '-products', '*', 
        '-requires',
        'Microsoft.VisualStudio.Component.VC.Tools.x86.x64',
        '-property', 
        'installationPath'
    ])

    if not ok:
        raise LinkConfigError('cannot find Microsoft Visual C++ build tools:\n' + output, True)

    # Calculate the path to the MSVC tools from the output of `vswhere.exe`
    vc_tools_path = output.strip()

    # Calculate the path to the MSVC tools version file and make sure it exists.
    version_file_path = os.path.join(vc_tools_path, 'VC/Auxiliary/Build/Microsoft.VCToolsVersion.default.txt')
    if not os.path.exists(version_file_path):
        raise LinkConfigError('missing Microsoft Visual C++ version file', True)

    # Read the version from the version file.
    with open(version_file_path) as f:
        version = f.readline().strip()

    # Use the MSVC tools path and version to calculate the path to `link.exe`
    # and validate that it exists at the expected location.
    link_path = os.path.join(vc_tools_path, 'VC/Tools/MSVC', version, 'bin/Hostx64/x64/link.exe')
    if not os.path.exists(link_path):
        raise LinkConfigError('cannot find `link.exe`', True)

    return link_path

def get_windows_default_link_args(output_path: str) -> Tuple[List[str], List[str]]:
    '''
    Calculates the default link flags and libraries to pass to the Windows
    native linker.

    Params
    ------
    output_path: str
        The path to output the resulting executable to.

    Returns
    -------
    (link_flags: List[str], link_libs: List[str])
        link_flags: The default link flags.
        link_libs: The default link libraries.
    '''

    # Get the path to the Windows SDK which we will pass as the library path
    # argument to the linker.
    win_sdk_lib_path = get_windows_sdk_lib_path()

    # Get the path to the Windows libraries.
    win_lib_path = os.path.join(win_sdk_lib_path, 'um/x64')
    
    # Get the path to the UCRT libraries.
    ucrt_lib_path = os.path.join(win_sdk_lib_path, 'ucrt/x64')

    # Define the default link flags.
    link_flags = [
        '/entry:_start',             # Set the entry point.
        '/subsystem:console',        # Set the executable to be a console app.
        '/nologo',                   # Turn of the logo banner for error reporting.
        '/libpath:' + win_lib_path,  # Add the Windows Libraries to the libpath.
        '/libpath:' + ucrt_lib_path, # Add the UCRT libraries to the libpath.
        '/out:' + output_path,       # Set the executable output path.
    ]

    # Add the needed Windows import libraries.
    link_libs = ['kernel32.lib', 'ucrt.lib']

    # Return the calculated flags and libaries.
    return link_flags, link_libs

def get_windows_sdk_lib_path() -> str:
    '''
    Calculates the path to the Windows SDK libraries.  These libraries are
    needed to access import libraries like `kernel32.lib` which provide
    essential system calls like `ExitProcess`.
    '''

    try:
        # Open the registry key corresponding the Windows SDK information. The
        # full path to this key is as follows:
        #
        # HKEY_LOCAL_MACHINE
        #   > SOFTWARE
        #       > Wow6432Node
        #           > Microsoft
        #               > Microsoft SDKSs
        #                   > Windows
        #                       > v10.0
        #
        # We use context managers to ensure each key gets closed once it is
        # opened even in the case of an exception.
        with winreg.OpenKey(winreg.HKEY_LOCAL_MACHINE, 'SOFTWARE') as rk_software,\
            winreg.OpenKey(rk_software, 'WOW6432Node') as rk_wow6432node,\
            winreg.OpenKey(rk_wow6432node, 'Microsoft') as rk_microsoft,\
            winreg.OpenKey(rk_microsoft, 'Microsoft SDKs') as rk_microsoft_sdks,\
            winreg.OpenKey(rk_microsoft_sdks, 'Windows') as rk_windows,\
            winreg.OpenKey(rk_windows, 'v10.0') as rk_v10_0:
        
            # Read the values of `InstallationFolder` and `ProductVersion` from
            # the keys which should both be strings (type of `KEY_CZ`).
            install_folder, _ = winreg.QueryValueEx(rk_v10_0, 'InstallationFolder')
            version, _ = winreg.QueryValueEx(rk_v10_0, 'ProductVersion')

            # Use the extracted registry values to make the SDK library path.
            return os.path.join(install_folder, 'Lib', version + '.0')
    except OSError as oe:
        # OSErrors always imply that we couldn't access the registry key and its
        # values which we just report as a generic Windows SDK error.
        raise LinkConfigError('failed to locate Windows SDK:' + str(oe))

# ---------------------------------------------------------------------------- #

def exec_command(cmd: List[str]) -> Tuple[str, bool]:
    '''
    Executes the given system command.

    Params
    ------
    cmd: List[str]
        The list of values that form the command to run.
    
    Returns
    -------
    (stdout: str, ok: bool)
        stdout: the captured standard output of the executed command
        ok: whether or not the command executed successfully
    '''
    
    proc = subprocess.run(cmd, stdout=subprocess.PIPE)
    return proc.stdout.decode(), proc.returncode == 0