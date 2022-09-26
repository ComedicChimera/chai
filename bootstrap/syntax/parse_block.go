package syntax

import (
	"chaic/ast"
	"chaic/report"
)

// block := '{' stmt {stmt} '}' | ':' stmt ;
func (p *Parser) parseBlock() *ast.Block {
	if p.has(TOK_LBRACE) {
		startSpan := p.tok.Span
		p.next()

		var stmts []ast.ASTNode
		for !p.has(TOK_RBRACE) && !p.has(TOK_EOF) {
			stmts = append(stmts, p.parseStmt())
		}

		endSpan := p.want(TOK_RBRACE).Span

		return &ast.Block{
			ASTBase: ast.NewASTBaseOver(startSpan, endSpan),
			Stmts:   stmts,
		}
	} else if p.has(TOK_COLON) {
		startSpan := p.tok.Span
		p.next()

		stmt := p.parseStmt()

		return &ast.Block{
			ASTBase: ast.NewASTBaseOver(startSpan, p.lookbehind.Span),
			Stmts:   []ast.ASTNode{stmt},
		}
	} else {
		p.reject()
		return nil
	}
}

// if_stmt := 'if' cond_branch {'elif' cond_branch} ['else' block] ;
// cond_branch := [var_decl ';'] expr block ;
func (p *Parser) parseIfStmt() *ast.IfTree {
	startSpan := p.want(TOK_IF).Span

	var condBranches []ast.CondBranch
	for {
		var headerVarDecl *ast.VarDecl
		if p.has(TOK_LET) || p.has(TOK_CONST) {
			headerVarDecl = p.parseVarDecl()

			p.want(TOK_SEMI)
		}

		outerExprNestDepth := p.exprNestDepth
		p.exprNestDepth = -1
		condition := p.parseExpr()
		p.exprNestDepth = outerExprNestDepth

		body := p.parseBlock()

		condBranches = append(condBranches, ast.CondBranch{
			HeaderVarDecl: headerVarDecl,
			Condition:     condition,
			Body:          body,
		})

		if p.has(TOK_ELIF) {
			p.next()

			continue
		}

		break
	}

	var elseBranch *ast.Block
	if p.has(TOK_ELSE) {
		p.next()

		elseBranch = p.parseBlock()
	}

	return &ast.IfTree{
		ASTBase:      ast.NewASTBaseOver(startSpan, p.lookbehind.Span),
		CondBranches: condBranches,
		ElseBranch:   elseBranch,
	}
}

// while_loop := 'while' [var_decl ';'] expr block [loop_else] ;
// loop_else := 'else' block ;
func (p *Parser) parseWhileLoop() *ast.WhileLoop {
	startSpan := p.want(TOK_WHILE).Span

	var headerVarDecl *ast.VarDecl
	if p.has(TOK_LET) || p.has(TOK_CONST) {
		headerVarDecl = p.parseVarDecl()
	}

	outerExprNestDepth := p.exprNestDepth
	p.exprNestDepth = -1
	condition := p.parseExpr()
	p.exprNestDepth = outerExprNestDepth

	body := p.parseBlock()

	var elseBlock *ast.Block
	if p.has(TOK_ELSE) {
		p.next()

		elseBlock = p.parseBlock()
	}

	return &ast.WhileLoop{
		ASTBase:       ast.NewASTBaseOver(startSpan, p.lookbehind.Span),
		HeaderVarDecl: headerVarDecl,
		Condition:     condition,
		Body:          body,
		ElseBlock:     elseBlock,
	}
}

// for_loop := 'for' c_for_loop ;
func (p *Parser) parseForLoop() ast.ASTNode {
	startSpan := p.want(TOK_FOR).Span

	var iterVarDecl *ast.VarDecl
	switch p.tok.Kind {
	case TOK_LET, TOK_CONST:
		iterVarDecl = p.parseVarDecl()
		fallthrough
	case TOK_SEMI:
		p.want(TOK_SEMI)
		return p.parseCForLoop(startSpan, iterVarDecl)
	default:
		// TODO: for iter loop
		p.reject()
		return nil
	}
}

// c_for_loop := [var_decl] ';' expr ';' [expr_assign_stmt] block [loop_else] ;
func (p *Parser) parseCForLoop(startSpan *report.TextSpan, iterVarDecl *ast.VarDecl) *ast.CForLoop {
	condition := p.parseExpr()
	p.want(TOK_SEMI)

	var updateStmt ast.ASTNode
	if !p.has(TOK_COLON) && !p.has(TOK_LBRACE) {
		outerExprNestDepth := p.exprNestDepth
		p.exprNestDepth = -1
		updateStmt = p.parseExprAssignStmt()
		p.exprNestDepth = outerExprNestDepth
	}

	body := p.parseBlock()

	var elseBlock *ast.Block
	if p.has(TOK_ELSE) {
		p.next()

		elseBlock = p.parseBlock()
	}

	return &ast.CForLoop{
		ASTBase:     ast.NewASTBaseOver(startSpan, p.lookbehind.Span),
		IterVarDecl: iterVarDecl,
		Condition:   condition,
		UpdateStmt:  updateStmt,
		Body:        body,
		ElseBlock:   elseBlock,
	}
}

// do_while_loop := 'do' block 'while' expr (';' | loop_else) ;
func (p *Parser) parseDoWhileLoop() *ast.DoWhileLoop {
	startSpan := p.want(TOK_DO).Span

	body := p.parseBlock()

	p.want(TOK_WHILE)

	condition := p.parseExpr()

	var elseBlock *ast.Block
	if p.has(TOK_ELSE) {
		p.next()

		elseBlock = p.parseBlock()
	} else {
		p.want(TOK_SEMI)
	}

	return &ast.DoWhileLoop{
		ASTBase:   ast.NewASTBaseOver(startSpan, p.lookbehind.Span),
		Body:      body,
		Condition: condition,
		ElseBlock: elseBlock,
	}
}
