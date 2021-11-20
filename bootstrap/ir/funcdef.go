package ir

import "strings"

// FuncDef represents a function defined in IR.  It is composed of a function
// declaration and a body.
type FuncDef struct {
	Decl *FuncDecl
	Body *Block
}

func (fd *FuncDef) Repr() string {
	return fd.Decl.Repr() + ":\n" + fd.Body.Repr()
}

// -----------------------------------------------------------------------------

// FuncDecl is the declaration of a function.  It describes the function's
// signature as well as any relevant attributes for it.
type FuncDecl struct {
	Name       string
	Args       []FuncArg
	ReturnType Type

	// CallConv must be one of the enumerated calling conventions.
	CallConv int

	// Inline indicates whether or not the function should be inlined where
	// possible.  This option is not valid for external functions.
	Inline bool
}

// FuncArg represents a single function argument.
type FuncArg struct {
	Name string
	Typ  Type
}

// Enumeration of calling conventions
const (
	ChaiCC  = iota // Standard Chai calling convention (C calling convention)
	Win64CC        // Windows 64-bit calling convention
)

func (fd *FuncDecl) Repr() string {
	sb := strings.Builder{}
	sb.WriteString("func ")

	if fd.Inline {
		sb.WriteString("inline ")
	}

	if fd.CallConv == Win64CC {
		sb.WriteString("win64cc ")
	}

	sb.WriteRune('@')
	sb.WriteString(fd.Name)

	sb.WriteRune('(')

	for i, arg := range fd.Args {
		sb.WriteString(arg.Typ.Repr())
		sb.WriteString(" $")
		sb.WriteString(arg.Name)

		if i < len(fd.Args)-1 {
			sb.WriteString(", ")
		}
	}

	sb.WriteString(") ")

	sb.WriteString(fd.ReturnType.Repr())

	return sb.String()
}
