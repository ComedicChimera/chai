package generate

import (
	"chai/ast"
	"chai/typing"

	"github.com/llir/llvm/ir"
	"github.com/llir/llvm/ir/value"
)

// genIfExpr generates an if expression and conditionally returns a value if its
// result is used and is not nothing.
func (g *Generator) genIfExpr(ifExpr *ast.IfExpr) value.Value {
	// incoming will be used to produce the resulting `phi` node if this block
	// yields a value.
	var incoming []*ir.Incoming

	// endBlock is the block that all the branches will jump to to end the block
	endBlock := g.appendBlock()

	for i, condBranch := range ifExpr.CondBranches {
		ifBlock := g.appendBlock()

		// if there is no else, then the final "else" block is the ending block.
		var elseBlock *ir.Block
		if i < len(ifExpr.CondBranches) && ifExpr.ElseBranch == nil {
			elseBlock = endBlock
		} else {
			elseBlock = g.appendBlock()
		}

		// push a local scope for the block
		g.pushScope()

		// handle the variable declaration if it exists
		if condBranch.HeaderVarDecl != nil {
			g.genVarDecl(condBranch.HeaderVarDecl)
		}

		// generate the conditional expression
		cexpr := g.genExpr(condBranch.Cond)

		// add the conditional branch
		g.block.NewCondBr(cexpr, ifBlock, elseBlock)

		// if body
		g.block = ifBlock
		bodyResult := g.genExpr(condBranch.Body)
		incoming = append(incoming, ir.NewIncoming(bodyResult, g.block))
		g.block.NewBr(endBlock)

		// position the generator in the else => that is where the next branch's
		// logic will be generated: translating `if`, `elif` into `if` and `else
		// if`
		g.block = elseBlock

		// pop the local scope
		g.popScope()
	}

	// handle else expressions
	if ifExpr.ElseBranch != nil {
		// we know the generator is already positioned on the else block so we
		// can just generate it here.
		bodyResult := g.genExpr(ifExpr.ElseBranch)
		incoming = append(incoming, ir.NewIncoming(bodyResult, g.block))
		g.block.NewBr(endBlock)

		// move the generator to the end block (which in this case is not the
		// same as the else block)
		g.block = endBlock
	}

	// if the block returns a value which is not nothing than we need to generate
	// and return a phi node accumulating the results of the different branches.
	// Otherwise, we return nothing.
	if ifExpr.Type() != nil && !typing.IsNothing(ifExpr.Type()) {
		return g.block.NewPhi(incoming...)
	}

	// otherwise, we just return `nil`
	return nil
}

// genWhileExpr generates a while loop (NOT handling loop generators).
func (g *Generator) genWhileExpr(whileExpr *ast.WhileExpr) {
	// TODO: loop patterns and update clause, after blocks

	loopHeader := g.block
	bodyBlock := g.appendBlock()
	endBlock := g.appendBlock()

	// push a scope for the loop body
	g.pushScope()

	// generate the header variable declaration as necessary
	if whileExpr.HeaderVarDecl != nil {
		g.genVarDecl(whileExpr.HeaderVarDecl)
	}

	// generate the condition and conditional branch
	condVal := g.genExpr(whileExpr.Cond)
	g.block.NewCondBr(condVal, bodyBlock, endBlock)

	// generate the body
	g.block = bodyBlock
	g.genExpr(whileExpr.Body)
	g.block.NewBr(loopHeader)

	// set the block to the end block of the loop
	g.block = endBlock

	// pop the local scope
	g.popScope()
}
