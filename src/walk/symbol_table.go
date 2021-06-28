package walk

import (
	"chai/logging"
	"chai/sem"
	"fmt"
)

// lookup looks up a symbol and returns it if it exists.
func (w *Walker) lookup(name string) (*sem.Symbol, bool) {
	// iterate through local scopes backwards to facilitate shadowing
	for i := len(w.exprContextStack) - 1; i > -1; i-- {
		if sym, ok := w.exprContextStack[i].Scope[name]; ok {
			return sym, true
		}
	}

	// import table and global table can't have conflicts so we can just look up
	// the two in either order -- next in priority after imported symbols
	if sym, ok := w.SrcFile.Parent.GlobalTable[name]; ok {
		return sym, true
	}

	if sym, ok := w.SrcFile.ImportedSymbols[name]; ok {
		return sym, true
	}

	return nil, false
}

// defineGlobal defines a symbol in the global symbol table if possible.  It
// returns false and throws an appropriate error if it can't
func (w *Walker) defineGlobal(sym *sem.Symbol) bool {
	if _, ok := w.lookup(sym.Name); ok {
		w.logError(
			fmt.Sprintf("symbol named `%s` already defined in the global scope", sym.Name),
			logging.LMKName,
			sym.Position,
		)

		return false
	}

	w.SrcFile.Parent.GlobalTable[sym.Name] = sym
	return true
}

// -----------------------------------------------------------------------------

// defineLocal defines a local symbol in the most local scope
func (w *Walker) defineLocal(sym *sem.Symbol) bool {
	if len(w.exprContextStack) == 0 {
		logging.LogFatal("attempted to declare local symbol with no local scope")
	}

	currScope := w.currExprContext().Scope
	if _, ok := currScope[sym.Name]; ok {
		w.logError(
			fmt.Sprintf("symbol named `%s` already defined in immediate local scope", sym.Name),
			logging.LMKName,
			sym.Position,
		)

		return false
	}

	currScope[sym.Name] = sym
	return true
}
