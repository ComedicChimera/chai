from typing import Tuple

from . import ChaiCompileError, text_pos_from_range
from .source import ChaiFile, ChaiModule, ChaiPackage, ChaiPackageImport
from .mod_loader import BuildProfile
from .lexer import *
from .ast import *
from .symbol import *
from .types import *

# ImportCallback is my sloppy way of getting around Python's recursive imports.
ImportCallback = Callable[[ChaiModule, str, str], ChaiPackage]

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
    scopes: List[Dict[str, Symbol]] = []

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
            if sym.name in curr_scope:
                raise ChaiCompileError(self.ch_file.rel_path, sym.def_pos, f'symbol defined multiple times: `{sym.name}`')
            
            curr_scope[sym.name] = sym
            return sym
        else:
            return self.table.define(sym, self.ch_file.rel_path)

    # _lookup attempts to find a symbol by its name
    def _lookup(self, name: str, pos: TextPosition, def_kind: DefKind = DefKind.ValueDef, mutability: Mutability = Mutability.NeverMutated) -> Symbol:
        # only want to perform local look ups for symbols that are values
        if def_kind == DefKind.ValueDef and self.scopes:
            # backwards for shadowing
            for scope in reversed(self.scopes):
                if name in scope:
                    return scope[name]

        # scope lookups failed => global lookup
        return self.table.lookup(
            self.ch_file.parent_id, 
            self.ch_file.rel_path, 
            pos,
            name,
            def_kind,
            mutability
            )

    # _push_scope begins a new local scope
    def _push_scope(self):
        self.scopes.append([])
    
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

    # file = [metadata] {import_stmt} {definition | pub_definition | pub_block}
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

        # {definition | pub_definition | pub_block}
        defs = []
        while True:
            # TODO: publics

            if defin := self._maybe_parse_definition():
                defs.append(defin)
                self._newlines()
            else:
                break
            
        self._assert(TokenKind.EndOfFile)

        return defs

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
        self._want(TokenKind.Identifier)
        first = self.tok
        self._next()

        def parse_pkg_path_tail() -> List[Token]:
            pkg_path_toks = []
            while self._got(TokenKind.Dot):
                self._want(TokenKind.Identifier)
                pkg_path_toks.append(self.tok)
                self._next()

            return pkg_path_toks

        def import_package(mod_tok: Token, pkg_path_toks: List[Token]) -> ChaiPackage:
            mod_name = mod_tok.value

            pkg_path = '.'.join(tok.value for tok in pkg_path_toks)
            pkg = self.import_callback(self.parent_mod, mod_name, pkg_path)
            if not pkg:
                if pkg_path_toks:
                    raise ChaiCompileError(
                        self.ch_file.rel_path, 
                        text_pos_from_range(first.position, pkg_path_toks[-1].position),
                        f'unable to import package `{mod_name}.{pkg_path}`'
                        )
                else:
                    raise ChaiCompileError(
                        self.ch_file.rel_path,
                        first.position,
                        f'unable to import package `{mod_name}`'
                    )

            return pkg

        # 'import' pkg_name ['as' 'ID']
        if self._got(TokenKind.Dot):
            # collect the elements of the package path
            pkg_path_toks = parse_pkg_path_tail()

            # import the package
            pkg = import_package(first, pkg_path_toks)

            # check for renames
            pkg_import_name = pkg.name
            pkg_import_name_pos = pkg_path_toks[-1].position
            if self._got(TokenKind.As):
                self._want(TokenKind.Identifier)
                pkg_import_name = self.tok.value
                pkg_import_name_pos = self.tok.position
                self._next()

            # confirm that the package name doesn't conflict
            if pkg_import_name in self.ch_file.visible_packages:
                raise ChaiCompileError(
                    self.ch_file.rel_path, 
                    pkg_import_name_pos,
                    f'package already imported with name `{pkg_import_name}`' 
                    )
            
            self.table.check_conflict(self.ch_file.rel_path, pkg_import_name, pkg_import_name_pos)
            
            # add it to the list of visible packages
            self.ch_file.visible_packages[pkg_import_name] = pkg
        # 'import' id_list 'from' pkg_name
        elif self._got(TokenKind.Comma):
            imported_symbols = {first.value: first}
            while self._got(TokenKind.Comma):
                self._want(TokenKind.Identifier)
                if self.tok.value in imported_symbols:
                    self._error(f'another symbol with name `{self.tok.value}` already imported')

            self._want(TokenKind.From)

            # TODO: import package
        # 'import' 'ID' 'from' pkg_name
        elif self._got(TokenKind.From):
            imported_symbols = {first.value: first}

            # TODO: import package
        # 'import' 'ID' 'as' 'ID'
        elif self._got(TokenKind.As):
            pass
        # 'import' 'ID'
        elif self._got(TokenKind.NewLine):
            pass
        # else => invalid
        else:
            self._reject()

        # mark the package as imported
        if pkg.id in self.ch_file.parent.import_table:
            self.ch_file.parent.import_table[pkg.id] = ChaiPackageImport(pkg)

        # close off import stmt
        self._assert(TokenKind.NewLine)
        self._next()

    # definition = func_def | type_def
    def _maybe_parse_definition(self) -> Optional[ASTDef]:
        # func_def
        if self._got(TokenKind.Def):
            return self._parse_func_def()
        # TODO: type_def

    # func_def = 'def' 'ID' '(' args_decl ')' [type] ('=' expr | block | 'end')
    # Semantic Actions: defines function
    def _parse_func_def(self) -> ASTDef:
        # get the function's name
        self._want(TokenKind.Identifier)
        name_tok = self.tok

        # push a scope for the arguments and function body
        self._push_scope()

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

    # type_label = prim_type | ref_type | tuple_type | named_type
    # Semantic Actions: lookup named types
    def _parse_type_label(self) -> DataType:
        prim_val = self.tok.kind.value - TokenKind.U8.value + 1
        if 0 < prim_val <= PrimType.NOTHING.value:
            self._next()
            return PrimType(prim_val)

        self._reject()

       
        


    



        

