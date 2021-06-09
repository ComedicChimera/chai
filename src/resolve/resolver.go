package resolve

import (
	"chai/mods"
)

// Resolver is main data structure used to facilitate symbol resolution within a
// single module.
type Resolver struct {
	// mod is the module this resolver is working on
	mod *mods.ChaiModule

	// depGraph is the module dependency graph
	depGraph map[uint]*mods.ChaiModule
}

func NewResolver(mod *mods.ChaiModule, depg map[uint]*mods.ChaiModule) *Resolver {
	return &Resolver{
		mod:      mod,
		depGraph: depg,
	}
}

// ResolveAll attempts to resolve all top level definitions within a single
// module and generates HIR trees for all those definitions.  It returns a
// boolean indicating whether or not resolution succeeded.
func (r *Resolver) ResolveAll() bool {
	for _, pkg := range r.mod.Packages() {
		for _, file := range pkg.Files {
			// extract and process all imported definitions
			for _, sym := range file.ImportedSymbols {
				if !r.processSymbolImport(file, sym) {
					return false
				}
			}

			// TODO: extract all independent definitions (types, classes, etc.)
		}
	}

	// NOTE: the global imported symbol references will CHANGE

	// TODO: resolve all independent definitions
	// TODO: resolve all dependent definitions
	return true
}
