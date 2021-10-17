from typing import Union

from . import ChaiCompileError, ChaiModuleError, TextPosition

# Reporter is class responsible for displaying compile information to the user
class Reporter:
    # errors_count stores the number of errors thrown
    errors_count: int = 0

    # warnings_count stores the number of warnings thrown 
    warnings_count: int = 0

    def report_compile_error(self, cc: ChaiCompileError) -> None:
        self.errors_count += 1
        print(f'Compile Error in {cc.rel_path} at {cc.position.start_line}:{cc.position.start_col}: {cc.message}')

    def report_module_error(self, cm: ChaiModuleError) -> None:
        self.errors_count += 1
        print(f'Module Error in `{cm.module_name}`: {cm.message}')

    def report_fatal_error(self, e: Union[Exception, str]) -> None:
        self.errors_count += 1
        print('Fatal Error:', e)

    def report_compile_warning(self, rel_path: str, position: TextPosition, msg: str) -> None:
        self.warnings_count += 1
        print(f'Compile Warning in {rel_path} at {position.start_line}:{position.start_col}: {msg}')

    def report_module_warning(self, mod_name: str, msg: str) -> None:
        self.warnings_count += 1
        print(f'Module Warning in `{mod_name}`: {msg}')

    def report_compile_status(self, output_path: str) -> None:
        if self.errors_count > 0:
            print(f'\ncompilation failed ({self.errors_count} errors, {self.warnings_count} warnings)')
        else:
            print(f'\ncompilation succeeded (0 errors, {self.warnings_count} warnings)')
            print('output written to:', output_path)

    def should_proceed(self) -> bool:
        return self.errors_count == 0

report = Reporter()