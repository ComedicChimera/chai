package syntax

import (
	"chai/ast"
	"chai/report"
	"log"
	"strconv"
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
			if !p.advance() {
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
	switch p.tok.Kind {
	case IF, MATCH, FOR, WHILE, DO:
		return p.parseBlockExpr()
	}

	return p.parseSimpleExpr()
}

// simple_expr = or_expr 'as' type_label
func (p *Parser) parseSimpleExpr() (ast.Expr, bool) {
	if expr, ok := p.parseBinOpExpr(); ok {
		if p.got(AS) {
			if !p.advance() {
				return nil, false
			}

			if dest, ok := p.parseTypeLabel(); ok {
				expr = &ast.Cast{
					ExprBase: ast.NewExprBase(dest, expr.Category()),
					Src:      expr,
					Pos: report.TextPositionFromRange(
						expr.Position(),
						p.lookbehind.Position,
					),
				}
			} else {
				return nil, false
			}
		}

		return expr, true
	}

	return nil, false
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

		if !p.advance() {
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
	// check for prefix operators
	var prefixOpTok *Token
	switch p.tok.Kind {
	// TODO: other supported prefix unary operators
	case MINUS, AMP:
		prefixOpTok = p.tok
		if !p.advance() {
			return nil, false
		}
	}

	// parse the atom expression
	expr, ok := p.parseAtomExpr()
	if !ok {
		return nil, false
	}

	// TODO: postfix operators

	// apply operators
	if prefixOpTok != nil {
		switch prefixOpTok.Kind {
		// handle any special operators like referencing and dereferncing
		case AMP:
			// indirection
			return &ast.Indirect{
				ExprBase: ast.NewExprBase(nil, ast.RValue),
				Operand:  expr,
				Pos:      report.TextPositionFromRange(prefixOpTok.Position, expr.Position()),
			}, true
		case STAR:
			// TODO: dereference
		default:
			// regular unary operator
			return &ast.UnaryOp{
				ExprBase: ast.NewExprBase(nil, ast.RValue),
				Operand:  expr,
				Op: ast.Oper{
					Kind: prefixOpTok.Kind,
					Name: prefixOpTok.Value,
					Pos:  prefixOpTok.Position,
				},
				Pos: report.TextPositionFromRange(prefixOpTok.Position, expr.Position()),
			}, true
		}
	}

	// no operator to apply
	return expr, true
}

// -----------------------------------------------------------------------------

// atom_expr = atom {trailer}
// trailer = '(' expr_list ')'
// 	| struct_init
//	| '[' slice_or_index ']'
//  | dot_expr
func (p *Parser) parseAtomExpr() (ast.Expr, bool) {
	if atomExpr, ok := p.parseAtom(); ok {
		switch p.tok.Kind {
		case LPAREN:
			// func call

			if !p.advance() {
				return nil, false
			}

			// handle arguments (including options)
			var args []ast.Expr
			if !p.got(RPAREN) {
				args, ok = p.parseExprList()
				if !ok {
					return nil, false
				}

				// skip newlines at end of a function call
				if !p.newlines() {
					return nil, false
				}
			}

			// assert the closing paren
			if !p.assertAndNext(RPAREN) {
				return nil, false
			}

			atomExpr = &ast.Call{
				ExprBase: ast.NewExprBase(nil, ast.RValue),
				Func:     atomExpr,
				Args:     args,
				Pos: report.TextPositionFromRange(
					atomExpr.Position(),
					p.lookbehind.Position,
				),
			}
		case DOT:
			return p.parseDotExpr(atomExpr)
		case LBRACE:
			return p.parseStructInit(atomExpr)
		}

		return atomExpr, true
	}

	return nil, false
}

// dot_expr = '.' ('IDENTIFIER' | 'INT_LIT' | 'generic_tag)
func (p *Parser) parseDotExpr(rootExpr ast.Expr) (ast.Expr, bool) {
	if !p.advance() {
		return nil, false
	}

	// tuple dot
	if p.got(INTLIT) {
		nTok := p.tok

		nValue, err := strconv.ParseInt(nTok.Value, 0, 32)
		if err != nil {
			log.Fatalln("Failed to convert tuple index:", err)
		}

		return &ast.TupleDot{
			ExprBase:  ast.NewExprBase(nil, rootExpr.Category()),
			Tuple:     rootExpr,
			FieldN:    int(nValue),
			FieldNPos: nTok.Position,
			Pos:       report.TextPositionFromRange(rootExpr.Position(), nTok.Position),
		}, p.next()
	} else if !p.assert(IDENTIFIER) /* named dot */ {
		return nil, false
	}

	idTok := p.tok
	return &ast.Dot{
		ExprBase: ast.NewExprBase(nil, rootExpr.Category()),
		Root:     rootExpr,
		Field:    idTok.Value,
		FieldPos: idTok.Position,
		Pos:      report.TextPositionFromRange(rootExpr.Position(), idTok.Position),
	}, p.next()
}

// struct_init = '{' ['...' expr ',' struct_field_inits | struct_field_inits] '}'
// struct_field_inits = 'IDENTIFIER' initializer {',' 'IDENTIFIER' initializer}
func (p *Parser) parseStructInit(rootExpr ast.Expr) (ast.Expr, bool) {
	if !p.assertAndAdvance(LBRACE) {
		return nil, false
	}

	si := &ast.StructInit{
		ExprBase:   ast.NewExprBase(nil, ast.RValue),
		TypeExpr:   rootExpr,
		FieldInits: make(map[string]ast.FieldInit),
	}

	// if it is struct an empty initializer, we can just return here
	if p.got(RBRACE) {
		return si, true
	}

	// spread initialization
	if p.got(ELLIPSIS) {
		if !p.next() {
			return nil, false
		}

		spreadExpr, ok := p.parseExpr()
		if !ok {
			return nil, false
		}

		si.SpreadInit = spreadExpr

		if !p.assertAndAdvance(COMMA) {
			return nil, false
		}
	}

	// regular initializers
	idTok := p.tok
	if !p.assertAndAdvance(IDENTIFIER) {
		return nil, false
	}

	expr, ok := p.parseInitializer()
	if !ok {
		return nil, false
	}

	si.FieldInits[idTok.Value] = ast.FieldInit{
		NamePos: idTok.Position,
		Init:    expr,
	}

	for p.got(COMMA) {
		if !p.advance() {
			return nil, false
		}

		idTok = p.tok
		if !p.assertAndAdvance(IDENTIFIER) {
			return nil, false
		}

		if _, ok = si.FieldInits[idTok.Value]; ok {
			p.errorOn(idTok, "multiple initializers for the same field: `%s`", idTok.Value)
			return nil, false
		}

		expr, ok = p.parseInitializer()
		if !ok {
			return nil, false
		}

		si.FieldInits[idTok.Value] = ast.FieldInit{
			NamePos: idTok.Position,
			Init:    expr,
		}
	}

	// closing rbrace
	if !p.assertAndNext(RBRACE) {
		return nil, false
	}

	return si, true
}

// -----------------------------------------------------------------------------

// atom = 'INTLIT' | 'FLOATLIT' | 'NUMLIT' | 'STRINGLIT' | 'RUNELIT'
//   | 'BOOLLIT' | 'IDENTIFIER' | 'NULL' | tupled_expr | sizeof_expr
//   | ...
func (p *Parser) parseAtom() (ast.Expr, bool) {
	switch p.tok.Kind {
	case INTLIT, FLOATLIT, NUMLIT, STRINGLIT, RUNELIT, BOOLLIT, NULL:
		p.next()
		return &ast.Literal{
			ExprBase: ast.NewExprBase(nil, ast.RValue),
			Kind:     p.lookbehind.Kind,
			Value:    p.lookbehind.Value,
			Pos:      p.lookbehind.Position,
		}, true
	case IDENTIFIER:
		p.next()
		return &ast.Identifier{
			ExprBase: ast.NewExprBase(nil, ast.LValue),
			Name:     p.lookbehind.Value,
			Pos:      p.lookbehind.Position,
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
	if !p.advance() {
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

	exprs, ok := p.parseExprList()
	if !ok {
		return nil, false
	}

	// skip any newlines at the end of a tuple
	if !p.newlines() {
		return nil, false
	}

	endTok := p.tok
	if !p.assertAndNext(RPAREN) {
		return nil, false
	}

	cat := exprs[0].Category()
	if len(exprs) > 1 {
		cat = ast.RValue
	}

	return &ast.Tuple{
		ExprBase: ast.NewExprBase(nil, cat),
		Exprs:    exprs,
		Pos:      report.TextPositionFromRange(startTok.Position, endTok.Position),
	}, true
}
