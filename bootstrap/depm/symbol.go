package depm

import (
	"chai/report"
	"chai/typing"
)

// Symbol represents a Chai symbol.
type Symbol struct {
	Name string

	// Pkg is the  package this symbol is defined in.
	Pkg *ChaiPackage

	// DefPosition is the position of the identifier that defines the symbol.
	DefPosition *report.TextPosition

	// Type is the symbol's defined type.
	Type typing.DataType

	// DefKind indicates what type of symbol this is: ie. does it correspond to
	// a value such as a function or a type.  Must be one of the enumerated def
	// kinds.
	DefKind int

	// Mutability indicates whether the symbol can be mutated, and if it can be,
	// has it been for purposes of implicit constancy.  Must be one of the
	// enumerated mutabilities.
	Mutability int

	// Public indicates whether or not this symbol is externally visible.
	Public bool
}

// Enumeration of definition kinds.
const (
	DKValueDef = iota // Variables, Functions
	DKTypeDef
	DKUnknown
)

// ReprDefKind generates a representative string for a definition kind.
func ReprDefKind(dkind int) string {
	switch dkind {
	case DKValueDef:
		return "value"
	case DKTypeDef:
		return "type"
	default:
		// should never occur, but...
		return "unknown"
	}
}

// Enumeration of mutabilities.
const (
	NeverMutated = iota // default mutability
	Mutable             // Symbol has been explicitly mutated at least once
	Immutable           // Symbol is defined as immutable (eg. a function or typedef)
)
