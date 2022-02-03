package syntax

import (
	"chai/ast"
	"chai/report"
)

// block = 'NEWLINE' {stmt 'NEWLINE'}
// stmt = var_decl | const_decl | block_expr | tupled_expr | expr_stmt | control_stmt
// control_stmt = 'break' | 'continue' | 'fallthrough' | 'return' [expr {',' expr}]
func (p *Parser) parseBlock() (ast.Expr, bool) {
	// skip leading newlines
	if !p.assertAndAdvance(NEWLINE) {
		return nil, false
	}

	var stmts []ast.Expr

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
			// block_expr
			if stmt, ok := p.parseBlockExpr(); ok {
				stmts = append(stmts, stmt)
			} else {
				return nil, false
			}
		case LPAREN:
			// tupled_expr
			if tupledExpr, ok := p.parseTupledExpr(); ok {
				stmts = append(stmts, tupledExpr)
			} else {
				return nil, false
			}
		case BREAK:
			// TODO
		case CONTINUE:
			// TODO
		case FALLTHROUGH:
			// TODO
		case RETURN:
			// TODO
		case END, ELIF, ELSE, CASE, AFTER:
			// closers: we do NOT want to the consume these: leave them for the
			// caller to consume as needed
			return &ast.Block{
				Stmts: stmts,
			}, true
		default:
			// expr_stmt
			if exprStmt, ok := p.parseExprStmt(); ok {
				stmts = append(stmts, exprStmt)
			} else {
				return nil, false
			}
		}

		// skip newlines between statements
		if !p.assertAndAdvance(NEWLINE) {
			return nil, false
		}
	}
}

// -----------------------------------------------------------------------------

// block_expr = do_expr | if_expr | while_expr | match_expr | for_expr
func (p *Parser) parseBlockExpr() (ast.Expr, bool) {
	switch p.tok.Kind {
	case DO:
		if !p.next() {
			return nil, false
		}

		result, ok := p.parseBlock()
		if ok {
			return result, p.assertAndNext(END)
		} else {
			return nil, false
		}
	case IF:
		return p.parseIfExpr()
	case WHILE:
		return p.parseWhileExpr()
	case FOR:
		// TODO
	case MATCH:
		// TODO
	}

	// unreachable
	return nil, false
}

// if_expr = 'if' cond_branch {'elif' cond_branch} ['else' block_body] 'end
func (p *Parser) parseIfExpr() (ast.Expr, bool) {
	// save the position of the initial `if` for calculating the final position
	// of the if expression
	ifPos := p.tok.Position

	if !p.assertAndAdvance(IF) {
		return nil, false
	}

	// create an initial if expression to populate
	ifExpr := &ast.IfExpr{
		ExprBase: ast.NewExprBase(nil, ast.RValue),
	}

	// main if branch
	if !p.parseCondBranch(ifExpr) {
		return nil, false
	}

	// subsequent elif branches
	for p.got(ELIF) {
		if !p.advance() {
			return nil, false
		}

		if !p.parseCondBranch(ifExpr) {
			return nil, false
		}
	}

	// final else branch
	if p.got(ELSE) {
		if !p.advance() {
			return nil, false
		}

		if body, ok := p.parseBlockBody(); ok {
			ifExpr.ElseBranch = body
		} else {
			return nil, false
		}
	}

	// final `end` of expression
	endPos := p.tok.Position
	if !p.assertAndNext(END) {
		return nil, false
	}

	// set the position of the if expression
	ifExpr.Pos = report.TextPositionFromRange(ifPos, endPos)

	// return the final, crafted if expression
	return ifExpr, true
}

// cond_branch = block_header block_body
func (p *Parser) parseCondBranch(ifExpr *ast.IfExpr) bool {
	vd, cond, ok := p.parseBlockHeader()
	if !ok {
		return false
	}

	body, ok := p.parseBlockBody()
	if !ok {
		return false
	}

	ifExpr.CondBranches = append(ifExpr.CondBranches, ast.CondBranch{
		HeaderVarDecl: vd,
		Cond:          cond,
		Body:          body,
	})

	// newlines in between branches
	return p.newlines()
}

// while_expr = 'while' block_header block_body [loop_after] 'end'
func (p *Parser) parseWhileExpr() (ast.Expr, bool) {
	// save the position of the initial `while` for calculating the final
	// position of the while expression
	whilePos := p.tok.Position

	if !p.assertAndAdvance(WHILE) {
		return nil, false
	}

	// main part of while loop
	headerVarDecl, cond, ok := p.parseBlockHeader()
	if !ok {
		return nil, false
	}

	body, ok := p.parseBlockBody()
	if !ok {
		return nil, false
	}

	// TODO: after branches

	// final `end` of expression
	endPos := p.tok.Position
	if !p.assertAndNext(END) {
		return nil, false
	}

	// return the final while loop
	return &ast.WhileExpr{
		ExprBase:      ast.NewExprBase(nil, ast.RValue),
		HeaderVarDecl: headerVarDecl,
		Cond:          cond,
		Body:          body,
		Pos:           report.TextPositionFromRange(whilePos, endPos),
	}, true
}

// -----------------------------------------------------------------------------

// block_header = [var_decl ';'] expr
func (p *Parser) parseBlockHeader() (*ast.VarDecl, ast.Expr, bool) {
	// collect the optional variable declaration
	var vd *ast.VarDecl
	if p.got(LET) {
		_vd, ok := p.parseVarDecl(false, nil, false)
		if !ok {
			return nil, nil, false
		}
		vd = _vd.(*ast.VarDecl)

		if !p.assertAndAdvance(SEMICOLON) {
			return nil, nil, false
		}
	}

	// collect the main expression
	expr, ok := p.parseExpr()
	if !ok {
		return nil, nil, false
	}

	return vd, expr, true
}

// block_body = '=>' expr | block
func (p *Parser) parseBlockBody() (ast.Expr, bool) {
	if p.got(ARROW) {
		if !p.advance() {
			return nil, false
		}

		return p.parseExpr()
	}

	return p.parseBlock()
}
