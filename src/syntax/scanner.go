package syntax

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"chai/logging"
)

// NewScanner creates a scanner for the given file
func NewScanner(fpath string, lctx *logging.LogContext) (*Scanner, bool) {
	f, err := os.Open(fpath)

	if err != nil {
		logging.LogConfigError("File", "error opening file: "+err.Error())
		return nil, false
	}

	s := &Scanner{fh: f, file: bufio.NewReader(f), fpath: fpath, line: 1, lctx: lctx}
	return s, true
}

// IsLetter tests if a rune is an ASCII character
func IsLetter(r rune) bool {
	return r > '`' && r < '{' || r > '@' && r < '[' // avoid using <= and >= by checking characters on boundaries (same for IsDigit)
}

// IsDigit tests if a rune is an ASCII digit
func IsDigit(r rune) bool {
	return r > '/' && r < ':'
}

// Scanner works like an io.Reader for a file (outputting tokens)
type Scanner struct {
	lctx *logging.LogContext

	fh    *os.File
	file  *bufio.Reader
	fpath string

	line int
	col  int

	tokBuilder strings.Builder

	curr rune

	indentLevel int

	// set after the scanner reads a newline to prompt it to update the
	// indentation level and produce tokens accordingly
	updateIndentLevel bool

	// indentMode is used to store how the user is indenting their code mode 0 =
	// no set mode, mode -1 = tabs, mode n > 1 = number of spaces per indent.
	indentMode int

	// lookahead is a token that was processed while scanning in another token
	// (ie. when an extra DEDENT needs to emitted after a linebreak).  This
	// token is set during one call to ReadToken() and outputted the next.
	lookahead *Token

	// auxLookahead is an auxilliary lookahead used in the situation where a
	// DEDENT and some other token need to be emitted b/c the next line was
	// determined to hold content but its first token was already consumed
	auxLookahead *Token
}

