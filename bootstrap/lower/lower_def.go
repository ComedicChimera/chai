package lower

import (
	"chaic/ast"
	"chaic/common"
	"chaic/depm"
	"chaic/mir"
	"chaic/report"
	"chaic/types"
	"chaic/util"
)

// lowerDef lowers an AST definition.
func (l *Lowerer) lowerDef(chfile *depm.ChaiFile, def ast.ASTNode) {
	switch v := def.(type) {
	case *ast.FuncDef:
		l.lowerFuncDef(chfile, v)
	case *ast.OperDef:
	case *ast.StructDef:
	default:
		report.ReportICE("lowering not implemented for definition")
	}
}

// lowerFuncDef lowers an AST function definition.
func (l *Lowerer) lowerFuncDef(chfile *depm.ChaiFile, fd *ast.FuncDef) {
	// Skip compilation for intrinsic functions.
	if hasAnnot(fd.Annotations, "intrinsic") {
		return
	}

	// Create the base MIR function.
	fn := &mir.Function{
		Name:      fd.Symbol.Name,
		Signature: fd.Symbol.Type.(*types.FuncType),
		ParamVars: util.Map(fd.Params, func(param *common.Symbol) *mir.Identifier {
			return &mir.Identifier{
				// No type simplification needed here.
				ValueBase:         mir.NewValueBase(param.DefSpan, param.Type),
				Name:              param.Name,
				IsImplicitPointer: !types.IsPtrWrappedType(param.Type),
			}
		}),
		SrcFileAbsPath: chfile.AbsPath,
		Span:           fd.Span(),
		// TODO: public
	}

	// Handle annotations.
	for annotName, annotValue := range fd.Annotations {
		switch annotName {
		case "abientry", "extern":
			fn.Attrs[mir.AttrKindPrototype] = ""
			fn.Public = true
		case "callconv":
			fn.Attrs[mir.AttrKindCallConv] = annotValue.Value
		}
	}

	// TODO: handle the body.

	l.bundle.Functions = append(l.bundle.Functions, fn)
}

/* -------------------------------------------------------------------------- */

// hasAnnot returns whether the annotation named name exists in annots.
func hasAnnot(annots map[string]ast.AnnotValue, name string) bool {
	_, ok := annots[name]
	return ok
}
