package syntax

import (
	"chaic/ast"
	"chaic/common"
	"chaic/report"
	"chaic/types"
	"chaic/util"
)

// expr := or_expr ['as' type_label] ;
func (p *Parser) parseExpr() ast.ASTExpr {
	baseExpr := p.parseLeftAssocBinOpExpr(0)

	if p.has(TOK_AS) {
		p.next()

		destTyp := p.parseTypeLabel()

		return &ast.TypeCast{ExprBase: ast.NewTypedExprBase(
			report.NewSpanOver(baseExpr.Span(), p.lookbehind.Span),
			destTyp,
		), SrcExpr: baseExpr}
	}

	return baseExpr
}

// -----------------------------------------------------------------------------

// predTable organizes the operator precedence for binary operators.
var predTable = [][]int{
	{TOK_LAND, TOK_BWAND},
	{TOK_BWXOR},
	{TOK_LOR, TOK_BWOR},
	{TOK_EQ, TOK_NEQ, TOK_LT, TOK_GT, TOK_LTEQ, TOK_GTEQ},
	{TOK_LSHIFT, TOK_RSHIFT},
	{TOK_PLUS, TOK_MINUS},
	{TOK_STAR, TOK_DIV, TOK_MOD},
}

const compOpPredLevel = 3

// or_expr := xor_expr {('||' | '|') xor_expr} ;
// xor_expr := and_expr {'^' and_expr} ;
// and_expr := comp_expr {('&&' | '&') comp_expr} ;
// ... [comp_expr] ...
// shift_expr := arith_expr {('<<' | '>>') arith_expr} ;
// arith_expr := term {('+' | '-') term} ;
// term := factor {('*' | '/' | '%') factor} ;
func (p *Parser) parseLeftAssocBinOpExpr(predLevel int) ast.ASTExpr {
	if predLevel == len(predTable) {
		return p.parsePowerOpExpr()
	} else if predLevel == compOpPredLevel {
		return p.parseCompOpExpr()
	} else {
		lhs := p.parseLeftAssocBinOpExpr(predLevel + 1)

		for util.Contains(predTable[predLevel], p.tok.Kind) {
			opTok := p.tok
			p.next()

			rhs := p.parseLeftAssocBinOpExpr(predLevel + 1)

			lhs = &ast.BinaryOpApp{
				ExprBase: ast.NewExprBase(report.NewSpanOver(lhs.Span(), rhs.Span())),
				Op:       newAppliedOper(opTok),
				LHS:      lhs,
				RHS:      rhs,
			}
		}

		return lhs
	}
}

// comp_expr := shift_expr {('==' | '!=' | '<' | '>' | '<=' | '>=') shift_expr} ;
func (p *Parser) parseCompOpExpr() ast.ASTExpr {
	lhs := p.parseLeftAssocBinOpExpr(compOpPredLevel + 1)
	var prevOperand ast.ASTExpr

	for util.Contains(predTable[compOpPredLevel], p.tok.Kind) {
		opTok := p.tok
		p.next()

		rhs := p.parseLeftAssocBinOpExpr(compOpPredLevel + 1)

		if prevOperand == nil {
			lhs = &ast.BinaryOpApp{
				ExprBase: ast.NewExprBase(report.NewSpanOver(lhs.Span(), rhs.Span())),
				Op:       newAppliedOper(opTok),
				LHS:      lhs,
				RHS:      rhs,
			}
		} else {
			rhs = &ast.BinaryOpApp{
				ExprBase: ast.NewExprBase(report.NewSpanOver(prevOperand.Span(), rhs.Span())),
				Op:       newAppliedOper(opTok),
				LHS:      prevOperand,
				RHS:      rhs,
			}

			lhs = &ast.BinaryOpApp{
				ExprBase: ast.NewExprBase(report.NewSpanOver(lhs.Span(), rhs.Span())),
				Op: &common.AppliedOperator{
					Kind:   TOK_LAND,
					OpRepr: "&&",
					Span:   report.NewSpanOver(lhs.Span(), rhs.Span()),
				},
				LHS: lhs,
				RHS: rhs,
			}
		}

		prevOperand = rhs
	}

	return lhs
}

// factor := {atom_expr '**'} atom_expr ;
func (p *Parser) parsePowerOpExpr() ast.ASTExpr {
	operands := []ast.ASTExpr{p.parseUnaryExpr()}
	var opTokens []*Token

	for p.has(TOK_POW) {
		opTokens = append(opTokens, p.tok)
		p.next()

		operands = append(operands, p.parseUnaryExpr())
	}

	if len(operands) == 1 {
		return operands[0]
	}

	rhs := operands[len(operands)-1]
	for i := len(operands) - 2; i >= 0; i-- {
		lhs := operands[i]

		rhs = &ast.BinaryOpApp{
			ExprBase: ast.NewExprBase(report.NewSpanOver(lhs.Span(), rhs.Span())),
			Op:       newAppliedOper(opTokens[i]),
			LHS:      lhs,
			RHS:      rhs,
		}
	}

	return rhs
}

