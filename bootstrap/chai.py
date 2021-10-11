import argparse

from chai import ChaiCompileError, ChaiModuleError
from chai.compile import compile_module

def build(args):
    try:
        compile_module(args.mod_path)
    except ChaiCompileError as cce:
        print(cce.report())
    except ChaiModuleError as cme:
        print(cme.report())
    except Exception as e:
        print(e)


if __name__ == '__main__':
    parser = argparse.ArgumentParser(prog='chai', description='the Chai compiler command line utility')
    subparsers = parser.add_subparsers(title="subcommands")

    parser_build = subparsers.add_parser('build', help='build Chai programs')
    parser_build.add_argument('mod_path', type=str, help='the path to the root module')
    parser_build.set_defaults(func=build)
    
    result = parser.parse_args()
    result.func(result)



