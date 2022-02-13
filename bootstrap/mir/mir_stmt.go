package mir

import "chai/typing"

// BlockElem represents an item that can be contained in a MIR block.
type BlockElem interface {
	// Term returns true if the statement terminates a block
	Term() bool
}

// -----------------------------------------------------------------------------

// SimpleStmt represents a simple (keyword) statement in MIR.
type SimpleStmt struct {
	StmtKind int
	Exprs    []Expr
}

// Enumeration of MIR statement kinds
const (
	SStmtReturn = iota // Return statement
	SStmtYield         // Yield statement
	SStmtAssign        // Assign statement
	SExprStmt          // Expression as statement
)

func (ss *SimpleStmt) Term() bool {
	switch ss.StmtKind {
	case SStmtReturn, SStmtYield:
		return true
	default:
		return false
	}
}

// Bind represents a temporary, immutable binding of a value to a name.
type Bind struct {
	Name  string
	Type  typing.DataType
	Value Expr
}
