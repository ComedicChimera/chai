package syntax

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"chai/logging"
)

// simple scanner/parser to read in and create grammar
type gramLoader struct {
	file    *bufio.Reader
	grammar Grammar
	curr    rune
	line    uint
}

// load the grammar using a gramLoader and return whether or not loading was
// successful (fails if grammar is syntactically invalid)
func loadGrammar(path string) (Grammar, error) {
	// open the file and check for errors
	f, err := os.Open(path)

	if err != nil {
		return nil, err
	}

	// create a gramLoader using bufio.Reader to load the file
	gl := &gramLoader{file: bufio.NewReader(f), grammar: make(map[string]Production), line: 1}

	// check for gramLoader errors
	err = gl.load()

	if err != nil {
		return nil, err
	}

	// return the loaded grammar if no errors
	return gl.grammar, nil
}

// load the grammar into a grammar struct
func (gl *gramLoader) load() error {
	for gl.next() {
		switch gl.curr {
		// skip whitespace and byte order marks, lines counted in next()
		case ' ', '\t', '\n', '\r', 65279:
			// break
		// read comments (starting with `(*`)
		case '(':
			if b, berr := gl.peek(); berr == nil && b == '*' {
				gl.skipComment()
				break
			}

			fallthrough
		// outer loading algorithm only expects whitespace, comments and
		// productions, so we only look to see if we have a valid beginning to a
		// production here instead of checking more generally
		default:
			if IsLetter(gl.curr) {
				perr := gl.readProduction()

				if perr != nil {
					return perr
				}
			} else {
				return gl.unexpectedToken()
			}
		}

	}

	return nil
}

// read a rune from the stream and store it
func (gl *gramLoader) next() bool {
	r, _, err := gl.file.ReadRune()

	if err != nil {
		if err == io.EOF {
			return false
		}

		logging.LogFatal(err.Error())
	}

	if r == '\n' {
		gl.line++
	}

	gl.curr = r
	return true
}

// peek and convert to rune if successful, return error if not (rune is 0 then)
func (gl *gramLoader) peek() (rune, error) {
	b, berr := gl.file.Peek(1)

	if berr != nil {
		return 0, berr
	}

	return rune(b[0]), nil
}

// read a comment to conclusion
func (gl *gramLoader) skipComment() {
	// skip opening '*'
	gl.next()

	for gl.next() {
		if gl.curr == '*' {
			ahead, err := gl.peek()

			if err == nil && ahead == ')' {
				gl.next()
				return
			}
		}
	}
}

// returns an unexpected token error
func (gl *gramLoader) unexpectedToken() error {
	return fmt.Errorf("unexpected token `%c` on line %d", gl.curr, gl.line)
}

// load and parse a production
func (gl *gramLoader) readProduction() error {
	prodNameBuilder := strings.Builder{}

	// know first character is valid so use "do-while" pattern here
	for ok := true; ok; ok = gl.next() {
		// accepts letters and underscores in production name and adds them to
		// the production name builder
		if IsLetter(gl.curr) || gl.curr == '_' {
			prodNameBuilder.WriteRune(gl.curr)
			// if we encounter a '=' we have reached end of production name so
			// we begin parsing content
		} else if gl.curr == '=' {
			// production is just group ending in ';'
			gelems, err := gl.parseGroupContent(';')

			// if there was a production error, return it
			if err != nil {
				return err
			}

			// otherwise, assume everything is fine, add production to grammar
			// and return
			gl.grammar[prodNameBuilder.String()] = gelems
			return nil

		} else if gl.curr == ' ' {
			// ignore trailing spaces (written for syntactic appeal)
			continue
		} else {
			// all other runes here are invalid, so we mark as unexpected
			return gl.unexpectedToken()
		}
	}

	// if we reach here, the loop did not exit properly (ran out of tokens)
	return errors.New("unexpected EOF")
}

