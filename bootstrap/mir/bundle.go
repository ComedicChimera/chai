package mir

import "chai/typing"

// Bundle represents a single package once converted into MIR.  It is a
// simplified and organized representation of the language.
type Bundle struct {
	// ID is the unique identifier for the bundle.  This is the same as ID of
	// the package that generated this bundle.
	ID uint64

	// Funcs is the list of functions defined in the bundle.
	Funcs []*FuncDef

	// Types is the list of types defined in the bundle.
	Types []*TypeDef

	// GlobalVars is the list of global variables defined in the bundle.
	GlobalVars []*GlobalVarDef
}

// NewBundle creates a new bundle with a given ID.
func NewBundle(id uint64) *Bundle {
	return &Bundle{ID: id}
}

// -----------------------------------------------------------------------------

// FuncDef represents a function definition in MIR.
type FuncDef struct {
	ParentID    uint64
	Name        string
	Public      bool
	Annotations map[string]string

	Args       []FuncParam
	ReturnType typing.DataType

	Body *FuncBody
}

// FuncParam represents a function parameter in MIR.
type FuncParam struct {
	Name     string
	Type     typing.DataType
	Constant bool
}

// FuncBody represents a function body in MIR.
type FuncBody struct {
	// Locals stores the local variables in the function.
	Locals map[string]typing.DataType

	BodyExpr Expr
}

// -----------------------------------------------------------------------------

// TypeDef represents a type definition in MIR.
type TypeDef struct {
	ParentID    uint64
	Name        string
	Public      bool
	Annotations map[string]string

	DefType typing.DataType
}

// -----------------------------------------------------------------------------

// GlobalVarDef represents a global variable declarations
type GlobalVarDef struct {
	ParentID uint64
	Names    []string
	Type     typing.DataType
	Public   bool

	Initializer *GlobalVarInit
}

// GlobalVarInit represents a global variable initializer.
type GlobalVarInit struct {
	Locals   map[string]typing.DataType
	InitExpr Expr
}
