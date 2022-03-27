package walk

import (
	"chai/ast"
	"chai/depm"
	"chai/syntax"
	"chai/typing"
	"fmt"
	"log"
	"strings"
)

// walkExpr walks an AST expression.  It updates the AST with types (mostly type
// variables to be determined by the solver).
func (w *Walker) walkExpr(expr ast.Expr, yieldsValue bool) bool {
	// just switch over the different kinds of expressions
	switch v := expr.(type) {
	case *ast.Block:
		w.pushScope()
		defer w.popScope()
		return w.walkBlock(v, yieldsValue)
	case *ast.IfExpr:
		return w.walkIfExpr(v, yieldsValue)
	case *ast.WhileExpr:
		return w.walkWhileExpr(v, yieldsValue)
	case *ast.Cast:
		if !w.walkExpr(v.Src, true) {
			return false
		}

		w.solver.AssertCast(v.Src.Type(), v.Type(), v.Position())
		return true
	case *ast.BinaryOp:
		return w.walkBinaryOp(v)
	case *ast.MultiComparison:
		// TODO
		log.Fatalln("multicomparison not implemented yet")
	case *ast.UnaryOp:
		return w.walkUnaryOp(v)
	case *ast.Indirect:
		return w.walkIndirect(v)
	case *ast.Call:
		return w.walkCall(v)
	case *ast.StructInit:
		return w.walkStructInit(v)
	case *ast.Dot:
		return w.walkDot(v)
	case *ast.Literal:
		w.walkLiteral(v)

		// always succeed
		return true
	case *ast.Identifier:
		// just lookup the identifier for now
		if sym, ok := w.lookup(v.Name, v.Pos); ok {
			// check that the definition kinds match up
			// TODO: handle constants and constraints
			if sym.DefKind == depm.DKTypeDef {
				w.reportError(v.Pos, "cannot use %s as value", depm.ReprDefKind(sym.DefKind))
				return false
			}

			// update the type with the type of the matching symbol :)
			v.ExprBase.SetType(sym.Type)
			v.Mutability = &sym.Mutability
			return true
		} else {
			return false
		}
	case *ast.Tuple:
		// just walk all the sub expressions and collect types
		tupleTypes := make([]typing.DataType, len(v.Exprs))
		for i, expr := range v.Exprs {
			// tuples will always require their elements to yield a value;
			// however, parenthesized sub-expressions only require it if the
			// enclosing expression yields a value.
			if !w.walkExpr(expr, yieldsValue || len(v.Exprs) > 1) {
				return false
			}

			tupleTypes[i] = expr.Type()
		}

		// set the return type of the tuple
		if len(v.Exprs) == 1 {
			v.SetType(tupleTypes[0])
		} else {
			v.SetType(typing.TupleType(tupleTypes))
		}

		return true
	}

	// unreachable
	return false
}

// walkBinaryOp walks a binary operator application
func (w *Walker) walkBinaryOp(bop *ast.BinaryOp) bool {
	// walk the LHS and RHS
	if !w.walkExpr(bop.Lhs, true) || !w.walkExpr(bop.Rhs, true) {
		return false
	}

	// create the operator overloaded function
	ftTypeVar, ok := w.makeOverloadFunc(bop.Op, 2)
	if !ok {
		return false
	}

	// return type variable
	rtv := w.solver.NewTypeVar(bop.Position(), "{_}")

	// create operator template to constrain to overload operator type
	operTemplate := &typing.FuncType{
		Args:       []typing.DataType{bop.Lhs.Type(), bop.Rhs.Type()},
		ReturnType: rtv,
	}

	// set the return type of the operator equal to type of the expression
	bop.SetType(rtv)

	// apply the equality constraint between operator and the template
	w.solver.MustBeEquiv(ftTypeVar, operTemplate, bop.Position())

	// set the operator signature equal to the ftTypeVariable
	bop.Op.Signature = ftTypeVar

	return true
}

// walkUnaryOp walks a unary operator application (not special unary operators)
func (w *Walker) walkUnaryOp(uop *ast.UnaryOp) bool {
	// walk the operand
	if !w.walkExpr(uop.Operand, true) {
		return false
	}

	// create the operator overloaded function
	ftTypeVar, ok := w.makeOverloadFunc(uop.Op, 1)
	if !ok {
		return false
	}

	// return type variable
	rtv := w.solver.NewTypeVar(uop.Position(), "{_}")

	// create operator template to constrain to overload operator type
	operTemplate := &typing.FuncType{
		Args:       []typing.DataType{uop.Operand.Type()},
		ReturnType: rtv,
	}

	// set the return type of the operator equal to type of the expression
	uop.SetType(rtv)

	// apply the equality constraint between operator and the template
	w.solver.MustBeEquiv(ftTypeVar, operTemplate, uop.Position())

	// set the operator signature equal to the ftTypeVariable
	uop.Op.Signature = ftTypeVar

	return true
}

