package typing

import (
	"chai/report"
	"fmt"
	"log"
)

// Constraint represents a Hindley-Milner type constraint: it asserts that two
// types are equivalent to each other.
type Constraint struct {
	Lhs, Rhs DataType

	// Position is the position of the expression that applied the constraint.
	Position *report.TextPosition
}

// TypeVar represents a Hindley-Milner type variable.  It is an instance of the
// DataType interface so that it can be used as a data type.  Each type
// variable has an ID that is unique to its solution context.
type TypeVar struct {
	ID       int
	Value    DataType
	Position *report.TextPosition
}

func (tv *TypeVar) Equals(other DataType) bool {
	if tv.Value != nil {
		return tv.Value.Equals(other)
	}

	// Equals cannot be used on a type variable until it is determined.
	log.Fatalln("Equals used on an undetermined type variable")
	return false
}

func (tv *TypeVar) Equiv(other DataType) bool {
	if tv.Value != nil {
		return tv.Value.Equiv(other)
	}

	// Equiv cannot be used on a type variable until it is determined.
	log.Fatalln("Equiv used on an undetermined type variable")
	return false
}

func (tv *TypeVar) Repr() string {
	if tv.Value != nil {
		return tv.Value.Repr()
	}

	return fmt.Sprintf("T%d", tv.ID)
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
	constraints []*Constraint

	// TODO: assertions, overloads
}

// NewSolver creates a new type solver.
func NewSolver(ctx *report.CompilationContext) *Solver {
	return &Solver{ctx: ctx}
}

// NewTypeVar creates a new type variable in the given solution context.
func (s *Solver) NewTypeVar(pos *report.TextPosition) *TypeVar {
	tv := &TypeVar{ID: len(s.vars), Position: pos}
	s.vars = append(s.vars, tv)
	return tv
}

// Constrain adds a new equivalency constraint between types.
func (s *Solver) Constrain(lhs, rhs DataType, pos *report.TextPosition) {
	s.constraints = append(s.constraints, &Constraint{
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
	}()

	// unify constraints
	for _, cons := range s.constraints {
		if !s.unify(cons.Lhs, cons.Rhs, cons.Position) {
			return false
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

		// if the RHS doesn't have a substitution
		if rhTypeVar.Value == nil {
			rhTypeVar.Value = lhs
			return true
		} else {
			// otherwise, unify against the current substitution
			return s.unify(lhs, rhTypeVar.Value, pos)
		}
	}

	// switch over the values of LHS knowing RHS is not a type variable
	switch v := lhs.(type) {
	case *TypeVar:
		// if LHS doesn't have a substitution
		if v.Value == nil {
			v.Value = rhs
			return true
		} else {
			// otherwise, unify against the current substitution
			return s.unify(v.Value, rhs, pos)
		}
	case PrimType:
		if rpt, ok := rhs.(PrimType); ok {
			return v == rpt
		}
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
				if arg.ByRef != rft.Args[i].ByRef || !s.unify(arg.Type, rft.Args[i].Type, pos) {
					return false
				}
			}

			return s.unify(v.ReturnType, rft.ReturnType, pos)
		}
	}

	// if we reach here, the types didn't match: unification fails
	report.ReportCompileError(
		s.ctx,
		pos,
		fmt.Sprintf("type mismatch: `%s` v `%s`", lhs.Repr(), rhs.Repr()),
	)

	return false
}
