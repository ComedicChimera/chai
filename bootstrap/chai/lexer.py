from enum import Enum, auto
from dataclasses import dataclass
from io import TextIOWrapper
from os import symlink
import struct
from typing import Optional

from . import ChaiCompileError, TextPosition
from .source import ChaiFile

# TokenKind is an enumeration of the different kinds of Chai tokens.
class TokenKind(Enum):
    # Keyword Tokens
    Def = auto()
    Type = auto()
    Closed = auto()
    End = auto()

    Import = auto()
    From = auto()
    Pub = auto()
    Priv = auto()

    If = auto()
    Elif = auto()
    Else = auto()
    While = auto()
    After = auto()
    Break = auto()
    Continue = auto()
    Do = auto()
    Match = auto()
    Case = auto()

    # Operator Tokens
    Plus = auto()
    Minus = auto()
    Star = auto()
    Div = auto()
    FloorDiv = auto()
    Power = auto()
    Modulo = auto()

    Amp = auto()
    Pipe = auto()
    Carret = auto()
    LShift = auto()
    RShift = auto()
    Compl = auto()

    And = auto()
    Or = auto()
    Not = auto()

    Eq = auto()
    NEq = auto()
    Gt = auto()
    Lt = auto()
    GtEq = auto()
    LtEq = auto()

    Assign = auto()
    Increment = auto()
    Decrement = auto()

    Dot = auto()
    RangeTo = auto()
    Ell = auto()

    # Punctuation Tokens
    LParen = auto()
    RParen = auto()
    LBrace = auto()
    RBrace = auto()
    LBracket = auto()
    RBracket = auto()
    Comma = auto()
    Semicolon = auto()
    Colon = auto()
    Arrow = auto()

    # Value Tokens
    Identifier = auto()
    StringLit = auto()
    RuneLit = auto()
    IntLit = auto()
    FloatLit = auto()
    NumLit = auto()
    BoolLit = auto()
    Null = auto()

    # Special Tokens
    NewLine = auto()
    EndOfFile = auto()

# Token represents a single lexical element of Chai source code.
@dataclass
class Token:
    kind: TokenKind
    value: str
    position: TextPosition

# ESCAPE_CODES is mapping of supported basic escape codes in Chai
ESCAPE_CODES = {
    'a': '\a',
    'b': '\b',
    'f': '\f',
    'n': '\n',
    'r': '\r',
    't': '\t',
    'v': '\v',
    '0': '\0',
    '\\': '\\',
    '\"': '\"',
    '\'': '\''
}

# KEYWORD_PATTERNS is a map of different patterns of keywords
KEYWORD_PATTERNS = {
    'def': TokenKind.Def,
    'type': TokenKind.Type,
    'closed': TokenKind.Closed,
    'end': TokenKind.End,

    'import': TokenKind.Import,
    'from': TokenKind.From,
    'pub': TokenKind.Pub,
    'priv': TokenKind.Priv,

    'if': TokenKind.If,
    'elif': TokenKind.Elif,
    'else': TokenKind.Else,
    'while': TokenKind.While,
    'after': TokenKind.After,
    'break': TokenKind.Break,
    'continue': TokenKind.Continue,
    'do': TokenKind.Do,
    'match': TokenKind.Match,
    'case': TokenKind.Case,

    'true': TokenKind.BoolLit,
    'false': TokenKind.BoolLit,
    'null': TokenKind.Null,
}

# SYMBOL_PATTERNS is a map of the different patterns for operators and
# punctuation (ie. symbols)
SYMBOL_PATTERNS = {
    '+': TokenKind.Plus,
    '-': TokenKind.Minus,
    '*': TokenKind.Star,
    '/': TokenKind.Div,
    '//': TokenKind.FloorDiv,
    '**': TokenKind.Power,
    '%': TokenKind.Modulo,

    '&': TokenKind.Amp,
    '|': TokenKind.Pipe,
    '^': TokenKind.Carret,
    '>>': TokenKind.LShift,
    '<<': TokenKind.RShift,
    '~': TokenKind.Compl,

    '&&': TokenKind.And,
    '||': TokenKind.Or,
    '!': TokenKind.Not,

    '==': TokenKind.Eq,
    '!=': TokenKind.NEq,
    '<': TokenKind.Lt,
    '>': TokenKind.Gt,
    '<=': TokenKind.LtEq,
    '>=': TokenKind.GtEq,

    '=': TokenKind.Assign,
    '++': TokenKind.Increment,
    '--': TokenKind.Decrement,

    '.': TokenKind.Dot,
    '..': TokenKind.RangeTo,
    '...': TokenKind.Ell,

    '(': TokenKind.LParen,
    ')': TokenKind.RParen,
    '[': TokenKind.LBracket,
    ']': TokenKind.RBracket,
    '{': TokenKind.LBrace,
    '}': TokenKind.RBrace,
    ',': TokenKind.Comma,
    ';': TokenKind.Semicolon,
    ':': TokenKind.Colon,
    '=>': TokenKind.Arrow,
}

