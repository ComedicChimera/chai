package mir

import (
	"chaic/report"
	"chaic/types"
)

// Function represents a MIR function which can correspond to a function, a
// method, a property, or a operator definition.
type Function struct {
	// The name of the function.
	Name string

	// The type signature of the function.
	Signature *types.FuncType

	// The parameter variables of the function.
	ParamVars []*Identifier

	// Whether the function should be visible outside the bundle.
	Public bool

	// The absolute path to the source file defining the function.
	SrcFileAbsPath string

	// The text span of the function definition.
	Span *report.TextSpan

	// The set of attributes applied to the function (if any).
	Attrs map[FuncAttrKind]string

	// The body of the function.
	Body []Statement

	// The definitive identifier of the function.
	Ident *Identifier
}

// FuncAttrKind is a kind of a function attribute.
type FuncAttrKind int

// Enumeration of attribute kinds.
const (
	AttrKindPrototype FuncAttrKind = iota // Function has no body.
	AttrKindCallConv                      // Function has a special calling convention.
	// TODO: add more as needed
)

/* -------------------------------------------------------------------------- */

// Struct represents a MIR struct definition.
type Struct struct {
	// The type of the struct.
	Type *types.StructType

	// The absolute path to the source file defining the struct.
	SrcFileAbsPath string

	// The span the struct is defined over.
	Span *report.TextSpan
}
