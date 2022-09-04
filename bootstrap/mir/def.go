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

	// Whether the function should be visible outside the bundle.
	Public bool

	// The bundle-relative path to the source file defining the function.
	SrcFilePath string

	// The text span of the function definition.
	Span *report.TextSpan

	// The list of attributes applied to the function (if any).
	Attrs []FuncAttribute

	// The body of the function.
	Body []Statement
}

// FuncAttribute represents a special attribute applied to the function.  These
// attributes are typically extracted from annotations.
type FuncAttribute struct {
	// Indicates the kind of attribute applied.  This must be one of the
	// enumerated attribute kinds.
	Kind int

	// The (optional) value of the attribute.
	Value string
}

// Enumeration of attribute kinds.
const (
	AttrKindPrototype = iota // Function has no body.
	AttrKindCallConv         // Function has a special calling convention.
	// TODO: add more as needed
)
