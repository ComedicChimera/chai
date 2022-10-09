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
	case *ast.StructDef:
		l.lowerStructDef(v)
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
		Signature: types.Simplify(fd.Symbol.Type).(*types.FuncType),
		ParamVars: util.Map(fd.Params, func(param *common.Symbol) *mir.Identifier {
			return &mir.Identifier{
				ExprBase: mir.NewExprBase(param.DefSpan),
				Symbol: &mir.MSymbol{
					Name:              param.Name,
					Type:              types.Simplify(param.Type),
					IsImplicitPointer: !types.IsPtrWrappedType(param.Type),
				},
			}
		}),
		SrcFileAbsPath: l.chfile.AbsPath,
		Span:           fd.Span(),
		Attrs:          make(map[mir.FuncAttrKind]string),
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

	// Set all the parameter MIR symbols.
	for i, param := range fd.Params {
		param.MIRSymbol = fn.ParamVars[i].Symbol
	}

	// Add the body to the deferred list if one exists.
	if fd.Body != nil {
		l.deferred = append(l.deferred, deferredPredicate{
			fn:   fn,
			pred: fd.Body,
		})
	}

	l.bundle.Functions = append(l.bundle.Functions, fn)
}

// lowerStructDef lowers a struct definition.
func (l *Lowerer) lowerStructDef(sd *ast.StructDef) {
	sst := types.Simplify(sd.Symbol.Type).(*types.StructType)
	msym := &mir.MSymbol{
		Name:              sd.Symbol.Name,
		Type:              sst,
		IsImplicitPointer: !types.IsPtrWrappedType(sst),
	}

	l.bundle.Structs = append(l.bundle.Structs, &mir.Struct{
		Type:           sst,
		SrcFileAbsPath: l.pkg.Files[sd.Symbol.FileNumber].AbsPath,
		Span:           sd.Span(),
		Symbol:         msym,
	})

	sd.Symbol.MIRSymbol = msym
}

/* -------------------------------------------------------------------------- */

// hasAnnot returns whether the annotation named name exists in annots.
func hasAnnot(annots map[string]ast.AnnotValue, name string) bool {
	_, ok := annots[name]
	return ok
}
