package depm

import (
	"chaic/common"
	"chaic/report"
	"chaic/types"
)

// ResolveOpaques resolves all opaque global symbol references in a file.  It
// returns false if not all opaque references could be resolved.
func ResolveOpaques(chFile *ChaiFile) bool {
	allResolved := true

	for name, orefs := range chFile.OpaqueRefs {
		if sym, ok := chFile.Parent.SymbolTable[name]; ok {
			if sym.DefKind == common.DefKindType {
				for _, oref := range orefs {
					oref.Value = sym.Type.(types.NamedType)
				}
			} else {
				for _, oref := range orefs {
					report.ReportCompileError(chFile.AbsPath, chFile.ReprPath, oref.Span, "%s is not a type", name)
				}

				allResolved = false
			}
		} else {
			for _, oref := range orefs {
				report.ReportCompileError(chFile.AbsPath, chFile.ReprPath, oref.Span, "undefined symbol: %s", name)
			}

			allResolved = false
		}
	}

	return allResolved
}
