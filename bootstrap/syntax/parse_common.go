package syntax

import (
	"chai/ast"
	"chai/report"
	"chai/typing"
)

// type_ext = ':' type_label
func (p *Parser) parseTypeExt() (typing.DataType, bool) {
	if !p.assertAndAdvance(COLON) {
		return nil, false
	}

	return p.parseTypeLabel()
}

// type_label = ['&'] value_type
func (p *Parser) parseTypeLabel() (typing.DataType, bool) {
	// check for reference types
	if p.got(AMP) {
		if !p.advance() {
			return nil, false
		}

		if vt, ok := p.parseValueType(); ok {
			return &typing.RefType{ElemType: vt}, true
		} else {
			return nil, false
		}
	}

	return p.parseValueType()
}

// value_type = prim_type | named_type | tuple_type
func (p *Parser) parseValueType() (typing.DataType, bool) {
	switch p.tok.Kind {
	case IDENTIFIER:
		// named_type
		return p.parseNamedType()
	case LPAREN:
		return p.parseTupleType()
	default:
		// prim_type
		if U8 <= p.tok.Kind && p.tok.Kind <= NOTHING {
			// use the fact that the token kind is numerically aligned with the
			// different primitive kinds -- just need to remove an offset.
			pt := typing.PrimType(p.tok.Kind - U8)
			return pt, p.next()
		}
	}

	p.reject()
	return nil, false
}

// named_type = 'IDENTIFIER' ['.' 'IDENTIFIER'] [generic_tag]
func (p *Parser) parseNamedType() (typing.DataType, bool) {
	firstIDTok := p.tok

	if !p.next() {
		return nil, false
	}

	if p.got(DOT) {
		if !p.wantAndNext(IDENTIFIER) {
			return nil, false
		}

		return p.res.AddOpaqueTypeRef(
			p.chFile,
			firstIDTok.Value+"."+p.tok.Value,
			report.TextPositionFromRange(firstIDTok.Position, p.tok.Position),
		), true
	} else {
		return p.res.AddOpaqueTypeRef(p.chFile, firstIDTok.Value, firstIDTok.Position), true
	}

	// TODO: generic tag
}

// tuple_type = '(' type_label ',' type_label {',' type_label} ')'
func (p *Parser) parseTupleType() (typing.DataType, bool) {
	if !p.advance() {
		return nil, false
	}

	firstTyp, ok := p.parseTypeLabel()
	if !ok {
		return nil, false
	}

	types := []typing.DataType{firstTyp}
	for p.got(COMMA) {
		if !p.advance() {
			return nil, false
		}

		if nextTyp, ok := p.parseTypeLabel(); ok {
			types = append(types, nextTyp)
		} else {
			return nil, false
		}
	}

	// single type tuples are not allowed
	if len(types) == 1 {
		p.reject()
	}

	if !p.assertAndNext(RPAREN) {
		return nil, false
	}

	return typing.TupleType(types), true
}

// -----------------------------------------------------------------------------

// parseOperator parses any valid operator token and returns it.
func (p *Parser) parseOperator() (*Token, bool) {
	opToken := p.tok
	switch opToken.Kind {
	case PLUS, MINUS, STAR, IDIV, FDIV, MOD, POWER, AND, OR, AMP,
		PIPE, CARRET, COMPL, NOT, EQ, GT, LT, GTEQ, LTEQ, NEQ:
		if p.next() {
			return opToken, true
		}
	}

	return nil, false
}

// initializer = '=' expr
func (p *Parser) parseInitializer() (ast.Expr, bool) {
	if !p.assertAndAdvance(ASSIGN) {
		return nil, false
	}

	return p.parseExpr()
}

// parseIdentList parses a series of identifiers separated by a given separator.
func (p *Parser) parseIdentList(sep int) ([]*ast.Identifier, bool) {
	var idents []*ast.Identifier

	for {
		if !p.assertAndNext(IDENTIFIER) {
			return nil, false
		}

		idents = append(idents, &ast.Identifier{
			ExprBase: ast.NewExprBase(nil, ast.LValue),
			Name:     p.lookbehind.Value,
			Pos:      p.lookbehind.Position,
		})

		if p.got(sep) {
			if !p.advance() {
				return nil, false
			}
		} else {
			break
		}
	}

	return idents, true
}
