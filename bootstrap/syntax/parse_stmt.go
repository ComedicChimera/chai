package syntax

import (
	"chai/ast"
	"chai/depm"
	"chai/report"
	"chai/typing"
)

// var_decl = 'let' var {',' var}
// var = id_list (type_ext [initializer] | initializer)
func (p *Parser) parseVarDecl(global bool, annotations map[string]string, public bool) (ast.Expr, bool) {
	letTokPos := p.tok.Position

	if !p.assertAndNext(LET) {
		return nil, false
	}

	var varLists []*ast.VarList

	for {
		// parse var
		idents, ok := p.parseIdentList(COMMA)
		if !ok {
			return nil, false
		}

		// type extension: only optional if the variable is local
		var varType typing.DataType
		if p.got(COLON) {
			varType, ok = p.parseTypeExt()
			if !ok {
				return nil, false
			}
		} else if global {
			p.rejectWithMsg("global variables must have a type extension")
		}

		// initializer: only optional if there is already a variable type
		var initializer ast.Expr
		if p.got(ASSIGN) {
			initializer, ok = p.parseInitializer()
			if !ok {
				return nil, false
			}
		} else if varType == nil {
			// variable must either have a type or an initializer
			p.reject()
			return nil, false
		}

		// declare the variable symbols if the variable is global
		if global {
			for _, ident := range idents {
				if !p.defineGlobal(&depm.Symbol{
					Name:        ident.Name,
					PkgID:       p.chFile.Parent.ID,
					DefPosition: ident.Pos,
					Type:        varType, // will never be `nil`
					DefKind:     depm.DKValueDef,
					// global variables are always mutable since they need to be mutated by their initializer (run in `init`)
					Mutability: depm.Mutable,
					Public:     public,
				}) {
					return nil, false
				}
			}
		}

		// build the var list
		varNames := make([]string, len(idents))
		varPositions := make([]*report.TextPosition, len(idents))
		for i, ident := range idents {
			varNames[i] = ident.Name
			varPositions[i] = ident.Pos
		}

		varLists = append(varLists, &ast.VarList{
			Names:         varNames,
			NamePositions: varPositions,
			// these will be filled in during walking
			Mutabilities: make([]int, len(varNames)),
			Type:         varType,
			Initializer:  initializer,
		})

		// `,` between successive statements
		if p.got(COMMA) {
			if !p.next() {
				return nil, false
			}
		} else {
			break
		}
	}

	// calculate the end of the variable declaration
	lastList := varLists[len(varLists)-1]
	var lastPos *report.TextPosition
	if lastList.Initializer != nil {
		lastPos = lastList.Initializer.Position()
	} else {
		lastPos = lastList.NamePositions[len(lastList.NamePositions)-1]
	}

	// return produced variable declaration
	return &ast.VarDecl{
		ExprBase: ast.NewExprBase(typing.PrimType(typing.PrimNothing), ast.RValue),
		DefBase:  ast.NewDefBase(annotations, public),
		VarLists: varLists,
		Pos:      report.TextPositionFromRange(letTokPos, lastPos),
	}, true
}

// expr_stmt = lhs_expr {',' lhs_expr} [[operator] '=' expr_list | '++' | '--']
// lhs_expr = ['*'] atom_expr
func (p *Parser) parseExprStmt() (ast.Expr, bool) {
	// collect the LHS expressions
	var lhsExprs []ast.Expr
	for {
		// TODO: ['*']

		atomExpr, ok := p.parseAtomExpr()
		if ok {
			lhsExprs = append(lhsExprs, atomExpr)
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

	// collect the operator
	var asnOp *Token
	var compoundOp *Token
	switch p.tok.Kind {
	case ASSIGN, INCREM, DECREM:
		asnOp = p.tok
		if !p.next() {
			return nil, false
		}
	case NEWLINE:
		// end of line -- standalone expression
		asnOp = p.tok
	default:
		if _cpdOp, ok := p.parseOperator(); ok {
			compoundOp = _cpdOp
		} else {
			return nil, false
		}

		asnOp = p.tok
		if !p.assertAndNext(ASSIGN) {
			return nil, false
		}
	}

	// determine if expressions should be collected
	var rhsExprs []ast.Expr
	if asnOp.Kind == ASSIGN {
		if _rhsExprs, ok := p.parseExprList(); ok {
			rhsExprs = _rhsExprs
		} else {
			return nil, false
		}
	}

	// check that the number of expressions is appropriate for the operator and
	// build the appropriate expression to return
	switch asnOp.Kind {
	case NEWLINE:
		if len(lhsExprs) > 1 {
			p.reportError(lhsExprs[1].Position(), "more than one expression on a line without assignment")
			return nil, false
		}

		// just return the expression raw in this case
		return lhsExprs[0], true
	case INCREM, DECREM:
		if len(lhsExprs) > 1 {
			p.reportError(lhsExprs[1].Position(), "the `%s` operator may only be applied to one operand", asnOp.Value)
		}

		return &ast.UnaryUpdate{
			ExprBase: ast.NewExprBase(typing.PrimType(typing.PrimNothing), ast.RValue),
			Operand:  lhsExprs[0],
			Oper: &ast.Oper{
				Kind: asnOp.Kind,
				Name: asnOp.Value,
				Pos:  asnOp.Position,
			},
		}, true
	default: // ASSIGN
		// if there is more than one rhsExpr, we do a linear match for the
		// number of expressions on each side: ie. no tuple unpacking allowed.
		// Otherwise, we perform tuple pattern matching during walking.
		if len(rhsExprs) > 1 {
			if len(lhsExprs) != len(rhsExprs) {
				p.reportError(
					report.TextPositionFromRange(lhsExprs[0].Position(), rhsExprs[len(rhsExprs)-1].Position()),
					"cannot assign %d values to %d locations",
					len(rhsExprs),
					len(lhsExprs),
				)

				return nil, false
			}
		}

		// build the final assignment
		var compoundASTOper *ast.Oper
		if compoundOp != nil {
			compoundASTOper = &ast.Oper{
				Kind: compoundOp.Kind,
				Name: compoundOp.Value,
				Pos:  compoundOp.Position,
			}
		}

		return &ast.Assign{
			ExprBase: ast.NewExprBase(typing.PrimType(typing.PrimNothing), ast.RValue),
			LHSExprs: lhsExprs,
			RHSExprs: rhsExprs,
			Oper:     compoundASTOper,
		}, true
	}
}
