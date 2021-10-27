package ast

// Definition represents a top level definition in user source code.
type Definition interface {
	// Names returns the list of names that this definition defines.
	Names() []string

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
