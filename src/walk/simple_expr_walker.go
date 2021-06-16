package walk

import (
	"chai/sem"
	"chai/syntax"
)

// walkSimpleExpr walks a simple expression (`simple_expr`)
func (w *Walker) walkSimpleExpr(branch *syntax.ASTBranch) (sem.HIRExpr, bool) {
	return nil, false
}