// ReadToken reads a single token from the stream.  True indicates that there is
// a token to be read/processed
func (s *Scanner) ReadToken() (*Token, bool) {
	// check the lookahead before yielding a token
	if next := s.readLookahead(); next != nil {
		return next, true
	}

	for s.readNext() {
		tok := &Token{}
		malformed := false

		switch s.curr {
		//  ignore non-meaningful characters (eg. BOM, form-feeds etc.)
		case '\r', '\f', '\v', 65279:
			s.tokBuilder.Reset()
			continue
		// handle newlines where they are relevant
		case '\n':
			// line counting handled in readNext
			tok = s.getToken(NEWLINE)
			s.tokBuilder.Reset()

			// handle any errors that occur from processing the NEWLINE
			if !s.processNewline() {
				return nil, false
			}

			return tok, true
		// handle space-based indentation (and spaces generally)
		case ' ':
			// if we are expecting some kind of space-based indentation
			// or need to determine the indentation level
			if s.updateIndentLevel && s.indentMode > -1 {
				s.updateIndentLevel = false

				// greedily, read as many spaces as is possible
				for ahead, more := s.peek(); more && ahead == ' '; ahead, more = s.peek() {
					s.readNext()
				}

				// save the space count and clear the tok builder before we
				// apply blank line rule (otherwise it thinks the current
				// content in the tokBuilder is a token it read in)
				spaceCount := s.tokBuilder.Len()
				s.tokBuilder.Reset()

				// check for blank line rule before we process the indentation
				// we just read in -- we want to track the indentation so we can
				// use it if we need it, but if we encounter a blank line then
				// the identation should be completely ignored.  We want to make
				// sure NOT to update the indentation level in this case so the
				// indentation can be processed later
				if newline, ok := s.applyBlankLineRule(); ok {
					if newline != nil {
						return newline, true
					}
				} else {
					return nil, false
				}

				// if we are determining the indentation mode, the number of
				// spaces until the first non-space becomes the indentation mode
				// for the rest of the program (regardless of how many it is)
				// NOTE: see `\t` case for explanation of INDENT and DEDENT
				// token values (ie. why we are casting numbers to strings)
				if s.indentMode == 0 {
					s.indentMode = spaceCount

					// if we are determining indentation mode, then we will be
					// indenting to level 1 (since it is only possible to have
					// one indent on the first measured indentation)
					s.indentLevel = 1

					tok = s.makeToken(INDENT, string(1))
				} else {
					// otherwise, calculate the equivalent indentation based on
					// the known space-based indentation mode
					level := spaceCount / s.indentMode

					// determine by how much the indentation changed and in what
					// direction (directed distance) and update the indentLevel
					// since we longer need its value
					levelDiff := level - s.indentLevel
					s.indentLevel = level

					// the change is negative, the indent level decreased
					if levelDiff < 0 {
						tok = s.makeToken(DEDENT, string(-levelDiff))
					} else if levelDiff > 0 {
						// if it is positive, indent level increased
						tok = s.makeToken(INDENT, string(levelDiff))
					} else if next := s.readLookahead(); next != nil {
						// if there was no change, we need to check to see if
						// the lookahead is populated, if it is, we should
						// return it here (so it is not skipped)
						return next, true
					} else {
						// otherwise, the scanner should continue reading (this
						// indentation did not change the level and so no token
						// should be produced).
						continue
					}
				}

				// no need to check for malformed tokens or clear the tokBuilder
				return tok, true
			} else {
				// discard all non-meaningful spaces
				s.tokBuilder.Reset()
				continue
			}
		// handle tab-based indentation
		case '\t':
			if s.updateIndentLevel && s.indentMode < 1 {
				s.updateIndentLevel = false

				// greedily read all the tabs we can
				for ahead, more := s.peek(); more && ahead == '\t'; ahead, more = s.peek() {
					s.readNext()
				}

				// determine the level and clear the tokBuilder before applying
				// blank rule (otherwise it thinks the current content in the
				// tokBuilder is a token it read in)
				level := s.tokBuilder.Len()
				s.tokBuilder.Reset()

				// preemptively apply blank line rule for tab based to
				// indentation as well (see comment in space-based tabing case
				// for a more thorough explanation)
				if newline, ok := s.applyBlankLineRule(); ok {
					if newline != nil {
						return newline, true
					}
				} else {
					return nil, false
				}

				// if we need to set the indentation mode to TAB, do it but only
				// after we have confirmed the indentation is meaningful
				if s.indentMode == 0 {
					s.indentMode = -1
				}

				// rune value of first character of INDENT and DEDENT tokens
				// indicates by how much the level changed (parser should handle
				// interpreting this -- avoids creating a bunch of useless
				// INDENT and DEDENT tokens and weird control flow in ReadToken)
				levelDiff := level - s.indentLevel
				s.indentLevel = level // update level now that we don't need it

				if levelDiff < 0 {
					tok = s.makeToken(DEDENT, string(-levelDiff))
				} else if levelDiff > 0 {
					tok = s.makeToken(INDENT, string(levelDiff))
				} else if next := s.readLookahead(); next != nil {
					// if there was no change, we need to check to see if the
					// lookahead is populated, if it is, we should return it
					// here (so it is not skipped)
					return next, true
				} else {
					// otherwise, the scanner should continue reading (this
					// indentation did not change the level and so no token
					// should be produced).
					continue
				}

				// no need to check for malformed tokens or clear the tokBuilder
				return tok, true
			} else {
				// drop the lingering tab
				s.tokBuilder.Reset()
				continue
			}
		// handle string-like
		case '"':
			// trim off leading `"`
			s.tokBuilder.Reset()
			tok, malformed = s.readStdStringLiteral()
		case '\'':
			// trim off leading `'`
			s.tokBuilder.Reset()
			tok, malformed = s.readRuneLiteral()
		case '`':
			// trim off leading ```
			s.tokBuilder.Reset()
			tok, malformed = s.readRawStringLiteral()
		// handle comments
		case '#':
			ahead, more := s.peek()

			if !more {
				// some weird error happened -- not just an eof
				return nil, false
			} else if ahead == '!' {
				s.tokBuilder.Reset() // get rid of lingering `#`
				s.skipBlockComment()
				continue
			} else {
				s.tokBuilder.Reset() // get rid of lingering `#`

				// handle any errors that occur from applying the blank line
				// rule (don't need to discard tokBuilder in the event of an
				// error since errors stop parsing for this file)
				if !s.skipLineComment() {
					return nil, false
				}

				// return a newline so that the indentation is measured
				// correctly by the parser (position is still accurate)
				return &Token{Kind: NEWLINE, Value: "\n", Line: s.line, Col: s.col}, true
			}
		// handle the split-join character ('\')
		case '\\':
			// read through any whitespace until a newline is encountered. We
			// can use regular readNext since we want to include the newline.
		loop:
			for s.readNext() {
				switch s.curr {
				// BOM can also be ignored if it occurs here
				case '\r', '\f', '\v', 65279, ' ', '\t':
					continue
				case '\n':
					break loop
				default:
					// if we encounter something that is not a whitespace
					// character before we encounter a newline, the character is
					// invalid and we mark only that character as erroneous
					logging.LogCompileError(
						s.lctx,
						"expecting a line break before next non-whitespace character",
						logging.LMKToken,
						&logging.TextPosition{StartLn: s.line, StartCol: s.col - 1, EndLn: s.line, EndCol: s.col},
					)
					return nil, false
				}
			}

			// if we reach this point, the newline has been consumed and our
			// split-join is valid.  To apply its effects, we simply discard the
			// current builder contents including the newline and continue on as
			// normal.  Since from the scanner's POV, we are not on a separate
			// line (it does count it, but doesn't act on it), it will not be
			// checking for indentation tokens and so the user can indent or
			// dedent as much as they want so long as they return to correct
			// indentation level after the next, non-split-join newline.
			s.tokBuilder.Reset()
			continue
		default:
			// check for identifiers
			if IsLetter(s.curr) || s.curr == '_' {
				tok = s.readWord()
			} else if IsDigit(s.curr) {
				// check numeric literals
				tok, malformed = s.readNumberLiteral()
			} else if kind, ok := symbolPatterns[string(s.curr)]; ok {
				// all compound tokens begin with valid single tokens so the
				// check above will match the start of any symbolic token

				// keep reading as long as our lookahead is valid: avoids
				// reading extra tokens (ie. in place of readNext)
				for ahead, more := s.peek(); more; ahead, more = s.peek() {
					if skind, ok := symbolPatterns[s.tokBuilder.String()+string(ahead)]; ok {
						kind = skind
						s.readNext()
					} else {
						break
					}
				}

				// turn whatever we managed to read into a token
				tok = s.getToken(kind)
			} else {
				// any other token must be malformed in some way
				malformed = true
			}
		}

		// discard the built contents for the current scanned token
		s.tokBuilder.Reset()

		// error out on any malformed tokens (using contents of token builder)
		if malformed {
			logging.LogCompileError(
				s.lctx,
				fmt.Sprintf("malformed token: `%s`", s.tokBuilder.String()),
				logging.LMKToken,
				&logging.TextPosition{StartLn: s.line, StartCol: s.col, EndLn: s.line, EndCol: s.col + s.tokBuilder.Len()},
			)
			return nil, false
		}

		// if we reach here, we do not need to update the indentation (another
		// meaningful token was encountered => no more indentation counting)
		s.updateIndentLevel = false

		return tok, true
	}

	// if we're are at the end of the file, we need to first return a NEWLINE to
	// exit any blocks we may be in or end any statements found at the end of
	// the file.  Then, we return an appropriate DEDENT to get us to the top of
	// the file followed by an EOF token.  We do this by storing the last two
	// tokens in our lookaheads and returning our initial NEWLINE.  This section
	// should only run once.
	s.auxLookahead = &Token{Kind: EOF}
	s.lookahead = &Token{Kind: DEDENT, Value: string(s.indentLevel)}

	return s.makeToken(NEWLINE, ""), true
}

