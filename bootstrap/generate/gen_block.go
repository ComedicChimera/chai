package generate

import (
	"chai/ast"
	"chai/depm"
	"chai/typing"
	"log"

	"github.com/llir/llvm/ir/constant"
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
		var lhsVars []value.Value
		for _, lhsExpr := range as.LHSExprs {
			if lhsVar := g.genLHSExpr(lhsExpr); lhsVar != nil {
				lhsVars = append(lhsVars, lhsVar)
			}
		}

		// evaluate all the RHS expressions
		var rhsVals []value.Value
		for _, rhsExpr := range as.RHSExprs {
			if rhsVal := g.genExpr(rhsExpr); rhsVal != nil {
				rhsVals = append(rhsVals, rhsVal)
			}
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
			g.assignTo(rhsVals[i], lhsVar, as.LHSExprs[i].Type())
		}
	} else {
		log.Fatalln("unpacking assign not implement yet")
	}
}

// genLHSExpr generates an expression on the left-hand side of an assigment or
// unary update.
func (g *Generator) genLHSExpr(expr ast.Expr) value.Value {
	if typing.IsNothing(expr.Type()) {
		return nil
	}

	// TODO: add more LHS exprs as necessary
	switch v := expr.(type) {
	case *ast.Identifier:
		// we know the identifier must be mutable (ie. it is a pointer) so we
		// just return that pointer.
		val, _ := g.lookup(v.Name)
		return val
	case *ast.Dot:
		if v.DotKind == typing.DKStaticGet {
			val, _ := g.lookup(v.Root.(*ast.Identifier).Name + v.FieldName)
			return val
		}

		// TODO: method calls

		// struct access
		structType := typing.InnerType(v.Root.Type()).(*typing.StructType)

		return g.block.NewGetElementPtr(
			g.pureConvType(structType),
			g.genExpr(v.Root),
			constant.NewInt(types.I32, 0),
			constant.NewInt(types.I32, int64(structType.FieldsByName[v.FieldName])),
		)
	}

	log.Fatalln("other lhs expressions not yet implemented")
	return nil
}