# Lexer is Chai's lexical analyzer.  It produces a sequential "stream" of tokens
# via the `next_token` method.
class Lexer:
    ch_file: ChaiFile
    line: int = 1
    col: int = 1

    # start_line and start_col are used to keep track of the start and end of tokens
    start_line: int = 1
    start_col: int = 1

    # fp is the file handle for the file being lexed
    fp: TextIOWrapper

    # tok_buff is holds the current content of the token as it is built up
    tok_buff: str = ""

    def __init__(self, ch_file: ChaiFile, fp: TextIOWrapper) -> None:
        self.ch_file = ch_file
        self.fp = fp

    # next_token reads the next token from the input stream
    def next_token(self) -> Token:
        while ahead := self._peek():
            # skip whitespace
            if ahead in {'\t', ' ', '\v', '\f', '\r'}:
                self._skip()
            # handle newlines
            elif ahead == '\n':
                self._mark()
                self._read()
                return self._make_token(TokenKind.NewLine)
            # handle split-joins
            elif ahead == '\\':
                # mark for error reporting
                self._mark()

                self._skip()

                if self._skip() == '\n':
                    continue
                else:
                    self._error('expected newline immediately after split-join')
            # handle comments
            elif ahead == '#':
                if tok := self._skip_comment():
                    return tok
            # handle strings
            elif ahead == '\"':
                return self._lex_std_string_lit()
            elif ahead == '`':
                return self._lex_raw_string_lit()
            elif ahead == '\'':
                return self._lex_rune_lit()
            elif ahead.isdigit():
                return self._lex_num_lit()
            elif ahead.isidentifier():
                # read in the leading character
                self._read()

                # identifiers are read in greedily: as long as there are more
                # valid identifier characters we keep reading.
                while (ahead := self._peek()) and (ahead.isalnum() or ahead == '_'):
                    self._read()

                if self.tok_buff in KEYWORD_PATTERNS:
                    return self._make_token(KEYWORD_PATTERNS[self.tok_buff])
                else:
                    return self._make_token(TokenKind.Identifier)
            else:
                self._mark()

                # set a default value so we can error if the token doesn't match
                # any of our values
                kind = None

                # for operators and punctuation we want the longest possible
                # match -- until no match is found
                while self.tok_buff + self._peek() in SYMBOL_PATTERNS:
                    self._read()
                    kind = SYMBOL_PATTERNS[self.tok_buff]

                # if there is no kind => no match => lexical error
                if not kind:
                    self._error('unknown symbol')
                else:
                    return self._make_token(kind)

        # if we reach here, we have no more characters to read => EOF
        return Token(TokenKind.EndOfFile, '', None)

    # ---------------------------------------------------------------------------- #

    # _skip_comment skips a line or block comment.  It conditionally returns a
    # newline token if the comment is a line comment.

    def _skip_comment(self) -> Optional[Token]:
        # skip the leading `#`
        self._skip()

        # test if to see if the next token is an exclamation mark => block comment
        if self._peek() == '!':
            # we can skip here since it always takes two tokens to exit a block
            # comment
            while ahead := self._skip():
                # exit if we encountered `!#`
                if ahead == '!' and self._skip() == '#':
                        return
        else:
            while ahead := self._peek():
                if ahead == '\n':
                    break

                self._skip()
            else:
                # we exited from an EOF => no newline
                return

            # return a newline since we hit the end of a line comment
            self._mark()
            self._read()
            return self._make_token(TokenKind.NewLine)

    # _lex_escape_code lexes an escape sequence assuming the leading `\` hasn't
    # been read in yet.  It inserts the escaped character value into the token
    # buffer
    def _lex_escape_code(self) -> None:
        # skip the leading `\`
        self._skip()

        if ahead := self._skip():
            if ahead in ESCAPE_CODES:
                self.tok_buff += ESCAPE_CODES[ahead]
            else:
                self._error(f"unknown escape code: {ahead}")
        else:
            self._error("expected escape sequence not end of file")

    # _lex_std_string_lit lexes a standard string literal.
    def _lex_std_string_lit(self) -> Token:
        self._mark()
        self._skip()

        while ahead := self._peek():
            if ahead == '\"':
                self._skip()
                return self._make_token(TokenKind.StringLit)
            elif ahead == '\\':
                self._lex_escape_code()
            elif ahead == '\n':
                self._error("unexpected newline in string literal")
            else:
                self._read()
        else:
            # if we exit normally => EOF in string
            self._error("unexpected end of file in string literal")

    # _lex_raw_string_lit lexes a raw/multi-line string literal.
    def _lex_raw_string_lit(self) -> Token:
        self._mark()
        self._skip()

        while ahead := self._peek():
            if ahead == '`':
                self._skip()
                return self._make_token(TokenKind.StringLit)
            else:
                self._read()
        else:
            # if we exit normally, => EOF in raw string
            self._error("unexpected end of file in raw string literal")

    # _lex_rune_lit lexes a rune literal.
    def _lex_rune_lit(self) -> Token:
        self._mark()
        self._skip()

        if ahead := self._peek():
            if ahead == '\\':
                self._lex_escape_code()
            elif ahead == '\'':
                self._error("empty rune literal")
            else:
                self._read()
        else:
            self._error("unexpected end of file in rune literal")

        if self._skip() == '\'':
            return self._make_token(TokenKind.RuneLit)
        else:
            self._error("expected closing single quote")

    # _lex_num_lit lexes a number literal.
    def _lex_num_lit(self) -> Token:
        self._mark()

        # can't be None (we peeked it b4)
        first = self._read()

        # set the initial base
        base = 10

        # check for base prefixes
        if first == '0':
            if ahead := self._peek():
                if ahead == 'x':
                    base = 16
                    self._read()
                elif ahead == 'b':
                    base = 2
                    self._read()
                elif ahead == 'o':
                    base = 8
                    self._read()                
        
        # flags collecting analyzing numeric literal
        is_float, has_exp = False, False
        is_long, is_uns = False, False

        # loop over the remaining characters. we don't error if we encounter an
        # unexpected character; rather, we just stop reading (as it integers are
        # commonly nestled next to other more complex tokens)
        while ahead := self._peek():
            if base == 2:
                if ahead == '0' or ahead == '1':
                    self._read()
                    continue
            elif base == 8:
                if '0' <= ahead <= '7':
                    self._read()
                    continue
            elif base == 16:
                if ahead.isdigit() or 'a' <= ahead <= 'f' or 'A' <= ahead <= 'F':
                    self._read()
                    continue
            elif base == 10:
                if ahead.isdigit():
                    self._read()
                    continue
                elif ahead == '.':
                    if has_exp:
                        self._error('float literal cannot have decimal exponent')
                    elif is_float:
                        self._error('float literal cannot have multiple decimals')
                    
                    self._read()
                    is_float = True
                    continue
                elif ahead == 'e' or ahead == 'E':
                    if has_exp:
                        self._error('float literal cannot contain multiple exponents')
                    
                    self._read()
                    has_exp = True
                    is_float = True
                    
                    # test for `-` after exponent
                    if self._peek() == '-':
                        self._read()

                    continue
            
            # handle int suffixes
            if not is_float:
                if ahead == 'u':
                    is_uns = True
                    self._read()

                    if self._peek() == 'l':
                        is_long = True
                        self._read()
                elif ahead == 'l':
                    is_long = True
                    self._read()

                    if self._peek() == 'u':
                        is_uns = True
                        self._read()

            # if we reach here, our ahead didn't match anything so we stop lexing
            break

        if is_float:
            return self._make_token(TokenKind.FloatLit)
        # non-decimal bases, unsigned, and long only apply to integers
        elif base != 10 or is_uns or is_long:
            return self._make_token(TokenKind.IntLit)
        else:
            return self._make_token(TokenKind.NumLit)

    # ---------------------------------------------------------------------------- #

    # _make_token makes a new token from the token buffer of the given kind.
    def _make_token(self, kind: TokenKind) -> Token:
        tok_value = self.tok_buff
        self.tok_buff = ''

        return Token(kind, tok_value, TextPosition(
            self.start_line,
            self.start_col,
            self.line,
            self.col
        ))

    # _error raises a compile error lexing a file
    def _error(self, msg: str) -> None:
        raise ChaiCompileError(
            self.ch_file.rel_path, 
            TextPosition(self.start_line, self.start_col, self.line, self.col),
            msg,
            )

    # _mark marks the current position as the start position for a new token.
    def _mark(self) -> None:
        self.start_line = self.line
        self.start_col = self.col

    # ---------------------------------------------------------------------------- #

    # _read reads a new character into the token buffer.  If the end of a file
    # is encountered, `None` is returned.  Otherwise, the read in character is
    # returned.
    def _read(self) -> Optional[str]:
        c = self.fp.read(1)

        # EOF
        if c == '':
            return None

        self.tok_buff += c
        self._update_pos(c)

        return c

    # _peek peeks one character ahead in the file without actually moving the
    # file forward.  If there is an end of file ahead, `None` is returned.
    # Otherwise, the peeked character is returned.
    def _peek(self) -> Optional[str]:
        before = self.fp.tell()
        c = self.fp.read(1)

        # EOF
        if c == '':
            return None

        # seek back one character if there is a character
        self.fp.seek(before)

        return c

    # _skip skips one character from the file and returns the skipped character
    # if there is one.  If not, `None` is returned.
    def _skip(self) -> Optional[str]:
        c = self.fp.read(1)

        # EOF
        if c == '':
            return None

        self._update_pos(c)
        return c
        
    # _update_pos updates the position of the lexer based on the position of a
    # character
    def _update_pos(self, c: str):
        if c == '\n':
            self.line += 1
            self.col = 1
        elif c == '\t':
            self.col += 4
        else:
            self.col += 1