package syntax

import "chai/ast"

// block = 'NEWLINE' {stmt 'NEWLINE'} 'end'
// stmt = var_decl | const_decl | block_expr | tupled_expr | expr_stmt
func (p *Parser) parseBlock() (ast.Expr, bool) {
	if !p.assertAndNext(NEWLINE) {
		return nil, false
	}

	var stmts []ast.Expr

	// skip leading newlines
	if !p.newlines() {
		return nil, false
	}

	for {
		switch p.tok.Kind {
		case LET:
			// var_decl
			if stmt, ok := p.parseVarDecl(false, nil, false); ok {
				stmts = append(stmts, stmt)
			} else {
				return nil, false
			}
		case CONST:
			// TODO: const_decl
		case IF, MATCH, FOR, WHILE, DO:
			// TODO: block_expr
		case LPAREN:
			// tupled_expr
			if tupledExpr, ok := p.parseTupledExpr(); ok {
				stmts = append(stmts, tupledExpr)
			} else {
				return nil, false
			}
		default:
			// expr_stmt
			if exprStmt, ok := p.parseExprStmt(); ok {
				stmts = append(stmts, exprStmt)
			} else {
				return nil, false
			}
		}

		if p.assertAndNext(NEWLINE) {
			// skip newlines between statements
			if !p.newlines() {
				return nil, false
			}

			// exit condition: `end`
			if p.got(END) {
				if !p.next() {
					return nil, false
				}

				break
			}
		} else {
			return nil, false
		}
	}

	return &ast.Block{
		Stmts: stmts,
	}, true
}
