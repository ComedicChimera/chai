package lower

import (
	"chai/ast"
	"chai/mir"
)

// lowerBlock lowers a series of expression statements (ie. a do block)
func (l *Lowerer) lowerBlock(exprs []ast.Expr) (mir.Value, []mir.Stmt) {
	l.pushScope()
	defer l.popScope()

	var stmts []mir.Stmt
	for _, exprStmt := range exprs {
		_ = exprStmt
	}

	return nil, stmts
}
