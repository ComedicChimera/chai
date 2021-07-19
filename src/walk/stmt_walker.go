package walk

import (
	"chai/logging"
	"chai/sem"
	"chai/syntax"
	"chai/typing"
)

// walkStmt walks an `stmt` node
func (w *Walker) walkStmt(branch *syntax.ASTBranch) (sem.HIRExpr, bool) {
	stmt := branch.BranchAt(0)
	switch stmt.Name {
	case "variable_decl":
		return w.walkVarDecl(stmt, false, sem.ModNone)
	case "break_stmt":
		if w.currExprContext().LoopContext {
			w.updateControl(CFLoop)
			return sem.NewControlStmt(sem.CSBreak), true
		}

		w.logError(
			"unable to use `break` outside of loop",
			logging.LMKUsage,
			stmt.Position(),
		)
	case "continue_stmt":
		if w.currExprContext().LoopContext {
			w.updateControl(CFLoop)
			return sem.NewControlStmt(sem.CSContinue), true
		}

		w.logError(
			"unable to use `continue` outside of loop",
			logging.LMKUsage,
			stmt.Position(),
		)
	case "fallthrough_stmt":
		if w.currExprContext().MatchContext {
			w.updateControl(CFMatch)

			if stmt.Len() == 3 {
				// fallthrough to match
				return sem.NewControlStmt(sem.CSFallMatch), true
			}

			// fallthrough
			return sem.NewControlStmt(sem.CSFallthrough), true
		}

		w.logError(
			"unable to use `fallthrough` outside of match expression",
			logging.LMKUsage,
			stmt.Position(),
		)
	case "return_stmt":
		if len(w.exprContextStack) == 0 {
			w.logError(
				"return used outside of function",
				logging.LMKUsage,
				stmt.Position(),
			)

			return nil, false
		}

		w.updateControl(CFReturn)

		fc := w.currExprContext().FuncContext

		if stmt.Len() == 1 {
			// no return values
			n := makeLiteral(nothingType(), "")
			w.solver.AddEqConstraint(fc.ReturnType, n.Type(), stmt.Position())

			return &sem.HIRReturnStmt{
				Value: n,
			}, true
		} else if exprs, ok := w.walkExprList(stmt.BranchAt(1)); ok {
			if len(exprs) > 1 {
				// TODO
			} else {
				e := exprs[0]
				w.solver.AddSubConstraint(fc.ReturnType, e.Type(), stmt.Content[0].Position())

				return &sem.HIRReturnStmt{
					Value: e,
				}, true
			}
		}
	}

	return nil, false
}

// -----------------------------------------------------------------------------

// walkVarDecl walks a `variable_decl` node
func (w *Walker) walkVarDecl(branch *syntax.ASTBranch, global bool, modifiers int) (*sem.HIRVarDecl, bool) {
	type variable struct {
		dt          typing.DataType
		initializer sem.HIRExpr
		pos         *logging.TextPosition
	}

	variables := make(map[string]*variable)
	symbolMods := 0

	// extract data from variable branch
	for _, item := range branch.Content {
		switch v := item.(type) {
		case *syntax.ASTLeaf:
			if v.Kind == syntax.VOL {
				symbolMods |= sem.ModVolatile
			}
		case *syntax.ASTBranch:
			if v.Name == "var" {
				var names map[string]*logging.TextPosition
				var initializer sem.HIRExpr
				var dt typing.DataType

				for _, elem := range v.Content {
					elembranch := elem.(*syntax.ASTBranch)

					switch elembranch.Name {
					case "identifier_list":
						_, _names, err := WalkIdentifierList(elembranch)
						if err != nil {
							w.logRepeatDef(err.Error(), _names[err.Error()])
							return nil, false
						}

						names = _names
					case "type_ext":
						if _dt, ok := w.walkTypeExt(elembranch); ok {
							dt = _dt
						} else {
							return nil, false
						}
					case "var_initializer":
						if _initializer, ok := w.walkVarInitializer(elembranch); ok {
							initializer = _initializer
						} else {
							return nil, false
						}
					}
				}

				if initializer != nil {
					if dt != nil {
						w.solver.AddSubConstraint(dt, initializer.Type(), v.Last().Position())
					} else {
						dt = initializer.Type()
					}
				} else {
					// known `type_ext` => known type
					// TODO: add nullable constraint
				}

				for name, pos := range names {
					if _, ok := variables[name]; ok {
						w.logRepeatDef(name, pos)
					}

					variables[name] = &variable{
						dt:          dt,
						initializer: initializer,
						pos:         pos,
					}
				}
			} else /* unpack_var; var_initializer is never reached */ {
				// TODO
			}
		}
	}

	// declare variables and create variable decl node
	hirVarDecl := &sem.HIRVarDecl{Initializers: make(map[string]sem.HIRExpr)}
	for name, varData := range variables {
		sym := &sem.Symbol{
			Name:       name,
			SrcPackage: w.SrcFile.Parent,
			Type:       varData.dt,
			DefKind:    sem.DefKindValueDef,
			Modifiers:  symbolMods,
			Immutable:  false,
			Position:   varData.pos,
		}

		if global {
			if !w.defineGlobal(sym) {
				return nil, false
			}
		} else if !w.defineLocal(sym) {
			return nil, false
		}

		hirVarDecl.Variables = append(hirVarDecl.Variables, sym)

		if varData.initializer != nil {
			hirVarDecl.Initializers[name] = varData.initializer
		}
	}

	return hirVarDecl, true
}

// walkVarInitializer walks a `var_initializer` node
func (w *Walker) walkVarInitializer(branch *syntax.ASTBranch) (sem.HIRExpr, bool) {
	if branch.LeafAt(0).Kind == syntax.ASSIGN {
		return w.walkExpr(branch.BranchAt(1), true)
	} else {
		// TODO: handle monad stuff
		return nil, false
	}
}

