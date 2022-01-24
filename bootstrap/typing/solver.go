package typing

import (
	"chai/report"
	"fmt"
	"log"
)

// TypeVar represents a Hindley-Milner type variable.  It is an instance of the
// DataType interface so that it can be used as a data type.  Each type
// variable has an ID that is unique to its solution context.
type TypeVar struct {
	ID       int
	Value    DataType
	Position *report.TextPosition

	// displayName is the name that is displayed when `Repr` is called on this
	// type.
	displayName string

	// shouldDefault indicates that this type variable should default to the
	// first overload that remains after solving completes if there is no valid
	// substitution remaining for it.
	shouldDefault bool
}

func (tv *TypeVar) Repr() string {
	if tv.Value != nil {
		return tv.Value.Repr()
	}

	return tv.displayName
}

func (tv *TypeVar) equals(other DataType) bool {
	if tv.Value != nil {
		return Equals(tv.Value, other)
	}

	// Equals cannot be used on a type variable until it is determined.
	log.Fatalln("Equals used on an undetermined type variable")
	return false
}

func (tv *TypeVar) equiv(other DataType) bool {
	if tv.Value != nil {
		return Equiv(tv.Value, other)
	}

	// Equiv cannot be used on a type variable until it is determined.
	log.Fatalln("Equiv used on an undetermined type variable")
	return false
}

// -----------------------------------------------------------------------------

// solutionState represents the substitutions and overloads being determined for
// types by the solver: ie. the solver's current state.
type solutionState struct {
	// Substitutions is a map of type variable IDs to the type that they will be
	// set equal to.
	Substitutions map[int]DataType

	// OverloadSets is a map of type variable IDs to the set of possible
	// substitutions for the type variable.  These act as a special kind of
	// constraint that restricts what kinds of substitutions are possible.
	OverloadSets map[int][]DataType
}

// newState creates a new solution state.
func newState() *solutionState {
	return &solutionState{
		Substitutions: map[int]DataType{},
		OverloadSets:  map[int][]DataType{},
	}
}

// copyState duplicates completely this solution state.
func (ss *solutionState) copyState() *solutionState {
	newSS := newState()

	for tvid, sub := range ss.Substitutions {
		newSS.Substitutions[tvid] = sub
	}

	for tvid, overloads := range ss.OverloadSets {
		newOverloads := make([]DataType, len(overloads))
		copy(newOverloads, overloads)
		newSS.OverloadSets[tvid] = newOverloads
	}

	return newSS
}

// constraint represents a Hindley-Milner type constraint: it asserts that two
// types are equivalent to each other.
type constraint struct {
	Lhs, Rhs DataType

	// Position is the position of the expression that applied the constraint.
	Position *report.TextPosition
}

// -----------------------------------------------------------------------------

// Solver is the type solver for Chai: it is responsible for determining the
// types of all expressions in the language and for checking that those types
// are correct.  The solver uses Hindley-Milner type inferencing with some
// augmentations for overloading and assertions.  The solver operates in a
// solution context which is the small area of the program it is currently
// considering.  For example, the body of a function is considered a single
// solution context.  One solver per file.
type Solver struct {
	// ctx is the compilation context of the solver.
	ctx *report.CompilationContext

	// vars is the list of type variables in the solver's current solution
	// context.  All these variables must be given substitutions in order for
	// the context to be considered solved.  The substitutions are stored in
	// the type variable's value field.  The type variables ID corresponds
	// to its position within this list.
	vars []*TypeVar

	// constraints is the list of type constraints applied in the solution
	// context.
	constraints []*constraint

	// globalOverloadSets is the global map of known overload sets.
	globalOverloadSets map[int][]DataType

	// localState is the working state of the solver to be updated and copied
	// while a constraint is being unified.
	localState *solutionState

	// shouldError indicates whether the solver should report an error on
	// unification failure.  This is useful for test unification.  This flag is
	// only considered during constraint unification.
	shouldError bool

	// asserts is the list of assertions to be applied after type solving.
	asserts []typeAssert
}

// NewSolver creates a new type solver.
func NewSolver(ctx *report.CompilationContext) *Solver {
	return &Solver{
		ctx:                ctx,
		globalOverloadSets: make(map[int][]DataType),
		shouldError:        true,
	}
}

