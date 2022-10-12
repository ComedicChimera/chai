package walk

import (
	"chaic/ast"
	"chaic/common"
	"chaic/syntax"
	"chaic/types"
)

// walkExpr walks an AST expression.
func (w *Walker) walkExpr(expr ast.ASTExpr) {
	switch v := expr.(type) {
	case *ast.TypeCast:
		w.walkExpr(v.SrcExpr)

		w.mustCast(v.SrcExpr.Type(), v.Type(), v.Span())
	case *ast.BinaryOpApp:
		w.walkExpr(v.LHS)
		w.walkExpr(v.RHS)

		v.NodeType = w.checkOperApp(v.Op, v.Span(), v.LHS, v.RHS)
	case *ast.UnaryOpApp:
		w.walkExpr(v.Operand)

		v.NodeType = w.checkOperApp(v.Op, v.Span(), v.Operand)
	case *ast.Deref:
		{
			w.walkExpr(v.Ptr)

			if ptr, ok := types.InnerType(v.Ptr.Type()).(*types.PointerType); ok {
				v.NodeType = ptr.ElemType
			} else {
				w.error(v.Ptr.Span(), "%s is not a pointer", v.Ptr.Type().Repr())
			}
		}
	case *ast.Indirect:
		w.walkExpr(v.Elem)

		v.NodeType = &types.PointerType{
			ElemType: v.Elem.Type(),
			Const:    v.Const || v.Elem.Constant(),
		}
	case *ast.PropertyAccess:
		w.walkExpr(v.Root)

		if prop := types.GetProperty(v.Root.Type(), v.PropName); prop != nil {
			v.NodeType = prop.Type
			v.PropKind = prop.Kind
		} else {
			w.error(v.PropSpan, "%s has no property named %s", v.Root.Type().Repr(), v.PropName)
		}
	case *ast.StructLiteral:
		w.walkStructLit(v)
	case *ast.FuncCall:
		w.walkFuncCall(v)
	case *ast.Index:
		w.walkIndex(v, nil)
	case *ast.Slice:
		w.walkSlice(v, nil)
	case *ast.ArrayLiteral:
		w.walkArrayLit(v)
	case *ast.Null:
		v.NodeType = w.newUntypedNull(v.Span())
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

// walkStructLit walks a struct literal.
func (w *Walker) walkStructLit(lit *ast.StructLiteral) {
	// Validate that the type name is a type.
	if st, ok := types.InnerType(lit.Type()).(*types.StructType); ok {
		// Validate the spread initializer if it exists.
		if lit.SpreadInit != nil {
			w.walkExpr(lit.SpreadInit)

			w.mustUnify(st, lit.SpreadInit.Type(), lit.SpreadInit.Span())
		}

		// Validate the field initializers.
		for _, init := range lit.FieldInits {
			w.walkExpr(init.InitExpr)

			// Validate that the field exists.
			if field, ok := st.GetFieldByName(init.Name); ok {
				// Assert that the types match.
				w.mustUnify(field.Type, init.InitExpr.Type(), init.InitExpr.Span())
			} else {
				w.recError(init.NameSpan, "%s has no field named %s", st.Repr(), init.Name)
			}
		}
	} else {
		w.error(lit.Span(), "%s is not a struct type", lit.Type().Repr())
	}
}

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

			w.mustUnify(ft.ParamTypes[i], arg.Type(), arg.Span())
		}

		// Set the resultant type of the function call.
		call.NodeType = ft.ReturnType
	} else {
		w.error(call.Func.Span(), "%s is not a function", call.Func.Type().Repr())
	}
}

// TODO: figure out how to handle LHS index and slice expressions correctly.

