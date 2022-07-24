package walk

import (
	"chaic/ast"
	"chaic/common"
	"chaic/types"
)

// doWalkDef walks a definition in a Chai file.  This should only be called from
// `walkDef`.
func (w *Walker) doWalkDef(def ast.ASTNode) {
	switch v := def.(type) {
	case *ast.FuncDef:
		w.validateFuncAnnots(v)
		w.walkFuncBody(v.Params, v.Symbol.Type.(*types.FuncType).ReturnType, v.Body)
	case *ast.OperDef:
		w.validateOperAnnots(v)
		w.walkFuncBody(v.Params, v.Overload.Signature.(*types.FuncType).ReturnType, v.Body)
	}
}

// The table of valid intrinsic function names.
var intrinsicFuncs = map[string]struct{}{}

// validateFuncAnnots validates the annotations of a function definition.
func (w *Walker) validateFuncAnnots(fd *ast.FuncDef) {
	expectsBody := true

	for aname, aval := range fd.Annotations {
		switch aname {
		case "intrinsic":
			if _, ok := intrinsicFuncs[fd.Symbol.Name]; !ok {
				w.recError(fd.Symbol.DefSpan, "no intrinsic function named `%s`", fd.Symbol.Name)
			}

			fallthrough
		case "extern":
			expectsBody = false
			fallthrough
		case "abientry":
			if len(aval.Value) != 0 {
				w.recError(aval.ValSpan, "@%s does not take an argument when applied to a function definition", aname)
			}
		}
	}

	if expectsBody && fd.Body == nil {
		w.recError(fd.Span(), "function must have a body")
	} else if !expectsBody && fd.Body != nil {
		w.recError(fd.Span(), "function must not have a body")
	}
}

// The table of intrinsic operator names.
var intrinsicOps = map[string]struct{}{
	"iadd":  {},
	"fadd":  {},
	"isub":  {},
	"fsub":  {},
	"imul":  {},
	"fmul":  {},
	"sdiv":  {},
	"udiv":  {},
	"fdiv":  {},
	"smod":  {},
	"umod":  {},
	"fmod":  {},
	"slt":   {},
	"ult":   {},
	"flt":   {},
	"sgt":   {},
	"ugt":   {},
	"fgt":   {},
	"slteq": {},
	"ulteq": {},
	"flteq": {},
	"sgteq": {},
	"ugteq": {},
	"fgteq": {},
	"ieq":   {},
	"feq":   {},
	"ineq":  {},
	"fneq":  {},
	"land":  {},
	"lor":   {},
	"lnot":  {},
	"ineg":  {},
	"fneg":  {},
	"band":  {},
	"bor":   {},
	"bxor":  {},
	"shl":   {},
	"ashr":  {},
	"lshr":  {},
	"compl": {},
}

// validateOperAnnots validates the annotations of an operator definition.
func (w *Walker) validateOperAnnots(od *ast.OperDef) {
	expectsBody := true

	for aname, aval := range od.Annotations {
		switch aname {
		case "intrinsic":
			if len(aval.Value) == 0 {
				w.recError(aval.NameSpan, "@intrinsic requires an argument when applied to an operator definition")
			}

			if _, ok := intrinsicOps[aval.Value]; !ok {
				w.recError(aval.ValSpan, "no intrinsic operator named `%s`", aval.Value)
			}

			od.Overload.IntrinsicName = aval.Value
			expectsBody = false
		}
	}

	if expectsBody && od.Body == nil {
		w.recError(od.Span(), "operator must have a body")
	} else if !expectsBody && od.Body != nil {
		w.recError(od.Span(), "operator must not have a body")
	}
}

// walkFuncBody walks a function or operator body.
func (w *Walker) walkFuncBody(params []*common.Symbol, rtType types.Type, body ast.ASTNode) {
	// Push the enclosing scope of the function.
	w.pushScope()
	defer w.popScope()

	// Declare all parameter symbols.
	for _, paramSym := range params {
		w.defineLocal(paramSym)
	}

	// Set the function return type.
	w.enclosingReturnType = rtType

	// Actually, walk the block/expression making up the function body.
	if bodyBlock, ok := body.(*ast.Block); ok { // Block
		cm := w.walkBlock(bodyBlock)

		// Make sure the function returns.
		if !types.IsUnit(rtType) && cm != ControlReturn && cm != ControlNoExit {
			if len(bodyBlock.Stmts) > 0 {
				w.error(
					bodyBlock.Stmts[len(bodyBlock.Stmts)-1].Span(),
					"missing return statement",
				)
			} else {
				w.error(
					bodyBlock.Span(),
					"missing return statement",
				)
			}
		}
	} else if bodyExpr, ok := body.(ast.ASTExpr); ok { // Expression
		w.walkExpr(bodyExpr)

		// Check the type of the expression matches the type of the function.
		w.mustUnify(rtType, bodyExpr.Type(), body.Span())
	}

	// Clear the function return type.
	w.enclosingReturnType = nil
}
