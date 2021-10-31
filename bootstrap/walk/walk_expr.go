package walk

import (
	"chai/ast"
	"chai/syntax"
	"chai/typing"
	"strings"
)

// walkExpr walks an AST expression.  It updates the AST with types (mostly type
// variables to be determined by the solver).
func (w *Walker) walkExpr(expr ast.Expr) bool {
	// just switch over the different kinds of expressions
	switch v := expr.(type) {
	case *ast.Literal:
		w.walkLiteral(v)

		// always succeed
		return true
	case *ast.Identifier:
		// just lookup the identifier for now
		if sym, ok := w.lookup(v.Name, v.Pos); ok {
			// update the type with the type of the matching symbol :)
			v.ExprBase.SetType(sym.Type)
			return true
		} else {
			return false
		}
	case *ast.Tuple:
		// just walk all the sub expressions
		for _, expr := range v.Exprs {
			if !w.walkExpr(expr) {
				return false
			}
		}

		return true
	case *ast.BinaryOp:
		return w.walkBinaryOp(v)
	case *ast.MultiComparison:
		// TODO
	}

	// unreachable
	return false
}

// walkBinaryOp walks a binary operator applications
func (w *Walker) walkBinaryOp(bop *ast.BinaryOp) bool {
	op, ok := w.lookupOperator(&bop.Op)
	if !ok {
		return false
	}

	// TODO: rest of this
	_ = op
	return false
}

// -----------------------------------------------------------------------------

// walkLiteral walks a literal value.
func (w *Walker) walkLiteral(lit *ast.Literal) {
	switch lit.Kind {
	case syntax.NULL:
		t := w.solver.NewTypeVar(lit.Pos)
		lit.SetType(t)
	case syntax.NUMLIT:
		t := w.solver.NewTypeVarWithOverloads(
			lit.Pos,
			true,
			// the order determines the defaulting preference: eg. this number
			// will default first to an `i32` and if that is not possible, then
			// to an `f32`, etc.
			typing.PrimType(typing.PrimI32),
			typing.PrimType(typing.PrimF32),
			typing.PrimType(typing.PrimI64),
			typing.PrimType(typing.PrimF64),
			typing.PrimType(typing.PrimU32),
			typing.PrimType(typing.PrimU64),
			typing.PrimType(typing.PrimI16),
			typing.PrimType(typing.PrimU16),
			typing.PrimType(typing.PrimI8),
			typing.PrimType(typing.PrimU8),
		)
		lit.SetType(t)
	case syntax.FLOATLIT:
		t := w.solver.NewTypeVarWithOverloads(
			lit.Pos,
			true,
			// default first to f32
			typing.PrimType(typing.PrimF32),
			typing.PrimType(typing.PrimF64),
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
				true,
				// default first to i64
				typing.PrimType(typing.PrimI64),
				typing.PrimType(typing.PrimU64),
			)

			lit.SetType(t)
		} else if isUnsigned {
			t := w.solver.NewTypeVarWithOverloads(
				lit.Pos,
				true,
				// default first to u32
				typing.PrimType(typing.PrimU32),
				typing.PrimType(typing.PrimU64),
				typing.PrimType(typing.PrimI16),
				typing.PrimType(typing.PrimU16),
				typing.PrimType(typing.PrimI8),
				typing.PrimType(typing.PrimU8),
			)

			lit.SetType(t)
		} else {
			t := w.solver.NewTypeVarWithOverloads(
				lit.Pos,
				true,
				// default first to i32
				typing.PrimType(typing.PrimI32),
				typing.PrimType(typing.PrimU32),
				typing.PrimType(typing.PrimI64),
				typing.PrimType(typing.PrimU64),
				typing.PrimType(typing.PrimU16),
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
