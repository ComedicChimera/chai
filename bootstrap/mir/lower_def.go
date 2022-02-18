package mir

import (
	"chai/ast"
	"chai/typing"
)

// visitDef visits a definition and recursively evaluates its dependencies
// before determining whether or not to lower it.  This ensures that the
// definitions are placed in the right order. The predicates of definitions are
// also lowered.
func (l *Lowerer) visitDef(def ast.Def) {
	// check that the definition has not already been visited
	if inProgress, ok := l.alreadyVisited[def]; ok {
		// if it is has not finished lowering, then the definition recursively
		// depends on itself and needs to be forward declared.
		if inProgress {
			l.addForwardDecl(def)
		}

		// in both cases, we do not continue with lowering of this definition
		// since doing so would constitute a repeat definition.
		return
	}

	// mark the current definition as in progress
	l.alreadyVisited[def] = true

	// recursively visit its dependencies to ensure they are all fully declared
	// before it (to prevent out of order declarations)
	for dep := range def.Dependencies() {
		l.visitDef(l.defDepGraph[dep])
	}

	// lower the definition itself now that its dependencies have resolved
	l.lowerDef(def)

	// mark it as having been lowered
	l.alreadyVisited[def] = false
}

// addForwardDecl adds a forward declaration for a given definition.  This is
// used when definitions are cyclic.
func (l *Lowerer) addForwardDecl(def ast.Def) {

}

// -----------------------------------------------------------------------------

// lowerDef lowers a definition and adds it to the bundle.
func (l *Lowerer) lowerDef(def ast.Def) {
	switch v := def.(type) {
	case *ast.FuncDef:
		l.lowerFuncDef(v.Name, v.Annotations(), v.Public(), v.Args, v.Signature.ReturnType, v.Body)
	}
}

// lowerFuncDef lowers a function definition.
func (l *Lowerer) lowerFuncDef(name string, annots map[string]string, public bool, args []*ast.FuncArg, rtType typing.DataType, body ast.Expr) {
	// return early if the function is intrinsic
	if _, ok := annots["instrinic"]; ok {
		return
	} else if _, ok := annots["intrinsicop"]; ok {
		return
	}

	// generate the function parameters
	var fparams []*FuncParam

	for _, arg := range args {
		if !typing.IsNothing(arg.Type) {
			fparams = append(fparams, &FuncParam{
				Name:     arg.Name,
				Type:     typing.InnerType(arg.Type),
				Constant: arg.Constant,
			})
		}
	}

	// generate the enclosing function
	fnDef := &FuncDef{
		Name:        name,
		Annotations: annots,
		Public:      public,
		Params:      fparams,
		ReturnType:  typing.InnerType(rtType),
	}

	// generate the function body
	var mirBody *FuncBody
	if body != nil {
		mirBody = &FuncBody{Locals: make(map[string]typing.DataType)}
		fnDef.Body = mirBody
		l.enclosingFunc = fnDef

		expr := l.lowerExpr(body)

		if expr != nil {
			mirBody.Body = append(mirBody.Body, &SimpleStmt{
				StmtKind: SStmtReturn,
				Exprs:    []Expr{expr},
			})
		} else {
			// if body evaluates to nil, then
			mirBody.Body = append(mirBody.Body, &SimpleStmt{StmtKind: SStmtReturn})
		}
	}

	l.b.Funcs = append(l.b.Funcs, fnDef)
}
