package syntax

import (
	"bufio"
	"chai/report"
	"fmt"
	"io"
	"strings"
)

// Lexer lexes a source file into tokens.
type Lexer struct {
	ctx *report.CompilationContext

	file *bufio.Reader

	line, col int

	// startLine and startCol are used to make where tokens begin
	// for purposes of calculating their text position.
	startLine, startCol int

	// tokBuff is where the token contents are stored until they are built.
	tokBuff strings.Builder
}

// NewLexer creates a new lexer for a given file passed in as a reader.
func NewLexer(ctx *report.CompilationContext, file *bufio.Reader) *Lexer {
	return &Lexer{
		ctx:       ctx,
		file:      file,
		line:      1,
		col:       1,
		startLine: 1,
		startCol:  1,
	}
}

// NextToken retrieves a single token from the lexer.
func (l *Lexer) NextToken() (*Token, bool) {
	for ahead, ok := l.peek(); ok; ahead, ok = l.peek() {
		switch ahead {
		// ignore non-meaningful characters (eg. BOM, tabs, spaces, etc.)
		case '\f', '\v', ' ', '\t', '\r', 65279:
			l.skip()
		// handle newlines
		case '\n':
			l.mark()
			l.read()
			return l.makeToken(NEWLINE), true
		// handle split-join
		case '\\':
			l.mark()
			l.skip()
			if ahead, ok := l.peek(); ok {
				if ahead == '\n' {
					l.skip()
				} else {
					l.fail("expected newline after split-join")
					return nil, false
				}
			} else {
				l.fail("unexpected end of file")
				return nil, false
			}
		// comments
		case '#':
			if tok := l.skipComment(); tok != nil {
				return tok, true
			}
		// handle string-like
		case '"':
			return l.lexStdString()
		case '`':
			return l.lexRawString()
		case '\'':
			return l.lexRune()
		default:
			l.mark()

			// numeric literals
			if isDigit(ahead) {
				return l.lexNumber()
			} else if isAlpha(ahead) || ahead == '_' /* identifiers and keywords */ {
				l.read()

				// greedily read in as large an identifier as possible
				for ahead, ok := l.peek(); ok; ahead, ok = l.peek() {
					if isAlpha(ahead) || isDigit(ahead) || ahead == '_' {
						l.read()
					} else {
						break
					}
				}

				// check for keywords
				if keyword, ok := keywordPatterns[l.tokBuff.String()]; ok {
					return l.makeToken(keyword), true
				}

				return l.makeToken(IDENTIFIER), true
			} else /* punctuation and operators */ {
				l.read()

				op, ok := symbolPatterns[l.tokBuff.String()]
				if !ok {
					l.fail(fmt.Sprintf("unknown rune `%c`", ahead))
					return nil, false
				}

				for ahead, ok := l.peek(); ok; ahead, ok = l.peek() {
					if nextOp, ok := symbolPatterns[l.tokBuff.String()+string(ahead)]; ok {
						l.read()
						op = nextOp
					} else {
						break
					}
				}

				return l.makeToken(op), true
			}
		}
	}

	// if we reach here, then there are no more tokens to be consumed: we stop
	// lexing and return an EOF.
	return &Token{Kind: EOF}, true
}

// -----------------------------------------------------------------------------

// skipComment skips a line or block comment.  It optionally returns a token to
// return (if the comment counts as a newline).
func (l *Lexer) skipComment() *Token {
	l.skip()

	ahead, ok := l.peek()
	if !ok {
		return nil
	}

	// multi-line comment
	if ahead == '{' {
		for {
			l.skip()

			ahead, ok = l.peek()
			if !ok {
				break
			}

			// check for the ending sequence
			if ahead == '}' {
				l.skip()

				ahead, ok = l.peek()
				if !ok {
					break
				}

				if ahead == '#' {
					l.skip()
					break
				}
			}
		}
	} else /* single-line comment */ {
		for ; ok && ahead != '\n'; ahead, ok = l.peek() {
			l.skip()
		}

		if !ok {
			return nil
		}

		// emit the newline at the end of the line comment
		l.mark()
		l.read()
		return l.makeToken(NEWLINE)
	}

	return nil
}

// lexStdString lexes a standard (double-quoted) string literal.
func (l *Lexer) lexStdString() (*Token, bool) {
	// skip leading `"` -- we don't want it in the value of token, only in the
	// position.
	l.mark()
	l.skip()

	// read in the string
	for next, ok := l.peek(); ok; next, ok = l.peek() {
		switch next {
		case '"':
			l.skip()
			return l.makeToken(STRINGLIT), true
		case '\\':
			if !l.lexEscapeSequence() {
				return nil, false
			}
		case '\n':
			l.fail("standard string literals can't contain newlines")
			return nil, false
		default:
			l.read()
		}
	}

	// string ended unexpectedly
	l.fail("unexpected end of file before closing double quote")
	return nil, false
}

