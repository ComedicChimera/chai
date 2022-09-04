package mir

// IfTree represents an if statement tree.
type IfTree struct {
	StmtBase

	// The conditional branches of the if tree (if/elif).
	CondBranches []CondBranch

	// The else branch of the if tree.
	ElseBranch []Statement
}

// CondBranch represents a conditional branch.
type CondBranch struct {
	// The condition of the branch.
	Condition Expr

	// The body of the branch.
	Body []Statement
}

// WhileLoop represents a while loop.
type WhileLoop struct {
	StmtBase
	CondBranch

	// The else branch of the while loop.
	ElseBranch []Statement

	// The unconditional update statement of the loop.
	UpdateStmt Statement
}

// DoWhileLoop represents a do-while loop.
type DoWhileLoop struct {
	StmtBase
	CondBranch

	// The else branch of the do-while loop.
	ElseBranch []Statement
}