// UnreadToken is used to undo the preprocessor read the occurs at the start of
// every file.  This function should ONLY be called at the start of a file
func (s *Scanner) UnreadToken(tok *Token) {
	// if the lookahead is empty, we can just dump it in there to be read next
	if s.lookahead == nil {
		s.lookahead = tok
	} else {
		// we know that since this is called at the start of the file, the
		// auxilliary lookahead should never populated.  So if there is an item
		// in the lookahead, we can move it to the auxilliary lookahead and
		// store our main token in the primary lookahead
		s.auxLookahead = s.lookahead
		s.lookahead = tok
	}
}

// Context returns the scanner's log context
func (s *Scanner) Context() *logging.LogContext {
	return s.lctx
}

// Close closes the open file handle the scanner is processing
func (s *Scanner) Close() error {
	return s.fh.Close()
}

// create a token at the current position from the provided data
func (s *Scanner) makeToken(kind int, value string) *Token {
	tok := &Token{Kind: kind, Value: value, Line: s.line, Col: s.col}
	return tok
}

// collect the contents of the token builder into a string and create a token at
// the current position with the provided kind and token string as its value
func (s *Scanner) getToken(kind int) *Token {
	tokValue := s.tokBuilder.String()
	return s.makeToken(kind, tokValue)
}

