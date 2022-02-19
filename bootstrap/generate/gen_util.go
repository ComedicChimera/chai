package generate

import (
	"chai/typing"

	"github.com/llir/llvm/ir/value"
)

// assignTo generates a code snippet for assigning an RHS value into an LHS value.
func (g *Generator) assignTo(lhs, rhs value.Value, dt typing.DataType) {
	if isPtrType(dt) {
		// TODO: memcopy
	} else {
		g.block.NewStore(rhs, lhs)
	}
}

// returnFromFunc generates a code snippet for returning a value from a
// function.
func (g *Generator) returnFromFunc(val value.Value, dt typing.DataType) {
	if isPtrType(dt) {
		// TODO: memcpy

		g.block.NewRet(nil)
	} else {
		g.block.NewRet(val)
	}
}
