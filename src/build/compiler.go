package build

import (
	"chai/common"
	"chai/logging"
	"chai/mods"
	"chai/resolve"
	"chai/syntax"
	"path/filepath"
	"sync"
)

// Compiler is the data structure responsible for maintaining all high-level
// state of the Chai compiler
type Compiler struct {
	// rootMod is the root module of the project being built
	rootMod *mods.ChaiModule

	// buildProfile is the profile that is being used to build the project
	buildProfile *mods.BuildProfile

	// depGraph is the graph of all the modules in the project organized by ID
	depGraph map[uint]*mods.ChaiModule

	// parsingTable is the global, shared parsing table used by all instances of
	// the Chai LALR(1) parser
	parsingTable *syntax.ParsingTable

	// coreMod is a reference to the `core` module that is imported as a part of
	// the prelude by all files
	coreMod *mods.ChaiModule
}

// NewCompiler creates a new compiler for a given root module and build profile
func NewCompiler(rootMod *mods.ChaiModule, buildProfile *mods.BuildProfile) *Compiler {
	return &Compiler{
		rootMod:      rootMod,
		buildProfile: buildProfile,
		depGraph: map[uint]*mods.ChaiModule{
			rootMod.ID: rootMod,
		},
	}
}

// Compile runs the full compilation algorithm on the root module and build
// profile. It handles all compilation errors appropriately.
func (c *Compiler) Compile() {
	if c.Analyze() {
		// TODO
	}
}

// Analyze runs just the analysis portion of the compilation algorithm.  It
// handles all errors appropriately.  This is exported for usage in the CLI (for
// editors/IDEs, etc.).  It returns a boolean indicating whether or not analysis
// was successful.
func (c *Compiler) Analyze() bool {
	// initialize the global parsing table
	// TODO: removing flag forcing grammatical rebuild every time
	ptable, err := syntax.NewParsingTable(filepath.Join(common.ChaiPath, common.GrammarPath), true)
	if err != nil {
		logging.LogConfigError("Grammar", "error building parsing table: "+err.Error())
		return false
	}
	c.parsingTable = ptable

	// load and initialize the core module
	c.coreMod, err = mods.LoadModule(filepath.Join(common.ChaiPath, "lib/std/core"), "", c.buildProfile)
	if err != nil {
		logging.LogConfigError("Module", "error loading core module: "+err.Error())
		return false
	}

	if _, ok := c.initPackage(c.coreMod, c.coreMod.ModuleRoot); !ok {
		return false
	}

	// then initialize the root package
	_, ok := c.initPackage(c.rootMod, c.rootMod.ModuleRoot)
	if !ok {
		return false
	}

	// resolve type defs, class defs, and imported symbols and process all
	// dependent definitions (functions, operators, etc.)
	batches := c.createResolutionBatches(c.rootMod)
	for _, batch := range batches {
		// each batch is resolved concurrently -- this makes the compiler far
		// more performant on large projects and allows it to take advantage of
		// any concurrent architecture provided to it.
		var wg *sync.WaitGroup
		resolutionSucceeded := true

		for _, mod := range batch {
			wg.Add(1)
			go func(mod *mods.ChaiModule) {
				defer wg.Done()
				r := resolve.NewResolver(mod)

				// don't need to use a mutex here since we are always setting
				// this boolean flag to the same value -- even if two goroutines
				// write to it at the same time, we know that the correct value
				// will always be written
				if !r.ResolveAll() {
					resolutionSucceeded = false
				}
			}(mod)
		}

		wg.Wait()

		// we don't want to continue with resolution since other modules will
		// fail to load dependencies from the batch that failed to resolve
		if !resolutionSucceeded {
			return false
		}
	}

	// TODO: validate expressions (func bodies, initializers, etc.)

	return logging.ShouldProceed()
}

// createResolutionBatches creates a list of batches of modules whose symbols
// can be resolved concurrently.  The batches at the front should be evaluated
// first.  The root module is the module to start created batches from.
func (c *Compiler) createResolutionBatches(rootMod *mods.ChaiModule) [][]*mods.ChaiModule {
	var currentBatch []*mods.ChaiModule
	var nextBatches [][]*mods.ChaiModule

	for _, mod := range rootMod.DependsOn {
		currentBatch = append(currentBatch, mod)

		for i, batch := range c.createResolutionBatches(mod) {
			if i < len(nextBatches) {
				nextBatches[i] = append(nextBatches[i], batch...)
			} else {
				nextBatches = append(nextBatches, batch)
			}
		}
	}

	return append(nextBatches, currentBatch)
}
