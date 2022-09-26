package syntax

import (
	"bufio"
	"chaic/common"
	"chaic/depm"
	"chaic/report"
	"chaic/types"
	"os"
)

// Parser is responsible for converting a stream of tokens into an AST and
// validating the syntactic correctness of those source tokens.  This parser
// also declared all global symbols so that they are primed for resolution. Chai
// uses a simple recursive descent parser so most of the functions of the parser
// are simply responsible for parsing a particular production of Chai's grammar.
type Parser struct {
	// The Chai source file.
	chFile *depm.ChaiFile

	// The lexer used to generate the token stream for parser.
	lexer *Lexer

	// The current token the parser is positioned on.
	tok *Token

	// The lookbehind the token the parser was previously positioned on.
	lookbehind *Token

	// The expression nesting depth.  This is used to determine whether `{` belongs
	// to a block or a struct literal.  This will be set to `-1` if a `{` opening
	// a block is expected.
	exprNestDepth int
}

// ParseFile parses a Chai file and returns whether or not parsing was
// successful.  All parsing and lexing error are handled internally: they don't
// bubble beyond this function call.
func ParseFile(chFile *depm.ChaiFile) {
	// Open the provided file.
	file, err := os.Open(chFile.AbsPath)
	if err != nil {
		report.ReportStdError(chFile.ReprPath, err)
		return
	}
	defer file.Close()

	// Create a new lexer for the file.
	lexer := NewLexer(bufio.NewReader(file))

	// Create a new parser for the file.
	p := Parser{
		chFile:        chFile,
		lexer:         lexer,
		tok:           nil,
		exprNestDepth: 0,
	}

	// Catch any errors that occur while parsing the file.
	defer report.CatchErrors(chFile.AbsPath, chFile.ReprPath)

	// Run the parser.
	p.parseFile()
}

// -----------------------------------------------------------------------------

// NOTE: Most parsing functions will simply have the EBNF corresponding to the
// production(s) they parse as their documentation comment.

// parseFile is entry/start function for the parsing algorithm.
//
// file := package_decl {definition} ;
func (p *Parser) parseFile() error {
	p.next()

	// TODO: process the package path
	p.parsePackageDecl()

	for !p.has(TOK_EOF) {
		p.parseDefinition()
	}

	return nil
}

// package_decl := 'package' pkg_path ';' ;
func (p *Parser) parsePackageDecl() []*Token {
	p.want(TOK_PACKAGE)

	pkgPath := p.parsePkgPath()

	p.want(TOK_SEMI)

	return pkgPath
}

// pkg_path := 'IDENT' {'.' 'IDENT'} ;
func (p *Parser) parsePkgPath() []*Token {
	pkgPath := []*Token{p.want(TOK_IDENT)}

	for p.has(TOK_COMMA) {
		p.next()

		pkgPath = append(pkgPath, p.want(TOK_IDENT))
	}

	return pkgPath
}

// -----------------------------------------------------------------------------

// defineGlobalSymbol defines a new global symbol.  It reports an error if
// declaration fails, but does not abort parsing.
func (p *Parser) defineGlobalSymbol(sym *common.Symbol) {
	// Acquire and release the table lock.
	p.chFile.Parent.TableMutex.Lock()
	defer p.chFile.Parent.TableMutex.Unlock()

	// Try to define the global symbol: check for conflicts.
	if _, ok := p.chFile.Parent.SymbolTable[sym.Name]; ok {
		p.recError(sym.DefSpan, "multiple symbols named `%s` defined in global scope", sym.Name)
	} else {
		p.chFile.Parent.SymbolTable[sym.Name] = sym
	}
}

// newOpaqueType creates a new opaque type reference.
func (p *Parser) newOpaqueType(name string, span *report.TextSpan) *types.OpaqueType {
	otype := &types.OpaqueType{
		Name:         name,
		ContainingID: p.chFile.Parent.ID,
		Span:         span,
	}

	p.chFile.OpaqueRefs[name] = append(p.chFile.OpaqueRefs[name], otype)

	return otype
}

// -----------------------------------------------------------------------------

// next moves the parser forward one token if possible.
func (p *Parser) next() {
	tok, err := p.lexer.NextToken()
	if err != nil {
		panic(err)
	}

	p.lookbehind = p.tok
	p.tok = tok
}

// has checks if the parser is currently holding a token of the given kind.
func (p *Parser) has(kind int) bool {
	return p.tok.Kind == kind
}

// recError reports a recoverable error: eg. an error that should not abort
// parsing but should cause compilation not to proceed to the next step.
func (p *Parser) recError(span *report.TextSpan, msg string, args ...interface{}) {
	report.ReportCompileError(
		p.chFile.AbsPath,
		p.chFile.ReprPath,
		span,
		msg,
		args...,
	)
}

// error reports an error message on a given token.
func (p *Parser) error(tok *Token, msg string, args ...interface{}) {
	panic(report.Raise(tok.Span, msg, args...))
}

// reject marks the current token as unexpected and aborts the parser.
func (p *Parser) reject() {
	p.error(p.tok, "unexpected token: %s", p.tok.Value)
}

// want asserts that the parser is positioned over a token of the given kind. If
// that assert succeeds, the parser is moved forward one token.  Otherwise, an
// error is returned.
func (p *Parser) want(kind int) *Token {
	if p.has(kind) {
		p.next()
		return p.lookbehind
	}

	p.reject()
	return nil
}
