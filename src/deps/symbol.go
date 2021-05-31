package deps

// Symbol represents a named symbol (globally or locally)
type Symbol struct {
	// Name is the name of the symbol (as it is referenced in source code)
	Name string

	// DefKind is the kind of definition that produced this definition. This
	// must be one of the enumerated definition kinds below
	DefKind int

	// Public indicates whether or not this symbol is public (exported)
	Public bool

	// Immutable indicates whether or not this symbol can be mutated
	Immutable bool
}

// Enumeration of symbol definition kinds
const (
	DefKindTypeDef    = iota // Type, Class, and Constraint definitions
	DefKindFuncDef           // Function and operator definitions
	DefKindNamedValue        // Variables and other identifiers
)
