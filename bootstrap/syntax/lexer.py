'''Provides Chai's lexical analyzer.'''

from io import TextIOWrapper
from typing import List, Optional

from .token import Token
from report import TextSpan
from report.reporter import CompileError
from depm.source import SourceFile

# The set of valid escape codes.
ESCAPE_CODES = {
    'a',
    'b',
    'f',
    'n',
    'r',
    't',
    'v',
    '0',
    '\"',
    '\'',
}

# Maps valid keyword strings to their token kinds.
KEYWORDS = {
    'package': Token.Kind.PACKAGE,
    'def': Token.Kind.DEF,
    'oper': Token.Kind.OPER,
    
    'let': Token.Kind.LET,
    'const': Token.Kind.CONST,
    
    'if': Token.Kind.IF,
    'elif': Token.Kind.ELIF,
    'else': Token.Kind.ELSE,
    'while': Token.Kind.WHILE,
    'end': Token.Kind.END,

    'break': Token.Kind.BREAK,
    'continue': Token.Kind.CONTINUE,
    'return': Token.Kind.RETURN,

    'as': Token.Kind.AS,

    'bool': Token.Kind.BOOL,
    'i8': Token.Kind.I8,
    'u8': Token.Kind.U8,
    'i16': Token.Kind.I16,
    'u16': Token.Kind.U16,
    'i32': Token.Kind.I32,
    'u32': Token.Kind.U32,
    'i64': Token.Kind.I64,
    'u64': Token.Kind.U64,
    'f32': Token.Kind.F32,
    'f64': Token.Kind.F64,
    'unit': Token.Kind.UNIT,

    'true': Token.Kind.BOOLLIT,
    'false': Token.Kind.BOOLLIT,
    'null': Token.Kind.NULL,
}

# Maps valid symbol strings to their token kinds.
SYMBOLS = {
    '+': Token.Kind.PLUS,
    '-': Token.Kind.MINUS,
    '*': Token.Kind.STAR,
    '/': Token.Kind.DIV,
    '%': Token.Kind.MOD,
    '**': Token.Kind.POWER,

    '<': Token.Kind.LT,
    '>': Token.Kind.GT,
    '<=': Token.Kind.LTEQ,
    '>=': Token.Kind.GTEQ,
    '==': Token.Kind.EQ,
    '!=': Token.Kind.NEQ,

    '<<': Token.Kind.LSHIFT,
    '>>': Token.Kind.RSHIFT,
    '~': Token.Kind.COMPL,

    '&': Token.Kind.AMP,
    '|': Token.Kind.PIPE,
    '^': Token.Kind.CARRET,

    '&&': Token.Kind.AND,
    '||': Token.Kind.OR,
    '!': Token.Kind.NOT,

    '=': Token.Kind.ASSIGN,
    '++': Token.Kind.INCREMENT,
    '--': Token.Kind.DECREMENT,

    '(': Token.Kind.LPAREN,
    ')': Token.Kind.RPAREN,
    '{': Token.Kind.LBRACE,
    '}': Token.Kind.RBRACE,
    '[': Token.Kind.LBRACKET,
    ']': Token.Kind.RBRACKET,
    ',': Token.Kind.COMMA,
    ':': Token.Kind.COLON,
    ';': Token.Kind.SEMICOLON,
    '.': Token.Kind.DOT,
    '=>': Token.Kind.ARROW,
    '@': Token.Kind.ATSIGN,
}

