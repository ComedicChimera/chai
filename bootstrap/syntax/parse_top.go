package syntax

import (
	"chai/ast"
	"chai/depm"
	"chai/typing"
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
// function arguments, evaluates type labels for arguments and return type,
// [maybe] pushes and pops a generic context.
func (p *Parser) parseFuncDef(public bool) (ast.Def, bool) {
	if !p.want(IDENTIFIER) {
		return nil, false
	}

	funcID := p.tok

	// TODO: generic tag

	if !p.want(LPAREN) || !p.next() {
		return nil, false
	}

	// parse args_decl
	var args []typing.FuncArg
	if !p.got(RPAREN) {
		_args, ok := p.parseArgsDecl()
		if !ok {
			return nil, false
		}

		args = _args
	}

	if !p.assert(RPAREN) || !p.next() {
		return nil, false
	}

	// type_label (if we don't have an ending token)
	var rtType typing.DataType
	if !p.gotOneOf(END, NEWLINE, ASSIGN) {
		if dt, ok := p.parseTypeLabel(); ok {
			rtType = dt
		} else {
			return nil, false
		}
	} else {
		// no return type => nothing type
		rtType = typing.PrimType(typing.PrimNothing)
	}

	// make the function type
	ft := &typing.FuncType{
		Args:       args,
		ReturnType: rtType,
	}

	// prepare and declare function symbol
	sym := &depm.Symbol{
		Name:        funcID.Value,
		PkgID:       p.chFile.Parent.ID,
		DefPosition: funcID.Position,
		Type:        ft,
		DefKind:     depm.DKValueDef,
		Mutability:  depm.Immutable,
		Public:      public,
	}

	if !p.defineGlobal(sym) {
		return nil, false
	}

	// func body
	var funcBody ast.Expr
	switch p.tok.Kind {
	case END:
		// no body
		if !p.want(NEWLINE) || !p.next() {
			return nil, false
		}
	case ASSIGN:
		// TODO
	case NEWLINE:
		// TODO
	}

	// make the function AST
	return &ast.FuncDef{
		Name: sym.Name,
		// TODO: update to collect annotations from parser
		Annots:   make(map[string]string),
		FuncType: ft,
		Body:     funcBody,
	}, true
}

// args_decl = arg_decl {',' arg_decl}
// arg_decl = arg_id {',' arg_id} type_ext
// Semantic Actions: check for argument name conflicts
func (p *Parser) parseArgsDecl() ([]typing.FuncArg, bool) {
	takenArgNames := make(map[string]struct{})
	var args []typing.FuncArg

	for {
		var groupedArgs []typing.FuncArg

		for {
			argId, byRef, ok := p.parseArgID()
			if !ok {
				return nil, false
			}

			if _, ok := takenArgNames[argId.Value]; ok {
				p.errorOn(argId, "multiple arguments named `%s`", argId.Value)
			} else {
				takenArgNames[argId.Value] = struct{}{}
			}

			groupedArgs = append(groupedArgs, typing.FuncArg{
				Name:  argId.Value,
				ByRef: byRef,
			})

			if p.got(COMMA) {
				if !p.next() {
					return nil, false
				}
			} else {
				break
			}
		}

		if dt, ok := p.parseTypeExt(); ok {
			for _, arg := range groupedArgs {
				arg.Type = dt
				args = append(args, arg)
			}
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

	return args, true
}

// arg_id = ['&'] 'IDENTIFIER'
func (p *Parser) parseArgID() (*Token, bool, bool) {
	byRef := false
	if p.got(AMP) {
		byRef = true

		if !p.next() {
			return nil, false, false
		}
	}

	if !p.assert(IDENTIFIER) {
		return nil, false, false
	}

	idTok := p.tok
	if !p.next() {
		return nil, false, false
	}

	return idTok, byRef, true
}
