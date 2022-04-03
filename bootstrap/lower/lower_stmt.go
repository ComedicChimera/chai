package lower

import (
	"chai/ast"
	"chai/mir"
)

// lowerStmt lowers a statement in an AST block.
func (l *Lowerer) lowerStmt(stmt ast.Expr) mir.Stmt {
	switch v := stmt.(type) {
	case *ast.VarDecl:
		return l.lowerLocalVarDecl(v)
	case *ast.Assign:
		return l.lowerAssign(v)
	case *ast.UnaryUpdate:
		return l.lowerUnaryUpdate(v)
	default:
		// all other statements are just expressions
		return &mir.SimpleStmt{
			Kind: mir.SSKindExpr,
			Arg:  l.lowerExpr(v),
		}
	}
}

// lowerLocalVarDecl lowers a local variable declaration.
func (l *Lowerer) lowerLocalVarDecl(vd *ast.VarDecl) mir.Stmt {
	// TODO
	return nil
}

// lowerAssign lowers an assignment statement.
func (l *Lowerer) lowerAssign(as *ast.Assign) mir.Stmt {
	// TODO
	return nil
}

// lowerUnaryUpdate lowers a unary update statement.
func (l *Lowerer) lowerUnaryUpdate(uu *ast.UnaryUpdate) mir.Stmt {
	// TODO
	return nil
}
