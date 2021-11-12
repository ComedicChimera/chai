package mir

import (
	"chai/ast"
	"chai/typing"
	"strings"
)

// Def represents a MIR definition.  On its own, such definitions are
// interpreted as mere forward declarations: they do not provided
// implementations of the things they define.
type Def interface {
	// Repr returns the textual representation of the definition.
	Repr() string

	// Public returns if the given definition is public.
	Public() bool
}

// -----------------------------------------------------------------------------

// FuncDef is a MIR function definition.  Note that operators also get converted
// into functions that are marked to be inlined automatically.  The naming
// convention for operator functions is as follows:
// `<global-prefix>.oper[<op-name>: <op-signature>]` followed by any other
// trailing information.
type FuncDef struct {
	Name       string
	Args       []ast.FuncArg
	ReturnType typing.DataType
	Pub        bool
	Inline     bool
}

func (mfd *FuncDef) Repr() string {
	sb := strings.Builder{}
	sb.WriteString("func ")

	if mfd.Inline {
		sb.WriteString("inline ")
	}

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

func (mfd *FuncDef) Public() bool {
	return mfd.Pub
}
