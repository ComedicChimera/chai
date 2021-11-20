package ir

import (
	"fmt"
	"strconv"
	"strings"
)

// Block represents a series of IR statements used to define a function.
type Block struct {
	Stmts []Statement

	// Labels is a list of defined labels within the block that can be jumped
	// to.  The ID of the label is its index within this list.  The value in
	// this list is the statement this label jumps to.
	Labels []int
}

func (b *Block) Repr() string {
	sb := strings.Builder{}

	nextLabelID := 0
	for i, stmt := range b.Stmts {
		if i == b.Labels[nextLabelID] {
			sb.WriteRune('@')
			sb.WriteString(strconv.Itoa(nextLabelID))
			sb.WriteString(":\n")

			nextLabelID++
		}

		sb.WriteString("  ")
		sb.WriteString(stmt.Repr())
		sb.WriteRune('\n')
	}

	sb.WriteRune('\n')
	return sb.String()
}

// -----------------------------------------------------------------------------

// Statement represents a single IR statement used in a block.
type Statement interface {
	// Repr returns a representative string from the statement.
	Repr() string

	// Line returns the source line number this statement corresponds to.
	Line() int
}

// StmtBase is the base structure for all IR statements.
type StmtBase struct {
	line int
}

func NewStmtBase(line int) StmtBase {
	return StmtBase{line: line}
}

func (sb *StmtBase) Line() int {
	return sb.line
}

// -----------------------------------------------------------------------------

// Binding represents the binding of an instruction to an SSA value.
type Binding struct {
	StmtBase
	ValueID int
	Instr   *Instruction
}

func (b *Binding) Repr() string {
	return fmt.Sprintf("$%d = %s", b.ValueID, b.Instr.Repr())
}

// -----------------------------------------------------------------------------

// Instruction represents a single operation within the IR.
type Instruction struct {
	// OpCode must be one of the enumerated instruction op codes.
	OpCode int

	// TypeSpec is the type specifier for the instruction: this is used to
	// determine what code to generate for the instruction based on the type it
	// operates on. For example, the `add` instruction's type specifier
	// indicates what type of numbers it operates on.
	TypeSpec Type

	// TODO: Values
}

func (instr *Instruction) Repr() string {
	// TODO
	return ""
}
