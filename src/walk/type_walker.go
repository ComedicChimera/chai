package walk

import (
	"chai/logging"
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

// getBuiltinOverloads gets all the overloads based on a builtin name (eg.
// `Numeric`)
func (w *Walker) getBuiltinOverloads(name string) []typing.DataType {
	switch name {
	case "Integral":
		return []typing.DataType{
			typing.PrimType(typing.PrimKindI8),
			typing.PrimType(typing.PrimKindI16),
			typing.PrimType(typing.PrimKindI32),
			typing.PrimType(typing.PrimKindI64),
			typing.PrimType(typing.PrimKindU8),
			typing.PrimType(typing.PrimKindU16),
			typing.PrimType(typing.PrimKindU32),
			typing.PrimType(typing.PrimKindU64),
		}
	case "Floating":
		return []typing.DataType{
			typing.PrimType(typing.PrimKindF32),
			typing.PrimType(typing.PrimKindF64),
		}
	case "Numeric":
		return []typing.DataType{
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
		}
	}

	logging.LogFatal("unknown builtin overload set: " + name)
	return nil
}

func (w *Walker) lookupNamedBuiltin(name string) typing.DataType {
	// TODO: replace with actual look up logic
	switch name {
	case "int":
		return typing.PrimType(typing.PrimKindI32)
	case "uint":
		return typing.PrimType(typing.PrimKindU32)
	case "byte":
		return typing.PrimType(typing.PrimKindU8)
	}

	return nil
}
