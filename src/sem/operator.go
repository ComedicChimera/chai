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

	// ArgsForm is the list of arguments to this operator where the number
	// corresponds to that argument's quantifier.
	ArgsForm []int

	// ReturnForm is the number corresponding the return type quantifier
	ReturnForm int
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
			if len(operator.ArgsForm) == argCount {
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

	// Quantifiers is the order list of quantifiers for this overload. A
	// quantifier corresponds to a constraint on the input (or output) type of
	// an operator.  The set of types is the set of valid overloads for the
	// given quantifier
	Quantifiers [][]typing.DataType

	// Public indicates whether or not this overload is externally visible
	Public bool
}

// CollidesWith checks whether or not two operator overloads collide
func (oo *OperatorOverload) CollidesWith(other *OperatorOverload) bool {
	for i, q := range oo.Quantifiers {
		oq := other.Quantifiers[i]

		for _, qType := range q {
			for _, oqType := range oq {
				// collision only occurs with equivalency
				if typing.Equivalent(qType, oqType) {
					return true
				}
			}
		}
	}

	return false
}
