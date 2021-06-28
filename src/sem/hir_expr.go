package sem

import (
	"chai/logging"
	"chai/syntax"
	"chai/typing"
)

// HIRExpr is the parent interface for all HIR (high-level intermediate
// representation) expressions
type HIRExpr interface {
	// Type returns the data type yielded by an expression
	Type() typing.DataType

	// Category returns the value category of the expression.  It must be one of
	// the enumerated categories below.
	Category() int

	// Constant indicates whether or not the expression is mutable
	Constant() bool
}

// Enumeration of value categories
const (
	LValue = iota
	RValue
)

// ExprBase is the base struct for all expressions
type ExprBase struct {
	dt       typing.DataType
	cat      int
	constant bool
}

func NewExprBase(dt typing.DataType, cat int, constant bool) ExprBase {
	return ExprBase{
		dt:       dt,
		cat:      cat,
		constant: constant,
	}
}

func (eb *ExprBase) Type() typing.DataType {
	return eb.dt
}

func (eb *ExprBase) Category() int {
	return eb.cat
}

func (eb *ExprBase) Constant() bool {
	return eb.constant
}

func (eb *ExprBase) SetType(dt typing.DataType) {
	eb.dt = dt
}

// HIRIncomplete represents an AST branch that hasn't been evaluated yet
type HIRIncomplete syntax.ASTBranch

func (hi *HIRIncomplete) Type() typing.DataType {
	return typing.PrimType(typing.PrimKindNothing)
}

func (hi *HIRIncomplete) Category() int {
	return RValue
}

func (hi *HIRIncomplete) Constant() bool {
	return false
}

// -----------------------------------------------------------------------------

// HIRDoBlock represents a do block expression
type HIRDoBlock struct {
	ExprBase

	Statements []HIRExpr
}

// -----------------------------------------------------------------------------

// HIRApply represents a function application
type HIRApply struct {
	ExprBase

	Func HIRExpr
	Args map[string]HIRExpr

	// VarArgs contains all of the non-spread variadic arguments to the function
	VarArgs []HIRExpr

	// SpreadArg is the argument used for variadic spread initialization
	SpreadArg HIRExpr
}

// -----------------------------------------------------------------------------

// HIRIdentifier represents an identifier
type HIRIdentifier struct {
	Sym *Symbol

	// IdPos is the position of the identifier token
	IdPos *logging.TextPosition
}

func (hi *HIRIdentifier) Type() typing.DataType {
	return hi.Sym.Type
}

func (hi *HIRIdentifier) Category() int {
	// all identifiers are LValues
	return LValue
}

func (hi *HIRIdentifier) Constant() bool {
	return hi.Sym.Mutability == Immutable
}

// HIRLiteral represents a literal
type HIRLiteral struct {
	ExprBase

	// Value can be `null` if this is a null literal; otherwise, it is just the
	// value of the literal (eg. `12`, `0b101`, etc.)
	Value string
}
