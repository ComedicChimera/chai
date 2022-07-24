package types

import "chaic/report"

// UntypedNumber is used to represent an untyped numeric constant.
type UntypedNumber struct {
	// The name used to represent this type in source code.
	DisplayName string

	// The list of types which are valid for this number.
	ValidTypes []PrimitiveType

	// The type inferred for this numeric constant.
	InferredType Type
}

func (un *UntypedNumber) equals(other Type) bool {
	if un.InferredType == nil {
		for _, typ := range un.ValidTypes {
			if typ.equals(other) {
				return true
			}
		}
	}

	return Equals(un.InferredType, other)
}

func (un *UntypedNumber) Size() int {
	if un.InferredType == nil {
		return un.ValidTypes[0].Size()
	}

	return un.InferredType.Size()
}

func (un *UntypedNumber) Align() int {
	if un.InferredType == nil {
		return un.ValidTypes[0].Align()
	}

	return un.InferredType.Align()
}

func (un *UntypedNumber) Repr() string {
	if un.InferredType == nil {
		return un.DisplayName
	}

	return un.InferredType.Repr()
}

// -----------------------------------------------------------------------------

// UntypedNull is used to represent an untyped null value.
type UntypedNull struct {
	// The source location where the null occurs.
	Span *report.TextSpan

	// The type that is inferred for null.
	InferredType Type
}

func (un *UntypedNull) equals(other Type) bool {
	if un.InferredType == nil {
		return true
	}

	return Equals(un.InferredType, other)
}

func (un *UntypedNull) Size() int {
	if un.InferredType == nil {
		report.ReportICE("Size() called directly on an undetermined null value")
	}

	return un.InferredType.Size()
}

func (un *UntypedNull) Align() int {
	if un.InferredType == nil {
		report.ReportICE("Align() called directly on an undetermined null value")
	}

	return un.InferredType.Align()
}

func (un *UntypedNull) Repr() string {
	if un.InferredType == nil {
		return "untyped null"
	}

	return un.InferredType.Repr()
}
