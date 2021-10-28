package syntax

import (
	"chai/ast"
	"chai/depm"
)

// parseFile parses the top level of a file.  This is the grammatical start
// symbol for the parser.
func (p *Parser) parseFile() ([]ast.Def, bool) {
	// skip leading newlines
	p.newlines()

	// TODO: metadata parsing

	// TODO: import parsing

	// parse definitions until we get an EOF
	var defs []ast.Def
	for !p.got(EOF) {
		switch p.tok.Kind {
		case PUB:
			// TODO: publics
		case ANNOTSTART:
			// TODO: annotations
		default:
			// assume definition
			if def, ok := p.parseDefinition(false); ok {
				defs = append(defs, def)
			} else {
				return nil, false
			}

			// skip newlines after a definition
			p.newlines()
		}
	}

	return defs, true
}

// definition = func_def | type_def | space_def | var_def | const_def
func (p *Parser) parseDefinition(public bool) (ast.Def, bool) {
	switch p.tok.Kind {
	case DEF:
		// func_def
		return p.parseFuncDef(public)
	case TYPE:
		// TODO: type_def
	case SPACE:
		// TODO: space_def
	case LET:
		// TODO: var_def
	case CONST:
		// TODO: const_def
	}

	p.reject()
	return nil, false
}

// func_def = `def` `IDENTIFIER` [generic_tag] `(` args_decl `)` [type_label] func_body
// Semantic Actions: declares function symbol, pushes and pops scope for
// function arguments, checks that argument names don't conflict, evaluates type
// labels for arguments and return type, [maybe] pushes and pops a generic
// context.
func (p *Parser) parseFuncDef(public bool) (ast.Def, bool) {
	if !p.want(IDENTIFIER) {
		return nil, false
	}

	funcID := p.tok

	// TODO: generic tag

	if !p.want(LPAREN) || !p.next() {
		return nil, false
	}

	// TODO: parse args_decl

	if !p.assert(RPAREN) {
		return nil, false
	}

	// TODO: type_label

	// prepare and declare function symbol
	sym := &depm.Symbol{
		Name:        funcID.Value,
		PkgID:       p.chFile.Parent.ID,
		DefPosition: funcID.Position,
		DefKind:     depm.DKValueDef,
		Mutability:  depm.Immutable,
		Public:      public,
	}

	if !p.defineGlobal(sym) {
		return nil, false
	}

	// TODO: func body

	// make the function AST
	return &ast.FuncDef{
		Name: sym.Name,
		// TODO: update to collect annotations from parser
		Annots: make(map[string]string),
		// TODO: rest
	}, true
}
