package depm

import "chai/report"

// Symbol represents a Chai symbol.
type Symbol struct {
	Name string

	// PkgID is the ID of the package this symbol is defined in.
	PkgID uint

	// DefPosition is the position of the identifier that defines the symbol.
	DefPosition *report.TextPosition

	// TODO: Type

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
	ValueDef = iota // Variables, Functions
	TypeDef
)

// Enumeration of mutabilities.
const (
	NeverMutated = iota // default mutability
	Mutable             // Symbol has been explicitly mutated at least once
	Immutable           // Symbol is defined as immutable (eg. a function or typedef)
)
