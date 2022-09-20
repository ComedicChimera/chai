package syntax

import "chaic/report"

// Token represents a single lexical token.
type Token struct {
	// The kind of the token.  This must be one of the enumerated token kinds.
	Kind int

	// The string value of the token.
	Value string

	// The text span over which the token exists.  This may not directly
	// correspond to its value: eg. the value of a string token has the leading
	// quotes trimmed off for convenience.
	Span *report.TextSpan
}

// Enumeration of token kinds.
const (
	TOK_PACKAGE = iota

	TOK_FUNC
	TOK_OPER
	TOK_STRUCT

	TOK_LET
	TOK_CONST

	TOK_IF
	TOK_ELIF
	TOK_ELSE
	TOK_FOR
	TOK_WHILE
	TOK_DO
	TOK_BREAK
	TOK_CONTINUE
	TOK_RETURN

	TOK_AS
	TOK_NULL

	TOK_I8
	TOK_I16
	TOK_I32
	TOK_I64
	TOK_U8
	TOK_U16
	TOK_U32
	TOK_U64
	TOK_F32
	TOK_F64
	TOK_BOOL
	TOK_UNIT

	TOK_PLUS
	TOK_MINUS
	TOK_STAR
	TOK_DIV
	TOK_MOD
	TOK_POW

	TOK_EQ
	TOK_NEQ
	TOK_LT
	TOK_GT
	TOK_LTEQ
	TOK_GTEQ

	TOK_BWAND
	TOK_BWOR
	TOK_BWXOR
	TOK_LSHIFT
	TOK_RSHIFT

	TOK_COMPL

	TOK_NOT
	TOK_LAND
	TOK_LOR

	TOK_ASSIGN
	TOK_INC
	TOK_DEC

	TOK_LPAREN
	TOK_RPAREN
	TOK_LBRACE
	TOK_RBRACE
	TOK_LBRACKET
	TOK_RBRACKET
	TOK_COMMA
	TOK_DOT
	TOK_RANGETO
	TOK_ELLIPSIS
	TOK_SEMI
	TOK_COLON
	TOK_ATSIGN

	TOK_IDENT
	TOK_NUMLIT
	TOK_INTLIT
	TOK_FLOATLIT
	TOK_BOOLLIT
	TOK_RUNELIT
	TOK_STRINGLIT

	TOK_EOF
)
