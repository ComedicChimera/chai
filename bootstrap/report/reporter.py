'''Provides all the classes used to handle errors.'''

__all__ = [
    'CompileError',
    'Reporter',
    'LogLevel'
]

from dataclasses import dataclass
from enum import IntEnum, auto
import traceback
import itertools

from . import TextSpan
from depm.source import SourceFile

@dataclass
class CompileError(Exception):
    '''
    Indicates an error that occurred during compilation.

    Attributes
    ----------
    message: str
        The error message.
    pkg_display_path: str
        The package display path used to identify the package in which the error
        occurred.
    rel_path: str
        The package-root-relative path to the file where the error occurred.
    span: Position
        The text span of the erroneous source text.
    '''

    message: str
    src_file: SourceFile
    span: TextSpan

class LogLevel(IntEnum):
    '''Enumerates the different possible reporter log levels.'''

    SILENT = auto()
    ERROR = auto()
    WARN = auto()
    VERBOSE = auto()

class Reporter:
    '''Responsible for handling all error reporting during compilation.'''

    # The reporter's log level.
    log_level: LogLevel

    # Whether or not there were an errors.
    errored: bool

    def __init__(self, log_level: LogLevel):
        '''
        Params
        ------
        log_level: LogLevel
            The reporter's log level.
        '''

        self.log_level = log_level
        self.errored = False

    def report_compile_error(self, cerr: CompileError):
        '''
        Reports a compilation error.

        Params
        ------
        cerr: CompileError
            The compilation error to report.
        '''

        self.errored = True

        if self.log_level != LogLevel.SILENT:
            display_compile_error(cerr)

    def report_fatal_error(self, ferr: Exception):
        '''
        Reports an unexpected fatal error and exits the compiler.

        Params
        ------
        ferr: Exception
            The exception to report as a fatal error.
        '''

        if self.log_level != LogLevel.SILENT:
            print('[fatal error]:')
            traceback.print_exception(ferr)
            exit(-1)

    def report_error(self, msg: str, kind: str):
        '''
        Reports a normal/top-level error that doesn't refer to specific place in
        user source text.
        '''

        self.errored = True

        if self.log_level != LogLevel.SILENT:
            print(f'[{kind} error] {msg}')

    @property
    def return_code(self) -> int:
        '''Gets the return code for compiler program.'''

        return 1 if self.errored else 0

# ---------------------------------------------------------------------------- #

def display_compile_error(cerr: CompileError):
    '''
    Displays a compile error.
    
    Params
    ------
    cerr: CompileError
        The compile error to display.
    '''

    # Print the error banner.
    print(f'[error] ({cerr.src_file.parent.display_path}) {cerr.src_file.rel_path}:', end='')

    # Print the error message on the same line as the banner while escaping any
    # newlines that occur within the message.
    escaped_msg = cerr.message.replace('\n', '\\n')
    print(f'{cerr.span.start_line}:{cerr.span.start_col}: {escaped_msg}', end='\n\n')

    # Find all the source lines the file over which the error occurs.
    lines = []
    with open(cerr.src_file.abs_path) as f:
        for n, line in enumerate(f):
            if cerr.span.start_line <= n + 1 <= cerr.span.end_line:
                # Replace all tabs with four spaces for indentation calculation.
                lines.append(line.replace('\t', ' ' * 4))

    # Calculate the minimum indentation of the lines (so we can cut off all
    # leading indentation shared by the lines).
    min_indent = min(len(list(itertools.takewhile(lambda c: c == ' ', line))) for line in lines)

    # Calculate the maximum length of the line numbers converted to strings.
    # This is used for padding the line numbers.  Since the end line number
    # always has the greatest value (and thus length), it is sufficient to
    # simply use its length for this metric.
    max_line_num_len = len(str(cerr.span.end_line))

    for i, line in enumerate(lines):
        # Print the line number and padding bar.
        print(str(i + cerr.span.start_line).ljust(max_line_num_len) + ' | ', end='')

        # Print the line with the shared minimum indent trimmed off (so we
        # aren't) printing lines halfway across the screen.
        print(line[min_indent:], end='')

        # Print line and padding bar used for the carret underlining line beneat
        # the erroneous source line.
        print(' ' * max_line_num_len + ' | ', end='')

        # Calculate the number of spaces before carret underlining begins. For
        # any line which is not the starting line, this is always zero since the
        # underlining is always continuing from the previous line. For all other
        # lines, it is start column - the minimum indent. The extra - 1 corrects
        # for the fact that Chai stores column numbers as starting at 1.
        if i == 0:
            carret_prefix_count = cerr.span.start_col - min_indent - 1
        else:
            carret_prefix_count = 0

        # Calculate the number of characters at the end of the source line that
        # should not be highlighted.  For all lines except the last line, this
        # is zero, since underlining should span until the end of the line and
        # over onto the next line.  For the last line, it is length of the line
        # - the end column of the errorenous source text. The extra + 1 corrects
        # for the fact the Chai stores column numbers as starting at 1.
        if i == len(lines) - 1:
            carret_suffix_count = len(line) - cerr.span.end_col + 1
        else:
            carret_suffix_count = 0

        # Print the number of spaces that come before the carret (ie. skip
        # underlining until the start column).
        print(' ' * carret_prefix_count, end='')

        # Print the underlining carrets for the given line.  This number is
        # always the length of the line - (the number of carret that are skipped
        # at the start + the number of carrets that are skipped at the end).  We
        # also subtract off the minimum indentation to account for the fact that
        # that part of the line is neglected by the prefix count (ie. to cancel
        # the - minimum indent inside the calculate for carret prefix count).
        print('^' * (len(line) - carret_prefix_count - carret_suffix_count - min_indent))
