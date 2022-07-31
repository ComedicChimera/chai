package types

import (
	"chaic/report"
	"fmt"
)

// NB: See `docs/type_solver.md` for a reasonably complete explanation of the
// type solving algorithm.

// Solver is the primary mechanism for performing type checking and deduction.
type Solver struct {
	// The list of type variable nodes in the solution graph.
	typeVarNodes []*typeVarNode

	// The map of substitution nodes in the solution graph keyed by ID.
	subNodes map[uint64]*subNode

	// The counter used to generate substitution IDs.
	subIDCounter uint64

	// The set of IDs corresponding to type variables that cannot have more
	// substitutions added to them: they are completely inferred.
	completes map[uint64]struct{}

	// The list of applied cast assertions.
	castAsserts []*castAssert
}

// NewSolver creates a new type solver.
func NewSolver() *Solver {
	return &Solver{
		subNodes:  make(map[uint64]*subNode),
		completes: make(map[uint64]struct{}),
	}
}

// NewTypeVar creates a new type variable in the solution context.
func (s *Solver) NewTypeVar(name string, span *report.TextSpan) *TypeVariable {
	tv := &TypeVariable{
		ID:     uint64(len(s.typeVarNodes)),
		Name:   name,
		Span:   span,
		parent: s,
	}

	s.typeVarNodes = append(s.typeVarNodes, &typeVarNode{Var: tv})

	return tv
}

// AddLiteralOverloads binds an overload set for a literal (ie. a defaulting
// overload set) comprised of the overloads to the type variable tv.
func (s *Solver) AddLiteralOverloads(tv *TypeVariable, overloads []Type) {
	// Get the type variable node associated with tv.
	tnode := s.typeVarNodes[tv.ID]

	// Set it to default (since literal overloads always default).
	tnode.Default = true

	// Add all the substitutions to the type variable node.
	for _, overload := range overloads {
		s.addSubstitution(tnode, basicSubstitution{typ: overload})
	}

	// Mark the type variable node as complete.
	s.completes[tv.ID] = struct{}{}
}

// AddOperatorOverloads binds an overload set for an operator application
// comprised of overloads to the type variable tv.
func (s *Solver) AddOperatorOverloads(tv *TypeVariable, overloads []Type, setOverload func(int)) {
	// Get the type variable node associated with tv.
	tnode := s.typeVarNodes[tv.ID]

	// Add all the substitutions to the type variable node.
	for i, overload := range overloads {
		s.addSubstitution(tnode, &operatorSubstitution{ndx: i, signature: overload, setOverload: setOverload})
	}

	// Mark the type variable node as complete.
	s.completes[tv.ID] = struct{}{}
}

// MustEqual asserts that two types are equivalent.
func (s *Solver) MustEqual(lhs, rhs Type, span *report.TextSpan) {
	// TODO
}

// MustCast asserts that src must be castable to dest.
func (s *Solver) MustCast(src, dest Type, span *report.TextSpan) {
	s.castAsserts = append(s.castAsserts, &castAssert{
		Src:  src,
		Dest: dest,
		Span: span,
	})
}

// Solve prompts the solver to make its finali type deductions based on all the
// constraints it has been given -- this assumes no more constraints will be
// provided.  This resets the solver when done.
func (s *Solver) Solve() {
	// Unify the first type substitution for any type variable nodes which should
	// default and have more than one remaining possible substitution.
	for _, tnode := range s.typeVarNodes {
		if tnode.Default && len(tnode.Nodes) > 1 {
			// We use `MustEqual` to perform the unification so we can avoid
			// rewriting all the boilerplate code inside `MustEqual` for top
			// level unification, but we pass in a `nil` position since
			// operation *should* never fail.
			s.MustEqual(tnode.Var, tnode.Nodes[0].Sub.Type(), nil)
		}
	}

	// As a final step of type deduction, apply all cast assertions.
	for _, ca := range s.castAsserts {
		if !s.unifyCast(ca) {
			s.error(ca.Span, "cannot cast %s to %s", ca.Src, ca.Dest)
		}
	}

	// Go through each type variable and make final deductions based on remaining
	// nodes in the solution graph.
	for _, tnode := range s.typeVarNodes {
		// Any remaining type variable which has exactly one substitution
		// associated with it is considered solved.
		if len(tnode.Nodes) == 1 {
			tnode.Var.Value = tnode.Nodes[0].Sub.Type()
			tnode.Nodes[0].Sub.Finalize()
		} else {
			// Otherwise, report an appropriate error.
			s.error(tnode.Var.Span, "unable to infer type for %s", tnode.Var.Name)
		}
	}

	// Reset the solver to its default state.
	s.typeVarNodes = nil
	s.subNodes = make(map[uint64]*subNode)
	s.subIDCounter = 0
	s.completes = make(map[uint64]struct{})
}

/* -------------------------------------------------------------------------- */

// error reports a compile error indicating a type solution failure.
func (s *Solver) error(span *report.TextSpan, message string, args ...any) {
	panic(&report.LocalCompileError{
		Message: fmt.Sprintf(message, args...),
		Span:    span,
	})
}
