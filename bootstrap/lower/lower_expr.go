package lower

import (
	"chai/ast"
	"chai/mir"
	"chai/typing"
	"log"
)

// lowerExpr lowers an expression and returns the a value referencing the last
// item generated from the expression unless the expression evaluated to
// `nothing` in which case `nil` is returned.  It takes a pointer to a list of
// statements to append onto as it lowers.
func (l *Lowerer) lowerExpr(stmts *[]mir.Stmt, expr ast.Expr) mir.Value {
	var result mir.Instruction
	switch v := expr.(type) {
	case *ast.Cast:
		// you can't cast from nothing so no reason to nothing check here
		result = &mir.Cast{
			Value:    l.lowerExpr(stmts, v.Src),
			DestType: v.Type(),
		}
	case *ast.Call:
		result = l.lowerCall(stmts, v)
	case *ast.Identifier:
		return &mir.Identifier{
			Name:    v.Name,
			Mutable: l.isMutable(v.Name),
		}
	case *ast.Literal:
		return &mir.Constant{
			Value: v.Value,
			Type:  v.Type(),
		}
	default:
		// TODO: other expressions
		log.Fatalln("unsupported expression")
		return nil
	}

	// nothing pruning
	if pt, ok := expr.Type().(typing.PrimType); ok && pt == typing.PrimNothing {
		return nil
	}

	// bind the result to a temporary and return a reference to that temporary
	temp := l.getTempName()
	*stmts = append(*stmts, &mir.Binding{
		Name: temp,
		RHS:  result,
	})
	return &mir.Identifier{Name: temp, Mutable: false}
}

// lowerCall lowers a call expression to a MIR instruction.
func (l *Lowerer) lowerCall(stmts *[]mir.Stmt, acall *ast.Call) mir.Instruction {
	// we know functions must be identifiers
	funcId := l.lowerExpr(stmts, acall.Func).(*mir.Identifier)

	// lower the arguments
	var args []mir.Value
	for _, arg := range acall.Args {
		val := l.lowerExpr(stmts, arg)

		// in arguments, nils are essentially treated like they aren't there at
		// all: all nothing arguments and values get completely pruned
		if val != nil {
			args = append(args, val)
		}
	}

	return &mir.Call{
		FuncName: funcId.Name,
		Args:     args,
	}
}
