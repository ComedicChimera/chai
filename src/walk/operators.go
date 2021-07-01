package walk

import (
	"chai/logging"
	"chai/sem"
	"chai/syntax"
	"chai/typing"
	"errors"
	"fmt"
)

// defineOperator attempts to define an operator in the global scope of a file
func (w *Walker) defineOperator(opCode int, opName string, overload *sem.OperatorOverload, argCount int, pos *logging.TextPosition) bool {
	if operator, ok := sem.GetOperatorFromTable(w.SrcFile.ImportedOperators, opCode, argCount); ok {
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

	if operator, ok := sem.GetOperatorFromTable(w.SrcFile.Parent.GlobalOperators, opCode, argCount); ok {
		if !operator.AddOverload(overload) {
			w.logError(
				"operator's signature conflicts with that of globally defined operator",
				logging.LMKDef,
				pos,
			)

			return false
		}
	} else {
		argsForm, returnForm, _ := getOperatorForm(opCode, argCount)

		if operatorSet, ok := w.SrcFile.Parent.GlobalOperators[opCode]; ok {
			w.SrcFile.Parent.GlobalOperators[opCode] = append(operatorSet, &sem.Operator{
				Name:       opName,
				Overloads:  []*sem.OperatorOverload{overload},
				ArgsForm:   argsForm,
				ReturnForm: returnForm,
			})
		} else {
			w.SrcFile.Parent.GlobalOperators[opCode] = []*sem.Operator{{
				Name:       opName,
				Overloads:  []*sem.OperatorOverload{overload},
				ArgsForm:   argsForm,
				ReturnForm: returnForm,
			}}
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

// getOperatorForm gets the form of the expected quantifiers for an operator. It
// takes in an argument count to get a form and returns an error that contains
// the "humanized" text describing the expecting arguments
func getOperatorForm(opCode int, argCount int) ([]int, int, error) {
	switch opCode {
	case syntax.COLON:
		// slice operator
		if argCount == 3 {
			return []int{0, 1, 1}, 2, nil
		} else if argCount == 4 {
			// mutation form
			return []int{0, 1, 1, 0}, 2, nil
		} else {
			return nil, -1, errors.New("3 or 4 arguments")
		}
	case syntax.LBRACKET:
		// subscript/index operator
		if argCount == 2 {
			return []int{0, 1}, 2, nil
		} else if argCount == 3 {
			// mutation form
			return []int{0, 1, 2}, 3, nil
		} else {
			return nil, -1, errors.New("2 or 3 arguments")
		}
	case syntax.MINUS:
		if argCount == 1 {
			// negate
			return []int{0}, 0, nil
		} else if argCount == 2 {
			// subtract
			return []int{0, 0}, 0, nil
		} else {
			return nil, -1, errors.New("1 or 2 arguments")
		}
	case syntax.NOT, syntax.COMPL:
		if argCount == 1 {
			return []int{0}, 0, nil
		} else {
			return nil, -1, errors.New("1 argument")
		}
	case syntax.GT, syntax.LT, syntax.EQ, syntax.NEQ, syntax.GTEQ, syntax.LTEQ:
		if argCount == 2 {
			return []int{0, 0}, 1, nil
		} else {
			return nil, -1, errors.New("2 arguments")
		}
	default:
		if argCount == 2 {
			return []int{0, 0}, 0, nil
		} else {
			return nil, -1, errors.New("2 arguments")
		}
	}
}

// lookupOperator looks up an operator based on the opcode and arity
func (w *Walker) lookupOperator(opCode int, arity int) (*sem.Operator, bool) {
	// we can look up operators in any order since they can not collide
	if opSet, ok := w.SrcFile.ImportedOperators[opCode]; ok {
		for _, op := range opSet {
			if len(op.ArgsForm) == arity {
				return op, true
			}
		}
	}

	if opSet, ok := w.SrcFile.Parent.GlobalOperators[opCode]; ok {
		for _, op := range opSet {
			if len(op.ArgsForm) == arity {
				return op, true
			}
		}
	}

	return nil, false
}
