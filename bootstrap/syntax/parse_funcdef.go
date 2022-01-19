package syntax

import (
	"chai/ast"
	"chai/depm"
	"chai/typing"
)

// func_def = `def` `IDENTIFIER` [generic_tag] `(` [args_decl] `)` [type_label] func_body
func (p *Parser) parseFuncDef(annotations map[string]string, public bool) (ast.Def, bool) {
	if !p.validateAnnotUsage(annotations, "function") {
		return nil, false
	}

	// parse the function name
	if !p.wantAndNext(IDENTIFIER) {
		return nil, false
	}

	funcID := p.lookbehind

	// TODO: generic tag

	if !p.assertAndNext(LPAREN) {
		return nil, false
	}

	// parse args_decl
	var args []*ast.FuncArg
	if !p.got(RPAREN) {
		_args, ok := p.parseArgsDecl()
		if !ok {
			return nil, false
		}

		args = _args
	}

	if !p.newlines() {
		return nil, false
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
	argTypes := make([]typing.DataType, len(args))
	for i, arg := range args {
		argTypes[i] = arg.Type
	}

	ft := &typing.FuncType{
		Args:       argTypes,
		ReturnType: rtType,
	}

	// prepare and define function symbol
	sym := &depm.Symbol{
		Name:        funcID.Value,
		Pkg:         p.chFile.Parent,
		DefPosition: funcID.Position,
		Type:        ft,
		DefKind:     depm.DKFuncDef,
		Mutability:  depm.Immutable,
		Public:      public,
	}

	// if the function is intrinsic, it is added to the universe instead of the
	// global symbol table.  It also has its function type modified to include
	// the intrinsic name
	if _, ok := annotations["intrinsic"]; ok {
		ft.IntrinsicName = sym.Name

		if p.isGloballyDefined(sym.Name) {
			p.reportError(sym.DefPosition, "multiple symbols named `%s` defined in scope", sym.Name)
			return nil, false
		}

		p.uni.IntrinsicFuncs[sym.Name] = sym
	} else /* otherwise, define it globally */ if !p.defineGlobal(sym) {
		return nil, false
	}

	// func body
	funcBody, ok := p.parseFuncBody(annotations)
	if !ok {
		return nil, false
	}

	// make the function AST
	return &ast.FuncDef{
		DefBase:   ast.NewDefBase(annotations, public),
		Name:      sym.Name,
		Signature: ft,
		Args:      args,
		Body:      funcBody,
	}, true
}

// args_decl = arg_decl {',' arg_decl}
// arg_decl = arg_id {',' arg_id} type_ext
func (p *Parser) parseArgsDecl() ([]*ast.FuncArg, bool) {
	takenArgNames := make(map[string]struct{})
	var args []*ast.FuncArg

	for {
		var groupedArgs []*ast.FuncArg

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

			groupedArgs = append(groupedArgs, &ast.FuncArg{
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

	if !p.assertAndNext(IDENTIFIER) {
		return nil, false, false
	}

	return p.lookbehind, byRef, true
}

// func_body = 'end' 'NEWLINE' | block | '=' expr 'NEWLINE'
func (p *Parser) parseFuncBody(annotations map[string]string) (ast.Expr, bool) {
	// determine if the function should or should not have a body based on the
	// provided annotations.
	needsBody := true
	for _, specialName := range []string{"extern", "intrinsic", "dllimport", "intrinsicop"} {
		if _, ok := annotations[specialName]; ok {
			needsBody = false
			break
		}
	}

	switch p.tok.Kind {
	case END:
		if needsBody {
			p.reject()
			return nil, false
		}

		// no body
		if !p.wantAndNext(NEWLINE) {
			return nil, false
		}

		return nil, true
	case ASSIGN:
		if !needsBody {
			p.reject()
			return nil, false
		}

		// pure expression body
		if !p.next() {
			return nil, false
		}

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
		if !needsBody {
			p.reject()
			return nil, false
		}

		block, ok := p.parseBlock()

		// the closing end + final newline
		if ok && p.assertAndNext(END) && p.assertAndNext(NEWLINE) {
			return block, true
		}

		return nil, false
	}

	p.reject()
	return nil, false
}

// -----------------------------------------------------------------------------

// oper_def = 'oper' '(' operator ')' [generic_tag] '(' args_decl ')' type_label func_body
func (p *Parser) parseOperDef(annotations map[string]string, public bool) (ast.Def, bool) {
	if !p.validateAnnotUsage(annotations, "operator") {
		return nil, false
	}

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
	argTypes := make([]typing.DataType, len(args))
	for i, arg := range args {
		argTypes[i] = arg.Type
	}

	ft := &typing.FuncType{
		Args:       argTypes,
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

	// NOTE: overload collisions will be checked later when all types are known
	// (so the equivalency test can work properly)

	// if the operator is intrinsic, it is added to the universe instead of the
	// global symbol table.  It also has its function type modified to include
	// the intrinsic name
	if opIname, ok := annotations["intrinsicop"]; ok {
		ft.IntrinsicName = opIname

		p.addToOperTable(p.uni.IntrinsicOperators, opToken.Kind, opToken.Value, opOverload)
	} else {
		p.addToOperTable(p.chFile.Parent.OperatorTable, opToken.Kind, opToken.Value, opOverload)
	}

	// parse the operator function body
	funcBody, ok := p.parseFuncBody(annotations)
	if !ok {
		return nil, false
	}

	// return the operator AST
	return &ast.OperDef{
		DefBase: ast.NewDefBase(annotations, public),
		Op: &ast.Oper{
			Kind:      opToken.Kind,
			Name:      opToken.Value,
			Pos:       opToken.Position,
			Signature: ft,
		},
		Args: args,
		Body: funcBody,
	}, true
}

// addToOperTable adds an operator to a given operator table.
func (p *Parser) addToOperTable(table map[int]*depm.Operator, kind int, name string, overload *depm.OperatorOverload) {
	if operator, ok := table[kind]; ok {
		operator.Overloads = append(operator.Overloads, overload)
	} else {
		operator := &depm.Operator{
			Pkg:       p.chFile.Parent,
			OpName:    name,
			Overloads: []*depm.OperatorOverload{overload},
		}

		table[kind] = operator
	}
}
