package mir

import (
	"chai/ast"
	"chai/depm"
)

// Lowerer is responsible for converting the typed AST into Chai MIR.
type Lowerer struct {
	pkg *depm.ChaiPackage
	b   *Bundle

	// defDepGraph is a graph of all of the definitions in Chai organized by the
	// names they define.  This graph is used to put the definitions in the
	// right order in the MIR bundle and resulting LLVM module.
	defDepGraph map[string]ast.Def

	// alreadyVisited stores the definitions that have already been lowered or
	// are in the process of being lowered.  The key is the definition pointer
	// and the value is a boolean flag indicating whether or not the definition
	// is still undergoing lowering: true if in progress, false if done.
	alreadyVisited map[ast.Def]bool

	// enclosingFunc is the definition of the current enclosing function.
	enclosingFunc *FuncDef
}

// Lower converts a package into a MIR bundle.
func Lower(pkg *depm.ChaiPackage) *Bundle {
	l := &Lowerer{
		pkg:            pkg,
		b:              NewBundle(pkg.ID),
		defDepGraph:    make(map[string]ast.Def),
		alreadyVisited: make(map[ast.Def]bool),
	}

	return l.Lower()
}

// Lower generates a MIR bundle using the Lowerer.
func (l *Lowerer) Lower() *Bundle {
	// TODO: add all the imports to the bundle

	// add all the definitions in the package to the dependency graph
	for _, file := range l.pkg.Files {
		for _, def := range file.Defs {
			for _, name := range def.Names() {
				l.defDepGraph[name] = def
			}
		}
	}

	// lower the package
	for _, def := range l.defDepGraph {
		l.visitDef(def)
	}

	return l.b
}
