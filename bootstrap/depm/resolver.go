package depm

import (
	"chai/report"
	"chai/typing"
	"fmt"
)

// TODO: decide whether recursive type checking should occur in the resolver or
// through another mechanism.

// Resolver is responsible for resolving all global symbol dependencies: namely,
// those on imported symbols and globally-defined types.  This is run before
// type checking so local symbols are not processed until after resolution is
// completed.
type Resolver struct {
	pkgList []*ChaiPackage
}

// NewResolver creates a new resolver.
func NewResolver() *Resolver {
	return &Resolver{}
}

// Resolve runs the main resolution algorithm.
func (r *Resolver) Resolve() bool {
	if !r.checkImportCollisions() {
		return false
	}

	if !r.resolveImports() {
		return false
	}

	if !r.resolveNamedTypes() {
		return false
	}

	return true
}

// AddPackageList adds the globally determined package list to the resolver.
func (r *Resolver) AddPackageList(pkgList []*ChaiPackage) {
	r.pkgList = pkgList
}

// AddOpaqueTypeRef adds a reference to named, defined type.  This should be
// called by the parser as named types are used as type labels.  A new opaque
// type is returned to be used in place of the named type.
func (r *Resolver) AddOpaqueTypeRef(file *ChaiFile, name string, pos *report.TextPosition) *typing.OpaqueType {
	// TODO
	return nil
}

// -----------------------------------------------------------------------------

func (r *Resolver) resolveNamedTypes() bool {
	// TODO
	return false
}

// -----------------------------------------------------------------------------

// checkImportGlobalCollisions checks to make sure imported symbols don't
// collide with global symbols.  This has to be done after all the global
// symbols are already defined.
func (r *Resolver) checkImportCollisions() bool {
	for _, pkg := range r.pkgList {
		for _, file := range pkg.Files {
			for _, isym := range file.ImportedSymbols {
				if _, ok := pkg.SymbolTable[isym.Name]; ok {
					report.ReportCompileError(
						file.Context,
						isym.DefPosition,
						fmt.Sprintf("imported name `%s` collides with global name", isym.Name),
					)
				}
			}
		}
	}

	return report.ShouldProceed()
}

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
			importedOperators := make(map[int]*Operator)
			for opKind, op := range file.ImportedOperators {
				importedPkgPath := op.Pkg.Path()
				importPosition := op.Overloads[0].Position

				// first, check that the package defines an operator matching
				// the operator kind.
				if impOp, ok := op.Pkg.OperatorTable[opKind]; ok {
					// go through the list of overloads and copy over only the
					// public overloads
					var overloads []*OperatorOverload
					for _, overload := range impOp.Overloads {
						if overload.Public {
							// change the position to reflect that of the
							// imported operator (so that errors can be handled
							// appropriately)
							overloads = append(overloads, &OperatorOverload{
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
					importedOperators[opKind] = &Operator{
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
