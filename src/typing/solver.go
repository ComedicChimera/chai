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

	// typeVars contains all of the type variables defined in the solution context
	// where the ID of the type variable corresponds to its index.
	typeVars []*TypeVariable

	// constraints contains all of the type constraints defined in the solution
	// context.  These constraints are in no particular order.
	constraints []*TypeConstraint

	// stateStack is the stack of solution states for the solver.  This stack is
	// pushed to and popped from the facilitate unification testing efficiently
	stateStack []*SolutionState

	// assertions is the list of applied type assertions
	assertions []*TypeAssertion
}

// NewSolver creates a new type solver in a given log context
func NewSolver(lctx *logging.LogContext) *Solver {
	s := &Solver{
		lctx: lctx,
	}
	s.pushState()
	return s
}

// CreateTypeVar creates a new type variable with a given default type and a
// handler which is called when the type can't be inferred.  The default type
// may be `nil` if there is none.
func (s *Solver) CreateTypeVar(defaultType DataType, handler func()) *TypeVariable {
	s.typeVars = append(s.typeVars, &TypeVariable{
		s:                  s,
		ID:                 len(s.typeVars),
		DefaultType:        defaultType,
		HandleUndetermined: handler,
		Required:           true,
	})

	return s.typeVars[len(s.typeVars)-1]
}

