package mir

import "chai/typing"

type Block struct {
	Stmts     []Stmt
	YieldType typing.DataType
	TermMode  int
}

func (b *Block) Type() typing.DataType {
	return b.YieldType
}

type Stmt interface {
	// Term returns the control flow termination mode of the statement.  If the
	// statement is not a terminator, this is `BTNone`.
	Term() int
}

const (
	BTLoop = iota
	BTCase
	BTFunc
	BTBlock
	BTNone
)

// -----------------------------------------------------------------------------

// IfTree represents and if/elif/else tree.
type IfTree struct {
	CondBranches []CondBranch
	ElseBranch   *Block
	TermMode     int
}

func (ift *IfTree) Term() int {
	return ift.TermMode
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
	TermMode  int

	// PostIter runs after every loop iteration (eg. the `i++`)
	PostIter *Block
}

func (loop *Loop) Term() int {
	return loop.TermMode
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

func (ss *SimpleStmt) Term() int {
	switch ss.Kind {
	case SSKindExpr:
		return BTNone
	case SSKindYield:
		return BTBlock
	case SSKindBreak, SSKindContinue:
		return BTLoop
	case SSKindFallthrough:
		return BTCase
	default: // return
		return BTFunc
	}
}

// AssignStmt represents an assignment statement.  This expression does perform
// simple multi-assignment (for sake of brevity in the MIR).  But it does not
// represent compound or pattern-matching-based assignment.
type AssignStmt struct {
	LHS, RHS []Expr
}

func (as *AssignStmt) Term() int {
	return BTNone
}

// BindStmt creates a constant, named binding to a value.  This is used to
// represent implicit constants and intermediate values.
type BindStmt struct {
	Name string
	Val  Expr
}

func (bs *BindStmt) Term() int {
	return BTNone
}
