package walk

import (
	"chaic/ast"
	"chaic/common"
	"chaic/types"
)

// walkExpr walks an AST expression.
func (w *Walker) walkExpr(expr ast.ASTExpr) {
	switch v := expr.(type) {
	case *ast.TypeCast:
		w.walkExpr(v.SrcExpr)

		w.solver.MustCast(v.SrcExpr.Type(), v.Type(), v.Span())
	case *ast.BinaryOpApp:
		w.walkExpr(v.LHS)
		w.walkExpr(v.RHS)

		v.NodeType = w.checkOperApp(v.Op, v.Span(), v.LHS, v.RHS)
	case *ast.UnaryOpApp:
		w.walkExpr(v.Operand)

		v.NodeType = w.checkOperApp(v.Op, v.Span(), v.Operand)
	case *ast.FuncCall:
		w.walkFuncCall(v)
	case *ast.Deref:
		{
			w.walkExpr(v.Ptr)

			// TODO: handle non-const/const pointers.
		}
	case *ast.Indirect:
		w.walkExpr(v.Elem)

		v.NodeType = &types.PointerType{
			ElemType: v.Elem.Type(),
			Const:    v.Const || v.Elem.Constant(),
		}
	case *ast.StructLiteral:
		if st, ok := types.InnerType(v.Type()).(*types.StructType); ok {
			_ = st
		} else {
			w.error(v.Span(), "%s is not a struct type", v.Type().Repr())
		}
	case *ast.Null:
		v.NodeType = w.solver.NewTypeVar("untyped null", v.Span())
	case *ast.Literal:
		w.walkLiteral(v)
	case *ast.Identifier:
		{
			sym := w.lookup(v.Name, v.Span())

			if sym.DefKind == common.DefKindType {
				w.error(v.Span(), "%s cannot be used as a value", sym.Name)
			}

			v.Sym = sym
		}
	}
}

// -----------------------------------------------------------------------------

// walkFuncCall walks a function call.
func (w *Walker) walkFuncCall(call *ast.FuncCall) {
	w.walkExpr(call.Func)

	// Check if we know the value being called is a function.
	if ft, ok := types.InnerType(call.Func.Type()).(*types.FuncType); ok {
		// Check the number of arguments.
		if len(ft.ParamTypes) != len(call.Args) {
			w.error(call.Span(), "expected %d parameters but received %d arguments", len(ft.ParamTypes), len(call.Args))
		}

		// Walk the arguments and check the argument types.
		for i, arg := range call.Args {
			w.walkExpr(arg)

			w.solver.MustEqual(ft.ParamTypes[i], arg.Type(), arg.Span())
		}

		// Set the resultant type of the function call.
		call.NodeType = ft.ReturnType
	} else {
		// Otherwise, we are likely either dealing with a non-callable type or a
		// type variable.  In either case, we have to use the type solver.

		// Wrap the arguments into a function type.
		returnType := w.solver.NewTypeVar("unknown return type", call.Span())

		argTypes := make([]types.Type, len(call.Args))
		for i, arg := range call.Args {
			argTypes[i] = arg.Type()
		}

		funcType := &types.FuncType{ParamTypes: argTypes, ReturnType: returnType}

		// Assert it to be equal to the function value's type.
		w.solver.MustEqual(call.Func.Type(), funcType, call.Span())

		// Set the resultant type of the function call.
		call.NodeType = returnType
	}
}
