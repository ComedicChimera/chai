'''The entry point for the Chai compiler.'''

__all__ = ['run_compiler']

import argparse
import sys
import os

from compiler import *

def run_compiler(root_dir: str, build_options: BuildOptions):
    '''
    Run the Chai compiler.
    
    Params
    ------
    root_dir: str
        The root compilation directory.
    build_options: BuildOptions
        The build options for the compiler.
    '''

    # Create and run the compiler.
    c = Compiler(root_dir, build_options)
    if isinstance(rt_code := c.compile(), int):
        exit(rt_code)
    else:
        exit(-1)


if __name__ == '__main__':
    # Declare the command-line argument parser.
    parser = argparse.ArgumentParser()
    parser.add_argument('root_dir', help='the package directory to compile')
    parser.add_argument('-o', '--output_path', help='the output path for the binary')    

    # Parse the command-line arguments.
    try:
        parse_result = parser.parse_args()
    except Exception:
        exit(1)

    # Extract the build config from the parse result.
    root_dir = parse_result.root_dir

    if parse_result.output_path:
        output_path = parse_result.output_path
    else:
        # output_path = os.path.join(root_dir, os.path.basename(root_dir))
        # TEMP CODE
        output_path = os.path.join(os.path.dirname(root_dir), os.path.basename(root_dir).replace('.chai', '.exe'))

    if sys.platform == 'win32' and os.path.splitext(output_path)[1] == '':
        output_path += '.exe'

    build_options = BuildOptions(output_path)

    # Run the compiler.
    run_compiler(root_dir, build_options)  
