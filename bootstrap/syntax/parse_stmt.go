package syntax

import (
	"chaic/ast"
	"chaic/common"
	"chaic/report"
	"chaic/types"
	"chaic/util"
)

// stmt := block_stmt | (simple_stmt | var_decl | expr_assign_stmt) ';' ;
// block_stmt := if_stmt | while_loop | for_loop ;
// simple_stmt := 'break' | 'continue' | 'return' [expr_list] ;
func (p *Parser) parseStmt() ast.ASTNode {
	var stmt ast.ASTNode

	switch p.tok.Kind {
	case TOK_LET, TOK_CONST:
		stmt = p.parseVarDecl()
	case TOK_BREAK, TOK_CONTINUE:
		p.next()
		stmt = &ast.KeywordStmt{
			ASTBase: ast.NewASTBaseOn(p.lookbehind.Span),
			Kind:    p.lookbehind.Kind,
		}
	case TOK_RETURN:
		{
			p.next()
			startSpan := p.lookbehind.Span

			var exprs []ast.ASTExpr
			if !p.has(TOK_SEMI) {
				exprs = p.parseExprList()
			}

			stmt = &ast.ReturnStmt{
				ASTBase: ast.NewASTBaseOver(startSpan, p.lookbehind.Span),
				Exprs:   exprs,
			}
		}
	case TOK_IF:
		return p.parseIfStmt()
	case TOK_WHILE:
		return p.parseWhileLoop()
	case TOK_FOR:
		return p.parseForLoop()
	default:
		stmt = p.parseExprAssignStmt()
	}

	p.want(TOK_SEMI)
	return stmt
}

// var_decl := ('let' | 'const') var_list {',' var_list} ;
// var_list := ident_list (initializer | type_ext [initializer]) ;
func (p *Parser) parseVarDecl() *ast.VarDecl {
	var startSpan *report.TextSpan
	var constant bool
	if p.has(TOK_LET) {
		constant = false
		p.next()
		startSpan = p.lookbehind.Span
	} else {
		constant = true
		startSpan = p.want(TOK_CONST).Span
	}

	var varLists []ast.VarList
	for {
		identToks := p.parseIdentList()

		var typ types.Type
		var init ast.ASTExpr
		if p.has(TOK_COLON) {
			typ = p.parseTypeExt()

			if p.has(TOK_ASSIGN) {
				init = p.parseInitializer()
			}
		} else {
			init = p.parseInitializer()
		}

		varList := ast.VarList{Initializer: init}
		for _, identTok := range identToks {
			ident := &ast.Identifier{
				ASTBase: ast.NewASTBaseOn(identTok.Span),
				Name:    identTok.Value,
				Sym: &common.Symbol{
					Name:       identTok.Value,
					ParentID:   p.chFile.Parent.ID,
					FileNumber: p.chFile.FileNumber,
					DefSpan:    identTok.Span,
					Type:       typ,
					DefKind:    common.DefKindValue,
					Constant:   constant,
				},
			}

			varList.Vars = append(varList.Vars, ident)
		}

		varLists = append(varLists, varList)

		if p.has(TOK_COMMA) {
			p.next()

			continue
		}

		break
	}

	return &ast.VarDecl{
		ASTBase:  ast.NewASTBaseOver(startSpan, p.lookbehind.Span),
		VarLists: varLists,
	}
}

var compoundAssignOps = []int{
	TOK_PLUS, TOK_MINUS, TOK_STAR, TOK_DIV, TOK_MOD, TOK_POW,
	TOK_LSHIFT, TOK_RSHIFT, TOK_BWAND, TOK_BWOR, TOK_BWXOR,
	TOK_LAND, TOK_LOR,
}

// expr_assign_stmt := lhs_expr [{',' lhs_expr} cpd_assign_ops '=' expr_list | '++' | '--'] ;
// lhs_expr := ['*'] atom_expr ;
// cpd_assign_ops := '+' | '-' | '*' | '/' | '%' | '**' | '<<' | '>>'
//				  | '&' | '|' | '^' | '&&' | '||' ;
func (p *Parser) parseExprAssignStmt() ast.ASTNode {
	var lhsExprs []ast.ASTExpr
	for {
		if p.has(TOK_STAR) {
			startSpan := p.tok.Span
			p.next()

			atomExpr := p.parseAtomExpr()
			lhsExprs = append(lhsExprs, &ast.Deref{
				ExprBase: ast.NewExprBase(report.NewSpanOver(startSpan, atomExpr.Span())),
				Ptr:      atomExpr,
			})
		} else {
			lhsExprs = append(lhsExprs, p.parseAtomExpr())
		}

		if p.has(TOK_COMMA) {
			p.next()

			continue
		}

		break
	}

	if len(lhsExprs) == 1 {
		lhsExpr := lhsExprs[0]

		switch p.tok.Kind {
		case TOK_INC, TOK_DEC:
			p.next()

			// Overwrite the lookbehind so that the applied operator is correct.
			p.lookbehind.Kind += TOK_PLUS - TOK_INC
			p.lookbehind.Value = p.lookbehind.Value[:1]

			return &ast.IncDecStmt{
				ASTBase:    ast.NewASTBaseOver(lhsExpr.Span(), p.lookbehind.Span),
				LHSOperand: lhsExpr,
				Op:         newAppliedOper(p.lookbehind),
			}
		case TOK_ASSIGN:
			p.next()

			rhsExpr := p.parseExpr()

			return &ast.Assignment{
				ASTBase:  ast.NewASTBaseOver(lhsExpr.Span(), rhsExpr.Span()),
				LHSVars:  lhsExprs,
				RHSExprs: []ast.ASTExpr{rhsExpr},
			}
		default:
			if util.Contains(compoundAssignOps, p.tok.Kind) {
				compoundAssignOp := p.tok
				p.next()

				p.want(TOK_ASSIGN)

				rhsExpr := p.parseExpr()

				return &ast.Assignment{
					ASTBase:    ast.NewASTBaseOver(lhsExpr.Span(), rhsExpr.Span()),
					LHSVars:    lhsExprs,
					RHSExprs:   []ast.ASTExpr{rhsExpr},
					CompoundOp: newAppliedOper(compoundAssignOp),
				}
			} else {
				return lhsExpr
			}
		}
	} else {
		var compoundOp *common.AppliedOperator
		if p.has(TOK_ASSIGN) {
			p.next()
		} else if util.Contains(compoundAssignOps, p.tok.Kind) {
			compoundOp = newAppliedOper(p.tok)
			p.next()
			p.want(TOK_ASSIGN)
		} else {
			p.reject()
		}

		rhsExprs := p.parseExprList()

		return &ast.Assignment{
			ASTBase:    ast.NewASTBaseOver(lhsExprs[0].Span(), rhsExprs[len(rhsExprs)-1].Span()),
			LHSVars:    lhsExprs,
			RHSExprs:   rhsExprs,
			CompoundOp: compoundOp,
		}
	}
}
