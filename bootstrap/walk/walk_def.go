package walk

import (
	"chai/ast"
	"chai/depm"
	"chai/typing"
)

// WalkDef walks a single definition of a source file.  This function allows
// the caller to control when a given definition is evaluated as well as
// what part of it to support generic evaluation.
func (w *Walker) WalkDef(def ast.Def) bool {
	// set the map of dependencies to match the definition being walked
	w.deps = def.Dependencies()

	switch v := def.(type) {
	case *ast.FuncDef:
		return w.walkFuncLike(v.Signature, v.Args, v.Body)
	case *ast.OperDef:
		return w.walkFuncLike(v.Op.Signature.(*typing.FuncType), v.Args, v.Body)
	case *ast.VarDecl:
		return w.WalkGlobalVarDecl(v)
	}

	// clear the map of dependencies
	w.deps = nil

	// TODO: other definitions
	return false
}

// -----------------------------------------------------------------------------

// walkFuncLike walks a function like (ie. a function or an operator: semantics
// are the same for both from an analysis perspective at this point).
func (w *Walker) walkFuncLike(signature *typing.FuncType, args []*ast.FuncArg, body ast.Expr) bool {
	// nil body => nothing to walk => all good
	if body == nil {
		return true
	}

	// push a scope for the function
	w.pushFuncScope(signature, args)

	// make sure the scope is popped before we exit
	defer w.popScope()

	// walk the function body expression (functions that return nothing do not
	// yield a value)
	if !w.walkExpr(body, !typing.IsNothing(signature.ReturnType)) {
		return false
	}

	// add a constraint to the body's return value to ensure that is matches
	// the function's return value if the function actually returns a value.
	if !typing.IsNothing(signature.ReturnType) {
		w.solver.MustBeEquiv(signature.ReturnType, body.Type(), body.Position())
	}

	// type solve the function body
	if !w.solver.Solve() {
		return false
	}

	// update the constancy of the function arguments
	for i, argSym := range w.topScope().LocalArgs {
		// cannot become `Immutable`, but if they are never mutated, then they
		// become constant.
		args[i].Constant = argSym.Mutability == depm.NeverMutated
	}

	// body has been checked -- good to go
	return true
}

// walkGlobalVarDecl walks a global variable declaration.
func (w *Walker) WalkGlobalVarDecl(vd *ast.VarDecl) bool {
	// TODO: make sure global variables properly update any type definition
	// dependencies

	// we don't care about the dependencies of the initializers since they
	// are always run after the global variables themselves.
	w.deps = make(map[string]struct{})

	for _, varList := range vd.VarLists {
		// handle initializers
		if varList.Initializer != nil {
			if !w.walkExpr(varList.Initializer, true) {
				return false
			}

			// if there are multiple names in the variable list, then we have
			// tuple unpacking (need to use different checking semantics)
			if len(varList.Names) > 1 {
				// create a tuple type to match against the one returned from
				// the initializer to appropriately extract types
				varTupleTemplate := make(typing.TupleType, len(varList.Names))

				// the type of the variable type label gets filled in for all
				// the fields of our tuple template (we know the variable has
				// type label)
				for i := range varTupleTemplate {
					varTupleTemplate[i] = varList.Type
				}

				// constrain the tuple returned to match the tuple template
				w.solver.MustBeEquiv(varTupleTemplate, varList.Initializer.Type(), varList.Initializer.Position())

				// return early so we don't declare variables multiple times
				return true
			} else {
				// we know the variable has an type label
				w.solver.MustBeEquiv(varList.Type, varList.Initializer.Type(), varList.Initializer.Position())
			}
		}
	}

	// solve variable context since this the end of the definition.  Note that
	// we technically only need to do this if there is an initializer, but it is
	// cheaper to just call an empty solver once than to call the solver
	// multiple times for multiple initializers.
	return w.solver.Solve()
}
