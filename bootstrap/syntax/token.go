package syntax

import "chai/report"

// Token represents a token read in by the lexer.
type Token struct {
	Kind     int
	Value    string
	Position *report.TextPosition
}

// The various kinds of a tokens supported by the lexer.
const (
	// variables
	LET = iota
	CONST
	VOL

	// control flow
	IF
	ELIF
	ELSE
	FOR
	WHILE
	BREAK
	CONTINUE
	AFTER
	MATCH
	CASE
	WHEN
	FALLTHROUGH
	DO
	END

	// function terminators
	RETURN

	// function definitions
	DEF
	OPER

	// type definitions
	TYPE
	SPACE
	WITH
	CLOSED

	// package keywords
	IMPORT
	PUB
	FROM

	// expression utils
	NULL
	AS
	IN

	// whitespace
	NEWLINE

	// type keywords
	U8
	U16
	U32
	U64
	I8
	I16
	I32
	I64
	F32
	F64
	BOOL
	STRING
	NOTHING

	// arithmetic/function operators
	PLUS
	MINUS
	STAR
	IDIV
	FDIV
	MOD
	POWER
	INCREM
	DECREM

	// boolean operators
	LT
	GT
	LTEQ
	GTEQ
	EQ
	NEQ
	NOT
	AND
	OR

	// bitwise operators
	AMP
	PIPE
	CARRET
	LSHIFT
	RSHIFT
	COMPL

	// assignment/declaration operators
	ASSIGN  // =
	EXTRACT // <-

	// monadic operators
	BIND // >>=
	ACC  // ?

	// dots
	DOT
	RANGETO
	ELLIPSIS

	// punctuation
	ANNOTSTART
	LPAREN
	RPAREN
	LBRACE
	RBRACE
	LBRACKET
	RBRACKET
	COMMA
	SEMICOLON
	COLON
	ARROW

	// literals (and identifiers)
	IDENTIFIER
	STRINGLIT
	INTLIT
	FLOATLIT
	NUMLIT
	RUNELIT
	BOOLLIT

	// used in parsing algorithm
	EOF
)

// token patterns (matching strings) for keywords.
var keywordPatterns = map[string]int{
	"let":   LET,
	"const": CONST,
	"vol":   VOL,

	"if":          IF,
	"elif":        ELIF,
	"else":        ELSE,
	"for":         FOR,
	"while":       WHILE,
	"break":       BREAK,
	"continue":    CONTINUE,
	"after":       AFTER,
	"match":       MATCH,
	"case":        CASE,
	"when":        WHEN,
	"fallthrough": FALLTHROUGH,
	"do":          DO,
	"end":         END,

	"return": RETURN,

	"def":  DEF,
	"oper": OPER,

	"type":   TYPE,
	"space":  SPACE,
	"closed": CLOSED,
	"with":   WITH,

	"import": IMPORT,
	"pub":    PUB,
	"from":   FROM,

	"null": NULL,
	"as":   AS,
	"in":   IN,

	"u8":      U8,
	"u16":     U16,
	"u32":     U32,
	"u64":     U64,
	"i8":      I8,
	"i16":     I16,
	"i32":     I32,
	"i64":     I64,
	"f32":     F32,
	"f64":     F64,
	"bool":    BOOL,
	"string":  STRING,
	"nothing": NOTHING,

	"true":  BOOLLIT,
	"false": BOOLLIT,
}

// token patterns for symbolic items - longest match wins.
var symbolPatterns = map[string]int{
	"+":  PLUS,
	"++": INCREM,
	"-":  MINUS,
	"--": DECREM,
	"*":  STAR,
	"/":  FDIV,
	"//": IDIV,
	"%":  MOD,
	"**": POWER,

	"<":  LT,
	">":  GT,
	"<=": LTEQ,
	">=": GTEQ,
	"==": EQ,
	"!=": NEQ,

	"!":  NOT,
	"&&": AND,
	"||": OR,

	"&":  AMP,
	"|":  PIPE,
	"^":  CARRET,
	"<<": LSHIFT,
	">>": RSHIFT,
	"~":  COMPL,

	"=":  ASSIGN,
	"<-": EXTRACT,

	">>=": BIND,
	"?":   ACC,

	".":   DOT,
	"..":  RANGETO,
	"...": ELLIPSIS,

	"@":  ANNOTSTART,
	"(":  LPAREN,
	")":  RPAREN,
	"{":  LBRACE,
	"}":  RBRACE,
	"[":  LBRACKET,
	"]":  RBRACKET,
	",":  COMMA,
	";":  SEMICOLON,
	":":  COLON,
	"->": ARROW,
}