// NewTypeVar creates a new type variable in the given solution context.
func (s *Solver) NewTypeVar(pos *report.TextPosition, displayName string) *TypeVar {
	tv := &TypeVar{ID: len(s.vars), Position: pos, displayName: displayName}
	s.vars = append(s.vars, tv)
	return tv
}

// NewTypeVarWithOverloads creates a new overloaded type variable in the current
// solution context.
func (s *Solver) NewTypeVarWithOverloads(pos *report.TextPosition, displayName string, shouldDefault bool, overloads ...DataType) *TypeVar {
	tv := &TypeVar{
		ID:            len(s.vars),
		Position:      pos,
		displayName:   displayName,
		shouldDefault: shouldDefault,
	}

	s.vars = append(s.vars, tv)
	s.globalOverloadSets[tv.ID] = overloads
	return tv
}

// Constrain adds a new equivalency constraint between types.
func (s *Solver) Constrain(lhs, rhs DataType, pos *report.TextPosition) {
	s.constraints = append(s.constraints, &constraint{
		Lhs:      lhs,
		Rhs:      rhs,
		Position: pos,
	})
}

// Solve solves the given solution context and returns if the solution was
// successful.  It reports errors as necessary.  It also clears the
// solution context for the next solve.
func (s *Solver) Solve() bool {
	// ensure the solution context is cleared
	defer func() {
		s.constraints = nil
		s.vars = nil
		s.globalOverloadSets = make(map[int][]DataType)
		s.localState = nil
		s.asserts = nil
	}()

	// unify constraints
	for _, cons := range s.constraints {
		// create a new local state for the constraint
		s.localState = newState()

		if !s.unify(cons.Lhs, cons.Rhs, cons.Position) {
			return false
		}

		// merge the completed local state into the global state
		s.mergeState()
	}

	// apply any default substitutions for type variables before checking
	// undetermined to make sure all possible substitutions are fully considered
	// before erroring.
	for _, tv := range s.vars {
		if tv.shouldDefault {
			if tv.Value == nil {
				// we know that overloads exist since the type was marked
				// as defaulting, it has no substitution, and we didn't
				// exit from a unification failure earlier.
				overloads := s.globalOverloadSets[tv.ID]

				// create a new local state for the unification
				s.localState = newState()

				// unification here should never fail because invalid overloads
				// have already been pruned out.
				s.unify(tv, overloads[0], nil)

				// merge the completed local state into the global state
				s.mergeState()
			}
		}
	}

	// check for undetermined type variables
	allSolved := true
	for _, tv := range s.vars {
		if tv.Value == nil {
			report.ReportCompileError(
				s.ctx,
				tv.Position,
				fmt.Sprintf("undetermined type variable: `T%d`", tv.ID),
			)

			allSolved = false
		}
	}

	// check type assertions
	if allSolved {
		for _, assert := range s.asserts {
			if !assert.Apply() {
				report.ReportCompileError(
					s.ctx,
					assert.Position(),
					assert.FailMsg(),
				)

				allSolved = false
			}
		}
	}

	return allSolved
}

// -----------------------------------------------------------------------------

// unify unifies a given pair of types -- asserting that they are equivalent.
func (s *Solver) unify(lhs, rhs DataType, pos *report.TextPosition) bool {
	// first check for type variables: start with RHS since we will switch over
	// the type of LHS and check for its type variables then.
	if rhTypeVar, ok := rhs.(*TypeVar); ok {
		// double type variable case: if the two type variables have the same
		// ID, we know they are equivalent -- this check prevents infinite
		// recursion in `unify`.
		if lhTypeVar, ok := lhs.(*TypeVar); ok && lhTypeVar.ID == rhTypeVar.ID {
			return true
		}

		return s.unifyTypeVar(rhTypeVar.ID, lhs, pos)
	}

	// switch over the values of LHS knowing RHS is not a type variable
	switch v := lhs.(type) {
	case *TypeVar:
		return s.unifyTypeVar(v.ID, rhs, pos)
	case TupleType:
		if rtt, ok := rhs.(TupleType); ok && len(v) == len(rtt) {
			for i, elemType := range v {
				if !s.unify(elemType, rtt[i], pos) {
					return false
				}
			}

			return true
		}
	case *FuncType:
		if rft, ok := rhs.(*FuncType); ok && len(v.Args) == len(rft.Args) {
			for i, arg := range v.Args {
				if !s.unify(arg, rft.Args[i], pos) {
					return false
				}
			}

			return s.unify(v.ReturnType, rft.ReturnType, pos)
		}
	default:
		if Equiv(lhs, rhs) {
			return true
		}
	}

	// if we reach here, the types didn't match: unification fails
	if s.shouldError {
		report.ReportCompileError(
			s.ctx,
			pos,
			fmt.Sprintf("type mismatch: `%s` v `%s`", lhs.Repr(), rhs.Repr()),
		)
	}

	return false
}

