package mir

import "chaic/types"

// Bundle is the MIR representation of a package.
type Bundle struct {
	// The unique ID of the MIR bundle (same as its source package ID).
	ID uint64

	// The absolute path to the package which created this bundle.
	PkgAbsPath string

	// The list of all the functions in the bundle.
	Functions []*Function

	// The list of all the structs in the bundle.
	Structs []*types.StructType
}
