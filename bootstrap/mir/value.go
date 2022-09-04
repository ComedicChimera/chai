package mir

import (
	"chaic/report"
	"chaic/types"

	llvalue "github.com/llir/llvm/ir/value"
)

// The base struct for all values.
type valueBase struct {
	exprBase

	// The type of the value.
	typ types.Type
}

// NewValueBase creates a new value base with the span span and type typ.
func NewValueBase(span *report.TextSpan, typ types.Type) valueBase {
	return valueBase{
		exprBase: exprBase{span: span},
		typ:      typ,
	}
}

/* -------------------------------------------------------------------------- */

// Identifier represents an reference to a symbol.
type Identifier struct {
	valueBase

	// The name of the identifier.
	Name string

	// Whether or not the identifier represents an implicit pointer to a value:
	// eg. if the identifier refers to a mutable variable, then it is actually a
	// pointer to the value and will need to be loaded.
	IsImplicitPointer bool

	// The LLVM value of the identifier.
	LLValue llvalue.Value
}

func (ident *Identifier) LValue() bool {
	return true
}

// ConstInt represents an integer constant.
type ConstInt struct {
	valueBase

	// The integer value of the integer constant.
	IntValue int64
}

// ConstReal represents a real (floating-point) constant.
type ConstReal struct {
	valueBase

	// The floating-point value of the integer constant.
	FloatValue float64
}

// ConstUnit represents the unit value.
type ConstUnit struct {
	valueBase
}

// ConstNullPtr represents the null pointer.
type ConstNullPtr struct {
	valueBase
}
