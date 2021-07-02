package sem

import (
	"chai/logging"
	"chai/syntax"
	"chai/typing"
)

// HIRExpr is the parent interface for all HIR (high-level intermediate
// representation) expressions
type HIRExpr interface {
	// Type returns the data type yielded by an expression.  This field can be
	// `nil` if the expression doesn't yield a type (eg. continue, return, etc)
	Type() typing.DataType

	// Category returns the value category of the expression.  It must be one of
	// the enumerated categories below.
	Category() int

	// Immutable indicates whether or not the expression is immutable
	Immutable() bool
}

// Enumeration of value categories
const (
	LValue = iota
	RValue
)

// ExprBase is the base struct for all expressions
type ExprBase struct {
	dt    typing.DataType
	cat   int
	immut bool
}

func NewExprBase(dt typing.DataType, cat int, immut bool) ExprBase {
	return ExprBase{
		dt:    dt,
		cat:   cat,
		immut: immut,
	}
}

func (eb *ExprBase) Type() typing.DataType {
	return eb.dt
}

func (eb *ExprBase) Category() int {
	return eb.cat
}

func (eb *ExprBase) Immutable() bool {
	return eb.immut
}

func (eb *ExprBase) SetType(dt typing.DataType) {
	eb.dt = dt
}

// HIRIncomplete represents an AST branch that hasn't been evaluated yet
type HIRIncomplete syntax.ASTBranch

func (hi *HIRIncomplete) Type() typing.DataType {
	return typing.PrimType(typing.PrimKindNothing)
}

func (hi *HIRIncomplete) Category() int {
	return RValue
}

func (hi *HIRIncomplete) Immutable() bool {
	return false
}

// -----------------------------------------------------------------------------

// HIRDoBlock represents a do block expression
type HIRDoBlock struct {
	ExprBase

	Statements []HIRExpr
}

// -----------------------------------------------------------------------------

type stmtBase struct {
}

func (*stmtBase) Type() typing.DataType {
	return typing.PrimType(typing.PrimKindNothing)
}

func (*stmtBase) Category() int {
	return RValue
}

func (*stmtBase) Immutable() bool {
	return false
}

// HIRVarDecl is a variable declaration expression
type HIRVarDecl struct {
	stmtBase
	Variables    []*Symbol
	Initializers map[string]HIRExpr
}

// HIRAssignment represents some form of assignment or mutation to a value
type HIRAssignment struct {
	stmtBase

	// If Rhs contains only one value and Lhs contains multiple, then this is an
	// unpacking assignment.  If this is a unary assignment, Rhs is empty
	Lhs, Rhs []HIRExpr

	// AssignKind is one of the enumerated assignment kinds below
	AssignKind int

	// Oper is the operator being used in any compound or unary assignments (eg.
	// +=, *=, ++, etc.).  This field can be `nil` if no operator is used.
	Oper *Operator
}

const (
	AKEq       = iota // `=`
	AKBind            // `<-`
	AKCompound        // [oper]=
	AKUnary           // `++` or `--`
)

// -----------------------------------------------------------------------------

// HIRApply represents a function application
type HIRApply struct {
	ExprBase

	Func HIRExpr
	Args map[string]HIRExpr

	// VarArgs contains all of the non-spread variadic arguments to the function
	VarArgs []HIRExpr

	// SpreadArg is the argument used for variadic spread initialization
	SpreadArg HIRExpr
}

// HIROperApply represents an operator application
type HIROperApply struct {
	ExprBase

	Oper     *Operator
	Operands []HIRExpr

	// OperFunc is just the signature of the operator overload; it does not
	// contain flags indicating intrinsics or boxed status -- those have to be
	// loaded from the overload after
	OperFunc *typing.FuncType
}

// HIRCast denotes an explicit type cast.  The destination type is stored in the
// yield type of the expression
type HIRCast struct {
	ExprBase

	Root HIRExpr
}

// -----------------------------------------------------------------------------

// HIRIdentifier represents an identifier
type HIRIdentifier struct {
	Sym *Symbol

	// IdPos is the position of the identifier token
	IdPos *logging.TextPosition
}

func (hi *HIRIdentifier) Type() typing.DataType {
	return hi.Sym.Type
}

func (hi *HIRIdentifier) Category() int {
	// all identifiers are LValues
	return LValue
}

func (hi *HIRIdentifier) Immutable() bool {
	return hi.Sym.Immutable
}

// HIRLiteral represents a literal
type HIRLiteral struct {
	ExprBase

	// Value can be `null` if this is a null literal; otherwise, it is just the
	// value of the literal (eg. `12`, `0b101`, etc.)
	Value string
}