// -----------------------------------------------------------------------------

// walkExprStmt walks an `expr_stmt` node
func (w *Walker) walkExprStmt(branch *syntax.ASTBranch) (sem.HIRExpr, bool) {
	var lhs, rhs []sem.HIRExpr

	// assignKind should be one of the assignment kinds enumerated in `sem` or
	// `-1` if this is not an assignment (its default value)
	assignKind := -1

	// op is the operator used assignment (if any -- this can be `nil`)
	var op *sem.Operator

	// we want to iterate through the branch backwards since we do not know
	// whether or not this is a mutation or simply an expression statement (such
	// as a function call)
	for i := branch.Len() - 1; i >= 0; i-- {
		switch v := branch.Content[i].(type) {
		case *syntax.ASTBranch:
			switch v.Name {
			case "expr_list":
				// `expr_list` only occurs on the rhs
				if exprs, ok := w.walkExprList(v); ok {
					rhs = exprs
				} else {
					return nil, false
				}
			case "assign_op":
				// compound operator
				if v.Len() == 2 {
					assignKind = sem.AKCompound
					if _op, ok := w.lookupOperator(v.LeafAt(0).Kind, 2); ok {
						op = _op
					} else {
						w.logMissingOpOverload(v.LeafAt(0).Value, 2, v.Content[0].Position())
						return nil, false
					}
				} else if v.LeafAt(0).Kind == syntax.BINDTO {
					assignKind = sem.AKBind
				} else {
					assignKind = sem.AKEq
				}
			case "mut_expr":
				if mutExpr, ok := w.walkMutExpr(v, len(rhs) > 0); ok {
					// push front to preserve the ordering of the expressions
					lhs = append([]sem.HIRExpr{mutExpr}, lhs...)
				} else {
					return nil, false
				}
			}
		case *syntax.ASTLeaf:
			switch v.Kind {
			case syntax.INCREM:
				if _op, ok := w.lookupOperator(syntax.PLUS, 2); ok {
					op = _op
					assignKind = sem.AKUnary
				} else {
					w.logMissingOpOverload("+", 2, v.Position())
					return nil, false
				}
			case syntax.DECREM:
				if _op, ok := w.lookupOperator(syntax.MINUS, 2); ok {
					op = _op
					assignKind = sem.AKUnary
				} else {
					w.logMissingOpOverload("-", 2, v.Position())
					return nil, false
				}
			}
		}
	}

	if len(rhs) == 0 {
		if assignKind == sem.AKUnary {
			// unary assignment
			// TODO: add appropriate operator constraint
			return &sem.HIRAssignment{
				// TODO: should unary assignment yield a value?
				Lhs:        lhs,
				AssignKind: assignKind,
				Oper:       op,
			}, true
		} else {
			// simple expr stmt -- eg. a function call
			return lhs[0], true
		}
	} else if len(lhs) > 1 && len(rhs) == 1 {
		// unpacking
		// TODO
		return nil, false
	} else if len(lhs) == len(rhs) {
		// regular assignment
		for i, v := range lhs {
			w.solver.AddSubConstraint(v.Type(), rhs[i].Type(), branch.LastBranch().Content[i*2].Position())
		}

		return &sem.HIRAssignment{
			Lhs:        lhs,
			Rhs:        rhs,
			AssignKind: assignKind,
			Oper:       op,
		}, true
	} else if len(lhs) > len(rhs) {
		// not unpacking; mismatched value counts
		w.logError(
			"too many values on the left side",
			logging.LMKPattern,
			branch.Position(),
		)

		return nil, false
	} else /* rhs > lhs */ {
		// not unpacking; mismatched value counts
		w.logError(
			"too many values on the right side",
			logging.LMKPattern,
			branch.Position(),
		)

		return nil, false
	}
}

// walkMutExpr walks a `mut_expr` node
func (w *Walker) walkMutExpr(branch *syntax.ASTBranch, isMutated bool) (sem.HIRExpr, bool) {
	var mutValue sem.HIRExpr
	var rootNode syntax.ASTNode

	for i, item := range branch.Content {
		switch v := item.(type) {
		case *syntax.ASTBranch:
			if v.Name == "expr" {
				if expr, ok := w.walkExpr(v, true); ok {
					mutValue = expr
					rootNode = v
				} else {
					return nil, false
				}
			} else /* trailer */ {
				switch v.LeafAt(0).Kind {
				case syntax.LPAREN:
					if call, ok := w.walkFuncCall(mutValue, syntax.TextPositionOfSpan(rootNode, branch.Content[i-1]), v.BranchAt(1)); ok {
						mutValue = call
					} else {
						return nil, false
					}
				}
			}
		case *syntax.ASTLeaf:
			switch v.Kind {
			case syntax.IDENTIFIER:
				if sym, ok := w.lookup(v.Value); ok {
					mutValue = &sem.HIRIdentifier{
						Sym:   sym,
						IdPos: v.Position(),
					}
					rootNode = v
				} else {
					w.logUndefined(sym.Name, sym.Position)
					return nil, false
				}
			case syntax.AWAIT:
				// TODO
			case syntax.STAR:
				// TODO
			}
		}
	}

	if isMutated {
		if mutValue.Category() == sem.RValue {
			w.logError(
				"unable to mutate an R-value",
				logging.LMKImmut,
				branch.Position(),
			)

			return nil, false
		} else if mutValue.Immutable() {
			w.logError(
				"value is immutable",
				logging.LMKImmut,
				branch.Position(),
			)

			return nil, false
		}
	}

	return mutValue, true
}
