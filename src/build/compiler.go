package build

import (
	"chai/common"
	"chai/logging"
	"chai/mods"
	"chai/resolve"
	"chai/syntax"
	"chai/walk"
	"fmt"
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
	logging.LogCompileHeader(
		c.buildProfile.TargetOS+"/"+c.buildProfile.TargetArch,
		c.rootMod.ShouldCache,
	)

	if c.Analyze() {
		// TODO
	} else {
		// close unfinished working phase
		logging.LogEndPhase()
	}

	// log ending compilation message
	logging.LogCompilationFinished()
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
	c.depGraph[c.coreMod.ID] = c.coreMod

	// log begin phase 1
	logging.LogBeginPhase("Parsing")

	// initialize the core module
	if _, ok := c.initPackage(c.coreMod, c.coreMod.ModuleRoot); !ok {
		return false
	}

	// then initialize the root package
	_, ok := c.initPackage(c.rootMod, c.rootMod.ModuleRoot)
	if !ok {
		return false
	}

	// log end phase 1; begin phase 2
	logging.LogEndPhase()
	logging.LogBeginPhase("Transforming")

	// resolve type defs, class defs, and imported symbols and process all
	// dependent definitions (functions, operators, etc.), then validate
	// predicates (function bodies, etc) using the same resolution batch
	// grouping (enabling us to process them concurrently)
	batches := c.createResolutionBatches(c.rootMod)
	for _, batch := range batches {
		// each batch is resolved concurrently -- this makes the compiler far
		// more performant on large projects and allows it to take advantage of
		// any concurrent architecture provided to it.
		wg := &sync.WaitGroup{}
		resolutionSucceeded := true

		for _, mod := range batch {
			wg.Add(1)
			go func(mod *mods.ChaiModule) {
				defer wg.Done()
				r := resolve.NewResolver(mod, c.depGraph)

				if r.ResolveAll() {
					// if we can resolve all symbols, then we know all operators
					// can be imported and all predicates can be walked
					for _, pkg := range mod.Packages() {
						for _, file := range pkg.Files {
							// import public operators -- do this after resolution since
							// only then will all the operators actually be defined
							for ipkg, pathPos := range file.ImportedPackages {
								if err := file.ImportOperators(ipkg); err != nil {
									logging.LogCompileError(
										file.LogContext,
										fmt.Sprintf("multiple conflicting overloads for `%s` operator", err.Error()),
										logging.LMKImport,
										pathPos,
									)
								}

								// we can actually continue walking here since
								// operator conflicts generally don't cause a
								// cascade of errors
							}

							// walk all predicates
							w := walk.NewWalker(file)
							w.WalkPredicates(file.Root)
						}
					}
				} else {
					// don't need to use a mutex here since we are always
					// setting this boolean flag to the same value -- even if
					// two goroutines write to it at the same time, we know that
					// the correct value will always be written
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

	// log end of the transformation phase
	logging.LogEndPhase()

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
