package generate

import (
	"chai/ast"
	"chai/report"
	"chai/typing"
	"fmt"
	"log"

	"github.com/llir/llvm/ir"
	"github.com/llir/llvm/ir/constant"
	"github.com/llir/llvm/ir/enum"
	"github.com/llir/llvm/ir/value"
)

// visitDef visits a definition and recursively evaluates its dependencies
// before determining whether or not the generate it.  This ensures that the
// definitions are placed in the right order. The predicates of definitions are
// also generated.
func (g *Generator) visitDef(def ast.Def) {
	// check that the definition has not already been visited
	if inProgress, ok := g.alreadyVisited[def]; ok {
		// if it is has not finished generating, then the definition recursively
		// depends on itself and needs to be forward declared.
		if inProgress {
			g.genForwardDecl(def)
		}

		// in both cases, we do not continue with generation of this definition
		// since doing so would constitute a repeat definition.
		return
	}

	// mark the current definition as in progress
	g.alreadyVisited[def] = true

	// recursively visit its dependencies to ensure they are all fully declared
	// before it (to prevent out of order declarations)
	for dep := range def.Dependencies() {
		g.visitDef(g.defDepGraph[dep])
	}

	// generate the definition itself now that its dependencies have resolved
	g.genDef(def)

	// mark it as having completed generation
	g.alreadyVisited[def] = false
}

// genDef generates a definition and adds it to the current module.
func (g *Generator) genDef(def ast.Def) {
	switch v := def.(type) {
	case *ast.FuncDef:
		g.genFunc(v.Name, v.Args, v.Signature.ReturnType, v.Body, v.Public(), v.Annotations())
	case *ast.OperDef:
		// operators just compile to specially named functions:
		// `oper[<operator>]`
		g.genFunc(fmt.Sprintf("oper[%s]", v.Op.Name), v.Args, v.Op.Signature.(*typing.FuncType).ReturnType, v.Body, v.Public(), v.Annotations())
	case *ast.VarDecl:
		g.genGlobalVar(v)
	}
}

// noMangleAnnotations is a list of annotations that cause a function name not
// be mangled (for linking purposes)
var noMangleAnnotations = []string{
	"entry",
	"dllexport",
	"dllimport",
	"extern",
}

