package mir

import "chai/typing"

// Bundle is a package represented in MIR.  It contains all the content of the
// individual files of the package but in a far more condensed format.
type Bundle struct {
	// ID the ID of the bundle.  This is the same as the ID of the package.
	ID uint64

	// Imports is a list of all the bundles imported by this bundle.
	Imports []*Import

	// Funcs is a list of function definitions defined in the bundle.
	Funcs []*FuncDef

	// TypeDefs is a list of type definitions defined in the bundle.
	TypeDefs []*TypeDef
}

// NewBundle creates a new MIR bundle.
func NewBundle(id uint64) *Bundle {
	return &Bundle{ID: id}
}

// -----------------------------------------------------------------------------

// Import represents an import of a specific MIR bundle.  This contains all the
// symbols imported from that bundle.
type Import struct {
	BundleID uint64
	Symbols  map[string]*SymbolImport
}

// SymbolImport represents a symbol imported by a MIR bundle.
type SymbolImport struct {
	Name    string
	Type    typing.DataType
	SymKind int
}

const (
	ISKindFunc = iota // Function
	ISKindType        // Type Definition
	ISKindVar         // Global Variable
)
