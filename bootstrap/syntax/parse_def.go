package syntax

import (
	"chaic/ast"
	"chaic/common"
	"chaic/report"
	"chaic/types"
)

// definition := [annot_decl] def_core ;
// def_core := func_def | oper_def ;
func (p *Parser) parseDefinition() {
	var annots map[string]ast.AnnotValue
	if p.has(TOK_ATSIGN) {
		annots = p.parseAnnotDecl()
	}

	var def ast.ASTNode

	switch p.tok.Kind {
	case TOK_FUNC:
		def = p.parseFuncDef(annots)
	case TOK_OPER:
		def = p.parseOperDef(annots)
	case TOK_STRUCT:
		def = p.parseStructDef(annots)
	default:
		p.reject()
	}

	p.chFile.Definitions = append(p.chFile.Definitions, def)
}

// annot_decl := '@' (annot_elem | '[' annot_elem {',' annot_elem} ']') ;
// annot_elem := 'IDENT' ['(' 'STRINGLIT' ')'] ;
func (p *Parser) parseAnnotDecl() map[string]ast.AnnotValue {
	p.want(TOK_ATSIGN)

	multiAnnot := false
	if p.has(TOK_LBRACKET) {
		p.next()
		multiAnnot = true
	}

	annots := make(map[string]ast.AnnotValue)
	for {
		nameTok := p.want(TOK_IDENT)

		var valueTok *Token
		if p.has(TOK_LPAREN) {
			p.next()

			valueTok = p.want(TOK_STRINGLIT)

			p.want(TOK_RPAREN)
		}

		if _, ok := annots[nameTok.Value]; ok {
			p.error(nameTok, "annotation %s specified multiple times", nameTok.Value)
		}

		if valueTok == nil {
			annots[nameTok.Value] = ast.AnnotValue{NameSpan: nameTok.Span}
		} else {
			annots[nameTok.Value] = ast.AnnotValue{
				Value:    valueTok.Value,
				NameSpan: nameTok.Span,
				ValSpan:  valueTok.Span,
			}
		}

		if multiAnnot && p.has(TOK_COMMA) {
			p.next()

			continue
		}

		break
	}

	if multiAnnot {
		p.want(TOK_RBRACKET)
	}

	return annots
}

// -----------------------------------------------------------------------------

// func_def := 'def' 'IDENT' func_signature ;
func (p *Parser) parseFuncDef(annots map[string]ast.AnnotValue) *ast.FuncDef {
	startSpan := p.want(TOK_FUNC).Span

	funcIdent := p.want(TOK_IDENT)

	funcParams, funcType, funcBody := p.parseFuncSignature(false)

	funcSym := &common.Symbol{
		Name:       funcIdent.Value,
		ParentID:   p.chFile.Parent.ID,
		FileNumber: p.chFile.FileNumber,
		DefSpan:    funcIdent.Span,
		Type:       funcType,
		DefKind:    common.DefKindFunc,
		Constant:   true,
	}

	p.defineGlobalSymbol(funcSym)

	return &ast.FuncDef{
		ASTBase:     ast.NewASTBaseOver(startSpan, p.lookbehind.Span),
		Symbol:      funcSym,
		Params:      funcParams,
		Body:        funcBody,
		Annotations: annots,
	}
}

// The table of overloadable operators.  The key is the operator kind and the
// value is a bit pattern indicating the possible arities of the overloaded
// operator.  The bit pattern works as follows: if the value of nth bit from the
// right is 1, then the operator can accept n arguments.
var overloadableOperators = map[int]int{
	TOK_PLUS:   0b10,
	TOK_MINUS:  0b11,
	TOK_STAR:   0b10,
	TOK_DIV:    0b10,
	TOK_MOD:    0b10,
	TOK_POW:    0b10,
	TOK_EQ:     0b10,
	TOK_NEQ:    0b10,
	TOK_LT:     0b10,
	TOK_GT:     0b10,
	TOK_LTEQ:   0b10,
	TOK_GTEQ:   0b10,
	TOK_BWAND:  0b10,
	TOK_BWOR:   0b10,
	TOK_BWXOR:  0b10,
	TOK_LSHIFT: 0b10,
	TOK_RSHIFT: 0b10,
	TOK_LAND:   0b10,
	TOK_LOR:    0b10,
	TOK_NOT:    0b01,
	TOK_COMPL:  0b01,
}

// oper_def := 'oper' '(' operator ')' func_signature ;
// operator := '+' | '-' | '*' | '/' | '%' | '**' | '<' | '>'
//          | '<=' | '>=' | '==' | '!=' | '&' | '|' | '^' |
//          | '>>' | '<<' | '~' | '&&' | '||' | '!' ;
func (p *Parser) parseOperDef(annots map[string]ast.AnnotValue) *ast.OperDef {
	startSpan := p.want(TOK_OPER).Span

	p.want(TOK_LPAREN)

	var operTok *Token
	var arityPattern int
	if _arityPattern, ok := overloadableOperators[p.tok.Kind]; ok {
		operTok = p.tok
		arityPattern = _arityPattern
		p.next()
	} else {
		p.reject()
	}

	p.want(TOK_RPAREN)

	funcParams, funcType, funcBody := p.parseFuncSignature(true)

	// Check the operator arity.
	if arityPattern>>(len(funcParams)-1)&1 == 0 {
		p.recError(
			report.NewSpanOver(funcParams[0].DefSpan, funcParams[len(funcParams)-1].DefSpan),
			"%s operator does not have a form with arity %d",
			operTok.Value,
			len(funcParams),
		)
	}

	overload := &common.OperatorOverload{
		ID:         common.GetNewOverloadID(),
		ParentID:   p.chFile.Parent.ID,
		FileNumber: p.chFile.FileNumber,
		Signature:  funcType,
		DefSpan:    operTok.Span,
	}

	p.defineOperatorOverload(operTok.Kind, operTok.Value, len(funcParams), overload)

	return &ast.OperDef{
		ASTBase:     ast.NewASTBaseOver(startSpan, p.lookbehind.Span),
		OpRepr:      operTok.Value,
		Overload:    overload,
		Params:      funcParams,
		Body:        funcBody,
		Annotations: annots,
	}
}

