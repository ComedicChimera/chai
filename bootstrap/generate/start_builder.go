package generate

import (
	"chai/depm"
	"sync"

	"github.com/llir/llvm/ir"
)

// StartBuilder is responsible for constructing the main initialization function
// (ie. the initialization function that calls all other initialization
// functions).
type StartBuilder struct {
	// m is used to prevent race conditions to during start building.
	m *sync.Mutex

	// rootPkgID is the ID of the root package of the project.
	rootPkgID uint

	// pkgTable is a table of all the packages in the project organized by
	// package ID.  This is used to determine in what order to call the
	// initialization functions.
	pkgTable map[uint]*depm.ChaiPackage

	// initFuncs is a map of the names of the initialization functions used in
	// each package.  These are used to generate external references so that
	// these functions can be called from the main initialization function.
	initFuncs map[uint]string

	// initMod is the LLVM module that contains the source for the main
	// initialization function.  This is set to the root LLVM module for the
	// project.
	initMod *ir.Module
}

// NewStartBuilder returns a new StartBuilder for the project.
func NewStartBuilder(depGraph map[uint]*depm.ChaiModule, rootPkgID uint) *StartBuilder {
	pkgTable := make(map[uint]*depm.ChaiPackage)
	for _, mod := range depGraph {
		for _, pkg := range mod.Packages() {
			pkgTable[pkg.ID] = pkg
		}
	}

	return &StartBuilder{
		rootPkgID: rootPkgID,
		pkgTable:  pkgTable,
		initFuncs: make(map[uint]string),
		initMod:   nil,
	}
}

// IsRoot returns whether or not a package is the root package for the project.
func (sb *StartBuilder) IsRoot(pkg *depm.ChaiModule) bool {
	return sb.rootPkgID == pkg.ID
}

// SetInitMod sets the initialization LLVM module for the project which should
// be the root module.
func (sb *StartBuilder) SetInitMod(rootMod *ir.Module) {
	sb.initMod = rootMod
}

// AddInitFunc adds a new initialization function to the start builder.
func (sb *StartBuilder) AddInitFunc(pkg *depm.ChaiPackage, initFunc *ir.Func) {
	sb.m.Lock()

	sb.initFuncs[pkg.ID] = initFunc.GlobalName

	sb.m.Unlock()
}

// -----------------------------------------------------------------------------

// BuildMainInitFunc builds the main initialization function.
func (sb *StartBuilder) BuildMainInitFunc() {
	// TODO
}
