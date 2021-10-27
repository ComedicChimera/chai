package depm

import (
	"chai/report"
	"fmt"
)

// symbolRef is a usage of a symbol in source code that is hasn't been resolved.
type symbolRef struct {
	pkgID          uint
	context        *report.CompilationContext
	position       *report.TextPosition
	assertsDefKind bool
	assertsMutable bool
}

// importConflict is used to represent an imported symbol that may conflict with
// other globally defined symbols.
type importConflict struct {
	context  *report.CompilationContext
	position *report.TextPosition
}

// SymbolTable is the global symbol table for a Chai package.  It is not only
// responsible for providing an index of defined symbols but also for resolving
// and handling undefined symbols as they are used.  Chai attempts to fully
// parse and symbolically analyze in one pass -- the table keeps track of the
// symbolic state of the program during this process.  Once package
// initialization is finished, this symbol table can be used to report all the
// undefined symbols.
type SymbolTable struct {
	// pkgID is the ID of the package this symbol table belongs to.
	pkgID uint

	// lookupTable is a map of all visible symbols OR all global symbols whose
	// definitions have been requested (so that all dependencies can share the
	// same symbol -- when it is resolved, it will be resolved for all
	// dependencies.
	lookupTable map[string]*Symbol

	// unresolved is a map of all references to unresolved symbols.  As symbols
	// are resolved, entries are removed from this table.  All entries remaining
	// in this table after package initialization are reported unresolved.
	unresolved map[string][]*symbolRef

	// resolutionFailed is a set of symbol names for which resolution failed.
	// This prevents symbols from being defined multiple times if their initial
	// resolution failed.
	resolutionFailed map[string]struct{}

	// importConficts is a map of imported symbols/package names associated with
	// their positions that should be errored upon if a conflicting global
	// symbol is defined.
	importConflicts map[string][]importConflict
}

// NewSymbolTable creates a new symbol table for the given package ID.
func NewSymbolTable(pkgid uint) *SymbolTable {
	return &SymbolTable{
		pkgID:            pkgid,
		lookupTable:      make(map[string]*Symbol),
		unresolved:       make(map[string][]*symbolRef),
		resolutionFailed: make(map[string]struct{}),
		importConflicts:  make(map[string][]importConflict),
	}
}

// -----------------------------------------------------------------------------

// Define defines a new global symbol and resolves all those symbols depending
// on it. This function returns false if the definition fails -- it will report
// errors in the case of failures.  It also requires the compilation context of
// the symbol definition for purposes of error reporting.  The returned symbol
// is the shared reference to the symbol that should be used as the definitive
// reference to this symbol henceforth.
func (st *SymbolTable) Define(ctx *report.CompilationContext, sym *Symbol) (*Symbol, bool) {
	// report import conflicts
	if _, ok := st.importConflicts[sym.Name]; ok {
		for _, ic := range st.importConflicts[sym.Name] {
			report.ReportCompileError(
				ic.context,
				ic.position,
				fmt.Sprintf("imported name `%s` conflicts with globally defined symbol", sym.Name),
			)
		}

		delete(st.importConflicts, sym.Name)
	}

	// if resolution has already been attempted and failed for this symbol then
	// trying to define it again is equivalent to a multi-definition
	if _, ok := st.resolutionFailed[sym.Name]; ok {
		report.ReportCompileError(
			ctx,
			sym.DefPosition,
			fmt.Sprintf("symbol defined multiple times: `%s`", sym.Name),
		)

		return nil, false
	} else if _, ok := st.unresolved[sym.Name]; ok {
		// if the symbol is unresolved, then its definition is ok and this
		// definition completes the previously defined symbol and removes
		// unresolved entries corresponding to the symbol.
		return st.resolve(sym)
	} else {
		// otherwise, just put the completed symbol in the lookup table
		st.lookupTable[sym.Name] = sym
		return sym, true
	}
}

// Lookup attempts to retrieve a globally declared symbol matching the given
// characteristics. It may fail if it finds a partial match (ie. a symbol that
// matches in name but not other properties).  However, if it does not find a
// symbol matching in name, then it will NOT fail, but rather record the symbol
// reference to see if a matching symbol is defined later -- a declared-by-usage
// symbol is returned in this case.  It accepts the package ID of the package
// *accessing* the symbol not the package the symbol is defined in.
func (st *SymbolTable) Lookup(pkgID uint, ctx *report.CompilationContext, name string, pos *report.TextPosition, defKind, mutability int) (*Symbol, bool) {
	// if the symbol is already defined, we have to match the lookup
	// parameters with those of the symbol to avoid conflicts.
	if declSym, ok := st.lookupTable[name]; ok {
		// first check that the mutability of the declared symbol does not
		// conflict with the lookup parameters -- ie. lookup isn't expecting a
		// mutable symbol.
		if declSym.Mutability == Immutable && mutability == Mutable {
			report.ReportCompileError(ctx, pos, "cannot mutate an immutable value")
			return nil, false
		}

		// then, we need to check if the symbol is unresolved or not
		if _, isUnresolved := st.unresolved[name]; isUnresolved {
			// check for matching definition kinds knowing that the symbol's
			// definition kind might not be well defined.
			if defKind != DKUnknown && declSym.DefKind != DKUnknown && declSym.DefKind != defKind {
				report.ReportCompileError(ctx, pos, fmt.Sprintf("cannot use %s as %s", reprDefKind(declSym.DefKind), reprDefKind(defKind)))
				return nil, false
			}

			// now, because the symbol has not yet been resolved, we can check
			// to see if we need to assert that it is public or not -- we set
			// the `Public` field to true if the symbol is used externally.
			if pkgID != st.pkgID {
				declSym.Public = true
			}

			// next we update the symbol's definition kind and mutability as
			// necessary
			if declSym.Mutability == NeverMutated && mutability != NeverMutated {
				declSym.Mutability = mutability
			}

			if declSym.DefKind == DKUnknown && defKind != DKUnknown {
				declSym.DefKind = defKind
			}

			// add the symbol reference to the list of unresolved
			st.unresolved[name] = append(st.unresolved[name], &symbolRef{
				pkgID:          pkgID,
				context:        ctx,
				position:       pos,
				assertsMutable: mutability == Mutable,
				assertsDefKind: defKind != DKUnknown,
			})
		} else {
			// check for matching definition kinds knowing that the symbol's
			// definition kind is well defined.
			if defKind != DKUnknown && declSym.DefKind != defKind {
				report.ReportCompileError(ctx, pos, fmt.Sprintf("cannot use %s as %s", reprDefKind(declSym.DefKind), reprDefKind(defKind)))
				return nil, false
			}

			// check to see if the symbol is being used externally and if so, is
			// it visible.
			if pkgID != st.pkgID && !declSym.Public {
				report.ReportCompileError(ctx, pos, fmt.Sprintf("symbol `%s` is not publically visible", name))
				return nil, false
			}
		}

		return declSym, true
	} else {
		// this is the first usage of this symbol
		declSym := &Symbol{
			Name:        name,
			PkgID:       pkgID,
			DefPosition: pos,
			DefKind:     defKind,
			Mutability:  mutability,
			Public:      pkgID != st.pkgID,
		}

		st.lookupTable[name] = declSym

		st.unresolved[declSym.Name] = []*symbolRef{{
			pkgID:          pkgID,
			context:        ctx,
			position:       pos,
			assertsMutable: mutability == Mutable,
			assertsDefKind: defKind != DKUnknown,
		}}

		return declSym, true
	}
}

