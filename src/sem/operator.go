package sem

import "chai/typing"

// Operator represents a global operator definition.  All operators in Chai are
// "consistent" across definitions which essentially means they can be
// accumulated to an explicit operator representation.  For example,`+` always
// takes two types `T` and returns some other type, generally `T`; therefore,
// the explicit form of `+` is `(T, T) -> R` where `T` and `R` are quantified
// appropriately.
type Operator struct {
	// Name is the name of the operator as a string
	Name string

	// Overloads is the list of defined overloads for this operator
	Overloads []*OperatorOverload

	// Arity is the number of arguments this overload takes
	Arity int
}

// AddOverload adds a new operator overload to this operator
func (op *Operator) AddOverload(newOverload *OperatorOverload) bool {
	for _, overload := range op.Overloads {
		if overload.CollidesWith(newOverload) {
			return false
		}
	}

	op.Overloads = append(op.Overloads, newOverload)
	return true
}

// GetOperatorFromTable looks up an operator by both opcode and form
// and returns it if it exists in an operator table.
func GetOperatorFromTable(table map[int][]*Operator, opCode int, argCount int) (*Operator, bool) {
	if operatorSet, ok := table[opCode]; ok {
		for _, operator := range operatorSet {
			if operator.Arity == argCount {
				return operator, true
			}
		}
	}

	return nil, false
}

// OperatorOverload represents a single overload of a given operator
type OperatorOverload struct {
	// SrcPackage is the package this operator is defined in
	SrcPackage *ChaiPackage

	// Signature is the function signature of the overload (either `FuncType` or
	// `GenericType`)
	Signature typing.DataType

	// Public indicates whether or not this overload is externally visible
	Public bool

	// InplaceForm indicates whether or not this operator has an inplace form
	InplaceForm bool
}

// CollidesWith checks whether or not two operator overloads collide
func (oo *OperatorOverload) CollidesWith(other *OperatorOverload) bool {
	// TODO: handle generic collision
	return typing.Equivalent(oo.Signature, other.Signature)
}