// reads a rune from the file stream into the token builder and returns whether
// or not there are more runes to be read (true = no EOF, false = EOF)
func (s *Scanner) readNext() bool {
	r, _, err := s.file.ReadRune()

	if err != nil {
		if err != io.EOF {
			logging.LogConfigError("File", fmt.Sprintf("error reading file %s: %s", s.fpath, err.Error()))
		}

		return false
	}

	// do line and column counting after the newline token
	// as been processed (so as to avoid positioning errors)
	if s.curr == '\n' {
		s.line++
		s.col = 0
	}

	s.tokBuilder.WriteRune(r)
	s.curr = r

	if r == '\t' {
		// Whirlwind makes the executive decision to count tabs as four spaces
		// (this is only really meaningful for display purposes -- Whirlwind
		// highlights the error)
		s.col += 4
	} else {
		s.col++
	}

	return true
}

// same behavior as readNext but doesn't populate the token builder used for
// comments where it makes sense
func (s *Scanner) skipNext() bool {
	r, _, err := s.file.ReadRune()

	if err != nil {
		if err != io.EOF {
			logging.LogConfigError("File", fmt.Sprintf("error reading file %s: %s", s.fpath, err.Error()))
		}

		return false
	}

	// do line and column counting after the newline token
	// as been processed (so as to avoid positioning errors)
	if s.curr == '\n' {
		s.line++
		s.col = 0
	}

	s.curr = r
	return true
}

// peek a rune ahead on the scanner (used to test for malformed tokens)
func (s *Scanner) peek() (rune, bool) {
	r, _, err := s.file.ReadRune()

	if err != nil {
		return 0, false
	}

	s.file.UnreadRune()

	return r, true
}

// reads an identifier or a keyword from the input stream determines based on
// contents of stream (matches to all possible keywords)
func (s *Scanner) readWord() *Token {
	// if our word starts with an '_', it cannot be a keyword (simple check here)
	keywordValid := s.curr != '_'

	// to read a word, we assume that current character is valid and already in
	// the token builder (guaranteed by caller or previous loop cycle). we then
	// use a look-ahead to check if the next token will be valid. If it is, we
	// continue looping (and the logic outlined above holds). If not, we exit.
	// Additionally, if at any point in the middle of the word, we encounter a
	// digit or an underscore, we know we are not reading a keyword and set the
	// corresponding flag.  This function is never called on words that begin
	// with numbers so no need to check for first-character rules in it.
	for {
		c, more := s.peek()

		if !more {
			break
		} else if IsDigit(c) || c == '_' {
			keywordValid = false
		} else if !IsLetter(c) {
			break
		}

		s.readNext()
	}

	tokValue := s.tokBuilder.String()

	// if a keyword is possible and our current token value matches a keyword
	// pattern, create a new keyword token from the token builder
	if keywordValid {
		if kind, ok := keywordPatterns[tokValue]; ok {
			return s.makeToken(kind, tokValue)
		}
	}

	// otherwise, assume that it is just an identifier and act accordingly
	return s.makeToken(IDENTIFIER, tokValue)
}

