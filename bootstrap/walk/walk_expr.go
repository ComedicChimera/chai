package walk

import "chai/ast"

// walkExpr walks an AST expression.  It updates the AST with types (mostly type
// variables to be determined by the solver).
func (w *Walker) walkExpr(expr ast.Expr) bool {
	// just switch over the different kinds of expressions
	switch v := expr.(type) {
	case *ast.Literal:
		w.walkLiteral(v)

		// always succeed
		return true
	case *ast.Identifier:
		// just lookup the identifier for now
		if sym, ok := w.lookup(v.Name, v.Pos); ok {
			// update the type with the type of the matching symbol :)
			v.ExprBase.SetType(sym.Type)
			return true
		} else {
			return false
		}
	case *ast.Tuple:
		// just walk all the sub expressions
		for _, expr := range v.Exprs {
			if !w.walkExpr(expr) {
				return false
			}
		}

		return true
	case *ast.BinaryOp:
		// TODO
	case *ast.MultiComparison:
		// TODO
	}

	// unreachable
	return false
}

// -----------------------------------------------------------------------------

// walkLiteral walks a literal value.
func (w *Walker) walkLiteral(lit *ast.Literal) {
	// TODO
}
