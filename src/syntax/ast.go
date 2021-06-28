package syntax

import (
	"chai/logging"
)

// ASTNode represents a piece of the Abstract Syntax Tree (AST)
type ASTNode interface {
	// Position should span the entire ASTNode (meaningfully)
	Position() *logging.TextPosition
}

// ASTLeaf is simply a token in the AST (at the end of branch)
type ASTLeaf Token

// Position of a leaf is just the position of the token it contains
func (a *ASTLeaf) Position() *logging.TextPosition {
	return TextPositionOfToken((*Token)(a))
}

// TextPositionOfToken takes in a token and returns its text position
func TextPositionOfToken(tok *Token) *logging.TextPosition {
	return &logging.TextPosition{StartLn: tok.Line, StartCol: tok.Col - len(tok.Value), EndLn: tok.Line, EndCol: tok.Col}
}

// ASTBranch is a named set of leaves and branches
type ASTBranch struct {
	Name    string
	Content []ASTNode
}

// Position of a branch is the starting position of its first node and the
// ending position of its last node (node can be leaf or branch)
func (a *ASTBranch) Position() *logging.TextPosition {
	// Note: empty AST nodes SHOULD never occur, but we check anyway (so if they
	// do, we see the error)
	if len(a.Content) == 0 {
		logging.LogFatal("Unable to take position of empty AST node")
		// if there is just one item in the branch, just return the position of
		// that item
	} else if len(a.Content) == 1 {
		return a.Content[0].Position()
		// otherwise, it is the positions that border the leaves (they occur in
		// order so we can take their position as such)
	} else {
		first, last := a.Content[0].Position(), a.Content[len(a.Content)-1].Position()

		return &logging.TextPosition{StartLn: first.StartLn, StartCol: first.StartCol, EndLn: last.EndLn, EndCol: last.EndCol}
	}

	// unreachable
	return nil
}

// TextPositionOfSpan takes two nodes and returns a text position that spans them
func TextPositionOfSpan(start, end ASTNode) *logging.TextPosition {
	return &logging.TextPosition{
		StartLn:  start.Position().StartLn,
		StartCol: start.Position().StartCol,
		EndLn:    end.Position().EndLn,
		EndCol:   end.Position().EndCol,
	}
}

// BranchAt gets and casts the specified element to an AST branch
func (a *ASTBranch) BranchAt(ndx int) *ASTBranch {
	return a.Content[ndx].(*ASTBranch)
}

// LeafAt gets and casts the specified element to an AST leaf
func (a *ASTBranch) LeafAt(ndx int) *ASTLeaf {
	return a.Content[ndx].(*ASTLeaf)
}

// Len returns the length of the branch's content
func (a *ASTBranch) Len() int {
	return len(a.Content)
}

// Last returns the last element of the branch
func (a *ASTBranch) Last() ASTNode {
	return a.Content[len(a.Content)-1]
}

// LastBranch returns the last element of a branch and casts it
// to a branch (assumes it is one)
func (a *ASTBranch) LastBranch() *ASTBranch {
	return a.Content[len(a.Content)-1].(*ASTBranch)
}
