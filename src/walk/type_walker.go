package walk

import (
	"chai/logging"
	"chai/sem"
	"chai/syntax"
	"chai/typing"
	"fmt"
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

		// TODO: the remaining value types
	case "ref_type":
		if elemType, ok := w.walkTypeLabelCore(branch.BranchAt(1)); ok {
			return &typing.RefType{
				ElemType: elemType,
			}, true
		}
	case "named_type":
		return w.walkNamedType(branch)
	}

	// unreachable
	return nil, false
}

// walkNamedType walks a `named_type` branch
func (w *Walker) walkNamedType(branch *syntax.ASTBranch) (typing.DataType, bool) {
	// TODO: check for cyclic types

	var rootName, accessedName string
	var rootPos, accessedPos *logging.TextPosition
	for _, item := range branch.Content {
		switch v := item.(type) {
		case *syntax.ASTBranch:
			switch v.Name {
			case "generic_spec":
				// TODO
			case "size_spec":
				// TODO
			}
		case *syntax.ASTLeaf:
			if v.Kind == syntax.IDENTIFIER {
				if rootName == "" {
					rootName = v.Value
					rootPos = v.Position()
				} else {
					accessedName = v.Value
					accessedPos = v.Position()
				}
			}
		}
	}

	// look the symbol up in the global scope
	var typeSym *sem.Symbol
	if accessedName == "" {
		if sym, ok := w.lookupGlobal(rootName); ok {
			typeSym = sym
		} else {
			w.logUndefined(rootName, rootPos)
			return nil, false
		}
	} else /* package lookup */ {
		if pkg, ok := w.SrcFile.VisiblePackages[rootName]; ok {
			if sym, ok := pkg.ImportSymbol(accessedName); ok {
				typeSym = sym
			} else {
				w.logError(
					fmt.Sprintf("package `%s` contains no public symbol named `%s`", rootName, accessedName),
					logging.LMKName,
					accessedPos,
				)

				return nil, false
			}
		} else {
			w.logError(
				fmt.Sprintf("no package visible by name `%s`", pkg.Name),
				logging.LMKName,
				rootPos,
			)

			return nil, false
		}
	}

	// check to make sure symbol is a valid type definition
	if typeSym.DefKind != sem.DefKindTypeDef {
		// TODO: handle constraint sets

		// determine the position of the name component of the named type
		var namePos *logging.TextPosition
		if accessedName == "" {
			namePos = rootPos
		} else {
			namePos = syntax.TextPositionOfSpan(branch.Content[0], branch.Content[2])
		}

		w.logError(
			fmt.Sprintf("`%s` is not a type", typeSym.Name),
			logging.LMKUsage,
			namePos,
		)

		return nil, false
	}

	// TODO: handle generics and size generics

	return typeSym.Type, true
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

	logging.LogFatal(fmt.Sprintf("missing required builtin type: `%s`", name))
	return nil
}
