'''The entry point for the Chai compiler.'''

import argparse
import sys
import os

from compiler import Compiler, BuildOptions

def run_build_cmd(parse_result):
    '''Run the build command.'''
    
    # Extract the build config from the parse result.
    root_dir = parse_result.root_dir

    if parse_result.output_path:
        output_path = parse_result.output_path
    else:
        output_path = root_dir / os.path.basename(root_dir)

    if sys.platform == 'win32' and os.path.splitext(output_path)[1] == '':
        output_path += '.exe'

    build_options = BuildOptions(output_path)

    # Create and run the compiler.
    c = Compiler(root_dir, build_options)
    c.compile()
    

if __name__ == '__main__':
    # Declare the command-line argument parser.
    parser = argparse.ArgumentParser()
    cmd_parsers = parser.add_subparsers(title='commands', dest='command')

    # Declare the parser for the `build`` command.
    build_parser = cmd_parsers.add_parser('build', help='compile source code')

    build_parser.add_argument('root_dir', help='the package directory to compile', required=True)
    build_parser.add_argument('-o', '--output_path', help='the output path for the binary')

    # Parse the command-line arguments.
    try:
        parse_result = parser.parse_args()
    except Exception:
        exit(1)

    # Run the specified command.
    match parse_result.command:
        case 'build':
            run_build_cmd(parse_result)




    
