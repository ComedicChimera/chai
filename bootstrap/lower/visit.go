package lower

import (
	"chai/ast"
	"chai/mir"
	"fmt"
)

// visit lowers a single definition by first recursively visiting all its
// dependencies and then it adds the fully lowered definition to the MIR bundle.
func (l *Lowerer) visit(def ast.Def) {
	// intrinsics are not compiled as definitions: they use a separate
	// instruction when encountered in MIR blocks
	if _, ok := def.Annotations()["intrinsic"]; ok {
		return
	} else if _, ok := def.Annotations()["intrinsicop"]; ok {
		return
	}

	// if the definition has already been visited
	if mdef, ok := l.alreadyVisited[def]; ok {
		if mdef != nil {
			// the definition is in the process of being added meaning it recursively
			// depends on itself.  Therefore, we need to add it as a forward declaration
			// to break the cycle and return
			l.bundle.Forwards = append(l.bundle.Forwards, mdef)
		}

		// if it has already been added, then we just return (nothing more to do)
		return
	}

	// convert the definition to a MIR definition
	mdef := l.lowerDef(def)

	// add it to the map of already visited to prevent infinite recursion and
	// allow for forward definitions as necessary
	l.alreadyVisited[def] = mdef

	// visit its dependencies
	for name := range def.Dependencies() {
		l.visit(l.defDepGraph[name])
	}

	// define global types and functions
	switch v := def.(type) {
	case *ast.FuncDef:
		// handle externals and DLL imports
		if v.Body == nil {
			l.bundle.Externals = append(l.bundle.Externals, mdef)
		} else {
			// lower the function body
			l.bundle.Functions = append(l.bundle.Functions, &mir.FuncImpl{
				Def:  mdef.(*mir.FuncDef),
				Body: l.lowerBody(v.Body),
			})
		}
	case *ast.OperDef:
		// lower the function body
		l.bundle.Functions = append(l.bundle.Functions, &mir.FuncImpl{
			Def:  mdef.(*mir.FuncDef),
			Body: l.lowerBody(v.Body),
		})
	}

	// flag the definition as completed before returning
	l.alreadyVisited[def] = nil
}

// lowerDef lowers a definition into a MIR definition *without* lowering its
// body.  This function assumes the definition is non-intrinsic.
func (l *Lowerer) lowerDef(def ast.Def) mir.Def {
	// TODO: other definitions
	switch v := def.(type) {
	case *ast.FuncDef:
		// check for the inline tag
		_, inline := def.Annotations()["inline"]

		// the entry point does not get mangled (so it can be seen by linker)
		var fullName string
		_, entry := def.Annotations()["entry"]
		if entry {
			fullName = v.Name
		} else {
			fullName = l.globalPrefix + v.Name
		}

		return &mir.FuncDef{
			Name:       fullName,
			Args:       v.Args,
			ReturnType: v.Signature.ReturnType,
			Pub:        v.Public(),
			Inline:     inline,
		}
	case *ast.OperDef:
		// operators get converted into functions that are always inlined
		return &mir.FuncDef{
			Name:       fmt.Sprintf("%s.oper[%s: %s]", l.globalPrefix, v.Op.Name, v.Op.Signature.Repr()),
			Args:       v.Args,
			ReturnType: v.Op.Signature.ReturnType,
			Pub:        v.Public(),
			Inline:     true,
		}
	}

	// unreachable
	return nil
}

// lowerBody lowers the function or a operator.
func (l *Lowerer) lowerBody(body ast.Expr) []mir.Stmt {
	// TODO: push a local scope for the body and arguments

	// reset the temporary counter
	l.tempCounter = 0

	var bodyStmts []mir.Stmt
	rtVal := l.lowerExpr(&bodyStmts, body)

	// TODO: handle returns at the end of blocks

	// return needs to be added
	if rtVal == nil {
		// return value
		bodyStmts = append(bodyStmts, &mir.Return{})
	} else {
		// return value
		bodyStmts = append(bodyStmts, &mir.Return{RetVal: rtVal})
	}

	return bodyStmts
}
