package sem

// HIRRoot is the high-level intermediate representation node for a file
type HIRRoot struct {
	Defs []HIRDef
}

// HIRDef is the parent interface for all definition HIR Nodes
type HIRDef interface {
	Sym() *Symbol
	Annotations() map[string]*Annotation
}

// DefBase is the base struct for all definitions
type DefBase struct {
	sym    *Symbol
	annots map[string]*Annotation
}

func NewDefBase(sym *Symbol, annots map[string]*Annotation) DefBase {
	return DefBase{
		sym:    sym,
		annots: annots,
	}
}

func (db *DefBase) Sym() *Symbol {
	return db.sym
}

func (db *DefBase) Annotations() map[string]*Annotation {
	return db.annots
}

// HIRFuncDef represents a function definition
type HIRFuncDef struct {
	DefBase

	// ArgumentInitializers stores the HIR expressions initializing all optional
	// arguments of a function.
	ArgumentInitializers map[string]HIRExpr

	// Body is the body of the function
	Body HIRExpr
}

// HIROperDef represents an operator definition
type HIROperDef struct {
	DefBase

	// OperCode is the operator code that is used to look up an operator. These
	// are derived from the tokens used for operators.
	OperCode int

	// Body is the body of the function
	Body HIRExpr
}
