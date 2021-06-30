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
		return w.walkVarDecl(branch, false)
	}

	return nil, false
}

// -----------------------------------------------------------------------------

// walkVarDecl walks a `variable_decl` node
func (w *Walker) walkVarDecl(branch *syntax.ASTBranch, global bool) (sem.HIRExpr, bool) {
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

		if !w.defineLocal(sym) {
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
	for i := branch.Len() - 1; i >= 0; i++ {
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
				if mutExpr, ok := w.walkMutExpr(v); ok {
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
		// unary assignment
		w.solver.AddSubConstraint(w.lookupNamedBuiltin("Numeric"), lhs[0].Type(), branch.Position())
		return &sem.HIRAssignment{
			// TODO: should unary assignment yield a value?
			Lhs:        lhs,
			AssignKind: assignKind,
			Oper:       op,
		}, true
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
func (w *Walker) walkMutExpr(branch *syntax.ASTBranch) (sem.HIRExpr, bool) {
	var mutValue sem.HIRExpr

	for _, item := range branch.Content {
		switch v := item.(type) {
		case *syntax.ASTBranch:
			if v.Name == "expr" {
				if expr, ok := w.walkExpr(v, true); ok {
					mutValue = expr
				} else {
					return nil, false
				}
			} else /* trailer */ {
				// TODO
			}
		case *syntax.ASTLeaf:
			switch v.Kind {
			case syntax.IDENTIFIER:
				if sym, ok := w.lookup(v.Value); ok {
					mutValue = &sem.HIRIdentifier{
						Sym:   sym,
						IdPos: v.Position(),
					}
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

	return mutValue, true
}
