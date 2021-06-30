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
			return typing.PrimType(valueTypeBranch.LeafAt(0).Kind - syntax.U8), true
		}
	}

	return nil, false
}

// -----------------------------------------------------------------------------

func (w *Walker) lookupNamedBuiltin(name string) typing.DataType {
	// TODO: replace with actual look up logic
	switch name {
	case "Integral":
		return &typing.ConstraintSet{
			Name:         "Integral",
			SrcPackageID: w.SrcFile.Parent.ID,
			Set: []typing.DataType{
				typing.PrimType(typing.PrimKindI8),
				typing.PrimType(typing.PrimKindI16),
				typing.PrimType(typing.PrimKindI32),
				typing.PrimType(typing.PrimKindI64),
				typing.PrimType(typing.PrimKindU8),
				typing.PrimType(typing.PrimKindU16),
				typing.PrimType(typing.PrimKindU32),
				typing.PrimType(typing.PrimKindU64),
			},
		}
	case "Floating":
		return &typing.ConstraintSet{
			Name:         "Floating",
			SrcPackageID: w.SrcFile.Parent.ID,
			Set: []typing.DataType{
				typing.PrimType(typing.PrimKindF32),
				typing.PrimType(typing.PrimKindF64),
			},
		}
	case "Numeric":
		return &typing.ConstraintSet{
			Name:         "Integral",
			SrcPackageID: w.SrcFile.Parent.ID,
			Set: []typing.DataType{
				typing.PrimType(typing.PrimKindI8),
				typing.PrimType(typing.PrimKindI16),
				typing.PrimType(typing.PrimKindI32),
				typing.PrimType(typing.PrimKindI64),
				typing.PrimType(typing.PrimKindU8),
				typing.PrimType(typing.PrimKindU16),
				typing.PrimType(typing.PrimKindU32),
				typing.PrimType(typing.PrimKindU64),
				typing.PrimType(typing.PrimKindF32),
				typing.PrimType(typing.PrimKindF64),
			},
		}
	case "int":
		return typing.PrimType(typing.PrimKindI32)
	case "uint":
		return typing.PrimType(typing.PrimKindU32)
	}

	return nil
}
