package types

import (
	"chaic/report"
)

// Unify attempts find a substitution which makes the inputted types equal. If
// the type is concrete, then the only valid substitutions are the type itself.
// However, if the type is untyped or a type variable, then more than one
// substitution may be possible.
func Unify(lhs, rhs Type) bool {
	// Check for any special types in the RHS (since we type switch primarily
	// over the LHS).
	switch v := rhs.(type) {
	case *UntypedNull:
		return unifyUntypedNull(v, lhs)
	case *UntypedNumber:
		return unifyUntypedNumber(v, lhs)
	}

	switch v := lhs.(type) {
	case *UntypedNull:
		return unifyUntypedNull(v, rhs)
	case *UntypedNumber:
		return unifyUntypedNumber(v, rhs)
	case *PointerType:
		if rpt, ok := rhs.(*PointerType); ok {
			return v.Const == rpt.Const && Unify(v.ElemType, rpt.ElemType)
		}

		return false
	case *FuncType:
		if rft, ok := rhs.(*FuncType); ok {
			if len(v.ParamTypes) != len(rft.ParamTypes) {
				return false
			}

			for i, lparam := range v.ParamTypes {
				if !Unify(lparam, rft.ParamTypes[i]) {
					return false
				}
			}

			return Unify(v.ReturnType, rft.ReturnType)
		}

		return false
	case PrimitiveType:
		if rpt, ok := rhs.(PrimitiveType); ok {
			return v.equals(rpt)
		}

		return false
	}

	report.ReportICE("Unify() called with unimplemented types: %s, %s", lhs.Repr(), rhs.Repr())
	return false
}

// unifyUntypedNumber unified an untyped number with another type.
func unifyUntypedNumber(un *UntypedNumber, other Type) bool {
	// If the untyped number already has an inferred type, then we just unify
	// with that value: the untyped number is no longer untyped.
	if un.InferredType != nil {
		return Unify(un.InferredType, other)
	}

	if oun, ok := other.(*UntypedNumber); ok {
		// If the other untyped number already has an inferred type, then it
		// becomes the inferred type of primary untyped number if it is in the
		// set of valid types for that number.  Otherwise, we can't unify and
		// have to return false.
		if oun.InferredType != nil {
			for _, vtype := range un.ValidTypes {
				if Equals(vtype, oun.InferredType) {
					// Unify to make sure all type updates are performed.
					Unify(vtype, oun.InferredType)

					un.InferredType = oun.InferredType
					return true
				}
			}

			return false
		}

		// Intersect the valid types of the two untyped numbers.
		var commonTypes []PrimitiveType
		for _, typ := range un.ValidTypes {
			for _, otyp := range oun.ValidTypes {
				if typ.equals(otyp) {
					commonTypes = append(commonTypes, typ)
					break
				}
			}
		}

		switch len(commonTypes) {
		case 0:
			// No common types => can't unify.
			return false
		case 1:
			// Only one common type => that must be the inferred type.
			un.InferredType = commonTypes[0]
			oun.InferredType = commonTypes[0]
			return true
		default:
			// Otherwise, we just arbitrary make one of the numbers the inferred
			// type of the other and update the valid types of the number which
			// is now acting in place of both numbers.
			un.InferredType = oun
			oun.ValidTypes = commonTypes
			return true
		}

	} else {
		for _, vtype := range un.ValidTypes {
			if Equals(vtype, other) {
				// Unify to make sure all type updates are performed.
				Unify(vtype, other)

				un.InferredType = other
				return true
			}
		}

		return false
	}
}

// unifyUntypedNull unifies an untyped null value with another type.
func unifyUntypedNull(un *UntypedNull, other Type) bool {
	// If the untyped null already has an inferred type, then we just unify with
	// that type: it is the true type of the untyped null.
	if un.InferredType != nil {
		return Unify(un.InferredType, other)
	}

	// Otherwise, we make the inferred type of null the other type.  This is
	// always valid since the other type is always at least as specific as an
	// untyped null so no typing information can ever be lost (untyped null is
	// among the least specific possible types).
	un.InferredType = other
	return true
}
