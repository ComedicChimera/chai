package typing

import (
	"chai/logging"
	"fmt"
)

// Solver is the state machine responsible for performing Hindley-Milner type
// inference within expression bodies.  It is the main mechanism by which the
// Walker interacts with the type system.
type Solver struct {
	// lctx is the log context of the parent walker to this solver
	lctx *logging.LogContext

	// vars contains all of the type variables defined in the solution context
	// where the ID of the type variable corresponds to its index.
	vars []*TypeVariable

	// constraints contains all of the type constraints defined in the solution
	// context.  These constraints are in no particular order.
	constraints []*TypeConstraint

	// substitutions is the map of global type substitutions applied to type
	// variables.  This will contain the solver's final deductions for the types
	// of type variables.
	substitutions map[int]*TypeSubstitution
}

// NewSolver creates a new type solver in a given log context
func NewSolver(lctx *logging.LogContext) *Solver {
	return &Solver{
		lctx:          lctx,
		substitutions: make(map[int]*TypeSubstitution),
	}
}

// CreateTypeVar creates a new type variable with a given default type and a
// handler which is called when the type can't be inferred.  The default type
// may be `nil` if there is none.
func (s *Solver) CreateTypeVar(defaultType DataType, handler func()) *TypeVariable {
	s.vars = append(s.vars, &TypeVariable{
		s:                  s,
		ID:                 len(s.vars),
		DefaultType:        defaultType,
		HandleUndetermined: handler,
	})

	return s.vars[len(s.vars)-1]
}

// AddEqConstraint adds a new equality constraint to the solver
func (s *Solver) AddEqConstraint(lhs, rhs DataType, pos *logging.TextPosition) {
	s.constraints = append(s.constraints, &TypeConstraint{
		Lhs:  lhs,
		Rhs:  rhs,
		Kind: TCEquiv,
		Pos:  pos,
	})
}

// AddSubConstraint adds a new subtype constraint to the solver
func (s *Solver) AddSubConstraint(lhs, rhs DataType, pos *logging.TextPosition) {
	s.constraints = append(s.constraints, &TypeConstraint{
		Lhs:  lhs,
		Rhs:  rhs,
		Kind: TCSubType,
		Pos:  pos,
	})
}

// Solve runs the main solution algorithm on the given solution context. This
// context will be cleared after Solve has completed.  It returns a boolean
// indicating whether solution succeeded.
func (s *Solver) Solve() bool {
	// attempt initial unification of the type constraints
	for _, cons := range s.constraints {
		if !s.unify(cons.Lhs, cons.Rhs, cons.Kind) {
			// we don't want to continue solving here since otherwise our type
			// errors may cascade and cause a bunch of other non-related type
			// errors that will just confuse the user
			s.logTypeError(cons.Lhs, cons.Rhs, cons.Kind, cons.Pos)
			return false
		}

	}

	// TODO: apply all type assertions to the types

	// determine final values for all our unknown types
	allEvaluated := true
	for _, tvar := range s.vars {
		if sub, ok := s.substitutions[tvar.ID]; ok {
			if sub.equivTo != nil {
				tvar.EvalType = sub.equivTo
				continue
			} else if SubTypeOf(sub.superTypeOf, sub.subTypeOf) {
				// we always want to deduce the lowest possible type so if the
				// super type bound satisfies the subtype bound, then, we can
				// just place in the type variable.
				tvar.EvalType = sub.superTypeOf
				continue
			}
		}

		// if we reach here, then there was no substitution so we check for
		// default types -- even in the case where we have an unresolved
		// constraint set
		if tvar.DefaultType != nil {
			tvar.EvalType = tvar.DefaultType
		} else {
			tvar.HandleUndetermined()
			allEvaluated = false
		}
	}

	// if any of the types were indeterminate then solving fails
	if !allEvaluated {
		return false
	}

	// simplify all the performed substitutions for the type variables
	for _, tvar := range s.vars {
		if s, ok := s.simplify(tvar.EvalType); ok {
			tvar.EvalType = s
		}
	}

	// clear the solution context for the next solve
	s.vars = nil
	s.constraints = nil
	return true
}

// -----------------------------------------------------------------------------

// unify performs unification on a pair of types based on the given type
// constraint between them (equivalent or subtype).  All substitutions are
// placed in `s.preview` by default but will be checked against the previous map
// of substitutions.  It returns  a boolean indicating whether or not the
// unification was possible.
func (s *Solver) unify(lhs, rhs DataType, consKind int) bool {
	// check for type variables on the right before switching of the left
	if rhTypeVar, ok := rhs.(*TypeVariable); ok {
		// all we need to do to handle type variables alone in unification is
		// add a new substitution for them.  `union` will catch any conflicts in
		// substitution.
		return s.substitute(rhTypeVar.ID, lhs, consKind, false)
	}

	// note: we now know right is not a type var, and that operators can only
	// appear on the right (so left is not an operator)
	switch v := lhs.(type) {
	case *TypeVariable:
		// same logic as for rhs type vars
		return s.substitute(v.ID, rhs, consKind, true)
	case *FuncType:
		if rft, ok := rhs.(*FuncType); ok {
			if v.Async != rft.Async {
				return false
			}

			if len(v.Args) != len(rft.Args) {
				return false
			}

			for i, arg := range v.Args {
				rarg := rft.Args[i]

				if arg.Name != "" && rarg.Name != "" && arg.Name != rarg.Name {
					return false
				}

				if arg.Indefinite != rarg.Indefinite || arg.Optional != rarg.Optional || arg.ByReference != rarg.ByReference {
					return false
				}

				if !s.unify(arg.Type, rarg.Type, TCEquiv) {
					return false
				}

			}

			return s.unify(v.ReturnType, rft.ReturnType, TCEquiv)
		}
	default:
		switch consKind {
		case TCEquiv:
			return Equivalent(lhs, rhs)
		case TCSubType:
			return SubTypeOf(rhs, lhs)
		}
	}

	// if we reach here, we had a type error
	return false
}

