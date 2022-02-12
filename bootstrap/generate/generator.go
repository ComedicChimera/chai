package generate

import (
	"chai/ast"
	"chai/depm"
	"chai/report"
	"fmt"

	"github.com/llir/llvm/ir"
	"github.com/llir/llvm/ir/enum"
	"github.com/llir/llvm/ir/types"
	"github.com/llir/llvm/ir/value"
)

// GlobalInit represents a global initializer (for any number of variables
// which are initialized to the same value)
type GlobalInit struct {
	Globals []value.Value
	Expr    ast.Expr
}

// LLVMIdent is the type used for LLVM identifiers.  It stores the value
// as well as whether or not the value has to loaded explicitly to be used.
type LLVMIdent struct {
	Val     value.Value
	Mutable bool
}

// ASTWrappedLLVMVal is special AST node that wraps an already defined LLVM
// value as an AST expression.  This is used when the compiler wants to "inject"
// a value that it has calculated into the user AST (such as in compound
// assignment) often to avoid redundant computation.
type ASTWrappedLLVMVal struct {
	ast.ExprBase
	Val value.Value
}

func (awlv *ASTWrappedLLVMVal) Position() *report.TextPosition {
	return nil
}

// -----------------------------------------------------------------------------

// Generator is responsible for converting the Chai Typed AST into LLVM IR. It
// converts each package into a single LLVM module.
type Generator struct {
	// startBuilder is the global start builder for the project.
	startBuilder *StartBuilder

	// pkg is the source package being converted.
	pkg *depm.ChaiPackage

	// isRoot is a boolean indicating if this package is the root package.
	isRoot bool

	// globalPrefix is prefix prepended to all global symbols to prevent
	// namespace collisions between the symbols of different packages during
	// linking.
	globalPrefix string

	// mod is the LLVM module being generated.
	mod *ir.Module

	// defDepGraph is a graph of all of the definitions in Chai organized by the
	// names they define.  This graph is used to put the definitions in the
	// right order in the resulting LLVM module.
	defDepGraph map[string]ast.Def

	// alreadyVisited stores the definitions that have already been generated or
	// are in the process of being generated.  The key is the definition pointer
	// and the value is a boolean flag indicating whether or not the definition
	// is still undergoing generation: true if in progress, false if done.
	alreadyVisited map[ast.Def]bool

	// stringType stores the type used `string`.
	stringType types.Type

	// globalCounter is a counter used to generate anonymous globals such as
	// those for interned strings.
	globalCounter int

	// enclosingFunc is function enclosing the block being compiled.
	enclosingFunc *ir.Func

	// globalScope is the scope containing all global values.
	globalScope map[string]LLVMIdent

	// localScopes is the stack of local scopes used during generation.
	localScopes []map[string]LLVMIdent

	// globalTypes is a table containing all the globally defined types in the
	// package.
	globalTypes map[string]types.Type

	// initFunc is the global initialization function for the package. It
	// initializes global variables and calls the package's `init` function as
	// necessary.  It is called by the main initialization function.
	initFunc *ir.Func

	// globalInits is the list of global globalInits to be generated.
	globalInits []GlobalInit

	// block stores the current block begin generated.
	block *ir.Block
}

// NewGenerator creates a new generator for the given package.
func NewGenerator(sb *StartBuilder, pkg *depm.ChaiPackage, isRoot bool) *Generator {
	return &Generator{
		startBuilder:   sb,
		pkg:            pkg,
		isRoot:         isRoot,
		globalPrefix:   fmt.Sprintf("p%d.", pkg.ID),
		mod:            ir.NewModule(),
		defDepGraph:    make(map[string]ast.Def),
		alreadyVisited: make(map[ast.Def]bool),
		globalScope:    make(map[string]LLVMIdent),
		globalTypes:    make(map[string]types.Type),
	}
}

// Generate runs the main generation algorithm for the source package. This
// generation process is assumed to always succeed: any errors here are
// considered fatal.
func (g *Generator) Generate() *ir.Module {
	// generate the global package init function
	g.initFunc = g.mod.NewFunc(g.globalPrefix+"$__init", types.Void)
	g.initFunc.Linkage = enum.LinkageExternal
	g.startBuilder.AddInitFunc(g.initFunc.GlobalName)
	g.initFunc.NewBlock("entry")

	// add all the definitions in the package to the dependency graph
	for _, file := range g.pkg.Files {
		for _, def := range file.Defs {
			for _, name := range def.Names() {
				g.defDepGraph[name] = def
			}
		}
	}

	// TEMPORARY: define the string type: {*i8, u32}
	g.stringType = g.mod.NewTypeDef("string", types.NewStruct(types.I8Ptr, types.I32))

	// declare all imports
	for _, pkgImport := range g.pkg.ImportedPackages {
		// calculate the global prefix for the imported package's symbols
		importPrefix := fmt.Sprintf("p%d.", pkgImport.Pkg.ID)

		// imported symbols
		// TODO: fix to handle implicitly imported symbols
		for _, sym := range pkgImport.Symbols {
			g.genSymbolImport(importPrefix, sym)
		}

		// TODO: imported operators
	}

	// generate the package
	for _, def := range g.defDepGraph {
		g.visitDef(def)
	}

	// generate global initializers at the very end
	initBlock := g.initFunc.Blocks[0]
	for _, ginit := range g.globalInits {
		// generate the initialization expression itself
		g.block = initBlock
		initExpr := g.genExpr(ginit.Expr)

		// store it in all the variables
		for _, global := range ginit.Globals {
			initBlock.NewStore(initExpr, global)
		}
	}

	// TODO: add in call to `init` if the package has one.

	// terminate the global package init function
	initBlock.NewRet(nil)

	// return the completed module
	return g.mod
}

// -----------------------------------------------------------------------------

// pushScope pushes a new local scope onto the scope stack.
func (g *Generator) pushScope() {
	g.localScopes = append(g.localScopes, make(map[string]LLVMIdent))
}

// popScope pops a local scope off of the local scope stack.
func (g *Generator) popScope() {
	g.localScopes = g.localScopes[:len(g.localScopes)-1]
}

// defineLocal defines a local variable.
func (g *Generator) defineLocal(name string, val value.Value, mutable bool) {
	g.localScopes[len(g.localScopes)-1][name] = LLVMIdent{val, mutable}
}

// lookup looks up a new symbol.  The returned boolean indicates if the returned
// value is mutable.
func (g *Generator) lookup(name string) (value.Value, bool) {
	// iterate through scopes in reverse order to implement shadowing.
	for i := len(g.localScopes) - 1; i >= 0; i-- {
		if ident, ok := g.localScopes[i][name]; ok {
			return ident.Val, ident.Mutable
		}
	}

	globalIdent := g.globalScope[name]
	return globalIdent.Val, globalIdent.Mutable
}

// -----------------------------------------------------------------------------

// appendBlock adds a new basic block to the current function.  It does *not*
// set the current block to this new block.
func (g *Generator) appendBlock() *ir.Block {
	return g.enclosingFunc.NewBlock(fmt.Sprintf("bb%d", len(g.enclosingFunc.Blocks)))
}
