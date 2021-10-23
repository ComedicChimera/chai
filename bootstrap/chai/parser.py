from typing import Tuple, Optional
from dataclasses import dataclass

from . import ChaiCompileError, text_pos_from_range
from .source import ChaiFile, ChaiModule, ChaiPackage, ChaiPackageImport
from .mod_loader import BuildProfile
from .lexer import *
from .ast import *
from .symbol import *
from .types import *

# ImportCallback is my sloppy way of getting around Python's recursive imports.
ImportCallback = Callable[[ChaiModule, str, str], Optional[ChaiPackage]]

# LocalScope represents a single lexical scope in Chai along with relevant
# contextual information such as the enclosing function.
@dataclass
class LocalScope:
    # symbols is the dictionary of symbols local to the immediate enclosing
    # scope.  These can be the arguments of a function if the function is the
    # enclosing scope of a function.
    symbols: Dict[str, Symbol]

    # rt_type is the return type of the enclosing function.
    rt_type: DataType

    # TODO: control flow flags

# Parser is the parser for Chai -- one parser per package.  This is recursive
# descent parser that acts a state machine -- moving forward one toekn at a time
# and considering that token.  It also performs semantic actions as it parses
# such as interacting with the symbol table, checking for duplicate names, etc.
# It does NOT perform any type checking.
class Parser:
    parent_mod: ChaiModule
    profile: BuildProfile
    
    table: SymbolTable

    ch_file: ChaiFile
    lexer: Lexer

    # tok is the current token being considered during parsing
    tok: Token

    # public indicates whether the top level defined symbols should be public or
    # not -- this primarily facilitates the behavior of public blocks.
    public: bool = False

    # scopes is the stack of subscopes declared as the parser parses
    scopes: List[LocalScope] = []

    # import_callback is a callback to the `import_package` function
    import_callback: ImportCallback

    def __init__(self, parent_mod: ChaiModule, profile: BuildProfile, table: SymbolTable, import_callback: ImportCallback) -> None:
        self.parent_mod = parent_mod
        self.table = table
        self.profile = profile
        self.import_callback = import_callback

    # parse is the main entry point for the parser.  It takes the ChaiFile being
    # parsed and an opened file pointer to parse over.  The contents of the Chai
    # file are updated with the file AST if parsing succeeds.
    def parse(self, ch_file: ChaiFile, fp: TextIOWrapper) -> bool:
        # initialize parser state
        self.ch_file = ch_file
        self.lexer = Lexer(ch_file, fp)

        self._next()
        if defs := self._parse_file():
            ch_file.defs = defs
            return True

        return False

    # ---------------------------------------------------------------------------- #

    # The following are parsing utility functions.  The exist to quickly
    # manipulate or test the parser's current state.  Due to desire for
    # concision, the following symbols are employed in these function's comments
    # to denote certain behaviors, namely: (!) denotes that the function can
    # raise an error, (>>) denotes that function moves the parser forward one or
    # more tokens.

    # _next (>>) moves the parser forward one token.
    def _next(self):
        self.tok = self.lexer.next_token()

    # _got returns true if the current token is of the given kind.
    def _got(self, kind: TokenKind) -> bool:
        return self.tok.kind == kind

    # _assert rejects the current token if it isn't of the given kind.
    def _assert(self, kind: TokenKind) -> None:
        if not self._got(kind):
            self._reject()

    # _ahead (>>) returns true if the next token is of the given kind.
    def _ahead(self, kind: TokenKind) -> bool:
        self._next()
        return self._got(kind)

    # _want (>>, !) rejects the next token if it isn't of the given kind.
    def _want(self, kind: TokenKind) -> None:
        if not self._ahead(kind):
            self._reject()

    # _got_one_of returns true if the current token is one of the given kinds.
    def _got_one_of(self, *kinds: TokenKind) -> bool:
        return self.tok.kind in kinds

    # _assert_one_of rejects the current token if it isn't one of the given
    # kinds.
    def _assert_one_of(self, *kinds: TokenKind) -> None:
        if not self._got_one_of(*kinds):
            self._reject()

    # _ahead_one_of (>>) returns true if the next token is one of the given
    # kinds.
    def _ahead_one_of(self, *kinds: TokenKind) -> bool:
        self._next()
        return self._got_one_of(*kinds)

    # _want_one_of (>>, !) rejects the next token if it isn't of one of the
    # given kinds.
    def _want_one_of(self, *kinds: TokenKind) -> None:
        if not self._ahead_one_of(*kinds):
            self._reject()

    # _newlines (>>) moves the parser forward until a token is encountered that
    # isn't a newline.  The current token is considered.  This function should
    # NOT be called at the end of a parsing function -- the function generally
    # won't know if the following whitespace is significant.
    def _newlines(self) -> None:
        while self._got(TokenKind.NewLine):
            self._next()

    # ---------------------------------------------------------------------------- #

    # _error produces an error on the current token while parsing a file.
    def _error(self, msg: str):
        raise ChaiCompileError(self.ch_file.rel_path, self.tok.position, msg)

    # _reject reports the current token as unexpected
    def _reject(self) -> None:
        if self.tok.kind == TokenKind.EndOfFile:
            self._error('unexpected end of file')
        elif self.tok.kind == TokenKind.NewLine:
            self._error('unexpected newline')
        else:
            self._error(f'unexpected token: `{self.tok.value}`') 

    # ---------------------------------------------------------------------------- #

    # _define defines in a new symbol in the current scope.  It returns the
    # symbol to use for the defined symbol in all future usages (see the
    # SymbolTable.define comment for more explanation)
    def _define(self, sym: Symbol) -> Symbol:
        if self.scopes:
            curr_scope = self.scopes[-1]
            if sym.name in curr_scope.symbols:
                raise ChaiCompileError(self.ch_file.rel_path, sym.def_pos, f'symbol defined multiple times: `{sym.name}`')
            
            curr_scope.symbols[sym.name] = sym
            return sym
        else:
            return self.table.define(sym, self.ch_file.rel_path)

    # _lookup attempts to find a symbol by its name
    def _lookup(self, name: str, pos: TextPosition, def_kind: DefKind = DefKind.ValueDef, mutability: Mutability = Mutability.NeverMutated) -> Symbol:
        # only want to perform local look ups for symbols that are values
        if def_kind == DefKind.ValueDef and self.scopes:
            # backwards for shadowing
            for scope in reversed(self.scopes):
                if name in scope.symbols:
                    return scope.symbols[name]

        # scope lookups failed => global lookup
        return self.table.lookup(
            self.ch_file.parent_id, 
            self.ch_file.rel_path, 
            pos,
            name,
            def_kind,
            mutability
            )

    # _push_scope begins a new local scope (optionally specifying the enclosing
    # function, if not the return type of the parent scope is used)
    def _push_scope(self, func: Optional[FuncType] = None):
        # caller specified a function => this is enclosing scope of that function
        if func:
            self.scopes.append(LocalScope({
                # add the arguments to the local scope
                Symbol(arg.name, arg.typ, None, self.ch_file.parent.id, False, DefKind.ValueDef, Mutability.NeverMutated) for arg in func.args
            }, func.rt_type))
        else:
            assert len(self.scopes) > 0, "local scope must define an enclosing function"
            self.scopes.append(LocalScope({}, self.scopes[-1].rt_type))
    
    # _pop_scope pops a new local scope
    def _pop_scope(self):
        self.scopes.pop()

    # ---------------------------------------------------------------------------- #

    # The following functions describe the parser's grammar.  Each function is
    # commented with the (approximate) EBNF notation of its grammar.
    # Additionally, any function that (itself not its nonterminals) performs a
    # semantic action is commented with that action.  
    #
    # The convention for these parsing function is that they consume all the
    # tokens associated with their grammar and leave the parser positioned on
    # the token immediately after their production.  The functions should also
    # expect to begin with the parser positioned on the first token of their
    # production.  All the non-maybe functions assume that the first token is
    # correct unless otherwise specified.
    # 
    # The ``maybe` prefix is used for parsing functions that will parse their
    # production if it exists or will simple do nothing if it doesn't.  They
    # will return their AST node if they parsed or None.

    # file = [metadata] {import_stmt} {definition | ['pub'] definition | annotated_def | pub_block}
    def _parse_file(self) -> Optional[List[ASTDef]]:
        self._newlines()

        # [metadata]
        if self._maybe_parse_metadata():
            return None

        self._newlines()

        # {import_stmt}
        while self._got(TokenKind.Import):
            self._parse_import_stmt()
            self._newlines()

        # {definition | ['pub'] definition | annotated_def | pub_block}
        defs = []
        while True:
            # TODO: publics, annotated_def

            if defin := self._maybe_parse_definition():
                defs.append(defin)
                self._newlines()
            else:
                break
            
        self._assert(TokenKind.EndOfFile)

        return defs

    # FILE HEADER PARSING:

    # metadata = '!' '!' metadata_tag {',' metadata_tag}
    # metadata_tag = 'ID' '=' 'STRING'
    # This function returns True if the file should not be compiled.
    def _maybe_parse_metadata(self) -> bool:
        if self._got(TokenKind.Not):
            self._want(TokenKind.Not)

            while True:
                self._want(TokenKind.Identifier)
                meta_name = self.tok.value
                meta_value = ""

                # check for `no_compile`
                if meta_name == 'no_compile':
                    return True

                if meta_name in self.ch_file.metadata:
                    self._error(f'metadata named `{meta_name}` declared multiple times')

                self._next()

                # handle metadata with values
                if self._got(TokenKind.Assign):
                    self._want(TokenKind.StringLit)

                    meta_value = self.tok.value

                    # check for os/arch conditions
                    if meta_name == 'os' and meta_value != self.profile.target_os:
                        return True

                    if meta_name == 'arch' and meta_value != self.profile.target_arch:
                        return True

                    self._next()

                if self._got(TokenKind.Comma):
                    self.ch_file.metadata[meta_name] = meta_value
                elif self._got(TokenKind.NewLine):
                    self.ch_file.metadata[meta_name] = meta_value
                    self._next()
                    break
                else:
                    self._reject()
                    
            return False

        return False

    # import_stmt = 'import' (pkg_name ['as' 'ID'] | id_list 'from' pkg_name)
    # pkg_name = 'ID' {'.' 'ID'}
    # Semantic Actions: imports a package, declares imported symbols or imported
    # package.
    def _parse_import_stmt(self) -> None:
        # collect leading identifier
        self._want(TokenKind.Identifier)
        first = self.tok
        self._next()

        # utility function to build package path
        def build_package_path(pkg_path_root_pos: TextPosition, mod_name_tok: Token) -> Tuple[str, TextPosition, Token]:
            pkg_path = []
            pkg_import_name_tok = mod_name_tok
            while self._got(TokenKind.Dot):
                self._want(TokenKind.Identifier)
                pkg_path.append(self.tok.value)
                pkg_import_name_tok = self.tok
                pkg_path_root_pos = text_pos_from_range(pkg_path_root_pos, self.tok.position)
                self._next()

            return '.'.join(pkg_path), pkg_path_root_pos, pkg_import_name_tok

        # collect data from AST as follows:
        imported_symbols = {}
        
        # 'import' id_list 'from' pkg_path
        if self._got(TokenKind.Comma):
            # consider the first token an imported symbol
            imported_symbols[first.value] = first

            # read in the ID list
            while self._got(TokenKind.Comma):
                self._want(TokenKind.Identifier)

                if self.tok.value in imported_symbols:
                    self._error(f'symbol `{self.tok.value}` imported multiple times')
                
                imported_symbols[self.tok.value] = self.tok

                self._next()

            # get the package path
            self._assert(TokenKind.From)
            self._want(TokenKind.Identifier)
            pkg_import_name_tok = self.tok
            mod_name = self.tok.value
            pkg_path_pos = self.tok.position
            self._next()

            pkg_path, pkg_path_pos, pkg_import_name_tok = build_package_path(pkg_path_pos, pkg_import_name_tok)
        # 'import' pkg_name ['as' 'ID']
        elif self._got(TokenKind.Dot):
            mod_name = first.value
            pkg_path, pkg_path_pos, pkg_import_name_tok = build_package_path(first.position, first)

            # check for alias
            if self._got(TokenKind.As):
                self._next()
                self._want(TokenKind.Identifier)

                pkg_import_name_tok = self.tok

                self._next()
        # 'import' 'ID' 'as' 'ID'
        elif self._got(TokenKind.As):
            mod_name = first.value
            pkg_path, pkg_path_pos = "", first.position

            self._want(TokenKind.Identifier)
            pkg_import_name_tok = self.tok
            self._next()
        # 'import' 'ID' 'from' pkg_name
        elif self._got(TokenKind.From):
            imported_symbols[first.value] = first

            self._want(TokenKind.Identifier)
            mod_name = self.tok.value
            mod_name_tok = self.tok
            self._next()

            pkg_path, pkg_path_pos, pkg_import_name_tok = build_package_path(mod_name_tok.position, mod_name_tok)
        # 'import' 'ID' 
        elif self._got(TokenKind.NewLine):
            mod_name = first.value
            pkg_path, pkg_path_pos = "", first.position
        # otherwise => reject
        else:
            self._reject()

        # import the package
        pkg = self.import_callback(self.ch_file.parent.parent, mod_name, pkg_path)
        if not pkg:
            raise ChaiCompileError(self.ch_file.rel_path, pkg_path_pos, 'unable to import package')

        # import symbols from package
        if imported_symbols:
            for symbol_tok in imported_symbols.values():
                if symbol_tok.value in self.ch_file.imported_symbols:
                    raise ChaiCompileError(
                        self.ch_file.rel_path,
                        symbol_tok.position,
                        f'multiple symbols imported with name `{symbol_tok.value}`'
                    )
                elif symbol_tok.value in self.ch_file.visible_packages:
                    raise ChaiCompileError(
                        self.ch_file.rel_path,
                        symbol_tok.position,
                        f'imported symbol name `{symbol_tok.value}` conflicts with name of imported package'
                    )
                
                self.table.check_conflict(self.ch_file.rel_path, symbol_tok.value, symbol_tok.position)

                self.ch_file.imported_symbols[symbol_tok.value] = pkg.global_table.lookup(
                    self.ch_file.parent.id,
                    self.ch_file.rel_path,
                    symbol_tok.position,
                    symbol_tok.value,
                    DefKind.Unknown,
                    Mutability.NeverMutated
                )
        # import package by name
        else:
            if pkg_import_name_tok.value in self.ch_file.visible_packages:
                raise ChaiCompileError(
                    self.ch_file.rel_path, 
                    pkg_import_name_tok.position, 
                    f'multiple packages imported with name `{pkg_import_name_tok.value}`')
            elif pkg_import_name_tok.value in self.ch_file.imported_symbols:
                raise ChaiCompileError(
                    self.ch_file.rel_path,
                    pkg_import_name_tok.position,
                    f'imported package name `{pkg_import_name_tok.value}` conflicts with name of imported symbol'
                )

            self.table.check_conflict(self.ch_file.rel_path, pkg_import_name_tok.value, pkg_import_name_tok.position)
            self.ch_file.visible_packages[pkg_import_name_tok.value] = pkg

        # mark the package as imported
        if pkg.id in self.ch_file.parent.import_table:
            self.ch_file.parent.import_table[pkg.id] = ChaiPackageImport(pkg)

        # close off import stmt
        self._assert(TokenKind.NewLine)
        self._next()

    # DEFINITION PARSING:

    # definition = func_def | type_def | space_def | oper_def
    def _maybe_parse_definition(self) -> Optional[ASTDef]:
        # func_def
        if self._got(TokenKind.Def):
            return self._parse_func_def(True)
        # TODO: type_def, space_def, oper_def

    # func_def = 'def' 'ID' '(' args_decl ')' [type] ('=' expr | block | 'end')
    # Semantic Actions: defines function
    def _parse_func_def(self, expects_body: bool) -> ASTDef:
        # get the function's name
        self._want(TokenKind.Identifier)
        name_tok = self.tok

        # parse the arguments
        self._want(TokenKind.LParen)
        self._next()

        args = self._maybe_parse_args_decl()

        self._assert(TokenKind.RParen)

        # handle return type if it exists
        if not self._ahead_one_of(TokenKind.NewLine, TokenKind.Assign, TokenKind.End):
            rt_type = self._parse_type_label()
        else:
            rt_type = PrimType.NOTHING

        typ = FuncType(args, rt_type)

        # push the enclosing scope of the body
        self._push_scope(typ)

        # handle the body
        if self._got(TokenKind.End):
            self._want(TokenKind.NewLine)
        else:
            self._reject()

        # pop the enclosing scope of the function
        self._pop_scope()

        # define the symbol
        sym = self._define(Symbol(
            name_tok.value, 
            typ, 
            name_tok.position, 
            self.ch_file.parent_id, 
            self.public, 
            DefKind.ValueDef, 
            Mutability.Immutable
            ))

        return FuncDef(sym)

    # args_decl = arg_decl {',' arg_decl}
    # arg_decl = arg_name {',' arg_name} ':' type_label
    # arg_name = ['&'] 'ID'
    # Semantic Actions: checks for duplicate arguments, defines arguments
    def _maybe_parse_args_decl(self) -> List[FuncArg]:
        args = []

        # arg_names stores each argument's name paired with a bool indicating
        # whether or not it is by reference
        arg_names = {}

        def _parse_arg_name() -> Tuple[str, bool]:
            by_ref = False
            if self._got(TokenKind.Amp):
                by_ref = True
                self._next()

            self._assert(TokenKind.Identifier)

            name = self.tok.value
            if name in arg_names:
                self._error(f'multiple arguments declared with name: `{name}`')
            
            arg_names[name] = by_ref

            self._next()
            return name, by_ref
                
        # parse only if there is an argument
        if self._got_one_of(TokenKind.Identifier, TokenKind.Amp):
            while True:
                arg_front = [_parse_arg_name()]
                while self._got(TokenKind.Comma):
                    self._next()
                    arg_front.append(_parse_arg_name())

                self._assert(TokenKind.Colon)
                self._next()

                typ = self._parse_type_label()

                for name, by_ref in arg_front:
                    args.append(FuncArg(name, typ, by_ref))

                if self._got(TokenKind.Comma):
                    self._next()
                    self._newlines()
                else:
                    break

        return args

    # EXPR PARSING:

    # expr = simple_expr | block_expr
    def _parse_expr(self) -> ASTExpr:
        # TODO: check for block expr keywords

        # no block expr keyword => simple expr
        return self._parse_simple_expr()

    # simple_expr = or_expr ['as' type_label]
    # Semantic Actions: apply type cast assertions
    def _parse_simple_expr(self) -> ASTExpr:
        # TODO: ['as' type_label]
        return self._parse_bin_op()

    # or_expr = xor_expr {('||' | '|') xor_expr}
    # xor_expr = and_expr {'^' and_expr}
    # and_expr = comp_expr {('&&' | '&') comp_expr}
    # comp_expr = shift_expr {('==' | '!=' | '<' | '>' | '<=' | '>=') shift_expr}
    # shift_expr = arith_expr {('>>' | '<<') arith_expr}
    # arith_expr = term {('+' | '-') term}
    # term = factor {('*' | '/' | '//') factor}
    # factor = unary_expr {'**' unary_expr}
    # Semantic Actions: lookup operator overloads, apply operator type
    # constraints
    def _parse_bin_op(self) -> ASTExpr:
        # TODO: parse binary operators
        return self._parse_unary_expr()

    # unary_expr = ['&' | '*' | '-' | '!' | '~'] atom_expr ['?']
    # Semantic Actions: lookup operators overloads, apply operator type
    # constraints
    def _parse_unary_expr(self) -> ASTExpr:
        # TODO: parse unary exprs
        return self._parse_atom_expr()

    # atom_expr = atom {trailer}
    # trailer = '.' ('IDENTIFIER' | 'INT_LIT') 
    #   | '(' args ')' 
    #   | '[' (expr [':' expr] | ':' expr) ']' 
    #   | '{' struct_init '}'
    # struct_init = ('...' 'IDENTIFIER' | 'ID' initializer) {',' 'ID' initializer}
    # Semantic Actions: get fields from records (structs, hybrids, tuples,
    # packages, etc.), apply call constraints and struct constraints, lookup
    # operator overloads, apply operator type constraints
    def _parse_atom_expr(self) -> ASTExpr:
        # TODO: parse trailer
        return self._parse_atom()

    # atom = 'INT_LIT' | 'FLOAT_LIT' | 'NUM_LIT' | 'STRING_LIT' | 'RUNE_LIT'
    #   | 'BOOL_LIT' | 'ID' | 'NULL' | '(' [expr_list] ')'
    # Semantic Actions: look up identifiers
    def _parse_atom(self) -> ASTExpr:
        # if self._got(TokenKind.BoolLit):
        #     return ASTLit()
        # TODO
        return None

    # COMMON PARSING

    # expr_list = expr {',' expr}
    def _parse_expr_list(self) -> List[ASTExpr]:
        exprs = [self._parse_expr()]

        while self.got(TokenKind.Comma):
            self._next()
            exprs.append(self._parse_expr())

        return exprs   

    # type_label = prim_type | ref_type | tuple_type | named_type
    # Semantic Actions: lookup named types
    def _parse_type_label(self) -> DataType:
        prim_val = self.tok.kind.value - TokenKind.U8.value + 1
        if 0 < prim_val <= PrimType.NOTHING.value:
            self._next()
            return PrimType(prim_val)

        # TODO: ref_type, tuple_type, named_type

        self._reject()

       
        


    



        

