package mir

import (
	"chai/typing"
	"fmt"
)

// Value is an interface representing a single discrete value such as a constant
// or identifier.  These are used as operands to instructions.
type Value interface {
	// Repr returns the string representation of the value.
	Repr() string

	// Pointer indicates whether or not this value is a pointer value such as a
	// variable that may need to be dereferenced (in LLVM).
	Pointer() bool
}

// Constant represents a literal constant value such as `2`.
type Constant struct {
	Value string
	Type  typing.DataType
}

func (c *Constant) Repr() string {
	return fmt.Sprintf("(const %s: %s)", c.Value, c.Type.Repr())
}

func (c *Constant) Pointer() bool {
	// constants are never pointer values
	return false
}

// Identifier represents a named identifier as a value.
type Identifier struct {
	Name    string
	Mutable bool

	// No data type since it will already be known and converted to LLVM in the
	// generator (via the global or local table).
}

func (id *Identifier) Repr() string {
	return "$" + id.Name
}

func (id *Identifier) Pointer() bool {
	return id.Mutable
}
