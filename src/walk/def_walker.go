package walk

import (
	"chai/logging"
	"chai/sem"
	"chai/syntax"
	"chai/typing"
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
					// TODO: handle walking the return type

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

// walkFuncArgsDecl walks an `args_decl` node and returns a list of valid arguments if possible
func (w *Walker) walkFuncArgsDecl(branch *syntax.ASTBranch) ([]*typing.FuncArg, map[string]sem.HIRExpr, bool) {
	return nil, nil, false
}
