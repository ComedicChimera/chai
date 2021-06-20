package typing

import (
	"chai/logging"
	"fmt"
)

// TypeVariable is a special data type that represents an unknown type to be
// determined by the solver.  It is somewhat of a placeholder, but it does
// eventually have a known value.
type TypeVariable struct {
	// ID is a unique value identifying this type variable
	ID int

	// EvalType is the type that this type variable has been evaluated to. This
	// field is `nil` until the type variable is determined
	EvalType DataType

	// DefaultType is the type this type variable will be substituted with if no
	// other constraints are placed on it.  This field can be `nil` indicating
	// that there is no default type.  This is primarily used for numeric
	// literals when no more specific type for them can be determined
	DefaultType DataType

	// Substitution is the type substitution currently applied to this data type.
	Substitution *TypeSubstitution

	// HandleUndetermined is called whenever a type value cannot be determined
	// for this type variable, but there was no other type error.
	HandleUndetermined func()
}

func (tv *TypeVariable) Repr() string {
	if tv.EvalType == nil {
		return "<unknown type>"
	}

	return tv.EvalType.Repr()
}

func (tv *TypeVariable) equals(other DataType) bool {
	logging.LogFatal("`equals` called directly on unevaluated type variable")
	return false
}

// TypeConstraint represents a single Hindley-Milner type constraint: a
// statement of some relation one type has to another.
type TypeConstraint struct {
	Lhs, Rhs DataType

	// kind must be one of the enumerated constraint kinds
	Kind int

	// Pos is the position of the element(s) that generated this constraint
	Pos *logging.TextPosition
}

// Enumeration of constraint kind
const (
	TCEquiv  = iota // LHS and RHS are equivalent
	TCCoerce        // RHS coerces to LHS
	TCCast          // RHS casts to LHS
)

// TypeSubstitution a "guess" the solver is making about one type to infer for a
// given variable and on what basis it is making that guess (coercion, casting,
// etc.)
type TypeSubstitution struct {
	// Values is the type currently being substituted
	Value DataType

	// ConsKind indicates the type of constraint that determined the value
	ConsKind int

	// IsLHS indicates whether this type was on the left (true) or right (false)
	// of the type constraint
	IsLHS bool
}

// -----------------------------------------------------------------------------

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

// AddConstraint adds a new type constraint to the solver
func (s *Solver) AddConstraint(lhs, rhs DataType, consKind int, pos *logging.TextPosition) {
	s.constraints = append(s.constraints, &TypeConstraint{
		Lhs:  lhs,
		Rhs:  rhs,
		Kind: consKind,
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
	// attempt initial unification of the type constraints
	for _, cons := range s.constraints {
		if !s.unify(cons.Lhs, cons.Rhs, cons.Kind) {
			s.logTypeError(cons.Lhs, cons.Rhs, cons.Kind, cons.Pos)
			return false
		}
	}

	// determine final values for all our unknown types
	allEvaluated := true
	for _, tvar := range s.vars {
		if tvar.Substitution.Value != nil {
			subVal := tvar.Substitution.Value

			// poly cons sets can never be used as raw data types; all other
			// types are acceptable
			if _, ok := subVal.(*ConstraintSet); !ok {
				tvar.EvalType = tvar.Substitution.Value
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

	// clear the solution context for the next solve
	s.vars = nil
	s.constraints = nil
	return true
}

// -----------------------------------------------------------------------------

// unify attempts to unify a pair of types based on a given type constraint
// between them.  This will apply substitutions as necessary and is the primary
// mechanism for Hindley-Milner type inference.  The boolean returned indicates
// whether or not unification was successful.
func (s *Solver) unify(lhs, rhs DataType, consKind int) bool {
	// check for type variables on the right before switching of the left
	if rhTypeVar, ok := rhs.(*TypeVariable); ok {
		// we can just apply the substitution the right type since in doing so
		// we will handle the cast where both left and right are tvars.
		return s.substitute(rhTypeVar, lhs, consKind, false)
	}

	// note: we now know right is not a type var or mono cons set
	switch v := lhs.(type) {
	case *TypeVariable:
		return s.substitute(v, rhs, consKind, true)
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
				if !s.unify(arg.Type, rarg.Type, TCEquiv) {
					return false
				}

				if arg.Name != "" && rarg.Name != "" && arg.Name != rarg.Name {
					return false
				}

				if arg.Indefinite != rarg.Indefinite || arg.Optional != rarg.Optional || arg.ByReference != rarg.ByReference {
					return false
				}
			}

			return s.unify(v.ReturnType, rft.ReturnType, TCEquiv)
		}
	default:
		switch consKind {
		case TCEquiv:
			return Equivalent(lhs, rhs)
		case TCCoerce:
			return CoerceTo(rhs, lhs)
		case TCCast:
			return CastTo(rhs, lhs)
		}
	}

	// if we reach here, we had a type error
	return false
}

// substitute attempts to apply a type substitution to a given type. The
// `subType` may be another type variable.  The flag `isLhs` refers to the type
// variable not the argument.  Note that the `subType` may still be valid, but
// not become the new substituted type -- eg. coercion for most general type.
// This function returns a boolean indicating success.
func (s *Solver) substitute(tv *TypeVariable, subType DataType, consKind int, isLhs bool) bool {
	if tv.Substitution == nil {
		tv.Substitution = &TypeSubstitution{
			Value:    subType,
			ConsKind: consKind,
			IsLHS:    isLhs,
		}

		return true
	}

	if tv.Substitution.IsLHS && s.unify(tv.Substitution.Value, subType, tv.Substitution.ConsKind) {
		// TODO: determine whether a new substitution should occur or not
		return true
	} else if s.unify(subType, tv.Substitution.Value, tv.Substitution.ConsKind) {
		// TODO: determine whether a new substitution should occur or not
		return true
	} else {
		return false
	}
}

// logTypeError logs a type error between two data types
func (s *Solver) logTypeError(lhs, rhs DataType, consKind int, pos *logging.TextPosition) {
	var msg string
	switch consKind {
	case TCEquiv:
		msg = fmt.Sprintf("type mismatch: %s v. %s", lhs.Repr(), rhs.Repr())
	case TCCoerce:
		msg = fmt.Sprintf("invalid coercion: %s to %s", rhs.Repr(), lhs.Repr())
	case TCCast:
		msg = fmt.Sprintf("invalid cast: %s to %s", rhs.Repr(), lhs.Repr())
	}

	logging.LogCompileError(
		s.lctx,
		msg,
		logging.LMKTyping,
		pos,
	)
}
