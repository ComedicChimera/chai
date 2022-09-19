package common

import "chaic/types"

// OperatorMethod represents an operator method.
type OperatorMethod struct {
	// The unique ID of the operator method.
	ID uint64

	// The signature of the operator method.
	Signature *types.FuncType
}

// The enumeration of intrinsic operator IDs.
const (
	OP_ID_IADD uint64 = iota
	OP_ID_ISUB
	OP_ID_IMUL
	OP_ID_SDIV
	OP_ID_UDIV
	OP_ID_SMOD
	OP_ID_UMOD
	OP_ID_FADD
	OP_ID_FSUB
	OP_ID_FMUL
	OP_ID_FDIV
	OP_ID_FMOD

	OP_ID_INEG
	OP_ID_FNEG

	OP_ID_BWAND
	OP_ID_BWOR
	OP_ID_BWXOR
	OP_ID_BWCOMPL
	OP_ID_BWSHL
	OP_ID_BWSHR

	OP_ID_EQ
	OP_ID_NEQ
	OP_ID_SLT
	OP_ID_ULT
	OP_ID_FLT
	OP_ID_SGT
	OP_ID_UGT
	OP_ID_FGT
	OP_ID_SLTEQ
	OP_ID_ULTEQ
	OP_ID_FLTEQ
	OP_ID_SGTEQ
	OP_ID_UGTEQ
	OP_ID_FGTEQ

	OP_ID_LAND
	OP_ID_LOR
	OP_ID_LNOT

	OP_ID_UNKNOWN // Operator ID not yet determined (used untypeds).

	OP_ID_END_RESERVED // End of reserved intrinsic operator IDs.
)
