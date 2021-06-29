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
			Mutability: sem.NeverMutated,
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
	return nil, false
}
