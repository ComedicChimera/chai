package ast

import "chai/typing"

// Def represents a top level definition in user source code.
type Def interface {
	// Names returns the list of names that this definition defines.
	Names() []string

	// Annotations returns a map of the annotations applied to this definition.
	Annotations() map[string]string

	// Dependencies is the map of global names this definition depends on.
	Dependencies() map[string]struct{}
}

// DefBase is the base type for all definition types.
type DefBase struct {
	annots map[string]string
	deps   map[string]struct{}
}

func NewDefBase(annots map[string]string) DefBase {
	return DefBase{
		annots: annots,
		deps:   make(map[string]struct{}),
	}
}

func (db *DefBase) Annotations() map[string]string {
	return db.annots
}

func (db *DefBase) Dependencies() map[string]struct{} {
	return db.deps
}

// -----------------------------------------------------------------------------

// FuncDef is an AST node for a function.
type FuncDef struct {
	DefBase

	Name      string
	Signature *typing.FuncType
	Args      []FuncArg
	Body      Expr
}

// FuncArg represents a function argument.
type FuncArg struct {
	Name     string
	Type     typing.DataType
	ByRef    bool
	Constant bool
}

func (fd *FuncDef) Names() []string {
	return []string{fd.Name}
}

// -----------------------------------------------------------------------------

// OperDef is an AST node for an operator definition.
type OperDef struct {
	DefBase

	// OpKind corresponds to the token kind of the operator.
	OpKind int

	Signature *typing.FuncType
	Args      []FuncArg
	Body      Expr
}

func (od *OperDef) Names() []string {
	return nil
}
