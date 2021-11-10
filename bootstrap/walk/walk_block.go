package walk

import (
	"chai/ast"
	"chai/depm"
	"chai/typing"
	"fmt"
)

// walkBlock walks a ast.Block node.
func (w *Walker) walkBlock(b *ast.Block) bool {
	for i, stmt := range b.Stmts {
		switch v := stmt.(type) {
		case *ast.VarDecl:
			if !w.walkVarDecl(v) {
				return false
			}
		case *ast.Assign:
			if !w.walkAssign(v) {
				return false
			}
		case *ast.UnaryUpdate:
			// TODO
		default:
			if !w.walkExpr(stmt) {
				return false
			}
		}

		if i == len(b.Stmts)-1 {
			b.SetType(stmt.Type())
		}
	}

	return true
}

// walkVarDecl walks a variable declaration.
func (w *Walker) walkVarDecl(vd *ast.VarDecl) bool {
	for _, varList := range vd.VarLists {
		// handle initializers
		if varList.Initializer != nil {
			if !w.walkExpr(varList.Initializer) {
				return false
			}

			// if there are multiple names in the variable list, then we have
			// tuple unpacking (need to use different checking semantics)
			if len(varList.Names) > 1 {
				// create a tuple type to match against the one returned from
				// the initializer to appropriately extract types
				varTupleTemplate := make(typing.TupleType, len(varList.Names))

				// if the variables have an type label, then that type gets
				// filled in for all of the fields in the template
				if varList.Type != nil {
					for i := range varTupleTemplate {
						varTupleTemplate[i] = varList.Type
					}
				} else {
					// create new type variables to extract the corresponding
					// tuple elements into
					for i := range varTupleTemplate {
						varTupleTemplate[i] = w.solver.NewTypeVar(
							varList.NamePositions[i],
							fmt.Sprintf("{typeof %s}", varList.Names[i]),
						)
					}
				}

				// constrain the tuple returned to match the tuple template
				w.solver.Constrain(varTupleTemplate, varList.Initializer.Type(), varList.Initializer.Position())

				// declare variables according to the fields in the tuple template
				for i, name := range varList.Names {
					// skip `_`
					if name == "_" {
						continue
					}

					if !w.defineLocal(&depm.Symbol{
						Name:        name,
						PkgID:       w.chFile.Parent.ID,
						DefPosition: varList.NamePositions[i],
						Type:        varTupleTemplate[i],
						DefKind:     depm.DKValueDef,
						Mutability:  depm.NeverMutated,
						Public:      false,
					}) {
						return false
					}
				}

				// return early so we don't declare variables multiple times
				return true
			} else {
				if varList.Type != nil {
					w.solver.Constrain(varList.Type, varList.Initializer.Type(), varList.Initializer.Position())
				} else {
					varList.Type = varList.Initializer.Type()
				}
			}
		}

		// declare local variables
		for i, name := range varList.Names {
			// skip `_`
			if name == "_" {
				continue
			}

			if !w.defineLocal(&depm.Symbol{
				Name:        name,
				PkgID:       w.chFile.Parent.ID,
				DefPosition: varList.NamePositions[i],
				Type:        varList.Type,
				DefKind:     depm.DKValueDef,
				Mutability:  depm.NeverMutated,
				Public:      false,
			}) {
				return false
			}
		}
	}

	return true
}

// walkAssign walks an assignment expression
func (w *Walker) walkAssign(asn *ast.Assign) bool {
	// walk all expressions
	for _, rexpr := range asn.RHSExprs {
		if !w.walkExpr(rexpr) {
			return false
		}
	}

	// and assert that all LHS expressions are mutable
	// TODO: handle `_` on the LHS
	for _, lexpr := range asn.LHSExprs {
		if !w.walkExpr(lexpr) {
			return false
		}

		if !w.assertMutable(lexpr) {
			return false
		}
	}

	// TODO: check compound operators

	// if number of variables match, no pattern matching
	if len(asn.LHSExprs) == len(asn.RHSExprs) {
		// constrain LHS and RHS
		for i, rexpr := range asn.RHSExprs {
			w.solver.Constrain(asn.LHSExprs[i].Type(), rexpr.Type(), rexpr.Position())
		}
	} else {
		// pattern matching => one RHS variable
		tupleTemplate := make([]typing.DataType, len(asn.LHSExprs))
		for i, lexpr := range asn.LHSExprs {
			tupleTemplate[i] = lexpr.Type()
		}

		// constrain to fit pattern
		w.solver.Constrain(typing.TupleType(tupleTemplate), asn.RHSExprs[0].Type(), asn.Position())
	}

	return true
}

// assertMutable asserts that a given LHS expression is mutable.
func (w *Walker) assertMutable(expr ast.Expr) bool {
	if expr.Category() == ast.RValue {
		w.reportError(expr.Position(), "cannot mutate an R value")
	}

	// TODO: support other LHS expressions as necessary
	switch v := expr.(type) {
	case *ast.Identifier:
		// we know the symbol exists
		sym, _ := w.lookup(v.Name, nil)

		switch sym.Mutability {
		case depm.NeverMutated:
			sym.Mutability = depm.Mutable
			fallthrough
		case depm.Mutable:
			return true
		case depm.Immutable:
			w.reportError(expr.Position(), "cannot mutate an immutable value")
		}
	}

	// unreachable
	return false
}
