package syntax

import (
	"bufio"
	"chai/depm"
	"chai/report"
	"fmt"
)

// NOTE: All parsing functions (that are not utility/API functions) are
// commented with the EBNF notation of the grammar they parse as well as any
// semantic actions they perform during parsing.

// Parser is the parser for a Chai source file. They perform three primary
// tasks: syntax analysis, AST generation, and symbol/import resolution. The
// parser itself acts as a state machine the moves over the file token by token
// and deciding what to parse based on the token it currently positioned over
// and its context (implicit from the callstack of parsing functions): it is a
// recursive descent parser.  All parsing functions assume that they begin with
// the parser centered on the first token of their production and must consume
// all tokens (including the last) of their production, leaving the parser on
// the next token.  Parsers are created once per file.
type Parser struct {
	// chFile is the Chai source file being parsed.
	chFile *depm.ChaiFile

	// lexer is the Lexer this parser is using to lex the source file.
	lexer *Lexer

	// tok is the current token the parser is positioned on.
	tok *Token
}

// NewParser creates a new parser for the given file and file reader.
func NewParser(chFile *depm.ChaiFile, r *bufio.Reader) *Parser {
	return &Parser{
		chFile: chFile,
		lexer:  NewLexer(chFile.Context, r),
	}
}

// Parse parses a file and writes the resulting AST to the Chai file if it
// succeeds.  It returns whether or not the file should be added to its parent
// package.  This boolean does NOT necessarily indicate whether parsing
// succeeded or failed.
func (p *Parser) Parse() bool {
	// move the parser onto the first token
	if p.next() {
		return false
	}

	// parse the file
	if defs, ok := p.parseFile(); ok {
		// store the defs in the Chai file
		p.chFile.Defs = defs

		return true
	}

	return false
}

// -----------------------------------------------------------------------------

// next moves the parser forward one token.
func (p *Parser) next() bool {
	if tok, ok := p.lexer.NextToken(); ok {
		p.tok = tok
		return true
	}

	return false
}

// got returns true if the parser is on a token of a given kind.
func (p *Parser) got(kind int) bool {
	return p.tok.Kind == kind
}

// assert checks if the parser is on a token of a given kind and rejects the
// token if not.  It returns a boolean indicating whether or not the parser is
// on a matching token kind (and should continue).
func (p *Parser) assert(kind int) bool {
	if p.got(kind) {
		return true
	}

	p.reject()
	return false
}

// want moves the parser forward one and then asserts that the token the parser
// has moved to its of a given kind.  It returns a boolean indicating if the
// move forward was successful and the assertion passed.
func (p *Parser) want(kind int) bool {
	if p.next() {
		return p.assert(kind)
	}

	return false
}

// newlines moves the parser forward until a non-newline token is encountered.
// This will leave the parser positioned on the first non-newline character. The
// current token is considered.  This function returns false if at any point the
// parser fails to move forward.
func (p *Parser) newlines() bool {
	for p.got(NEWLINE) {
		if !p.next() {
			return false
		}
	}

	return true
}

// -----------------------------------------------------------------------------

// reject reports an unexpected token error on the current token.
func (p *Parser) reject() {
	report.ReportCompileError(
		p.chFile.Context,
		p.tok.Position,
		fmt.Sprintf("unexpected token: `%s`", p.tok.Value),
	)
}

// rejectWithMsg rejects the current token with a specific message.  The
// function takes a message and arguments to format into it.
func (p *Parser) rejectWithMsg(msg string, a ...interface{}) {
	report.ReportCompileError(
		p.chFile.Context,
		p.tok.Position,
		fmt.Sprintf(msg, a...),
	)
}

// errorOn reports an error on a given token.  The function takes a message and
// arguments to format into it.
func (p *Parser) errorOn(tok *Token, msg string, a ...interface{}) {
	report.ReportCompileError(
		p.chFile.Context,
		tok.Position,
		fmt.Sprintf(msg, a...),
	)
}

// warnOn reports a warning on a given token.  The function takes a message and
// arguments to format into it.
func (p *Parser) warnOn(tok *Token, msg string, a ...interface{}) {
	report.ReportCompileWarning(
		p.chFile.Context,
		tok.Position,
		fmt.Sprintf(msg, a...),
	)
}

// -----------------------------------------------------------------------------

// defineGlobal defines a global symbol.
func (p *Parser) defineGlobal(sym *depm.Symbol) bool {
	return p.chFile.Parent.GlobalTable.Define(p.chFile.Context, sym)
}
