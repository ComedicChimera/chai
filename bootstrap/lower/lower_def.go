package lower

import (
	"chai/ast"
	"chai/mir"
	"chai/typing"
	"fmt"
	"log"
)

// lowerDef lowers a definition and adds it to the MIR bundle.
func (l *Lowerer) lowerDef(def ast.Def) {
	// Note: we ignore intrinsics since they are only defined in the Universe.
	switch v := def.(type) {
	case *ast.FuncDef:
		{
			var body *mir.FuncBody
			if v.Body != nil {
				l.pushScope()
				defer l.popScope()

				l.locals = make(map[string]typing.DataType)
				body = &mir.FuncBody{
					Locals:   l.locals,
					BodyExpr: l.lowerExpr(v.Body),
				}
			}
			l.b.Funcs = append(l.b.Funcs, &mir.FuncDef{
				ParentID:    l.pkg.ID,
				Name:        v.Name,
				Annotations: v.Annotations(),
				Public:      v.Public(),
				Args:        convFuncArgs(v.Args),
				ReturnType:  typing.Simplify(v.Signature.ReturnType),
				Body:        body,
			})
		}
	case *ast.OperDef:
		// operators just compile to functions
		{
			var body *mir.FuncBody
			if v.Body != nil {
				l.pushScope()
				defer l.popScope()

				l.locals = make(map[string]typing.DataType)
				body = &mir.FuncBody{
					Locals:   l.locals,
					BodyExpr: l.lowerExpr(v.Body),
				}
			}

			l.b.Funcs = append(l.b.Funcs, &mir.FuncDef{
				ParentID:    l.pkg.ID,
				Name:        fmt.Sprintf("oper[%s]", v.Op.Name),
				Annotations: v.Annotations(),
				Public:      v.Public(),
				Args:        convFuncArgs(v.Args),
				ReturnType:  typing.Simplify(v.Op.Signature).(*typing.FuncType).ReturnType,
				Body:        body,
			})
		}
	case *ast.VarDecl:
		// each global variable list compiles as a separate global variable
		// declaration (all have same initializer and so are processed together)
		for _, vlist := range v.VarLists {
			var init *mir.GlobalVarInit
			if vlist.Initializer != nil {
				l.locals = make(map[string]typing.DataType)
				init = &mir.GlobalVarInit{
					Locals:   l.locals,
					InitExpr: l.lowerExpr(vlist.Initializer),
				}
			}

			l.b.GlobalVars = append(l.b.GlobalVars, &mir.GlobalVarDef{
				ParentID:    l.pkg.ID,
				Names:       vlist.Names,
				Public:      v.Public(),
				Type:        vlist.Type,
				Initializer: init,
			})
		}
	case *ast.StructDef:
		l.b.Types = append(l.b.Types, &mir.TypeDef{
			ParentID:    l.pkg.ID,
			Name:        v.Name,
			Public:      v.Public(),
			Annotations: v.Annotations(),
			DefType:     typing.SimplifyStructTypeDef(v.Type),
		})
	}
}

// forwardDecl adds a forward declaration to the MIR bundle.
func (l *Lowerer) forwardDecl(def ast.Def) {
	switch v := def.(type) {
	case *ast.FuncDef:
		l.b.Funcs = append(l.b.Funcs, &mir.FuncDef{
			ParentID:    l.pkg.ID,
			Name:        v.Name,
			Annotations: v.Annotations(),
			Public:      v.Public(),
			Args:        convFuncArgs(v.Args),
			ReturnType:  typing.Simplify(v.Signature.ReturnType),
			// no body
		})
	case *ast.OperDef:
		l.b.Funcs = append(l.b.Funcs, &mir.FuncDef{
			ParentID:    l.pkg.ID,
			Name:        fmt.Sprintf("oper[%s]", v.Op.Name),
			Annotations: v.Annotations(),
			Public:      v.Public(),
			Args:        convFuncArgs(v.Args),
			ReturnType:  typing.Simplify(v.Op.Signature).(*typing.FuncType).ReturnType,
			// no body
		})
	case *ast.StructDef:
		l.b.Types = append(l.b.Types, &mir.TypeDef{
			ParentID:    l.pkg.ID,
			Name:        v.Name,
			Public:      v.Public(),
			Annotations: v.Annotations(),
			// no type -- translate to an opaque type
		})
	default:
		// many constructs cannot be forward declarations and should never be:
		// eg. global variables.  This is either because doing so would be
		// nonsensical or because their representation in LLVM separates their
		// declaration from any dependencies which could be recursive.
		log.Fatalln("Unsupported forward declaration")
	}
}

// -----------------------------------------------------------------------------

// convFuncArgs converts a slice of function arguments to MIR function arguments.
func convFuncArgs(args []*ast.FuncArg) []mir.FuncParam {
	var params []mir.FuncParam
	for _, arg := range args {
		stype := typing.Simplify(arg.Type)
		if !typing.IsNothing(stype) {
			params = append(params, mir.FuncParam{
				Name:     arg.Name,
				Type:     stype,
				Constant: arg.Constant,
			})
		}
	}

	return params
}
