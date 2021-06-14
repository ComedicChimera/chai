package walk

import (
	"chai/logging"
	"chai/sem"
	"chai/syntax"
	"chai/typing"
	"fmt"
)

// WalkDef walks a core definition node (eg. `type_def` or `variable_decl`)
func (w *Walker) WalkDef(branch *syntax.ASTBranch, public bool, annots map[string]*sem.Annotation) bool {
	switch branch.Name {
	case "func_def":
		return w.walkFuncDef(branch, public, annots)
	case "oper_def":
		return w.walkOperDef(branch, public, annots)
	}

	// TODO: handle generics

	return false
}

// -----------------------------------------------------------------------------

// walkOperDef walks an operator definition
func (w *Walker) walkOperDef(branch *syntax.ASTBranch, public bool, annots map[string]*sem.Annotation) bool {
	ft := &typing.FuncType{}
	var body *syntax.ASTBranch

	// we will process the operator value after we will walk the contents of the
	// operator definition.  That way we can check operator definitions against
	// the arity of the operator after we are done with it.
	for _, item := range branch.Content[2:] {
		if itembranch, ok := item.(*syntax.ASTBranch); ok {
			switch itembranch.Name {
			case "generic_tag":
				// TODO
			case "signature":
				if rtType, args, _, ok := w.walkFuncSignature(itembranch, true); ok {
					ft.ReturnType = rtType
					ft.Args = args
				} else {
					return false
				}
			case "decl_func_body":
				body = itembranch
			}
		}
	}

	if intrinsicName, ok := w.validateOperatorAnnotations(annots); ok {
		if intrinsicName == "" && body == nil {
			w.logError(
				"operator definition required a body",
				logging.LMKDef,
				branch.LastBranch().Last().Position(), // error on the parenthesis
			)

			return false
		}

		// maps to the first token of the `oper_value` branch unless the
		// operator is `[:]` in which case it maps to the second token
		operCode := branch.BranchAt(1).LeafAt(branch.BranchAt(1).Len() / 2).Kind

		// validate the operator takes the correct number of arguments
		switch operCode {
		// slice operator is ternary
		case syntax.COLON:
			if len(ft.Args) != 3 {
				w.logError(
					"the slice operator takes 3 arguments",
					logging.LMKDef,
					branch.Content[1].Position(),
				)

				return false
			}
		// minus for negation and for subtraction
		case syntax.MINUS:
			if len(ft.Args) != 1 && len(ft.Args) != 2 {
				w.logError(
					"the `-` operator may take either 1 or 2 arguments",
					logging.LMKDef,
					branch.Content[1].Position(),
				)

				return false
			}
		// not and complement operator
		case syntax.NOT, syntax.COMPL:
			if len(ft.Args) != 1 {
				w.logError(
					fmt.Sprintf("the `%s` operator takes 1 argument", branch.BranchAt(1).LeafAt(0).Value),
					logging.LMKDef,
					branch.Content[1].Position(),
				)

				return false
			}
		// all other operators are binary
		default:
			if len(ft.Args) != 2 {
				w.logError(
					fmt.Sprintf("the `%s` operator takes 2 arguments", branch.BranchAt(1).LeafAt(0).Value),
					logging.LMKDef,
					branch.Content[1].Position(),
				)

				return false
			}
		}

		// check that the operator isn't already defined
		ft.IntrinsicName = intrinsicName
		operSym := &sem.Symbol{
			Type:       ft,
			Public:     public,
			Mutability: sem.Immutable,
			DefKind:    sem.DefKindFuncDef,
			Position:   branch.Content[1].Position(),
		}

		if !w.defineOperator(operCode, operSym, branch.Content[1].Position()) {
			return false
		}

		// create and add the operator definition
		w.SrcFile.AddDefNode(&sem.HIROperDef{
			DefBase:  sem.NewDefBase(operSym, annots),
			OperCode: operCode,
			Body:     (*sem.HIRIncomplete)(body),
		})

	}

	return false
}

// -----------------------------------------------------------------------------

// walkFuncDef walks a function definition
func (w *Walker) walkFuncDef(branch *syntax.ASTBranch, public bool, annots map[string]*sem.Annotation) bool {
	if sym, namePos, argInits, ok := w.walkFuncHeader(branch, public); ok {
		if w.defineGlobal(sym, namePos) {
			needsBody, ok := w.validateFuncAnnotations(annots)
			if !ok {
				return false
			}

			if _, ok := annots["intrinsic"]; ok {
				(sym.Type).(*typing.FuncType).IntrinsicName = sym.Name
			}

			if dfn, ok := branch.Last().(*syntax.ASTBranch); ok {
				w.SrcFile.AddDefNode(&sem.HIRFuncDef{
					DefBase:              sem.NewDefBase(sym, annots),
					ArgumentInitializers: argInits,
					Body:                 (*sem.HIRIncomplete)(dfn),
				})
			} else if needsBody {
				w.logError(
					fmt.Sprintf("function `%s` must provide a body", sym.Name),
					logging.LMKDef,
					branch.LastBranch().Last().Position(), // error on the parenthesis
				)

				return false
			} else {
				w.SrcFile.AddDefNode(&sem.HIRFuncDef{
					DefBase:              sem.NewDefBase(sym, annots),
					ArgumentInitializers: argInits,
					Body:                 nil,
				})
			}

			return true
		}
	}

	return false
}