class Lexer:
    '''
    Responsible for converting a source file into a stream of tokens (ie.
    performing lexical analysis).

    .. note: All the `lex_*` methods assume that the character beginning them
        has not been read in but is correct -- ie. the lookahead is valid.
    '''

    # The file being tokenized.
    file: TextIOWrapper

    # The source file being tokenized.
    src_file: SourceFile

    # The line beginning the current token.
    start_line: int = 1

    # The column beginning the current token.
    start_col: int = 1

    # The current line position of the lexer within the input stream.
    line: int = 1

    # The current column position of the lexer within the input stream.
    col: int = 1

    # The buffer storing the contents of the token as it is constructed.
    tok_buff: List[str]

    def __init__(self, src_file: SourceFile):
        '''
        Params
        ------
        pkg: Package
            The package containing the file to tokenize.
        file_abs_path: str
            The absolute path to the file to tokenize.
        '''

        self.file = open(src_file.abs_path, 'r')
        self.src_file = src_file
        self.tok_buff = []

    def close(self):
        '''Disposes of the resources associated with the lexer.'''

        self.file.close()

    def next_token(self) -> Token:
        '''
        Returns the next token read from the stream if there is a token.  If
        not, it returns an EOF token. It can produce compile errors if any
        tokens are malformed.
        '''

        # Use the lookahead to determine what token to lex.
        while c := self.peek():
            match c:
                case ' ' | '\t' | '\v' | '\f' | '\r':
                    self.skip()
                case '\n':
                    return self.lex_newline()
                case '#':
                    if tok := self.skip_comment():
                        return tok
                case '\"':
                    return self.lex_std_string()
                case '\'':
                    return self.lex_rune()
                case '`':
                    return self.lex_raw_string()
                case _:
                    if c.isalpha() or c == '_':
                        return self.lex_ident_or_keyword()
                    elif c.isdigit():
                        return self.lex_number(c)
                    else:
                        return self.lex_symbol_or_fail(c)

        # If we reach the end of the input stream, then we just make and return
        # a new EOF token.  Once this line is reached, it will be reached by all
        # subsequent calls the `next_token` and so an EOF token will be returned
        # from then on out.
        self.mark()
        return Token(Token.Kind.EOF, '', self.get_span())

    # ---------------------------------------------------------------------------- #

    def lex_newline(self) -> Token:
        '''Lexes a newline token.'''

        self.mark()
        self.read()

        # This loop implements newline merging: multiple newlines in sequence
        # are all simply merged together into one newline token.
        while c := self.peek():
            match c:
                case '\n':
                    self.read()
                case ' ' | '\r' | '\t' | '\v' | '\f':
                    self.skip()
                case _:
                    break

        return self.make_token(Token.Kind.NEWLINE)

    def skip_comment(self) -> Optional[Token]:
        '''
        Skips a block or line comment.
        
        .. note: This function only returns a token if it is reading a line comment
            in which case it returns the newline token ending it.
        '''

        self.skip()

        if self.peek() == '{': # Block comment ...
            self.skip()

            while c := self.peek():
                self.skip()

                if c == '}' and self.peek() == '#':
                    break

            return None

        # Line comment...           
        while (c := self.peek()) and c != '\n':
            self.skip()

        # Line comment ended with an EOF, not a newline => return nothing.
        if not c:
            return None

        # Line comments always end with a newline.
        return self.lex_newline()

    def lex_std_string(self) -> Token:
        '''Lexes a standard string literal.'''

        # Skip the opening double quote.
        self.mark()
        self.skip()

        # Read the content of the string literal.
        while c := self.peek():
            match c:
                case '\n':
                    self.error('standard string literals cannot contain newlines')
                case '\"':
                    self.skip()
                    return self.make_token(Token.Kind.STRINGLIT)
                case '\\':
                    self.read_escape_sequence()
                case _:
                    self.read()
    
        # If we reach here, the input loop for the string didn't return, so we
        # must assume that it exited based on an EOF which we means the string
        # is unclosed.
        self.error('unclosed string literal')

    def lex_rune(self) -> Token:
        '''Lexes a rune literal.'''

        # Skip the opening single quote.
        self.mark()
        self.skip()

        # Read the content of the rune literal.
        if c := self.peek():
            match c:
                case '\\':
                    self.read_escape_sequence()
                case '\'':
                    self.error('empty rune literal')
                case _:
                    self.read()
        else:
            self.error('unclosed rune literal')

        # Skip the closing single quote.
        if c := self.peek():
            if c == '\'':
                self.skip()
                return self.make_token(Token.Kind.RUNELIT)
            else:
                self.error(f'expected closing quote not `{c}`')
        else:
            self.error('unclosed rune literal')

    def lex_raw_string(self) -> Token:
        '''Lexes a raw string literal.'''

        # Skip the opening backtick.
        self.mark()
        self.skip()

        # Read the content of the raw string.
        while c := self.peek():
            match c:
                case '\\':
                    self.skip()

                    match self.peek():
                        case '\\' | '`':
                            self.read()
                        case None:
                            self.error('unclosed string literal')
                        case _:
                            # The backslash wasn't escaping anything so we need
                            # to "reconsume" it.
                            self.tok_buff.append('\\')
                            self.read()
                case '`':
                    self.skip()
                    return self.make_token(Token.Kind.STRINGLIT)
                case _:
                    self.read()

        # If we reach here, the reading loop exited on an EOF and so we have an
        # unclosed string literal.
        self.error('unclosed string literal')

    def read_escape_sequence(self):
        '''Reads and validates an escape sequence.'''

        # Read in the opening backslash.
        self.read()

        def read_unicode_sequence(num_chars: int):
            '''
            Reads a unicode point hex value from the input stream.

            Params
            ------
            num_chars: int
                The number of digits in the unicode point.
            '''

            for _ in range(num_chars):
                if c := self.read():
                    if not is_hex(c):
                        self.error('expected a hexadecimal value')
                else:
                    self.error('unexpected end of escape sequence')

        # Read the escape code itself.
        if c := self.read():
            match c:
                case 'x':
                    read_unicode_sequence(2)
                case 'u':
                    read_unicode_sequence(4)
                case 'U':
                    read_unicode_sequence(8)
                case _:
                    if c not in ESCAPE_CODES:
                        self.error(f'unknown escape code: \{c}')                        
        else:
            self.error('expected escape sequence not end of file')

    def lex_ident_or_keyword(self) -> Token:
        '''Lexes an identifier or keyword.'''

        # Read the raw identifier.
        self.mark()
        self.read()

        while (c := self.peek()) and (c.isalnum() or c == '_'):
            self.read()

        tok = self.make_token(Token.Kind.IDENTIFIER)

        # If the token value is a keyword, then we convert to the token kind to
        # a keyword kind but leave all the other data the same.
        if tok.value in KEYWORDS:
            tok.kind = KEYWORDS[tok.value]

        return tok

    def lex_number(self, c: str) -> Token:
        '''
        Lexes a numeric literal.

        Params
        ------
        c: str
            The character ahead -- beginning the number, yet to be read in.    
        '''

        self.mark()
        self.read()

        # Handle base prefixes.
        if c == '0':
            match self.peek():
                case 'b':
                    self.read()
                    base = 2
                case 'o':
                    self.read()
                    base = 8
                case 'x':
                    self.read()
                    base = 16
                case None:
                    return self.make_token(Token.Kind.NUMLIT)
                case _:
                    base = 10
        else:
            base = 10

        # Whether or not the number literal is a floating point literal.
        is_float = False

        # Whether the floating point literal has an exponent.
        has_exp = False

        # Whether the lexer is expecting a `-` sign for a negative exponent.
        expect_neg = False

        # Whether the lexer requires that the next character be a digit.
        requires_digit = False

        # Whether or not the integer literal is 64 bit; is unsigned.
        is_long, is_uns = False, False

        while c := self.peek():
            # Skip any `_` used for readability.
            if c == '_':
                self.skip()
                continue

            # Handle integer suffixes.
            if is_uns:
                if self.peek() == 'l':
                    self.read()
                    is_long = True

                break
            elif is_long:
                if self.peek() == 'u':
                    self.read()
                    is_uns = True

                break
                
            # For all of the non-decimal bases, the only possiblity is an
            # integer literal, so simply read in the appropriate digits until no
            # more digits are in the input stream. Base 10 can also have
            # floating point numbers and so requires more complex logic.
            match base:
                case 2:
                    if c == '0' or c == '1':
                        self.read()
                    else:
                        break
                case 8:
                    if '0' <= c <= '7':
                        self.read()
                    else:
                        break
                case 16:
                    if is_hex(c):
                        self.read()
                    else:
                        break
                case 10:
                    if c.isdigit():
                        # Digits always clear the expectation of negative sign
                        # so you can't write something like `1e2-3` as a
                        # literal.
                        expect_neg = False
                        requires_digit = False
                        self.read()
                    elif requires_digit:
                        # `-` signs allow the requirement of a digit to be
                        # postponed.
                        if expect_neg:
                            expect_neg = False
                            self.read()
                        else:
                            self.error('expected digit')
                    else:
                        match c:
                            case '.':
                                # Can't have multiple decimals.
                                if is_float:
                                    break

                                is_float = True

                                # Must have a digit after a decimal.
                                requires_digit = True
                                self.read()
                            case 'e' | 'E':
                                if has_exp:
                                    break

                                # Only floats can use exponents.
                                is_float = True
                                has_exp = True

                                # Exponents allow `-` signs but also require
                                # digits after.
                                expect_neg = True
                                requires_digit = True
                                self.read()
                            case 'u':
                                if not is_float:
                                    is_uns = True
                                else:
                                    break
                            case 'l':
                                if not is_float:
                                    is_long = True
                                else:
                                    break                                 
                            case _:
                                # Not a valid number character, exit the loop.
                                break

        # If there is a digit requirement that was no met, we need to raise an
        # error -- token is malformed.
        if requires_digit:
            self.error('expecting digit')

        # Determine what kind of token to produce.  
        match base:     
            case 2 | 8 | 16:
                # Non-decimal bases => integer literal.
                kind = Token.Kind.INTLIT
            case _:
                if is_float:
                    kind = Token.Kind.FLOATLIT
                elif is_long or is_uns:
                    # Long or unsigned => integer literal.
                    kind = Token.Kind.INTLIT
                else:
                    # Default to number literal.
                    kind = Token.Kind.NUMLIT

        return self.make_token(kind)

    def lex_symbol_or_fail(self, c: str) -> Token:
        '''
        Lexes a punctuation or operator token. If the input stream does not
        contain a valid punctuation or operator token, then an unknown token
        error is raised.

        Params
        ------
        c: str
            The character ahead -- beginning the symbol, yet to be read in.
        '''

        self.mark() 
        self.read()

        # We need to first character to be a valid symbol -- this method is
        # called as a "last resort" for the lexer when it can't match anything
        # else.
        if c in SYMBOLS:
            symbol_kind = SYMBOLS[c]

            # Consume until the ahead is no longer a symbol character.
            while (c := self.peek()) and (sym_value := ''.join(self.tok_buff) + c) in SYMBOLS:
                self.read()
                symbol_kind = SYMBOLS[sym_value]

            return self.make_token(symbol_kind)
        else:
            self.error(f'unknown symbol: `{c}`')

    # ---------------------------------------------------------------------------- #

    def make_token(self, kind: Token.Kind) -> Token:
        '''
        Creates a new token based on the state of the lexer.  Clears the token
        buffer for the next token to be read in.

        Params
        ------
        kind: Token.Kind
            The kind of token to produce.
        '''

        tok = Token(kind, ''.join(self.tok_buff), self.get_span())

        self.tok_buff.clear()

        return tok

    def mark(self):
        '''
        Marks the current position of the lexer as the beginning of a new token.
        '''

        self.start_line = self.line
        self.start_col = self.col

    def error(self, msg: str):
        '''
        Raises a compile error over the lexer's current token span.

        Params
        ------
        msg: str
            The error message.
        '''

        raise CompileError(msg, self.src_file, self.get_span())

    def get_span(self) -> TextSpan:
        '''
        Returns the text span of the current token being lexed: from the marked
        start position to the current position of the lexer.
        '''

        return TextSpan(self.start_line, self.start_col, self.line, self.col)

    # ---------------------------------------------------------------------------- #

    def read(self) -> Optional[str]:
        '''
        Consumes the next character from the input stream into the token buffer
        if another character exists.  Otherwise, nothing is added to the token
        buffer and `None` is returned.
        '''

        c = self.file.read(1)

        if not c:
            return None

        self.update_position(c)
        self.tok_buff.append(c)

        return c

    def peek(self) -> Optional[str]:
        '''
        Returns the next character in the input stream if it exists without
        consuming it or moving the lexer forward.  If there is no token ahead,
        then `None` is returned.
        '''

        c = self.file.read(1)

        if not c:
            return None

        # Unread the character.
        self.file.seek(self.file.tell() - 1)

        return c

    def skip(self) -> bool:
        '''
        Moves the lexer forward one character without consuming it.  This method
        returns whether or not the lexer was able to move forward (ie. is the
        lexer at the end of the input stream -- true if no, false if yes).
        '''

        c = self.file.read(1)

        if not c:
            return False

        self.update_position(c)

        return True

    def update_position(self, c: str):
        '''
        Updates the position of the lexer based on a character.
        
        Params
        ------
        c: str
            The character read in.
        '''

        match c:
            case '\n':
                self.line += 1
                self.col = 1
            case '\t':
                # Tabs count as four columns.
                self.col += 4
            case _:
                self.col += 1

def is_hex(c: str) -> bool:
    '''
    Returns whether a character is a valid hexadecimal digit.

    Params
    ------
    c: str
        The character to test.
    '''
    return c.isdigit() or 'a' <= c <= 'z' or 'A' <= c <= 'Z'