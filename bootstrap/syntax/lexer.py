from io import TextIOWrapper
import os
from typing import List, Optional

from . import Token
from report import Position, CompileError
from source import Package

class Lexer:
    file: TextIOWrapper
    rel_path: str

    start_line: int = 1
    start_col: int = 1
    line: int = 1
    col: int = 1

    tok_buff: List[str] = []

    def __init__(self, pkg: Package, file_abs_path: str):
        self.file = open(file_abs_path, 'r')
        self.rel_path = os.path.relpath(pkg.abs_path, file_abs_path)

    def close(self):
        self.file.close()

    def next_token(self) -> Token:
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

        self.mark()
        return Token(Token.Kind.EndOfFile, '', self.get_position())

    def lex_newline(self) -> Token:
        self.mark()
        self.read()

        while self.peek() == '\n':
            self.read()

        return self.make_token(Token.Kind.Newline)

    def skip_comment(self) -> Optional[Token]:
        self.skip()

        if self.peek() == '{':
            self.skip()

            while c := self.peek():
                self.skip()

                if c == '}' and self.peek() == '#':
                    break

            return None
                    
        while (c := self.peek()) and c != '\n':
            self.skip()

        if not c:
            return None

        return self.lex_newline()

    def lex_std_string(self) -> Token:
        self.mark()
        self.skip()

        while c := self.peek():
            match c:
                case '\n':
                    self.error('standard string literals cannot contain newlines')
                case '\"':
                    self.skip()
                    return self.make_token(Token.Kind.StringLit)
                case '\\':
                    self.read_escape_sequence()
                case _:
                    self.read()
    
        self.error('unclosed string literal')

    def lex_rune(self) -> Token:
        self.mark()
        self.skip()

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

        if c := self.read():
            if c == '\'':
                self.skip()
                return self.make_token(Token.Kind.RuneLit)
            else:
                self.error(f'expected closing quote not `{c}`')
        else:
            self.error('unclosed rune literal')

    def lex_raw_string(self) -> Token:
        self.mark()
        self.skip()

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
                            self.tok_buff.append('\\')
                            self.read()
                case '`':
                    self.skip()
                    return self.make_token(Token.Kind.StringLit)
                case _:
                    self.read()

        self.error('unclosed string literal')

    def read_escape_sequence(self):
        self.read()

        def read_unicode_sequence(num_chars: int):
            for _ in range(num_chars):
                if c := self.read():
                    if not is_hex(c):
                        self.error('expected a hexadecimal value')
                else:
                    self.error('unexpected end of escape sequence')

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
        self.mark()
        self.read()

        while (c := self.peek()) and (c.isalnum() or c == '_'):
            self.read()

        tok = self.make_token(Token.Kind.Identifier)
        if tok.value in KEYWORDS:
            tok.kind = KEYWORDS[tok.value]
            return tok
        else:
            return tok

    def lex_number(self, c: str) -> Token:
        self.mark()
        self.read()

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
                    return self.make_token(Token.Kind.NumLit)
                case _:
                    base = 10
        else:
            base = 10

        is_float = False
        has_exp = False
        expect_neg = False
        requires_digit = False
        is_long, is_uns = False, False

        while c := self.peek():
            if c == '_':
                self.skip()

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
                        expect_neg = False
                        requires_digit = False
                        self.read()
                    elif requires_digit:
                        if expect_neg:
                            expect_neg = False
                            self.read()
                        else:
                            self.error('expected digit')
                    else:
                        match c:
                            case '.':
                                if is_float:
                                    break

                                is_float = True
                                requires_digit = True
                                self.read()
                            case 'e' | 'E':
                                if has_exp:
                                    break

                                is_float = True
                                has_exp = True
                                expect_neg = True
                                requires_digit = True
                                self.read()                                 
                            case _:
                                break

        match base:
            case 2 | 8 | 16:
                kind = Token.Kind.IntLit
            case _:
                if is_float:
                    kind = Token.Kind.FloatLit
                elif is_long or is_uns:
                    kind = Token.Kind.IntLit
                else:
                    kind = Token.Kind.NumLit

        return self.make_token(kind)

    def lex_symbol_or_fail(self, c: str) -> Token:
        self.mark() 
        self.read()

        if c in SYMBOLS:
            symbol_kind = SYMBOLS[c]

            while (c := self.peek()) and (sym_value := ''.join(self.tok_buff) + c) in SYMBOLS:
                self.read()
                symbol_kind = SYMBOLS[sym_value]

            return self.make_token(symbol_kind)
        else:
            self.error(f'unknown symbol: `{c}`')

    # ---------------------------------------------------------------------------- #

    def make_token(self, kind: Token.Kind) -> Token:
        tok = Token(kind, ''.join(self.tok_buff), self.get_position())

        self.tok_buff.clear()

        return tok

    def mark(self):
        self.start_line = self.line
        self.start_col = self.col

    def error(self, msg: str):
        raise CompileError(msg, self.rel_path, self.get_position())

    def get_position(self) -> Position:
        return Position(self.start_line, self.start_col, self.line, self.col)

    # ---------------------------------------------------------------------------- #

    def read(self) -> Optional[str]:
        c = self.file.read(1)

        if not c:
            return None

        self.update_position(c)
        self.tok_buff.append(c)

        return c

    def peek(self) -> Optional[str]:
        c = self.file.read(1)

        if not c:
            return None

        self.file.seek(self.file.tell() - 1)

        return c

    def skip(self) -> bool:
        c = self.file.read(1)

        if not c:
            return False

        self.update_position(c)

        return True

    def update_position(self, c: str):
        match c:
            case '\n':
                self.line += 1
                self.col = 1
            case '\t':
                self.col += 4
            case _:
                self.col += 1

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
    '\''
}

KEYWORDS = {
    'def': Token.Kind.Def,
    'let': Token.Kind.Let,
    'if': Token.Kind.If
}

SYMBOLS = {
    '+': Token.Kind.Plus,
    '++': Token.Kind.Increment
}

def is_hex(c: str) -> bool:
    return c.isdigit() or 'a' <= c <= 'z' or 'A' <= c <= 'Z'