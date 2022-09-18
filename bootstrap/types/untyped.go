package types

import "chaic/report"

// UntypedNull represents a null value which is untyped.
type UntypedNull struct {
	// The concrete type of the null.
	Value Type

	// The span over which the null occurs in source test.
	Span *report.TextSpan
}

func (*UntypedNull) equals(other Type) bool {
	return true
}

func (*UntypedNull) Size() int {
	report.ReportICE("Size() called on untyped null")
	return 0
}

func (*UntypedNull) Align() int {
	report.ReportICE("Align() called on untyped null")
	return 0
}

func (un *UntypedNull) Repr() string {
	if un.Value != nil {
		return un.Value.Repr()
	}

	return "untyped null"
}

/* -------------------------------------------------------------------------- */

// UntypedNumber represents an untyped number type: it is a type corresponding
// to a numeric literal that can be one of many numeric types.
type UntypedNumber struct {
	// The concrete type of the number.
	Value Type

	// The display name of the untyped number.
	DisplayName string

	// The set of possible types for this untyped number.
	PossibleTypes []PrimitiveType
}

func (un *UntypedNumber) equals(other Type) bool {
	for _, pt := range un.PossibleTypes {
		if Equals(pt, other) {
			return true
		}
	}

	return false
}

func (un *UntypedNumber) Size() int {
	report.ReportICE("Size() called directly on untyped number")
	return 0
}

func (un *UntypedNumber) Align() int {
	report.ReportICE("Align() called directly on untyped number")
	return 0
}

func (un *UntypedNumber) Repr() string {
	if un.Value != nil {
		return un.Value.Repr()
	}

	return un.DisplayName
}
