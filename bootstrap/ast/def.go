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

	// Public returns if the definition is public.
	Public() bool
}

// DefBase is the base type for all definition types.
type DefBase struct {
	annots map[string]string
	deps   map[string]struct{}
	public bool
}

func NewDefBase(annots map[string]string, public bool) DefBase {
	return DefBase{
		annots: annots,
		deps:   make(map[string]struct{}),
		public: public,
	}
}

func (db *DefBase) Annotations() map[string]string {
	return db.annots
}

func (db *DefBase) Dependencies() map[string]struct{} {
	return db.deps
}

func (db *DefBase) Public() bool {
	return db.public
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

	Op   *Oper
	Args []FuncArg
	Body Expr
}

func (od *OperDef) Names() []string {
	return nil
}
