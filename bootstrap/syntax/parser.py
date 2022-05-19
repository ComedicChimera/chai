from typing import List, Optional, Tuple

from report import CompileError, TextSpan
from depm import Symbol
from depm.source import SourceFile
from typecheck import *
from .ast import *
from .token import Token
from .lexer import Lexer

SWALLOW_TOKENS = {
    Token.Kind.LPAREN,
    Token.Kind.RPAREN,
    Token.Kind.LBRACE,
    Token.Kind.RBRACE,
    Token.Kind.LBRACKET,
    Token.Kind.RBRACKET,
    Token.Kind.COMMA,
}

class Parser:
    srcfile: SourceFile
    lexer: Lexer

    _tok: Optional[Token] = None
    _lookbehind: Optional[Token] = None

    def __init__(self, srcfile: SourceFile):
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
                annots = self.parse_annotation_def()
            else:
                annots = {}

            self.srcfile.definitions.append(self.parse_definition(annots))

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
    
    def parse_annotation_def(self) -> Annotations:
        '''
        annotation_def := '@' (annotation | '[' annotation {',' annotation} ']') ;
        annotation := 'IDENTIFIER' ['(' 'STRINGLIT' ')'] ;
        '''

        annots = {}

        def parse_annotation():
            annot_id = self.want(Token.Kind.IDENTIFIER)

            if annot_id.value in annots:
                self.error(f'annotation {annot_id.value} specified multiple times', annot_id.span)

            if self.has(Token.Kind.LPAREN):
                self.advance()

                annot_value = self.want(Token.Kind.STRINGLIT)

                rparen = self.want(Token.Kind.RPAREN)

                annots[annot_id.value] = (
                    annot_value.value, 
                    TextSpan.over(annot_id.span, rparen.span),
                )
            else:
                annots[annot_id.value] = ("", annot_id.span)         

        self.want(Token.Kind.ATSIGN)

        if self.has(Token.Kind.LBRACKET):
            self.advance()

            parse_annotation()

            while self.has(Token.Kind.COMMA):
                self.advance()

                parse_annotation()

            self.want(Token.Kind.RBRACKET)
        else:
            parse_annotation()

        self.want(Token.Kind.NEWLINE)

        return annots
        

    # ---------------------------------------------------------------------------- #

    def parse_definition(self, annots: Annotations) -> ASTNode:
        '''definition := func_def ;'''

        match (tok := self.tok()).kind:
            case Token.Kind.DEF:
                return self.parse_func_def(annots)
            case _:
                self.reject()

    def parse_func_def(self, annots: Annotations) -> ASTNode:
        '''
        func_def := 'def' 'IDENTIFIER' '(' [func_params] ')' [type_label] func_body ;
        '''

        self.want(Token.Kind.DEF)

        func_id = self.want(Token.Kind.IDENTIFIER)

        self.want(Token.Kind.LPAREN)

        if self.has(Token.Kind.IDENTIFIER):
            func_params = self.parse_func_params()
        else:
            func_params = []

        self.want(Token.Kind.RPAREN)

        match self.tok(False).kind:
            case Token.Kind.NEWLINE | Token.Kind.ASSIGN | Token.Kind.END:
                func_rt_type = PrimitiveType.NOTHING
            case _:
                func_rt_type = self.parse_type_label()

        func_body, end_pos = self.parse_func_body()

        func_type = FuncType([p.type for p in func_params], func_rt_type)
        func_sym = Symbol(
            func_id.value, 
            self.srcfile.parent.id, 
            func_type, 
            Symbol.Kind.FUNC, 
            Symbol.Mutability.IMMUTABLE, 
            func_id.span,
            'intrinsic' in annots,
        )

        self.define_global(func_sym)

        return FuncDef(
            Identifier(func_sym.name, func_sym.def_span, symbol=func_sym),
            func_params,
            func_body,
            annots,
            TextSpan.over(func_id.span, end_pos),
        )

    def parse_func_params(self) -> List[Symbol]:
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
                    params.append(Symbol(
                        ident.name,
                        self.srcfile.parent.id,
                        param_type,
                        Symbol.Kind.VALUE,
                        Symbol.Mutability.NEVER_MUTATED,
                        ident.span
                    ))
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
            case Token.Kind.NEWLINE:
                return self.parse_block(), self.behind.span
            case Token.Kind.ASSIGN:
                self.advance()

                expr = self.parse_expr()

                return expr, expr.span
            case _:
                self.reject()

    # ---------------------------------------------------------------------------- #

    def parse_block(self) -> ASTNode:
        '''
        block := 'NEWLINE' stmt {stmt} 'end' ;
        stmt := (var_decl | expr_assign_stmt) 'NEWLINE' ;
        '''

        self.want(Token.Kind.NEWLINE)

        stmts = []

        while not self.has(Token.Kind.END):
            match self.tok().kind:
                case Token.Kind.LET:
                    stmts.append(self.parse_var_decl())
                case _:
                    stmts.append(self.parse_expr_assign_stmt())

            self.want(Token.Kind.NEWLINE)

        self.want(Token.Kind.END)
        
        return Block(stmts)

    def parse_var_decl(self) -> ASTNode:
        '''
        var_decl := 'let' var_list {',' var_list} ;
        var_list := id_list (type_ext [initializer] | initializer) ;
        '''

        start_pos = self.tok().span
        self.want(Token.Kind.LET)

        def parse_var_list() -> List[VarList]:
            id_list = self.parse_id_list()

            if self.has(Token.Kind.COLON):
                typ = self.parse_type_ext()

                if self.has(Token.Kind.ASSIGN):
                    init = self.parse_initializer()
                else:
                    init = None
            else:
                init = self.parse_initializer()
                typ = None

            for ident in id_list:
                ident.symbol = Symbol(
                    ident.name,
                    self.srcfile.parent.id,
                    typ,
                    Symbol.Kind.VALUE,
                    Symbol.Mutability.NEVER_MUTATED,
                    ident.span
                )

            return VarList([ident.symbol for ident in id_list], init)

        var_lists = [parse_var_list()]

        while self.has(Token.Kind.COMMA, False):
            self.advance()

            var_lists.append(parse_var_list())

        return VarDecl(var_lists, TextSpan.over(start_pos, self.behind.span))

    def parse_expr_assign_stmt(self) -> ASTNode:
        '''
        expr_assign_stmt := atom_expr ;
        '''

        return self.parse_atom_expr()

    # ---------------------------------------------------------------------------- #

    def parse_expr(self) -> ASTNode:
        '''
        expr := unary_expr [expr_suffix] ;
        expr_suffix := 'as' type_label ;
        '''

        expr = self.parse_unary_expr()

        if self.has(Token.Kind.AS, False):
            self.advance()

            dest_type = self.parse_type_label()

            return TypeCast(expr, dest_type, TextSpan.over(expr.span, self.behind.span))
        else:
            return expr

    def parse_unary_expr(self) -> ASTNode:
        '''
        unary_expr := ['&' | '*'] atom_expr ;
        '''

        match (tok := self.tok()).kind:
            case Token.Kind.AMP:
                self.advance()

                atom_expr = self.parse_atom_expr()

                unary_expr = Indirect(atom_expr, TextSpan.over(tok.span, atom_expr.span))
            case Token.Kind.STAR:
                self.advance()

                atom_expr = self.parse_atom_expr()

                unary_expr = Dereference(atom_expr, TextSpan.over(tok.span, atom_expr.span))
            case _:
                unary_expr = self.parse_atom_expr()

        return unary_expr

    def parse_atom_expr(self) -> ASTNode:
        '''
        atom_expr := atom {trailer}
        trailer := '(' expr_list ')' ;
        '''

        atom_expr = self.parse_atom()

        match self.tok(False).kind:
            case Token.Kind.LPAREN:
                self.advance()

                if not self.has(Token.Kind.RPAREN):    
                    args_list = self.parse_expr_list()
                else:
                    args_list = []

                rparen = self.want(Token.Kind.RPAREN)

                atom_expr = FuncCall(
                    atom_expr, 
                    args_list,
                    TextSpan.over(atom_expr.span, rparen.span)
                )

        return atom_expr

    def parse_atom(self) -> ASTNode:
        '''
        atom := 'INTLIT' | 'NUMLIT' | 'RUNELIT' | 'BOOLLIT' | 'IDENTIFIER' | 'null' | '(' expr ')' ;
        '''

        match (tok := self.tok()).kind:
            case Token.Kind.INTLIT | Token.Kind.NUMLIT:
                self.advance()
                return Literal(tok.kind, tok.value, tok.span)
            case Token.Kind.RUNELIT:
                self.advance()
                return Literal(tok.kind, tok.value, tok.span, PrimitiveType.U32)
            case Token.Kind.BOOLLIT:
                self.advance()
                return Literal(tok.kind, tok.value, tok.span, PrimitiveType.BOOL)
            case Token.Kind.IDENTIFIER:
                self.advance()
                return Identifier(tok.value, tok.span)
            case Token.Kind.NULL:
                self.advance()
                return Null(tok.span)
            case Token.Kind.LPAREN:
                self.advance()

                expr = self.parse_expr()

                self.want(Token.Kind.RPAREN)

                return expr
            case _:
                self.reject()

    # ---------------------------------------------------------------------------- #

    def parse_expr_list(self) -> List[ASTNode]:
        '''expr_list := expr {',' expr} ;'''

        exprs = [self.parse_expr()]

        while self.has(Token.Kind.COMMA):
            self.advance()
            exprs.append(self.parse_expr())

        return exprs

    def parse_initializer(self) -> ASTNode:
        '''initializer := '=' expr ;'''

        self.want(Token.Kind.ASSIGN)

        return self.parse_expr()

    def parse_id_list(self) -> List[Identifier]:
        '''id_list := 'IDENTIFIER' {',' 'IDENTIFIER'} ;'''
        
        id_tok = self.want(Token.Kind.IDENTIFIER)
        ids = [Identifier(id_tok.value, id_tok.span)]

        while self.has(Token.Kind.COMMA):
            self.advance()
            id_tok = self.want(Token.Kind.IDENTIFIER)
            ids.append(Identifier(id_tok.value, id_tok.span))

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

        match self.tok().kind:
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

    def define_global(self, sym: Symbol):
        if sym.name in self.srcfile.parent.symbol_table:
            self.error(f'multiple symbols named `{sym.name}` defined in scope', sym.def_span)

        self.srcfile.parent.symbol_table[sym.name] = sym

    # ---------------------------------------------------------------------------- #

    def advance(self):
        '''Moves the parser forward one token.'''

        self._lookbehind = self._tok
        self._tok = self.lexer.next_token()

    def has_swallow_behind(self):
        '''Returns whether the lookbehind is a swallow token.'''

        return self._lookbehind and self._lookbehind.kind in SWALLOW_TOKENS

    def swallow(self):
        '''Consumes newlines until the parser reaches a non-newline token.'''

        while self._tok.kind == Token.Kind.NEWLINE:
            self.advance()

    def has(self, kind: Token.Kind, swallow: bool = True) -> bool:
        '''
        Returns whether or no the parser is on a token of the given kind.

        Params
        ------
        kind: Token.Kind
            The kind of token to check for.
        swallow: bool
            Whether or not to try to swallow before checking.
        '''

        if kind == Token.Kind.NEWLINE or not swallow:
            return self._tok.kind == kind

        return self.tok().kind == kind

    def want(self, kind: Token.Kind) -> Token:
        '''
        Asserts that the token the parser is on is of the given kind. This
        function automatically tries to swallow newlines if the expected token
        is not a newline.

        Params
        ------
        kind: Token.Kind
            The kind of token to check for.

        Returns
        -------
        Token
            The matched token.
        '''

        if kind == Token.Kind.NEWLINE:
            if self._tok.kind != kind and self._tok.kind != Token.Kind.EOF:
                self.reject()
        else:
            if kind in SWALLOW_TOKENS or self.has_swallow_behind():
                self.swallow()        

            if self._tok.kind != kind:
                self.reject()

        self.advance()
        return self._lookbehind
    
    def tok(self, swallow: bool = True) -> Token:
        '''
        Returns the token the parser is currently on.

        Params
        ------
        swallow: bool
            Whether or not to try to swallow before returning the token.
        '''

        if swallow and self.has_swallow_behind():
            self.swallow()

        return self._tok

    @property
    def behind(self) -> Token:
        '''Returns the parser's lookbehind.'''

        return self._lookbehind

    def reject(self):
        '''Rejects the current token giving a default unexpected message.'''

        self.reject_with_msg(f'unexpected token: `{self._tok.value}`')

    def reject_with_msg(self, msg: str):
        '''
        Rejects the current token using the given error message.

        Params
        ------
        msg: str
            The error message.
        '''

        self.error(msg, self._tok.span)

    def error(self, msg: str, span: TextSpan):
        '''
        Raises a parsing compile error.

        Params
        ------
        msg: str
            The error message.
        span: TextSpan
            The span of the error.
        '''

        raise CompileError(msg, self.srcfile.rel_path, span)    
