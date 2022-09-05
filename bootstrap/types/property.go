package types

import "chaic/report"

// propertyConstraint is an assertion that a particular type has a property.
type propertyConstraint struct {
	// The name of the property.
	Name string

	// Whether the property must be mutable.
	Mutable bool

	// The span over which the property name occurs.
	Span *report.TextSpan

	// The type variable to store the property type into.
	PropTypeVar Type

	// The access kind performed to resolve the type constraint.  This is a
	// pointer to the corresponding field on the AST so it can be set by the
	// solver.
	AccessKind *PropertyAccessKind
}

// The kind of property access that resolves the constraint.
type PropertyAccessKind int

// Enumeration of property access kinds.
const (
	PropAccessStructField PropertyAccessKind = iota
)

// getProperty attempts to return the property of innerTyp named name which
// occurs over span.  If mutable is true, then the property must be mutable.
// This function assumes that typ is an inner type.
func (s *Solver) getProperty(innerTyp Type, name string, mutable bool, span *report.TextSpan, akind *PropertyAccessKind) Type {
	switch v := innerTyp.(type) {
	case *StructType:
		if field, ok := v.GetFieldByName(name); ok {
			// Struct fields are always mutable if the struct is mutable.
			*akind = PropAccessStructField
			return field.Type
		}
	case *PointerType:
		// Implement automatic dereferencing.
		return s.MustHaveProperty(v.ElemType, name, mutable, span, akind)
	}

	s.error(span, "%s has no property named %s", innerTyp.Repr(), name)
	return nil
}
