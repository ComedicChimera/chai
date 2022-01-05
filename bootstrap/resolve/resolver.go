package resolve

import "chai/depm"

// Resolver is responsible for resolving all global symbol dependencies: namely,
// those on imported symbols and globally-defined types.  It also checks for
// recursive types.  This is run before type checking so local symbols are not
// processed until after resolution is completed.
type Resolver struct {
	pkgList []*depm.ChaiPackage
}

// NewResolver creates a new resolver for the given dependency graph.
func NewResolver(pkgList []*depm.ChaiPackage) *Resolver {
	return &Resolver{pkgList: pkgList}
}

// Resolve runs the main resolution algorithm.
func (r *Resolver) Resolve() bool {
	// resolve imports
	if !r.resolveImports() {
		return false
	}

	// TODO: resolved named types and check for recursive types

	return true
}
