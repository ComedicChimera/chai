package depm

// Universe is the set of symbols and operators defined in every package. This
// also contains all of the definitions for intrinsics: all intrinsics are
// visible in all packages without being imported.
type Universe struct {
	// IntrinsicFuncs is a map of all intrinsic functions by name.
	IntrinsicFuncs map[string]*Symbol

	// IntrinsicOperators is a map of all intrinsic operators by kind.
	IntrinsicOperators map[int]*Operator

	// TODO: prelude stuff
}

// NewUniverse creates a new universe for a project.
func NewUniverse() *Universe {
	return &Universe{
		IntrinsicFuncs:     make(map[string]*Symbol),
		IntrinsicOperators: make(map[int]*Operator),
	}
}

// GetSymbol attempts to get a symbol with a specific name from the universe.
func (u *Universe) GetSymbol(name string) (*Symbol, bool) {
	// TODO: prelude stuff

	if sym, ok := u.IntrinsicFuncs[name]; ok {
		return sym, true
	}

	return nil, false
}

// GetOperator attempts to get an operator with a specific kind from the
// universe.
func (u *Universe) GetOperator(kind int) (*Operator, bool) {
	// TODO: prelude stuff

	if oper, ok := u.IntrinsicOperators[kind]; ok {
		return oper, true
	}

	return nil, false
}
