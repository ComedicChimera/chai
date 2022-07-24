package syntax

import (
	"bufio"
	"chaic/report"
	"io"
	"strings"
	"unicode"
)

// Lexer is responsible for tokenizing a source file.
type Lexer struct {
	file    *bufio.Reader
	tokBuff *strings.Builder

	line, col           int
	startLine, startCol int
}

// NewLexer creates a new lexer for the given source file.
func NewLexer(file *bufio.Reader) *Lexer {
	return &Lexer{
		file:    file,
		tokBuff: &strings.Builder{},
		line:    0,
		col:     0,
	}
}

// NextToken retrieves the next token from the input file. If the file has
// ended, this will be an EOF token.
func (l *Lexer) NextToken() (*Token, error) {
	for {
		c, err := l.peek()
		if err != nil {
			return nil, err
		} else if c == -1 {
			break
		}

		switch c {
		case '\n', '\t', ' ', '\r', '\v', '\f':
			l.skip()
		case '/':
			if tok, err := l.lexCommentOrDiv(); tok != nil || err != nil {
				return tok, err
			}
		case '\'':
			return l.lexRuneLit()
		case '"':
			return l.lexStdStringLit()
		case '`':
			return l.lexRawStringLit()
		default:
			if isDecimalDigit(c) {
				return l.lexNumericLit()
			} else if isFirstIdentChar(c) {
				return l.lexIdentOrKeyword()
			} else {
				return l.lexPunctOrOper()
			}
		}
	}

	return &Token{Kind: TOK_EOF}, nil
}

// EnableShiftSplit tells the lexer to automatically split `<<` and `>>` into
// their component `<` and `>` tokens.  This is used when parsing generic tags.
func (l *Lexer) EnableShiftSplit() {
	delete(symbolPatterns, "<<")
	delete(symbolPatterns, ">>")
}

// DisableShiftSplit tells the lexer to stop automatically splitting `<<` and
// `>>`.  This is the default setting for the lexer.
func (l *Lexer) DisableShiftSplit() {
	symbolPatterns["<<"] = TOK_LSHIFT
	symbolPatterns[">>"] = TOK_RSHIFT
}

// -----------------------------------------------------------------------------

// symbolPatterns maps symbol strings (patterns) to their punctuation/operator
// token kind.
var symbolPatterns = map[string]int{
	"+": TOK_PLUS,
	"-": TOK_MINUS,
	"*": TOK_STAR,
	// Division operator is handled with comment logic.
	"%":  TOK_MOD,
	"**": TOK_POW,

	"&":  TOK_BWAND,
	"|":  TOK_BWOR,
	"^":  TOK_BWXOR,
	"~":  TOK_COMPL,
	"<<": TOK_LSHIFT,
	">>": TOK_RSHIFT,

	"==": TOK_EQ,
	"!=": TOK_NEQ,
	"<":  TOK_LT,
	"<=": TOK_LTEQ,
	">":  TOK_GT,
	">=": TOK_GTEQ,

	"&&": TOK_LAND,
	"||": TOK_LOR,
	"!":  TOK_NOT,

	"=":  TOK_ASSIGN,
	"++": TOK_INC,
	"--": TOK_DEC,

	"(": TOK_LPAREN,
	")": TOK_RPAREN,
	"{": TOK_LBRACE,
	"}": TOK_RBRACE,
	"[": TOK_LBRACKET,
	"]": TOK_RBRACKET,
	",": TOK_COMMA,
	".": TOK_DOT,
	";": TOK_SEMI,
	":": TOK_COLON,
	"@": TOK_ATSIGN,
}

// lexPunctOrOper lexes a punctuation or operator symbol.
func (l *Lexer) lexPunctOrOper() (*Token, error) {
	l.mark()
	l.eat()

	kind, ok := symbolPatterns[l.tokBuff.String()]
	if !ok {
		return nil, report.Raise(l.getSpan(), "unknown rune")
	}

	for {
		c, err := l.peek()
		if err != nil {
			return nil, err
		}

		if c == -1 {
			break
		}

		if _kind, ok := symbolPatterns[l.tokBuff.String()+string(c)]; ok {
			l.eat()
			kind = _kind
		} else {
			break
		}
	}

	return l.makeToken(kind), nil
}

// -----------------------------------------------------------------------------

