package lower

import (
	"chai/ast"
	"chai/mir"
)

// visit lowers a single definition by first recursively visiting all its
// dependencies and then it adds the fully lowered definition to the MIR bundle.
func (l *Lowerer) visit(def ast.Def) {
	// if the definition has already been visited
	if mdef, ok := l.alreadyVisited[def]; ok {
		if mdef != nil {
			// the definition is in the process of being added meaning it recursively
			// depends on itself.  Therefore, we need to add it as a forward declaration
			// to break the cycle and return.
			l.bundle.Forwards = append(l.bundle.Forwards, mdef)
		}

		// if it has already been added, then we just return (nothing more to do)
		return
	}

	// convert the definition to a MIR definition
	mdef := l.lowerDef(def)

	// add it to the map of already visited to prevent infinite recursion and
	// allow for forward definitions as necessary
	l.alreadyVisited[def] = mdef

	// visit its dependencies
	for name := range def.Dependencies() {
		l.visit(l.defDepGraph[name])
	}

	// TODO: lower function bodies and define MIR functions

	// flag the definition as completed before returning
	l.alreadyVisited[def] = nil
}

// lowerDef lowers a definition into a MIR definition *without* lowering its body.
func (l *Lowerer) lowerDef(def ast.Def) mir.MIRDef {
	// TODO
	return nil
}
