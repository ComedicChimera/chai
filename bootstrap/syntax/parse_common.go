package syntax

// type_label = prim_type | ref_type | named_type | tuple_type
// ref_type = '&' (prim_type | named_type | tuple_type)
func (p *Parser) parseTypeLabel() bool {
	// TODO: check for reference types

	switch p.tok.Kind {
	case IDENTIFIER:
		// TODO: named_type
	case LPAREN:
		// TODO: tuple_type
	default:
		// TODO: prim_type
	}

	return false
}
