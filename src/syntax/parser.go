package syntax

import (
	"fmt"

	"chai/logging"
)

// Parser is a modified LALR(1) parser designed for Whirlwind
type Parser struct {
	// `lctx` is updated every call to `parse`
	lctx *logging.LogContext

	// scanner containing reference to the file being parsed
	sc *Scanner

	// standard LALR(1) parser state
	lookahead  *Token
	ptable     *ParsingTable
	stateStack []int

	// semantic stack is used to build the AST and accumulate tokens as necessary
	semanticStack []ASTNode

	// used to store the parser's stack of indentation frames.  The parser
	// starts with a frame on top of the stack to represent the frame of the
	// whole program.  This frame is indentation aware with an entry level of -1
	// (meaning it will never close => total enclosing frame).
	indentFrames []*IndentFrame
}

// IndentFrame represents an indentation context frame.  An indentation context
// frame is a region of a particular kind of indentation processing: indentation
// blind (eg. inside `[]`) or indentation aware.  They are used to represent the
// contextual state of the parser as it relates to indentation.  As the parser
// switches between the two modes, frames are created and removed.  The parser
// decides when to switch and what to do when it switches based on the mode of
// the indentation frame and the entry level.  If it is in indentation aware
// mode, it exits when it returns to its entry level.  If it is not, it exits
// when all of the unbalanced openers (ie. `(`, `[`, and `{`) have been closed.
// When an indentation blind frame is closed, it updates the scanner's
// indentation level to be its entry level so that all changes to indentation
// inside the ignored block become irrelevant and the parser checks with the
// intended behavior.  A new indentation blind frame is created when an opener
// is encountered in an indentation aware frame.  A new indentation aware frame
// is created when an INDENT token is shifted in an indentation blind frame.
type IndentFrame struct {
	// mode == -1 => indentation aware
	// mode >= 0 => indentation blind, mode = number of unbalanced openers
	Mode int

	// the indentation level that the indentation frame began at
	EntryLevel int
}

// NewParser creates a new parser for the given parsing table and the
// given scannner (file).  Context is extracted from scanner
func NewParser(ptable *ParsingTable, sc *Scanner) *Parser {
	return &Parser{
		ptable: ptable,
		sc:     sc,
		lctx:   sc.Context(),
		// set the state stack to the starting/initial state
		stateStack: []int{0},
		// parser starts with one indentation-aware indent frame that will never close
		indentFrames: []*IndentFrame{{Mode: -1, EntryLevel: -1}},
	}
}

// Parse runs the main parsing algorithm on the given scanner
func (p *Parser) Parse() (ASTNode, bool) {
	// initialize the lookahead
	if !p.consume() {
		return nil, false
	}

	for {
		state := p.ptable.Rows[p.stateStack[len(p.stateStack)-1]]

		if action, ok := state.Actions[p.lookahead.Kind]; ok {
			switch action.Kind {
			case AKShift:
				if !p.shift(action.Operand) {
					return nil, false
				}
			case AKReduce:
				p.reduce(action.Operand)
			case AKAccept:
				return p.semanticStack[0], true
			}
		} else {
			switch p.lookahead.Kind {
			case NEWLINE:
				// unexpected newlines can be accepted whenever
				break
			// generate descriptive error messages for special tokens
			case EOF:
				logging.LogCompileError(
					p.lctx,
					"unexpected end of file",
					logging.LMKSyntax,
					nil, // EOFs only happen in one place :)
				)
				return nil, false
			case INDENT, DEDENT:
				// check indentation changes only if we are not in a blind frame
				if p.topIndentFrame().Mode == -1 {
					// if the scanner's next token is an EOF, then this
					// is actually an unexpected EOF not an unexpected
					// indentation change (caused by closing DEDENT)
					if tok, ok := p.sc.ReadToken(); ok && tok.Kind == EOF {
						// if an EOF would be acceptable at this point instead
						// of an indentation, then we can simply ignore the
						// indentation and end the file (probably just
						// whitespace floating around at the end)
						if _, ok := state.Actions[EOF]; ok {
							p.lookahead = &Token{
								Kind:  EOF,
								Value: "",
								Line:  p.sc.line,
								Col:   p.sc.col,
							}

							continue
						}

						logging.LogCompileError(
							p.lctx,
							"unexpected end of file",
							logging.LMKSyntax,
							nil, // EOFs only happen in one place :)
						)
						return nil, false
					}

					logging.LogCompileError(
						p.lctx,
						"unexpected indentation change",
						logging.LMKSyntax,
						TextPositionOfToken(p.lookahead),
					)
					return nil, false
				}
			// handle lexical edge case where `<<` and `>>` are actually two
			// separate tokens (ie. for generics) - we do this here instead of
			// in the grammar b/c treating them as separate tokens creates an
			// unresolvable Shift/Reduce conflict between shift_expr and
			// comp_expr (cheating with GOTO b/c no inline comparison)
			case LSHIFT, RSHIFT:
				if p.splitShift(state, p.lookahead.Kind) {
					continue
				}

				fallthrough
			default:
				logging.LogCompileError(
					p.lctx,
					fmt.Sprintf("unexpected token: `%s`", p.lookahead.Value),
					logging.LMKSyntax,
					TextPositionOfToken(p.lookahead),
				)
				return nil, false
			}

			// if we reach here, whatever token was errored on can be ignored
			if !p.consume() {
				return nil, false
			}
		}
	}
}

