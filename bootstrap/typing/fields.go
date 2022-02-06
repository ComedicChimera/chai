package typing

import (
	"chai/report"
	"fmt"
)

// fieldConstraint is a constraint asserting that a value of a given type has a
// specific field.  These are only used for struct field accesses and implicit
// method calls.
type fieldConstraint struct {
	RootType DataType
	RootPos  *report.TextPosition

	FieldName string
	FieldPos  *report.TextPosition

	// FieldType is where the result of the field constraint is stored.
	FieldType DataType
}

// MustHaveField adds a new field constraint.
func (s *Solver) MustHaveField(rootType DataType, fieldName string, rootPos, fieldPos *report.TextPosition) {
	s.constraints = append(s.constraints, &fieldConstraint{
		RootType:  rootType,
		RootPos:   rootPos,
		FieldName: fieldName,
		FieldPos:  fieldPos,
	})
}

// -----------------------------------------------------------------------------

func (fc *fieldConstraint) Unify(s *Solver) bool {
	// if the root is an undetermined type variable, then we convert the field
	// constraint into a field bound.
	if tv, ok := fc.RootType.(*TypeVar); ok {
		if sub, ok := s.getSubstitution(tv.ID); ok {
			return s.resolveField(&fieldConstraint{
				RootType:  sub,
				RootPos:   fc.RootPos,
				FieldName: fc.FieldName,
				FieldPos:  fc.FieldPos,
				FieldType: fc.FieldType,
			})
		}

		if fbs, ok := tv.fieldBounds[fc.FieldName]; ok {
			tv.fieldBounds[fc.FieldName] = append(fbs, fieldBound{
				FieldPos:  fc.FieldPos,
				FieldType: fc.FieldType,
			})
		} else {
			tv.fieldBounds[fc.FieldName] = []fieldBound{
				{FieldPos: fc.FieldPos, FieldType: fc.FieldType},
			}
		}

		return true
	}

	return s.resolveField(fc)
}

// resolveField performs a field access and unifies its result into the solution state.
func (s *Solver) resolveField(fc *fieldConstraint) bool {
	rit := InnerType(fc.RootType)

	// handle structs
	if st, ok := rit.(*StructType); ok {
		if fieldNdx, ok := st.FieldsByName[fc.FieldName]; ok {
			field := st.Fields[fieldNdx]

			// check for visiblity
			if st.ParentID() != s.pkgID && !field.Public {
				report.ReportCompileError(
					s.ctx,
					fc.FieldPos,
					fmt.Sprintf("struct has no public field named `%s`", field.Name),
				)
				return false
			}

			// apply the final constraint
			s.unify(field.Type, fc.FieldType, fc.FieldPos)
		}
	}

	// TODO: implicit method call

	report.ReportCompileError(
		s.ctx,
		fc.FieldPos,
		fmt.Sprintf("variable has no field or method named `%s`", fc.FieldName),
	)
	return false
}
