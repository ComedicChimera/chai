package resolve

import "chai/mods"

// Resolver is main data structure used to facilitate symbol resolution within a
// single module.
type Resolver struct {
	// mod is the module this resolver is working on
	mod *mods.ChaiModule
}

func NewResolver(mod *mods.ChaiModule) *Resolver {
	return &Resolver{
		mod: mod,
	}
}

// ResolveAll attempts to resolve all top level definitions within a single
// module and generates HIR trees for all those definitions.  It returns a
// boolean indicating whether or not resolution succeeded.
func (r *Resolver) ResolveAll() bool {
	return false
}
