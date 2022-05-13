from typing import List, Optional, Tuple

from report import CompileError
from depm.source import Package, SourceFile
from typecheck import FuncType, PointerType, PrimitiveType
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

        file := package_stmt {definition 'NEWLINE'} ;
        '''

        self.advance()
        
        self.parse_package_stmt()

        while not self.has(Token.Kind.EOF):
            if self.has(Token.Kind.ATSIGN):
                # TODO annotations
                pass

            self.srcfile.definitions.append(self.parse_definition())

            self.want(Token.Kind.NEWLINE)

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

        func_params = self.parse_func_params()

        self.want(Token.Kind.RPAREN)

        match self.tok(False).kind:
            case Token.Kind.NEWLINE | Token.Kind.ASSIGN | Token.Kind.END:
                func_rt_type = PrimitiveType.NOTHING
            case _:
                func_rt_type = self.parse_type_ext()

        func_body, end_pos = self.parse_func_body()

        func_type = FuncType([p.type for p in func_params], func_rt_type)
        func_sym = Symbol(
            func_id.value, 
            self.pkg.id, 
            func_type, 
            Symbol.Kind.FUNC, 
            Symbol.Mutability.IMMUTABLE, 
            func_id.span,
        )

        self.define(func_sym)

        return FuncDef(
            Identifier(func_sym.name, func_sym.def_span, symbol=func_sym),
            func_body,
            TextSpan.over(func_id.span, end_pos),
            func_params
        )

    def parse_func_params(self) -> List[FuncParam]:
        '''
        func_params := func_param {',' func_param} ;
        func_param := id_list type_ext ;
        '''

        id_list = self.parse_id_list()
        param_type = self.parse_type_ext()

        params = []
        param_names = set()

        def add_params():
            for ident in id_list:
                if ident.name not in param_names:
                    param_names.add(ident.name)
                    params.append(FuncParam(ident.name, param_type))
                else:
                    self.error(f'multiple parameters defined with same name: `{ident.name}`', ident.span)

        add_params()

        while self.has(Token.Kind.COMMA):
            self.advance()

            id_list = self.parse_id_list()
            param_type = self.parse_type_ext()

            add_params()
        
        return params

    def parse_func_body(self) -> Tuple[Optional[ASTNode], TextSpan]:
        match self.tok(False).kind:
            case Token.Kind.END:
                end_pos = self.tok().span
                self.advance()
                return None, end_pos
            case _:
                self.reject()

    # ---------------------------------------------------------------------------- #

    def parse_id_list(self) -> List[Identifier]:
        '''id_list := 'IDENTIFIER' {',' 'IDENTIFIER'} ;'''
        
        id_tok = self.want(Token.Kind.IDENTIFIER)
        ids = [Identifier(id_tok.value, id_tok.span)]

        while self.has(Token.Kind.COMMA):
            self.advance()
            id_tok = self.want(Token.Kind.IDENTIFIER)
            ids.append(id_tok.value, id_tok.span)

        return ids

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

        match self.tok():
            case Token.Kind.BOOL:
                typ = PrimitiveType.BOOL
            case Token.Kind.I8:
                typ = PrimitiveType.I8
            case Token.Kind.U8:
                typ = PrimitiveType.U8
            case Token.Kind.I16:
                typ = PrimitiveType.I16
            case Token.Kind.U16:
                typ = PrimitiveType.U16
            case Token.Kind.I32:
                typ = PrimitiveType.I32
            case Token.Kind.U32:
                typ = PrimitiveType.U32
            case Token.Kind.I64:
                typ = PrimitiveType.I64
            case Token.Kind.U64:
                typ = PrimitiveType.U64
            case Token.Kind.F32:
                typ = PrimitiveType.F32
            case Token.Kind.F64:
                typ = PrimitiveType.F64
            case Token.Kind.NOTHING:
                typ = PrimitiveType.NOTHING
            case Token.Kind.STAR:
                self.advance()
                return PointerType(self.parse_type_label())
            case _:
                self.reject()

        self.advance()
        return typ

    # ---------------------------------------------------------------------------- #

    def define(self, sym: Symbol):
        # TODO local table

        if sym.name in self.pkg.symbol_table:
            self.error(f'multiple symbols named `{sym.name}` defined in scope', sym.def_span)

        self.pkg.symbol_table[sym.name] = sym

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
        self.reject_with_msg(f'unexpected token: `{self._tok.value}`')

    def reject_with_msg(self, msg: str):
        raise CompileError(msg, self.srcfile.rel_path, self._tok.span)

    def error(self, msg: str, span: TextSpan):
        raise CompileError(msg, self.srcfile.rel_path, span)

        

SWALLOW_TOKENS = {
    Token.Kind.LPAREN,
    Token.Kind.RPAREN,
    Token.Kind.LBRACE,
    Token.Kind.RBRACE,
    Token.Kind.LBRACKET,
    Token.Kind.RBRACKET,
    Token.Kind.COMMA,
}

    



    
