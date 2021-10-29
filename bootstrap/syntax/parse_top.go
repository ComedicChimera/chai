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

// definition = func_def | oper_def | type_def | space_def | var_def | const_def
func (p *Parser) parseDefinition(public bool) (ast.Def, bool) {
	switch p.tok.Kind {
	case DEF:
		// func_def
		return p.parseFuncDef(public)
	case OPER:
		// oper_def
		return p.parseOperDef(public)
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

// func_def = `def` `IDENTIFIER` [generic_tag] `(` [args_decl] `)` [type_label] func_body
// Semantic Actions: declares function symbol
func (p *Parser) parseFuncDef(public bool) (ast.Def, bool) {
	if !p.want(IDENTIFIER) {
		return nil, false
	}

	funcID := p.tok
	if !p.next() {
		return nil, false
	}

	// TODO: generic tag

	if !p.assertAndNext(LPAREN) {
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

	if !p.assertAndNext(RPAREN) {
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
	funcBody, ok := p.parseFuncBody()
	if !ok {
		return nil, false
	}

	// make the function AST
	return &ast.FuncDef{
		Name: sym.Name,
		// TODO: update to collect annotations from parser
		Annots:    make(map[string]string),
		Signature: ft,
		Body:      funcBody,
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

// func_body = 'end' 'NEWLINE' | block | '=' expr 'NEWLINE'
func (p *Parser) parseFuncBody() (ast.Expr, bool) {
	switch p.tok.Kind {
	case END:
		// no body
		if !p.wantAndNext(NEWLINE) {
			return nil, false
		}

		return nil, true
	case ASSIGN:
		// pure expression body
		p.next()

		// newlines can be placed after equals sign
		if p.got(NEWLINE) {
			if !p.next() {
				return nil, false
			}
		}

		expr, ok := p.parseExpr()
		if !ok || !p.assertAndNext(NEWLINE) {
			return nil, false
		}

		return expr, true
	case NEWLINE:
		// TODO
	}

	p.reject()
	return nil, false
}

// -----------------------------------------------------------------------------

// oper_def = 'oper' '(' operator ')' [generic_tag] '(' args_decl ')' type_label func_body
// Semantic Actions: define operator
func (p *Parser) parseOperDef(public bool) (ast.Def, bool) {
	// operator token
	if !p.wantAndNext(LPAREN) {
		return nil, false
	}

	opToken, ok := p.parseOperator()
	if !ok {
		return nil, false
	}

	if !p.assertAndNext(RPAREN) {
		return nil, false
	}

	// TODO: generic tag

	// arguments
	if !p.assertAndNext(LPAREN) {
		return nil, false
	}

	args, ok := p.parseArgsDecl()
	if !ok {
		return nil, false
	}

	if !p.assertAndNext(RPAREN) {
		return nil, false
	}

	// return type
	rtType, ok := p.parseTypeLabel()
	if !ok {
		return nil, false
	}

	// create the operator function type
	ft := &typing.FuncType{
		Args:       args,
		ReturnType: rtType,
	}

	// check to see if the number of arguments is valid with the known arity of
	// the operator.
	switch opToken.Kind {
	case MINUS:
		if len(args) != 1 && len(args) != 2 {
			p.errorOn(opToken, "the `-` operator accepts 1 or 2 operands not %d", len(args))
			return nil, false
		}
	case COMPL, NOT:
		if len(args) != 1 {
			p.errorOn(opToken, "ths `%s` operator accepts 1 operand not %d", opToken.Value, len(args))
			return nil, false
		}
	default:
		// all other operators are binary
		if len(args) != 2 {
			p.errorOn(opToken, "the `%s` operator accepts 2 operands not %d", opToken.Value, len(args))
			return nil, false
		}
	}

	// create and define the operator overload
	opOverload := &depm.OperatorOverload{
		Signature: ft,
		Context:   p.chFile.Context,
		Position:  opToken.Position,
		Public:    public,
	}

	if operator, ok := p.chFile.Parent.OperatorTable[opToken.Kind]; ok {
		// NOTE: overload collisions will be checked later when all types are
		// known (so the equivalency test can work properly)
		operator.Overloads = append(operator.Overloads, opOverload)
	} else {
		operator := &depm.Operator{
			OpName:    opToken.Value,
			Overloads: []*depm.OperatorOverload{opOverload},
		}

		p.chFile.Parent.OperatorTable[opToken.Kind] = operator
	}

	// parse the operator function body
	funcBody, ok := p.parseFuncBody()
	if !ok {
		return nil, false
	}

	// return the operator AST
	return &ast.OperDef{
		OpKind: opToken.Kind,
		// TODO: update to use annotations from parser
		Annots:    make(map[string]string),
		Signature: ft,
		Body:      funcBody,
	}, true
}