// shift performs a shift operation and returns an error indicating whether or
// not it was able to successfully get the next token (from scanner or self) and
// whether or not the shift operation on the current token is valid (ie. indents)
func (p *Parser) shift(state int) bool {
	p.stateStack = append(p.stateStack, state)
	p.semanticStack = append(p.semanticStack, (*ASTLeaf)(p.lookahead))

	// used in all branches of switch (small operation - low cost)
	topFrame := p.topIndentFrame()

	switch p.lookahead.Kind {
	// handle indentation
	case INDENT, DEDENT:
		// implement frame control rules for indentation-aware frames
		if p.lookahead.Kind == INDENT && topFrame.Mode > -1 {
			p.pushIndentFrame(-1, p.sc.indentLevel-1)
		} else if p.lookahead.Kind == DEDENT && topFrame.Mode == -1 && topFrame.EntryLevel == p.sc.indentLevel {
			p.popIndentFrame()
		}

		levelChange := int(p.lookahead.Value[0])

		// no need to update lookahead if full level change hasn't been handled
		if levelChange > 1 {
			p.lookahead.Value = string(levelChange - 1)
			return true
		}
	// handle indentation blind frame openers
	case LPAREN, LBRACE, LBRACKET:
		if topFrame.Mode == -1 {
			p.pushIndentFrame(1, p.sc.indentLevel)
		} else {
			topFrame.Mode++
		}
	// handle indentation blind frame closers
	case RPAREN, RBRACE, RBRACKET:
		// if there are closers, there must be openers so we can treat the
		// current frame as if we know it is a blind frame (because we do :D)
		if topFrame.Mode == 1 {
			p.popIndentFrame()
		} else {
			topFrame.Mode--
		}
	}

	return p.consume()
}

// consume reads the next token from the scanner into the lookahead REGARDLESS
// OF WHAT IS IN THE LOOKAHEAD (not whitespace aware - should be used as such)
func (p *Parser) consume() bool {
	tok, ok := p.sc.ReadToken()

	fmt.Println(tok.Kind)
	if tok.Kind == NEWLINE {
		fmt.Println("debug")
	}

	if ok {
		p.lookahead = tok
		return true
	} else {
		return false
	}
}

