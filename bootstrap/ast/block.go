package ast

import "chai/report"

// Block is a collection of statements to be executed in sequence.
type Block struct {
	ExprBase

	Stmts []Expr
}

func (b *Block) Position() *report.TextPosition {
	return report.TextPositionFromRange(
		b.Stmts[0].Position(),
		b.Stmts[len(b.Stmts)-1].Position(),
	)
}