// keywordPatterns maps keyword strings (patterns) to their keyword token kind.
var keywordPatterns = map[string]int{
	"package": TOK_PACKAGE,

	"def":  TOK_DEF,
	"oper": TOK_OPER,

	"let":   TOK_LET,
	"const": TOK_CONST,

	"if":       TOK_IF,
	"elif":     TOK_ELIF,
	"else":     TOK_ELSE,
	"while":    TOK_WHILE,
	"for":      TOK_FOR,
	"break":    TOK_BREAK,
	"continue": TOK_CONTINUE,
	"return":   TOK_RETURN,

	"as":   TOK_AS,
	"null": TOK_NULL,
	"unit": TOK_UNIT,

	"i8":   TOK_I8,
	"i16":  TOK_I16,
	"i32":  TOK_I32,
	"i64":  TOK_I64,
	"u8":   TOK_U8,
	"u16":  TOK_U16,
	"u32":  TOK_U32,
	"u64":  TOK_U64,
	"f32":  TOK_F32,
	"f64":  TOK_F64,
	"bool": TOK_BOOL,
}

// lexIdentOrKeyword lexes an identifier or a keyword.
func (l *Lexer) lexIdentOrKeyword() (*Token, error) {
	l.mark()
	l.eat()

	for {
		c, err := l.peek()
		if err != nil {
			return nil, err
		} else if !isFirstIdentChar(c) && !isDecimalDigit(c) {
			break
		}

		l.eat()
	}

	var kind int
	if _kind, ok := keywordPatterns[l.tokBuff.String()]; ok {
		kind = _kind
	} else {
		kind = TOK_IDENT
	}

	return l.makeToken(kind), nil
}

// -----------------------------------------------------------------------------

// lexNumericLit lexes a numeric literal (number lit, int lit, or float lit).
func (l *Lexer) lexNumericLit() (*Token, error) {
	l.mark()
	c, _ := l.eat()

	// Determine the base of the literal.
	base := 10
	mustHaveDigit := false
	if c == '0' {
		c, err := l.peek()
		if err != nil {
			return nil, err
		}

		switch c {
		case 'x':
			base = 16
			l.eat()
			mustHaveDigit = true
		case 'o':
			base = 8
			l.eat()
			mustHaveDigit = true
		case 'b':
			base = 2
			l.eat()
		}
	}

	// Integer literal suffixes.
	var unsigned, long bool

	// Floating-point data.
	var isFloat, hasExp, expectNeg bool

numLexLoop:
	for {
		c, err := l.peek()
		if err != nil {
			return nil, err
		} else if c == -1 {
			break
		} else if c == '_' {
			// Skip all _ that occur in the literal.
			l.skip()
			continue
		}

		// Handle integer suffixes.
		if !isFloat {
			if c == 'u' {
				l.eat()
				unsigned = true

				c, err = l.peek()
				if err != nil {
					return nil, err
				}

				if c == 'l' {
					l.eat()
					long = true
				}

				break
			} else if c == 'l' {
				l.eat()
				long = true

				c, err = l.peek()
				if err != nil {
					return nil, err
				}

				if c == 'u' {
					l.eat()
					unsigned = true
				}

				break
			}
		}

		// Handle base specific logic.
		switch base {
		case 2: // Base 2 can only be number or int: no float logic.
			if c == '0' || c == '1' {
				l.eat()
			} else {
				break numLexLoop
			}
		case 8: // Base 8 can only be a number or int: no float logic.
			if '0' <= c && c <= '7' {
				l.eat()
			} else {
				break numLexLoop
			}
		case 10: // Base 10 can be number, int or float
			switch c {
			case '.':
				if mustHaveDigit || isFloat {
					break numLexLoop
				}

				l.eat()

				isFloat = true
				mustHaveDigit = true
				continue
			case 'e', 'E':
				if mustHaveDigit || hasExp {
					break numLexLoop
				}

				l.eat()

				isFloat = true
				hasExp = true
				expectNeg = true
				mustHaveDigit = true
				continue
			case '-':
				if mustHaveDigit || !expectNeg {
					break numLexLoop
				}

				l.eat()

				expectNeg = false
				continue
			default:
				if isDecimalDigit(c) {
					l.eat()
					expectNeg = false
				} else {
					break numLexLoop
				}
			}
		case 16: // Base 16 can be number, int, or float.
			switch c {
			case '.':
				if mustHaveDigit || isFloat {
					break numLexLoop
				}

				l.eat()

				isFloat = true
				mustHaveDigit = true
				continue
			case 'p', 'P':
				if mustHaveDigit || hasExp {
					break numLexLoop
				}

				l.eat()

				isFloat = true
				hasExp = true
				expectNeg = true
				mustHaveDigit = true
				continue
			case '-':
				if mustHaveDigit || !expectNeg {
					break numLexLoop
				}

				l.eat()

				expectNeg = false
				continue
			default:
				if isHexDigit(c) {
					l.eat()
					expectNeg = false
				} else {
					break numLexLoop
				}
			}
		}

		// Indicate that a value was received.
		mustHaveDigit = false
	}

	// Ensure that the literal is not malformed.
	if mustHaveDigit {
		return nil, report.Raise(l.getSpan(), "incomplete numeric literal")
	}

	// Determine the type of literal to return.
	kind := TOK_NUMLIT
	if isFloat {
		kind = TOK_FLOATLIT
	} else if unsigned || long {
		kind = TOK_INTLIT
	}

	return l.makeToken(kind), nil
}

