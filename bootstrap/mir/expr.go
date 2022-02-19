package mir

import "chai/typing"

// Expr represents an expression in MIR.
type Expr interface {
	// Type returns the result type of the expression.
	Type() typing.DataType
}

// -----------------------------------------------------------------------------

// OperExpr is an expression comprised of an operation being applied to some
// number of values.  It contains only the operation and the sub-expressions.
type OperExpr struct {
	OpCode     int
	Operands   []Expr
	ResultType typing.DataType
}

func (oe *OperExpr) Type() typing.DataType {
	return oe.ResultType
}

// Enumeration of op codes
const (
	OCCall = iota

	OCIndirect
	OCDeref

	OCAdd
	OCSub
	OCMul
	OCDiv
	OCMod
	OCPow
	OCNeg

	OCEq
	OCNEq
	OCLt
	OCGt
	OCLtEq
	OCGtEq

	OCNot
	OCAnd
	OCOr

	OCCompl
	OCLShift
	OCRShift
	OCBWAnd
	OCBWOr
	OCBWXor
)

// CastExpr represents a type cast expression.
type CastExpr struct {
	Src      Expr
	DestType typing.DataType
}

func (ce *CastExpr) Type() typing.DataType {
	return ce.DestType
}

// FieldExpr represents an expression to access the field of a struct.
type FieldExpr struct {
	Struct    Expr
	FieldName string
	FieldType typing.DataType
}

func (fe *FieldExpr) Type() typing.DataType {
	return fe.FieldType
}

// -----------------------------------------------------------------------------

// LocalIdent represents a locally-defined identifier.
type LocalIdent struct {
	Name    string
	IdType  typing.DataType
	Mutable bool
}

func (lid *LocalIdent) Type() typing.DataType {
	return lid.IdType
}

// GlobalIdent represents a globally-defined identifier.
type GlobalIdent struct {
	ParentID uint64
	Name     string
	IdType   typing.DataType
	Mutable  bool
}

func (gid *GlobalIdent) Type() typing.DataType {
	return gid.IdType
}

// Literal represents a literal value.
type Literal struct {
	Value   string
	LitType typing.DataType
}

func (lit *Literal) Type() typing.DataType {
	return lit.LitType
}
