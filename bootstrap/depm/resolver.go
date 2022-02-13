package depm

import (
	"chai/report"
	"chai/typing"
	"fmt"
	"strings"
)

// typeRef represents a collection of references to a defined type.
type typeRef struct {
	TypePtr   *typing.DataType
	Positions []*report.TextPosition
}

// symbolNode is a node in the symbol graph.
type symbolNode struct {
	// Color is the color of the node in the three color algorithm.
	// white = 0, grey = 1, black = 2
	Color int

	Type    typing.DataType
	Context *report.CompilationContext
	Pos     *report.TextPosition
}

// Resolver is responsible for resolving all global symbol dependencies: namely,
// those on imported symbols and globally-defined types.  It also performs
// recursive type checking.  This is run before type checking so local symbols
// are not processed until after resolution is completed.
type Resolver struct {
	pkgList []*ChaiPackage

	// typeRefs is the table of type references organized by file.
	typeRefs map[*ChaiFile]map[string]typeRef

	// symGraph is used to check for recursive type definitions.
	symGraph map[uint64]map[string]*symbolNode
}

// NewResolver creates a new resolver.
func NewResolver() *Resolver {
	return &Resolver{
		typeRefs: make(map[*ChaiFile]map[string]typeRef),
		symGraph: make(map[uint64]map[string]*symbolNode),
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
// type is returned to be used in place of the named type.  If the type is
// accessed through a dot operation then the name should be passed in the form
// `pkg.Type`.
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
	// handle "dot" lookups
	if strings.ContainsRune(name, '.') {
		contents := strings.Split(name, ".")

		if pkg, ok := chFile.VisiblePackages[contents[0]]; ok {
			if sym, ok := pkg.SymbolTable[contents[1]]; ok && sym.Public {
				// mark the symbol is imported
				chFile.Parent.ImportedPackages[pkg.ID].Symbols[contents[1]] = &ChaiSymbolImport{
					Sym:      sym,
					Implicit: true,
				}

				return sym, true
			}
		}

		return nil, false
	}

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

// checkRecursiveTypes checks for recursive type definitions.
func (r *Resolver) checkRecursiveTypes() bool {
	r.buildSymbolGraph()

	// return if there is no symbol graph
	if len(r.symGraph) == 0 {
		return true
	}

	// get the first symbol in the graph
	pkgID := r.pkgList[0].ID
	var name string
	for _name := range r.symGraph[pkgID] {
		name = _name
		break
	}

	// use the "three color" algorithm to check for recursive types.
	return r.threeColorDFS(pkgID, name, nil)
}

// threeColorDFS runs the three color algorithm on the symbol graph.
// The parentPos is used to report errors based on the previous node in the
// graph rather than the definition itself: if there is a cycle, we mark the
// node that makes it not the node that begins it.
func (r *Resolver) threeColorDFS(pkgID uint64, name string, parentPos *report.TextPosition) bool {
	node := r.symGraph[pkgID][name]

	switch node.Color {
	case 0: // white => mark as grey, process node
		node.Color = 1

		ok := true
		switch v := node.Type.(type) {
		case *typing.AliasType:
			for _, nt := range getNamedTypes(v.Type) {
				ok = r.threeColorDFS(nt.ParentID(), nt.Name(), node.Pos) && ok
			}
		case *typing.StructType:
			for _, field := range v.Fields {
				for _, nt := range getNamedTypes(field.Type) {
					ok = r.threeColorDFS(nt.ParentID(), nt.Name(), node.Pos) && ok
				}
			}
		}

		return ok
	case 1: // grey => mark as black, report error
		node.Color = 2

		report.ReportCompileError(
			node.Context,
			// we report an error on the parent not the node
			parentPos,
			"illegal recursive type definition",
		)

		fallthrough // fail
	default: // black => do nothing, fail
		return false
	}
}

// buildSymbolGraph constructs the symbol graph.
func (r *Resolver) buildSymbolGraph() {
	for _, pkg := range r.pkgList {
		for _, sym := range pkg.SymbolTable {
			if sym.DefKind == DKTypeDef {
				r.symGraph[pkg.ID][sym.Name] = &symbolNode{
					Type:    sym.Type,
					Context: sym.File.Context,
					Pos:     sym.DefPosition,
				}
			}
		}
	}
}

// getNamedTypes extracts all the named types from a type.
func getNamedTypes(dt typing.DataType) []typing.NamedType {
	dt = typing.InnerType(dt)

	var nts []typing.NamedType

	switch v := dt.(type) {
	case typing.NamedType:
		nts = append(nts, v)
	case typing.TupleType:
		for _, tt := range v {
			nts = append(nts, getNamedTypes(tt)...)
		}
	case *typing.FuncType:
		for _, arg := range v.Args {
			nts = append(nts, getNamedTypes(arg)...)
		}

		nts = append(nts, getNamedTypes(v.ReturnType)...)
	case *typing.RefType:
		return getNamedTypes(v.ElemType)
	}

	return nts
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
