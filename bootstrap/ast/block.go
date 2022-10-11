package ast

import "chaic/common"

// Block represents a list of AST statements.
type Block struct {
	ASTBase

	// The statements of the block.
	Stmts []ASTNode
}

// -----------------------------------------------------------------------------

// IfTree represents an if/elif/else tree.
type IfTree struct {
	ASTBase

	// The list of conditional branches which make up the block.
	CondBranches []CondBranch

	// The else branch of the block.
	ElseBranch *Block
}

// CondBranch represents a single conditional branch of if/elif/else tree.
type CondBranch struct {
	// The (optional) header variable declaration.
	HeaderVarDecl *VarDecl

	// The condition of the branch.
	Condition ASTExpr

	// The body of the branch.
	Body *Block
}

// WhileLoop represents a while loop.
type WhileLoop struct {
	ASTBase

	// The (optional) header variable declaration.
	HeaderVarDecl *VarDecl

	// The condition of loop.
	Condition ASTExpr

	// The body of the loop.
	Body *Block

	// The (optional) else block of the while loop.
	ElseBlock *Block
}

// CForLoop represents a C-style for loop.
type CForLoop struct {
	ASTBase

	// The (optional) iterator variable declaration.
	IterVarDecl *VarDecl

	// The loop condition.
	Condition ASTExpr

	// The (optional) iterator update statement.
	UpdateStmt ASTNode

	// The body of the loop.
	Body *Block

	// The (optional) else block of the for loop.
	ElseBlock *Block
}

// DoWhileLoop represents a do-while loop.
type DoWhileLoop struct {
	ASTBase

	// The body of the loop.
	Body *Block

	// The loop condition.
	Condition ASTExpr

	// The (optional) else block of the do-while loop.
	ElseBlock *Block
}

// ForEachLoop represents for-each loop.
type ForEachLoop struct {
	ASTBase

	// The iterator variable.
	IterVarSym *common.Symbol

	// The sequence or iterable.
	SeqExpr ASTExpr

	// The body of the loop.
	Body *Block

	// The (optional) else block of the for-each loop.
	ElseBlock *Block
}
