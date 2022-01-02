package generate

import (
	"chai/depm"

	"github.com/llir/llvm/ir"
	"github.com/llir/llvm/ir/enum"
	"github.com/llir/llvm/ir/types"
)

// StartBuilder is responsible for constructing the main initialization function
// (ie. the initialization function that calls all other initialization
// functions).
type StartBuilder struct {
	// initFuncNamesCh is a channel of the names of package initialization
	// functions as they are being written.  These names are consumed at the end
	// to generate the main initialization function.
	initFuncNamesCh chan string

	// initMod is the LLVM module that contains the source for the main
	// initialization function.
	initMod *ir.Module
}

// NewStartBuilder returns a new StartBuilder for the project.
func NewStartBuilder(depGraph map[uint]*depm.ChaiModule, rootPkgID uint) *StartBuilder {
	initFuncCount := 0
	for _, mod := range depGraph {
		for range mod.Packages() {
			initFuncCount++
		}
	}

	return &StartBuilder{
		initMod:         ir.NewModule(),
		initFuncNamesCh: make(chan string, initFuncCount),
	}
}

// AddInitFunc adds a new initialization function to the start builder.
func (sb *StartBuilder) AddInitFunc(initFuncName string) {
	sb.initFuncNamesCh <- initFuncName
}

// GetInitMod gets the initialization LLVM module for the project.
func (sb *StartBuilder) GetInitMod() *ir.Module {
	return sb.initMod
}

// -----------------------------------------------------------------------------

// BuildMainInitFunc builds the main initialization function.
func (sb *StartBuilder) BuildMainInitFunc() {
	// close the channel => all the names should be in by now
	close(sb.initFuncNamesCh)

	// define a new function called `__chai_init` which is the main
	// initialization function.
	mainInitFunc := sb.initMod.NewFunc("__chai_init", types.Void)
	mainInitFunc.Linkage = enum.LinkageExternal
	mainInitFunc.FuncAttrs = append(mainInitFunc.FuncAttrs, enum.FuncAttrNoUnwind)

	// call the initialization functions.  These are called in an essentially
	// random order since it is impossible to know which packages need their
	// initialization performed first (especially if we have a cyclic
	// dependency).  The compiler will produce an appropriate warning if this
	// "randomness" will cause an issue
	entryBlock := mainInitFunc.NewBlock("entry")
	for {
		// drain the names channel to get all the names to call
		initFuncName, ok := <-sb.initFuncNamesCh
		if !ok {
			break
		}

		// generate the external reference to the initialization function
		initFunc := sb.initMod.NewFunc(initFuncName, types.Void)
		initFunc.Linkage = enum.LinkageExternal
		initFunc.FuncAttrs = append(initFunc.FuncAttrs, enum.FuncAttrNoUnwind)

		// call it from the main init function
		entryBlock.NewCall(initFunc)
	}

	// terminate the main init function
	entryBlock.NewRet(nil)
}
