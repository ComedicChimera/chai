package ast

import "chai/typing"

// Def represents a top level definition in user source code.
type Def interface {
	// Names returns the list of names that this definition defines.
	Names() []string

	// Annotations returns a map of the annotations applied to this definition.
	Annotations() map[string]string
}

// FuncDef is an AST node for a function.
type FuncDef struct {
	Name     string
	Annots   map[string]string
	FuncType *typing.FuncType
	Body     Expr
}

func (fd *FuncDef) Names() []string {
	return []string{fd.Name}
}

func (fd *FuncDef) Annotations() map[string]string {
	return fd.Annots
}