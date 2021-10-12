import argparse

from chai import ChaiCompileError, ChaiModuleError
from chai.compile import compile_module
from chai.report import report

def build(args):
    output_dir = None

    try:
        output_dir = compile_module(args.mod_path)
    except ChaiCompileError as cce:
        report.report_compile_error(cce)
    except ChaiModuleError as cme:
        report.report_module_error(cme)
    except Exception as e:
        report.report_fatal_error(e)

    report.report_compile_status(output_dir)   


if __name__ == '__main__':
    parser = argparse.ArgumentParser(prog='chai', description='the Chai compiler command line utility')
    subparsers = parser.add_subparsers(title="subcommands")

    parser_build = subparsers.add_parser('build', help='build Chai programs')
    parser_build.add_argument('mod_path', type=str, help='the path to the root module')
    parser_build.set_defaults(func=build)
    
    result = parser.parse_args()
    result.func(result)