// makeOverloadFunc creates a new function type matching an operator overload.
func (w *Walker) makeOverloadFunc(aop ast.Oper, arity int) (typing.DataType, bool) {
	operators, ok := w.lookupOperators(aop)
	if !ok {
		return nil, false
	}

	var opOverloads []typing.OperatorOverload
	for _, op := range operators {
		for _, overload := range op.Overloads {
			if len(overload.Signature.Args) == arity {
				opOverloads = append(opOverloads, typing.OperatorOverload{
					PkgID:     op.Pkg.ID,
					Signature: overload.Signature,
				})
			}
		}
	}

	return w.solver.NewOperatorTypeVar(aop.Pos, aop.Name, &aop.PkgID, opOverloads...), true
}

// -----------------------------------------------------------------------------

// walkIndirect walks an indirection (referencing).
func (w *Walker) walkIndirect(ind *ast.Indirect) bool {
	// walk the operand
	if !w.walkExpr(ind.Operand, true) {
		return false
	}

	// if the operand is an L-value, we need to mark it as mutable because even
	// if it isn't actually mutated through the reference, an allocation is
	// still necessary for the reference to be usable on the backend (need a
	// pointer, not just a value).
	if ind.Operand.Category() == ast.LValue {
		// TODO: handle L-value references to immutable values (eg. constants)
		if !w.assertMutable(ind.Operand) {
			return false
		}
	}

	// TODO: indirection kinds, non-reference assertion

	// set the resulting type of the indirection
	ind.SetType(&typing.RefType{ElemType: ind.Operand.Type()})
	return true
}

// -----------------------------------------------------------------------------

// walkCall walks a function call.
func (w *Walker) walkCall(call *ast.Call) bool {
	// walk the argument and function expressions
	if !w.walkExpr(call.Func, true) {
		return false
	}

	for _, arg := range call.Args {
		if !w.walkExpr(arg, true) {
			return false
		}
	}

	// create a template function type to match against the actual function
	// using our known arguments: this is how we will constrain the shapes of
	// the two functions.
	argVars := make([]typing.DataType, len(call.Args))
	for i, arg := range call.Args {
		argVar := w.solver.NewTypeVar(arg.Position(), arg.Type().Repr())

		argVars[i] = argVar
	}

	rtTypeVar := w.solver.NewTypeVar(call.Position(), "{_}")
	funcTemplate := &typing.FuncType{
		Args:       argVars,
		ReturnType: rtTypeVar,
	}

	// apply the function constraint
	w.solver.MustBeEquiv(call.Func.Type(), funcTemplate, call.Position())

	// constrain the argument type variables to match the actual argument types
	// so that we can type check arguments.  Doing this checking in this sort of
	// "roundabout" way allows us to provide more descriptive error messages
	// pointing to specific arguments as opposed to the whole call.  Note that
	// the constaints must be applied AFTER the primary function constraint so
	// that the argument type variables get given values based on those in
	// actual function first so that failures happen on the individual arguments
	// (we check the actual argument type v. the expected so we have determine
	// the expected first).
	for i, argVar := range argVars {
		w.solver.MustBeEquiv(argVar, call.Args[i].Type(), call.Args[i].Position())
	}

	// set the yield type of the function
	call.SetType(rtTypeVar)

	return true
}

