package walk

import (
	"chai/logging"
	"chai/sem"
	"chai/syntax"
	"chai/typing"
	"fmt"
)

// defineOperator attempts to define an operator in the global scope of a file
func (w *Walker) defineOperator(opCode int, opName string, overload *sem.OperatorOverload, pos *logging.TextPosition) bool {
	if operator, ok := w.SrcFile.ImportedOperators[opCode]; ok {
		for _, loverload := range operator.Overloads {
			if overload.CollidesWith(loverload) {
				w.logError(
					fmt.Sprintf("operator definition conflicts with operator defined in `%s`", loverload.SrcPackage.Name),
					logging.LMKDef,
					pos,
				)

				return false
			}
		}
	}

	if operator, ok := w.SrcFile.Parent.GlobalOperators[opCode]; ok {
		if !operator.AddOverload(overload) {
			w.logError(
				"operator's signature conflicts with that of globally defined operator",
				logging.LMKDef,
				pos,
			)

			return false
		}
	} else {
		argsForm, returnForm := getOperatorForm(opCode, len(overload.Quantifiers) > 2)

		w.SrcFile.Parent.GlobalOperators[opCode] = &sem.Operator{
			Name:       opName,
			Overloads:  []*sem.OperatorOverload{overload},
			ArgsForm:   argsForm,
			ReturnForm: returnForm,
		}
	}

	return true
}

// getOperatorFromSignature takes in the signature of the operator and produces
// an appropriate operator overload and form for that signature.
func (w *Walker) getOverloadFromSignature(signature *typing.FuncType) (*sem.OperatorOverload, []int, int) {
	overload := &sem.OperatorOverload{
		SrcPackage: w.SrcFile.Parent,
	}

	argsForm := make([]int, len(signature.Args))

outerloop:
	for i, arg := range signature.Args {
		for j, q := range overload.Quantifiers {
			if typing.Equivalent(q.QType, arg.Type) {
				argsForm[i] = j
				continue outerloop
			}
		}

		argsForm[i] = len(overload.Quantifiers)

		overload.Quantifiers = append(overload.Quantifiers, &sem.OperatorQuantifier{
			QType: arg.Type,
			// TODO: handle type parameters
		})
	}

	for j, q := range overload.Quantifiers {
		if typing.Equivalent(q.QType, signature.ReturnType) {
			return overload, argsForm, j
		}
	}

	overload.Quantifiers = append(overload.Quantifiers, &sem.OperatorQuantifier{
		QType: signature.ReturnType,
		// TODO: handle type parameters
	})
	return overload, argsForm, len(overload.Quantifiers) - 1
}

// getOperatorForm gets the form of the expected quantifiers for an operator
func getOperatorForm(opCode int, binary bool) ([]int, int) {
	switch opCode {
	case syntax.COLON:
		// slice operator
		return []int{0, 1, 1}, 2
	case syntax.LBRACKET:
		// subscript/index operator
		if binary {
			return []int{0, 1}, 2
		} else {
			// ternary/mutation form
			return []int{0, 1, 2}, 3
		}

	default:
		if binary {
			return []int{0, 0}, 0
		} else {
			return []int{0}, 0
		}
	}
}
