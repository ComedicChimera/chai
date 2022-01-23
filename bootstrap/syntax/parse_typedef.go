package syntax

import (
	"chai/ast"
	"chai/depm"
	"chai/report"
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

	// TODO: [generic_tag]

	if !p.assertAndNext(ASSIGN) {
		return nil, false
	}

	switch p.tok.Kind {
	case LBRACE: // struct_body (structure)
		return p.parseStructBody(id.Value, id.Position, annotations, public)
	case NEWLINE: // enum_body (algebraic/hybrid)
	default: // type (type alias)

	}

	// TODO
	return nil, false
}

// struct_body = '{' {struct_field} '}'
// struct_field = ['PUB'] ident_list type_ext [initializer] 'NEWLINE'
func (p *Parser) parseStructBody(name string, namePos *report.TextPosition, annotations map[string]string, public bool) (*ast.StructDef, bool) {
	// '{'
	if !p.assertAndNext(LBRACE) {
		return nil, false
	}

	// collect the data from the body
	st := &typing.StructType{
		NamedType:    typing.NewNamedType(p.chFile.Parent.Name, name, p.chFile.Parent.ID),
		FieldsByName: make(map[string]int),
	}
	fieldInits := make(map[string]ast.Expr)

	for p.got(IDENTIFIER) {
		// ['PUB']
		fieldIsPublic := false
		if p.got(PUB) {
			if !p.next() {
				return nil, false
			}

			fieldIsPublic = true
		}

		// ident_list type_ext
		ids, ok := p.parseIdentList(COMMA)
		if !ok {
			return nil, false
		}

		dt, ok := p.parseTypeExt()
		if !ok {
			return nil, false
		}

		// [initializer]
		var init ast.Expr
		if p.got(COLON) {
			init, ok = p.parseInitializer()
			if !ok {
				return nil, false
			}
		}

		for _, id := range ids {
			// check for duplicate identifiers
			if _, ok := st.FieldsByName[id.Name]; ok {
				p.reportError(id.Pos, "multiple fields declared with name `%s`", id.Name)
				return nil, false
			}

			st.FieldsByName[id.Name] = len(st.Fields)
			st.Fields = append(st.Fields, typing.StructField{
				Name:        id.Name,
				Type:        dt,
				Public:      fieldIsPublic,
				Initialized: init != nil,
			})

			// handle any initializers
			if init != nil {
				fieldInits[id.Name] = init
			}
		}
	}

	// `}`
	if !p.assertAndNext(RBRACE) {
		return nil, false
	}

	// create and declare the struct symbol
	sym := &depm.Symbol{
		Name:        name,
		Pkg:         p.chFile.Parent,
		DefPosition: namePos,
		Type:        st,
		DefKind:     depm.DKTypeDef,
		Mutability:  depm.Immutable,
		Public:      public,
	}

	if !p.defineGlobal(sym) {
		return nil, false
	}

	// build and return the struct definition
	return &ast.StructDef{
		DefBase:    ast.NewDefBase(annotations, public),
		Name:       name,
		Type:       st,
		FieldInits: fieldInits,
	}, true
}
