package sem

import (
	"chai/logging"
	"chai/syntax"
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

	// Modifiers is a bit field that is used to store the modifiers of the
	// symbol.  The various bit field values are enumerated below.
	Modifiers int

	// Mutability indicates whether or not this symbol can and has been mutated
	Mutability int

	// Position is the text position where this symbol is defined
	Position *logging.TextPosition
}

// HasModifier checks if the symbol has a given modifier
func (sym *Symbol) HasModifier(modifier int) bool {
	return sym.Modifiers&modifier != 0
}

// Enumeration of symbol definition kinds
const (
	DefKindTypeDef  = iota // Type and Class definitions
	DefKindConsDef         // Constraint Definitions
	DefKindFuncDef         // Function and operator definitions
	DefKindValueDef        // Variables and other identifiers
)

// Enumeration of symbol modifiers.  These are used as bitfield values
const (
	ModPublic   = 1
	ModVolatile = 2
)

// Enumeration of mutabilities
const (
	Immutable    = iota // Cannot be mutated
	NeverMutated        // Can be mutated, never has been
	Mutable             // Can and has been mutated
)

// Annotation represents an annotation (name and value)
type Annotation struct {
	Name    string
	NamePos *logging.TextPosition

	// Values stores leaves so we can keep their position
	Values []*syntax.ASTLeaf
}
