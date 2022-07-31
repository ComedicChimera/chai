package types

// substitution represents the solver's guess for what type should be inferred
// for a given type variable.
type substitution interface {
	// Type returnes the guessed type.
	Type() Type

	// Finalize is called hen this substitution's value is inferred as the final
	// value for a type variable.  This method should be used to handle any
	// side-effects of type inference (eg. setting the package ID of the
	// operator that was selected).
	Finalize()
}

/* -------------------------------------------------------------------------- */

// basicSubstitution is a general purpose substitution: used for all cases which
// don't require any special substitution logic.
type basicSubstitution struct {
	typ Type
}

func (bs basicSubstitution) Type() Type {
	return bs.typ
}

func (bs basicSubstitution) Finalize() {
	// Do Nothing
}

/* -------------------------------------------------------------------------- */

// operatorSubstitution is a specialized kind of substitution used for operator
// overloading: it contains additional information needed to determine which
// overload was selected.
type operatorSubstitution struct {
	// The index of the overload associated with this substitution.
	ndx int

	// The signature type associated with the overload.
	signature Type

	// A callback function used to indicate the index of the overload that was
	// chosen. This is a very hacky way to perform dependency injection since Go
	// is go.
	setOverload func(int)
}

func (os *operatorSubstitution) Type() Type {
	return os.signature
}

func (os *operatorSubstitution) Finalize() {
	os.setOverload(os.ndx)
}
