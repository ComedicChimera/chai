package mir

import (
	"chaic/report"
	"chaic/types"
)

// The base struct for all values.
type ValueBase struct {
	ExprBase

	// The type of the value.
	typ types.Type
}

// NewValueBase creates a new value base with the span span and type typ.
func NewValueBase(span *report.TextSpan, typ types.Type) ValueBase {
	return ValueBase{
		ExprBase: ExprBase{span: span},
		typ:      typ,
	}
}

func (vb ValueBase) Type() types.Type {
	return vb.typ
}

/* -------------------------------------------------------------------------- */

// Identifier represents an reference to a symbol.
type Identifier struct {
	ExprBase

	// The MIR symbol associated with the identifier.
	Symbol *MSymbol
}

func (ident *Identifier) Type() types.Type {
	return ident.Symbol.Type
}

func (ident *Identifier) LValue() bool {
	return true
}

// ConstInt represents an integer constant.
type ConstInt struct {
	ValueBase

	// The integer value of the integer constant.
	IntValue int64
}

// ConstReal represents a real (floating-point) constant.
type ConstReal struct {
	ValueBase

	// The floating-point value of the integer constant.
	FloatValue float64
}

// ConstUnit represents the unit value.
type ConstUnit struct {
	ValueBase
}

// ConstNullPtr represents the null pointer.
type ConstNullPtr struct {
	ValueBase
}
