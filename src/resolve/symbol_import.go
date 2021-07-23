package resolve

import (
	"chai/logging"
	"chai/sem"
	"fmt"
)

// processSymbolImport attempts to resolve a symbol import directly -- ie. by
// simply importing it from the given package.  If it fails, it will add this
// symbol to the resolution queue.
func (r *Resolver) processSymbolImport(file *sem.ChaiFile, sym *sem.Symbol) bool {
	if sym.SrcPackage.ParentID == r.mod.ID {
		// TODO
		return false
	}

	// if the module of the imported symbol is in is not the current module,
	// then we know that that module has already been loaded and should be able
	// to force a symbol import to resolve
	return r.resolveSymbolImport(file, sym)
}

// resolveSymbolImport forces a symbol import to resolve and errors if it is
// unable to resolve it
func (r *Resolver) resolveSymbolImport(file *sem.ChaiFile, sym *sem.Symbol) bool {
	if importedSym, ok := sym.SrcPackage.ImportSymbol(sym.Name); ok {
		// update the sym in the local file's import table
		file.ImportedSymbols[sym.Name] = importedSym
		return true
	}

	logging.LogCompileError(
		file.LogContext,
		fmt.Sprintf("symbol `%s` not publicly visible in package `%s`", sym.Name, r.depGraph[sym.SrcPackage.ParentID].BuildPackagePathString(sym.SrcPackage)),
		logging.LMKImport,
		sym.Position,
	)
	return false
}
