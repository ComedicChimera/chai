package lower

import (
	"chai/ast"
	"chai/depm"
	"chai/mir"
	"chai/typing"
	"log"
)

// lowerExpr converts an AST expression into a MIR expression.
func (l *Lowerer) lowerExpr(expr ast.Expr) mir.Expr {
	switch v := expr.(type) {
	case *ast.BinaryOp:
		return l.lowerOperApp(v.Op, v.Type(), v.Lhs, v.Rhs)
	case *ast.Cast:
		// casts between equivalent types are unnecessary
		if typing.Equiv(v.Src.Type(), v.Type()) {
			return l.lowerExpr(v.Src)
		}

		return &mir.CastExpr{
			Src:      l.lowerExpr(v.Src),
			DestType: typing.Simplify(v.Type()),
		}
	case *ast.Identifier:
		{
			// if the symbol is in the local scope stack, then we return a local
			// identifier; otherwise, we return a global identifier.
			for i := len(l.scopeStack) - 1; i >= 0; i-- {
				if loweredName, ok := l.scopeStack[i][v.Name]; ok {
					return &mir.LocalIdent{
						Name:    loweredName,
						IdType:  typing.Simplify(v.Type()),
						Mutable: *v.Mutability == depm.Mutable,
					}
				}
			}

			// global identifier
			return &mir.GlobalIdent{
				ParentID: l.pkg.ID,
				Name:     v.Name,
				IdType:   v.Type(),
				Mutable:  *v.Mutability == depm.Mutable,
			}
		}
	case *ast.Literal:
		return &mir.Literal{
			Value:   v.Value,
			LitType: typing.Simplify(v.Type()),
		}
	}

	// TODO
	return nil
}

// intrinsicTable is a table mapping intrinsic names to their MIR Op Codes
var intrinsicTable map[string]int = map[string]int{
	"iadd": mir.OCIAdd,
	"isub": mir.OCISub,
	"ineg": mir.OCINeg,
	"slt":  mir.OCSLt,
	"sgt":  mir.OCSGt,
	"sge":  mir.OCSGtEq,
	"eq":   mir.OCEq,
	"sdiv": mir.OCSDiv,
	"smod": mir.OCSMod,
	"lor":  mir.OCOr,
}

// lowerOperApp lowers an operator application.
func (l *Lowerer) lowerOperApp(op ast.Oper, resultType typing.DataType, operands ...ast.Expr) mir.Expr {
	operFt := typing.InnerType(op.Signature).(*typing.FuncType)

	// TODO: handle nothing pruning of operands

	// handle intrinsic operators
	if operFt.IntrinsicName != "" {
		opCode, ok := intrinsicTable[operFt.IntrinsicName]
		if !ok {
			log.Fatalf("unknown intrinsic named: `%s`\n", operFt.IntrinsicName)
			return nil
		}

		return &mir.OperExpr{
			OpCode: opCode,
			// Operands: ,
		}
	}

	// TODO
	return nil
}
