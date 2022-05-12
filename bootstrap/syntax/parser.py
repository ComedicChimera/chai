from typing import List, Optional

from report import CompileError
from depm.source import Package, SourceFile
from .ast import *
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
        '''
        The main entry point for the parsing algorithm.

        file := package_stmt {definition} ;
        '''

        self.advance()
        
        self.parse_package_stmt()

        while not self.has(Token.Kind.EOF):
            if self.has(Token.Kind.ATSIGN):
                # TODO annotations
                pass

            self.srcfile.definitions.append(self.parse_definition())

    # ---------------------------------------------------------------------------- #

    def parse_package_stmt(self):
        '''package_stmt := 'package' package_path ;'''

        self.want(Token.Kind.PACKAGE)

        # TODO determine the root package
        self.parse_package_path()

        self.want(Token.Kind.NEWLINE)

    def parse_package_path(self) -> List[Token]:
        '''package_path := 'IDENTIFIER' {'.' 'IDENTIFIER'} ;'''

        id_toks = [self.want(Token.Kind.IDENTIFIER)]

        while self.has(Token.Kind.DOT):
            self.advance()

            self.want(Token.Kind.IDENTIFIER)

        return id_toks

    # ---------------------------------------------------------------------------- #

    def parse_definition(self) -> ASTNode:
        '''definition := func_def ;'''

        match (tok := self.tok()).kind:
            case Token.Kind.DEF:
                self.parse_func_def()
            case _:
                self.reject()

    def parse_func_def(self) -> ASTNode:
        '''
        func_def := 'def' 'IDENTIFIER' '(' func_def_args ')' ['type_label'] func_body ;
        '''

        self.want(Token.Kind.DEF)

        func_id = self.want(Token.Kind.IDENTIFIER)

        self.want(Token.Kind.LPAREN)

        # TODO func params

        self.want(Token.Kind.RPAREN)

        # TODO type label

        # TODO func body

    def parse_func_params(self) -> List[FuncParam]:
        '''
        func_params := func_param {',' func_param} ;
        func_param := id_list type_ext ;
        '''

    # ---------------------------------------------------------------------------- #

    def parse_id_list(self) -> List[Identifier]:
        '''id_list := 'IDENTIFIER' {',' 'IDENTIFIER'} ;'''

    def parse_type_ext(self) -> Type:
        '''type_ext := ':' type_label ;'''

        self.want(Token.Kind.COLON)

        return self.parse_type_label()

    def parse_type_label(self) -> Type:
        '''
        type_label := prim_type_label | ptr_type_label ;
        prim_type_label := 'bool' | 'i8' | 'u8' | 'u16' | 'i32' | 'u32'
            | 'i64' | 'u64' | 'f32' | 'f64' | 'nothing' ;
        ptr_type_label := '*' type_label ;
        '''

        


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
        if kind == Token.Kind.NEWLINE:
            if self._tok.kind != kind and self._tok.kind != Token.Kind.EOF:
                self.reject()
        else:
            if kind in SWALLOW_TOKENS or self._lookbehind in SWALLOW_TOKENS:
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

    



    
