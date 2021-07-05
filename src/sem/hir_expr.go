package sem

import (
	"chai/logging"
	"chai/syntax"
	"chai/typing"
)

// HIRExpr is the parent interface for all HIR (high-level intermediate
// representation) expressions
type HIRExpr interface {
	// Type returns the data type yielded by an expression
	Type() typing.DataType

	// Category returns the value category of the expression.  It must be one of
	// the enumerated categories below.
	Category() int

	// Immutable indicates whether or not the expression is immutable
	Immutable() bool

	// Control returns the effect this expression has on control flow.  This
	// method will only return a control flow effect if that effect is
	// unconditional: eg. an if statement that only returns on some branches
	// will have a control flow kind of CFNone not CFReturn
	Control() int
}

// Enumeration of value categories
const (
	LValue = iota
	RValue
)

// Enumeration of control flow types
const (
	CFNone   = iota // No change to control flow
	CFReturn        // Return from function
	CFYield         // Yield from a block
	CFLoop          // Change the control flow of a loop
	CFMatch         // Changes the control flow of a match
	CFPanic         // Panic and exit block
)

// ExprBase is the base struct for all expressions
type ExprBase struct {
	dt      typing.DataType
	cat     int
	immut   bool
	control int
}

func NewExprBase(dt typing.DataType, cat int, immut bool) ExprBase {
	return ExprBase{
		dt:      dt,
		cat:     cat,
		immut:   immut,
		control: CFNone,
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

func (eb *ExprBase) Control() int {
	return eb.control
}

func (eb *ExprBase) SetType(dt typing.DataType) {
	eb.dt = dt
}

func (eb *ExprBase) SetControl(control int) {
	eb.control = control
}

// -----------------------------------------------------------------------------

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

func (hi *HIRIncomplete) Control() int {
	return CFNone
}

// -----------------------------------------------------------------------------

// HIRWhileLoop represents a while loop/expressions
type HIRWhileLoop struct {
	ExprBase

	// HeaderDecl is a variable declaration that occurs in the loop header. This
	// field can be `nil` if there was no header variable declaration
	HeaderDecl *HIRVarDecl

	// HeaderCond is the condition expression for the while loop
	HeaderCond HIRExpr

	// HeaderUpdate is the update statement to be run at the end of each loop.
	// This field can be `nil` if there was no update statement
	HeaderUpdate HIRExpr

	// LoopBody is the body of the while loop
	LoopBody HIRExpr

	// NoBreakClause is the body of the `nobreak` following a loop.  This field
	// can be `nil` if there was no `nobreak`
	NoBreakClause HIRExpr
}

// HIRDoBlock represents a do block expression
type HIRDoBlock struct {
	ExprBase

	Statements []HIRExpr
}

// -----------------------------------------------------------------------------

type stmtBase struct {
	// control default to zero (CFNone)
	control int
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

func (sb *stmtBase) Control() int {
	return sb.control
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

// HIRControlStmt is a control flow statement (break, continue, etc)
type HIRControlStmt struct {
	stmtBase

	// Kind indicates the kind of control statement: must be one of the control
	// kinds enumerated below
	Kind int
}

// NewControlStmt returns a new control flow statement based on the control flow
// statement kind passed in
func NewControlStmt(kind int) *HIRControlStmt {
	switch kind {
	case CSUnimplemented:
		return &HIRControlStmt{
			Kind:     kind,
			stmtBase: stmtBase{control: CFPanic},
		}
	case CSBreak, CSContinue:
		return &HIRControlStmt{
			Kind:     kind,
			stmtBase: stmtBase{control: CFLoop},
		}
	default /* fallthrough variants */ :
		return &HIRControlStmt{
			Kind:     kind,
			stmtBase: stmtBase{control: CFMatch},
		}
	}
}

// Enumeration of control statement kinds
const (
	CSBreak         = iota // `break`
	CSContinue             // `continue`
	CSFallthrough          // `fallthrough`
	CSFallMatch            // `fallthrough to match`
	CSUnimplemented        // `...`
)

// HIRReturnStmt is a statement that returns a value from a function
type HIRReturnStmt struct {
	stmtBase

	Value HIRExpr
}

func (hrs *HIRReturnStmt) Control() int {
	return CFReturn
}

func (hrs *HIRReturnStmt) Type() typing.DataType {
	return hrs.Value.Type()
}

// HIRYieldStmt is a statement that yields a value from a block
type HIRYieldStmt struct {
	stmtBase

	Value HIRExpr
}

func (hys *HIRYieldStmt) Control() int {
	return CFYield
}

func (hys *HIRYieldStmt) Type() typing.DataType {
	return hys.Value.Type()
}

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

func (hi *HIRIdentifier) Control() int {
	return CFNone
}

// HIRLiteral represents a literal
type HIRLiteral struct {
	ExprBase

	// Value can be `null` if this is a null literal; otherwise, it is just the
	// value of the literal (eg. `12`, `0b101`, etc.)
	Value string
}