// unifyTypeVar performs unification for a type variable assuming that the type
// variable is on the LHS.  This shouldn't matter because equivalency is always
// commutative.  It assumes that the other data type is NOT a type variable.
func (s *Solver) unifyTypeVar(id int, other DataType, pos *report.TextPosition) bool {
	if sub, ok := s.getSubstitution(id); ok {
		// if the type var has a substitution, we just unify against that
		return s.unify(sub, other, pos)
	} else if overloads, ok := s.getOverloads(id); ok {
		// perform overload reduction since the type has no known substitution
		return s.reduceOverloads(id, overloads, other, pos)
	} else {
		// the type variable has no substitution and no overloads so we
		// just update the working state with a new substitution for it
		s.localState.Substitutions[id] = other
		return true
	}
}

// reduceOverloads determines which overloads if any of a given data type are
// valid based on a unification.  It returns `false` if there are no matching
// overloads.  It will update the local substitutions and overloads as
// necessary.
func (s *Solver) reduceOverloads(id int, overloads []DataType, other DataType, pos *report.TextPosition) bool {
	// copy the local state before performing test unification
	cachedLocalState := s.localState.copyState()

	// turn off error reporting for test unification and save the outer
	// error reporting flag so we can restore it once we're done.
	errorFlag := s.shouldError
	s.shouldError = false

	// attempt to unify with each overload and keep track of which overloads
	// are still valid as a result of these unifications.
	var validOverloads []DataType
	for _, overload := range overloads {
		if s.unify(overload, other, pos) {
			validOverloads = append(validOverloads, overload)
		}

		// reset state for next test unification; since we always reset this at
		// the end of the loop, we know that we don't need to restore it at the
		// end of this function.
		s.localState = cachedLocalState.copyState()
	}

	// restore outer error reporting flag
	s.shouldError = errorFlag

	// check the number of valid overloads to determine how to update local state
	switch len(validOverloads) {
	case 0:
		// no valid overloads => unification fails
		if s.shouldError {
			report.ReportCompileError(
				s.ctx,
				pos,
				fmt.Sprintf("no type overload of `%s` matches type `%s`", s.vars[id].Repr(), other.Repr()),
			)
		}

		return false
	case 1:
		// one matching overload => overload becomes substitution
		s.localState.Substitutions[id] = validOverloads[0]
		return s.unify(validOverloads[0], other, pos)
	default:
		// multiple matching overloads => update local overloads to remove any
		// invalid ones as necessary.
		s.localState.OverloadSets[id] = validOverloads
		return true
	}
}

// -----------------------------------------------------------------------------

// getSubstitution gets the substitution for a type variable if they exist.
func (s *Solver) getSubstitution(id int) (DataType, bool) {
	if sub, ok := s.localState.Substitutions[id]; ok {
		return sub, true
	}

	if s.vars[id].Value != nil {
		return s.vars[id].Value, true
	}

	return nil, false
}

// getOverloads gets the overloads for a type variable if they exist.
func (s *Solver) getOverloads(id int) ([]DataType, bool) {
	if overloads, ok := s.localState.OverloadSets[id]; ok {
		return overloads, true
	}

	if overloads, ok := s.globalOverloadSets[id]; ok {
		return overloads, true
	}

	return nil, false
}

// mergeState merges the local state into the global state.
func (s *Solver) mergeState() {
	for id, sub := range s.localState.Substitutions {
		s.vars[id].Value = sub
	}

	for id, overloadSet := range s.localState.OverloadSets {
		s.globalOverloadSets[id] = overloadSet
	}
}
