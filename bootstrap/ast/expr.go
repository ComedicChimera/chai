package ast

import (
	"chaic/common"
	"chaic/report"
	"chaic/types"
)

// ASTExpr represents an expression node in the AST.
type ASTExpr interface {
	ASTNode

	// Type returns the yielded type of the AST expression.
	Type() types.Type

	// Category() returns the value category of the yielded value of the AST
	// expression. This must be one of the enumerated value categories.
	Category() int

	// Constant() returns whether the given value is constant.
	Constant() bool
}

// Enumeration of value categories.
const (
	LVALUE = iota
	RVALUE
)

// The base type for all AST expressions.
type ExprBase struct {
	ASTBase

	// The yielded type of the AST node.
	NodeType types.Type
}

// NewExprBase creates a new expression base with no type.
func NewExprBase(span *report.TextSpan) ExprBase {
	return ExprBase{ASTBase: ASTBase{span: span}}
}

// NewTypedExprBase creates a new expression with a type.
func NewTypedExprBase(span *report.TextSpan, typ types.Type) ExprBase {
	return ExprBase{ASTBase: ASTBase{span: span}, NodeType: typ}
}

func (eb *ExprBase) Type() types.Type {
	return eb.NodeType
}

func (eb *ExprBase) Category() int {
	return RVALUE
}

func (eb *ExprBase) Constant() bool {
	return false
}

/* -------------------------------------------------------------------------- */

// AppliedOperator represents a particular application of an operator.
type AppliedOperator struct {
	// The token kind of the operator.
	OpKind int

	// The string name of the operator.
	OpName string

	// The span over which the operator occurs in source text.
	Span *report.TextSpan

	// The operator method used in this application.
	OpMethod *common.OperatorMethod
}

// -----------------------------------------------------------------------------

// TypeCast represents an AST type cast.
type TypeCast struct {
	ExprBase

	// The source expression being cast.
	SrcExpr ASTExpr
}

// -----------------------------------------------------------------------------

// BinaryOpApp represents an AST binary operator application.
type BinaryOpApp struct {
	ExprBase

	// The operator application.
	Op *AppliedOperator

	// The LHS operand.
	LHS ASTExpr

	// The RHS operand.
	RHS ASTExpr
}

// UnaryOpApp represents an AST unary operator application.
type UnaryOpApp struct {
	ExprBase

	// The operator application.
	Op *AppliedOperator

	// The operand.
	Operand ASTExpr
}

// -----------------------------------------------------------------------------

// Indirect represents an AST indirection expression.
type Indirect struct {
	ExprBase

	// The element being indirected.
	Elem ASTExpr

	// Whether a const reference is produced.
	Const bool
}

// Deref represents an AST dereference expression.
type Deref struct {
	ExprBase

	// The pointer being dereferenced.
	Ptr ASTExpr
}

func (d *Deref) Category() int {
	// Dereferences always result in lvalues.
	return LVALUE
}

func (d *Deref) Constant() bool {
	// Dereferences to constant pointers may yield constants.
	return types.InnerType(d.Ptr.Type()).(*types.PointerType).Const
}

// -----------------------------------------------------------------------------

// FuncCall represents an AST function call.
type FuncCall struct {
	ExprBase

	// The function being called.
	Func ASTExpr

	// The arguments to the function.
	Args []ASTExpr
}

// PropertyAccess represents a property access (ie. a `.` expr).
type PropertyAccess struct {
	ExprBase

	// The value whose property is being accessed.
	Root ASTExpr

	// The name of the property being accessed.
	PropName string

	// The span over which the property name occurs.
	PropSpan *report.TextSpan

	// The kind of property that was accessed.
	PropKind types.PropertyKind
}

// Index represents an index operation.
type Index struct {
	ExprBase

	// The index operator.
	Op *AppliedOperator

	// The value which is being indexed.
	Root ASTExpr

	// The subscript: index being accessed.
	Subscript ASTExpr
}

// Slice represents a slice operation.
type Slice struct {
	ExprBase

	// The slice operator.
	Op *AppliedOperator

	// The value being sliced.
	Root ASTExpr

	// The start value of the slice if it exists.
	Start ASTExpr

	// The end value of the slice if it exists.
	End ASTExpr
}

// -----------------------------------------------------------------------------

// StructLiteral represents a structure literal (initializer).
type StructLiteral struct {
	ExprBase

	// The field initializers.
	FieldInits map[string]StructLiteralFieldInit

	// The spread initializer if it exists.
	SpreadInit ASTExpr
}

// StructLiteralFieldInit represents a field initializer in a struct literal.
type StructLiteralFieldInit struct {
	// The name of the field being initialized.
	Name string

	// The span over which the name of the field being initialized occurs.
	NameSpan *report.TextSpan

	// The initializer expression.
	InitExpr ASTExpr
}

// ArrayLiteral represents an array literal.
type ArrayLiteral struct {
	ExprBase

	// The array elements.
	Elements []ASTExpr
}

// -----------------------------------------------------------------------------

// Identifier represents an AST identifier.
type Identifier struct {
	ASTBase

	// The name of the identifier.
	Name string

	// The symbol which this identifier refers to.
	Sym *common.Symbol
}

func (id *Identifier) Type() types.Type {
	return id.Sym.Type
}

func (id *Identifier) SetType(typ types.Type) {
	report.ReportICE("SetType for identifiers is not implemented")
}

func (id *Identifier) Category() int {
	return LVALUE
}

func (id *Identifier) Constant() bool {
	return id.Sym.Constant
}

// Literal represents a literal value on the AST.
type Literal struct {
	ExprBase

	// The token kind of the literal.
	Kind int

	// The string value of the literal.
	Text string

	// The interpreted Go value of the literal.
	Value interface{}
}

// Null represents a null value on the AST.
type Null struct {
	ExprBase
}
