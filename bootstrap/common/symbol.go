package common

import (
	"chaic/llvm"
	"chaic/report"
	"chaic/types"
)

// Symbol represents a semantic symbol: a named value or definition.
type Symbol struct {
	// The name of the symbol.
	Name string

	// The ID of the parent package to this symbol.
	ParentID int

	// The numbering identifying the file which defines this symbol.
	FileNumber int

	// Where the symbol was defined.
	DefSpan *report.TextSpan

	// The type of the value stored in the symbol.
	Type types.Type

	// The symbol's kind: what kind of things does this symbol represent. This
	// must be one the enumerated definition kinds.
	DefKind int

	// Whether or not the symbol is constant.
	Constant bool

	// Whether or not the symbol was actually used.
	Used bool

	// The symbol's LLVM value.
	LLValue llvm.Value

	// The symbol's LLVM type.  This may not be set on all symbols.
	LLType llvm.Type
}

// Enumeration of different symbol kinds.
const (
	DefKindValue = iota
	DefKindFunc
	DefKindType
)

// Enumeration of different mutabilities.
const (
	Immutable = iota
	Mutable
)
