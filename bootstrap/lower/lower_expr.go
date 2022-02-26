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
	case *ast.Dot:
		return l.lowerDotExpr(v)
	case *ast.Call:
		return l.lowerFuncCall(v)
	case *ast.UnaryOp:
		return l.lowerOperApp(v.Op, v.Type(), v.Operand)
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
	case *ast.Indirect:
		{
			opExpr := l.lowerExpr(v.Operand)

			return &mir.OperExpr{
				OpCode:     mir.OCIndirect,
				Operands:   []mir.Expr{opExpr},
				ResultType: &typing.RefType{ElemType: opExpr.Type()},
			}
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

// lowerDotExpr lowers a dot expression.
func (l *Lowerer) lowerDotExpr(dot *ast.Dot) mir.Expr {
	if dot.IsStaticGet {
		log.Fatalln("static get not implemented")
		// return &mir.GlobalIdent{
		// 	ParentID: dot.Root.(*ast.Identifier).Name,
		// }
	}

	// TODO: handle method calls

	// handle the struct field expressions
	return &mir.FieldExpr{
		Struct:    l.lowerExpr(dot.Root),
		FieldName: dot.FieldName,
		FieldType: typing.Simplify(dot.Type()),
	}
}

// intrinsicFuncTable maps intrinsic function names to MIR Op Codes.
var intrinsicFuncTable map[string]int = map[string]int{
	"__strbytes": mir.OCStrBytes,
	"__strlen":   mir.OCStrLen,
}

// lowerFuncCall lowers a function call.
func (l *Lowerer) lowerFuncCall(call *ast.Call) mir.Expr {
	ft := typing.Simplify(call.Func.Type()).(*typing.FuncType)

	// lower the arguments
	mirOperands := make([]mir.Expr, len(call.Args))
	for i, arg := range call.Args {
		mirOperands[i] = l.lowerExpr(arg)
	}

	// handle intrinsic functions
	if ft.IntrinsicName != "" {
		opCode, ok := intrinsicFuncTable[ft.IntrinsicName]
		if !ok {
			log.Fatalf("unknown intrinsic named: `%s`\n", ft.IntrinsicName)
			return nil
		}

		return &mir.OperExpr{
			OpCode:     opCode,
			Operands:   mirOperands,
			ResultType: ft.ReturnType,
		}
	}

	// lower the function itself and make it the first MIR operand
	mirOperands = append([]mir.Expr{l.lowerExpr(call.Func)}, mirOperands...)

	// return the finalized call expression
	return &mir.OperExpr{
		OpCode:     mir.OCCall,
		Operands:   mirOperands,
		ResultType: ft.ReturnType,
	}
}

// intrinsicOpTable maps intrinsic operator names to MIR Op Codes.
var intrinsicOpTable map[string]int = map[string]int{
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

	mirOperands := make([]mir.Expr, len(operands))
	for i, operand := range operands {
		mirOperands[i] = l.lowerExpr(operand)
	}

	// handle intrinsic operators
	if operFt.IntrinsicName != "" {
		opCode, ok := intrinsicOpTable[operFt.IntrinsicName]
		if !ok {
			log.Fatalf("unknown intrinsic named: `%s`\n", operFt.IntrinsicName)
			return nil
		}

		return &mir.OperExpr{
			OpCode:     opCode,
			Operands:   mirOperands,
			ResultType: typing.Simplify(resultType),
		}
	}

	log.Fatalln("non intrinsic operators are not yet implemented")

	// add the operator function to the operands
	mirOperands = append([]mir.Expr{
		&mir.GlobalIdent{
			// TODO: parent ID
			Name:    op.Name,
			IdType:  typing.Simplify(op.Signature),
			Mutable: false,
		},
	}, mirOperands...)

	return &mir.OperExpr{
		OpCode:     mir.OCCall,
		Operands:   mirOperands,
		ResultType: typing.Simplify(resultType),
	}
}
