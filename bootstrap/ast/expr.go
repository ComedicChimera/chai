package ast

import "chai/typing"

// Expr represents an expression simple or complex.
type Expr interface {
	// Type is the yielded type of the expression.
	Type() typing.DataType

	// Category is the value category of the expression. It should be one of the
	// enumerated value categories.
	Category() int
}
