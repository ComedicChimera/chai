package walk

import (
	"chai/ast"
	"chai/depm"
	"chai/typing"
	"fmt"
	"log"
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
	log.Fatalln("not implemented")
	return false
}