// CreateOptionalTypeVar creates a new optional type variable.
func (s *Solver) CreateOptionalTypeVar() *TypeVariable {
	s.typeVars = append(s.typeVars, &TypeVariable{
		s:        s,
		ID:       len(s.typeVars),
		Required: false,
	})

	return s.typeVars[len(s.typeVars)-1]
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

// AddTypeAssertion adds a new type assertion to the solver
func (s *Solver) AddTypeAssertion(kind int, operand DataType, data interface{}, pos *logging.TextPosition) {
	s.assertions = append(s.assertions, &TypeAssertion{
		Operand:    operand,
		AssertKind: kind,
		Data:       data,
		Pos:        pos,
	})
}

// AddOverload adds an overload to a given type variable
func (s *Solver) AddOverload(tvar *TypeVariable, overloadValues ...DataType) {
	// we know this function will be called before `Solve` so state[0] is always
	// the top state on the stack (initial state)
	if overload, ok := s.stateStack[0].overloads[tvar.ID]; ok {
		s.stateStack[0].overloads[tvar.ID].Values = append(overload.Values, overloadValues...)
	} else {
		s.stateStack[0].overloads[tvar.ID] = &TypeOverload{Values: overloadValues}
	}
}

// AddOverloadCorrespondence adds a correspondence between two overloaded type
// variables.  This is used to facilitate tandem reduction: used for operator
// overloading.
func (s *Solver) AddOverloadCorrespondence(aID, bID int) {
	s.stateStack[0].overloads[aID].Corresponds = append(s.stateStack[0].overloads[aID].Corresponds, bID)
	s.stateStack[0].overloads[bID].Corresponds = append(s.stateStack[0].overloads[bID].Corresponds, aID)
}

// Solve runs the main solution algorithm on the given solution context. This
// context will be cleared after Solve has completed.  It returns a boolean
// indicating whether solution succeeded.
func (s *Solver) Solve() bool {
	// always clear the solution context for the next solve after we exit
	defer s.reset()

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

	// fill in default types for all types that can't be determined if possible
	// -- this way those default types can help with deduction
	for _, tvar := range s.typeVars {
		// we know that this function is only run at the top level so we can
		// just treat the top most state as state[0]
		if sub, ok := s.stateStack[0].substitutions[tvar.ID]; ok {
			// if any of these cases are true, then we have a usable
			// substitution: no need to infer default value
			if sub.equivTo != nil || sub.upperBound != nil || len(sub.lowerBounds) == 1 {
				continue
			}
		}

		// no viable substitution: try to unify with default type if that
		// unification is possible, use the default type as the inferred value
		if tvar.DefaultType != nil {
			// testUnify will overwrite the current substitutions if it succeeds
			// so we don't need to test for success or failure
			s.testUnify(tvar, tvar.DefaultType, TCEquiv)
		}
	}

	// determine final values for all our unknown types
	solvingSucceeded := true
	for _, tvar := range s.typeVars {
		// through deduction multiple type variables will evaluate at once and
		// so we have to test to see whether or not the type has already been
		// evaluated
		if tvar.EvalType == nil && !tvar.EvalFailed {
			solvingSucceeded = solvingSucceeded && s.deduce(tvar)
		}
	}

	// now, we can check assertions after type deductions has already been
	// performed: they give no usable information about the underlying type
	if solvingSucceeded {
		for _, assert := range s.assertions {
			switch assert.AssertKind {
			case AssertCast:
				destType := assert.Data.(DataType)

				if !CastTo(assert.Operand, destType) {
					logging.LogCompileError(
						s.lctx,
						fmt.Sprintf("cannot cast `%s` to `%s`", assert.Operand.Repr(), destType.Repr()),
						logging.LMKTyping,
						assert.Pos,
					)

					solvingSucceeded = false
				}
			case AssertNonRef:
				if _, ok := InnerType(assert.Operand).(*RefType); ok {
					logging.LogCompileError(
						s.lctx,
						"unable to create a double reference",
						logging.LMKTyping,
						assert.Pos,
					)
					solvingSucceeded = false
				}
			}
		}
	}

	return solvingSucceeded
}

// -----------------------------------------------------------------------------

// SolutionState stores the current state variables used for type deduction
type SolutionState struct {
	// substitutions is the map of type substitutions applied to type variables
	substitutions map[int]*TypeSubstitution

	// overloads is the map of the set of values that each type variable can be
	// unified to.  Some type variables will have no overloads meaning they can
	// be unified to any value
	overloads map[int]*TypeOverload
}

// pushState pushes a new solution state onto the state stack
func (s *Solver) pushState() {
	s.stateStack = append(s.stateStack, &SolutionState{
		substitutions: make(map[int]*TypeSubstitution),
		overloads:     make(map[int]*TypeOverload),
	})
}

// discardState pops the top state off the state stack and discards it
func (s *Solver) discardState() {
	s.stateStack = s.stateStack[:len(s.stateStack)-1]
}

// mergeState pops the top state off the state stack and merges it into the
// state before it -- assuming correctness between the states
func (s *Solver) mergeState() {
	topState := s.topState()
	s.stateStack = s.stateStack[:len(s.stateStack)-1]
	s.mergeFrom(topState)
}

// mergeFrom merges the the current top state with another state
func (s *Solver) mergeFrom(state *SolutionState) {
	for tvarID, sub := range state.substitutions {
		s.topState().substitutions[tvarID] = sub
	}

	for tvarID, overloadSet := range state.overloads {
		s.topState().overloads[tvarID] = overloadSet
	}
}

// topState gets the state on top of the state stack
func (s *Solver) topState() *SolutionState {
	return s.stateStack[len(s.stateStack)-1]
}

// getSubstitution gets the top most substitution for a type variable
func (s *Solver) getSubstitution(tvarID int) (*TypeSubstitution, bool) {
	for i := len(s.stateStack) - 1; i > -1; i-- {
		if sub, ok := s.stateStack[i].substitutions[tvarID]; ok {
			return sub, true
		}
	}

	return nil, false
}

// getOverload gets the top most overload for a type variable
func (s *Solver) getOverload(tvarID int) (*TypeOverload, bool) {
	for i := len(s.stateStack) - 1; i > -1; i-- {
		if overload, ok := s.stateStack[i].overloads[tvarID]; ok {
			return overload, true
		}
	}

	return nil, false
}

// reset resets the solution context of the solver
func (s *Solver) reset() {
	s.typeVars = nil
	s.constraints = nil
	s.assertions = nil
	s.discardState()
	s.pushState()
}

// -----------------------------------------------------------------------------

// unify takes two types and a constraint relating them and attempts to find a
// substitution involving those two types that satisfies the constraint.  It
// returns a boolean indicating whether or not the unification was possible.
func (s *Solver) unify(lhs, rhs DataType, consKind int) bool {
	// check for type variables on the right before switching of the left
	if rhTypeVar, ok := rhs.(*TypeVariable); ok {
		// check to see if the type variable has already been evaluated (type
		// variable from a different solution context: eg. global variables or
		// argument exprs)
		if rhTypeVar.EvalType != nil {
			return s.unify(lhs, rhTypeVar.EvalType, consKind)
		}

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
		// check to see if the left hand variable has already been evaluated:
		// ie. from another solution context
		if v.EvalType != nil {
			return s.unify(v.EvalType, rhs, consKind)
		}

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
	case *RefType:
		if rrt, ok := rhs.(*RefType); ok {
			return s.unify(v.ElemType, rrt.ElemType, TCEquiv)
		}
	// TODO: alias unification :)
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
	if sub, ok := s.getSubstitution(tvarID); ok {
		// if the substitution was not from the current state, we need to copy
		// it and move it into the current state so that we can update it safely
		if _, ok := s.topState().substitutions[tvarID]; !ok {
			subCopy := &TypeSubstitution{}
			*subCopy = *sub
			s.topState().substitutions[tvarID] = subCopy
			sub = subCopy
		}

		// validate constraint and update substitutions as necessary
		return s.updateSubstitution(sub, value, consKind, isLhs)
	} else if overload, ok := s.getOverload(tvarID); ok {
		if _, ok := s.topState().overloads[tvarID]; !ok {
			// if the overloads come from an outer state, copy the overloads so
			// that append won't manipulate an overload set from the wrong state
			overload = overload.copy()
		}

		// reduce the overloads
		return s.reduceOverloads(tvarID, overload, value, consKind, isLhs)
	} else /* if there was no substitution, add a new substitution */ {
		if consKind == TCEquiv {
			s.topState().substitutions[tvarID] = &TypeSubstitution{
				equivTo: value,
			}
		} else if isLhs {
			// type var on lhs => super type
			s.topState().substitutions[tvarID] = &TypeSubstitution{
				lowerBounds: []DataType{value},
			}
		} else {
			// type var on rhs => sub type
			s.topState().substitutions[tvarID] = &TypeSubstitution{
				upperBound: value,
			}
		}
	}

	// any case that reaches here is a valid substitution
	return true
}

// updateSubstitution checks a given value against a substitution according to a
// given constraint and updates that substitution appropriately. This function
// does mutate the PASSED IN type substitution.
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
			// otherwise, we check to see if the type is already in the lower
			// bound.  If it is, then we just return true because we know that
			// the lower bound won't be affected by this unification
			for _, bound := range sub.lowerBounds {
				if Equivalent(value, bound) {
					return true
				}
			}

			// not in lower bounds => add it
			sub.lowerBounds = append(sub.lowerBounds, value)

			// we then attempt to generalize the lower bounds based on this new
			// addition to them; if it can be generalized, we update the lower
			// bounds.  If it can't, then we leave them as is
			if generalType, ok := s.generalize(sub.lowerBounds); ok {
				sub.lowerBounds = []DataType{generalType}
			}
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

		// if we know that this substitution is valid, we only override if the
		// upper bound of the type variable is a super type of the passed in
		// value -- narrowing the bounds.  We use `testUnify` here since we
		// conditionally decide to override the current state
		if sub.upperBound == nil || s.testUnify(sub.upperBound, value, TCSubType) {
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
func (s *Solver) reduceOverloads(tvarID int, overload *TypeOverload, value DataType, consKind int, isLhs bool) bool {
	remainingOverloads := make(map[int]DataType)

	// store the last state produced through successful unification.  This will
	// be merged later if there is only one overload remaining: thus we only
	// need to store one value for the state
	var state *SolutionState

	// iterate over the overload values and determine which values are valid and
	// which aren't.  store those values, organized by their original index, in
	// the remaining overloads
	for i, overVal := range overload.Values {
		// push a new state to do quick test unification (without forcing a merge)
		s.pushState()

		if isLhs && s.unify(overVal, value, consKind) {
			remainingOverloads[i] = overVal
			state = s.topState()
		} else if !isLhs && s.unify(value, overVal, consKind) {
			remainingOverloads[i] = overVal
			state = s.topState()
		}

		// discard the state for now: we will save it in the states slice if we
		// might want to merge it later
		s.discardState()
	}

	switch len(remainingOverloads) {
	case 0:
		// no valid overloads remain: unification fails
		return false
	case 1:
		// only one valid overload remains: we want to merge the single valid
		// state and then update it with the new substitution information
		s.mergeFrom(state)

		// get the first and only overload remaining
		var i int
		var overVal DataType
		for k, v := range remainingOverloads {
			i = k
			overVal = v
		}

		// perform the initial substitution
		s.topState().substitutions[tvarID] = &TypeSubstitution{equivTo: overVal}

		// then, substitute in all the corresponding overloads
		for _, corrID := range s.getCorresponds(tvarID, make(map[int]struct{})) {
			// should always succeed
			corrOverload, _ := s.getOverload(corrID)
			s.topState().substitutions[corrID] = &TypeSubstitution{
				equivTo: corrOverload.Values[i],
			}
		}
	default:
		// multiple overloads remain: reduce the current overload to only
		// include the remaining valid overloads
		overload.Values = make([]DataType, len(remainingOverloads))
		i := 0
		for _, overVal := range remainingOverloads {
			overload.Values[i] = overVal
			i++
		}

		// reduce all the corrrespondences accordingly
		for _, corrID := range s.getCorresponds(tvarID, make(map[int]struct{})) {
			// this always succeeds since the overload was added as a correspondence
			corrOverload, _ := s.getOverload(corrID)

			// reduce the old overload; this also copies the overload so that we
			// make sure we aren't manipulating state inappropriately
			var reducedValues []DataType
			for i := range remainingOverloads {
				reducedValues = append(reducedValues, corrOverload.Values[i])
			}

			// store the newly created overload back into the top state
			s.topState().overloads[corrID] = &TypeOverload{
				Values:      reducedValues,
				Corresponds: corrOverload.Corresponds,
			}
		}
	}

	// if we reach here, we know some overloading was successful
	return true
}

// getCorresponds gets all the correspondences for an overload
func (s *Solver) getCorresponds(tvarID int, alreadyFound map[int]struct{}) []int {
	alreadyFound[tvarID] = struct{}{}

	// should always succeed
	overload, _ := s.getOverload(tvarID)

	var corresponds []int
	for _, corrID := range overload.Corresponds {
		if _, ok := alreadyFound[corrID]; !ok {
			corresponds = append(corresponds, corrID)

			corresponds = append(corresponds, s.getCorresponds(corrID, alreadyFound)...)
		}
	}

	return corresponds
}

// testUnify tests if a given constraint is unifiable.  This function will
// only preserve its state changes if unification succeeds.
func (s *Solver) testUnify(lhs, rhs DataType, consKind int) bool {
	s.pushState()
	if s.unify(lhs, rhs, consKind) {
		s.mergeState()
		return true
	} else {
		s.discardState()
		return false
	}
}

// generalize takes a list of data types and attempts to produce a most general
// type from them.  For example, `generalize(i32, i64, i16) => i64`.  It does
// have limits on what it can generalize: it will never generalize to a type
// that wasn't in the original set of data types.  So `generalize(i33, rune) !=
// any`.  This function will only preserve its state changes if generalization
// succeeds.
func (s *Solver) generalize(types []DataType) (DataType, bool) {
	var generalType DataType

	// create a new state so that any unification merges (for testUnify or
	// unwrapped unify) do not affect the outer state
	s.pushState()

	for _, dt := range types {
		if generalType == nil {
			generalType = dt
			continue
		}

		// we use test unify since we are testing to see if the new most general
		// type could be the current type (current type is a super type of the
		// general type).  Because sub typing is transitive, we can always
		// override to a more general super type
		if s.testUnify(dt, generalType, TCSubType) {
			generalType = dt
		} else if s.unify(generalType, dt, TCSubType) {
			// we do not use `testUnify` above since this is a failure case: the
			// current type is a subtype of the general type meaning the general
			// type can't be determined for the current set of values.  Discard
			// the generalization state and carry on
			s.discardState()
			return nil, false
		}
	}

	s.mergeState()
	return generalType, true
}

// -----------------------------------------------------------------------------

// deduce determines the final type for a type variable.  It will also infer
// types for any other type variables used in the value of this type variable
func (s *Solver) deduce(tv *TypeVariable) bool {
	// only remaining state at the time of deduction is the top state
	if sub, ok := s.stateStack[0].substitutions[tv.ID]; ok {
		// if the type has an equivalency substitution then we evaluate to the
		// simplified substitution type
		if sub.equivTo != nil {
			if dt, ok := s.simplify(sub.equivTo); ok {
				tv.EvalType = dt
				return true
			}
		} else if len(sub.lowerBounds) == 1 {
			// if there is a lower bound containing one elements, then we know
			// that this lower bound is the most specific possible deduction for
			// this type that is still general enough to satisfy all constraints
			// placed on it.  Thus, we can take that lower type to be our result
			if dt, ok := s.simplify(sub.lowerBounds[0]); ok {
				tv.EvalType = dt
				return true
			}
		} else if sub.upperBound != nil {
			// if there is an upper bound, then we know that it is correct by
			// the lower bounds, and therefore we can always safely infer it.
			if dt, ok := s.simplify(sub.upperBound); ok {
				tv.EvalType = dt
				return true
			}
		}
	}

	// if we reach here, deduction (including default types) has completely
	// failed for this type variable: handle undetermined appropriately
	if tv.Required {
		tv.HandleUndetermined()
	}

	tv.EvalFailed = true
	return false
}

// simplify removes all nested types from the deduced type for a type parameter.
// This can cause other type variables to deduced.  If simplification fails then
// deduction for the type variable whose value is being simplified as also
// failed.
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