// genFunc generates an LLVM function definition.
func (g *Generator) genFunc(name string, args []*ast.FuncArg, rtType typing.DataType, body ast.Expr, public bool, annotations map[string]string) {
	// check for intrinsics (they aren't actually generated)
	if hasAnnot(annotations, "intrinsic") || hasAnnot(annotations, "intrinsicop") {
		return
	}

	// build the base LLVM function
	var params []*ir.Param
	for _, arg := range args {
		// prune nothing types from the arguments
		if typing.IsNothing(arg.Type) {
			continue
		}

		// TODO: by reference arguments
		if arg.ByRef {
			log.Fatalln("by reference arguments not implemented yet")
		}

		params = append(params, ir.NewParam(arg.Name, g.convType(arg.Type)))
	}

	// mangle name if necessary
	mangledName := g.globalPrefix + name
	for _, noMangleAnnot := range noMangleAnnotations {
		if hasAnnot(annotations, noMangleAnnot) {
			mangledName = name
			break
		}
	}

	llvmFunc := g.mod.NewFunc(mangledName, g.convType(rtType), params...)

	// set linkage based on visibility
	if public || hasAnnot(annotations, "extern") || hasAnnot(annotations, "entry") {
		llvmFunc.Linkage = enum.LinkageExternal
	} else {
		llvmFunc.Linkage = enum.LinkageInternal

		// TODO: consider marking internal functions as fastcc
	}

	// annotated properties
	if hasAnnot(annotations, "callconv") {
		switch annotations["callconv"] {
		case "win64":
			llvmFunc.CallingConv = enum.CallingConvWin64
		case "stdcall":
			llvmFunc.CallingConv = enum.CallingConvX86StdCall
		case "thiscall":
			llvmFunc.CallingConv = enum.CallingConvX86ThisCall
		case "c":
			llvmFunc.CallingConv = enum.CallingConvC
		default:
			report.ReportFatal("unsupported calling convention: %s", annotations["callconv"])
			return
		}
	}

	if hasAnnot(annotations, "inline") {
		llvmFunc.FuncAttrs = append(llvmFunc.FuncAttrs, enum.FuncAttrInlineHint)
	}

	// add the global declaration for the function
	g.globalScope[name] = LLVMIdent{Val: llvmFunc, Mutable: false}

	// generate the body if a body is provided
	if body != nil {
		// Chai does not use exceptions in any form and thus all functions are
		// marked `nounwind`
		llvmFunc.FuncAttrs = []ir.FuncAttribute{enum.FuncAttrNoUnwind}

		entry := llvmFunc.NewBlock("entry")

		// set the parent function of the block
		g.enclosingFunc = llvmFunc

		// declare arguments as local variables
		g.pushScope()
		defer g.popScope()

		n := 0
		for _, arg := range args {
			// nothing pruning
			if typing.IsNothing(arg.Type) {
				continue
			}

			if arg.Constant {
				g.defineLocal(arg.Name, llvmFunc.Params[n], false)
			} else {
				// mutable parameters need local allocas to be manipulated
				localArg := entry.NewAlloca(llvmFunc.Params[n].Type())
				entry.NewStore(llvmFunc.Params[n], localArg)
				g.defineLocal(
					arg.Name,
					localArg,
					true,
				)
			}

			n++
		}

		// parse the body
		g.block = entry
		result := g.genExpr(body)

		// set result to `nil` (ie. discard it) if the function returns void
		if typing.IsNothing(rtType) {
			result = nil
		}

		// generate the implicit return statement at the end.  Note that even
		// through `genExpr` may return `nil`, if the result is indeed `nil`,
		// then `NewRet` is defined to generate a ret void which is the desired
		// behavior.
		lastBlock := llvmFunc.Blocks[len(llvmFunc.Blocks)-1]
		lastBlock.NewRet(result)
	}
}

// genGlobalVar generates a global variable declaration.
func (g *Generator) genGlobalVar(vd *ast.VarDecl) {
	for _, vlist := range vd.VarLists {
		// generate the global variables themselves
		varIdents := make([]value.Value, len(vlist.Names))
		for i, name := range vlist.Names {
			glob := g.mod.NewGlobal(name, g.convType(vlist.Type))
			varIdents[i] = glob
			g.globalScope[name] = LLVMIdent{Val: glob, Mutable: true}

			// add a default `null` initializer to the global variable
			// TODO: amend to handle global variables that aren't null pointers
			glob.Init = constant.NewNull(glob.Typ)
		}

		// generate the initializers as necessary
		if vlist.Initializer != nil {
			// add the global initializer
			g.globalInits = append(g.globalInits, GlobalInit{varIdents, vlist.Initializer})
		}
	}
}

// -----------------------------------------------------------------------------

// genForwardDecl generates a forward declaration for a definition.
func (g *Generator) genForwardDecl(def ast.Def) {
	switch v := def.(type) {
	case *ast.FuncDef:
		// forward declaration for function just generates a function with no body
		g.genFunc(v.Name, v.Args, v.Signature.ReturnType, nil, v.Public(), v.Annotations())
	case *ast.OperDef:
		// same logic as functions but with operator naming scheme
		g.genFunc(fmt.Sprintf("oper[%s]", v.Op.Name), v.Args, v.Op.Signature.(*typing.FuncType).ReturnType, nil, v.Public(), v.Annotations())
	case *ast.VarDecl:
		// global variables should never be forward declared since they only
		// depend on type definitions which do not depend on them
		log.Fatalln("unable to forward declare a variable")
	}
}

// hasAnnot is a utility function to check a definition has an annotation.
func hasAnnot(annots map[string]string, name string) bool {
	_, ok := annots[name]
	return ok
}
