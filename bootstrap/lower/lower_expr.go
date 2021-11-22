package lower

import (
	"chai/ast"
	"chai/ir"
	"chai/typing"
	"log"
	"strconv"
)

// lowerExpr lowers an expression and returns a value that can be used to access
// the result if the `useResult` flag is set to true.  Otherwise, `nil` is
// returned, and the program can proceed as if the result was discarded.
func (l *Lowerer) lowerExpr(block *ir.Block, expr ast.Expr, useResult bool) ir.Value {
	var instr *ir.Instruction
	switch v := expr.(type) {
	case *ast.Call:
		// TODO
	case *ast.Identifier:
		if useResult {
			return l.lowerIdent(block, v)
		} else {
			return nil
		}
	case *ast.Literal:
		if useResult {
			return l.lowerLiteral(block, v)
		} else {
			return nil
		}
	}

	if useResult {
		return l.bindResult(block, instr)
	}

	return nil
}

// -----------------------------------------------------------------------------

// lowerIdent lowers an identifier and returns it as an IR value that can be
// directly manipulated.
func (l *Lowerer) lowerIdent(block *ir.Block, id *ast.Identifier) ir.Value {
	sval := l.lookup(id.Name)

	// handle mutable symbol loading
	if sval.IsMutable {
		return l.bindResult(block, &ir.Instruction{
			StmtBase: ir.NewStmtBase(id.Pos.StartLn),
			OpCode:   ir.OpLoad,
			TypeSpec: l.lowerType(id.Type()),
			Operands: []ir.Value{sval.Value},
		})
	}

	return sval.Value
}

// lowerLiteral lowers an AST literal and returns it as an IR value that can be
// directly manipulated.
func (l *Lowerer) lowerLiteral(block *ir.Block, lit *ast.Literal) ir.Value {
	if pt, ok := lit.Type().(typing.PrimType); ok {
		switch pt {
		case typing.PrimF32:
			{
				fval, err := strconv.ParseFloat(lit.Value, 32)
				if err != nil {
					log.Fatalln(err)
				}

				return &ir.ConstFloat{
					ValueBase: ir.NewValueBase(ir.PrimType(ir.PrimF32)),
					Val:       fval,
				}
			}
		case typing.PrimF64:
			{
				fval, err := strconv.ParseFloat(lit.Value, 64)
				if err != nil {
					log.Fatalln(err)
				}

				return &ir.ConstFloat{
					ValueBase: ir.NewValueBase(ir.PrimType(ir.PrimF64)),
					Val:       fval,
				}
			}
		case typing.PrimU8, typing.PrimI8:
			{
				ival, err := strconv.ParseInt(lit.Value, 0, 8)
				if err != nil {
					log.Fatalln(err)
				}

				return &ir.ConstInt{
					ValueBase: ir.NewValueBase(l.lowerPrimType(pt)),
					Val:       ival,
				}
			}
		case typing.PrimU16, typing.PrimI16:
			{
				ival, err := strconv.ParseInt(lit.Value, 0, 16)
				if err != nil {
					log.Fatalln(err)
				}

				return &ir.ConstInt{
					ValueBase: ir.NewValueBase(l.lowerPrimType(pt)),
					Val:       ival,
				}
			}
		case typing.PrimU32, typing.PrimI32:
			{
				ival, err := strconv.ParseInt(lit.Value, 0, 32)
				if err != nil {
					log.Fatalln(err)
				}

				return &ir.ConstInt{
					ValueBase: ir.NewValueBase(l.lowerPrimType(pt)),
					Val:       ival,
				}
			}
		case typing.PrimU64, typing.PrimI64:
			{
				ival, err := strconv.ParseInt(lit.Value, 0, 64)
				if err != nil {
					log.Fatalln(err)
				}

				return &ir.ConstInt{
					ValueBase: ir.NewValueBase(l.lowerPrimType(pt)),
					Val:       ival,
				}
			}
		case typing.PrimBool:
			{
				ival := 0
				if lit.Value == "true" {
					ival = 1
				}

				return &ir.ConstInt{
					ValueBase: ir.NewValueBase(ir.PrimType(ir.PrimBool)),
					Val:       int64(ival),
				}
			}
		case typing.PrimNothing:
			// nothings should all be pruned
			return nil
		case typing.PrimString:
			// string literals
			return l.internString(lit.Value)
		}
	}

	// only literal that is non-primitive is null
	// TODO: null
	return nil
}

// -----------------------------------------------------------------------------

// bindResult automatically binds the result of an IR instruction to a value.
func (l *Lowerer) bindResult(block *ir.Block, instr *ir.Instruction) ir.Value {
	binding := &ir.Binding{
		StmtBase: ir.NewStmtBase(instr.Line()),
		ValueID:  l.localIdentCounter,
		Instr:    instr,
	}

	l.localIdentCounter++

	block.Stmts = append(block.Stmts, binding)
	return &ir.LocalIdentifier{
		ValueBase: ir.NewValueBase(instr.TypeSpec),
		ID:        binding.ValueID,
	}
}
