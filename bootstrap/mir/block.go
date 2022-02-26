package mir

import "chai/typing"

type Block struct {
	Stmts     []Stmt
	YieldType typing.DataType
}

func (b *Block) Type() typing.DataType {
	return b.YieldType
}

type Stmt interface {
	// Term returns whether or not the statement is a block terminator.
	Term() bool
}

// -----------------------------------------------------------------------------

// IfTree represents and if/elif/else tree.
type IfTree struct {
	CondBranches []CondBranch
	ElseBranch   *Block
}

// CondBranch represents a single condition (if/elif) branch of an if tree.
type CondBranch struct {
	Condition Expr
	Body      *Block
}

// Loop represents a loop based on a condition.
type Loop struct {
	Condition Expr
	Body      *Block

	// PostIter runs after every loop iteration (eg. the `i++`)
	PostIter *Block
}

// -----------------------------------------------------------------------------

// SimpleStmt is an umbrella structure for several "simple statements" which can
// be represented by a single (optional) expression and an enum value
type SimpleStmt struct {
	Kind int
	Arg  Expr
}

// Enumeration of simple statement kinds
const (
	SSKindExpr = iota // Expression statement
	SSKindReturn
	SSKindYield
	SSKindBreak
	SSKindContinue
	SSKindFallthrough
)

func (ss *SimpleStmt) Term() bool {
	return ss.Kind != SSKindExpr
}

// AssignStmt represents an assignment statement.  This expression does perform
// simple multi-assignment (for sake of brevity in the MIR).  But it does not
// represent compound or pattern-matching-based assignment.
type AssignStmt struct {
	LHS, RHS []Expr
}

func (as *AssignStmt) Term() bool {
	return false
}

// BindStmt creates a constant, named binding to a value.  This is used to
// represent implicit constants and intermediate values.
type BindStmt struct {
	Name string
	Val  Expr
}

func (bs *BindStmt) Term() bool {
	return false
}
