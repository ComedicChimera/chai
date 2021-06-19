package walk

import (
	"chai/sem"
	"chai/syntax"
)

// walkStmt walks an `stmt` node
func (w *Walker) walkStmt(branch *syntax.ASTBranch) (sem.HIRExpr, bool) {
	stmt := branch.BranchAt(0)
	switch stmt.Name {

	}

	return nil, false
}

// -----------------------------------------------------------------------------

// walkExprStmt walks an `expr_stmt` node
func (w *Walker) walkExprStmt(branch *syntax.ASTBranch) (sem.HIRExpr, bool) {
	return nil, false
}
