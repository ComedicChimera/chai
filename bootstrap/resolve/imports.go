package resolve

import (
	"chai/report"
	"fmt"
)

// resolveImports resolves all imported symbols.
func (r *Resolver) resolveImports() bool {
	for _, mod := range r.depGraph {
		for _, pkg := range mod.Packages() {
			for _, file := range pkg.Files {
				// first lookup imported symbols
				for _, isym := range file.ImportedSymbols {
					if sym, ok := isym.Pkg.SymbolTable[isym.Name]; ok {
						// symbols must be public
						if sym.Public {
							// update imported symbol to mirror symbol from
							// other package
							*isym = *sym
						} else {
							report.ReportCompileError(
								file.Context,
								// DefPosition here is position of the symbol import
								isym.DefPosition,
								fmt.Sprintf("symbol named `%s` is not public in package `%s`", sym.Name, mod.Name+pkg.ModSubPath),
							)
						}
					} else {
						report.ReportCompileError(
							file.Context,
							// DefPosition here is position of the symbol import
							isym.DefPosition,
							fmt.Sprintf("no symbol named `%s` defined in package `%s`", sym.Name, mod.Name+pkg.ModSubPath),
						)
					}
				}

				// TODO: then add operators: collisions checked later
				for _, op := range file.ImportedOperators {
					// TODO
					_ = op
				}
			}
		}
	}

	return report.ShouldProceed()
}
