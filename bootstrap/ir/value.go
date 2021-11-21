package ir

import "fmt"

// Value represents an operand that can be used in an instruction.
type Value interface {
	Repr() string

	Type() Type
}

// ValueBase is the base struct for all values.
type ValueBase struct {
	typ Type
}

func NewValueBase(typ Type) ValueBase {
	return ValueBase{typ: typ}
}

func (vb *ValueBase) Type() Type {
	return vb.typ
}

// -----------------------------------------------------------------------------

// ConstInt is an integer, boolean, or pointer constant.
type ConstInt struct {
	ValueBase
	Val int64
}

func (ci *ConstInt) Repr() string {
	return fmt.Sprintf("const %d %s", ci.Val, ci.typ.Repr())
}

// ConstFloat is a floating-point constant.
type ConstFloat struct {
	ValueBase
	Val float64
}

func (cf *ConstFloat) Repr() string {
	return fmt.Sprintf("const %f %s", cf.Val, cf.typ.Repr())
}

// -----------------------------------------------------------------------------

// GlobalIdentifier is the name of a global identifier.
type GlobalIdentifier struct {
	ValueBase
	Name string
}

func (id *GlobalIdentifier) Repr() string {
	return "@" + id.Name
}

// LocalIdentifier is the name of a local SSA value.
type LocalIdentifier struct {
	ValueBase
	ID int
}

func (id *LocalIdentifier) Repr() string {
	return fmt.Sprintf("$%d", id.ID)
}
