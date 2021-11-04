package syntax

import "chai/ast"

// block = 'NEWLINE' {stmt 'NEWLINE'} 'end'
// stmt = var_decl | const_decl | block_expr | expr_stmt
func (p *Parser) parseBlock() (ast.Expr, bool) {
	if !p.assertAndNext(NEWLINE) {
		return nil, false
	}

	var stmts []ast.Expr

	// skip leading newlines

	for {

		switch p.tok.Kind {
		case LET:
			// var_decl
			if stmt, ok := p.parseVarDecl(false); ok {
				stmts = append(stmts, stmt)
			} else {
				return nil, false
			}
		case CONST:
			// TODO: const_decl
		case IF, MATCH, FOR, WHILE, DO:
			// TODO: block_expr
		default:
			// TODO: expr_stmt
		}

		if p.assertAndNext(NEWLINE) {
			// skip newlines between statements
			p.newlines()

			// exit condition: `end`
			if p.got(END) {
				if !p.next() {
					return nil, false
				}

				break
			}
		} else {
			return nil, false
		}
	}

	return &ast.Block{
		Stmts: stmts,
	}, true
}

// var_decl = 'let' var {',' var}
// var = id_list (type_ext [initializer] | initializer)
func (p *Parser) parseVarDecl(global bool) (ast.Expr, bool) {
	if !p.assertAndNext(LET) {
		return nil, false
	}

	for {
		// TODO: parse `var`

		// `,` between successive statements
		if p.got(COMMA) {
			if !p.next() {
				return nil, false
			}
		} else {
			break
		}
	}

	// TODO: return produced variable declaration
	return nil, false
}
