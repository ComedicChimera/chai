package depm

import (
	"chai/report"
	"chai/typing"
	"fmt"
)

// typeRef represents a collection of references to a defined type.
type typeRef struct {
	TypePtr   *typing.DataType
	Positions []*report.TextPosition
}

// Resolver is responsible for resolving all global symbol dependencies: namely,
// those on imported symbols and globally-defined types.  It also performs
// recursive type checking.  This is run before type checking so local symbols
// are not processed until after resolution is completed.
type Resolver struct {
	pkgList []*ChaiPackage

	// typeRefs is the table of type references organized by file.
	typeRefs map[*ChaiFile]map[string]typeRef
}

// NewResolver creates a new resolver.
func NewResolver() *Resolver {
	return &Resolver{
		typeRefs: make(map[*ChaiFile]map[string]typeRef),
	}
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

	return r.checkRecursiveTypes()
}

// AddPackageList adds the globally determined package list to the resolver.
func (r *Resolver) AddPackageList(pkgList []*ChaiPackage) {
	r.pkgList = pkgList
}

// AddOpaqueTypeRef adds a reference to named, defined type.  This should be
// called by the parser as named types are used as type labels.  A new opaque
// type is returned to be used in place of the named type.
func (r *Resolver) AddOpaqueTypeRef(file *ChaiFile, name string, pos *report.TextPosition) *typing.OpaqueType {
	// check to see if there is already a type reference in the table. If so,
	// just return an opaque type wrapping that pointer.
	if tRefSubTable, ok := r.typeRefs[file]; ok {
		if tref, ok := tRefSubTable[name]; ok {
			tref.Positions = append(tref.Positions, pos)

			return &typing.OpaqueType{
				Name:    name,
				TypePtr: tref.TypePtr,
			}
		}

		// create a new type pointer
		var dt typing.DataType = typing.NothingType()
		typePtr := &dt

		// add it to the preexisting sub table
		tRefSubTable[name] = typeRef{
			TypePtr:   typePtr,
			Positions: []*report.TextPosition{pos},
		}

		return &typing.OpaqueType{
			Name:    name,
			TypePtr: typePtr,
		}
	}

	// create a new type pointer
	var dt typing.DataType = typing.NothingType()
	typePtr := &dt

	// create a new table to store them and add the type ref to it
	r.typeRefs[file] = map[string]typeRef{
		name: {
			TypePtr:   typePtr,
			Positions: []*report.TextPosition{pos},
		},
	}

	return &typing.OpaqueType{
		Name:    name,
		TypePtr: typePtr,
	}
}

// -----------------------------------------------------------------------------

// resolveNamedTypes resolves all type references if possible and produces
// appropriate errors.
func (r *Resolver) resolveNamedTypes() bool {
	for chFile, fileTypeRefs := range r.typeRefs {
		for name, tref := range fileTypeRefs {
			if sym, ok := lookupType(chFile, name); ok {
				*tref.TypePtr = sym.Type
			} else {
				// report for every type reference
				for _, pos := range tref.Positions {
					report.ReportCompileError(
						chFile.Context,
						pos,
						fmt.Sprintf("undefined symbol: `%s`", name),
					)
				}

			}
		}
	}

	return report.ShouldProceed()
}

// lookupType looks up a type definition in a file.
func lookupType(chFile *ChaiFile, name string) (*Symbol, bool) {
	// lookup order: imports, global, universal

	if sym, ok := chFile.ImportedSymbols[name]; ok {
		return sym, true
	}

	if sym, ok := chFile.Parent.SymbolTable[name]; ok {
		return sym, true
	}

	// TODO: handle universal symbols

	return nil, false
}

// -----------------------------------------------------------------------------

// symbolNode is a node in the symbol graph used to check for recursive types.
type symbolNode struct {
	// Color is the color of the node in the three color algorithm.
	// white = 0, grey = 1, black = 2
	Color int

	Type    typing.DataType
	Context *report.CompilationContext
	Pos     *report.TextPosition
}

// checkRecursiveTypes checks for recursive type definitions.
func (r *Resolver) checkRecursiveTypes() bool {
	graph := r.buildSymbolGraph()

	// return if there is no symbol graph
	if len(graph) == 0 {
		return true
	}

	// get the first symbol in the graph
	pkgID := r.pkgList[0].ID
	var name string
	for _name := range graph[pkgID] {
		name = _name
		break
	}

	// use the "three color" algorithm to check for recursive types.
	return r.threeColorDFS(graph, pkgID, name)
}

// threeColorDFS runs the three color algorithm on the symbol graph.
func (r *Resolver) threeColorDFS(graph map[uint64]map[string]symbolNode, pkgID uint64, name string) bool {
	// node := graph[pkgID][name]

	return true
}

// buildSymbolGraph constructs a graph of all global types so that we can
// search for recursive types.
func (r *Resolver) buildSymbolGraph() map[uint64]map[string]symbolNode {
	graph := make(map[uint64]map[string]symbolNode)

	for _, pkg := range r.pkgList {
		for _, sym := range pkg.SymbolTable {
			if sym.DefKind == DKTypeDef {
				graph[pkg.ID][sym.Name] = symbolNode{
					Type:    sym.Type,
					Context: sym.File.Context,
					Pos:     sym.DefPosition,
				}
			}
		}
	}

	return graph
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
				importedPkgPath := isym.File.Parent.Path()

				if sym, ok := isym.File.Parent.SymbolTable[isym.Name]; ok {
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
