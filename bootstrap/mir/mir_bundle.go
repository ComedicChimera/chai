package mir

// See `github.com/ComedicChimera/chai-lang.dev/blob/main/content/chai_mir.md`
// for more information about the structure of Chai's MIR.

// MIRBundle is a single unit representing the lowered contents of a single
// ChaiPackage.  MIR bundles are converted directly into LLVM modules. The
// entirety of the package is stored and concisely represented in Chai's MIR.
// MIR bundles are considered distinct translation units -- all symbols that are
// not immediately defined in the package are considered external.
type MIRBundle struct {
	// Externals is a list of the external function definitions used by this
	// package.  These symbols may be defined in other MIR bundles (to be
	// resolved later by linker) or may be found in some external library.
	Externals []Def

	// Forwards is a list of the forward declarations of the package.
	Forwards []Def

	// Functions is a list a the function implementations of the package.
	Functions []*FuncImpl
}
