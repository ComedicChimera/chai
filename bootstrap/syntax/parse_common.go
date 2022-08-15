package syntax

import (
	"chaic/ast"
	"chaic/types"
)

// initializer := '=' expr ;
func (p *Parser) parseInitializer() ast.ASTExpr {
	p.want(TOK_ASSIGN)

	return p.parseExpr()
}

// expr_list := expr {',' expr} ;
func (p *Parser) parseExprList() []ast.ASTExpr {
	exprList := []ast.ASTExpr{p.parseExpr()}

	for p.has(TOK_COMMA) {
		p.next()

		exprList = append(exprList, p.parseExpr())
	}

	return exprList
}

// ident_list := 'IDENT' {',' IDENT} ;
func (p *Parser) parseIdentList() []*Token {
	identList := []*Token{p.want(TOK_IDENT)}

	for p.has(TOK_COMMA) {
		p.next()

		identList = append(identList, p.want(TOK_IDENT))
	}

	return identList
}

// type_ext := ':' type_label ;
func (p *Parser) parseTypeExt() types.Type {
	p.want(TOK_COLON)
	return p.parseTypeLabel()
}

// type_label := prim_type | pointer_type | named_type ;
// prim_type := 'i8' | 'u8' | 'i16' | 'u16' | 'i32' | 'u32'
//           | 'i64' | 'u64' | 'f32' | 'f64' | 'bool' | `unit` ;
// pointer_type := '*' ['const'] type_label ;
// named_type := 'IDENT' ['.' 'IDENT'] ;
func (p *Parser) parseTypeLabel() types.Type {
	switch p.tok.Kind {
	case TOK_I8:
		p.next()
		return types.PrimTypeI8
	case TOK_U8:
		p.next()
		return types.PrimTypeU8
	case TOK_I16:
		p.next()
		return types.PrimTypeI16
	case TOK_U16:
		p.next()
		return types.PrimTypeU16
	case TOK_I32:
		p.next()
		return types.PrimTypeI32
	case TOK_U32:
		p.next()
		return types.PrimTypeU32
	case TOK_I64:
		p.next()
		return types.PrimTypeI64
	case TOK_U64:
		p.next()
		return types.PrimTypeU64
	case TOK_F32:
		p.next()
		return types.PrimTypeF32
	case TOK_F64:
		p.next()
		return types.PrimTypeF64
	case TOK_BOOL:
		p.next()
		return types.PrimTypeBool
	case TOK_UNIT:
		p.next()
		return types.PrimTypeUnit
	case TOK_STAR:
		p.next()

		if p.has(TOK_CONST) {
			p.next()

			return &types.PointerType{
				ElemType: p.parseTypeLabel(),
				Const:    true,
			}
		} else {
			return &types.PointerType{
				ElemType: p.parseTypeLabel(),
				Const:    false,
			}
		}
	case TOK_IDENT:
		{
			typeIdent := p.tok
			p.next()

			// TODO: handle `.` package accesses
			parentPkg := p.chFile.Parent

			// TODO: add the opaque type to the resolver
			return &types.OpaqueType{
				NamedType: types.NamedType{
					Name:     typeIdent.Value,
					ParentID: parentPkg.ID,
				},
				Span: typeIdent.Span,
			}
		}
	default:
		p.reject()
		return nil
	}
}
