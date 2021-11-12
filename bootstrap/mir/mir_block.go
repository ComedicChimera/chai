package mir

import (
	"chai/typing"
	"fmt"
)

// FuncImpl represents the implementation of a function: it definition and its
// body together.
type FuncImpl struct {
	Def *FuncDef

	Body []Stmt
}

// -----------------------------------------------------------------------------

// Stmt is the parent interface for all sequentially defined statements such as
// instructions, if statements, loops, etc.  It is the element type of a block.
type Stmt interface {
	// Repr returns the indented string representation of the stmt.
	Repr(preindent string) string
}

// Binding is an immutable assignment of the result of an instruction to a
// temporary value.  Bindings follow an SSA-like paradigm.
type Binding struct {
	Name string
	RHS  *Instruction
}

func (b *Binding) Repr(preindent string) string {
	return fmt.Sprintf("%s$%s := %s", preindent, b.Name, b.RHS.Repr(""))
}

// Cast is a type cast of a value to a different type.  Type casts act both as
// bindings and instructions.
type Cast struct {
	// DestName is the name to bind the result of the cast to.
	DestName string

	// Value is the value to type cast.
	Value Value

	// DestType is the type being casted to.
	DestType typing.DataType
}

func (c *Cast) Repr(preindent string) string {
	return fmt.Sprintf("%s$%s := cast %s to %s;", preindent, c.DestName, c.Value.Repr(), c.DestType.Repr())
}
