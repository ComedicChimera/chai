package syntax

import (
	"bufio"
	"chai/depm"
)

// NOTE: AST should be defined in another package (b/c of Go's import rules:
// `syntax` depends on `ast` and `depm` which itself depends on `ast`)

// Parser is the parser for a Chai source file. They perform three primary
// tasks: syntax analysis, AST generation, and symbol/import resolution. The
// parser itself acts as a state machine the moves over the file token by token
// and deciding what to parse based on the token it currently positioned over
// and its context (implicit from the callstack of parsing functions): it is a
// recursive descent parser.  Parsers are created once per file.
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
	// TODO
	return false
}

// -----------------------------------------------------------------------------

// TODO: parsing utility functions: next, got, want, assert

// -----------------------------------------------------------------------------

// TODO: parsing error handling functions: reject, rejectWithMsg, errorOn, warnOn
