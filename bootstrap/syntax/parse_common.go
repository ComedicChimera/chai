package syntax

import "chai/typing"

// type_ext = ':' type_label
func (p *Parser) parseTypeExt() (typing.DataType, bool) {
	if !p.assert(COLON) || !p.next() {
		return nil, false
	}

	return p.parseTypeLabel()
}

// type_label = prim_type | ref_type | named_type | tuple_type
// ref_type = '&' (prim_type | named_type | tuple_type)
func (p *Parser) parseTypeLabel() (typing.DataType, bool) {
	// TODO: check for reference types

	switch p.tok.Kind {
	case IDENTIFIER:
		// TODO: named_type
	case LPAREN:
		return p.parseTupleType()
	default:
		// prim_type
		if U8 <= p.tok.Kind && p.tok.Kind <= NOTHING {
			if !p.next() {
				return nil, false
			}

			// use the fact that the token kind is numerically aligned with the
			// different primitive kinds -- just need to remove an offset.
			return typing.PrimType(p.tok.Kind - U8), true
		}
	}

	p.reject()
	return nil, false
}

// tuple_type = '(' type_label ',' type_label {',' type_label} ')'
func (p *Parser) parseTupleType() (typing.DataType, bool) {
	if !p.next() {
		return nil, false
	}

	firstTyp, ok := p.parseTypeLabel()
	if !ok {
		return nil, false
	}

	types := []typing.DataType{firstTyp}
	for p.got(COMMA) {
		if !p.next() {
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
