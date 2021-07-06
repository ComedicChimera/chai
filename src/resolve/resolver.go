package resolve

import (
	"chai/mods"
	"chai/sem"
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

	// walkers is the list of walkers for each file
	walkers map[*sem.ChaiFile]*walk.Walker
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
				publicBlock := false

				if branch.Name == "pub_block" {
					// this is at index 2 because whitespace gets pruned off
					branch = branch.BranchAt(2)
					publicBlock = true
				}

				for _, item := range branch.Content {
					defNodeInner := item.(*syntax.ASTBranch).BranchAt(0)

					if !r.extractDefinition(file, defNodeInner, publicBlock) {
						return false
					}
				}
			}
		}
	}

	// NOTE: the global imported symbol references will CHANGE

	// initialize the walkers
	r.walkers = make(map[*sem.ChaiFile]*walk.Walker)

	// resolve all independent definitions
	// TODO: implement an actually dependency resolution algorithm
	for _, def := range r.independents {
		if !r.walkDef(def) {
			return false
		}
	}

	// resolve all dependent definitions
	for _, def := range r.dependents {
		if !r.walkDef(def) {
			return false
		}
	}

	return true
}

// walkDef walks a definition
func (r *Resolver) walkDef(def *Definition) bool {
	var modifiers int
	if def.Public {
		modifiers = sem.ModPublic
	}

	if w, ok := r.walkers[def.SrcFile]; ok {
		return w.WalkDef(def.AST, modifiers, def.Annotations)
	} else {
		r.walkers[def.SrcFile] = walk.NewWalker(def.SrcFile)
		return r.walkers[def.SrcFile].WalkDef(def.AST, modifiers, def.Annotations)
	}
}
