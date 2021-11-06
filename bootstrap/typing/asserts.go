package typing

import (
	"chai/report"
	"fmt"
)

// typeAssert represents a condition that must be met in order for the program
// to be considered well typed that a) doesn't influence type deduction in a
// meaningful way and b) must be applied after all types are determined. An
// example of such an assertion would be a cast assertion which asserts that a
// given type cast is valid.
type typeAssert interface {
	// Apply actually tests the assertion.  It returns if assertion suceeds.
	Apply() bool

	// FailMsg returns the error message to report if the assertion fails
	FailMsg() string

	// Position returns the position of the assertion for error reporting.
	Position() *report.TextPosition
}

// -----------------------------------------------------------------------------

// castAssert asserts the one type is castable to another.
type castAssert struct {
	src, dest DataType
	pos       *report.TextPosition
}

func (s *Solver) AssertCast(src, dest DataType, pos *report.TextPosition) {
	s.asserts = append(s.asserts, &castAssert{
		src:  src,
		dest: dest,
		pos:  pos,
	})
}

func (ca *castAssert) Apply() bool {
	destInnerType := InnerType(ca.dest)

	switch v := InnerType(ca.src).(type) {
	case PrimType:
		if dpt, ok := destInnerType.(PrimType); ok {
			// all numbers can cast between each other
			if v < PrimBool && dpt < PrimBool {
				return true
			}

			// bool to int
			if v == PrimBool && dpt < PrimBool {
				return true
			}

			// rune to string
			if v == PrimU32 && dpt == PrimString {
				return true
			}

			// invalid cast
			return false
		}
	case *RefType:
		if _, ok := destInnerType.(*RefType); ok {
			// TEMPORARILY: all reference types are castable between each other
			return true
		}
	}

	// mismatched types or type that can't be cast to anything
	return false
}

func (ca *castAssert) FailMsg() string {
	return fmt.Sprintf("cannot cast `%s` to `%s`", ca.src.Repr(), ca.dest.Repr())
}

func (ca *castAssert) Position() *report.TextPosition {
	return ca.pos
}
