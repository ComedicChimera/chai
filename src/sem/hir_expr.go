package sem

import "chai/typing"

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

// exprBase is the base struct for all expressions
type exprBase struct {
	dt       typing.DataType
	cat      int
	constant bool
}

func (eb *exprBase) Type() typing.DataType {
	return eb.dt
}

func (eb *exprBase) Category() int {
	return eb.cat
}

func (eb *exprBase) Constant() bool {
	return eb.constant
}