// read in a floating point or integral number
func (s *Scanner) readNumberLiteral() (*Token, bool) {
	var isHex, isBin, isOct, isFloat, isUns, isLong bool

	// the first thing to do it to determine what kind of numeric literal we are
	// dealing with.  If our token starts with at `0`, then we peek ahead to see
	// if we need to process an integer literal prefix (ie. a 0x).  If it
	// doesn't, we do nothing and leave the next character to be scanned by the
	// main integer literal scanning loop (s.curr will be our starting char)
	if s.curr == '0' {
		ahead, more := s.peek()

		if more {
			switch ahead {
			case 'x':
				isHex = true
				s.readNext()
			case 'o':
				isOct = true
				s.readNext()
			case 'b':
				isBin = true
				s.readNext()
			}
		} else {
			// if we are out of tokens, then 0 is our only element
			return s.getToken(INTLIT), true
		}
	}

	// if we previous was an 'e' then we can expect a '-'
	expectNeg := false

	// if previous was a `.` then we expect a digit
	expectDigit := false

	// if we triggered a floating point using '.' instead of 'e' than 'e' could
	// still be valid
	eValid := false

	// use loop break label to break out loop from within switch case
loop:
	// We already know the first character is valid so we can simply ignore it.
	// Moreover, we don't need to check readNext since we know the next token
	// will be valid b/c it will be checked with s.peek() is called.  So instead
	// we just want to make sure that all `continue` because the checked
	// character to be read in.
	for ; true; s.readNext() {
		ahead, more := s.peek()

		// if we have nothing left, then we are out of tokens and should exit
		if !more {
			break
		}

		// if we have identified signage or sign, then we are not expecting
		// anymore values and so exit out if an additional values are
		// encountered besides sign and size specifiers
		if isLong && isUns {
			break
		} else if isLong {
			if ahead == 'u' {
				isUns = true
				continue
			} else {
				break
			}
		} else if isUns {
			if ahead == 'l' {
				isLong = true
				continue
			} else {
				break
			}
		}

		// if we are expecting a negative and get another character then we
		// simply update the state (no longer expecting a negative) and continue
		// on (expect is not a hard expectation)
		if expectNeg && ahead != '-' {
			expectNeg = false
		}

		// check to ensure that any binary literals are valid
		if isBin {
			if ahead == '0' || ahead == '1' {
				continue
			} else {
				break
			}
		}

		// check to ensure that any octal literals are valid
		if isOct {
			if ahead > '/' && ahead < '9' {
				continue
			} else {
				break
			}
		}

		if IsDigit(ahead) {
			expectDigit = false
			continue
		}

		// check for validity of hex literal
		if isHex && (ahead < 'A' || ahead > 'F') && (ahead < 'a' || ahead > 'f') {
			break
		} else if isFloat {
			// if we are expecting a digit and not getting one (ie. after the `.`)
			if expectDigit {
				// create the base integer literal
				tokValue := s.tokBuilder.String()
				intlit := s.makeToken(INTLIT, tokValue[:len(tokValue)-1])
				intlit.Col--

				// clear the token builder before we continue to build the dots
				s.tokBuilder.Reset()

				// accumulate `.` into larger tokens if necessary
				for ahead == '.' {
					s.readNext()
					ahead, more = s.peek()

					if !more {
						break
					}
				}

				// make the dots token into the current lookahead
				s.lookahead = s.getToken(symbolPatterns[s.tokBuilder.String()])
				s.tokBuilder.Reset()

				return intlit, false
			}

			// our scanning logic changes after we parse a float: we can longer
			// accept `u`, `l` or `.` (since even if an `e` (or `E`) was what
			// moved us into this state, the power must be an integer value).
			// So we employ alternative checking logic
			switch ahead {
			case 'e', 'E':
				// avoid duplicate e's
				if eValid {
					eValid = false
				} else {
					break loop
				}
			case '-':
				if expectNeg {
					// since we are already looking ahead, we have to assume that
					// the negative is part of this token.  If what comes after this
					// token is not a digit, then we would have something of the form
					// `0e-...`` which can never be valid even if you split the tokens
					// (b/c you get `0e`, `-`, and `...` and the `0e` is invalid)
					s.readNext()

					// check if there is a non-number ahead then we actually
					// have 3 tokens and have to scan the other two separately
					ahead, valid := s.peek()

					// hit EOF on peek, malformed token
					if !valid {
						return nil, true
					}

					// if it is not a digit, the token is malformed
					if !IsDigit(ahead) {
						return nil, true
					}

					expectNeg = false
				} else {
					break loop
				}
			default:
				// if expectNeg is true here, then we know the previous
				// character we read in was an `e` and since a numeric literal
				// cannot end with an `e` and there will never be context where
				// an identifier followed by an integer literal could be two
				// valid tokens (according to the grammar), we mark such
				// literals as one malformed token
				if expectNeg {
					return nil, true
				}

				break loop
			}
		}

		// we only reach this point, we are not a floating point value (and
		// might want to become one or need to act knowing we aren't one)
		switch ahead {
		case '.':
			isFloat = true
			eValid = true
			expectDigit = true
		case 'e', 'E':
			isFloat = true
			expectNeg = true
		case 'u':
			if isFloat {
				break
			}

			isUns = true
		case 'l':
			if isFloat {
				break
			}

			isLong = true
		default:
			break loop
		}
	}

	// binary, octal, decimal, and hexadecimal literals are all considered
	// integer literals and so the only decision here is whether or not to
	// create a floating point literal (use already accumulated information)
	if isFloat {
		return s.getToken(FLOATLIT), false
	}

	return s.getToken(INTLIT), false
}