// reduce performs a reduction and simplifies any elements it reduces
func (p *Parser) reduce(ruleRef int) {
	rule := p.ptable.Rules[ruleRef]

	// no need to perform any extra logic if we are reducing an epsilon rule
	if rule.Count == 0 {
		p.semanticStack = append(p.semanticStack, &ASTBranch{Name: rule.Name})
	} else {
		branch := &ASTBranch{Name: rule.Name, Content: make([]ASTNode, rule.Count)}
		copy(branch.Content, p.semanticStack[len(p.semanticStack)-rule.Count:])
		p.semanticStack = p.semanticStack[:len(p.semanticStack)-rule.Count]

		// calculate the added size of inlining all anonymous productions
		anonSize := 0

		// flag indicating whether or not any anonymous productions need to be
		// inlined (can't use anonSize since an anon production could contain a
		// single element which would mean no size change => anonSize = 0 even
		// though there would still be anonymous productions that need inlining)
		containsAnons := false

		// remove all empty trees and whitespace tokens from the new branch
		n := 0
		for _, item := range branch.Content {
			if subbranch, ok := item.(*ASTBranch); ok {
				if len(subbranch.Content) > 0 {
					branch.Content[n] = item
					n++

					// only want to inline anonymous production if they will
					// still exist after all empty trees are removed
					if subbranch.Name[0] == '$' {
						anonSize += len(subbranch.Content) - 1
						containsAnons = true
					}
				}
			} else if leaf, ok := item.(*ASTLeaf); ok {
				switch leaf.Kind {
				case NEWLINE, INDENT, DEDENT:
					continue
				default:
					branch.Content[n] = item
					n++
				}
			}
		}

		branch.Content = branch.Content[:n]

		// If no anonymous productions need to be inlined, slice the branch down
		// to the appropriate length and continue executing.  If anonymous
		// productions do need to inlined, allocate a new slice to store the new
		// branch and copy/inline as necessary (could do some trickery with
		// existing slice but that seems unnecessarily complex /shrug)
		if containsAnons {
			newBranchContent := make([]ASTNode, anonSize+n)
			k := 0

			for _, item := range branch.Content {
				if subbranch, ok := item.(*ASTBranch); ok && subbranch.Name[0] == '$' {
					for _, subItem := range subbranch.Content {
						newBranchContent[k] = subItem
						k++
					}
				} else {
					newBranchContent[k] = item
					k++
				}
			}

			branch.Content = newBranchContent
		}

		p.semanticStack = append(p.semanticStack, branch)
		p.stateStack = p.stateStack[:len(p.stateStack)-rule.Count]
	}

	// goto the next state
	currState := p.ptable.Rows[p.stateStack[len(p.stateStack)-1]]
	p.stateStack = append(p.stateStack, currState.Gotos[rule.Name])
}

// topIndentFrame gets the indent frame at the top of the frame stack (ie. the
// current indent frame)
func (p *Parser) topIndentFrame() *IndentFrame {
	return p.indentFrames[len(p.indentFrames)-1]
}

// pushIndentFrame pushes an indent frame onto the frame stack
func (p *Parser) pushIndentFrame(mode, level int) {
	p.indentFrames = append(p.indentFrames, &IndentFrame{Mode: mode, EntryLevel: level})
}

// popIndentFrame removes an indent frame from the top of the frame stack
// (clears the current indent frame)
func (p *Parser) popIndentFrame() {
	p.indentFrames = p.indentFrames[:len(p.indentFrames)-1]
}

// splitShift splits a binary shift token into two less than or greater than
// tokens where the input is either LSHIFT or RSHIFT (indicating which to split
// from) and attempts to use the first split token as an action for the current
// state. If it successful, the parser state is updated, the split is applied
// and true is returned.  Otherwise, false is returned and no state is changed.
func (p *Parser) splitShift(state *PTableRow, kind int) bool {
	if action, ok := state.Actions[kind]; ok {
		// Inline Comparison?  Go: NO!
		// Using a boolean as an index? Go: CAN'T CAST BOOL TO INT!
		// Me: (╯°□°）╯︵ ┻━┻
		var splitKind int
		if kind == LSHIFT {
			splitKind = LT
		} else {
			splitKind = GT
		}

		// thinking about spread initialization rn...
		// for line 2: &Token{...p.lookahead, Kind=splitKind}... smh
		t1 := &Token{Kind: splitKind, Value: p.lookahead.Value[:1], Line: p.lookahead.Line, Col: p.lookahead.Col - 1}
		t2 := &Token{Kind: splitKind, Value: p.lookahead.Value[:1], Line: p.lookahead.Line, Col: p.lookahead.Col}

		p.stateStack = append(p.stateStack, action.Operand)
		p.semanticStack = append(p.semanticStack, (*ASTLeaf)(t1))

		p.lookahead = t2

		return true
	}

	return false
}
