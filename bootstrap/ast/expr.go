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

// Cast represents a type cast.  The destination type is stored in the ExprBase.
type Cast struct {
	ExprBase

	Src Expr
	Pos *report.TextPosition
}

func (c *Cast) Position() *report.TextPosition {
	return c.Pos
}

// -----------------------------------------------------------------------------

// Oper is an operator used in the AST.
type Oper struct {
	Kind  int
	Name  string
	PkgID uint64
	Pos   *report.TextPosition

	// Signature must be a regular type since it will be stored in a type
	// variable (signature is inferred based on usage)
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

// UnaryOp represents a unary operator application.
type UnaryOp struct {
	ExprBase

	Operand Expr
	Op      Oper
	Pos     *report.TextPosition
}

func (uop *UnaryOp) Position() *report.TextPosition {
	return uop.Pos
}

// -----------------------------------------------------------------------------

// Indirect is a reference/indirection expression (ie. `&x`).
type Indirect struct {
	ExprBase

	Operand Expr
	// TODO: some way to distinguish different kinds of indirection to backend

	Pos *report.TextPosition
}

func (ind *Indirect) Position() *report.TextPosition {
	return ind.Pos
}

// -----------------------------------------------------------------------------

// Call is a function call expression.
type Call struct {
	ExprBase

	Func Expr
	Args []Expr

	Pos *report.TextPosition
}

func (c *Call) Position() *report.TextPosition {
	return c.Pos
}

// Dot represents a dot expression (x.f)
type Dot struct {
	ExprBase

	// Root represents the root expression from which the field is being
	// accessed. For implicit method accesses and fields, this is simply a
	// values.  For explicit method accesses, this is just an identifer with a
	// type corresponding to the type of the space being accessed.  For
	// packages, this is an identifier with a `nil` data type and a name
	// corresponding to the visible package.
	Root      Expr
	FieldName string
	FieldPos  *report.TextPosition

	// DotKind indicates the kind of dot operation this encodes.  This
	// should be one of the dot kinds enumerated in `typing`.
	DotKind int

	Pos *report.TextPosition
}

func (d *Dot) Position() *report.TextPosition {
	return d.Pos
}

// TupleDot represents a tuple field access (tuple.n)
type TupleDot struct {
	ExprBase

	Tuple     Expr
	FieldN    int
	FieldNPos *report.TextPosition
	Pos       *report.TextPosition
}

func (td *TupleDot) Position() *report.TextPosition {
	return td.Pos
}

// StructInit represents a struct initialization.
type StructInit struct {
	ExprBase

	TypeExpr   Expr
	SpreadInit Expr
	FieldInits map[string]FieldInit
	Pos        *report.TextPosition
}

// FieldInit is a structure field initializer.
type FieldInit struct {
	NamePos *report.TextPosition
	Init    Expr
}

func (si *StructInit) Position() *report.TextPosition {
	return si.Pos
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

	Name       string
	Pos        *report.TextPosition
	Mutability *int
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
