package walk

import (
	"chai/syntax"
	"chai/typing"
)

// walkTypeExt walks a `type_ext` label and produces the labeled type
func (w *Walker) walkTypeExt(branch *syntax.ASTBranch) (typing.DataType, bool) {
	return w.walkTypeLabel(branch.BranchAt(1))
}

// walkTypeLabel walks a type label and produces the labeled type
func (w *Walker) walkTypeLabel(branch *syntax.ASTBranch) (typing.DataType, bool) {
	return w.walkTypeLabelCore(branch.BranchAt(0))
}

// walkTypeLabelCore walks the node internal to the `type` node
func (w *Walker) walkTypeLabelCore(branch *syntax.ASTBranch) (typing.DataType, bool) {
	switch branch.Name {
	case "value_type":
		valueTypeBranch := branch.BranchAt(0)
		switch valueTypeBranch.Name {
		case "prim_type":
			// the primitive kinds are enumerated identically to tokens -- we
			// can just subtract the `syntax.U8` starting token
			return typing.PrimType(branch.LeafAt(0).Kind - syntax.U8), true
		}
	}

	return nil, false
}
