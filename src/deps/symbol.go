package deps

import (
	"chai/logging"
	"chai/typing"
)

// Symbol represents a named symbol (globally or locally)
type Symbol struct {
	// Name is the name of the symbol (as it is referenced in source code)
	Name string

	// SrcPackage is the package this symbol is defined in
	SrcPackage *ChaiPackage

	// Type stores the data type of this symbol
	Type typing.DataType

	// DefKind is the kind of definition that produced this definition. This
	// must be one of the enumerated definition kinds below
	DefKind int

	// Public indicates whether or not this symbol is public (exported)
	Public bool

	// Mutability indicates whether or not this symbol can and has been mutated
	Mutability int

	// Position is the text position where this symbol is defined
	Position *logging.TextPosition
}

// Enumeration of symbol definition kinds
const (
	DefKindTypeDef    = iota // Type, Class, and Constraint definitions
	DefKindFuncDef           // Function and operator definitions
	DefKindNamedValue        // Variables and other identifiers
)

// Enumeration of mutabilities
const (
	Immutable    = iota // Cannot be mutated
	NeverMutated        // Can be mutated, never has been
	Mutable             // Can and has been mutated
)
