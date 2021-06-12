package sem

// HIRRoot is the high-level intermediate representation node for a file
type HIRRoot struct {
	Defs []HIRDef
}

// HIRDef is the parent interface for all definition HIR Nodes
type HIRDef interface {
	Sym() *Symbol
}

// defBase is the base struct for all definitions
type defBase struct {
	sym *Symbol
}

func (db *defBase) Sym() *Symbol {
	return db.sym
}

// HIRFuncDef represents a function definition
type HIRFuncDef struct {
	defBase

	// ArgumentInitializers stores the HIR expressions initializing all optional
	// arguments of a function.
	ArgumentInitializers map[string]HIRExpr

	// Body is the body of the function
	Body HIRExpr
}
