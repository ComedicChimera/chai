package types

import (
	"chaic/report"
	"fmt"
	"strings"
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

	// The span of the current unification.
	currSpan *report.TextSpan

	// The set of type variable referenced during the current unification.
	usedTypeVars map[uint64]*typeVarNode
}

// NewSolver creates a new type solver.
func NewSolver() *Solver {
	return &Solver{
		subNodes:     make(map[uint64]*subNode),
		completes:    make(map[uint64]struct{}),
		usedTypeVars: make(map[uint64]*typeVarNode),
	}
}

// NewTypeVar creates a new type variable in the solution context.
func (s *Solver) NewTypeVar(name string, span *report.TextSpan) *TypeVariable {
	tv := &TypeVariable{
		ID:   uint64(len(s.typeVarNodes)),
		Name: fmt.Sprintf("{%s}", name),
		Span: span,
	}

	s.typeVarNodes = append(s.typeVarNodes, &typeVarNode{Var: tv, Nodes: make(map[uint64]*subNode)})

	return tv
}

// AddOverloads binds an overload set comprised of overalods to the type
// variable tv.
func (s *Solver) AddOverloads(tv *TypeVariable, overloads []Type) {
	// Get the type variable node associated with tv.
	tnode := s.typeVarNodes[tv.ID]

	// Indicate the type variable is known.
	tnode.Known = true

	// Add substitution nodes corresponding to the overloads.
	for _, overload := range overloads {
		s.addSubstitution(tnode, basicSubstitution{typ: overload})
	}

	// Mark the type variable node as complete.
	s.completes[tv.ID] = struct{}{}
}

// AddLiteralOverloads binds an overload set for a literal (ie. a defaulting
// overload set) comprised of overloads to the type variable tv.
func (s *Solver) AddLiteralOverloads(tv *TypeVariable, overloads []Type) {
	// Get the type variable node associated with tv.
	tnode := s.typeVarNodes[tv.ID]

	// Indicate the type variable is known.
	tnode.Known = true

	// Add all the substitutions to the type variable node.
	for _, overload := range overloads {
		// Add the substitution node's ID to the default order.
		tnode.DefaultOrder = append(
			tnode.DefaultOrder,
			s.addSubstitution(tnode, basicSubstitution{typ: overload}).ID,
		)
	}

	// Mark the type variable node as complete.
	s.completes[tv.ID] = struct{}{}
}

// AddOperatorOverloads binds an overload set for an operator application
// comprised of overloads to the type variable tv.
func (s *Solver) AddOperatorOverloads(tv *TypeVariable, overloads []Type, setOverload func(int)) {
	// Get the type variable node associated with tv.
	tnode := s.typeVarNodes[tv.ID]

	// Indicate the type variable is known.
	tnode.Known = true

	// Add all the substitutions to the type variable node.
	for i, overload := range overloads {
		s.addSubstitution(tnode, &operatorSubstitution{ndx: i, signature: overload, setOverload: setOverload})
	}

	// Mark the type variable node as complete.
	s.completes[tv.ID] = struct{}{}
}

