package generate

import (
	"chai/ast"
	"chai/depm"
	"log"

	"github.com/llir/llvm/ir"
	"github.com/llir/llvm/ir/value"
)

// genBlock generates an AST do block which appends its content onto a parent
// block (such as a function body or the basic block created by an control flow
// statement)
func (g *Generator) genBlock(parentBlock *ir.Block, block *ast.Block) value.Value {
	for i, stmt := range block.Stmts {
		switch v := stmt.(type) {
		case *ast.VarDecl:
			g.genVarDecl(parentBlock, v)
		case *ast.Assign:
		case *ast.UnaryUpdate:
			log.Fatalln("unary update not implemented yet")
		default:
			{
				ev := g.genExpr(parentBlock, v)

				// last expression => return it
				if i == len(block.Stmts)-1 {
					return ev
				}
			}
		}
	}

	// reach here => block yields nothing
	return nil
}

// genVarDecl generates a variable declaration.
func (g *Generator) genVarDecl(block *ir.Block, vd *ast.VarDecl) {
	for _, vlist := range vd.VarLists {
		// first determine the value to initialize the
		// variables with
		var init value.Value
		if vlist.Initializer == nil {
			log.Fatalln("null initialization not supported yet")
		} else {
			init = g.genExpr(block, vlist.Initializer)
		}

		// initialized with nothing => variable is to be pruned
		if init == nil {
			continue
		}

		for i, name := range vlist.Names {
			if vlist.Mutabilities[i] == depm.Mutable {
				// mutable variables require a stack allocation and a store
				varPtr := block.NewAlloca(init.Type())
				block.NewStore(init, varPtr)
				g.defineLocal(name, varPtr, true)
			} else {
				// immutable variables are just initialized as their
				// initializers value (basically just SSA registers)
				g.defineLocal(name, init, false)
			}
		}
	}
}

// genAssign generates an assignment.
func (g *Generator) genAssign(block *ir.Block, as *ast.Assign) {
	log.Fatalln("assign not implement yet")
}
