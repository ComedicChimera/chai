from typing import List, Optional

from report import CompileError
from depm.source import Package, SourceFile
from .ast import ASTNode
from .token import Token
from .lexer import Lexer

class Parser:
    pkg: Package
    srcfile: SourceFile
    lexer: Lexer

    _tok: Optional[Token] = None
    _lookbehind: Optional[Token] = None

    def __init__(self, pkg: Package, srcfile: SourceFile):
        self.pkg = pkg
        self.srcfile = srcfile
        self.lexer = Lexer(srcfile)

    def __enter__(self):
        return self

    def __exit__(self):
        self.lexer.close()

    def parse(self):
        self.advance()
        
        self.parse_package_stmt()


    # ---------------------------------------------------------------------------- #

    def parse_package_stmt(self):
        self.want(Token.Kind.PACKAGE)

        if self.has(Token.Kind.COMMA):
            pass

    def parse_package_path(self) -> List[ASTNode]:
        id_tok = self.want(Token.Kind.IDENTIFIER)

    # ---------------------------------------------------------------------------- #

    def advance(self):
        self._lookbehind = self._tok
        self._tok = self.lexer.next_token()

    def swallow(self):
        while self._tok.kind == Token.Kind.NEWLINE:
            self.advance()

    def has(self, kind: Token.Kind) -> bool:
        if kind == Token.Kind.NEWLINE:
            return self._tok.kind == kind

        return self.tok().kind == kind

    def want(self, kind: Token.Kind) -> Token:
        if kind in SWALLOW_TOKENS or kind != Token.Kind.NEWLINE and self._lookbehind in SWALLOW_TOKENS:
            self.swallow()

        if self._tok.kind != kind:
            self.reject()

        self.advance()
        return self._lookbehind
    
    def tok(self, swallow=True):
        if swallow and self._lookbehind in SWALLOW_TOKENS:
            self.swallow()

        return self._tok

    def reject(self):
        self.error(f'unexpected token: `{self._tok.value}`')

    def error(self, msg: str):
        raise CompileError(msg, self.srcfile.rel_path, self._tok.span)

        

SWALLOW_TOKENS = {
    Token.Kind.LPAREN,
    Token.Kind.RPAREN,
    Token.Kind.LBRACE,
    Token.Kind.RBRACE,
    Token.Kind.LBRACKET,
    Token.Kind.RBRACKET,
    Token.Kind.COMMA,
}

    



    