// MustEqual asserts that two types are equivalent.
func (s *Solver) MustEqual(lhs, rhs Type, span *report.TextSpan) {
	defer func() {
		// Clear the set of used type variables.
		for id := range s.usedTypeVars {
			delete(s.usedTypeVars, id)
		}
	}()

	// Set the solver's current span.
	s.currSpan = span

	// Attempt to unify the two types.
	result := s.unify(nil, lhs, rhs)

	// Raise an error if unification fails.
	if !result.Unified {
		sb := strings.Builder{}
		sb.WriteString("type mismatch: ")

		s.buildTraceback(&sb, s.usedTypeVars)

		sb.WriteString("type ")
		sb.WriteString(lhs.Repr())
		sb.WriteString(" does not match type ")
		sb.WriteString(rhs.Repr())

		s.error(span, sb.String())
	}

	// Prune all nodes which the unification algorithm marked for pruning.
	for id, prune := range result.Visited {
		// Note that we need to make sure the `id` has not already been pruned
		// through its connection to another pruned node.
		if _, ok := s.subNodes[id]; ok && prune {
			s.pruneSubstitution(s.subNodes[id], make(map[uint64]struct{}))
		}
	}

	// For any type variable nodes which now have a definite substitution,
	// resolve all all property constraints on that variable which has not been
	// resolved.
	for _, tnode := range s.usedTypeVars {
		if len(tnode.Nodes) == 1 && len(tnode.Properties) > 0 {
			// Clear the property constraints of the type variable node: we
			// don't want to apply them multiple times.
			props := tnode.Properties
			tnode.Properties = nil

			// Apply each of the property constraints.
			for _, prop := range props {
				propType := s.MustHaveProperty(tnode.Nodes[0].Sub.Type(), prop.Name, prop.Mutable, prop.Span)
				s.MustEqual(prop.PropTypeVar, propType, prop.Span)
			}
		}
	}
}

// MustCast asserts that src must be castable to dest.
func (s *Solver) MustCast(src, dest Type, span *report.TextSpan) {
	s.castAsserts = append(s.castAsserts, &castAssert{
		Src:  src,
		Dest: dest,
		Span: span,
	})
}

// MustHaveProperty asserts that typ has a property named name which occurs over
// span. If mutable is true, then the property must be mutable.  It returns the
// type of the accessed property.
func (s *Solver) MustHaveProperty(typ Type, name string, mutable bool, span *report.TextSpan) Type {
	innerTyp := InnerType(typ)

	if tv, ok := innerTyp.(*TypeVariable); ok {
		tnode := s.typeVarNodes[tv.ID]

		if len(tnode.Nodes) == 1 {
			return s.MustHaveProperty(tnode.Nodes[0].Sub.Type(), name, mutable, span)
		}

		propTypeVar := s.NewTypeVar("", span)
		tnode.Properties = append(tnode.Properties, propertyConstraint{
			Name:        name,
			Mutable:     mutable,
			Span:        span,
			PropTypeVar: propTypeVar,
		})

		return propTypeVar
	} else {
		return s.getProperty(innerTyp, name, mutable, span)
	}
}

// Solve prompts the solver to make its finali type deductions based on all the
// constraints it has been given -- this assumes no more constraints will be
// provided.  This does NOT reset the solver when done.
func (s *Solver) Solve() {
	// Unify the first type substitution for any type variable nodes which should
	// default and have more than one remaining possible substitution.
	for _, tnode := range s.typeVarNodes {
		if len(tnode.DefaultOrder) > 0 && len(tnode.Nodes) > 1 {
			// We use `MustEqual` to perform the unification so we can avoid
			// rewriting all the boilerplate code inside `MustEqual` for top
			// level unification, but we pass in a `nil` position since
			// operation *should* never fail.
			s.MustEqual(tnode.Var, tnode.Default().Sub.Type(), nil)
		}
	}

	// Go through each type variable and make final deductions based on remaining
	// nodes in the solution graph.
	for _, tnode := range s.typeVarNodes {
		// Any remaining type variable which has exactly one substitution
		// associated with it is considered solved.
		if len(tnode.Nodes) == 1 {
			tnode.Var.Value = tnode.First().Sub.Type()
			tnode.First().Sub.Finalize()
		} else {
			// Otherwise, report an appropriate error.
			s.error(tnode.Var.Span, "unable to infer type for %s", tnode.Var.Name)
		}
	}

	// Apply all cast assertions once deductions are made.
	for _, ca := range s.castAsserts {
		if !ca.tryCast() {
			s.error(ca.Span, "cannot cast %s to %s", ca.Src, ca.Dest)
		}
	}
}

// Reset resets the solver to its default state.
func (s *Solver) Reset() {
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
