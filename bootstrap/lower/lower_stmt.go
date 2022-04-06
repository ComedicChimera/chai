package lower

import (
	"chai/ast"
	"chai/depm"
	"chai/mir"
	"chai/typing"
	"fmt"
)

// lowerStmt lowers a statement in an AST block.
func (l *Lowerer) lowerStmt(parent *mir.Block, stmt ast.Expr) {
	switch v := stmt.(type) {
	case *ast.VarDecl:
		l.lowerLocalVarDecl(parent, v)
	case *ast.Assign:
		l.lowerAssign(parent, v)
	case *ast.UnaryUpdate:
		l.lowerUnaryUpdate(parent, v)
	default:
		// all other statements are just expressions
		parent.Stmts = append(parent.Stmts, &mir.SimpleStmt{
			Kind: mir.SSKindExpr,
			Arg:  l.lowerExpr(v),
		})
	}
}

// lowerLocalVarDecl lowers a local variable declaration.
func (l *Lowerer) lowerLocalVarDecl(parent *mir.Block, vd *ast.VarDecl) {
	for _, vlist := range vd.VarLists {
		mtype := typing.Simplify(vlist.Type)

		if vlist.Initializer != nil {
			mexpr := l.lowerExpr(vlist.Initializer)

			if len(vlist.Names) == 1 {
				mname := fmt.Sprintf("%s$%d", vlist.Names[0], len(l.locals))

				if vlist.Mutabilities[0] == depm.Mutable {
					l.locals[mname] = mtype
					l.currScope()[vlist.Names[0]] = mname

					// initialization
					parent.Stmts = append(parent.Stmts, &mir.AssignStmt{
						LHS: []mir.Expr{&mir.LocalIdent{Name: mname, IdType: mtype, Mutable: true}},
						RHS: []mir.Expr{mexpr},
					})
				} else {
					// constant initialization
					parent.Stmts = append(parent.Stmts, &mir.BindStmt{
						Name: mname,
						Val:  mexpr,
					})
				}
			} else {
				// TODO: pattern matching
			}
		} else {
			for i, name := range vlist.Names {
				mname := fmt.Sprintf("%s$%d", name, len(l.locals))

				if vlist.Mutabilities[i] == depm.Mutable {
					l.locals[mname] = mtype
					l.currScope()[name] = mname

					// null initialization
					parent.Stmts = append(parent.Stmts, &mir.AssignStmt{
						LHS: []mir.Expr{&mir.LocalIdent{Name: mname, IdType: mtype, Mutable: true}},
						RHS: []mir.Expr{&mir.Literal{Value: "null", LitType: mtype}},
					})
				} else {
					// constant null initialization
					parent.Stmts = append(parent.Stmts, &mir.BindStmt{
						Name: mname,
						Val:  &mir.Literal{Value: "null", LitType: mtype},
					})
				}

			}
		}
	}
}

// lowerAssign lowers an assignment statement.
func (l *Lowerer) lowerAssign(parent *mir.Block, as *ast.Assign) {
	if len(as.LHSExprs) != len(as.RHSExprs) {
		// TODO: pattern matching
	} else {
		massign := &mir.AssignStmt{
			LHS: make([]mir.Expr, len(as.LHSExprs)),
			RHS: make([]mir.Expr, len(as.RHSExprs)),
		}

		for i, lhsExpr := range as.LHSExprs {
			rhsExpr := as.RHSExprs[i]

			mLHS := l.lowerExpr(lhsExpr)
			massign.LHS[i] = mLHS

			if as.Oper != nil {
				// compound assignment
				massign.RHS[i] = l.lowerOperApp(*as.Oper, mLHS.Type(), lhsExpr, rhsExpr)
			} else {
				massign.RHS[i] = l.lowerExpr(rhsExpr)
			}
		}
	}
}

// lowerUnaryUpdate lowers a unary update statement.
func (l *Lowerer) lowerUnaryUpdate(parent *mir.Block, uu *ast.UnaryUpdate) {
	// TODO
}
