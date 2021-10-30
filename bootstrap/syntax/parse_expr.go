package syntax

import (
	"chai/ast"
	"chai/report"
)

// expr_list = expr {',' expr}
func (p *Parser) parseExprList() ([]ast.Expr, bool) {
	var exprs []ast.Expr

	for {
		if expr, ok := p.parseExpr(); ok {
			exprs = append(exprs, expr)
		} else {
			return nil, false
		}

		if p.got(COMMA) {
			if !p.next() {
				return nil, false
			}
		} else {
			break
		}
	}

	return exprs, true
}

// expr = simple_expr | block_expr
func (p *Parser) parseExpr() (ast.Expr, bool) {
	// TODO: simple_expr | block_expr
	return p.parseBinOpExpr()
}

// -----------------------------------------------------------------------------

// or_expr = xor_expr {('||' | '|') xor_expr}
// xor_expr = and_expr {'^' and_expr}
// and_expr = eq_expr {('&&' | '&') eq_expr}
// eq_expr = comp_expr {('==' | '!=') comp_expr}
// comp_expr = shift_expr {('<' | '>' | '<=' | '>=') shift_expr}
// shift_expr = arith_expr {('>>' | '<<') arith_expr}
// arith_expr = term {('+' | '-') term}
// term = factor {('*' | '/' | '//' | '%') factor}
// factor = unary_expr {'**' unary_expr}
func (p *Parser) parseBinOpExpr() (ast.Expr, bool) {
	lhs, ok := p.parseUnaryExpr()
	if !ok {
		return nil, false
	}

	return p.precedenceParse(lhs, len(precTable))
}

// precTable is the operator precedence table for binary operators. The table is
// ordered highest to lowest precedence.
var precTable [][]int = [][]int{
	{POWER},
	{STAR, IDIV, FDIV, MOD},
	{PLUS, MINUS},
	{LSHIFT, RSHIFT},
	{GT, LT, GTEQ, LTEQ},
	{EQ, NEQ},
	{AND, AMP},
	{CARRET},
	{OR, PIPE},
}

// precedenceParse is a helper function used to perform operator precedence
// parsing for binary operator -- it is essentially an augmented implementation
// of a Pratt parser.
func (p *Parser) precedenceParse(lhs ast.Expr, maxPrec int) (ast.Expr, bool) {
	for {
		// check to see if the lookahead matches any of the operators at or
		// above our precedence level.
		var op *Token
		var opPrec int
		for prec, precLevel := range precTable[:maxPrec] {
			if p.gotOneOf(precLevel...) {
				op = p.tok
				opPrec = prec
				break
			}
		}

		// no matching operator
		if op == nil {
			break
		}

		if !p.next() {
			return nil, false
		}

		rhs, ok := p.parseUnaryExpr()
		if !ok {
			return nil, false
		}

	nextOpLoop:
		for {
			var precBound int

			// `**` is right associative
			if opPrec == 0 {
				precBound = 1
			} else {
				precBound = opPrec
			}

			for _, precLevel := range precTable[:precBound] {
				if p.gotOneOf(precLevel...) {
					rhs, ok = p.precedenceParse(rhs, precBound)
					if !ok {
						return nil, false
					}

					continue nextOpLoop
				}
			}

			break nextOpLoop
		}

		lhs = &ast.BinaryOp{
			ExprBase: ast.NewExprBase(nil, ast.RValue),
			Op: ast.Oper{
				Kind: op.Kind,
				Name: op.Value,
				Pos:  op.Position,
			},
			Lhs: lhs,
			Rhs: rhs,
		}

		// check for ternary comparison operators
		if lbop, ok := lhs.(*ast.BinaryOp); ok {
			if LT <= lbop.Op.Kind && lbop.Op.Kind <= GTEQ {
				lhs = buildMultiCompare(lbop)
			}
		}
	}

	return lhs, true
}

// buildMultiCompare builds a multi-comparison expression.
func buildMultiCompare(root *ast.BinaryOp) ast.Expr {
	exprs := []ast.Expr{root.Rhs}
	ops := []ast.Oper{root.Op}

	for {
		if lbop, ok := root.Lhs.(*ast.BinaryOp); ok && LT <= lbop.Op.Kind && lbop.Op.Kind <= GTEQ {
			exprs = append([]ast.Expr{lbop.Rhs}, exprs...)
			ops = append([]ast.Oper{lbop.Op}, ops...)

			root = lbop
		} else if mc, ok := root.Lhs.(*ast.MultiComparison); ok {
			mc.Exprs = append(mc.Exprs, exprs...)
			mc.Ops = append(mc.Ops, ops...)
			return mc
		} else {
			exprs = append([]ast.Expr{root.Lhs}, exprs...)
			break
		}
	}

	// two expressions => just a regular comparison operation
	if len(exprs) == 2 {
		return root
	}

	return &ast.MultiComparison{
		ExprBase: ast.NewExprBase(nil, ast.RValue),
		Exprs:    exprs,
		Ops:      ops,
	}
}

// -----------------------------------------------------------------------------

// unary_expr = ['*' | '&' | '-' | '~' | '!'] atom_expr ['?']
func (p *Parser) parseUnaryExpr() (ast.Expr, bool) {
	// TODO: prefix and postfix unary operators
	return p.parseAtomExpr()
}

// -----------------------------------------------------------------------------

// atom_expr = atom {trailer}
// trailer = '(' expr_list ')'
// 	| '{' struct_init '}'
//	| '[' slice_or_index ']'
//  |  '.' ('IDENTIFIER' | 'NUM_LIT' | generic_spec)
func (p *Parser) parseAtomExpr() (ast.Expr, bool) {
	// TODO: {trailer}
	return p.parseAtom()
}

// atom = 'INTLIT' | 'FLOATLIT' | 'NUMLIT' | 'STRINGLIT' | 'RUNELIT'
//   | 'BOOLLIT' | 'IDENTIFIER' | 'NULL' | tupled_expr | sizeof_expr
//   | ...
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
