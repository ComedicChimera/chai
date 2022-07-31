package types

import "chaic/report"

// TypeVariable represents a type variable: a undetermined type.
type TypeVariable struct {
	// The unique node ID of the type variable.
	ID uint64

	// The name of the type variable.
	Name string

	// The span to error on the type variable cannot be inferred.
	Span *report.TextSpan

	// The deduced concrete type for the type variable.
	Value Type

	// The parent solver to the type variable.
	parent *Solver
}

func (tv *TypeVariable) equals(other Type) bool {
	// If we reach here, we know the type variable is undetermined.
	report.ReportICE("equals() cannot be directly called on an undetermined type variable")
	return false
}

func (tv *TypeVariable) Size() int {
	if tv.Value == nil {
		report.ReportICE("Size() cannot be directly called on an undetermined type variable")
	}

	return tv.Value.Size()
}

func (tv *TypeVariable) Align() int {
	if tv.Value == nil {
		report.ReportICE("Align() cannot be directly called on an undetermined type variable")
	}

	return tv.Value.Align()
}

func (tv *TypeVariable) Repr() string {
	if tv.Value == nil {
		// TODO
	}

	return tv.Value.Repr()
}
