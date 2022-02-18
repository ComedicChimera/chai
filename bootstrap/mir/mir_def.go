package mir

import "chai/typing"

// FuncDef represents a function definition in MIR.
type FuncDef struct {
	Name        string
	Annotations map[string]string
	Public      bool
	Params      []*FuncParam
	ReturnType  typing.DataType

	// Body with be `nil` if this function has none.
	Body *FuncBody
}

// FuncParam represents a MIR function parameter.
type FuncParam struct {
	Name     string
	Type     typing.DataType
	Constant bool
}

// FuncBody is the body of the function.
type FuncBody struct {
	Locals map[string]typing.DataType
	Body   []BlockElem
}

// -----------------------------------------------------------------------------

// TypeDef represents a type definition in MIR.
type TypeDef struct {
	Name        string
	Annotations map[string]string
	Public      bool

	Type typing.DataType // `nil` => opaque type
}

// -----------------------------------------------------------------------------

// GlobalVar represents a global variable definiton in MIR.
type GlobalVar struct {
	Name   string
	Public bool

	Type  typing.DataType
	Value Expr
}
