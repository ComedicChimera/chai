package walk

import (
	"chaic/ast"
	"chaic/report"
	"chaic/syntax"
	"chaic/types"
	"fmt"
	"math"
	"strconv"
	"strings"
	"unicode/utf8"
)

// walkLiteral walks a literal AST node.
func (w *Walker) walkLiteral(lit *ast.Literal) {
	switch lit.Kind {
	case syntax.TOK_INTLIT:
		w.walkIntLit(lit)
	case syntax.TOK_FLOATLIT:
		w.walkFloatLit(lit)
	case syntax.TOK_NUMLIT:
		w.walkNumLit(lit)
	case syntax.TOK_RUNELIT:
		w.walkRuneLit(lit)
	case syntax.TOK_STRINGLIT:
		lit.Value = lit.Text
	case syntax.TOK_BOOLLIT:
		lit.Value = lit.Text == "true"
	}
}

// The list of int types in default preferential order.
var intTypes = []types.PrimitiveType{
	types.PrimTypeI64,
	types.PrimTypeU64,
	types.PrimTypeI32,
	types.PrimTypeU32,
	types.PrimTypeI16,
	types.PrimTypeU16,
	types.PrimTypeI8,
	types.PrimTypeU8,
}

// walkIntLit walks an integer literal.
func (w *Walker) walkIntLit(lit *ast.Literal) {
	// Make sure the integer literal is not absurdly large.
	if len(lit.Text) > 1000 {
		w.recError(lit.Span(), "literal is excessively long", lit.Text)
	}

	// Trim off the integer suffix and determine whether it is long/unsigned.
	var long, unsigned bool
	trimmedText := strings.TrimRight(lit.Text, "ul")

	switch len(lit.Text) - len(trimmedText) {
	case 2:
		long = true
		unsigned = true
	case 1:
		long = lit.Text[len(lit.Text)-1] == 'u'
		unsigned = lit.Text[len(lit.Text)-1] == 'l'
	}

	// Convert the integer to its Go value (Chai's integers outside of suffixes
	// have the exact same syntax as Go's so their should be no issues here).
	x, err := strconv.ParseUint(trimmedText, 0, 64)
	if err != nil {
		nerr := err.(*strconv.NumError)
		if nerr.Err == strconv.ErrRange {
			w.recError(lit.Span(), "literal `%s` is too large be represented by any integer type", lit.Text)
			return
		} else {
			report.ReportICE("unexpected Go error when parsing integer literal: %s", err)
		}
	}

	// Set the value of the literal.
	lit.Value = x

	// Determine the type or set of types that can represent the integer.
	var validTypes []types.PrimitiveType
	var constKindName string
	if long && unsigned {
		lit.NodeType = types.PrimTypeU64
		return
	} else if unsigned {
		validTypes = validIntTypes(x, types.PrimTypeU64, types.PrimTypeU32, types.PrimTypeU16, types.PrimTypeU8)
		constKindName = "unsigned int"
	} else if long {
		validTypes = validIntTypes(x, types.PrimTypeI64, types.PrimTypeU64)
		constKindName = "long int"
	} else {
		validTypes = validIntTypes(x, intTypes...)
		constKindName = "int"
	}

	lit.NodeType = &types.UntypedNumber{
		DisplayName: fmt.Sprintf("untyped %s literal", constKindName),
		ValidTypes:  validTypes,
	}
}

// validIntTypes returns the valid integer types of those passed in that can
// represent the given value.  The integer types must be in ascending order by
// size.
func validIntTypes(x uint64, intTypes ...types.PrimitiveType) []types.PrimitiveType {
	newIntTypes := make([]types.PrimitiveType, len(intTypes))

	n := 0
	for _, typ := range intTypes {
		if x < 1<<typ {
			newIntTypes[n] = typ
			n++
		}
	}

	return intTypes
}

// walkFloatLit walks a floating literal.
func (w *Walker) walkFloatLit(lit *ast.Literal) {
	// Convert the floating-point value.
	x, err := strconv.ParseFloat(lit.Text, 64)
	if err != nil {
		nerr := err.(*strconv.NumError)
		if nerr.Err == strconv.ErrRange {
			w.recError(
				lit.Span(),
				"literal `%s` is more than 1/2 ULP away from the maximum floating point value representable in 64 bits",
				lit.Text,
			)
		} else {
			report.ReportICE("unexpected Go error parsing float literal: %s", err)
		}
	}

	// Set the value of the literal.
	lit.Value = x

	// Test if it the literal can be represented by a 32-bit value.
	if x < math.SmallestNonzeroFloat32 || x > math.MaxFloat32 { // Too large/small.
		lit.NodeType = types.PrimTypeF64
	} else { // Can be f32.
		lit.NodeType = &types.UntypedNumber{
			DisplayName: "untyped float literal",
			ValidTypes:  []types.PrimitiveType{types.PrimTypeF64, types.PrimTypeF32},
		}
	}

}

// walkNumLit walks a number literal.
func (w *Walker) walkNumLit(lit *ast.Literal) {
	// First, attempt to treat the literal as an integer.
	x, err := strconv.ParseUint(lit.Text, 0, 64)
	if err != nil {
		nerr := err.(*strconv.NumError)
		if nerr.Err == strconv.ErrRange {
			// It is too large to be an integer: try as a float
			w.walkFloatLit(lit)
			return
		} else {
			report.ReportICE("unexpected Go error parsing number literal: %s", err)
		}
	}

	// The value will default to an integer.
	lit.Value = x

	// Get all the valid integer types for x.
	validTypes := validIntTypes(x, intTypes...)

	// Add the floating point types.
	// TODO: check if f32 is valid here?
	validTypes = append(validTypes, types.PrimTypeF64, types.PrimTypeF32)

	// Set the untyped number for the number literal.
	lit.NodeType = &types.UntypedNumber{
		DisplayName: "untyped number literal",
		ValidTypes:  validTypes,
	}
}

// walkRuneLit walks a rune literal.
func (w *Walker) walkRuneLit(lit *ast.Literal) {
	if len(lit.Text) > 1 {
		switch lit.Text[:2] {
		case "\\0":
			lit.Value = 0
		case "\\n":
			lit.Value = '\n'
		case "\\t":
			lit.Value = '\t'
		case "\\r":
			lit.Value = '\r'
		case "\\\\":
			lit.Value = '\\'
		case "\\\"":
			lit.Value = '"'
		case "\\'":
			lit.Value = '\''
		case "\\v":
			lit.Value = '\v'
		case "\\f":
			lit.Value = '\f'
		case "\\b":
			lit.Value = '\b'
		case "\\a":
			lit.Value = '\a'
		case "\\x", "\\u", "\\U":
			{
				r, err := strconv.ParseUint(lit.Text[2:], 16, 32)
				if err != nil {
					report.ReportICE("unexpected Go error parsing rune literal: %s", err)
				}

				lit.Value = r
			}
		default: // Multi-byte character.
			{
				r, _ := utf8.DecodeRuneInString(lit.Text)
				if r == utf8.RuneError {
					w.recError(lit.Span(), "invalid character in rune literal")
					return
				}

				lit.Value = uint64(r)
			}
		}
	} else {
		lit.Value = uint64(lit.Text[0])
	}
}
