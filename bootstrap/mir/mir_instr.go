package mir

import (
	"chai/typing"
	"fmt"
	"strings"
)

// Instruction represents a standalone statement that performs an operation such
// as a call and an addition.  The results of instructions can be stored into a
// bind or used in an assignment.
type Instruction interface {
	// repr returns the unindent string representation of an instruction.
	Repr() string
}

// -----------------------------------------------------------------------------

// Call is an instruction to call a function.
type Call struct {
	FuncName string
	Args     []Value
}

func (c *Call) Repr() string {
	sb := strings.Builder{}
	sb.WriteString("call $")
	sb.WriteString(c.FuncName)

	sb.WriteRune('(')

	for i, arg := range c.Args {
		sb.WriteString(arg.Repr())

		if i < len(c.Args)-1 {
			sb.WriteString(", ")
		}
	}

	sb.WriteString(");")
	return sb.String()
}

// Cast is an instruction to cast of a value to a different type.
type Cast struct {
	// Value is the value to type cast.
	Value Value

	// DestType is the type being casted to.
	DestType typing.DataType
}

func (c *Cast) Repr() string {
	return fmt.Sprintf("cast %s to %s;", c.Value.Repr(), c.DestType.Repr())
}
