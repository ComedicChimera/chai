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

	// overloads is the map of the set of values that each type variable can be
	// unified to. Some type variables will have no overloads meaning they can
	// be unified to any value
	overloads map[int][]DataType
}

// NewSolver creates a new type solver in a given log context
func NewSolver(lctx *logging.LogContext) *Solver {
	return &Solver{
		lctx:          lctx,
		substitutions: make(map[int]*TypeSubstitution),
		overloads:     make(map[int][]DataType),
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

// AddOverload adds an overload to a given type variable
func (s *Solver) AddOverload(tvar *TypeVariable, overloads ...DataType) {
	if overloadSet, ok := s.overloads[tvar.ID]; ok {
		s.overloads[tvar.ID] = append(overloadSet, overloads...)
	} else {
		s.overloads[tvar.ID] = overloads
	}
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
		// through deduction multiple type variables will evaluate at once and
		// so we have to test to see whether or not the type has already been
		// evaluated
		if tvar.EvalType == nil && !tvar.EvalFailed {
			allEvaluated = allEvaluated && s.deduce(tvar)
		}

	}

	// clear the solution context for the next solve and return
	s.vars = nil
	s.constraints = nil
	s.substitutions = make(map[int]*TypeSubstitution)
	return allEvaluated
}

// -----------------------------------------------------------------------------

// unify takes two types and a constraint relating them and attempts to find a
// substitution involving those two types that satisfies the constraint.  It
// returns a boolean indicating whether or not the unification was possible.
func (s *Solver) unify(lhs, rhs DataType, consKind int) bool {
	// check for type variables on the right before switching of the left
	if rhTypeVar, ok := rhs.(*TypeVariable); ok {
		// check to see if both arguments are type variables, and return true if
		// they correspond to the same type variable
		if lhTypeVar, ok := lhs.(*TypeVariable); ok && lhTypeVar.ID == rhTypeVar.ID {
			return true
		}

		// otherwise, perform type variable unification on the right side
		return s.unifyTypeVar(rhTypeVar.ID, lhs, consKind, false)
	}

	switch v := lhs.(type) {
	case *TypeVariable:
		// since we know rhs is not a type variable, we can safely perform type
		// variable unification on the left side
		return s.unifyTypeVar(v.ID, rhs, consKind, true)
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

				if arg.Variadic != rarg.Variadic || arg.Optional != rarg.Optional || arg.ByReference != rarg.ByReference {
					return false
				}

				if !s.unify(arg.Type, rarg.Type, TCEquiv) {
					return false
				}

			}

			return s.unify(v.ReturnType, rft.ReturnType, TCEquiv)
		}
	case *VectorType:
		if rvt, ok := rhs.(*VectorType); ok {
			return s.unify(v.ElemType, rvt.ElemType, TCEquiv) &&
				v.IsRow == rvt.IsRow &&
				(v.Size == rvt.Size || v.Size == -1 || rvt.Size == -1)
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

// unifyTypeVar takes in type variable ID, a value the unify it with, the
// constraint kind applying the substitution, and a boolean indicating whether
// the type variable is on the left (true) or right (false) side of the
// constraint.  It checks to see if the value is unifiable with the known
// substitution for the type variable and updates that substitution if
// appropriate.  It returns a boolean indicating unification success.
func (s *Solver) unifyTypeVar(tvarID int, value DataType, consKind int, isLhs bool) bool {
	// check if there are any previous substitutions applied to the type
	// variable so the constraint can be checked against the current
	// subsitution; if there are no substitutions, check to see if the type
	// variable has any overloads. If it does, attempt to reduce the overloads
	// accordingly.  We can check for overloads after substitutions since we
	// know overloads will always be declared before substitutions (logically)
	// but are less common in the general case.  We can to priotize
	// substitutions but handle overloads first
	if sub, ok := s.substitutions[tvarID]; ok {
		// validate constraint and update substitutions as necessary
		return s.updateSubstitution(sub, value, consKind, isLhs)
	} else if overloadSet, ok := s.overloads[tvarID]; ok {
		overloadSet = s.reduceOverloads(overloadSet, value, consKind, isLhs)

		switch len(overloadSet) {
		case 0:
			// constraint eliminated all possible overloads: fail
			return false
		case 1:
			// only one overload remaining: it becomes the new equivalency
			// substitution; we know that no substitutions have been applied to
			// this type variable yet so we can just blindly override
			s.substitutions[tvarID] = &TypeSubstitution{equivTo: overloadSet[0]}
		default:
			// pruned out some possible overloads.  we know that all remaining
			// overloads will be valid by the known type constraints
			s.overloads[tvarID] = overloadSet
		}

		// there were still valid overloads so unification succeeds
		return true
	} else /* if there was no substitution, add a new substitution */ {
		if consKind == TCEquiv {
			s.substitutions[tvarID] = &TypeSubstitution{
				equivTo: value,
			}
		} else if isLhs {
			// type var on lhs => super type
			s.substitutions[tvarID] = &TypeSubstitution{
				lowerBounds: []DataType{value},
			}
		} else {
			// type var on rhs => sub type
			s.substitutions[tvarID] = &TypeSubstitution{
				upperBound: value,
			}
		}
	}

	// any case that reaches here is a valid substitution
	return true
}

// updateSubstitution checks a given value against a substitution according
// to a given constraint and updates that substitution appropriately
func (s *Solver) updateSubstitution(sub *TypeSubstitution, value DataType, consKind int, isLhs bool) bool {
	if consKind == TCEquiv {
		// if the type variable already has an equivalency constraint, then
		// the types must be equivalent; otherwise, the application is not
		// valid
		if sub.equivTo != nil {
			// ordering doesn't matter for equivalency
			return s.unify(sub.equivTo, value, TCEquiv)
		}

		// check that the value is in between the bounds of the type var
		if sub.upperBound != nil && !s.unify(sub.upperBound, value, TCSubType) {
			return false
		}

		if len(sub.lowerBounds) != 0 {
			for _, bound := range sub.lowerBounds {
				if !s.unify(value, bound, TCSubType) {
					return false
				}
			}
		}

		// if we reach here, we know it is within bounds, so we replace the
		// bounded substitution with an exact substitution
		sub.equivTo = value
	} else if isLhs /* type var is super type of value */ {
		// if the variable has a type that it is exactly equivalent to, then
		// we check the passed value against that substituted value
		if sub.equivTo != nil {
			return s.unify(sub.equivTo, value, TCSubType)
		}

		// in order for this substitution to be possible, either the value
		// has to be a sub type of the upper bound on the type substitution
		// or there has to be no upper bound for the variable
		if sub.upperBound != nil && !s.unify(sub.upperBound, value, TCSubType) {
			return false
		}

		// if there is no current lower bound, then this type becomes the new
		// lower bound
		if len(sub.lowerBounds) == 0 {
			sub.lowerBounds = []DataType{value}
		} else {
			// otherwise, we attempt to generalize the lower bound including
			// the new type: if such a generalization is possible
			for _, bound := range sub.lowerBounds {
				if Equivalent(value, bound) {
					return true
				}
			}

			sub.lowerBounds = append(sub.lowerBounds, value)
		}

		// we know the substitution is legal if we reach here
	} else /* type var is sub type of value */ {
		// if the variable has a type that it is exactly equivalent to, then
		// we check the passed value against that substituted value
		if sub.equivTo != nil {
			return s.unify(sub.equivTo, value, TCSubType)
		}

		// in order for this substitution to be possible, either the value
		// has to be a super type of the lower bounds of the type variable or
		// there has to be no lower bound for the variable
		if len(sub.lowerBounds) > 0 {
			// it is valid to unify here since the upper bound may inform the
			// lower bounds: this is not a repeat test
			for _, bound := range sub.lowerBounds {
				if !s.unify(value, bound, TCSubType) {
					return false
				}
			}
		}

		// if we know that this substitution is valid, we only override if
		// the upper bound of the type variable is a super type of the
		// passed in value -- narrowing the bounds
		if sub.upperBound == nil || s.unify(sub.upperBound, value, TCSubType) {
			sub.upperBound = value
			return true
		}

		// we know the substitution is legal whether or not we update it
	}

	return true
}

// reduceOverloads checks the set of overloads for a given type variable against
// a constraint.  It reduces the set of overloads, eliminating all overloads
// that don't satisfy that constraint.
func (s *Solver) reduceOverloads(overloadSet []DataType, value DataType, consKind int, isLhs bool) []DataType {
	var newOverloads []DataType

	for _, overload := range overloadSet {
		if isLhs && s.unify(overload, value, consKind) {
			newOverloads = append(newOverloads, overload)
		} else if !isLhs && s.unify(value, overload, consKind) {
			newOverloads = append(newOverloads, overload)
		}
	}

	return newOverloads
}

// deduce determines the final type for a type variable.  It will also infer
// types for any other type variables used in the value of this type variable
func (s *Solver) deduce(tv *TypeVariable) bool {
	if sub, ok := s.substitutions[tv.ID]; ok {
		// if the type has an equivalency substitution then we evaluate to the
		// simplified substitution type
		if sub.equivTo != nil {
			if dt, ok := s.simplify(sub.equivTo); ok {
				tv.EvalType = dt
				return true
			}

			return false
		}

		if sub.upperBound != nil {
			// if there is an upper bound, then we know that it is correct by
			// the lower bounds, and therefore we can always safely infer it.
			// It is possible for the lower bounds to also contain a valid
			// substitution but that is much less consistent and their
			// generalization is less accurate
			if dt, ok := s.simplify(sub.upperBound); ok {
				// if there is no lower bound, then we can just evaluate to the
				// simplified upper bound -- this is all the information we have
				// to make a deduction and have to work with what we know
				tv.EvalType = dt
				return true
			}
		}

		// if we reach here, either the upper bound wasn't useful to perform
		// deductions or it doesn't exist.  Therefore, we have to use the lower
		// bound.  If there is not exactly one lower bound, then we know the
		// lower bounds were not generalizable: we cannot use them to get the
		// final deduction
		if len(sub.lowerBounds) != 1 {
			if dt, ok := s.simplify(sub.lowerBounds[0]); ok {
				tv.EvalType = dt
				return true
			}
		}
	}

	// if we reach here, the substitution was not usable to determine a final
	// value for the type variable.  We can evaluate to the default type if it
	// exists (assumes that the default type is valid by all the constraints
	// applied to the type -- is this assumption correct?)
	if tv.DefaultType != nil {
		tv.EvalType = tv.DefaultType
		return true
	} else /* deduction has completely failed */ {
		tv.HandleUndetermined()
		tv.EvalFailed = true
		return false
	}
}

// simplify removes all nested types from the deduced type for a type parameter.
// It also checks for types such as constraint sets that are not valid types on
// their own.  This can cause other type variables to deduced
func (s *Solver) simplify(dt DataType) (DataType, bool) {
	switch v := dt.(type) {
	case *TypeVariable:
		if v.EvalFailed {
			// we can't simplify the value of a type variable that has already
			// failed to evaluate
			return nil, false
		} else if v.EvalType == nil && !s.deduce(v) {
			// we also can't simplify if deduction fails
			return nil, false
		}

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
					Variadic:    arg.Variadic,
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
		msg = fmt.Sprintf("`%s` is not a subtype of `%s`", rhs.Repr(), lhs.Repr())
	}

	logging.LogCompileError(
		s.lctx,
		msg,
		logging.LMKTyping,
		pos,
	)
}
