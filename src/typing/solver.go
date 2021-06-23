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
}

// NewSolver creates a new type solver in a given log context
func NewSolver(lctx *logging.LogContext) *Solver {
	return &Solver{lctx: lctx}
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

// CreateTypeVar creates a new type variable with a given default type and a
// handler which is called when the type can't be inferred.  The default type
// may be `nil` if there is none.
func (s *Solver) CreateTypeVar(defaultType DataType, handler func()) *TypeVariable {
	s.vars = append(s.vars, &TypeVariable{
		ID:                 len(s.vars),
		DefaultType:        defaultType,
		HandleUndetermined: handler,
	})

	return s.vars[len(s.vars)-1]
}

// Solve runs the main solution algorithm on the given solution context. This
// context will be cleared after Solve has completed.  It returns a boolean
// indicating whether solution succeeded.
func (s *Solver) Solve() bool {
	// create a global set of type substitutions for the given constraints
	substitutions := make(map[int]*TypeSubstitution)

	// attempt initial unification of the type constraints
	for _, cons := range s.constraints {
		if newSubs, ok := s.unify(cons.Lhs, cons.Rhs, cons.Kind); ok {
			if usubs, ok := s.union(substitutions, newSubs); ok {
				substitutions = usubs
				continue
			}
		}

		s.logTypeError(cons.Lhs, cons.Rhs, cons.Kind, cons.Pos)
	}

	// TODO: apply all type assertions to the types

	// determine final values for all our unknown types
	allEvaluated := true
	for _, tvar := range s.vars {
		if sub, ok := substitutions[tvar.ID]; ok {
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
		tvar.EvalType = s.simplify(tvar.EvalType)
	}

	// clear the solution context for the next solve
	s.vars = nil
	s.constraints = nil
	return true
}

// -----------------------------------------------------------------------------

// unify performs unification on a pair of types based on the given type
// constraint between them (equivalent or subtype).  It returns a map of the
// substitutions applied by this unification as well as a boolean indicating
// whether or not the unification was possible.  The returned map can be `nil`
// if no substitutions were performed.
func (s *Solver) unify(lhs, rhs DataType, consKind int) (map[int]*TypeSubstitution, bool) {
	// check for type variables on the right before switching of the left
	if rhTypeVar, ok := rhs.(*TypeVariable); ok {
		// all we need to do to handle type variables alone in unification is
		// add a new substitution for them.  `union` will catch any conflicts in
		// substitution.
		return s.apply(nil, rhTypeVar.ID, lhs, consKind, false)
	}

	// note: we now know right is not a type var or mono cons set
	switch v := lhs.(type) {
	case *TypeVariable:
		// same logic as for rhs type vars
		return s.apply(nil, v.ID, rhs, consKind, true)
	case *FuncType:
		if rft, ok := rhs.(*FuncType); ok {
			if v.Async != rft.Async {
				return nil, false
			}

			if len(v.Args) != len(rft.Args) {
				return nil, false
			}

			var subs map[int]*TypeSubstitution
			for i, arg := range v.Args {
				rarg := rft.Args[i]

				if arg.Name != "" && rarg.Name != "" && arg.Name != rarg.Name {
					return nil, false
				}

				if arg.Indefinite != rarg.Indefinite || arg.Optional != rarg.Optional || arg.ByReference != rarg.ByReference {
					return nil, false
				}

				if newSubs, ok := s.unify(arg.Type, rarg.Type, TCEquiv); ok {
					if usubs, ok := s.union(subs, newSubs); ok {
						subs = usubs
					} else {
						return nil, false
					}
				} else {
					return nil, false
				}

			}

			if rtsubs, ok := s.unify(v.ReturnType, rft.ReturnType, TCEquiv); ok {
				return s.union(subs, rtsubs)
			} else {
				return nil, false
			}
		}
	case *ConstraintSet:
		// TODO
	default:
		switch consKind {
		case TCEquiv:
			return nil, Equivalent(lhs, rhs)
		case TCSubType:
			return nil, SubTypeOf(rhs, lhs)
		}
	}

	// if we reach here, we had a type error
	return nil, false
}

// apply checks a type against the known type substitutions for a given type
// variable and performs a type substitution for a given type variable in a
// substitution set.  `isLhs` indicates whether the type variable is on the left
// or right side of the type constraint.  This function will mutate the input
// `set` as well as returning the new one; however, only the returned set is
// guaranteed to be valid.  It returns a boolean flag that indicates whether the
// value is acceptable by the current substitution.
func (s *Solver) apply(set map[int]*TypeSubstitution, tvarID int, value DataType, consKind int, isLhs bool) (map[int]*TypeSubstitution, bool) {
	// validate and combine substitutions as applicable
	if set != nil {
		if sub, ok := set[tvarID]; ok {
			if consKind == TCEquiv {
				// if the type variable already has an equivalency constraint,
				// then the types must be equivalent; otherwise, the application
				// is not valid
				if sub.equivTo != nil {
					// ordering doesn't matter for equivalency
					if usubs, ok := s.unify(sub.equivTo, value, TCEquiv); ok {
						return s.union(set, usubs)
					}

					return nil, false
				}

				// check that the value is in between the bounds of the type var
				if sub.subTypeOf != nil {
					if usubs, ok := s.unify(sub.subTypeOf, value, TCSubType); ok {
						set, ok = s.union(set, usubs)

						if !ok {
							return nil, false
						}
					} else {
						return nil, false
					}
				}

				if sub.superTypeOf != nil {
					if usubs, ok := s.unify(value, sub.superTypeOf, TCSubType); ok {
						set, ok = s.union(set, usubs)

						if !ok {
							return nil, false
						}
					} else {
						return nil, false
					}
				}

				// if we reach here, we know it is within bounds, so we replace
				// the bounded substitution with an exact substitution
				sub.equivTo = value
				return set, true
			} else if isLhs /* type var is super type of value */ {
				// if the variable has a type that it is exactly equivalent to,
				// then we check the passed value against that substituted value
				if sub.equivTo != nil {
					if usubs, ok := s.unify(sub.equivTo, value, TCSubType); ok {
						return s.union(set, usubs)
					}

					return nil, false
				}

				// in order for this substitution to be possible, either the
				// value has to be a sub type of the super type of the type
				// variable or there has to be no super type for the variable
				if sub.superTypeOf != nil {
					if usubs, ok := s.unify(sub.superTypeOf, value, TCSubType); ok {
						set, ok = s.union(set, usubs)

						if !ok {
							return nil, false
						}
					} else {
						return nil, false
					}
				}

				// if we know that this substitution is valid, we only override
				// if the sub type of the type variable is a sub type of the
				// passed in value -- narrowing the bounds
				if usubs, ok := s.unify(value, sub.subTypeOf, TCSubType); ok {
					set, ok = s.union(set, usubs)
					sub.subTypeOf = value
					return set, ok
				} else {
					sub.subTypeOf = value
					return set, true
				}
			} else /* type var is sub type of value */ {
				// if the variable has a type that it is exactly equivalent to,
				// then we check the passed value against that substituted value
				if sub.equivTo != nil {
					if usubs, ok := s.unify(value, sub.equivTo, TCSubType); ok {
						return s.union(set, usubs)
					}

					return nil, false
				}

				// in order for this substitution to be possible, either the
				// value has to be a super type of the sub type of the type
				// substitution or there has to be no sub type for the variable
				if sub.subTypeOf != nil {
					if usubs, ok := s.unify(value, sub.subTypeOf, TCSubType); ok {
						set, ok = s.union(set, usubs)

						if !ok {
							return nil, false
						}
					} else {
						return nil, false
					}
				}

				// if we know that this substitution is valid, we only override
				// if the super type of the type variable is a super type of the
				// passed in value -- narrowing the bounds
				if sub.superTypeOf != nil {
					if usubs, ok := s.unify(sub.superTypeOf, value, TCSubType); ok {
						set, ok = s.union(set, usubs)
						sub.superTypeOf = value
						return set, ok
					}
				} else {
					sub.superTypeOf = value
					return set, true
				}
			}
		}
	}

	// there is no preexisting substitution, so we can just add and return
	if consKind == TCEquiv {
		return map[int]*TypeSubstitution{
			tvarID: {equivTo: value},
		}, true
	} else if isLhs {
		// type var on lhs => super type
		return map[int]*TypeSubstitution{
			tvarID: {superTypeOf: value},
		}, true
	} else {
		// type var on rhs => sub type
		return map[int]*TypeSubstitution{
			tvarID: {subTypeOf: value},
		}, true
	}
}

// union combines two substitution sets and returns the combination if possible;
// the input substitutions can be `nil` (empty).  This may mutate the sets
// passed in, but only the returned set is guaranteed to be valid.
func (s *Solver) union(a, b map[int]*TypeSubstitution) (map[int]*TypeSubstitution, bool) {
	if a == nil {
		return b, true
	} else if b == nil {
		return a, true
	}

	for tvarID, sub := range b {
		if sub.equivTo != nil {
			// lhs v rhs doesn't matter for equivalency substitution
			if nextA, ok := s.apply(a, tvarID, sub.equivTo, TCEquiv, true); ok {
				a = nextA
			} else {
				return nil, false
			}
		} else {
			if sub.superTypeOf != nil {
				if nextA, ok := s.apply(a, tvarID, sub.superTypeOf, TCSubType, true); ok {
					a = nextA
				} else {
					return nil, false
				}
			}

			if sub.subTypeOf != nil {
				if nextA, ok := s.apply(a, tvarID, sub.subTypeOf, TCSubType, false); ok {
					a = nextA
				} else {
					return nil, false
				}
			}
		}
	}

	return a, true
}

// simplify removes all nested types from the deduced type for a type parameter
func (s *Solver) simplify(dt DataType) DataType {
	// TODO
	return dt
}

// logTypeError logs a type error between two data types
func (s *Solver) logTypeError(lhs, rhs DataType, consKind int, pos *logging.TextPosition) {
	// TODO: fix to give informative error messages involving substitutions

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