// walkIndex walks an LHS or RHS index expression.
func (w *Walker) walkIndex(index *ast.Index, rhsExpr ast.ASTExpr) {
	w.walkExpr(index.Root)
	w.walkExpr(index.Subscript)

	if arrTyp, ok := types.InnerType(index.Root.Type()).(*types.ArrayType); ok {
		w.mustUnify(types.PrimTypeI64, index.Subscript.Type(), index.Subscript.Span())

		index.NodeType = arrTyp.ElemType

		var opMethod *common.OperatorMethod
		if rhsExpr != nil {
			if index.Root.Constant() || arrTyp.Const {
				w.recError(index.Span(), "cannot mutate an immutable value")
			}

			opMethod = &common.OperatorMethod{
				ID: common.OP_ID_ARRAY_SET_INDEX,
				Signature: &types.FuncType{
					ParamTypes: []types.Type{arrTyp, index.Subscript.Type(), rhsExpr.Type()},
					ReturnType: arrTyp.ElemType,
				},
			}
		} else {
			opMethod = &common.OperatorMethod{
				ID: common.OP_ID_ARRAY_GET_INDEX,
				Signature: &types.FuncType{
					ParamTypes: []types.Type{arrTyp, index.Subscript.Type()},
					ReturnType: arrTyp.ElemType,
				},
			}
		}

		index.Op = &ast.AppliedOperator{
			OpKind:   syntax.TOK_LBRACKET,
			OpName:   "[]",
			Span:     index.Span(),
			OpMethod: opMethod,
		}
	} else {
		// TODO: lookup operator overload
		w.error(index.Root.Span(), "%s is not an array", index.Root.Type().Repr())
	}
}

// walkSlice walks an LHS or RHS slice expression.
func (w *Walker) walkSlice(slice *ast.Slice, rhsExpr ast.ASTExpr) {
	w.walkExpr(slice.Root)

	if slice.Start != nil {
		w.walkExpr(slice.Start)
	}

	if slice.End != nil {
		w.walkExpr(slice.End)
	}

	if arrTyp, ok := types.InnerType(slice.Root.Type()).(*types.ArrayType); ok {
		if slice.Start != nil {
			w.mustUnify(types.PrimTypeI64, slice.Start.Type(), slice.Start.Span())
		}

		if slice.End != nil {
			w.mustUnify(types.PrimTypeI64, slice.End.Type(), slice.End.Span())
		}

		slice.NodeType = arrTyp

		var opMethod *common.OperatorMethod
		if rhsExpr != nil {
			if slice.Root.Constant() || arrTyp.Const {
				w.recError(slice.Span(), "cannot mutate an immutable value")
			}

			opMethod = &common.OperatorMethod{
				ID: common.OP_ID_ARRAY_SET_SLICE,
				Signature: &types.FuncType{
					ParamTypes: []types.Type{arrTyp, types.PrimTypeI64, types.PrimTypeI64, rhsExpr.Type()},
					ReturnType: arrTyp,
				},
			}
		} else {
			opMethod = &common.OperatorMethod{
				ID: common.OP_ID_ARRAY_GET_SLICE,
				Signature: &types.FuncType{
					ParamTypes: []types.Type{arrTyp, types.PrimTypeI64, types.PrimTypeI64},
					ReturnType: arrTyp,
				},
			}
		}

		slice.Op = &ast.AppliedOperator{
			OpKind:   syntax.TOK_COLON,
			OpName:   "[:]",
			Span:     slice.Span(),
			OpMethod: opMethod,
		}
	} else {
		// TODO: lookup operator overload
		w.error(slice.Root.Span(), "%s is not an array", slice.Root.Type().Repr())
	}
}

/* -------------------------------------------------------------------------- */

// walkArrayLit walks an array literal.
func (w *Walker) walkArrayLit(arrLit *ast.ArrayLiteral) {
	for i, elem := range arrLit.Elements {
		w.walkExpr(elem)

		if i == 0 {
			arrLit.NodeType = elem.Type()
		} else {
			w.mustUnify(arrLit.NodeType, elem.Type(), elem.Span())
		}
	}
}
