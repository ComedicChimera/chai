package mir

import "strings"

// Instruction represents a standalone statement that performs an operation such
// as a call.  The results of instructions can be stored into a bind or used in
// an assignment.
type Instruction struct {
	// OpCode is the integer code used to designate the instruction.
	OpCode int

	// Operands are the values that this instruction operates upon.
	Operands []Value
}

// Enumeration of instruction Op Codes
const (
	OpCall = iota
	OpRet

	// Intrinsic arithmetic
	OpNeg

	// Intrinsic Functions
	OpStrBytes
	OpStrLen
)

// IntrinsicTable is a table of all valid intrinsic names and their mappings to
// instructions.  All function instrinsics are prefixed with `__`.
var IntrinsicTable = map[string]int{
	"neg":        OpNeg,
	"__strbytes": OpStrBytes,
	"__strlen":   OpStrLen,
}

// displayTable converts an op code into a displayable string for the instruction.
var displayTable = []string{
	"call",     // OpCall
	"ret",      // OpRet
	"neg",      // OpNeg
	"strbytes", // OpStrBytes
	"strlen",   // OpStrLen
}

func (instr *Instruction) Repr(preindent string) string {
	sb := strings.Builder{}
	sb.WriteString(preindent)

	sb.WriteString(displayTable[instr.OpCode])

	sb.WriteRune(' ')
	for i, operand := range instr.Operands {
		sb.WriteString(operand.Repr())

		if i < len(instr.Operands)-1 {
			sb.WriteString(", ")
		}
	}

	sb.WriteRune(';')
	return sb.String()
}
