package generate

import (
	"chai/ast"
	"chai/depm"

	"github.com/llir/llvm/ir"
	"github.com/llir/llvm/ir/types"
	"github.com/llir/llvm/ir/value"
)

// Generator is responsible for converting the Chai Typed AST into LLVM IR. It
// converts each package into a single LLVM module.
type Generator struct {
	// pkg is the source package being converted.
	pkg *depm.ChaiPackage

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

	// globalScope is the scope containing all global identifiers.
	globalScope map[string]LLVMIdent

	// localScopes is the stack of local scopes used during generation.
	localScopes []map[string]LLVMIdent
}

// LLVMIdent is the type used for LLVM identifiers.  It stores the value
// as well as whether or not the value has to loaded explicitly to be used.
type LLVMIdent struct {
	Val     value.Value
	Mutable bool
}

// NewGenerator creates a new generator for the given package.
func NewGenerator(pkg *depm.ChaiPackage) *Generator {
	return &Generator{
		pkg:            pkg,
		globalPrefix:   pkg.Parent.Name + pkg.ModSubPath + ".",
		mod:            ir.NewModule(),
		defDepGraph:    make(map[string]ast.Def),
		alreadyVisited: make(map[ast.Def]bool),
		globalScope:    make(map[string]LLVMIdent),
	}
}

// Generate runs the main generation algorithm for the source package. This
// generation process is assumed to always succeed: any errors here are
// considered fatal.
func (g *Generator) Generate() *ir.Module {
	// TODO: imports

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

	// generate the module
	for _, def := range g.defDepGraph {
		g.visitDef(def)
	}

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