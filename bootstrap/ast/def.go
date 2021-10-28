package ast

// Def represents a top level definition in user source code.
type Def interface {
	// Names returns the list of names that this definition defines.
	Names() []string

	// Annotations returns a map of the annotations applied to this definition.
	Annotations() map[string]string

	// Kind returns the kind of definition declared by this AST. It must be one
	// of the enumerated definition kinds below.
	Kind() int
}

// Enumeration of kinds of AST definition.
const (
	ASTFuncDef = iota
	ASTStructDef
	ASTAlgebraicDef
	ASTHybridDef
	ASTAliasDef
	ASTVarDef
	ASTConstDef
)

// FuncDef is an AST node for a function.
type FuncDef struct {
	Name   string
	Annots map[string]string
	// FuncType Type
	// Body Expr
}

func (fd *FuncDef) Names() []string {
	return []string{fd.Name}
}

func (fd *FuncDef) Annotations() map[string]string {
	return fd.Annots
}

func (fd *FuncDef) Kind() int {
	return ASTFuncDef
}
