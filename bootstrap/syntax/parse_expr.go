package syntax

import (
	"chai/ast"
	"chai/report"
)

// expr = simple_expr | block_expr
func (p *Parser) parseExpr() (ast.Expr, bool) {
	// TODO: simple_expr | block_expr
	return p.parseAtom()
}

// -----------------------------------------------------------------------------

// atom = 'INTLIT' | 'FLOATLIT' | 'NUMLIT' | 'STRINGLIT' | 'RUNELIT'
//   | 'BOOLLIT' | 'IDENTIFIER' | 'NULL' | tupled_expr | ...
func (p *Parser) parseAtom() (ast.Expr, bool) {
	startTok := p.tok

	switch p.tok.Kind {
	case INTLIT, FLOATLIT, NUMLIT, STRINGLIT, RUNELIT, BOOLLIT, NULL:
		p.next()
		return &ast.Literal{
			ExprBase: ast.NewExprBase(nil, ast.RValue),
			Kind:     startTok.Kind,
			Value:    startTok.Value,
			Pos:      startTok.Position,
		}, true
	case IDENTIFIER:
		p.next()
		return &ast.Identifier{
			ExprBase: ast.NewExprBase(nil, ast.LValue),
			Name:     startTok.Value,
			Pos:      startTok.Position,
		}, true
	case LPAREN:
		return p.parseTupledExpr()
	}

	p.reject()
	return nil, false
}

// tupled_expr = '(' [expr {',' expr}] ')'
func (p *Parser) parseTupledExpr() (ast.Expr, bool) {
	startTok := p.tok
	if !p.next() {
		return nil, false
	}

	if p.got(RPAREN) {
		lit := &ast.Literal{
			ExprBase: ast.NewExprBase(nil, ast.RValue),
			Kind:     NOTHING,
			Pos:      report.TextPositionFromRange(startTok.Position, p.tok.Position),
		}

		if !p.next() {
			return nil, false
		}

		return lit, true
	}

	expr, ok := p.parseExpr()
	if !ok {
		return nil, false
	}

	exprs := []ast.Expr{expr}
	for p.got(COMMA) {
		if !p.next() {
			return nil, false
		}

		if nextExpr, ok := p.parseExpr(); ok {
			exprs = append(exprs, nextExpr)
		} else {
			return nil, false
		}
	}

	endTok := p.tok
	if !p.assertAndNext(RPAREN) {
		return nil, false
	}

	cat := expr.Category()
	if len(exprs) > 1 {
		cat = ast.RValue
	}

	return &ast.Tuple{
		ExprBase: ast.NewExprBase(nil, cat),
		Exprs:    exprs,
		Pos:      report.TextPositionFromRange(startTok.Position, endTok.Position),
	}, true
}
