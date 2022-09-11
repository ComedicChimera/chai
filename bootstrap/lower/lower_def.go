package lower

import (
	"chaic/ast"
	"chaic/common"
	"chaic/mir"
	"chaic/report"
	"chaic/types"
	"chaic/util"
)

// lowerDef lowers an AST definition.
func (l *Lowerer) lowerDef(def ast.ASTNode) {
	switch v := def.(type) {
	case *ast.FuncDef:
		l.lowerFuncDef(v)
	case *ast.OperDef:
	case *ast.StructDef:
	default:
		report.ReportICE("lowering not implemented for definition")
	}
}

// lowerFuncDef lowers an AST function definition.
func (l *Lowerer) lowerFuncDef(fd *ast.FuncDef) {
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
				ExprBase: mir.NewExprBase(param.DefSpan),
				Symbol: &mir.MSymbol{
					Name:              param.Name,
					Type:              param.Type,
					IsImplicitPointer: !types.IsPtrWrappedType(param.Type),
				},
			}
		}),
		SrcFileAbsPath: l.chfile.AbsPath,
		Span:           fd.Span(),
		// TODO: public
	}

	// Create and bind the MIR symbol.
	msym := &mir.MSymbol{
		Name:              fn.Name,
		Type:              fn.Signature,
		IsImplicitPointer: false,
	}
	fn.Symbol = msym
	fd.Symbol.MIRSymbol = msym

	// Handle annotations.
	for annotName, annotValue := range fd.Annotations {
		switch annotName {
		case "abientry", "extern":
			fn.Attrs[mir.AttrKindExternal] = ""
			fn.Public = true
		case "callconv":
			fn.Attrs[mir.AttrKindCallConv] = annotValue.Value
		}
	}

	l.block = &fn.Body
	if bodyExpr, ok := fd.Body.(ast.ASTExpr); ok {
		l.appendStmt(l.lowerExpr(bodyExpr))
	} else {
		l.lowerBlock(fd.Body.(*ast.Block))
	}

	l.bundle.Functions = append(l.bundle.Functions, fn)
}

/* -------------------------------------------------------------------------- */

// hasAnnot returns whether the annotation named name exists in annots.
func hasAnnot(annots map[string]ast.AnnotValue, name string) bool {
	_, ok := annots[name]
	return ok
}