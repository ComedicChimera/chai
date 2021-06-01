package build

import (
	"chai/common"
	"chai/logging"
	"chai/mods"
	"chai/syntax"
	"path/filepath"
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

	// start by initializing the root package
	rootpkg, ok := c.initPackage(c.rootMod, c.rootMod.ModuleRoot)
	if !ok {
		return false
	}

	// initialize the root package's dependencies
	if !c.initDependencies(c.rootMod, rootpkg) {
		return false
	}

	// TODO: check for import cycles across modules

	// TODO: resolve type defs, class defs, and imported symbols and process all
	// dependent definitions (functions, operators, etc.)

	// TODO: validate expressions (func bodies, initializers, etc.)

	return logging.ShouldProceed()
}