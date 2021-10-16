from . import ChaiCompileError
from .source import ChaiFile
from .mod_loader import BuildProfile
from .lexer import *
from .ast import *

# Parser is the parser for Chai -- one parser per package
class Parser:
    profile: BuildProfile
    # TODO: symbol table

    ch_file: ChaiFile
    lexer: Lexer

    def __init__(self, profile: BuildProfile) -> None:
        self.profile = profile

    # parse is the main entry point for the parser.  It takes the ChaiFile being
    # parsed and an opened file pointer to parse over.  The contents of the Chai
    # file are updated with the file AST if parsing succeeds.
    def parse(self, ch_file: ChaiFile, fp: TextIOWrapper) -> None:
        # initialize parser state
        self.ch_file = ch_file
        self.lexer = Lexer(ch_file, fp)

        ch_file.defs = self._parse_file()

    # ---------------------------------------------------------------------------- #

    # _reject raises an unexpected token error parsing a file.
    def _reject(self, tok: Token) -> None:
        if tok.kind == TokenKind.EndOfFile:
            msg = 'unexpected end of file'
        else:
            msg = f'unexpected token: `{tok.value}`'

        raise ChaiCompileError(self.ch_file.rel_path, tok.position, msg)

    # _expect asserts that the next token is of a specific type.  If the next
    # token matches this criteria, it is returned.  Otherwise, an unexpected
    # token error is thrown.  Note that this function will skip any unexpected
    # newlines (provided NewLine is the not the expected token kind).
    def _expect(self, kind: TokenKind) -> Token:
        if kind == TokenKind.NewLine:
            if (tok := self.lexer.next_token()).kind == kind:
                return tok
            else:
                self._reject(tok)
        else:
            while (tok := self.lexer.next_token()).kind == TokenKind.NewLine:
                pass

            if tok.kind == kind:
                return tok
            else:
                self._reject(tok)

    # _next retrieves the next token in the token stream.  It optionally skips
    # newlines if the flag argument is set.  It rejects EOF tokens.
    def _next(self, skip_newlines: bool = False) -> Token:
        if skip_newlines:
            while (tok := self.lexer.next_token()).kind == TokenKind.NewLine:
                pass

            if tok.kind == TokenKind.EndOfFile:
                self._reject(tok)

            return tok
        else:
            tok = self.lexer.next_token()
            if tok.kind == TokenKind.NewLine:
                self._reject(tok)

            return tok

    # _next_maybe retrieves the next token in the token stream.  It optionally
    # skips newlines if the flag argument is set.
    def _next_maybe(self, skip_newlines: bool = False) -> Token:
        if skip_newlines:
            while (tok := self.lexer.next_token()).kind == TokenKind.NewLine:
                pass

            return tok
        else:
            return self.lexer.next_token()

    # _one_of returns a token if a the token has a kind supplied in the input
    # *token_kinds. Otherwise, it rejects the token.  It always skips newlines
    # unless a newline is in the *token_kinds.
    def _one_of(self, *token_kinds: TokenKind) -> Token:
        if TokenKind.NewLine in token_kinds:
            tok = self.lexer.next_token()
        else:
            tok = self._next_maybe()

        if tok.kind in token_kinds:
            return tok
        else:
            self._reject(tok)

    # ---------------------------------------------------------------------------- #
    # Below are the set of "production" parser routines.  These function solely
    # to "construct" programmatically the Chai grammar and apply its syntactic
    # rules and semantic actions.  Each parse function will have, in EBNF
    # notation, the production it parses above it.

    # file = [metadata] {import_stmt} {definition | pub_block}
    def _parse_file(self) -> List[ASTDef]:
        # TODO: metadata

        # TODO: import statement

        defs = []
        while (tok := self._one_of(TokenKind.Def, TokenKind.EndOfFile)).kind != TokenKind.EndOfFile:
            defs.append(self._parse_definition(tok))

        return defs


    # definition = func_def
    def _parse_definition(self, first: Token) -> ASTDef:
        return self._parse_func_def(True)

    # func_def = `def` `IDENTIFIER` `(` args_list `)` type func_body
    # `expect_body` indicates whether the parser expects the function to have a
    # body
    def _parse_func_def(self, expect_body: bool) -> ASTDef:
        name = self._expect(TokenKind.Identifier)

        # TODO: define symbol

        # TODO: parse rest


        

