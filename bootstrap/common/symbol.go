package common

import (
	"chaic/report"
	"chaic/types"

	llvalue "github.com/llir/llvm/ir/value"
)

// Symbol represents a semantic symbol: a named value or definition.
type Symbol struct {
	// The name of the symbol.
	Name string

	// The ID of the parent package to this symbol.
	ParentID uint64

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

	// The LLVM value of the symbol.
	LLValue llvalue.Value
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
