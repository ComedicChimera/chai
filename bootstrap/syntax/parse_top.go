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
	if !p.newlines() {
		return nil, false
	}

	// TODO: metadata parsing

	// TODO: import parsing

	// parse definitions until we get an EOF
	var defs []ast.Def
	annotations := make(map[string]string)
	for !p.got(EOF) {
		switch p.tok.Kind {
		case ANNOTSTART:
			// annotations
			if _annots, ok := p.parseAnnotations(); ok {
				annotations = _annots
			} else {
				return nil, false
			}

			// annotations => publics => definition
			fallthrough
		case PUB:
			// TODO: publics
			// publics => definition
			fallthrough
		default:
			// assume definition
			if def, ok := p.parseDefinition(annotations, false); ok {
				defs = append(defs, def)

				// assert newlines after definitions
				if !p.got(EOF) && !p.assertAndNext(NEWLINE) {
					return nil, false
				}
			} else {
				return nil, false
			}

			// skip additional newlines after a definition
			if !p.newlines() {
				return nil, false
			}
		}

		// reset annotations
		if len(annotations) != 0 {
			annotations = make(map[string]string)
		}
	}

	return defs, true
}

// -----------------------------------------------------------------------------

// annotations = '@' (annot_elem | '[' annot_elem ']') 'NEWLINE'
// annotation = 'IDENTIFIER' ['(' 'STRING_LIT' ')']
func (p *Parser) parseAnnotations() (map[string]string, bool) {
	if !p.assertAndNext(ANNOTSTART) {
		return nil, false
	}

	annotations := make(map[string]string)
	expectingMultiple := false

	if p.got(LBRACKET) {
		expectingMultiple = true
		if !p.next() {
			return nil, false
		}
	}

	// parse annotations
	for {
		// collect annotation pieces
		if !p.assertAndNext(IDENTIFIER) {
			return nil, false
		}

		idTok := p.lookbehind

		value := ""
		var valueTok *Token
		if p.got(LPAREN) {
			if !p.want(STRINGLIT) {
				return nil, false
			}

			value = p.tok.Value
			valueTok = p.tok

			if !p.wantAndNext(RPAREN) {
				return nil, false
			}
		}

		// check that the annotation is not duplicated
		if _, ok := annotations[idTok.Value]; ok {
			p.errorOn(idTok, "multiple values specified for annotation `%s`", idTok.Value)
			return nil, false
		}

		// validate special annotation arity
		if specialCfg, ok := specialAnnotations[idTok.Value]; ok {
			if specialCfg.ExpectsValue && value == "" {
				p.errorOn(idTok, "the special annotation `%s` requires a value", idTok.Value)
				return nil, false
			} else if !specialCfg.ExpectsValue && value != "" {
				p.errorOn(valueTok, "the special annotation `%s` doesn't take a value", idTok.Value)
				return nil, false
			}
		}

		// store new annotation
		annotations[idTok.Value] = value

		if expectingMultiple && p.got(COMMA) {
			if !p.next() {
				return nil, false
			}
		} else {
			break
		}
	}

	// make sure multi-annotation are closed and end the annotation
	if expectingMultiple && !p.assertAndNext(RBRACKET) || !p.assertAndNext(NEWLINE) {
		return nil, false
	}

	return annotations, true
}

// specialAnnotConfig specifies what usages are legal for special annotations.
type specialAnnotConfig struct {
	// Usage is a string indicating what type of code object can use this
	// annotation.  Valid values are: `function`, `operator`
	Usage string

	ExpectsValue bool
}

var specialAnnotations = map[string]specialAnnotConfig{
	"callconv":  {Usage: "function", ExpectsValue: true},
	"intrinsic": {Usage: "function", ExpectsValue: false},
	"entry":     {Usage: "function", ExpectsValue: false},
	"extern":    {Usage: "function", ExpectsValue: false},
	"inline":    {Usage: "function", ExpectsValue: false},

	"intrinsicop": {Usage: "operator", ExpectsValue: true},
}

// validateAnnotUsage checks that the given list of annotations is being used on
// the correct type of code object (eg. you can use `intrinsic` on an operator).
// It assumes the parser is currently positioned on a suitable token to error
// upon.
func (p *Parser) validateAnnotUsage(annotations map[string]string, usage string) bool {
	for name := range annotations {
		if specialCfg, ok := specialAnnotations[name]; ok {
			if specialCfg.Usage != usage {
				p.rejectWithMsg("the special annotation `%s` can only be applied to %s", name, specialCfg.Usage)
				return false
			}
		}
	}

	return true
}

// -----------------------------------------------------------------------------

// definition = func_def | oper_def | type_def | space_def | var_def | const_def
func (p *Parser) parseDefinition(annotations map[string]string, public bool) (ast.Def, bool) {
	switch p.tok.Kind {
	case DEF:
		// func_def
		return p.parseFuncDef(annotations, public)
	case OPER:
		// oper_def
		return p.parseOperDef(annotations, public)
	case TYPE:
		// TODO: type_def
	case SPACE:
		// TODO: space_def
	case LET:
		// global_var_def
		{
			vd, ok := p.parseVarDecl(true, annotations, public)
			return vd.(*ast.VarDecl), ok
		}
	case CONST:
		// TODO: const_def
	}

	p.reject()
	return nil, false
}

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

	intrinsicName := ""
	if _, ok := annotations["intrinsic"]; ok {
		intrinsicName = funcID.Value
	}

	ft := &typing.FuncType{
		Args:          argTypes,
		ReturnType:    rtType,
		IntrinsicName: intrinsicName,
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

	intrinsicName := ""
	if opIname, ok := annotations["intrinsicop"]; ok {
		intrinsicName = opIname
	}

	ft := &typing.FuncType{
		Args:          argTypes,
		ReturnType:    rtType,
		IntrinsicName: intrinsicName,
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