// lexRawString lexes a raw (backtick enclosed, multiline) string literal.
func (l *Lexer) lexRawString() (*Token, bool) {
	// skip the leading backtick
	l.mark()
	l.skip()

	// read in the string
	for next, ok := l.peek(); ok; next, ok = l.peek() {
		switch next {
		case '`':
			l.skip()
			return l.makeToken(STRINGLIT), true
		case '\\':
			l.skip()
			if next2, ok := l.peek(); ok {
				// only backticks can be escaped in raw strings
				if next2 != '`' {
					l.tokBuff.WriteRune('\\')
				}

				l.read()
			}
		default:
			l.read()
		}
	}

	l.fail("unexpected end of file before closing backtick")
	return nil, false
}

// lexRune lexes a rune literal.
func (l *Lexer) lexRune() (*Token, bool) {
	l.mark()
	l.skip()

	c, ok := l.peek()
	if !ok {
		l.fail("unexpected end of file in rune literal")
		return nil, false
	}

	switch c {
	case '\\':
		if !l.lexEscapeSequence() {
			return nil, false
		}
	case '\'':
		l.fail("empty rune literal")
		return nil, false
	default:
		l.read()
	}

	if closer, ok := l.peek(); ok {
		if closer != '\'' {
			l.fail(fmt.Sprintf("expected `'` not `%c`", closer))
			return nil, false
		}

		l.skip()
		return l.makeToken(RUNELIT), true
	} else {
		l.fail("unexpected end of file before closing single quote")
		return nil, false
	}
}

// lexEscapeSequence reads in an escape sequence in a string or rune.
func (l *Lexer) lexEscapeSequence() bool {
	// skip the leading `\`
	l.skip()

	// check to see if the code exists
	code, ok := l.peek()
	if !ok {
		l.fail("expected escape sequence not end of file")
		return false
	}

	// match it to the known codes and write the encoded character to the string
	switch code {
	case 'a':
		l.tokBuff.WriteRune('\a')
	case 'b':
		l.tokBuff.WriteRune('\b')
	case 'f':
		l.tokBuff.WriteRune('\f')
	case 'n':
		l.tokBuff.WriteRune('\n')
	case 'r':
		l.tokBuff.WriteRune('\r')
	case 't':
		l.tokBuff.WriteRune('\t')
	case 'v':
		l.tokBuff.WriteRune('\v')
	case '"', '\'':
		l.tokBuff.WriteRune(code)
	case '0':
		l.tokBuff.WriteRune(0)
	default:
		l.fail(fmt.Sprintf("unknown escape code: `%c`", code))
		return false
	}

	// skip the actual character
	l.skip()

	return true
}

// lexNumber lexes a number literals -- assumes mark has already been set.
func (l *Lexer) lexNumber() (*Token, bool) {
	var isFloat, isUns, isLong bool
	base := 10

	// we know there is a token ahead
	curr, _ := l.read()

	// handle base prefixes if we have a zero
	if curr == '0' {
		ahead, ok := l.peek()
		if ok {
			switch ahead {
			case 'x':
				base = 16
				l.read()
			case 'b':
				base = 2
				l.read()
			case 'o':
				base = 8
				l.read()
			}
		}
	}

	// keep reading while there are valid runes to consume; number literals just
	// stop as soon as they are unsatisfied.
	hasExp := false
	mustHaveDigit := false // only applies in base 10
loop:
	for ahead, ok := l.peek(); ok; ahead, ok = l.peek() {
		// for all bases that are not base 10, the only special handling is for
		// integer suffixes -- so if we get a valid digit for any of the bases,
		// we just continue.
		switch base {
		case 2:
			if ahead == '1' || ahead == '0' {
				l.read()
				continue
			} else {
				break loop
			}
		case 8:
			if '0' <= ahead && ahead < '8' {
				l.read()
				continue
			} else {
				break loop
			}
		case 16:
			if isDigit(ahead) || 'A' <= ahead && ahead <= 'F' || 'a' <= ahead && ahead <= 'f' {
				l.read()
				continue
			} else {
				break loop
			}
		case 10:
			if isDigit(ahead) {
				l.read()
				mustHaveDigit = false
				continue
			} else if mustHaveDigit {
				l.fail(fmt.Sprintf("expected digit not `%c`", ahead))
				return nil, false
			}

			// for float logic, we make the assumption that things like `e` that
			// immediately follow numeric literals are meant to be part of the
			// literal for sake of logical consistency.
			switch ahead {
			case 'e', 'E':
				if hasExp {
					l.fail("float literal cannot have multiple exponents")
					return nil, false
				}

				l.read()
				hasExp = true
				isFloat = true

				// check for negatives
				if ahead, ok := l.peek(); ok {
					if ahead == '-' {
						l.read()
					}

					mustHaveDigit = true
				} else {
					l.fail("expected digit or `-` not end of file")
					return nil, false
				}

				continue
			case '.':
				if isFloat {
					l.fail("float literal cannot have multiple decimals")
					return nil, false
				}

				l.read()
				isFloat = true

				continue
			case '_':
				// allow underscores to separate large numbers: eg. 100_000_000
				l.skip()

				continue
			}
		}

		// handle suffixes
		if !isFloat {
			if ahead == 'u' {
				l.read()
				isUns = true

				if ahead, ok := l.peek(); ok {
					if ahead == 'l' {
						l.read()
						isLong = true
					}
				}
			} else if ahead == 'l' {
				l.read()
				isLong = true

				if ahead, ok := l.peek(); ok {
					if ahead == 'u' {
						l.read()
						isUns = true
					}
				}
			}
		}

		// if we reach here, completely unknown symbol/all of number has been
		// read in
		break
	}

	if mustHaveDigit {
		l.fail("expected digit not end of file")
		return nil, false
	}

	if base != 10 || isUns || isLong {
		return l.makeToken(INTLIT), true
	} else if isFloat {
		return l.makeToken(FLOATLIT), true
	} else {
		// some tokens may not be determinable as integers or floats yet
		return l.makeToken(NUMLIT), true
	}
}

