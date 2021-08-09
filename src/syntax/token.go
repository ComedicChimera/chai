package syntax

// Token represents a token read in by the scanner
type Token struct {
	Kind  int
	Value string

	// Line is line number starting at 1
	Line int

	// Col is the column number counted tabs as four columns
	Col int
}

// The various kinds of a tokens supported by the scanner
const (
	// variables
	LET = iota
	VOL

	// control flow
	IF
	ELIF
	ELSE
	FOR
	BREAK
	CONTINUE
	WHEN
	NOBREAK
	WHILE
	FALLTHROUGH
	WITH
	DO
	OF
	MATCH
	CASE
	END

	// function terminators
	RETURN

	// function definitions
	DEF
	ASYNC
	OPER

	// type definitions
	TYPE
	CLASS
	SPACE
	CLOSED
	UNION

	// package keywords
	IMPORT
	PUB
	FROM

	// expression utils
	SUPER
	NULL
	IS
	AWAIT
	AS
	IN
	FN
	THEN

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
	RUNE
	STRING
	ANY
	NOTHING
	SUB

	// arithmetic/function operators
	PLUS
	MINUS
	STAR
	DIVIDE
	FDIVIDE
	MOD
	RAISETO
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
	BXOR
	LSHIFT
	RSHIFT
	COMPL

	// assignment/declaration operators
	ASSIGN // =
	BINDTO // <-

	// monadic operators
	ACC  // ?
	BIND // >>=

	// type proposition
	TYPEPROP // ::

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
	RUNELIT
	BOOLLIT

	// used in parsing algorithm
	EOF
)

// token patterns (matching strings) for keywords
var keywordPatterns = map[string]int{
	"let":         LET,
	"if":          IF,
	"elif":        ELIF,
	"else":        ELSE,
	"for":         FOR,
	"break":       BREAK,
	"continue":    CONTINUE,
	"when":        WHEN,
	"nobreak":     NOBREAK,
	"while":       WHILE,
	"fallthrough": FALLTHROUGH,
	"with":        WITH,
	"do":          DO,
	"of":          OF,
	"return":      RETURN,
	"vol":         VOL,
	"def":         DEF,
	"async":       ASYNC,
	"oper":        OPER,
	"type":        TYPE,
	"closed":      CLOSED,
	"class":       CLASS,
	"space":       SPACE,
	"union":       UNION,
	"import":      IMPORT,
	"pub":         PUB,
	"from":        FROM,
	"super":       SUPER,
	"null":        NULL,
	"is":          IS,
	"await":       AWAIT,
	"as":          AS,
	"match":       MATCH,
	"end":         END,
	"in":          IN,
	"fn":          FN,
	"then":        THEN,
	"case":        CASE,
	"i8":          I8,
	"i16":         I16,
	"i32":         I32,
	"i64":         I64,
	"u8":          U8,
	"u16":         U16,
	"u32":         U32,
	"u64":         U64,
	"f32":         F32,
	"f64":         F64,
	"string":      STRING,
	"bool":        BOOL,
	"rune":        RUNE,
	"any":         ANY,
	"nothing":     NOTHING,
	"sub":         SUB,
}

// token patterns for symbolic items - longest match wins
var symbolPatterns = map[string]int{
	"+":   PLUS,
	"++":  INCREM,
	"-":   MINUS,
	"--":  DECREM,
	"*":   STAR,
	"//":  FDIVIDE,
	"%":   MOD,
	"**":  RAISETO,
	"<":   LT,
	">":   GT,
	"<=":  LTEQ,
	">=":  GTEQ,
	"==":  EQ,
	"!=":  NEQ,
	"!":   NOT,
	"&&":  AND,
	"||":  OR,
	"&":   AMP,
	"|":   PIPE,
	"^":   BXOR,
	"<<":  LSHIFT,
	">>":  RSHIFT,
	"~":   COMPL,
	"=":   ASSIGN,
	".":   DOT,
	"..":  RANGETO,
	"...": ELLIPSIS,
	"@":   ANNOTSTART,
	"(":   LPAREN,
	")":   RPAREN,
	"{":   LBRACE,
	"}":   RBRACE,
	"[":   LBRACKET,
	"]":   RBRACKET,
	",":   COMMA,
	";":   SEMICOLON,
	":":   COLON,
	"::":  TYPEPROP,
	"->":  ARROW,
	"<-":  BINDTO,
	"/":   DIVIDE,
	"?":   ACC,
	">>=": BIND,
}