// -----------------------------------------------------------------------------

// lexRawStringLit lexes a raw/multiline string literal.
func (l *Lexer) lexRawStringLit() (*Token, error) {
	l.mark()
	l.skip()

	for {
		c, err := l.peek()
		if err != nil {
			return nil, err
		}

		switch c {
		case -1:
			return nil, report.Raise(l.getSpan(), "unclosed string literal")
		case '`':
			l.skip()
			return l.makeToken(TOK_STRINGLIT), nil
		case '\\':
			l.skip()

			{
				c, err = l.peek()
				if err != nil {
					return nil, err
				}

				switch c {
				case -1:
					return nil, report.Raise(l.getSpan(), "unclosed string literal")
				case '`', '\\':
					l.eat()
				default:
					l.tokBuff.WriteRune('\\')
					l.eat()
				}
			}
		default:
			l.eat()
		}
	}
}

// lexStdStringLit lexes a standard string literal.
func (l *Lexer) lexStdStringLit() (*Token, error) {
	l.mark()
	l.skip()

	for {
		c, err := l.peek()
		if err != nil {
			return nil, err
		}

		switch c {
		case -1:
			return nil, report.Raise(l.getSpan(), "unclosed string literal")
		case '"':
			l.skip()
			return l.makeToken(TOK_STRINGLIT), nil
		case '\\':
			l.eat()
			if err = l.eatEscapeSequence(); err != nil {
				return nil, err
			}
		case '\n':
			return nil, report.Raise(l.getSpan(), "standard string cannot contain a newline")
		default:
			l.eat()
		}
	}
}

// lexRuneLit lexes a rune literal.
func (l *Lexer) lexRuneLit() (*Token, error) {
	l.mark()
	l.skip()

	c, err := l.eat()
	if err != nil {
		return nil, err
	}

	switch c {
	case -1:
		return nil, report.Raise(l.getSpan(), "unclosed rune literal")
	case '\'':
		return nil, report.Raise(l.getSpan(), "empty rune literal")
	case '\n':
		return nil, report.Raise(l.getSpan(), "rune cannot contain a newline")
	case '\\':
		if err = l.eatEscapeSequence(); err != nil {
			return nil, err
		}
	}

	c, err = l.skip()
	if err != nil {
		return nil, err
	} else if c == -1 {
		return nil, report.Raise(l.getSpan(), "unclosed rune literal")
	} else if c != '\'' {
		return nil, report.Raise(l.getSpan(), "rune literal cannot contain multiple characters")
	}

	return l.makeToken(TOK_RUNELIT), nil
}

// eatEscapeSequence attempts to consume an escape sequence.  This assumes the
// leading `\` has already been consumed.
func (l *Lexer) eatEscapeSequence() error {
	c, err := l.eat()
	if err != nil {
		return err
	}

	eatUnicodeEscapeSequence := func(n int) error {
		for i := 0; i < n; i++ {
			c, err := l.eat()
			if err != nil {
				return err
			} else if c == -1 {
				return report.Raise(l.getSpan(), "expected %d digit hexadecimal value not end of file", n)
			} else if !isHexDigit(c) {
				return report.Raise(l.getSpan(), "unicode escape code may be comprised of hexadecimal digits only")
			}
		}

		return nil
	}

	switch c {
	case -1:
		return report.Raise(l.getSpan(), "expected escape sequence not end of file")
	case 'a', 'b', 'f', 'n', 'r', 't', 'v', '0', '\'', '\\', '"':
		return nil
	case 'x':
		return eatUnicodeEscapeSequence(2)
	case 'u':
		return eatUnicodeEscapeSequence(4)
	case 'U':
		return eatUnicodeEscapeSequence(8)
	default:
		return report.Raise(l.getSpan(), "unknown escape sequence: `\\%s`", c)
	}
}