// walkFuncHeader walks the header (top-branch not `signature`) of a function or
// method and returns an appropriate symbol based on that signature.  It does
// NOT define this symbol.
func (w *Walker) walkFuncHeader(branch *syntax.ASTBranch, public bool) (*sem.Symbol, *logging.TextPosition, map[string]sem.HIRExpr, bool) {
	ft := &typing.FuncType{}
	sym := &sem.Symbol{
		Type:       ft,
		Public:     public,
		SrcPackage: w.SrcFile.Parent,
		Mutability: sem.Immutable,
	}
	var namePos *logging.TextPosition
	var argInits map[string]sem.HIRExpr

	for _, item := range branch.Content {
		switch v := item.(type) {
		case *syntax.ASTBranch:
			switch v.Name {
			case "generic_tag":
				// TODO
			case "signature":
				var ok bool
				ft.ReturnType, ft.Args, argInits, ok = w.walkFuncSignature(v, false)
				if !ok {
					return nil, nil, nil, false
				}
			}
		case *syntax.ASTLeaf:
			switch v.Kind {
			case syntax.IDENTIFIER:
				sym.Name = v.Value
				namePos = v.Position()
			case syntax.ASYNC:
				ft.Async = true
			}
		}
	}

	return sym, namePos, argInits, true
}

// walkFuncSignature walks a `signature` node
func (w *Walker) walkFuncSignature(branch *syntax.ASTBranch, isOperator bool) (typing.DataType, []*typing.FuncArg, map[string]sem.HIRExpr, bool) {
	if branch.Len() == 1 {
		if args, argInits, ok := w.walkFuncArgsDecl(branch.BranchAt(0), isOperator); ok {
			// only arguments => return type of nothing
			return typing.PrimType(typing.PrimKindNothing), args, argInits, true
		}
	} else if rtType, ok := w.walkTypeLabel(branch.BranchAt(1)); ok {
		if args, argInits, ok := w.walkFuncArgsDecl(branch.BranchAt(0), isOperator); ok {
			return rtType, args, argInits, true
		}
	}

	return nil, nil, nil, false
}

// walkFuncArgsDecl walks an `args_decl` or a `next_arg_decl` node and returns a
// list of valid arguments if possible
func (w *Walker) walkFuncArgsDecl(branch *syntax.ASTBranch, isOperator bool) ([]*typing.FuncArg, map[string]sem.HIRExpr, bool) {
	var args []*typing.FuncArg
	argInits := make(map[string]sem.HIRExpr)

	for _, item := range branch.Content {
		if itembranch, ok := item.(*syntax.ASTBranch); ok {
			switch itembranch.Name {
			case "arg_decl":
				arg := &typing.FuncArg{}
				var names []string

				for _, elem := range itembranch.Content {
					if elembranch, ok := elem.(*syntax.ASTBranch); ok {
						switch elembranch.Name {
						case "identifier_list":
							_names, namePos, err := WalkIdentifierList(elembranch)
							if err == nil {
								names = _names

								for _, name := range names {
									argInits[name] = nil
								}
							} else {
								w.logError(
									fmt.Sprintf("multiple arguments named `%s`", err.Error()),
									logging.LMKArg,
									namePos[err.Error()],
								)

								return nil, nil, false
							}
						case "type_ext":
							if dt, ok := w.walkTypeExt(elembranch); ok {
								arg.Type = dt
							} else {
								return nil, nil, false
							}
						case "initializer":
							if isOperator {
								w.logError(
									"operators cannot take optional arguments",
									logging.LMKArg,
									elembranch.Position(),
								)

								return nil, nil, false
							}

							for _, name := range names {
								argInits[name] = (*sem.HIRIncomplete)(itembranch)
							}
						}
					} else {
						// only token in `arg_decl` is vol
						arg.Volatile = true
					}
				}

				for _, name := range names {
					fa := *arg
					fa.Name = name

					args = append(args, &fa)
				}
			case "var_arg_decl":
				if isOperator {
					w.logError(
						"operators cannot take indefinite arguments",
						logging.LMKArg,
						itembranch.Position(),
					)

					return nil, nil, false
				}

				if dt, ok := w.walkTypeExt(itembranch.LastBranch()); ok {
					name := itembranch.LeafAt(1).Value

					if _, ok := argInits[name]; ok {
						w.logError(
							fmt.Sprintf("multiple arguments named `%s`", name),
							logging.LMKArg,
							itembranch.Content[1].Position(),
						)

						return nil, nil, false
					} else {
						argInits[name] = nil
					}

					args = append(args, &typing.FuncArg{
						Name:       name,
						Type:       dt,
						Indefinite: true,
					})
				} else {
					return nil, nil, false
				}
			case "next_arg_decl":
				if nextArgs, newArgInits, ok := w.walkFuncArgsDecl(itembranch, isOperator); ok {
					for argName, argInitExpr := range newArgInits {
						argInits[argName] = argInitExpr
					}

					return append(args, nextArgs...), argInits, true
				} else {
					return nil, nil, false
				}
			}
		}
	}

	// remove empty initializers
	for argName, argInitExpr := range argInits {
		if argInitExpr == nil {
			delete(argInits, argName)
		}
	}

	return args, argInits, true
}
