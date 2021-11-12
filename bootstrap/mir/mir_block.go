package mir

import (
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
	RHS  Instruction
}

func (b *Binding) Repr(preindent string) string {
	return fmt.Sprintf("%s$%s := %s", preindent, b.Name, b.RHS.Repr())
}

// InstrStmt is an instruction used as a standalone statement with no binding.
type InstrStmt struct {
	Instr Instruction
}

func (is *InstrStmt) Repr(preindent string) string {
	return preindent + is.Instr.Repr()
}

// Return is a return statement.
type Return struct {
	// RetVal may be `nil` if no value is returned.
	RetVal Value
}

func (r *Return) Repr(preindent string) string {
	if r.RetVal == nil {
		return preindent + "ret;"
	}

	return fmt.Sprintf("%sret %s;", preindent, r.RetVal.Repr())
}