// func_signature := '(' [func_params] ')' [type_label] (func_body | ';') ;
func (p *Parser) parseFuncSignature(mustHaveParams bool) ([]*common.Symbol, *types.FuncType, ast.ASTNode) {
	p.want(TOK_LPAREN)

	var funcParams []*common.Symbol
	if mustHaveParams || !p.has(TOK_RPAREN) {
		funcParams = p.parseFuncParams()
	}

	p.want(TOK_RPAREN)

	var funcBody ast.ASTNode
	var returnTyp types.Type = types.PrimTypeUnit
	for {
		switch p.tok.Kind {
		case TOK_SEMI:
			p.next()
		case TOK_ASSIGN, TOK_LBRACE:
			funcBody = p.parseFuncBody()
		default:
			returnTyp = p.parseTypeLabel()
			continue
		}

		break
	}

	funcType := &types.FuncType{ReturnType: returnTyp}
	for _, param := range funcParams {
		funcType.ParamTypes = append(funcType.ParamTypes, param.Type)
	}

	return funcParams, funcType, funcBody
}

// func_params := func_param {',' func_param} ;
// func_param := ident_list type_ext ;
func (p *Parser) parseFuncParams() []*common.Symbol {
	var funcParams []*common.Symbol
	paramNames := make(map[string]struct{})

	for {
		idents := p.parseIdentList()
		typ := p.parseTypeExt()

		for _, ident := range idents {
			if _, ok := paramNames[ident.Value]; ok {
				p.error(ident, "multiple parameters named %s", ident.Value)
			} else {
				paramNames[ident.Value] = struct{}{}

				funcParams = append(funcParams, &common.Symbol{
					Name:       ident.Value,
					ParentID:   p.chFile.Parent.ID,
					FileNumber: p.chFile.FileNumber,
					DefSpan:    ident.Span,
					Type:       typ,
					DefKind:    common.DefKindValue,
					Constant:   false,
				})
			}
		}

		if p.has(TOK_COMMA) {
			p.next()

			continue
		}

		break
	}

	return funcParams
}

// func_body := '=' expr ';' | block ;
func (p *Parser) parseFuncBody() ast.ASTNode {
	switch p.tok.Kind {
	case TOK_ASSIGN:
		p.next()

		bodyExpr := p.parseExpr()

		p.want(TOK_SEMI)

		return bodyExpr
	case TOK_LBRACE:
		return p.parseBlock()
	default:
		p.reject()
		return nil
	}
}

/* -------------------------------------------------------------------------- */

// struct_def := 'struct' 'IDENTIFIER' '{' struct_field {struct_field} '}' ;
// struct_field := ident_list type_ext [initializer] ';'
func (p *Parser) parseStructDef(annots map[string]ast.AnnotValue) *ast.StructDef {
	startSpan := p.want(TOK_STRUCT).Span

	nameIdent := p.want(TOK_IDENT)

	p.want(TOK_LBRACE)

	var fields []types.StructField
	fieldIndices := make(map[string]int)
	fieldInits := make(map[string]ast.ASTExpr)
	fieldAnnots := make(map[string]map[string]ast.AnnotValue)
	for {
		var thisFieldAnnots map[string]ast.AnnotValue
		if p.has(TOK_ATSIGN) {
			thisFieldAnnots = p.parseAnnotDecl()
		}

		fieldIdents := p.parseIdentList()
		fieldType := p.parseTypeExt()

		var fieldInit ast.ASTExpr
		if p.has(TOK_ASSIGN) {
			fieldInit = p.parseInitializer()
		}

		p.want(TOK_SEMI)

		for _, fieldIdent := range fieldIdents {
			fieldName := fieldIdent.Value

			if _, ok := fieldIndices[fieldName]; ok {
				p.recError(fieldIdent.Span, "multiple fields named %s", fieldName)
			}

			fieldIndices[fieldName] = len(fields)

			fields = append(fields, types.StructField{
				Name: fieldName,
				Type: fieldType,
			})

			if fieldInit != nil {
				fieldInits[fieldName] = fieldInit
			}

			if thisFieldAnnots != nil {
				fieldAnnots[fieldName] = thisFieldAnnots
			}
		}

		if p.has(TOK_RBRACE) {
			break
		}
	}

	structType := &types.StructType{
		NamedType: types.NamedType{
			Name:      p.chFile.Parent.Name + "." + nameIdent.Value,
			ParentID:  p.chFile.Parent.ID,
			DeclIndex: len(p.chFile.Definitions),
		},
		Fields:  fields,
		Indices: fieldIndices,
	}

	structSym := &common.Symbol{
		Name:       nameIdent.Value,
		ParentID:   p.chFile.Parent.ID,
		FileNumber: p.chFile.FileNumber,
		DefSpan:    nameIdent.Span,
		Type:       structType,
		DefKind:    common.DefKindType,
		Constant:   true,
	}

	p.defineGlobalSymbol(structSym)

	endSpan := p.want(TOK_RBRACE).Span

	return &ast.StructDef{
		ASTBase:     ast.NewASTBaseOver(startSpan, endSpan),
		Symbol:      structSym,
		FieldInits:  fieldInits,
		Annotations: annots,
		FieldAnnots: fieldAnnots,
	}
}
