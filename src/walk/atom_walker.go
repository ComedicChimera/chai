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
	atomBranch := branch.BranchAt(0)

	if root, ok := w.walkAtom(atomBranch, yieldsValue); ok {
		for _, trailer := range branch.Content[1:] {
			// determine a position for the whole root
			rootPos := &logging.TextPosition{
				StartLn:  atomBranch.Position().StartLn,
				EndLn:    trailer.Position().EndLn,
				StartCol: atomBranch.Position().StartCol,
				EndCol:   trailer.Position().EndCol,
			}

			// walk the trailer branch
			trailerBranch := trailer.(*syntax.ASTBranch)
			switch trailerBranch.LeafAt(0).Kind {
			case syntax.LPAREN: // function call
				// function takes no arguments
				if trailerBranch.Len() == 2 {
					if call, ok := w.walkFuncCall(root, rootPos, nil); ok {
						root = call
					} else {
						return nil, false
					}
				} else /* function does take arguments */ {
					if call, ok := w.walkFuncCall(root, rootPos, trailerBranch.BranchAt(1)); ok {
						root = call
					} else {
						return nil, false
					}
				}
			}
		}

		// return the fully accumulated root; each trailer wraps around the
		// original root and stores the wrapped value in `root` to be the root
		// of the next trailer
		return root, true
	} else {
		return nil, false
	}
}

