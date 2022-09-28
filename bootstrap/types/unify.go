package types

import (
	"chaic/util"
	"fmt"
	"strings"
)

// Unify attempts to make two types equal.  It returns whether or not it was
// successful. This function may make any untyped that are passed into it
// concrete!
func Unify(lhs, rhs Type) bool {
	lhs = InnerType(lhs)
	rhs = InnerType(rhs)

	if run, ok := rhs.(*UntypedNull); ok {
		// Null is still not concrete so it just assumes the value of the LHS.
		// This always works because even if the LHS is not concrete the two
		// types are equivalent so `InnerType` handles nulls recursively.
		run.Value = lhs
		return true
	} else if run, ok := rhs.(*UntypedNumber); ok {
		// Check for two untyped numbers unified against each other.
		if lun, ok := lhs.(*UntypedNumber); ok {
			return unifyUntypedNumbers(lun, run)
		} else if _, ok := lhs.(*UntypedNull); !ok {
			// Handle a concrete LHS.
			return unifyUntypedNumberWithConcrete(run, lhs)
		}
	}

	switch v := lhs.(type) {
	case *UntypedNumber:
		// We know the RHS is concrete so we can just do concrete unification.
		return unifyUntypedNumberWithConcrete(v, rhs)
	case *UntypedNull:
		// Null is still not concrete so it just assumes the value of the RHS.
		v.Value = rhs
		return true
	case *PointerType:
		if rpt, ok := rhs.(*PointerType); ok {
			return v.Const == rpt.Const && Unify(v.ElemType, rpt.ElemType)
		}
	case *TupleType:
		if rtt, ok := rhs.(*TupleType); ok {
			if len(v.ElementTypes) != len(rtt.ElementTypes) {
				return false
			}

			for i, elemType := range v.ElementTypes {
				if !Unify(elemType, rtt.ElementTypes[i]) {
					return false
				}
			}

			return true
		}
	case *FuncType:
		if rft, ok := rhs.(*FuncType); ok {
			if len(v.ParamTypes) != len(rft.ParamTypes) {
				return false
			}

			for i, paramType := range v.ParamTypes {
				if !Unify(paramType, rft.ParamTypes[i]) {
					return false
				}
			}

			return Unify(v.ReturnType, rft.ReturnType)
		}
	default:
		return lhs.equals(rhs)
	}

	return false
}

// unifyUntypedNumbers tries unify two undetermined untyped numbers.
func unifyUntypedNumbers(a, b *UntypedNumber) bool {
	// If the two types have same display name, then we can skip combining
	// their possible types since they must be the same type.
	if a.DisplayName == b.DisplayName {
		// Make sure they aren't the same type :)
		if a != b {
			a.Value = b
		}

		return true
	} else {
		// Find all the possible types the two untyped numbers have in common.
		var commonTypes []PrimitiveType

		for _, apt := range a.PossibleTypes {
			for _, bpt := range b.PossibleTypes {
				if apt.equals(bpt) {
					commonTypes = append(commonTypes, apt)
					break
				}
			}
		}

		// If there are no common types, then unification fails.
		if len(commonTypes) == 0 {
			return false
		}

		// Otherwise, we make b the value a and update b's possible types to be
		// the common types of the two untyped numbers.  This ensures that the
		// two untypeds are held equivalent and that only the types that suit
		// both are possible.
		a.Value = b
		b.PossibleTypes = commonTypes

		// Update the display name of `b`.

		b.DisplayName = fmt.Sprintf(
			"untyped {%s}",
			strings.Join(
				util.Map(commonTypes, func(pt PrimitiveType) string {
					return pt.Repr()
				}),
				" | ",
			),
		)

		// Mixing succeeds!
		return true
	}
}

// unifyUntypedNumberWithConcrete attempts to make an untyped number into a
// concrete type by unifying it with aconcrete type.
func unifyUntypedNumberWithConcrete(un *UntypedNumber, concrete Type) bool {
	for _, pt := range un.PossibleTypes {
		if Equals(pt, concrete) {
			un.Value = pt
			return Unify(pt, concrete)
		}
	}

	return false
}
