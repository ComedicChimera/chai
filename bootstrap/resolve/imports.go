package resolve

import (
	"chai/depm"
	"chai/report"
	"fmt"
)

// resolveImports resolves all imported symbols.
func (r *Resolver) resolveImports() bool {
	for _, pkg := range r.pkgList {
		for _, file := range pkg.Files {
			// first lookup imported symbols
			for _, isym := range file.ImportedSymbols {
				importedPkgPath := isym.Pkg.Path()

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
							fmt.Sprintf("symbol named `%s` is not public in package `%s`", sym.Name, importedPkgPath),
						)
					}
				} else {
					report.ReportCompileError(
						file.Context,
						// DefPosition here is position of the symbol import
						isym.DefPosition,
						fmt.Sprintf("no symbol named `%s` defined in package `%s`", isym.Name, importedPkgPath),
					)
				}
			}

			// then add operators: collisions checked later
			importedOperators := make(map[int]*depm.Operator)
			for opKind, op := range file.ImportedOperators {
				importedPkgPath := op.Pkg.Path()
				importPosition := op.Overloads[0].Position

				// first, check that the package defines an operator matching
				// the operator kind.
				if impOp, ok := op.Pkg.OperatorTable[opKind]; ok {
					// go through the list of overloads and copy over only the
					// public overloads
					var overloads []*depm.OperatorOverload
					for _, overload := range impOp.Overloads {
						if overload.Public {
							// change the position to reflect that of the
							// imported operator (so that errors can be handled
							// appropriately)
							overloads = append(overloads, &depm.OperatorOverload{
								Signature: overload.Signature,
								Context:   file.Context,
								Position:  importPosition,
								Public:    false,
							})
						}
					}

					// no overloads => no public operators to import
					if len(overloads) == 0 {
						report.ReportCompileError(
							file.Context,
							importPosition,
							fmt.Sprintf("no public overloads for operator `%s` defined in package `%s`", op.OpName, importedPkgPath),
						)

						return false
					}

					// operator exists and has public overloads => add it
					importedOperators[opKind] = &depm.Operator{
						Pkg:       op.Pkg,
						OpName:    op.OpName,
						Overloads: overloads,
					}
				} else {
					report.ReportCompileError(
						file.Context,
						// first overload is position of imported operator
						importPosition,
						fmt.Sprintf("no definitions for operator `%s` defined in package `%s`", op.OpName, importedPkgPath),
					)

					return false
				}

			}

			file.ImportedOperators = importedOperators
		}
	}

	return report.ShouldProceed()
}
