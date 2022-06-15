from re import L
from typing import List, Optional, Tuple, Dict

from report import TextSpan
from report.reporter import CompileError
from depm import *
from depm.source import SourceFile
from depm.resolver import Resolver
from typecheck import *
from .ast import *
from .token import Token
from .lexer import Lexer

SWALLOW_BEFORE_TOKENS = {
    Token.Kind.RPAREN,
    Token.Kind.RBRACKET,
    Token.Kind.RBRACKET,
    Token.Kind.ELIF,
    Token.Kind.ELSE,
    Token.Kind.END
}

SWALLOW_AFTER_TOKENS = {
    Token.Kind.LPAREN,
    Token.Kind.LBRACE,
    Token.Kind.LBRACKET,
    Token.Kind.COMMA,
    Token.Kind.ASSIGN,
}

class Parser:
    src_file: SourceFile
    lexer: Lexer
    resolver: Resolver

    _tok: Optional[Token] = None
    _lookbehind: Optional[Token] = None

    def __init__(self, resolver: Resolver, src_file: SourceFile):
        self.src_file = src_file
        self.lexer = Lexer(src_file)
        self.resolver = resolver

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

            self.src_file.definitions.append(self.parse_definition(annots))

            self.want(Token.Kind.NEWLINE)

    # ---------------------------------------------------------------------------- #

    def parse_package_stmt(self):
        '''package_stmt := 'package' package_path ;'''

        self.want(Token.Kind.PACKAGE)

        # TODO determine the root package
        id_toks = self.parse_package_path()
        self.src_file.parent.pkg_path = '.'.join(id_tok.value for id_tok in id_toks)

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
        '''definition := func_def | oper_def ;'''

        match (tok := self.tok()).kind:
            case Token.Kind.DEF:
                return self.parse_func_def(annots)
            case Token.Kind.OPER:
                return self.parse_oper_def(annots)
            case _:
                self.reject()

    def parse_func_def(self, annots: Annotations) -> ASTNode:
        '''
        func_def := 'def' 'IDENTIFIER' '(' [func_params] ')' [type_label] func_body ;
        '''

        start_span = self.want(Token.Kind.DEF).span

        func_id = self.want(Token.Kind.IDENTIFIER)

        self.want(Token.Kind.LPAREN)

        if self.has(Token.Kind.IDENTIFIER):
            func_params = self.parse_func_params()
        else:
            func_params = []

        self.want(Token.Kind.RPAREN)

        match self.tok(False).kind:
            case Token.Kind.NEWLINE | Token.Kind.ASSIGN | Token.Kind.END:
                func_rt_type = PrimitiveType.UNIT
            case _:
                func_rt_type = self.parse_type_label()

        func_body, end_span = self.parse_func_body()

        func_type = FuncType([p.type for p in func_params], func_rt_type)
        func_sym = Symbol(
            func_id.value, 
            self.src_file.parent.id, 
            self.src_file.file_number,
            func_type, 
            Symbol.Kind.FUNC, 
            Symbol.Mutability.IMMUTABLE, 
            func_id.span,
            'intrinsic' in annots,
        )

        self.define_global(func_sym)

        return FuncDef(
            func_sym,
            func_params,
            func_body,
            annots,
            TextSpan.over(start_span, end_span),
        )

    def parse_func_params(self) -> List[Symbol]:
        '''
        func_params := func_param {',' func_param} ;
        func_param := id_list type_ext ;

        Returns
        -------
        func_params: List[Symbol]
            The list of symbols representing the parameters to the function.
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
                        self.src_file.parent.id,
                        self.src_file.file_number,
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
        '''
        func_body = 'end' | ended_block | '=' expr ;

        Returns
        -------
        (maybe_body: Optional[ASTNode], body_end: TextSpan)
            maybe_body: If the function has a body, then the expression node
                comprising it is returned.  If not, `None` is returned.
            body_end: The end token of the body.
        '''

        match self.tok(False).kind:
            case Token.Kind.END:
                end_pos = self.tok().span
                self.advance()
                return None, end_pos
            case Token.Kind.NEWLINE:
                return self.parse_ended_block(), self.behind.span
            case Token.Kind.ASSIGN:
                self.advance()

                expr = self.parse_expr()

                return expr, expr.span
            case _:
                self.reject()

    OVERLOADABLE_OPERATORS = {
        Token.Kind.PLUS: 2,
        Token.Kind.MINUS: [1, 2],
        Token.Kind.STAR: 2,
        Token.Kind.DIV: 2,
        Token.Kind.MOD: 2,
        Token.Kind.POWER: 2,
        Token.Kind.EQ: 2,
        Token.Kind.NEQ: 2,
        Token.Kind.LT: 2,
        Token.Kind.GT: 2,
        Token.Kind.LTEQ: 2,
        Token.Kind.GTEQ: 2,
        Token.Kind.AMP: 2,
        Token.Kind.PIPE: 2,
        Token.Kind.CARRET: 2,
        Token.Kind.LSHIFT: 2,
        Token.Kind.RSHIFT: 2,
        Token.Kind.AND: 2,
        Token.Kind.OR: 2,
        Token.Kind.NOT: 1,
        Token.Kind.COMPL: 1
    }

    def parse_oper_def(self, annots: Annotations) -> ASTNode:
        '''
        oper_def = 'oper' '(' operator ')' '(' func_params ')' [type_label] 'func_body' ;
        operator = '+' | '-' | '*' | '/' | '%' | '**' | '==' | '!=' | '<' | '>' | '<='
            | '>=' | '&' | '|' | '^' | '<<' | '>>' | '&&' | '||' | '!' | '~' ;
        '''

        start_span = self.want(Token.Kind.OPER).span

        self.want(Token.Kind.LPAREN)

        op_tok = self.tok()
        arities = self.OVERLOADABLE_OPERATORS.get(op_tok.kind)
        if not arities:
            self.reject()

        op_sym = op_tok.value
        op_span = op_tok.span

        self.advance()
        self.want(Token.Kind.RPAREN)

        self.want(Token.Kind.LPAREN)

        oper_params = self.parse_func_params()

        self.want(Token.Kind.RPAREN)

        match self.tok(False).kind:
            case Token.Kind.NEWLINE | Token.Kind.END | Token.Kind.ASSIGN:
                rt_type = PrimitiveType.UNIT
            case _:
                rt_type = self.parse_type_label()

        oper_body, end_span = self.parse_func_body()

        if isinstance(arities, int):
            arities = [arities]

        if not any(arity == len(oper_params) for arity in arities):
            if len(arities) == 1 and arities[0] == 1:
                expected_arity_repr = '1 argument'
            else:
                expected_arity_repr = ' or '.join(arities) + ' arguments'

            self.error(
                f'operator {op_sym} must take {expected_arity_repr}', 
                TextSpan.over(oper_params[0].def_span, oper_params[-1].def_span)
            )

        oper_type = FuncType([p.type for p in oper_params], rt_type)

        oper_overload = OperatorOverload(
            self.src_file.parent.id,
            self.src_file.file_number,
            oper_type,
            op_span
        )

        self.define_operator_overload(op_tok.kind, op_sym, oper_overload, len(oper_params))

        return OperDef(
            op_sym,
            oper_overload,
            oper_params,
            oper_body,
            annots,
            TextSpan.over(start_span, end_span)
        )

    def define_operator_overload(self, op_kind: Token.Kind, op_sym: str, overload: OperatorOverload, arity: int):
        '''
        Defines an operator overload.  Does NOT check for conflicts.  That is
        done by the resolver.

        Params
        ------
        op_kind: Token.Kind
            The operator kind of the operator being overloaded.
        op_sym: str
            The operator symbol of the operator being overloaded.
        overload: OperatorOverload
            The operator overload to add.
        arity: int
            The arity of the operator overload to define.
        '''

        if operators := self.src_file.parent.operator_table.get(op_kind):
            for operator in operators:
                if operator.arity == arity:
                    operator.overloads.append(overload)
                    break
            else:
                operators.append(Operator(
                    op_kind,
                    op_sym,
                    arity,
                    [overload]
                ))
        else:
            self.src_file.parent.operator_table[op_kind] = [Operator(
                op_kind,
                op_sym,
                arity,
                [overload]
            )]

    def parse_record_typedef(self, annots: Annotations) -> RecordTypeDef:
        '''
        record_type_def := 'record' 'IDENTIFIER' [':' type {',' type}] '=' record_body ;
        record_body := '{' record_field {record_field} '}' ;
        '''

        start_span = self.want(Token.Kind.RECORD).span

        id_tok = self.want(Token.Kind.IDENTIFIER)

        if self.has(Token.Kind.COLON):
            self.advance()

            extends = [self.parse_type_label()]
            while self.has(Token.Kind.COMMA):
                self.advance()
                
                extends.append(self.parse_type_label())
        else:
            extends = []

        self.want(Token.Kind.ASSIGN)

        self.want(Token.Kind.LBRACE)

        fields = {}
        field_inits = {}
        field_annots = {}

        self.parse_record_field(fields, field_inits, field_annots)

        while self.has(Token.Kind.IDENTIFIER):
            self.parse_record_field(fields, field_inits, field_annots)

        end_span = self.want(Token.Kind.RBRACE).span

        rec_type = RecordType(fields, extends)
        rec_sym = Symbol(
            id_tok.value,
            self.src_file.parent.id,
            self.src_file.file_number,
            rec_type,
            Symbol.Kind.TYPEDEF,
            Symbol.Mutability.IMMUTABLE,
            id_tok.span,
        )

        self.define_global(rec_sym)

        return RecordTypeDef(
            rec_sym,
            annots,
            TextSpan.over(start_span, end_span),
            field_inits,
            field_annots,
        )

    def parse_record_field(
        self, 
        fields: Dict[str, RecordField], 
        field_inits: Dict[str, ASTNode], 
        field_annots: Dict[str, Annotations]
    ):
        '''record_field := [annotation_def] ['const'] id_list type_ext [initializer] ;'''

        if self.has(Token.Kind.ATSIGN):
            annots = self.parse_annotation_def()
        else:
            annots = None

        if self.has(Token.Kind.CONST):
            self.advance()
            const = True
        else:
            const = False

        idents = self.parse_id_list()
        typ = self.parse_type_ext()

        if self.has(Token.Kind.ASSIGN):
            init = self.parse_initializer()
        else:
            init = None

        for ident in idents:
            if ident.name in fields:
                self.error(f'multiple fields named `{ident.name}`', ident.span)

            fields[ident.name] = RecordField(ident.name, typ, const, bool(init) or typ.is_nullable)

            if init:
                field_inits[ident.name] = init

            if annots:
                field_annots[ident.name] = annots

        self.want(Token.Kind.NEWLINE)

    # ---------------------------------------------------------------------------- #

    def parse_block_expr(self) -> ASTNode:
        '''block_expr := if_tree | while_loop ;'''

        match self.tok().kind:
            case Token.Kind.IF:
                return self.parse_if_tree()
            case Token.Kind.WHILE:
                return self.parse_while_loop()

    def parse_if_tree(self) -> ASTNode:
        '''
        if_tree := 'if' if_cond block_body {'elif' if_cond block_body} ['else' block_body] 'end' ;
        if_cond := [var_decl ';'] expr ;
        '''

        def parse_cond_branch() -> CondBranch:
            if self.tok().kind in {Token.Kind.LET, Token.Kind.CONST}:
                header_var_decl = self.parse_var_decl()
                self.want(Token.Kind.SEMICOLON)
            else:
                header_var_decl = None

            return CondBranch(
                header_var_decl,
                self.parse_expr(),
                self.parse_block_body(Token.Kind.ELIF, Token.Kind.ELSE, Token.Kind.END)
            )

        start_span = self.want(Token.Kind.IF).span

        cond_branches = [parse_cond_branch()]

        while self.has(Token.Kind.ELIF):
            self.advance()

            cond_branches.append(parse_cond_branch())

        if self.has(Token.Kind.ELSE):
            self.advance()

            else_branch = self.parse_block_body(Token.Kind.END)
        else:
            else_branch = None

        end_span = self.want(Token.Kind.END).span

        return IfTree(cond_branches, else_branch, TextSpan.over(start_span, end_span))

    def parse_while_loop(self) -> ASTNode:
        '''
        while_loop := 'while' [var_decl ';'] expr [';' expr_assign_stmt] block_body 'end' ;
        '''

        start_span = self.want(Token.Kind.WHILE).span

        if self.tok().kind in {Token.Kind.LET, Token.Kind.CONST}:
            header_var_decl = self.parse_var_decl()
            self.want(Token.Kind.SEMICOLON)
        else:
            header_var_decl = None

        condition = self.parse_expr()

        if self.has(Token.Kind.SEMICOLON):
            self.advance()

            update_stmt = self.parse_expr_assign_stmt()
        else:
            update_stmt = None

        block_body = self.parse_block_body(Token.Kind.END)

        end_span = self.want(Token.Kind.END).span

        return WhileLoop(
            header_var_decl,
            condition,
            update_stmt,
            block_body,
            TextSpan.over(start_span, end_span)
        )

    def parse_block_body(self, *end_tokens: Token.Kind) -> ASTNode:
        '''
        block_body := '=>' expr | block ;
        '''

        if self.has(Token.Kind.NEWLINE):
            return self.parse_block(*end_tokens)
        else:
            self.want(Token.Kind.ARROW)

            return self.parse_expr()

    # ---------------------------------------------------------------------------- #

    def parse_ended_block(self) -> ASTNode:
        '''ended_block := block 'end' ;'''

        block = self.parse_block(Token.Kind.END)
        self.want(Token.Kind.END)
        return block

    def parse_block(self, *end_tokens: Token.Kind) -> ASTNode:
        '''block := 'NEWLINE' stmt 'NEWLINE' {stmt 'NEWLINE'} ;'''

        self.want(Token.Kind.NEWLINE)

        stmts = []

        while not self.tok().kind in end_tokens:
            stmts.append(self.parse_stmt())

            self.want(Token.Kind.NEWLINE)
        
        return Block(stmts)

    def parse_stmt(self) -> ASTNode:
        '''
        stmt := var_decl | expr_assign_stmt | simple_stmt | block_expr ;
        simple_stmt := 'break' | 'continue' | 'return' [expr_list] ;
        '''

        match (tok := self.tok()).kind:
            case Token.Kind.LET | Token.Kind.CONST:
                return self.parse_var_decl()
            case Token.Kind.BREAK | Token.Kind.CONTINUE:
                self.advance()
                return KeywordStmt(tok)
            case Token.Kind.RETURN:
                self.advance()

                if self.has(Token.Kind.NEWLINE):
                    return ReturnStmt([], tok.span)
                else:
                    expr_list = self.parse_expr_list()
                    return ReturnStmt(expr_list, TextSpan.over(tok.span, expr_list[-1].span))
            case Token.Kind.IF | Token.Kind.WHILE:
                return self.parse_block_expr()
            case _:
                return self.parse_expr_assign_stmt()

    def parse_var_decl(self) -> ASTNode:
        '''
        var_decl := ('let' | 'const') var_list {',' var_list} ;
        var_list := id_list (type_ext [initializer] | initializer) ;
        '''

        start_pos = self.tok().span
        if self.has(Token.Kind.CONST):
            self.advance()
            const = True
        else:
            self.want(Token.Kind.LET)
            const = False

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
                    self.src_file.parent.id,
                    self.src_file.file_number,
                    typ,
                    Symbol.Kind.VALUE,
                    Symbol.Mutability.IMMUTABLE if const else Symbol.Mutability.NEVER_MUTATED,
                    ident.span
                )

            return VarList([ident.symbol for ident in id_list], init)

        var_lists = [parse_var_list()]

        while self.has(Token.Kind.COMMA, False):
            self.advance()

            var_lists.append(parse_var_list())

        return VarDecl(var_lists, TextSpan.over(start_pos, self.behind.span))

    # The set of valid compound assignment operators.
    COMPOUND_ASSIGN_OPS = {
        Token.Kind.PLUS, Token.Kind.MINUS, Token.Kind.STAR, Token.Kind.DIV,
        Token.Kind.MOD, Token.Kind.POWER, Token.Kind.AMP, Token.Kind.PIPE,
        Token.Kind.CARRET, Token.Kind.AND, Token.Kind.OR
    }

    def parse_expr_assign_stmt(self) -> ASTNode:
        '''
        expr_assign_stmt := lhs_expr [single_assign_suffix | multi_assign_suffix] ;
        single_assign_suffix := assign_op expr_list | '++' | '--' ; 
        multi_assign_suffix := ',' lhs_expr {',' lhs_expr} assign_op expr_list ;
        lhs_expr := ['*'] atom_expr ;
        assign_op := ['+' | '-' | '*' | '/' | '%' | '**' | '&' | '|' | '^' | '&&' | '||' ] '=' ;
        '''

        def parse_lhs_expr() -> ASTNode:
            if self.has(Token.Kind.STAR):
                star_tok = self.tok()
                self.advance()

                atom_expr = self.parse_atom_expr()
                return Dereference(atom_expr, TextSpan.over(star_tok.span, atom_expr.span))

            return self.parse_atom_expr()

        lhs_exprs = [parse_lhs_expr()]

        if self.has(Token.Kind.COMMA):
            self.advance()

            lhs_exprs.append(parse_lhs_expr())
        
        next_op_tok = self.tok(False)

        if next_op_tok.kind in self.COMPOUND_ASSIGN_OPS:
            self.advance()
            compound_op_tok = next_op_tok
        elif next_op_tok.kind == Token.Kind.ASSIGN:
            compound_op_tok = None  
        elif len(lhs_exprs) == 1:
            match next_op_tok.kind:
                case Token.Kind.INCREMENT:
                    self.advance()

                    return IncDecStmt(lhs_exprs[0], AppliedOperator(Token(Token.Kind.PLUS, '+', next_op_tok.span)))
                case Token.Kind.DECREMENT:
                    self.advance()

                    return IncDecStmt(lhs_exprs[0], AppliedOperator(Token(Token.Kind.MINUS, '-', next_op_tok.span)))
                case _:
                    return lhs_exprs[0]                
        else:
            self.reject()    

        self.want(Token.Kind.ASSIGN)

        rhs_exprs = self.parse_expr_list()

        return Assignment(
            lhs_exprs, 
            rhs_exprs, 
            [AppliedOperator(compound_op_tok) for _ in lhs_exprs] if compound_op_tok else []
        )

    # ---------------------------------------------------------------------------- #

    def parse_expr(self) -> ASTNode:
        '''
        expr := or_expr [expr_suffix] | block_expr ;
        expr_suffix := 'as' type_label ;
        '''

        match self.tok().kind:
            case Token.Kind.IF | Token.Kind.WHILE:
                return self.parse_block_expr()
            case _:
                expr = self.parse_binary_expr(0)

                if self.has(Token.Kind.AS):
                    self.advance()

                    dest_type = self.parse_type_label()

                    return TypeCast(expr, dest_type, TextSpan.over(expr.span, self.behind.span))
                else:
                    return expr

    # The table of binary operators ordered by precedence: lowest to highest.
    PRED_TABLE = [
        {Token.Kind.OR, Token.Kind.PIPE},
        {Token.Kind.CARRET},
        {Token.Kind.AND, Token.Kind.AMP},
        {Token.Kind.LT, Token.Kind.GT, Token.Kind.LTEQ, Token.Kind.GTEQ, Token.Kind.EQ, Token.Kind.NEQ},
        {Token.Kind.LSHIFT, Token.Kind.RSHIFT},
        {Token.Kind.PLUS, Token.Kind.MINUS},
        {Token.Kind.STAR, Token.Kind.DIV, Token.Kind.MOD},
        {Token.Kind.POWER}
    ]

    # The precedence level for comparison operators.
    COMP_PRED_LEVEL = 3

    def parse_binary_expr(self, pred_level: int) -> ASTNode:
        '''
        Parses one of the given binary expressions based on the precedence
        level. This function essentially handles all the logic common to all
        binary operators.

        Params
        ------
        pred_level: int
            The precedence level of the expression being parsed: 0 is lowest
            precedence.

        or_expr := xor_expr {('||' | '|') xor_expr} ;
        xor_expr := and_expr {'^' and_expr} ;
        and_expr := comp_expr {('&&' | '&') comp_expr} ;
        comp_expr := shift_expr {('==' | '!=' | '>' | '<' | '>=' | '<=') shift_expr} ;
        shift_expr := arith_expr {('<<' | '>>') arith_expr} ;
        arith_expr := term {('+' | '-') term} ;
        term := factor {('*' | '/' | '%') factor} ;
        factor := unary_expr {'**' unary_expr} ;
        '''

        if pred_level == len(self.PRED_TABLE):
            return self.parse_unary_expr()
        elif pred_level == self.COMP_PRED_LEVEL:
            return self.parse_multi_comparison()

        lhs = self.parse_binary_expr(pred_level + 1)

        while (op_tok := self.tok()).kind in self.PRED_TABLE[pred_level]:
            self.advance()

            rhs = self.parse_binary_expr(pred_level + 1)

            lhs = BinaryOpApp(AppliedOperator(op_tok), lhs, rhs)

        return lhs
        
    def parse_multi_comparison(self) -> ASTNode:
        '''
        Parses a multi-operator comparison (eg. `a < b < c`).  This is
        essentially just a binary operator parser that generates its tree
        differently.
        '''

        lhs = self.parse_binary_expr(self.COMP_PRED_LEVEL + 1)
        prev_operand = None

        while (op_tok := self.tok()).kind in self.PRED_TABLE[self.COMP_PRED_LEVEL]:
            self.advance()

            rhs = self.parse_binary_expr(self.COMP_PRED_LEVEL + 1)

            if prev_operand:
                rhs = BinaryOpApp(AppliedOperator(op_tok), prev_operand, rhs)
                lhs = BinaryOpApp(
                    AppliedOperator(Token(Token.Kind.AND, '&&', TextSpan.over(lhs.span, rhs.span))),
                    lhs, rhs
                )
            else:
                lhs = BinaryOpApp(AppliedOperator(op_tok), lhs, rhs)

            prev_operand = rhs

        return lhs

    def parse_unary_expr(self) -> ASTNode:
        '''
        unary_expr := ['&' ['const'] | '*' | '-' | '!' | '~'] atom_expr ;
        '''

        match (tok := self.tok()).kind:
            case Token.Kind.AMP:
                self.advance()

                if self.has(Token.Kind.CONST):
                    self.advance()
                    const = True
                else:
                    const = False

                atom_expr = self.parse_atom_expr()

                unary_expr = Indirect(atom_expr, const, TextSpan.over(tok.span, atom_expr.span))
            case Token.Kind.STAR:
                self.advance()

                atom_expr = self.parse_atom_expr()

                unary_expr = Dereference(atom_expr, TextSpan.over(tok.span, atom_expr.span))
            case Token.Kind.NOT | Token.Kind.COMPL | Token.Kind.MINUS:
                self.advance()

                atom_expr = self.parse_atom_expr()

                unary_expr = UnaryOpApp(AppliedOperator(tok), atom_expr, TextSpan.over(tok.span, atom_expr.span))
            case _:
                unary_expr = self.parse_atom_expr()

        return unary_expr

    # ---------------------------------------------------------------------------- #

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
        type_label := prim_type_label | ptr_type_label | defined_type_label ;
        prim_type_label := 'bool' | 'i8' | 'u8' | 'u16' | 'i32' | 'u32'
            | 'i64' | 'u64' | 'f32' | 'f64' | 'unit' ;
        ptr_type_label := '*' ['const'] type_label ;
        defined_type_label := 'IDENTIFIER' ['.' 'IDENTIFIER'] ;
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
            case Token.Kind.UNIT:
                typ = PrimitiveType.UNIT
            case Token.Kind.STAR:
                self.advance()

                if self.has(Token.Kind.CONST):
                    self.advance()
                    return PointerType(self.parse_type_label(), True)

                return PointerType(self.parse_type_label(), False)
            case Token.Kind.IDENTIFIER:
                first_id_tok = self.tok()
                self.advance()

                if self.advance(Token.Kind.DOT):
                    self.advance()

                    next_id_tok = self.tok()
                    self.advance()

                    # TODO imported types
                    raise NotImplementedError()
                else:
                    typ = OpaqueType(
                        first_id_tok.value, 
                        self.src_file.parent.id,
                        self.src_file.parent.name,
                        first_id_tok.span
                    )

                self.resolver.add_opaque_ref(self.src_file, typ)
                return typ
            case _:
                self.reject()

        self.advance()
        return typ

    # ---------------------------------------------------------------------------- #

    def define_global(self, sym: Symbol):
        if sym.name in self.src_file.parent.symbol_table:
            self.error(f'multiple symbols named `{sym.name}` defined in scope', sym.def_span)

        self.src_file.parent.symbol_table[sym.name] = sym

    # ---------------------------------------------------------------------------- #

    def advance(self):
        '''Moves the parser forward one token.'''

        self._lookbehind = self._tok
        self._tok = self.lexer.next_token()

    def has_swallow_behind(self):
        '''Returns whether the lookbehind is a swallow token.'''

        return self._lookbehind and self._lookbehind.kind in SWALLOW_AFTER_TOKENS

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

        if kind in SWALLOW_BEFORE_TOKENS:
            self.swallow()

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
            if kind in SWALLOW_BEFORE_TOKENS or self.has_swallow_behind():
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

        raise CompileError(msg, self.src_file, span)    
