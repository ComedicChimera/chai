package depm

import "log"

// Universe is the set of symbols and operators defined in every package. This
// also contains all of the definitions for intrinsics: all intrinsics are
// visible in all packages without being imported.
type Universe struct {
	// CorePkg is the core/prelude package.  This package is imported by default,
	// and it's public namespace is merged with all packages (except itself): you can
	// access symbols defined in its public namespace without using a `.`.
	CorePkg *ChaiPackage

	// PreludeSymbolImports stores a list of all the symbols imported by default
	// in the prelude.  Note that we don't need to include the package since the
	// individual symbols always store that field.
	PreludeSymbolImports map[string]*Symbol

	// PreludeOperatorImports stores a list of all the operators imported by
	// default in the prelude.  Note that we don't need to include the package
	// since the individual overloads store that field, and an operator
	// definition may be comprised of operator overloads from many packages.
	PreludeOperatorImports map[int]*Operator
}

// NewUniverse creates a new universe for a project.
func NewUniverse() *Universe {
	return &Universe{
		PreludeSymbolImports:   make(map[string]*Symbol),
		PreludeOperatorImports: make(map[int]*Operator),
	}
}

// GetSymbol attempts to get a symbol with a specific name from the universe.
func (u *Universe) GetSymbol(name string) (*Symbol, bool) {
	// start by looking up in the core package
	if sym, ok := u.CorePkg.SymbolTable[name]; ok && sym.Public {
		return sym, true
	}

	// then, check the table of symbol imports
	if sym, ok := u.PreludeSymbolImports[name]; ok {
		return sym, true
	}

	return nil, false
}

// GetOperator attempts to get an operator with a specific kind from the
// universe.
func (u *Universe) GetOperator(kind int) (*Operator, bool) {
	// start by looking up in the core package.
	if op, ok := u.CorePkg.OperatorTable[kind]; ok {
		// NOTE: technically, we should only return the public operator
		// definitions, but all operators defined in core are public anyway --
		// it is just less hassle to do it this way.  Maybe when I am feeling
		// less lazy I will revise this...
		return op, true
	}

	// then, check the table of operator imports.
	if op, ok := u.PreludeOperatorImports[kind]; ok {
		return op, true
	}

	return nil, false
}

// ReserveNames reserves a list of symbol names for use later (to prevent import
// conflicts between core packages).  NOTE: We may want to find a better way to
// address this issue, but this will have to work for now.
func (u *Universe) ReserveNames(names ...string) {
	for _, name := range names {
		u.PreludeSymbolImports[name] = nil
	}
}

// AddSymbolImports adds a list of imported symbols given by name from a given
// package to the prelude import table.
func (u *Universe) AddSymbolImports(pkg *ChaiPackage, names ...string) {
	for _, name := range names {
		if sym, ok := pkg.SymbolTable[name]; ok && sym.Public {
			u.PreludeSymbolImports[name] = sym
		} else {
			log.Fatalf("[prelude error] package `%s` has no public symbol named `%s`\n", pkg.Name, name)
		}
	}
}

// AddOperatorImports adds a list of imported operators given by kind from a
// given package to the prelude import table.
func (u *Universe) AddOperatorImports(pkg *ChaiPackage, kinds ...int) {
	for _, kind := range kinds {
		if op, ok := pkg.OperatorTable[kind]; ok {
			// we need to remove all the private overloads before we add them to
			// the import table. However, we do not mess with the position of
			// the operator (for now)
			var newOverloads []*OperatorOverload
			for _, overload := range op.Overloads {
				if overload.Public {
					newOverloads = append(newOverloads, overload)
				}
			}

			// if there are no public overloads, then we need to report an error.
			if len(newOverloads) == 0 {
				log.Fatalf("[prelude error] package `%s` has no public overload overloads for `%s`\n", pkg.Name, op.OpName)
			}

			// add the operator to the prelude import table.
			u.PreludeOperatorImports[kind] = &Operator{
				Pkg:       pkg,
				OpName:    op.OpName,
				Overloads: newOverloads,
			}
		} else {
			log.Fatalf("[prelude error] package `%s` has no operators defined with kind `%d`\n", pkg.Name, kind)
		}
	}
}
