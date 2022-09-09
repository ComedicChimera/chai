package mir

// IfTree represents an if statement tree.
type IfTree struct {
	StmtBase

	// The conditional branches of the if tree (if/elif).
	CondBranches []CondBranch

	// The else branch of the if tree.
	ElseBlock []Statement
}

// CondBranch represents a conditional branch.
type CondBranch struct {
	// The header block of the conditional branch: contains all the prelude
	// content to the conditional branch.
	HeaderBlock []Statement

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
	ElseBlock []Statement

	// The unconditional update block of the loop.
	UpdateBlock []Statement
}

// DoWhileLoop represents a do-while loop.
type DoWhileLoop struct {
	StmtBase
	CondBranch

	// The else branch of the do-while loop.
	ElseBlock []Statement
}
