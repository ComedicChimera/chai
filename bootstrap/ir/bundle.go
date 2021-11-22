package ir

import "strings"

// Bundle represents a single unit of Chai IR.  Such a bundle can be compiled
// into a single object file or merged with other bundles and compiled together
// into one larger object file.  Bundles generally represent a single Chai
// package but can represent larger or smaller units (thus the logical
// distinction between bundle and package).
type Bundle struct {
	// SymTable is the global symbol table for this bundle.
	SymTable map[string]*IRSymbol

	// Functions is the list of functions defined in this package.
	Functions []*FuncDef
}

// IRSymbol represents a global symbol defined in the IR.  These symbols are
// much closer to symbols used by the linker and are used to determine the final
// symbol table of the bundle.
type IRSymbol struct {
	// Linkage indicates how this symbol is to be linked and what storage class,
	// value, etc. is placed with this symbol's definition in the final symbol
	// table of the resulting object file from this bundle. It is should be a
	// combination of one of the linkage flags below.
	Linkage int

	// Typ is type of this symbol as a value.
	Typ Type

	// Decl is this symbol's declaration within this bundle.
	Decl Decl
}

// Linkage flags
const (
	Private   = 0x1  // Symbol is private to its bundle (not public)
	Public    = 0x2  // Symbol is public to its bundle (public, visible externally)
	External  = 0x4  // Symbol is externally defined (extern)
	DllImport = 0x8  // Symbol is defined in a DLL depended on by this bundle (dllimport)
	DllExport = 0x16 // Symbol is exported as part of a DLL (dllexport)
)

// IsDefined returns whether or not this symbol is externally defined: ie. will
// not be defined in the object file.
func (isym *IRSymbol) IsDefined() bool {
	return (isym.Linkage&External) > 0 || (isym.Linkage&DllImport) > 0
}

// -----------------------------------------------------------------------------

// Decl represents a declaration in the IR.
type Decl interface {
	// Repr returns the string representation of the declaration.
	Repr() string

	// Section returns the section that where this symbol will be found in the
	// object file if it exists.  If it doesn't exist in the current object
	// file, then this function's return value is meaningless.  It must be one
	// of the enumerated sections below.
	Section() int
}

// Enumeration of sections
const (
	SectionText = iota
	SectionData
	SectionBSS
)

// -----------------------------------------------------------------------------

func NewBundle() *Bundle {
	return &Bundle{SymTable: make(map[string]*IRSymbol)}
}

// -----------------------------------------------------------------------------

func (b *Bundle) Repr() string {
	sb := strings.Builder{}

	// write the external symbols
	for _, sym := range b.SymTable {
		// write the symbol linkage
		if sym.Linkage&External > 0 {
			sb.WriteString("extern ")
		} else {
			// we don't put the defined symbols at the top of the source file
			continue
		}

		if sym.Linkage&DllImport > 0 {
			sb.WriteString("dllimport ")
		}

		sb.WriteString(sym.Decl.Repr())
		sb.WriteRune('\n')
	}

	sb.WriteRune('\n')

	// TODO: global symbols

	// function definitions
	for _, fd := range b.Functions {
		sb.WriteString(fd.Repr())
		sb.WriteString("\n\n")
	}

	return sb.String()
}