// walkStructInit walks a struct initialization.
func (w *Walker) walkStructInit(init *ast.StructInit) bool {
	// extract the type symbol
	var sym *depm.Symbol
	switch v := init.TypeExpr.(type) {
	case *ast.Identifier:
		// look up the type name
		if _sym, ok := w.lookupGlobal(v.Name, v.Pos); ok {
			sym = _sym
		} else {
			w.reportError(v.Pos, "no struct type named `%s`", v.Name)
			return false
		}
	case *ast.Dot:
		// must be a package definition access
		pkgIdent, ok := v.Root.(*ast.Identifier)
		if !ok {
			w.reportError(init.TypeExpr.Position(), "expected a type")
			return false
		}

		if pkg, ok := w.chFile.VisiblePackages[pkgIdent.Name]; ok {
			if _sym, ok := pkg.SymbolTable[v.FieldName]; ok && _sym.Public {
				sym = _sym
			} else {
				w.reportError(v.FieldPos, "package `%s` has no publicly visible symbol named `%s`", pkg.Name, v.FieldName)
				return false
			}
		} else {
			w.reportError(init.TypeExpr.Position(), "expected a type")
			return false
		}
	}

	// assert that the symbol is a type
	if sym.DefKind != depm.DKTypeDef {
		w.reportError(init.TypeExpr.Position(), "expected a type not a %s", depm.ReprDefKind(sym.DefKind))
		return false
	}

	// assert that the type of the symbol is a struct type
	st, ok := typing.InnerType(sym.Type).(*typing.StructType)
	if !ok {
		w.reportError(init.TypeExpr.Position(), "struct initializer can only be applied to a struct type")
		return false
	}

	// constrain spread initializers
	if init.SpreadInit != nil {
		w.solver.MustBeEquiv(sym.Type, init.SpreadInit.Type(), init.SpreadInit.Position())
	}

	// walk and check the field initializers
	for name, fieldInit := range init.FieldInits {
		// check that the field exists and is visible
		fieldNdx, ok := st.FieldsByName[name]
		if !ok || w.chFile.Parent.ID != sym.File.Parent.ID && !st.Fields[fieldNdx].Public {
			w.reportError(fieldInit.NamePos, "struct `%s` has no public field named `%s`", st.Repr(), name)
			return false
		}

		// assert that the field type is equivalent to the type of the expression
		w.solver.MustBeEquiv(st.Fields[fieldNdx].Type, fieldInit.Init.Type(), fieldInit.Init.Position())
	}

	// set the return type of the expression to be the type of the struct
	init.SetType(st)
	return true
}

// walkDot walks a dot operator.
func (w *Walker) walkDot(dot *ast.Dot) bool {
	// if there is only an identifier in the root, then we may have a package
	// symbol access or an explicit method call which are handled here.
	if ident, ok := dot.Root.(*ast.Identifier); ok {
		// check first for package symbol accesses
		if pkg, ok := w.chFile.VisiblePackages[ident.Name]; ok {
			if sym, ok := pkg.SymbolTable[dot.FieldName]; ok && sym.Public {
				if sym.DefKind != depm.DKValueDef {
					w.reportError(ident.Pos, "cannot use %s as value", depm.ReprDefKind(sym.DefKind))
					return false
				}

				// TODO: figure out how to communicate the package ID to the
				// backend.

				// update the type of the yielded expression
				dot.SetType(sym.Type)
				dot.DotKind = typing.DKStaticGet
				return true
			}

			w.reportError(dot.FieldPos, "package `%s` has no publicly visible symbol named `%s`", pkg.Name, dot.FieldName)
			return false
		}

		// TODO: explicit method calls
	}

	// create a new type variable to hold the returned type of the field
	fieldType := w.solver.NewTypeVar(dot.FieldPos, fmt.Sprintf("{.%s}", dot.FieldName))
	dot.SetType(fieldType)

	// otherwise, it can only be a field access or an implicit method call
	w.solver.MustHaveField(&typing.FieldConstraint{
		RootType:   dot.Root.Type(),
		RootPos:    dot.Root.Position(),
		FieldName:  dot.FieldName,
		FieldPos:   dot.FieldPos,
		FieldType:  fieldType,
		DotKindPtr: &dot.DotKind,
	})

	return true
}

// -----------------------------------------------------------------------------

