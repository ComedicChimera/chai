package walk

import (
	"chai/ast"
	"chai/depm"
	"chai/syntax"
	"chai/typing"
	"strings"
)

// walkExpr walks an AST expression.  It updates the AST with types (mostly type
// variables to be determined by the solver).
func (w *Walker) walkExpr(expr ast.Expr) bool {
	// just switch over the different kinds of expressions
	switch v := expr.(type) {
	case *ast.Block:
		return w.walkBlock(v)
	case *ast.Cast:
		if !w.walkExpr(v.Src) {
			return false
		}

		w.solver.AssertCast(v.Src.Type(), v.Type(), v.Position())
		return true
	case *ast.BinaryOp:
		return w.walkBinaryOp(v)
	case *ast.MultiComparison:
		// TODO
	case *ast.Call:
		return w.walkCall(v)
	case *ast.Literal:
		w.walkLiteral(v)

		// always succeed
		return true
	case *ast.Identifier:
		// just lookup the identifier for now
		if sym, ok := w.lookup(v.Name, v.Pos); ok {
			// check that the definition kinds match up
			// TODO: handle constants
			if sym.DefKind != depm.DKValueDef {
				w.reportError(v.Pos, "cannot use %s as value", depm.ReprDefKind(sym.DefKind))
			}

			// update the type with the type of the matching symbol :)
			v.ExprBase.SetType(sym.Type)
			return true
		} else {
			return false
		}
	case *ast.Tuple:
		// just walk all the sub expressions and collect types
		tupleTypes := make([]typing.DataType, len(v.Exprs))
		for i, expr := range v.Exprs {
			if !w.walkExpr(expr) {
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

// walkBinaryOp walks a binary operator applications
func (w *Walker) walkBinaryOp(bop *ast.BinaryOp) bool {
	// walk the LHS and RHS
	if !w.walkExpr(bop.Lhs) || !w.walkExpr(bop.Rhs) {
		return false
	}

	op, ok := w.lookupOperator(&bop.Op)
	if !ok {
		return false
	}

	// create the operator overloaded function
	var ftOverloads []typing.DataType
	for _, overload := range op.Overloads {
		if len(overload.Signature.Args) == 2 {
			ftOverloads = append(ftOverloads, overload.Signature)
		}
	}

	ftTypeVar := w.solver.NewTypeVarWithOverloads(bop.Op.Pos, bop.Op.Name, false, ftOverloads...)

	// return type variable
	rtv := w.solver.NewTypeVar(bop.Position(), "{_}")

	// create operator template to constrain to overload operator type
	operTemplate := &typing.FuncType{
		Args: []typing.FuncArg{
			{
				Type: bop.Lhs.Type(),
			},
			{
				Type: bop.Rhs.Type(),
			},
		},
		ReturnType: rtv,
	}

	// set the return type of the operator equal to type of the expression
	bop.SetType(rtv)

	// apply the equality constraint between operator and the template
	w.solver.Constrain(ftTypeVar, operTemplate, bop.Position())
	return true
}

// -----------------------------------------------------------------------------

// walkCall walks a function call.
func (w *Walker) walkCall(call *ast.Call) bool {
	// walk the argument and function expressions
	if !w.walkExpr(call.Func) {
		return false
	}

	for _, arg := range call.Args {
		if !w.walkExpr(arg) {
			return false
		}
	}

	// create a template function type to match against the actual function
	// using our known arguments: this is how we will constrain the shapes of
	// the two functions.
	templateArgs := make([]typing.FuncArg, len(call.Args))
	argVars := make([]typing.DataType, len(call.Args))
	for i, arg := range call.Args {
		argVar := w.solver.NewTypeVar(arg.Position(), arg.Type().Repr())

		argVars[i] = argVar

		templateArgs[i] = typing.FuncArg{
			Type: argVar,
		}
	}

	rtTypeVar := w.solver.NewTypeVar(call.Position(), "{_}")
	funcTemplate := &typing.FuncType{
		Args:       templateArgs,
		ReturnType: rtTypeVar,
	}

	// apply the function constraint
	w.solver.Constrain(call.Func.Type(), funcTemplate, call.Position())

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
		w.solver.Constrain(argVar, call.Args[i].Type(), call.Args[i].Position())
	}

	// set the yield type of the function
	call.SetType(rtTypeVar)

	return true
}

// -----------------------------------------------------------------------------

// walkLiteral walks a literal value.
func (w *Walker) walkLiteral(lit *ast.Literal) {
	switch lit.Kind {
	case syntax.NULL:
		// TODO: null checking (but not until there is an official way to get a
		// `nullptr` for system APIs)
		t := w.solver.NewTypeVar(lit.Pos, "{_}")
		lit.SetType(t)
	case syntax.NUMLIT:
		t := w.solver.NewTypeVarWithOverloads(
			lit.Pos,
			"{number}",
			true,
			// the order determines the defaulting preference: eg. this number
			// will default first to an `i64` and if that is not possible, then
			// to an `u64`, etc.
			typing.PrimType(typing.PrimI64),
			typing.PrimType(typing.PrimU64),
			typing.PrimType(typing.PrimI32),
			typing.PrimType(typing.PrimU32),
			typing.PrimType(typing.PrimI16),
			typing.PrimType(typing.PrimU16),
			typing.PrimType(typing.PrimI8),
			typing.PrimType(typing.PrimU8),

			// floats can be at the bottom because the compiler should only
			// select floats if they are the only option (generally, the user is
			// going want to their number to be an integer)
			typing.PrimType(typing.PrimF64),
			typing.PrimType(typing.PrimF32),
		)
		lit.SetType(t)
	case syntax.FLOATLIT:
		t := w.solver.NewTypeVarWithOverloads(
			lit.Pos,
			"{float}",
			true,
			// default first to highest precision
			typing.PrimType(typing.PrimF64),
			typing.PrimType(typing.PrimF32),
		)
		lit.SetType(t)
	case syntax.INTLIT:
		isUnsigned := strings.Contains(lit.Value, "u")
		isLong := strings.Contains(lit.Value, "l")

		if isLong && isUnsigned {
			lit.SetType(typing.PrimType(typing.PrimU64))
		} else if isLong {
			t := w.solver.NewTypeVarWithOverloads(
				lit.Pos,
				"{long int}",
				true,
				// default first to signed
				typing.PrimType(typing.PrimI64),
				typing.PrimType(typing.PrimU64),
			)

			lit.SetType(t)
		} else if isUnsigned {
			t := w.solver.NewTypeVarWithOverloads(
				lit.Pos,
				"{unsigned int}",
				true,
				// default first to largest size
				typing.PrimType(typing.PrimU64),
				typing.PrimType(typing.PrimU32),
				typing.PrimType(typing.PrimU16),
				typing.PrimType(typing.PrimU8),
			)

			lit.SetType(t)
		} else {
			t := w.solver.NewTypeVarWithOverloads(
				lit.Pos,
				"{int}",
				true,
				// default first to signed in order by size
				typing.PrimType(typing.PrimI64),
				typing.PrimType(typing.PrimU64),
				typing.PrimType(typing.PrimI32),
				typing.PrimType(typing.PrimU32),
				typing.PrimType(typing.PrimI16),
				typing.PrimType(typing.PrimU16),
				typing.PrimType(typing.PrimI8),
				typing.PrimType(typing.PrimU8),
			)

			lit.SetType(t)
		}
	// string literals, rune literals, bool literals, and nothings only have one
	// possible type each.
	case syntax.STRINGLIT:
		lit.SetType(typing.PrimType(typing.PrimString))
	case syntax.RUNELIT:
		lit.SetType(typing.PrimType(typing.PrimU32))
	case syntax.BOOLLIT:
		lit.SetType(typing.PrimType(typing.PrimBool))
	case syntax.NOTHING:
		lit.SetType(typing.PrimType(typing.PrimNothing))
	}
}