// CheckConflict checks if an imported name conflicts with an already defined
// symbol. If it does, a conflict is reported.  Otherwise, it is added to the
// map of potential import conflicts.  This should be called whenever a name
// (package or symbol) is imported.
func (st *SymbolTable) CheckConflict(ctx *report.CompilationContext, name string, pos *report.TextPosition) {
	if _, ok := st.lookupTable[name]; ok {
		report.ReportCompileError(ctx, pos, fmt.Sprintf("imported name `%s` conflicts with globally defined symbol", name))
	} else if _, ok := st.importConflicts[name]; ok {
		st.importConflicts[name] = append(st.importConflicts[name], importConflict{
			context:  ctx,
			position: pos,
		})
	} else {
		st.importConflicts[name] = []importConflict{{context: ctx, position: pos}}
	}
}

// ReportUnresolved reports all remaining unresolved symbols as unresolved. This
// function should be called after all packages have been fully initialized.
func (st *SymbolTable) ReportUnresolved() {
	for name, srefs := range st.unresolved {
		for _, sref := range srefs {
			reportErrorOnSymRef(sref, "undefined symbol: `%s`", name)
		}
	}
}

// -----------------------------------------------------------------------------

// resolve attempts to resolve a declared-by-usage symbol with its actual
// definition. This function returns the declared symbol if resolution succeeds
// and false if it fails.
func (st *SymbolTable) resolve(sym *Symbol) (*Symbol, bool) {
	// we want to first check that the properties of the declared symbol match
	// up with those of the defined symbol.  If not, then we need to error on
	// all the unresolved usages appropriately.
	declSym := st.lookupTable[sym.Name]

	// if the declared symbol is marked as public, then it was required from
	// another package.  Thus, if the defined symbol is not public then we want
	// to error on all unresolved that are from another package and fail.
	if declSym.Public && !sym.Public {
		for _, unresolved := range st.unresolved[sym.Name] {
			if sym.PkgID != unresolved.pkgID {
				reportErrorOnSymRef(unresolved, "symbol `%s` is not publically visible", sym.Name)
			}
		}

		return nil, false
	}

	// if the definition kind of the globally declared symbol doesn't match the
	// definition kind of the symbol resolving it, then we have an error on all
	// the symbols that asserted the current definition kind since they no
	// longer match the defined kind.  However, if the definition kind of the
	// global symbol is unknown, then it doesn't matter.
	if declSym.DefKind != DKUnknown && declSym.DefKind != sym.DefKind {
		for _, unresolved := range st.unresolved[sym.Name] {
			if unresolved.assertsDefKind {
				reportErrorOnSymRef(
					unresolved,
					"cannot use %s as %s",
					reprDefKind(sym.DefKind),
					reprDefKind(declSym.DefKind),
				)
			}
		}
	}

	// mutability is only an issue if the declared symbol is marked as mutable,
	// but its definition requires to be immutable.
	if declSym.Mutability == Mutable && sym.Mutability == Immutable {
		for _, unresolved := range st.unresolved[sym.Name] {
			if unresolved.assertsMutable {
				reportErrorOnSymRef(unresolved, "cannot mutate an immutable value")
			}
		}
	}

	// we want to update the declared symbol's properties with those of its
	// actual definition.  However, we only want to update the mutability
	// if the defined symbol is not marked never mutated.
	if sym.Mutability != NeverMutated {
		*declSym = *sym
	} else {
		mut := declSym.Mutability
		*declSym = *sym
		declSym.Mutability = mut
	}

	// remove the unnecessary unresolved entries.
	delete(st.unresolved, sym.Name)

	// the declared symbol is still the shared symbol reference -- it supercedes
	// the symbol in the actual definition.
	return declSym, true
}

// reportErrorOnSymRef is a utility function to report an error related to a
// given symbol reference.
func reportErrorOnSymRef(sref *symbolRef, msg string, a ...interface{}) {
	report.ReportCompileError(
		sref.context,
		sref.position,
		fmt.Sprintf(msg, a...),
	)
}
