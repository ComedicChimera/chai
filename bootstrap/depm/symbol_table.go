package depm

import (
	"chai/report"
	"fmt"
)

// symbolRef is a usage of a symbol in source code that is hasn't been resolved.
type symbolRef struct {
	pkgID          int
	context        *report.CompilationContext
	position       *report.TextPosition
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
// the symbol definition for purposes of error reporting.
func (st *SymbolTable) Define(sym *Symbol, ctx *report.CompilationContext) bool {
	// TODO
	return false
}

// Lookup attempts to retrieve a globally declared symbol matching the given
// characteristics. It may fail if it finds a partial match (ie. a symbol that
// matches in name but not other properties).  However, if it does not find a
// symbol matching in name, then it will NOT fail, but rather record the symbol
// reference to see if a matching symbol is defined later -- a declared-by-usage
// symbol is returned in this case.  It accepts the package ID of the package
// *accessing* the symbol not the package the symbol is defined in.
func (st *SymbolTable) Lookup(pkgID uint, ctx *report.CompilationContext, pos *report.TextPosition, defKind, mutability int) (*Symbol, bool) {
	// TODO
	return nil, false
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
			report.ReportCompileError(sref.context, sref.position, fmt.Sprintf("undefined symbol: `%s`", name))
		}
	}
}
