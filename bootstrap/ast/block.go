package ast

import "chai/report"

// Block is a collection of statements to be executed in sequence.
type Block struct {
	ExprBase

	Stmts []Expr
}

// Position of block always returns the position of the last statement since
// that is the actual source of any type errors generated relating to this
// block.
func (b *Block) Position() *report.TextPosition {
	return b.Stmts[len(b.Stmts)-1].Position()
}