// unary_expr := ['~' | '!' | '-' | '*' | '&'] atom_expr ;
func (p *Parser) parseUnaryExpr() ast.ASTExpr {
	opTok := p.tok

	switch p.tok.Kind {
	case TOK_COMPL, TOK_NOT, TOK_MINUS:
		p.next()

		operand := p.parseAtomExpr()

		return &ast.UnaryOpApp{
			ExprBase: ast.NewExprBase(report.NewSpanOver(opTok.Span, operand.Span())),
			Op:       newAppliedOper(opTok),
			Operand:  operand,
		}
	case TOK_STAR:
		p.next()

		ptr := p.parseAtomExpr()

		return &ast.Deref{
			ExprBase: ast.NewExprBase(report.NewSpanOver(opTok.Span, ptr.Span())),
			Ptr:      ptr,
		}
	case TOK_BWAND:
		p.next()

		isConst := false
		if p.has(TOK_CONST) {
			p.next()
			isConst = true
		}

		elem := p.parseAtomExpr()

		return &ast.Indirect{
			ExprBase: ast.NewExprBase(report.NewSpanOver(opTok.Span, elem.Span())),
			Elem:     elem,
			Const:    isConst,
		}
	default:
		return p.parseAtomExpr()
	}
}

// -----------------------------------------------------------------------------

// atom_expr := atom {trailer} ;
// trailer := '(' expr_list ')' ;
func (p *Parser) parseAtomExpr() ast.ASTExpr {
	atomExpr := p.parseAtom()

	for {
		switch p.tok.Kind {
		case TOK_LPAREN:
			p.next()

			var args []ast.ASTExpr
			var endSpan *report.TextSpan
			if p.has(TOK_RPAREN) {
				endSpan = p.tok.Span
				p.next()
			} else {
				args = p.parseExprList()
				endSpan = p.want(TOK_RPAREN).Span
			}

			atomExpr = &ast.FuncCall{
				ExprBase: ast.NewExprBase(report.NewSpanOver(atomExpr.Span(), endSpan)),
				Func:     atomExpr,
				Args:     args,
			}
		default:
			return atomExpr
		}
	}
}

// atom := 'IDENT' | 'NUMLIT' | 'INTLIT' | 'FLOATLIT' | 'BOOLLIT' | 'RUNELIT'
// 		| 'null' | '(' expr ')' ;
func (p *Parser) parseAtom() ast.ASTExpr {
	switch p.tok.Kind {
	case TOK_IDENT:
		p.next()

		return &ast.Identifier{
			ASTBase: ast.NewASTBaseOn(p.lookbehind.Span),
			Name:    p.lookbehind.Value,
		}
	case TOK_NUMLIT, TOK_FLOATLIT, TOK_INTLIT:
		p.next()

		return &ast.Literal{
			ExprBase: ast.NewExprBase(p.lookbehind.Span),
			Kind:     p.lookbehind.Kind,
			Text:     p.lookbehind.Value,
		}
	case TOK_BOOLLIT:
		p.next()

		return &ast.Literal{
			ExprBase: ast.NewTypedExprBase(
				p.lookbehind.Span,
				types.PrimTypeBool,
			),
			Kind: TOK_BOOLLIT,
			Text: p.lookbehind.Value,
		}
	case TOK_RUNELIT:
		p.next()

		return &ast.Literal{
			ExprBase: ast.NewTypedExprBase(
				p.lookbehind.Span,
				types.RuneType,
			),
			Kind: TOK_RUNELIT,
			Text: p.lookbehind.Value,
		}
	case TOK_NULL:
		p.next()

		return &ast.Null{
			ExprBase: ast.NewExprBase(p.lookbehind.Span),
		}
	case TOK_LPAREN:
		p.next()

		expr := p.parseExpr()

		p.want(TOK_RPAREN)

		return expr
	default:
		p.reject()
		return nil
	}
}

// -----------------------------------------------------------------------------

// newAppliedOper creates a new applied operator from the given token.
func newAppliedOper(tok *Token) *common.AppliedOperator {
	return &common.AppliedOperator{
		Kind:   tok.Kind,
		OpRepr: tok.Value,
		Span:   tok.Span,
	}
}
