package syntax

import (
	"chai/ast"
	"chai/typing"
)

// type_def = 'type' 'IDENTIFIER' [generic_tag] '=' type_def_body
// type_def_body = type | struct_body | enum_body
func (p *Parser) parseTypeDef(annotations map[string]string, public bool) (ast.Def, bool) {
	if !p.assertAndNext(TYPE) {
		return nil, false
	}

	id := p.tok
	if !p.assertAndNext(IDENTIFIER) {
		return nil, false
	}

	// TODO: generic tag

	if !p.assertAndNext(ASSIGN) {
		return nil, false
	}

	var dt typing.DataType
	switch p.tok.Kind {
	case LBRACE: // struct_body (structure)
	case NEWLINE: // enum_body (algebraic/hybrid)
	default: // type (type alias)

	}

	_ = id
	_ = dt

	// TODO
	return nil, false
}

// struct_body = '{' {struct_field} '}'
// struct_field = id_list type_ext [initializer] 'NEWLINE'
func (p *Parser) parseStructBody(name string, annotations map[string]string, public bool) (*ast.StructDef, bool) {
	// TODO
	return nil, false
}
