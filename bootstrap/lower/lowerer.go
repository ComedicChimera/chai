package lower

import (
	"chai/ast"
	"chai/depm"
	"chai/mir"
	"strconv"
)

// Lowerer is the construct responsible for converting the AST into MIR.
type Lowerer struct {
	pkg    *depm.ChaiPackage
	bundle *mir.MIRBundle

	// defDepGraph is a graph of definitions organized by the names they define.
	defDepGraph map[string]ast.Def

	// alreadyVisited is the list of definitions already added used to determine
	// in what order definitions should be processed and to prevent definitions
	// from being visited multiple times.  The stored value of this map is the
	// MIR definition.  This will be `nil` if the body has *already* been
	// processed.
	alreadyVisited map[ast.Def]mir.Def

	// globalPrefix is the prefix that is added before all global symbols to
	// prevent name collisions.  It ends with a `.` and thus able to be directly
	// concatenated to the front of all global symbols.
	globalPrefix string

	// scopes is the stack of local variable scopes which each local variable
	// scope is map of the Chai name of the variable mapped to whether or not it
	// is a constant.
	scopes []map[string]bool

	// tempCounter is a counter for temporary names used in functions.
	tempCounter int
}

// NewLowerer creates a new lowerer for a given package.
func NewLowerer(pkg *depm.ChaiPackage) *Lowerer {
	return &Lowerer{
		pkg:            pkg,
		bundle:         &mir.MIRBundle{Name: pkg.Parent.Name + pkg.ModSubPath},
		defDepGraph:    make(map[string]ast.Def),
		alreadyVisited: make(map[ast.Def]mir.Def),
		globalPrefix:   pkg.Parent.Name + pkg.ModSubPath + ".",
	}
}

// Lower converts the package into a MIR bundle.
func (l *Lowerer) Lower() *mir.MIRBundle {
	// build the graph of definitions
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

	return l.bundle
}

// -----------------------------------------------------------------------------

// getTempName gets a temporary name from the counter.
func (l *Lowerer) getTempName() string {
	l.tempCounter++
	return strconv.Itoa(l.tempCounter)
}

// isMutable tests of a symbol with a given name is mutable.
func (l *Lowerer) isMutable(name string) bool {
	// scopes in reverse order to implement shadowing
	for i := len(l.scopes) - 1; i >= 0; i-- {
		if mut, ok := l.scopes[i][name]; ok {
			return mut
		}
	}

	// TODO: local symbol imports

	return l.pkg.SymbolTable[name].Mutability == depm.Mutable
}

// setMutable defines a new mutability for a local variable.
func (l *Lowerer) setMutable(name string, mut bool) {
	l.scopes[len(l.scopes)-1][name] = mut
}

// pushScope pushes a scope onto the local scope stack.
func (l *Lowerer) pushScope() {
	l.scopes = append(l.scopes, make(map[string]bool))
}

// popScope pops a scope from the local scope stack.
func (l *Lowerer) popScope() {
	l.scopes = l.scopes[:len(l.scopes)-1]
}
