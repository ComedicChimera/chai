package mir

import (
	"chaic/report"
	"chaic/types"
)

// Expr represents an expression in the MIR.
type Expr interface {
	// Type returns the type of the expression.
	Type() types.Type

	// LValue returns whether or not the expression is an L-value.
	LValue() bool

	// The span over which the expression occurs in source text.
	Span() *report.TextSpan
}

// The base type for all expressions.
type exprBase struct {
	span *report.TextSpan
}

// NewExprBase returns a new expr base with span span.
func NewExprBase(span *report.TextSpan) exprBase {
	return exprBase{span: span}
}

func (eb exprBase) LValue() bool {
	return false
}

func (eb exprBase) Span() *report.TextSpan {
	return eb.span
}

/* -------------------------------------------------------------------------- */

// Cast represents a type cast.
type Cast struct {
	exprBase

	// The source expression being cast.
	Src Expr

	// The destination type of the cast.
	DestType types.Type
}

func (ca *Cast) Type() types.Type {
	return ca.DestType
}

// FuncCall represents a function call.
type FuncCall struct {
	exprBase

	// The function being called.
	Func Expr

	// The arguments to the function.
	Args []Expr
}

func (call *FuncCall) Type() types.Type {
	return call.Func.Type().(*types.FuncType).ReturnType
}

// FieldAccess represents a struct or tuple field access.
type FieldAccess struct {
	exprBase

	// The struct or tuple whose field is being accessed.
	Struct Expr

	// The number of the field being accessed.
	FieldNumber int

	// The type of the field accessed.
	FieldType types.Type
}

func (fa *FieldAccess) LValue() bool {
	return fa.Struct.LValue()
}

func (fa *FieldAccess) Type() types.Type {
	return fa.FieldType
}

// BinaryOperApp is an intrinsic binary operator application: all non-intrinsic
// binary operator applications are converted into function calls.
type BinaryOperApp struct {
	exprBase

	// The name of the binary operator instruction generator.
	OpName string

	// The LHS and RHS expressions.
	LHS, RHS Expr

	// The result type of the operator application.
	ResultType types.Type
}

func (boa *BinaryOperApp) Type() types.Type {
	return boa.ResultType
}

// UnaryOperApp is an intrinsic unary operator application: all non-intrinsic
// unary operator applications are converted into function calls.
type UnaryOperApp struct {
	exprBase

	// The name of the unary operator instruction generator.
	OpName string

	// The LHS and RHS expressions.
	LHS, RHS Expr

	// The result type of the operator application.
	ResultType types.Type
}

func (uoa *UnaryOperApp) Type() types.Type {
	return uoa.ResultType
}

// AddressOf represents an L-value indirection.
type AddressOf struct {
	exprBase

	// The L-value whose address is being taken.
	Element Expr
}

func (ao *AddressOf) Type() types.Type {
	return &types.PointerType{ElemType: ao.Element.Type()}
}

// Deref represents a dereference.
type Deref struct {
	exprBase

	// The pointer being dereferenced.
	Ptr Expr
}

func (de *Deref) Type() types.Type {
	return de.Ptr.Type().(*types.PointerType).ElemType
}
