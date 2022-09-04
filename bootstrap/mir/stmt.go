package mir

import "chaic/report"

// Statement represents a MIR statement.
type Statement interface {
	// Span returns the source position of the statement.
	Span() *report.TextSpan
}

// The base type for all statements.
type stmtBase struct {
	// The span of the statement.
	span *report.TextSpan
}

// NewStmtBase creates a new statement base with span span.
func NewStmtBase(span *report.TextSpan) stmtBase {
	return stmtBase{span: span}
}

func (sb stmtBase) Span() *report.TextSpan {
	return sb.span
}

/* -------------------------------------------------------------------------- */

// VarDecl represents a variable or temporary declaration.
type VarDecl struct {
	stmtBase

	// The identifier of the variable.
	Ident *Identifier

	// The initializer of the variable.
	Initializer Expr

	// Whether the variable is a temporary.
	Temporary bool

	// Whether the variable is always explicitly returned.
	Returned bool
}

// StructDecl represents a struct declaration.
type StructDecl struct {
	stmtBase

	// The identifier of the struct.
	Ident *Identifier

	// The field initializers of the struct by field index.
	FieldInits []Expr

	// Whether the struct is always explicitly returned.
	Returned bool
}

// Assignment represents an L-value assignment.
type Assignment struct {
	stmtBase

	// The LHS expression.
	LHS Expr

	// The RHS expression.
	RHS Expr
}

// Return represents a return statement.
type Return struct {
	stmtBase

	// The value being returned (if any).
	Value Expr
}

// Break represents a break statement.
type Break struct {
	stmtBase
}

// Continue represents a continue statement.
type Continue struct {
	stmtBase
}
