package lower

import (
	"chaic/ast"
	"chaic/mir"
	"chaic/types"
)

func (l *Lowerer) lowerVarDecl(vd *ast.VarDecl) {
	for _, vlist := range vd.VarLists {
		init := l.lowerExpr(vlist.Initializer)

		// Struct literal variable initializers.
		if _, ok := vlist.Initializer.(*ast.StructLiteral); ok {
			for _, ident := range vlist.Vars {
				ident.Sym.MIRSymbol = init.(*mir.Identifier).Symbol
			}
		} else {
			for _, ident := range vlist.Vars {
				vd := &mir.VarDecl{
					StmtBase: mir.NewStmtBase(vd.Span()),
					Ident: &mir.Identifier{
						ExprBase: mir.NewExprBase(ident.Span()),
						Symbol: &mir.MSymbol{
							Name:              ident.Name,
							Type:              types.Simplify(ident.Type()),
							IsImplicitPointer: true,
						},
					},
					Initializer: init,
				}

				l.appendStmt(vd)
			}
		}
	}
}

func (l *Lowerer) lowerAssignment(assign *ast.Assignment) {
	// TODO
}

func (l *Lowerer) lowerIncDecStmt(incdec *ast.IncDecStmt) mir.Statement {
	// TODO
	return nil
}
