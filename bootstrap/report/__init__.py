'''
Provides all classes used to report compilation errors within user source text.
'''

from dataclasses import dataclass

@dataclass
class Position:
    '''
    Indicates a positional range/span within user source text.

    .. note:: Both line numbers and column numbers start at 1.

    Attributes
    ----------
    start_line: int
        The line number where the span begins.
    start_col: int
        The column number where the span begins.
    end_line: int
        The line number where the span ends.
    end_col: int
        The column number where the span ends.
    '''

    start_line: int
    start_col: int
    end_line: int
    end_col: int

@dataclass
class CompileError(Exception):
    '''
    Indicates an error that occurred during compilation.

    Attributes
    ----------
    message: str
        The error message.
    rel_path: str
        The package-root-relative path to the file where the error occurred.
    position: Position
        The position of the erroneous source text.
    '''

    message: str
    rel_path: str
    position: Position

class Reporter:
    '''Responsible for handling all error reporting during compilation.'''

    def __init__(self):
        pass
