package lower

import (
	"chai/ast"
	"chai/depm"
	"chai/mir"
	"chai/typing"
)

// Lowerer is responsible for converting a Chai package into a MIR bundle.
type Lowerer struct {
	pkg *depm.ChaiPackage
	b   *mir.Bundle

	// defDepGraph is a graph of all of the definitions in Chai organized by the
	// names they define.  This graph is used to put the definitions in the
	// right order in the resulting MIR bundle.
	defDepGraph map[string]ast.Def

	// alreadyVisited stores the definitions that have already been lowered or
	// are in the process of being lowered.  The key is the definition pointer
	// and the value is a boolean flag indicating whether or not the definition
	// is still undergoing lowering: true if in progress, false if done.
	alreadyVisited map[ast.Def]bool

	// locals is table of lowered local variables.
	locals map[string]typing.DataType

	// scopeStack is the stack of local scopes used during lowering: it maps
	// Chai identifiers to their actual local names.
	scopeStack []map[string]string
}

// Lower lowers a package into a MIR bundle.
func Lower(pkg *depm.ChaiPackage) *mir.Bundle {
	l := Lowerer{
		pkg:            pkg,
		b:              mir.NewBundle(pkg.ID),
		defDepGraph:    make(map[string]ast.Def),
		alreadyVisited: make(map[ast.Def]bool),
	}

	l.lower()

	return l.b
}

// -----------------------------------------------------------------------------

// lower runs the main lowering algorithm to convert a package into a MIR bundle.
func (l *Lowerer) lower() {
	// TODO: lower imports

	// build the definition graph
	for _, file := range l.pkg.Files {
		for _, def := range file.Defs {
			for _, name := range def.Names() {
				l.defDepGraph[name] = def
			}
		}
	}

	// visit all definitions
	for _, def := range l.defDepGraph {
		l.visit(def)
	}
}

// visit visits a definition node and recursively evaluates its dependencies
// before determining whether or not to lower it.  This ensures that the
// definitions are placed in the right order. The predicates of definitions are
// also lowered.
func (l *Lowerer) visit(def ast.Def) {
	// check that the definition has not already been visited
	if inProgress, ok := l.alreadyVisited[def]; ok {
		// if it is has not finished lowering, then the definition recursively
		// depends on itself and needs to be forward declared.
		if inProgress {
			l.forwardDecl(def)
		}

		// in both cases, we do not continue lowering this definition since
		// doing so would constitute a repeat definition.
		return
	}

	// mark the current definition as in progress
	l.alreadyVisited[def] = true

	// recursively visit its dependencies to ensure they are all fully declared
	// before it (to prevent out of order declarations)
	for dep := range def.Dependencies() {
		l.visit(l.defDepGraph[dep])
	}

	// lower the definition itself now that its dependencies have resolved
	l.lowerDef(def)

	// mark it as having been lowered
	l.alreadyVisited[def] = false
}

// -----------------------------------------------------------------------------

// currScope returns the current scope on the scope stack.
func (l *Lowerer) currScope() map[string]string {
	return l.scopeStack[len(l.scopeStack)-1]
}

// pushScope pushes a new local scope onto the scope stack.
func (l *Lowerer) pushScope() {
	l.scopeStack = append(l.scopeStack, make(map[string]string))
}

// popScope pops a local scope off of the local scope stack.
func (l *Lowerer) popScope() {
	l.scopeStack = l.scopeStack[:len(l.scopeStack)-1]
}
