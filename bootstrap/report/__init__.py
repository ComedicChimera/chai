'''Provides the class used to identify places in source text.'''

__all__ = ['TextSpan']

from dataclasses import dataclass

@dataclass
class TextSpan:
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

    @staticmethod
    def over(a: 'TextSpan', b: 'TextSpan') -> 'TextSpan':
        '''
        Returns a text span spanning over and between the two input spans.

        Params
        ------
        a: TextSpan
            The start span of the over text span.
        b: TextSpan
            The end span of the over text span.
        '''

        return TextSpan(a.start_line, a.start_col, b.end_line, b.end_col)
