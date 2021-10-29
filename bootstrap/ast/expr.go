package ast

import (
	"chai/report"
	"chai/typing"
)

// Expr represents an expression simple or complex.
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