// read in a standard string literal -- assuming leading `"` has been dropped
func (s *Scanner) readStdStringLiteral() (*Token, bool) {
	expectingEscape := false

	// use a lookahead pattern to avoid reading the closing quote
	for next, ok := s.peek(); ok; next, ok = s.peek() {
		// test for escape first
		if expectingEscape {
			// handle invalid escape sequences -- no need to read next here
			// since our `readEscapeSequence` does that for us
			if s.readEscapeSequence() {
				expectingEscape = false
			} else {
				return nil, true
			}
		}

		switch next {
		case '\\':
			expectingEscape = true
			s.readNext()
		case '"':
			// we don't want to read the closing quote, so we make
			// our string token, skip it (so the column is right)
			// and we return.  We know that this quote is valid since
			// escape sequences are handled above this switch
			tok := s.getToken(STRINGLIT)
			s.skipNext()
			return tok, false
		case '\n':
			// catch newlines in strings
			return nil, true
		default:
			// otherwise, just read in the next character
			s.readNext()
		}
	}

	// if we reach here, we didn't encounter a proper closing quotation and
	// therefore, we need to indicate that this token is malformed
	return nil, true
}

// read in a rune literal -- assuming leading `'` has been dropped
func (s *Scanner) readRuneLiteral() (*Token, bool) {
	// if the rune has no content then it is malformed
	if !s.readNext() {
		return nil, true
	}

	// if there is an escape sequence, read it and if it is invalid, rune lit is
	// malformed
	if s.curr == '\\' && !s.readEscapeSequence() {
		return nil, true
	}

	// if the next token after processing the escape sequence/rune body is not a
	// closing quote than the rune literal is too long on we are at EOF =>
	// malformed in either case
	if next, ok := s.peek(); !ok || next != '\'' {
		return nil, true
	}

	// assume it is properly formed and skip the closing single quote; skip
	// after creating the token so the column is correct
	tok := s.getToken(RUNELIT)
	s.skipNext()
	return tok, false
}

// read and interpret an escape sequence inside a string or rune literal
func (s *Scanner) readEscapeSequence() bool {
	if !s.readNext() {
		return false
	}

	readUnicodeSequence := func(count int) bool {
		for i := 0; i < count; i++ {
			if !s.readNext() {
				return false
			}

			r := s.curr

			if !IsDigit(r) && (r < 'A' || r > 'F') && (r < 'a' || r > 'f') {
				return false
			}
		}

		return true
	}

	switch s.curr {
	case 'a', 'b', 'n', 'f', 'r', 't', 'v', '0', 's', '"', '\'', '\\':
		return true
	case 'x':
		return readUnicodeSequence(2)
	case 'u':
		return readUnicodeSequence(4)
	case 'U':
		return readUnicodeSequence(8)
	}

	return true
}

// read in a raw string literal -- assuming leading backtick has been dropped
func (s *Scanner) readRawStringLiteral() (*Token, bool) {
	// used to signal when we need to escape a backtick
	escapeNext := true

	// use a lookahead pattern to avoid reading the closing backtick
	for next, ok := s.peek(); ok; next, ok = s.peek() {
		if next != '`' || escapeNext {
			s.readNext()
		} else {
			// skip closing backtick; creating stringlit token before skipping
			// to generate correct column position
			tok := s.getToken(STRINGLIT)
			s.skipNext()
			return tok, false
		}

		// check for escapes for backticks; update after we have handled the
		// exit condition
		escapeNext = next == '\\'
	}

	// only stringlits that reach here are incomplete
	return nil, true
}

func (s *Scanner) skipLineComment() bool {
	for s.skipNext() && s.curr != '\n' {
	}

	// make sure the scanner properly handles the newline
	return s.processNewline()
}

func (s *Scanner) skipBlockComment() {
	// skip opening '!'
	s.skipNext()

	for s.skipNext() {
		if s.curr == '!' {
			p, more := s.peek()

			if more && p == '#' {
				s.skipNext()
				return
			}
		}
	}
}

