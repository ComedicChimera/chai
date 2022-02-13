package mir

import (
	"chai/typing"
)

// Expr represents an expression in MIR.
type Expr interface {
	Type() typing.DataType
}

// -----------------------------------------------------------------------------

// IfTree represents an if expression.
type IfTree struct {
	CondBranches []*CondBranch
	ElseBranch   []BlockElem
}

// CondBranch represents a single condition branch of an if statement.
type CondBranch struct {
	Condition Expr
	Body      []BlockElem
}

// Loop represents a loop that repeats based on a condition in MIR. Both while
// and for loops compile to this kind of loop (in most cases).
type Loop struct {
	Condition Expr
	Body      []BlockElem

	// PostIter is an code that runs after every iteration.
	PostIter []BlockElem
}

// -----------------------------------------------------------------------------

// InstrExpr represents a MIR expression comprised of a single instruction.
type InstrExpr struct {
	InstrCode  int
	Operands   []Expr
	ResultType typing.DataType
}

// Enumeration of instruction codes.
const (
	InstrCall = iota // Call function

	InstrAdd // Add
	InstrSub // Subtract
	InstrMul // Multiply
	InstrDiv // Divide
	InstrMod // Modulo
	InstrNeg // Negate

	InstrEq   // Equal to
	InstrNEq  // Not equal to
	InstrLT   // Less than
	InstrGT   // Greater than
	InstrLTEq // Less than or equal to
	InstrGTEq // Greater than or equal to

	InstrNot // Logical NOT
	InstrAnd // Logical AND
	InstrOr  // Logical OR
)

func (ie *InstrExpr) Type() typing.DataType {
	return ie.ResultType
}

// TypeCast represents a type cast.
type TypeCast struct {
	SrcExpr  Expr
	DestType typing.DataType
}

func (tc *TypeCast) Type() typing.DataType {
	return tc.DestType
}

// GetField gets a field from a struct.
type GetField struct {
	Struct    Expr
	FieldName string
	FieldType typing.DataType
}

func (ge *GetField) Type() typing.DataType {
	return ge.FieldType
}

// StructInit represents a struct initialization.
type StructInit struct {
	StructType *typing.StructType
	FieldInits map[string]Expr
	SpreadInit Expr // `nil` => no spread init
}

func (si *StructInit) Type() typing.DataType {
	return si.StructType
}

// -----------------------------------------------------------------------------

// Ident represents an identifier in MIR.
type Ident struct {
	Name   string
	IdType typing.DataType
}

func (mid *Ident) Type() typing.DataType {
	return mid.IdType
}

// Literal represent a literal value in MIR.
type Literal struct {
	Value string
	VType typing.DataType
}

func (mv *Literal) Type() typing.DataType {
	return mv.VType
}
