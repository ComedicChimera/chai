package mir

import (
	"chai/typing"
	"strings"
)

// MIRDef represents a MIR definition.  On its own, such definitions are
// interpreted as mere forward declarations: they do not provided
// implementations of the things they define.
type MIRDef interface {
	// Repr returns the textual representation of the definition.
	Repr() string

	// Public returns if the given definition is public.
	Public() bool
}

// -----------------------------------------------------------------------------

// MIRFuncDef is a MIR function definition.
type MIRFuncDef struct {
	Name       string
	Args       []MIRFuncArg
	ReturnType typing.DataType
	Pub        bool
}

// MIRFuncArg is the MIR data specific to a function argument.
type MIRFuncArg struct {
	Name     string
	ByRef    bool
	Constant bool
	Type     typing.DataType
}

func (mfd *MIRFuncDef) Repr() string {
	sb := strings.Builder{}
	sb.WriteString("func ")

	sb.WriteRune('$')
	sb.WriteString(mfd.Name)

	sb.WriteRune('(')

	for i, arg := range mfd.Args {
		if arg.ByRef {
			sb.WriteString("byref ")
		}

		if arg.Constant {
			sb.WriteString("const ")
		}

		sb.WriteRune('$')
		sb.WriteString(arg.Name)
		sb.WriteString(": ")
		sb.WriteString(arg.Type.Repr())

		if i < len(mfd.Args)-1 {
			sb.WriteRune(',')
		}
	}

	sb.WriteString(") ")
	sb.WriteString(mfd.ReturnType.Repr())

	return sb.String()
}

func (mfd *MIRFuncDef) Public() bool {
	return mfd.Pub
}