// -----------------------------------------------------------------------------

// makeToken produces a new token of the given kind based on the lexer's current
// state.  This function also clears the token buffer.
func (l *Lexer) makeToken(kind int) *Token {
	defer l.tokBuff.Reset()
	return &Token{
		Kind:  kind,
		Value: l.tokBuff.String(),
		Position: &report.TextPosition{
			StartLn:  l.startLine,
			StartCol: l.startCol,
			EndLn:    l.line,
			EndCol:   l.col,
		},
	}
}

// mark marks the start of a token being read in (for positioning).
func (l *Lexer) mark() {
	l.startLine = l.line
	l.startCol = l.col
}

// fail reports an error to the user at the position of the current token.
func (l *Lexer) fail(msg string) {
	report.ReportCompileError(
		l.ctx,
		&report.TextPosition{
			StartLn:  l.startLine,
			StartCol: l.startCol,
			EndLn:    l.line,
			EndCol:   l.col,
		},
		msg,
	)
}

// isDigit returns if a given rune is a digit or not.
func isDigit(r rune) bool {
	return '0' <= r && r <= '9'
}

// isAlpha returns if a given rune is a letter or not.
func isAlpha(r rune) bool {
	return 'a' <= r && r <= 'z' || 'A' <= r && r <= 'Z'
}

// -----------------------------------------------------------------------------

// read reads one rune from the input stream into the token buffer and updates
// the Lexer's position if there is a rune to be read.  It returns false if the
// rune was an EOF.
func (l *Lexer) read() (rune, bool) {
	r, _, err := l.file.ReadRune()

	if err != nil {
		if err != io.EOF {
			report.ReportFatal("error reading file `%s`: %s", l.ctx.FileRelPath, err.Error())
		}

		return 0, false
	}

	l.tokBuff.WriteRune(r)
	l.updatePos(r)

	return r, true
}

// skip skips a rune from the input stream and updates the Lexer's position. It
// returns false if the skipped rune was an EOF.
func (l *Lexer) skip() bool {
	r, _, err := l.file.ReadRune()

	if err != nil {
		if err != io.EOF {
			report.ReportFatal("error reading file `%s`: %s", l.ctx.FileRelPath, err.Error())
		}

		return false
	}

	l.updatePos(r)
	return true
}

// peek looks one rune ahead in the input stream without moving the stream or
// the Lexer forward.  It returns `false` if there is an EOF.
func (l *Lexer) peek() (rune, bool) {
	r, _, err := l.file.ReadRune()

	if err != nil {
		if err != io.EOF {
			report.ReportFatal("error reading file `%s`: %s", l.ctx.FileRelPath, err.Error())
		}

		return 0, false
	}

	l.file.UnreadRune()

	return r, true
}

// updatePos updates the Lexer's position based on a rune.
func (l *Lexer) updatePos(r rune) {
	switch r {
	case '\n':
		l.line++
		l.col = 1
	case '\t':
		l.col += 4
	case '\r':
		// carriage returns can't move the lexer forward in column counting
	default:
		l.col++
	}
}
