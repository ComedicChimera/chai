package ast

import (
	"chai/report"
	"chai/typing"
)

// Expr represents an expression simple or complex. All expression nodes
// implement the `Expr` interface.
type Expr interface {
	// Type is the yielded type of the expression.
	Type() typing.DataType

	// SetType sets the type of the expression.
	SetType(typing.DataType)

	// Category is the value category of the expression. It should be one of the
	// enumerated value categories.
	Category() int

	// Position returns the spanning position of the whole expression.
	Position() *report.TextPosition
}

// Enumeration of value categories.
const (
	LValue = iota
	RValue
)

// ExprBase is the base struct for all expressions.
type ExprBase struct {
	typ typing.DataType
	cat int
}

func NewExprBase(typ typing.DataType, cat int) ExprBase {
	return ExprBase{
		typ: typ,
		cat: cat,
	}
}

func (eb *ExprBase) Type() typing.DataType {
	return eb.typ
}

func (eb *ExprBase) SetType(typ typing.DataType) {
	eb.typ = typ
}

func (eb *ExprBase) Category() int {
	return eb.cat
}

// -----------------------------------------------------------------------------

// Oper is an operator used in the AST.
type Oper struct {
	Kind      int
	Name      string
	Pos       *report.TextPosition
	Signature typing.DataType
}

// BinaryOp represents a binary operator application (specifically excluding the
// "ternary" forms of comparison operators).
type BinaryOp struct {
	ExprBase

	Op Oper

	Lhs, Rhs Expr
}

func (bo *BinaryOp) Position() *report.TextPosition {
	return report.TextPositionFromRange(
		bo.Lhs.Position(),
		bo.Rhs.Position(),
	)
}

// MultiComparison is a multi operand comparison expression such as `a < b < c`
// or `x <= y < z > b`.
type MultiComparison struct {
	ExprBase

	Exprs []Expr

	Ops []Oper
}

func (mc *MultiComparison) Position() *report.TextPosition {
	return report.TextPositionFromRange(
		mc.Exprs[0].Position(),
		mc.Exprs[len(mc.Exprs)-1].Position(),
	)
}

// -----------------------------------------------------------------------------

// Tuple represents an n-tuple of elements.  These tuples can be length 1 in
// which case they are simple compiled as sub-expressions.
type Tuple struct {
	ExprBase

	Exprs []Expr
	Pos   *report.TextPosition
}

func (t *Tuple) Position() *report.TextPosition {
	return t.Pos
}

// -----------------------------------------------------------------------------

// Identifier represents a named value.
type Identifier struct {
	ExprBase

	Name string
	Pos  *report.TextPosition
}

func (id *Identifier) Position() *report.TextPosition {
	return id.Pos
}

// Literal represents a single literal value.
type Literal struct {
	ExprBase

	// Kind should be a token kind.  Note that `()` is considered a literal
	// whose kind is `NOTHING`.
	Kind  int
	Value string
	Pos   *report.TextPosition
}

func (lit *Literal) Position() *report.TextPosition {
	return lit.Pos
}
