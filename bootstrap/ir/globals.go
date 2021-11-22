package ir

import (
	"strings"
)

// GlobalVar represents a global variable declaration. It can correspond to any
// of: static initialized variable, static uninitialized variable, or external
// variable.
type GlobalVar struct {
	// Name is the name of the global variable.  Although this is technically
	// redundant, it makes displaying the output IR source a lot easier and
	// helps to make this variable easier to identify in debugging.
	Name string

	// Typ is the data type of the global variable.
	Typ Type

	// Val is the literal value of the global variable.  It may be `nil` if
	// the global variable is uninitialized or external.
	Val Value
}

// NOTE: Repr only displays the actual declaration of the variable: not its initializer.
func (gv *GlobalVar) Repr() string {
	sb := strings.Builder{}

	sb.WriteString("var @")
	sb.WriteString(gv.Name)

	sb.WriteRune(' ')

	sb.WriteString(gv.Typ.Repr())

	return sb.String()
}

func (gv *GlobalVar) Section() int {
	if gv.Val == nil {
		return SectionBSS
	} else {
		return SectionData
	}
}