// substitute takes in type variable ID, a value to substitute, the constraint
// kind applying the substitution, and a boolean indicating whether the type
// variable is on the left (true) or right (false) side of the constraint.  It
// attempts to perform a substitution according to that constraint and returns a
// boolean indicating whether or not the constraint was valid by the current set
// of substitutions
func (s *Solver) substitute(tvarID int, value DataType, consKind int, isLhs bool) bool {
	// check if there are any previous substitutions applied to the type
	// variable so the constraint can be checked against the current subsitution
	if sub, ok := s.substitutions[tvarID]; ok {
		// validate constraint and update substitutions as necessary

		if consKind == TCEquiv {
			// if the type variable already has an equivalency constraint, then
			// the types must be equivalent; otherwise, the application is not
			// valid
			if sub.equivTo != nil {
				// ordering doesn't matter for equivalency
				return s.unify(sub.equivTo, value, TCEquiv)
			}

			// check that the value is in between the bounds of the type var
			if sub.subTypeOf != nil && !s.unify(sub.subTypeOf, value, TCSubType) {
				return false
			}

			if sub.superTypeOf != nil && !s.unify(value, sub.superTypeOf, TCSubType) {
				return false
			}

			// if we reach here, we know it is within bounds, so we replace the
			// bounded substitution with an exact substitution
			s.substitutions[tvarID].equivTo = value
		} else if isLhs /* type var is super type of value */ {
			// if the variable has a type that it is exactly equivalent to, then
			// we check the passed value against that substituted value
			if sub.equivTo != nil {
				return s.unify(sub.equivTo, value, TCSubType)
			}

			// in order for this substitution to be possible, either the value
			// has to be a sub type of the super type of the type variable or
			// there has to be no super type for the variable
			if sub.superTypeOf != nil && !s.unify(sub.superTypeOf, value, TCSubType) {
				return false
			}

			// if we know that this substitution is valid, we only override if
			// the sub type of the type variable is a sub type of the passed in
			// value -- narrowing the bounds
			if sub.subTypeOf == nil || s.unify(value, sub.subTypeOf, TCSubType) {
				s.substitutions[tvarID].subTypeOf = value
				return true
			}

			// we know the substitution is legal whether or not we update it
		} else /* type var is sub type of value */ {
			// if the variable has a type that it is exactly equivalent to, then
			// we check the passed value against that substituted value
			if sub.equivTo != nil {
				return s.unify(sub.equivTo, value, TCSubType)
			}

			// in order for this substitution to be possible, either the value
			// has to be a super type of the sub type of the type substitution
			// or there has to be no sub type for the variable
			if sub.subTypeOf != nil && !s.unify(value, sub.subTypeOf, TCSubType) {
				return false
			}

			// if we know that this substitution is valid, we only override if
			// the super type of the type variable is a super type of the passed
			// in value -- narrowing the bounds
			if sub.superTypeOf == nil || s.unify(sub.superTypeOf, value, TCSubType) {
				s.substitutions[tvarID].superTypeOf = value
			}

			// we know the substitution is legal whether or not we update it
		}
	} else /* if there was no substitution, add a new substitution */ {
		if consKind == TCEquiv {
			s.substitutions[tvarID] = &TypeSubstitution{
				equivTo: value,
			}
		} else if isLhs {
			// type var on lhs => super type
			s.substitutions[tvarID] = &TypeSubstitution{
				superTypeOf: value,
			}
		} else {
			// type var on rhs => sub type
			s.substitutions[tvarID] = &TypeSubstitution{
				subTypeOf: value,
			}
		}
	}

	// any case that reaches here is a valid substitution
	return true
}

// simplify removes all nested types from the deduced type for a type parameter.
// It also checks for types such as constraint sets that are not valid types on
// their own.
func (s *Solver) simplify(dt DataType) (DataType, bool) {
	// operators never appear here
	switch v := dt.(type) {
	case *TypeVariable:
		return s.simplify(v.EvalType)
	case *FuncType:
		newArgs := make([]*FuncArg, len(v.Args))
		for i, arg := range v.Args {
			if newAdt, ok := s.simplify(arg.Type); ok {
				newArgs[i] = &FuncArg{
					Type:        newAdt,
					Name:        arg.Name,
					ByReference: arg.ByReference,
					Optional:    arg.Optional,
					Indefinite:  arg.Indefinite,
				}
			} else {
				return nil, false
			}
		}

		if newRt, ok := s.simplify(v.ReturnType); ok {
			return &FuncType{
				Args:          newArgs,
				ReturnType:    newRt,
				Async:         v.Async,
				IntrinsicName: v.IntrinsicName,
				Boxed:         v.Boxed,
			}, true
		} else {
			return nil, false
		}
	case *ConstraintSet:
		return nil, false
	}

	return dt, true
}

// logTypeError logs a type error between two data types
func (s *Solver) logTypeError(lhs, rhs DataType, consKind int, pos *logging.TextPosition) {
	var msg string
	switch consKind {
	case TCEquiv:
		msg = fmt.Sprintf("type mismatch: `%s` v. `%s`", lhs.Repr(), rhs.Repr())
	case TCSubType:
		msg = fmt.Sprintf("`%s` must be a subtype of `%s`", rhs.Repr(), lhs.Repr())
	}

	logging.LogCompileError(
		s.lctx,
		msg,
		logging.LMKTyping,
		pos,
	)
}