// walkFuncCall walks a function call.  The `argsListBranch` can be `nil` if
// there are no arguments to the function
func (w *Walker) walkFuncCall(root sem.HIRExpr, rootPos *logging.TextPosition, argsListBranch *syntax.ASTBranch) (sem.HIRExpr, bool) {
	type positionedArg struct {
		value  sem.HIRExpr
		pos    *logging.TextPosition
		spread bool
	}

	// walk the arguments
	var posArgs []*positionedArg
	namedArgs := make(map[string]*positionedArg)
	receivedSpread := false
	for _, item := range argsListBranch.Content {
		// only branch is `arg`
		if itembranch, ok := item.(*syntax.ASTBranch); ok {
			// spread argument must be last argument in `args_list`
			if receivedSpread {
				w.logError(
					"no argument can follow a spread argument",
					logging.LMKArg,
					itembranch.Position(),
				)

				return nil, false
			}

			// branch length always indicates the kind of argument:
			// 1 => positioned, 2 => positioned spread, 3 => named
			switch itembranch.Len() {
			case 1:
				if len(namedArgs) > 0 {
					w.logError(
						"positioned arguments must come before named arguments",
						logging.LMKArg,
						itembranch.Position(),
					)

					return nil, false
				}

				if argExpr, ok := w.walkExpr(itembranch.BranchAt(0), true); ok {
					posArgs = append(posArgs, &positionedArg{
						value: argExpr,
						pos:   itembranch.Position(),
					})
				} else {
					return nil, false
				}
			case 2:
				if argExpr, ok := w.walkExpr(itembranch.BranchAt(1), true); ok {
					receivedSpread = true

					// TODO: validate that `argExpr` can be spread
					posArgs = append(posArgs, &positionedArg{
						value:  argExpr,
						pos:    itembranch.Position(),
						spread: true,
					})
				} else {
					return nil, false
				}
			case 3:
				if name, ok := w.extractIdentifier(itembranch.BranchAt(0)); ok {
					if argExpr, ok := w.walkExpr(itembranch.BranchAt(2), true); ok {
						if _, ok := namedArgs[name]; ok {
							w.logError(
								fmt.Sprintf("multiple values specified for argument `%s`", name),
								logging.LMKArg,
								itembranch.Position(),
							)
						}

						namedArgs[name] = &positionedArg{
							value: argExpr,
							pos:   itembranch.Content[2].Position(),
						}
					} else {
						return nil, false
					}
				} else {
					w.logError(
						"expected identifier not expression",
						logging.LMKSyntax,
						itembranch.Position(),
					)

					return nil, false
				}
			}
		}
	}

	// check to see if the type of the root is known
	rootType := typing.InnerType(root.Type())
	if ft, ok := rootType.(*typing.FuncType); ok {
		// finalArgs will store the final positioned or named arguments to the
		// function -- will go in HIRApply.Args
		finalArgs := make(map[string]sem.HIRExpr)

		// varArgs are variadic argument values passed directly to the function
		var varArgs []sem.HIRExpr

		// spreadArg is the spread argument value (if one exists)
		var spreadArg sem.HIRExpr

		for i, farg := range ft.Args {
			// handle variadic arguments as they have completely different
			// behavior for regular arguments (in terms of argument checking)
			if farg.Variadic {
				// variadic arguments can't be named
				if namedArg, ok := namedArgs[farg.Name]; ok {
					w.logError(
						"cannot specified named value for variadic argument",
						logging.LMKArg,
						namedArg.pos,
					)
				}

				// collect all the variadic argument elements; we know any
				// spread values will come at the end so we can break as soon as
				// we encounter a spread argument
				var varArgElems []*positionedArg
				for j := i; j < len(posArgs); j++ {
					if posArgs[j].spread {
						// TODO: check that the element types match

						spreadArg = posArgs[j].value
						break
					}

					varArgElems = append(varArgElems, posArgs[j])
				}

				// if there are varArgs, type check them and add them as
				// variadic argument values
				if len(varArgElems) > 0 {
					// create a type variable to store the accumulated type of
					// the variadic arguments; will never be undetermined
					vatv := w.solver.CreateTypeVar(nil, func() {})

					// constrain that type parameter to be exactly equal to the
					// argument type -- all of the other argument values will be
					// subtypes of this parameter.  we use the whole list of
					// variadic arguments as the position for this constraint
					w.solver.AddEqConstraint(farg.Type, vatv, &logging.TextPosition{
						StartLn:  posArgs[i].pos.StartLn,
						StartCol: posArgs[i].pos.StartCol,
						EndLn:    posArgs[i+len(varArgElems)-1].pos.EndLn,
						EndCol:   posArgs[i+len(varArgElems)-1].pos.EndCol,
					})

					varArgs = make([]sem.HIRExpr, len(varArgElems))
					for i, elem := range varArgElems {
						w.solver.AddSubConstraint(vatv, elem.value.Type(), elem.pos)
						varArgs[i] = elem.value
					}
				}

				continue
			}

			// consume positioned arguments first; we know they come before the
			// named arguments
			if i < len(posArgs) {
				posArg := posArgs[i]
				if posArg.spread {
					w.logError(
						"cannot pass spread value to non-variadic argument",
						logging.LMKArg,
						posArg.pos,
					)
				}

				// argument values need only be subtypes of the argument type
				w.solver.AddSubConstraint(farg.Type, posArg.value.Type(), posArg.pos)
				finalArgs[farg.Name] = posArg.value

				// check for duplicate named arguments
				if namedArg, ok := namedArgs[farg.Name]; ok {
					w.logError(
						fmt.Sprintf("multiple values specified for argument: `%s`", farg.Name),
						logging.LMKArg,
						namedArg.pos,
					)

					return nil, false
				}
			} else if namedArg, ok := namedArgs[farg.Name]; ok {
				// we don't need to check for duplicates here since positioned
				// arguments already performed the check for named v. positioned
				// clashes and earlier we checked for name v name clashes

				// same logic as for positioned arguments
				w.solver.AddSubConstraint(farg.Type, namedArg.value.Type(), namedArg.pos)
				finalArgs[farg.Name] = namedArg.value
			} else if !farg.Optional {
				// catch uninitialized required arguments
				w.logError(
					fmt.Sprintf("missing value for required argument: `%s`", farg.Name),
					logging.LMKArg,
					rootPos,
				)

				return nil, false
			}
		}

		return &sem.HIRApply{
			ExprBase: sem.NewExprBase(
				ft.ReturnType,
				sem.RValue,
				false,
			),
			Func:      root,
			Args:      finalArgs,
			VarArgs:   varArgs,
			SpreadArg: spreadArg,
		}, true
	} else if tv, ok := rootType.(*typing.TypeVariable); ok {
		// root type is unknown
		// TODO
		_ = tv
	}

	w.logError(
		fmt.Sprintf("unable to call type of `%s`", rootType.Repr()),
		logging.LMKTyping,
		rootPos,
	)
	return nil, false
}

// extractIdentifier attempts to treat an AST branch as really just an
// identifier.  It returns the extracted identifier if this assertion is true
// and a boolean flag.  It is used primarily for named arguments.
func (w *Walker) extractIdentifier(branch *syntax.ASTBranch) (string, bool) {
	if branch.Len() != 1 {
		return "", false
	}

	if subbranch, ok := branch.Content[0].(*syntax.ASTBranch); ok {
		return w.extractIdentifier(subbranch)
	}

	if leaf := branch.LeafAt(0); leaf.Kind == syntax.IDENTIFIER {
		return leaf.Value, true
	} else {
		return "", false
	}
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
