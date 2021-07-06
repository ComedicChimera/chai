package walk

import (
	"chai/logging"
	"chai/sem"
	"chai/syntax"
	"chai/typing"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// WalkDef walks a core definition node (eg. `type_def` or `variable_decl`)
func (w *Walker) WalkDef(branch *syntax.ASTBranch, symbolModifiers int, annots map[string]*sem.Annotation) bool {
	switch branch.Name {
	case "func_def":
		return w.walkFuncDef(branch, symbolModifiers, annots)
	case "oper_def":
		return w.walkOperDef(branch, symbolModifiers, annots)
	case "type_def":
		return w.walkTypeDef(branch, symbolModifiers, annots)
	case "variable_decl":
		if varDecl, ok := w.walkVarDecl(branch, true, symbolModifiers); ok {
			w.SrcFile.Root.Globals = append(w.SrcFile.Root.Globals, varDecl)
			return true
		}
	}

	// TODO: handle generics

	return false
}

// -----------------------------------------------------------------------------

// walkTypeDef walks a type definition
func (w *Walker) walkTypeDef(branch *syntax.ASTBranch, symbolModifiers int, annots map[string]*sem.Annotation) bool {
	sym := &sem.Symbol{
		DefKind:    sem.DefKindTypeDef,
		Modifiers:  symbolModifiers,
		Immutable:  true,
		SrcPackage: w.SrcFile.Parent,
	}
	// fieldInits := make(map[string]sem.HIRExpr)
	// closed := false
	// expectingInherit := false
	// var inheritPos *logging.TextPosition

	// TODO: declare self types

	for _, item := range branch.Content {
		switch v := item.(type) {
		case *syntax.ASTBranch:
			switch v.Name {
			case "generic_tag":
				// TODO
			case "type":
				// TODO: handle inherits

				if dt, ok := w.walkTypeLabel(v); ok {
					sym.Type = &typing.AliasType{
						Name:         sym.Name,
						SrcPackageID: sym.SrcPackage.ID,
						Type:         dt,
					}
				} else {
					return false
				}
			case "newtype":
				// TODO
			}
		case *syntax.ASTLeaf:
			switch v.Kind {
			case syntax.IDENTIFIER:
				sym.Name = v.Value
				sym.Position = v.Position()
			case syntax.OF:
				// TODO
			case syntax.CLOSED:
				// TODO
			}
		}
	}

	// define the symbol and add the definition node
	if !w.defineGlobal(sym) {
		return false
	}

	w.SrcFile.AddDefNode(&sem.HIRTypeDef{
		DefBase: sem.NewDefBase(sym, annots),
		// FieldInitializers: fieldInits,
	})
	return true
}

// -----------------------------------------------------------------------------

// walkOperDef walks an operator definition
func (w *Walker) walkOperDef(branch *syntax.ASTBranch, symbolModifiers int, annots map[string]*sem.Annotation) bool {
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
		// handle intrinsics
		if intrinsicName == "" && body == nil {
			w.logError(
				"operator definition requires a body",
				logging.LMKDef,
				branch.LastBranch().Last().Position(), // error on the parenthesis
			)

			return false
		}

		ft.IntrinsicName = intrinsicName

		// get the `operator` branch for use repeatedly later
		operBranch := branch.BranchAt(1)

		// maps to the first token of the `oper_value` branch unless the
		// operator is `[:]` in which case it maps to the second token
		operCode := operBranch.LeafAt(branch.BranchAt(1).Len() / 2).Kind

		// operName is the string name of the operator (for error messages)
		operName := ""
		for _, item := range operBranch.Content {
			operName += item.(*syntax.ASTLeaf).Value
		}

		// validate the operator against the expected form for it
		expectedArgsForm, expectedReturnForm, err := getOperatorForm(operCode, len(ft.Args))
		if err != nil {
			w.logError(
				fmt.Sprintf("the `%s` operator takes %s", operName, err.Error()),
				logging.LMKDef,
				operBranch.Position(),
			)

			return false
		}

		overload, argsForm, returnForm := w.getOverloadFromSignature(ft)
		overload.Public = symbolModifiers&sem.ModPublic != 0

		if !reflect.DeepEqual(argsForm, expectedArgsForm) || returnForm != expectedReturnForm {
			b := strings.Builder{}
			b.WriteRune('(')

			for i, n := range expectedArgsForm {
				b.WriteRune('T')
				b.WriteString(strconv.Itoa(n))

				if i != len(expectedArgsForm)-1 {
					b.WriteString(", ")
				}
			}

			b.WriteString(") -> T")
			b.WriteString(strconv.Itoa(expectedReturnForm))

			w.logError(
				fmt.Sprintf("signature for operator `%s` must be of form `%s`", operName, b.String()),
				logging.LMKDef,
				operBranch.Position(),
			)
		}

		// define the operator overload; errors handled inside
		if !w.defineOperator(operCode, operName, overload, len(ft.Args), operBranch.Position()) {
			return false
		}

		// create and add the operator definition
		w.SrcFile.AddDefNode(&sem.HIROperDef{
			DefBase: sem.NewDefBase(&sem.Symbol{
				Name:       operName,
				Type:       ft,
				SrcPackage: w.SrcFile.Parent,
				DefKind:    sem.DefKindFuncDef,
				Immutable:  true,
				Modifiers:  symbolModifiers,
				Position:   operBranch.Position(),
			}, annots),
			OperCode: operCode,
			Body:     (*sem.HIRIncomplete)(body),
		})

	}

	return false
}