// processNewline performs all necessary scanner logic to handle a newline
func (s *Scanner) processNewline() bool {
	s.updateIndentLevel = true
	var ok bool = true

	emitDedent := func() {
		next, bok := s.applyBlankLineRule()

		if !bok {
			ok = bok
			return
		}

		if next == nil {
			// make sure that the token following the non-blank line is not
			// discarded/lost by using the auxilliary lookahead
			s.auxLookahead = s.lookahead
			s.lookahead = s.makeToken(DEDENT, string(s.indentLevel))
			s.indentLevel = 0
		}

		// since next will only be a NEWLINE, we can ignore it since we are
		// already processing a NEWLINE (if a next exists)
	}

	// check to see if we have an appropriate indent character on the next line
	// (something that will trigger indentation logic).  If not, we need to emit
	// the appropriate DEDENT (if we are exiting to top indentation level).
	// However, we only need to do this, if we are not already at the top level.
	// NOTE: DEDENT emitted on next call.
	if s.indentLevel > 0 {
		ahead, more := s.peek()

		if more {
			// NOTE: in both cases, the DEDENT change is equivalent to the
			// current level if one should be emitted

			// if the mode is not determined then either spaces or tabs will
			// count as an indent and so we check for both
			if s.indentMode == 0 && ahead != ' ' && ahead != '\t' {
				emitDedent()
			} else if s.indentMode == -1 && ahead != '\t' {
				// if we are in TAB mode, check for tabs (above)
				emitDedent()
			} else if s.indentMode > 0 && ahead != ' ' {
				// we are in some SPACE mode, check for spaces (above)
				emitDedent()
			}
		}

		// regardless of any DEDENT emissions, continue as normal
	}

	// want to keep updateIndentLevel flag
	s.tokBuilder.Reset()

	return ok
}

// applyBlankLineRule checks if a given line is blank and if it is, it
// configures the scanner to ignore the content of the current line and returns
// the NEWLINE the scanner should return.  If the line is not blank, it returns
// `nil`.  This should be called before any kind of indentation is produced and
// fed to the parser (not doing so will confuse the parser and cause blank lines
// to be interpreted as syntactically significant).  NOTE: this function does
// override the current lookahead with the token the scanner should return on
// the next consumption.  It will not always, but it should be assumed that it
// will.  It will also return any errors it encounters while looking ahead.
func (s *Scanner) applyBlankLineRule() (*Token, bool) {
	tok, ok := s.ReadToken()

	if !ok {
		return nil, false
	}

	// if the next token is NEWLINE, the line was blank and should be skipped.
	// NOTE: there is a bit of implicit recursion here in that when a NEWLINE is
	// encountered, the blank line rule may be applied.  This is fine as we are
	// ok ignoring line breaks on blank lines: we can simply defer control
	// recursively as necessary.  Ultimately, no harm will be done.
	if tok.Kind == NEWLINE {
		// no need to set explicitly set the lookahead here: if it is needed,
		// (ie. to store an upcoming DEDENT), it will already be set.
		return tok, true
	}

	// if the token was not a line break, then we simply store what we read
	// ahead into the lookahead and return `nil` indicating that we don't want
	// to override the current scanner's return (not a blank line).  NOTE: b/c
	// INDENT and DEDENT never occur directly sequentially, we don't need to
	// worry about them here (all indentation changes are compressed into a
	// single INDENT or DEDENT token for simplicity and efficiency).  We also
	// can freely override the lookahead here as the only time we couldn't would
	// be in the context of a blank line which has already been handled.
	s.lookahead = tok
	return nil, true
}

// readLookahead checks to see if the scanner has a lookahead that it should
// return before processing more tokens.  If it does, it updates the lookahead
// and the auxilliary lookahead and returns the lookahead token.  If it does
// not, this function does nothing and returns `nil`.  This should ONLY be
// called if the scanner is intending to yield the lookahead token.
func (s *Scanner) readLookahead() *Token {
	if s.lookahead != nil {
		lhTok := s.lookahead

		// move the auxilliary lookahead into the current lookahead
		s.lookahead = s.auxLookahead

		// clear the auxilliary lookahead
		s.auxLookahead = nil

		return lhTok
	}

	return nil
}