// -----------------------------------------------------------------------------

// lexCommentOrDiv lexes a comment or a division token.
func (l *Lexer) lexCommentOrDiv() (*Token, error) {
	l.mark()
	l.skip()

	c, err := l.peek()
	if err != nil {
		return nil, err
	}

	switch c {
	case '/':
		for ; err == nil && c != '\n' && c != -1; c, err = l.skip() {
		}
	case '*':
		for {
			c, err = l.skip()
			if err != nil || c == -1 {
				break
			}

			if c == '*' {
				c, err = l.skip()
				if err != nil || c == -1 || c == '/' {
					break
				}
			}
		}
	default:
		{
			tok := l.makeToken(TOK_DIV)
			tok.Value = "/"
			return tok, nil
		}
	}

	return nil, err
}

// -----------------------------------------------------------------------------

// mark sets the lexer's stored start line and column to its current position.
func (l *Lexer) mark() {
	l.startLine = l.line
	l.startCol = l.col
}

// makeToken produces a new token of the given kind from the lexer's state and
// resets the lexer to begin building the next token.
func (l *Lexer) makeToken(kind int) *Token {
	value := l.tokBuff.String()
	l.tokBuff.Reset()

	return &Token{
		Kind:  kind,
		Value: value,
		Span:  l.getSpan(),
	}
}

// getSpan calculates a text span based on the lexer's current state.
func (l *Lexer) getSpan() *report.TextSpan {
	return &report.TextSpan{
		StartLine: l.startLine,
		StartCol:  l.startCol,
		EndLine:   l.line,
		EndCol:    l.col,
	}
}

// -----------------------------------------------------------------------------

// eat moves the lexer forward one rune and writes the rune to the token buffer.
// If the lexer encounters an EOF, -1 is returned as the rune value.
func (l *Lexer) eat() (rune, error) {
	c, _, err := l.file.ReadRune()
	if err != nil {
		if err == io.EOF {
			return -1, nil
		}

		return 0, err
	}

	l.updatePos(c)
	l.tokBuff.WriteRune(c)

	return c, nil
}

// skip moves the lexer forward one rune but does not write the rune to the
// token buffer.  If the lexer encounters an EOF, -1 is returned as the rune
// value.
func (l *Lexer) skip() (rune, error) {
	c, _, err := l.file.ReadRune()
	if err != nil {
		if err == io.EOF {
			return -1, nil
		}

		return 0, err
	}

	l.updatePos(c)

	return c, nil
}

// peek returns the next rune in the file without moving the lexer forward or
// writing the rune to the token buffer.  If the lexer encounters an EOF, -1 is
// returned as rune value.
func (l *Lexer) peek() (rune, error) {
	c, _, err := l.file.ReadRune()
	if err != nil {
		if err == io.EOF {
			return -1, nil
		}

		return 0, err
	}

	if err = l.file.UnreadRune(); err != nil {
		return 0, err
	}

	return c, nil
}

// updatePos updates the lexer's position based on input character.
func (l *Lexer) updatePos(c rune) {
	switch c {
	case '\n':
		l.line++
		l.col = 0
	case '\t':
		l.col += 4
	default:
		l.col++
	}
}

// -----------------------------------------------------------------------------

// isDecimalDigit returns whether c is a decimal digit.
func isDecimalDigit(c rune) bool {
	return '0' <= c && c <= '9'
}

// isHexDigit returns whether  c is a hexadecimal digit.
func isHexDigit(c rune) bool {
	return isDecimalDigit(c) || 'a' <= c && c <= 'f' || 'A' <= c && c <= 'F'
}

// isFirstIdentChar returns whether c could be the first rune of an identifier.
func isFirstIdentChar(c rune) bool {
	return unicode.IsLetter(c) || c == '_'
}
