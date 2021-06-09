package resolve

import (
	"chai/deps"
	"chai/mods"
	"chai/syntax"
	"chai/walk"
)

// Resolver is main data structure used to facilitate symbol resolution within a
// single module.
type Resolver struct {
	// mod is the module this resolver is working on
	mod *mods.ChaiModule

	// depGraph is the module dependency graph
	depGraph map[uint]*mods.ChaiModule

	// independents is the list of definitions that can resolve other symbols
	// such as type definitions or class definitions
	independents []*Definition

	// dependents is the list of definitions that can only be resolved by other
	// symbols such as functions and variables (which don't resolve other
	// symbols)
	dependents []*Definition
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

			// extract all definitions (types, functions, etc.)
			for _, item := range file.AST.Content {
				branch := item.(*syntax.ASTBranch)

				if branch.Name == "pub_block" {
					branch = branch.BranchAt(3)
				}

				for _, item := range branch.Content {
					defNodeInner := item.(*syntax.ASTBranch).BranchAt(0)

					if !r.extractDefinition(file, defNodeInner) {
						return false
					}
				}
			}
		}
	}

	// NOTE: the global imported symbol references will CHANGE

	walkers := make(map[*deps.ChaiFile]*walk.Walker)

	// TODO: resolve all independent definitions

	// resolve all dependent definitions
	for _, def := range r.dependents {
		if w, ok := walkers[def.SrcFile]; ok {
			w.WalkDef(def.AST, def.Public, def.Annotations)
		} else {
			walkers[def.SrcFile] = walk.NewWalker(def.SrcFile)
			walkers[def.SrcFile].WalkDef(def.AST, def.Public, def.Annotations)
		}
	}
	return true
}
