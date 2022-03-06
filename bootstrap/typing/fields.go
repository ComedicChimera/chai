package typing

import (
	"chai/report"
	"fmt"
)

// FieldConstraint is a constraint asserting that a value of a given type has a
// specific field.  These are only used for struct field accesses and implicit
// method calls.
type FieldConstraint struct {
	RootType DataType
	RootPos  *report.TextPosition

	FieldName string
	FieldPos  *report.TextPosition

	// FieldType is where the result of the field constraint is stored.
	FieldType DataType

	// DotKindPtr indicates the kind of field access that was performed.  It
	// stores a pointer back to a matching field on the AST.
	DotKindPtr *int
}

// Enumeration of dot kinds.
const (
	DKStructField = iota // Struct field
	DKStaticGet          // Package namespace access
)

// MustHaveField adds a new field constraint.
func (s *Solver) MustHaveField(fc *FieldConstraint) {
	s.fieldConstraints = append(s.fieldConstraints, fc)
}

// -----------------------------------------------------------------------------

// unifyFieldConstraint attempts to unify a field constraint with the already
// defined list of type constraints.  It returns two boolean values: one
// indicating whether or not there was sufficient data to resolve the field
// constraint (ie. should it's unification be deferred) and one indicating
// whether unifcation was successful if attempted.
func (s *Solver) unifyFieldConstraint(fc *FieldConstraint) (bool, bool) {
	// extract the root type
	var rootType DataType
	if tv, ok := fc.RootType.(*TypeVar); ok {
		if sub, ok := s.getSubstitution(tv.ID); ok {
			rootType = sub
		} else {
			// if we don't have a substitution, then we don't have enough
			// information to resolve the field constraint.
			return false, false
		}
	} else {
		rootType = InnerType(fc.RootType)
	}

	if st, ok := rootType.(*StructType); ok {
		if fieldNdx, ok := st.FieldsByName[fc.FieldName]; ok {
			field := st.Fields[fieldNdx]

			// check for visiblity
			if st.ParentID() != s.pkgID && !field.Public {
				report.ReportCompileError(
					s.ctx,
					fc.FieldPos,
					fmt.Sprintf("struct has no public field named `%s`", field.Name),
				)
				return true, false
			}

			// apply the final constraint
			if !s.unify(field.Type, fc.FieldType, fc.FieldPos) {
				return true, false
			}

			// mark it as a struct field access
			*fc.DotKindPtr = DKStructField

			return true, true
		}
	}

	// TODO: implicit method call

	report.ReportCompileError(
		s.ctx,
		fc.FieldPos,
		fmt.Sprintf("variable has no field or method named `%s`", fc.FieldName),
	)
	return true, false
}
