package walk

import (
	"chaic/report"
	"chaic/types"
)

// mustUify asserts that two types must be unifiable.
func (w *Walker) mustUnify(expected, actual types.Type, span *report.TextSpan) {
	if !types.Unify(expected, actual) {
		w.error(span, "type mismatch: expected %s but got %s", expected.Repr(), actual.Repr())
	}
}

// mustCast asserts that one type must be castable to another.
func (w *Walker) mustCast(src, dest types.Type, span *report.TextSpan) {
	if !types.Cast(src, dest) {
		w.error(span, "cannot cast type %s to %s", src.Repr(), dest.Repr())
	}
}

/* -------------------------------------------------------------------------- */

// newUntypedNumber creates a new untyped number.
func (w *Walker) newUntypedNumber(displayName string, possibleTypes []types.PrimitiveType) *types.UntypedNumber {
	numType := &types.UntypedNumber{
		DisplayName:   displayName,
		PossibleTypes: possibleTypes,
	}

	w.untypedNumbers = append(w.untypedNumbers, numType)

	return numType
}

// newUntypedNull creates a new untyped null.
func (w *Walker) newUntypedNull(span *report.TextSpan) *types.UntypedNull {
	nullType := &types.UntypedNull{Span: span}

	w.untypedNulls = append(w.untypedNulls, nullType)

	return nullType
}
