package walk

import (
	"chai/logging"
	"chai/sem"
	"chai/syntax"
	"errors"
	"fmt"
)

// defineOperator attempts to define an operator in the global scope of a file
func (w *Walker) defineOperator(opCode int, opName string, overload *sem.OperatorOverload, argCount int, pos *logging.TextPosition) bool {
	// TODO: handle inplace operators

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
		if operatorSet, ok := w.SrcFile.Parent.GlobalOperators[opCode]; ok {
			w.SrcFile.Parent.GlobalOperators[opCode] = append(operatorSet, &sem.Operator{
				Name:      opName,
				Overloads: []*sem.OperatorOverload{overload},
				Arity:     argCount,
			})
		} else {
			w.SrcFile.Parent.GlobalOperators[opCode] = []*sem.Operator{{
				Name:      opName,
				Overloads: []*sem.OperatorOverload{overload},
				Arity:     argCount,
			}}
		}
	}

	return true
}

// checkOperatorArity checks to see if the operator has a correct arity.  It
// takes in an argument count to get a form and returns an error that contains
// the "humanized" text describing the expecting arguments
func checkOperatorArity(opCode, arity int) error {
	switch opCode {
	case syntax.COLON:
		// slice operator
		if arity != 3 && arity != 4 {
			return errors.New("3 or 4 arguments")
		}
	case syntax.LBRACKET:
		// index operator
		if arity != 2 && arity != 3 {
			return errors.New("2 or 3 arguments")
		}
	case syntax.MINUS:
		// subtract or negate
		if arity != 1 && arity != 2 {
			return errors.New("1 or 2 arguments")
		}
	case syntax.NOT, syntax.COMPL:
		// unary operators
		if arity != 1 {
			errors.New("1 argument")
		}
	default:
		// binary operators
		if arity != 2 {
			errors.New("2 arguments")
		}
	}

	return nil
}

// lookupOperator looks up an operator based on the opcode and arity
func (w *Walker) lookupOperator(opCode int, arity int) (*sem.Operator, bool) {
	// we can look up operators in any order since they can not collide
	if opSet, ok := w.SrcFile.ImportedOperators[opCode]; ok {
		for _, op := range opSet {
			if op.Arity == arity {
				return op, true
			}
		}
	}

	if opSet, ok := w.SrcFile.Parent.GlobalOperators[opCode]; ok {
		for _, op := range opSet {
			if op.Arity == arity {
				return op, true
			}
		}
	}

	return nil, false
}
