package walk

import (
	"chai/logging"
	"chai/sem"
	"chai/syntax"
	"chai/typing"
	"fmt"
)

// WalkDef walks a core definition node (eg. `type_def` or `variable_decl`)
func (w *Walker) WalkDef(branch *syntax.ASTBranch, public bool, annots map[string]*Annotation) bool {
	switch branch.Name {
	case "func_def":
		return w.walkFuncDef(branch, public, annots)
	}

	// TODO: handle generics

	return false
}

// -----------------------------------------------------------------------------

// walkFuncDef walks a function definition
func (w *Walker) walkFuncDef(branch *syntax.ASTBranch, public bool, annots map[string]*Annotation) bool {
	if sym, namePos, argInits, ok := w.walkFuncHeader(branch, public); ok {
		_ = sym
		_ = namePos
		_ = argInits
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
				if v.Len() == 1 {
					// only arguments => return type of nothing
					ft.ReturnType = typing.PrimType(typing.PrimKindNothing)

					if args, _argInits, ok := w.walkFuncArgsDecl(v.BranchAt(0)); ok {
						ft.Args = args
						argInits = _argInits
					} else {
						return nil, nil, nil, false
					}
				} else {
					if rtType, ok := w.walkTypeLabel(v.BranchAt(1)); ok {
						ft.ReturnType = rtType
					} else {
						return nil, nil, nil, false
					}

					if args, _argInits, ok := w.walkFuncArgsDecl(v.BranchAt(0)); ok {
						ft.Args = args
						argInits = _argInits
					} else {
						return nil, nil, nil, false
					}
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

// walkFuncArgsDecl walks an `args_decl` or a `next_arg_decl` node and returns a
// list of valid arguments if possible
func (w *Walker) walkFuncArgsDecl(branch *syntax.ASTBranch) ([]*typing.FuncArg, map[string]sem.HIRExpr, bool) {
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
							}
						case "type_ext":
							if dt, ok := w.walkTypeExt(elembranch); ok {
								arg.Type = dt
							} else {
								return nil, nil, false
							}
						case "initializer":
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
				if nextArgs, newArgInits, ok := w.walkFuncArgsDecl(itembranch); ok {
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
