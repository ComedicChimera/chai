package syntax

import (
	"fmt"

	"chai/logging"
)

// Parser is a modified LALR(1) parser designed for Chai
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
				// unexpected newlines can be accepted wherever; just consume
				// the next token and keep parsing
				if !p.consume() {
					return nil, false
				}
			// generate descriptive error messages for EOFs
			case EOF:
				logging.LogCompileError(
					p.lctx,
					"unexpected end of file",
					logging.LMKSyntax,
					nil, // EOFs only happen in one place :)
				)
				return nil, false
			default:
				logging.LogCompileError(
					p.lctx,
					fmt.Sprintf("unexpected token: `%s`", p.lookahead.Value),
					logging.LMKSyntax,
					TextPositionOfToken(p.lookahead),
				)
				return nil, false
			}
		}
	}
}

// shift performs a shift operation and returns an error indicating whether or
// not it was able to successfully get the next token (from scanner or self) and
// whether or not the shift operation on the current token is valid (ie.
// indents)
func (p *Parser) shift(state int) bool {
	p.stateStack = append(p.stateStack, state)
	p.semanticStack = append(p.semanticStack, (*ASTLeaf)(p.lookahead))

	return p.consume()
}

// consume reads the next token from the scanner into the lookahead
func (p *Parser) consume() bool {
	tok, ok := p.sc.ReadToken()

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

		// remove all empty trees and whitespace tokens from the new branch --
		// none of those tokens are meaningful for analysis
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
				// end tags and newlines are not meaningful after the
				// parsing stage
				case NEWLINE, END:
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