// parse a group or production to a closer
func (gl *gramLoader) parseGroupContent(expectedCloser rune) ([]GrammaticalElement, error) {
	var groupContent []GrammaticalElement

	for gl.next() {
		switch gl.curr {
		case ' ', '\t', '\n', '\r':
			// ignore whitespace
			continue
		// for all grouping elements, parse their content up until the provided closer
		// and the collect the elements into a group of the desired kind and push it
		// onto the groupContent stack (doesn't really make sense as function)
		case '(':
			gelems, err := gl.parseGroupContent(')')

			if err != nil {
				return nil, err
			}

			groupContent = append(groupContent, NewGroupingElement(GKindGroup, gelems))
		case '[':
			gelems, err := gl.parseGroupContent(']')

			if err != nil {
				return nil, err
			}

			groupContent = append(groupContent, NewGroupingElement(GKindOptional, gelems))
		case '{':
			gelems, err := gl.parseGroupContent('}')

			if err != nil {
				return nil, err
			}

			groupContent = append(groupContent, NewGroupingElement(GKindRepeat, gelems))
		// alternators interrupt the current parsing group and create a new one
		// to the same closer so that they can combine the tailing elements with
		// the elements before them. if the tail is itself an alternator, then
		// we combine them into a single alternator over multiple values, if it
		// is not, then we simply combine the two sets of group elements across
		// a single alternator.  because alternators interrupt parsing and
		// bubble upward, the above-described behavior can happen (alternators
		// will only ever return as the only element in their group so if a
		// group starts with an alternator, we know that is all it is)
		case '|':
			tailContent, err := gl.parseGroupContent(expectedCloser)

			if err != nil {
				return nil, err
			}

			if tailContent[0].Kind() == GKindAlternator {
				alternator := tailContent[0].(*AlternatorElement)
				alternator.PushFront(groupContent)

				if len(groupContent) == 0 {
					return nil, fmt.Errorf("unable to empty alternator branch on line %d", gl.line)
				}

				return []GrammaticalElement{alternator}, nil
			}

			return []GrammaticalElement{NewAlternatorElement(groupContent, tailContent)}, nil
		case '\'':
			terminal, err := gl.readTerminal()

			if err != nil {
				return nil, err
			}

			groupContent = append(groupContent, Terminal(terminal))
		case expectedCloser:
			// if we encounter an empty production or group than we cannot close
			// on it
			if len(groupContent) == 0 {
				return nil, fmt.Errorf("unable to allow empty grammatical group on line %d", gl.line)
			}

			return groupContent, nil
		// the suite group check is implemented here since the closer is the
		// same as the opener and we want to prefer matching the closer.
		// Otherwise, it has the same logic as the other kinds of groups.
		case '?':
			gelems, err := gl.parseGroupContent('?')

			if err != nil {
				return nil, err
			}

			groupContent = append(groupContent, NewGroupingElement(GKindSuite, gelems))
		default:
			// nonterminals only contain letters and underscores
			if IsLetter(gl.curr) || gl.curr == '_' {
				nonTerminal := gl.readNonterminal()

				groupContent = append(groupContent, Nonterminal(nonTerminal))
				// if nothing else matched, then we have an unexpected token
				// (some kind of rogue particle or perhaps the residue of a
				// malformed production or group
			} else {
				// fmt.Println("Expected Closer: " + string(expectedCloser))
				return nil, gl.unexpectedToken()
			}
		}
	}

	return nil, errors.New("grammatical group not closed before EOF")
}

func (gl *gramLoader) readTerminal() (int, error) {
	// ignore the leading `'` in our terminal (start with empty builder)
	terminalBuilder := strings.Builder{}

	for gl.next() {
		// the ending `'` is skipped implicitly (never included in token,
		// dropped in next loop cycle)
		if gl.curr == '\'' {
			// if the terminal is blank, then we have an epsilon
			if terminalBuilder.Len() == 0 {
				return -1, nil
			}

			terminal := terminalBuilder.String()

			// match terminal to a keyword or symbol pattern if possible.
			if kind, ok := keywordPatterns[terminal]; ok {
				return kind, nil
			} else if kind, ok := symbolPatterns[terminal]; ok {
				return kind, nil
			} else {
				var kind int

				// no patterns match => must be a special token
				// (wouldn't a match expression be nice here ;) )
				switch terminal {
				case "IDENTIFIER":
					kind = IDENTIFIER
				case "NEWLINE":
					kind = NEWLINE
				case "STRINGLIT":
					kind = STRINGLIT
				case "BOOLLIT":
					kind = BOOLLIT
				case "INTLIT":
					kind = INTLIT
				case "FLOATLIT":
					kind = FLOATLIT
				case "RUNELIT":
					kind = RUNELIT
				case "INDENT":
					kind = INDENT
				case "DEDENT":
					kind = DEDENT
				default:
					// unknown terminal kind => undefined (0 is
					// meaningless/arbitrary - not token kind)
					return 0, fmt.Errorf("undefined terminal `%s` on line %d", terminalBuilder.String(), gl.line)
				}

				return kind, nil
			}
		}

		terminalBuilder.WriteRune(gl.curr)
	}

	// if the token is not closed before EOF, then it is malformed (and we reach
	// here - value returned is meaningless even though 0 is a token kind value)
	return 0, errors.New("unexpected EOF")
}

func (gl *gramLoader) readNonterminal() string {
	nonterminalBuilder := strings.Builder{}

	// to read a nonterminal, we assume the current character is valid
	// (guaranteed be caller or loop logic) and add it to the nonterminal. Then,
	// we peek the next character: if it is valid, we continue looping. If it is
	// not, we exit and avoid adding it.
	for {
		nonterminalBuilder.WriteRune(gl.curr)

		c, err := gl.peek()
		if err == nil && (IsLetter(c) || c == '_') {
			gl.next()
		} else {
			break
		}
	}

	return nonterminalBuilder.String()
}
