// Package cmd is the top-level "driver" package for the Chai compiler: it
// contains all the functionality for parsing command-line arguments, managing
// compiler state, and runnign all the various phases of the compiler.
package cmd

import (
	"chaic/depm"
)

// Compiler represents the overall state and configuration of compilation.
type Compiler struct {
	// The CHAI_PATH of this compiler instance.
	chaiPath string

	// The path to the root directory or file of compilation.
	rootPath string

	// The path to write output to.  This is a path to a directory if multiple
	// outputs are being generated.
	outputPath string

	// The output mode: the kind of output the compiler should produce.  This
	// must be one of the enumerated output modes.
	outputMode int

	// Whether the compiler should emit debug information.
	debug bool

	// depGraph is the compiler's package dependency graph.
	depGraph map[uint64]*depm.ChaiPackage
}

// Enumeration of compilation output modes.
const (
	OutModeExecutable = iota // Output an executable (default).
	OutModeStaticLib         // Output a static library.
	OutModeDynamicLib        // Output a dynamic library.
	OutModeObj               // Output compiled packages as object files.
	OutModeASM               // Output compiled packages as assembly files.
	OutModeLLVM              // Output compiled packages as LLVM IR text files.
)

// RunCompiler is the main entry point for the Chai compiler.  This should be
// called directly from main.
func RunCompiler() int {
	// Create a new compiler from the given command-line arguments.
	c := NewCompilerFromArgs()

	// TODO: handle scripts

	// Initialize the root package which will in turn initialize all
	// sub-packages.
	if _, ok := c.InitPackage(c.rootPath); !ok {
		return 1
	}

	// Perform symbol resolution, import resolution, and infinite type checking.
	if !c.ResolveSymbols() {
		return 1
	}

	// Perform semantic analysis.
	if !c.WalkPackages() {
		return 1
	}

	// Generate compilation output.
	c.CodeGen()
	return 0
}
