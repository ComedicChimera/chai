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

// -----------------------------------------------------------------------------

// IfExpr represents an entire "if tree" expression with all its branches.
type IfExpr struct {
	ExprBase

	CondBranches []CondBranch
	ElseBranch   Expr

	Pos *report.TextPosition
}

func (ie *IfExpr) Position() *report.TextPosition {
	return ie.Pos
}

// CondBranch represents a single conditional if/elif branch in an if tree.
type CondBranch struct {
	HeaderVarDecl *VarDecl
	Cond, Body    Expr
}

// -----------------------------------------------------------------------------

// WhileExpr represents a while loop with all its trailing blocks included.
type WhileExpr struct {
	ExprBase

	HeaderVarDecl *VarDecl
	Cond, Body    Expr

	// TODO: after blocks

	Pos *report.TextPosition
}

func (we *WhileExpr) Position() *report.TextPosition {
	return we.Pos
}