// walkLiteral walks a literal value.
func (w *Walker) walkLiteral(lit *ast.Literal) {
	switch lit.Kind {
	case syntax.NULL:
		// TODO: null checking (but not until there is an official way to get a
		// `nullptr` for system APIs)
		t := w.solver.NewTypeVar(lit.Pos, "{null}")
		lit.SetType(t)
	case syntax.NUMLIT:
		t := w.solver.NewLiteralTypeVar(
			lit.Pos,
			fmt.Sprintf("{%s}", lit.Value),
			"{number}",

			// the order determines the defaulting preference: eg. this number
			// will default first to an `i64` and if that is not possible, then
			// to an `u64`, etc.
			typing.LiteralOverload{Name: "{int}", Type: typing.PrimType(typing.PrimI64)},
			typing.LiteralOverload{Name: "{int}", Type: typing.PrimType(typing.PrimU64)},
			typing.LiteralOverload{Name: "{int}", Type: typing.PrimType(typing.PrimI32)},
			typing.LiteralOverload{Name: "{int}", Type: typing.PrimType(typing.PrimU32)},
			typing.LiteralOverload{Name: "{int}", Type: typing.PrimType(typing.PrimI16)},
			typing.LiteralOverload{Name: "{int}", Type: typing.PrimType(typing.PrimU16)},
			typing.LiteralOverload{Name: "{int}", Type: typing.PrimType(typing.PrimI8)},
			typing.LiteralOverload{Name: "{int}", Type: typing.PrimType(typing.PrimU8)},

			// floats can be at the bottom because the compiler should only
			// select floats if they are the only option (generally, the user is
			// going want to their number to be an integer)
			typing.LiteralOverload{Name: "{float}", Type: typing.PrimType(typing.PrimF64)},
			typing.LiteralOverload{Name: "{float}", Type: typing.PrimType(typing.PrimF32)},
		)
		lit.SetType(t)
	case syntax.FLOATLIT:
		t := w.solver.NewLiteralTypeVar(
			lit.Pos,
			fmt.Sprintf("{%s}", lit.Value),
			"{float}",
			// default first to highest precision
			typing.LiteralOverload{Name: "{float}", Type: typing.PrimType(typing.PrimF64)},
			typing.LiteralOverload{Name: "{float}", Type: typing.PrimType(typing.PrimF32)},
		)
		lit.SetType(t)
	case syntax.INTLIT:
		isUnsigned := strings.Contains(lit.Value, "u")
		isLong := strings.Contains(lit.Value, "l")

		if isLong && isUnsigned {
			lit.SetType(typing.PrimType(typing.PrimU64))
		} else if isLong {
			t := w.solver.NewLiteralTypeVar(
				lit.Pos,
				fmt.Sprintf("{%s}", lit.Value),
				"{long int}",
				// default first to signed
				typing.LiteralOverload{Name: "{int}", Type: typing.PrimType(typing.PrimI64)},
				typing.LiteralOverload{Name: "{int}", Type: typing.PrimType(typing.PrimU64)},
			)

			lit.SetType(t)
		} else if isUnsigned {
			t := w.solver.NewLiteralTypeVar(
				lit.Pos,
				fmt.Sprintf("{%s}", lit.Value),
				"{unsigned int}",
				// default first to largest size
				typing.LiteralOverload{Name: "{unsigned int}", Type: typing.PrimType(typing.PrimU64)},
				typing.LiteralOverload{Name: "{unsigned int}", Type: typing.PrimType(typing.PrimU32)},
				typing.LiteralOverload{Name: "{unsigned int}", Type: typing.PrimType(typing.PrimU16)},
				typing.LiteralOverload{Name: "{unsigned int}", Type: typing.PrimType(typing.PrimU8)},
			)

			lit.SetType(t)
		} else {
			t := w.solver.NewLiteralTypeVar(
				lit.Pos,
				fmt.Sprintf("{%s}", lit.Value),
				"{int}",
				// default first to signed in order by size
				typing.LiteralOverload{Name: "{signed int}", Type: typing.PrimType(typing.PrimI64)},
				typing.LiteralOverload{Name: "{unsigned int}", Type: typing.PrimType(typing.PrimU64)},
				typing.LiteralOverload{Name: "{signed int}", Type: typing.PrimType(typing.PrimI32)},
				typing.LiteralOverload{Name: "{unsigned int}", Type: typing.PrimType(typing.PrimU32)},
				typing.LiteralOverload{Name: "{signed int}", Type: typing.PrimType(typing.PrimI16)},
				typing.LiteralOverload{Name: "{unsigned int}", Type: typing.PrimType(typing.PrimU16)},
				typing.LiteralOverload{Name: "{signed int}", Type: typing.PrimType(typing.PrimI8)},
				typing.LiteralOverload{Name: "{unsigned int}", Type: typing.PrimType(typing.PrimU8)},
			)

			lit.SetType(t)
		}
	// string literals, rune literals, bool literals, and nothings only have one
	// possible type each.
	case syntax.STRINGLIT:
		lit.SetType(w.uni.PreludeSymbolImports["string"].Type)
	case syntax.RUNELIT:
		lit.SetType(typing.PrimType(typing.PrimI32))
	case syntax.BOOLLIT:
		lit.SetType(typing.BoolType())
	case syntax.NOTHING:
		lit.SetType(typing.NothingType())
	}
}