// -----------------------------------------------------------------------------

// walkFuncDef walks a function definition
func (w *Walker) walkFuncDef(branch *syntax.ASTBranch, symbolModifiers int, annots map[string]*sem.Annotation) bool {
	if sym, argInits, ok := w.walkFuncHeader(branch, symbolModifiers); ok {
		if w.defineGlobal(sym) {
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
					fmt.Sprintf("function `%s` requires a body", sym.Name),
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
func (w *Walker) walkFuncHeader(branch *syntax.ASTBranch, symbolModifiers int) (*sem.Symbol, map[string]sem.HIRExpr, bool) {
	ft := &typing.FuncType{}
	sym := &sem.Symbol{
		Type:       ft,
		Modifiers:  symbolModifiers,
		SrcPackage: w.SrcFile.Parent,
		Immutable:  true,
	}
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
					return nil, nil, false
				}
			}
		case *syntax.ASTLeaf:
			switch v.Kind {
			case syntax.IDENTIFIER:
				sym.Name = v.Value
				sym.Position = v.Position()
			case syntax.ASYNC:
				ft.Async = true
			}
		}
	}

	return sym, argInits, true
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

	// we are going to store `nil` values in this map to keep track of which
	// argument names have been taken -- even values that don't have an
	// initializer will be stored here.  We will prune them after we have
	// finished taking in all the arguments.
	argInits := make(map[string]sem.HIRExpr)

	for _, item := range branch.Content {
		if itembranch, ok := item.(*syntax.ASTBranch); ok {
			switch itembranch.Name {
			case "arg_decl":
				// use one funcArg to store shared argument data
				arg := typing.FuncArg{}

				// the values in this map indicate whether or not the argument
				// is passed by reference
				names := make(map[string]bool)

				allowRequiredArguments := true
				for _, elem := range itembranch.Content {
					// all elements of `arg_decl` are branches
					elembranch := elem.(*syntax.ASTBranch)

					switch elembranch.Name {
					case "arg_id_list":
						nextByRef := false
						for _, item := range elembranch.Content {
							itemleaf := item.(*syntax.ASTLeaf)
							switch itemleaf.Kind {
							case syntax.IDENTIFIER:
								if _, ok := argInits[itemleaf.Value]; ok {
									w.logError(
										fmt.Sprintf("multiple arguments named `%s`", itemleaf.Value),
										logging.LMKArg,
										item.Position(),
									)

									return nil, nil, false
								} else {
									names[itemleaf.Value] = nextByRef
									nextByRef = false
								}
							case syntax.AMP:
								nextByRef = true
							}
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

						allowRequiredArguments = false

						for name, byRef := range names {
							if byRef {
								w.logError(
									fmt.Sprintf("by-reference argument `%s` cannot have an initializer", name),
									logging.LMKArg,
									elem.Position(),
								)

								return nil, nil, false
							}

							argInits[name] = (*sem.HIRIncomplete)(itembranch)
						}
					}
				}

				// use our shared FuncArg by simply duplicating it for each
				// different name and update the fields that change.
				for name, byRef := range names {
					if !allowRequiredArguments && argInits[name] == nil {
						w.logError(
							"all required arguments must come before optional arguments",
							logging.LMKArg,
							itembranch.Position(),
						)

						return nil, nil, false
					}

					fa := arg
					fa.Name = name
					fa.ByReference = byRef

					args = append(args, &fa)
				}
			case "var_arg_decl":
				if isOperator {
					w.logError(
						"operators cannot take variadic arguments",
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
						Name:     name,
						Type:     dt,
						Variadic: true,
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
