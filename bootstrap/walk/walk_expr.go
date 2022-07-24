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

		if !types.Cast(v.SrcExpr.Type(), v.Type()) {
			w.error(v.Span(), "cannot cast %s to %s", v.SrcExpr.Type().Repr(), v.Type().Repr())
		}
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
		w.walkExpr(v.Ptr)

		if pt, ok := types.InnerType(v.Ptr.Type()).(*types.PointerType); ok {
			v.NodeType = pt.ElemType
		} else {
			w.error(v.Span(), "%s cannot be dereferenced because it is not a pointer", v.Ptr.Type().Repr())
		}
	case *ast.Indirect:
		w.walkExpr(v.Elem)

		v.NodeType = &types.PointerType{
			ElemType: v.Elem.Type(),
			Const:    v.Const || (v.Elem.Category() == ast.RVALUE && v.Elem.Constant()),
		}
	case *ast.Null:
		v.NodeType = &types.UntypedNull{Span: v.Span()}
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

	// Assert that the value being called is a function.
	ft, ok := types.InnerType(call.Func.Type()).(*types.FuncType)
	if !ok {
		w.error(call.Span(), "%s cannot be called because it is not a function", call.Func.Type().Repr())
	}

	// Check the number of arguments.
	if len(ft.ParamTypes) != len(call.Args) {
		w.error(call.Span(), "expected %d parameters but received %d arguments", len(ft.ParamTypes), len(call.Args))
	}

	// Walk the arguments and check the argument types.
	for i, arg := range call.Args {
		w.walkExpr(arg)

		w.mustUnify(ft.ParamTypes[i], arg.Type(), arg.Span())
	}

	// Set the resultant type of the function call.
	call.NodeType = ft.ReturnType
}
