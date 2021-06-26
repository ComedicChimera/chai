package walk

import (
	"chai/logging"
	"chai/sem"
	"chai/syntax"
	"chai/typing"
	"fmt"
	"strings"
)

// walkAtomExpr walks an `atom_expr` branch
func (w *Walker) walkAtomExpr(branch *syntax.ASTBranch, yieldsValue bool) (sem.HIRExpr, bool) {
	if root, ok := w.walkAtom(branch.BranchAt(0), yieldsValue); ok {
		for _, trailer := range branch.Content[1:] {
			trailerBranch := trailer.(*syntax.ASTBranch)
			switch trailerBranch.LeafAt(0).Kind {
			case syntax.LPAREN: // function call
				// function takes no arguments
				if trailerBranch.Len() == 2 {
					if call, ok := w.walkFuncCall(root, nil); ok {
						root = call
					} else {
						return nil, false
					}
				} else /* function does take arguments */ {
					if call, ok := w.walkFuncCall(root, trailerBranch.BranchAt(1)); ok {
						root = call
					} else {
						return nil, false
					}
				}
			}
		}

		return root, true
	} else {
		return nil, false
	}
}

// walkFuncCall walks a function call.  The `argsListBranch` can be `nil` if
// there are no arguments to the function
func (w *Walker) walkFuncCall(root sem.HIRExpr, argsListBranch *syntax.ASTBranch) (sem.HIRExpr, bool) {
	return nil, false
}

// -----------------------------------------------------------------------------

// walkAtom walks an `atom` branch
func (w *Walker) walkAtom(branch *syntax.ASTBranch, yieldsValue bool) (sem.HIRExpr, bool) {
	switch v := branch.Content[0].(type) {
	case *syntax.ASTLeaf:
		switch v.Kind {
		case syntax.BOOLLIT:
			return makeLiteral(typing.PrimType(typing.PrimKindBool), v.Value), true
		case syntax.STRINGLIT:
			return makeLiteral(typing.PrimType(typing.PrimKindString), v.Value), true
		case syntax.RUNELIT:
			return makeLiteral(typing.PrimType(typing.PrimKindRune), v.Value), true
		case syntax.FLOATLIT:
			// floatlits will never be undeducible because they have a default value
			ftv := w.solver.CreateTypeVar(typing.PrimType(typing.PrimKindF32), func() {})
			w.solver.AddSubConstraint(w.lookupNamedBuiltin("Floating"), ftv, branch.Position())
			return makeLiteral(ftv, v.Value), true
		case syntax.INTLIT:
			return w.makeIntLiteral(v.Value, branch.Position()), true
		case syntax.NULL:
			ntv := w.solver.CreateTypeVar(nil, func() {
				w.logError(
					"unable to deduce type for null",
					logging.LMKTyping,
					branch.Position(),
				)
			})
			// TODO: add nullable constraint/assertion
			return makeLiteral(ntv, v.Value), true
		case syntax.IDENTIFIER:
			if sym, ok := w.lookup(v.Value); ok {
				return &sem.HIRIdentifier{
					Sym:   sym,
					IdPos: branch.Position(),
				}, true
			} else {
				w.logError(
					fmt.Sprintf("undefined symbol: `%s`", v.Value),
					logging.LMKName,
					branch.Position(),
				)

				return nil, false
			}
		}

	case *syntax.ASTBranch:
		switch v.Name {
		case "tupled_expr":
			// if length 2 => `()`
			if v.Len() == 2 {
				return makeLiteral(nothingType(), ""), true
			} else {
				exprList := v.BranchAt(1)
				if exprList.Len() == 1 {
					return w.walkExpr(exprList.BranchAt(0), yieldsValue)
				} else {
					// TODO
				}
			}
		}
	}

	// unreachable/unimplemented
	return nil, false
}

// makeIntLiteral creates a new integral (can be numeric) literal
func (w *Walker) makeIntLiteral(value string, pos *logging.TextPosition) *sem.HIRLiteral {
	// TODO: deduce size based on value

	// intlits will never be undeducible because they have a default value
	if strings.HasSuffix(value, "ul") || strings.HasSuffix(value, "lu") {
		// trim off the suffix
		return makeLiteral(typing.PrimType(typing.PrimKindU64), value[:len(value)-2])
	} else if strings.HasSuffix(value, "l") {
		// trim off the suffix
		value := value[:len(value)-1]

		itv := w.solver.CreateTypeVar(typing.PrimType(typing.PrimKindI64), func() {})

		longCons := &typing.ConstraintSet{
			SrcPackageID: w.SrcFile.Parent.ID,
			Set: []typing.DataType{
				typing.PrimType(typing.PrimKindI64),
				typing.PrimType(typing.PrimKindU64),
			},
		}
		w.solver.AddSubConstraint(longCons, itv, pos)

		return makeLiteral(itv, value)
	} else if strings.HasSuffix(value, "u") {
		// trim off the suffix
		value = value[:len(value)-1]

		itv := w.solver.CreateTypeVar(w.lookupNamedBuiltin("uint"), func() {})

		unsCons := &typing.ConstraintSet{
			SrcPackageID: w.SrcFile.Parent.ID,
			Set: []typing.DataType{
				typing.PrimType(typing.PrimKindU8),
				typing.PrimType(typing.PrimKindU16),
				typing.PrimType(typing.PrimKindU32),
				typing.PrimType(typing.PrimKindU64),
			},
		}
		w.solver.AddSubConstraint(unsCons, itv, pos)

		return makeLiteral(itv, value)
	} else {
		itv := w.solver.CreateTypeVar(w.lookupNamedBuiltin("int"), func() {})
		w.solver.AddSubConstraint(w.lookupNamedBuiltin("Numeric"), itv, pos)
		return makeLiteral(itv, value)
	}
}

// makeLiteral creates a new HIRLiteral
func makeLiteral(dt typing.DataType, value string) *sem.HIRLiteral {
	return &sem.HIRLiteral{
		ExprBase: sem.NewExprBase(
			dt,
			sem.RValue,
			false,
		),
		Value: value,
	}
}
