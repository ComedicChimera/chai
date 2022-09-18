package ast

// VarDecl represents a variable declaration.
type VarDecl struct {
	ASTBase

	// The list of variable declaration lists.
	VarLists []VarList
}

// VarList represents a single collection of variables
// sharing a common type label and/or initializer.
type VarList struct {
	Vars        []*Identifier
	Initializer ASTExpr
}

// Assignment represents an assignment statement.
type Assignment struct {
	ASTBase

	// The list of LHS variables.
	LHSVars []ASTExpr

	// The list of RHS expressions.
	RHSExprs []ASTExpr

	// The optional compound binary operator application.
	// CompoundOp *common.AppliedOperator
}

// IncDecStmt represents an increment/decrement statement.
type IncDecStmt struct {
	ASTBase

	// The LHS variable being mutated.
	LHSOperand ASTExpr

	// The binary operator being applied.
	// Op *common.AppliedOperator
}

// -----------------------------------------------------------------------------

// KeywordStmt represents a single keyword control flow statement (eg. `break`).
type KeywordStmt struct {
	ASTBase

	// The token kind of the keyword.
	Kind int
}

// ReturnStmt represents a return statement.
type ReturnStmt struct {
	ASTBase

	// The expressions being returned.
	Exprs []ASTExpr
}
