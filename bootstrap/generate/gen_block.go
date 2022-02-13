package generate

import (
	"chai/ast"
	"chai/depm"
	"chai/typing"
	"log"

	"github.com/llir/llvm/ir/types"
	"github.com/llir/llvm/ir/value"
)

// genBlock generates an AST do block which appends its content onto a parent
// block (such as a function body or the basic block created by an control flow
// statement)
func (g *Generator) genBlock(block *ast.Block) value.Value {
	for i, stmt := range block.Stmts {
		switch v := stmt.(type) {
		case *ast.VarDecl:
			g.genVarDecl(v)
		case *ast.Assign:
			g.genAssign(v)
		case *ast.UnaryUpdate:
			log.Fatalln("unary update not implemented yet")
		default:
			{
				ev := g.genExpr(v)

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
func (g *Generator) genVarDecl(vd *ast.VarDecl) {
	for _, vlist := range vd.VarLists {
		// first determine the value to initialize the
		// variables with
		var init value.Value
		if vlist.Initializer == nil {
			init = g.genNull(vlist.Type)
			// log.Fatalln("null initialization not supported yet")
		} else {
			init = g.genExpr(vlist.Initializer)
		}

		// initialized with nothing => variable is to be pruned
		if init == nil {
			continue
		}

		for i, name := range vlist.Names {
			if vlist.Mutabilities[i] == depm.Mutable {
				// mutable variables require a stack allocation and a store.
				// However, the `alloca` is always placed at the start of the
				// function to avoid loop invariant computation.
				varPtr := g.enclosingFunc.Blocks[0].NewAlloca(g.convTypeNoPtr(vlist.Type))
				g.assignTo(init, varPtr, vlist.Type)
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
func (g *Generator) genAssign(as *ast.Assign) {
	// non-unpacking assignment
	if len(as.LHSExprs) == len(as.RHSExprs) {
		// evaluate all the LHS expressions.  We do have to evaluate these first
		// since it is technically possible for the RHS to effect the LHS via
		// indirection.
		lhsVars := make([]value.Value, len(as.LHSExprs))
		for i, lhsExpr := range as.LHSExprs {
			lhsVars[i] = g.genLHSExpr(lhsExpr)
		}

		// evaluate all the RHS expressions
		rhsVals := make([]value.Value, len(as.RHSExprs))
		for i, rhsExpr := range as.RHSExprs {
			rhsVals[i] = g.genExpr(rhsExpr)
		}

		// if this a compound operation, then we need to apply the operator
		// first.  Here is where we rewrite `a += b` as `a = a + b` (although we
		// do it a bit more cleverly than you might expect)
		if as.Oper != nil {
			for i, lhsVar := range lhsVars {
				rhsVals[i] = g.genOpCall(
					*as.Oper, &ASTWrappedLLVMVal{
						ExprBase: ast.NewExprBase(as.LHSExprs[i].Type(), ast.LValue),
						Val:      g.block.NewLoad(lhsVar.Type().(*types.PointerType).ElemType, lhsVar),
					}, as.RHSExprs[i],
				)
			}
		}

		// create the assignments
		for i, lhsVar := range lhsVars {
			if rhsVals[i] == nil {
				continue
			}

			g.assignTo(rhsVals[i], lhsVar, as.LHSExprs[i].Type())
		}
	} else {
		log.Fatalln("unpacking assign not implement yet")
	}
}

// genLHSExpr generates an expression on the left-hand side of an assigment or
// unary update.
func (g *Generator) genLHSExpr(expr ast.Expr) value.Value {
	// TODO: add more LHS exprs as necessary
	switch v := expr.(type) {
	case *ast.Identifier:
		// we know the identifier must be mutable (ie. it is a pointer) so we
		// just return that pointer.
		val, _ := g.lookup(v.Name)
		return val
	case *ast.Dot:
		if v.IsStaticGet {
			val, _ := g.lookup(v.Root.(*ast.Identifier).Name + v.FieldName)
			return val
		}

		// TODO: method calls

		// return g.block.NewGetElementPtr(g.convTypeNoPtr(v.Root.Type()), g.genExpr(v.Root))
	}

	log.Fatalln("other lhs expressions not yet implemented")
	return nil
}

// -----------------------------------------------------------------------------

// assignTo generates a code snippet for assigning an RHS value into an LHS value.
func (g *Generator) assignTo(lhs, rhs value.Value, dt typing.DataType) {
	if isPtrType(dt) {
		// TODO: memcopy
	} else {
		g.block.NewStore(rhs, lhs)
	}
}

// returnFromFunc generate a code snippet for returning a value from a function.
func (g *Generator) returnFromFunc(val value.Value, dt typing.DataType) {
	if isPtrType(dt) {
		// TODO: memcpy

		g.block.NewRet(nil)
	} else {
		g.block.NewRet(val)
	}
}
