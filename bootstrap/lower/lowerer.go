package lower

import (
	"chai/ast"
	"chai/depm"
	"chai/mir"
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
}

// NewLowerer creates a new lowerer for a given package.
func NewLowerer(pkg *depm.ChaiPackage) *Lowerer {
	return &Lowerer{
		pkg:            pkg,
		bundle:         &mir.MIRBundle{},
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
