package walk

import (
	"chai/logging"
	"chai/sem"
	"chai/typing"
	"fmt"
)

// lookup looks up a symbol and returns it if it exists.
func (w *Walker) lookup(name string) (*sem.Symbol, bool) {
	// TODO: local table

	// local table and global table can't have conflicts so we can just look up
	// the two in either order -- next in priority after local symbols
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
func (w *Walker) defineGlobal(sym *sem.Symbol, pos *logging.TextPosition) bool {
	if _, ok := w.lookup(sym.Name); ok {
		w.logError(
			fmt.Sprintf("symbol named `%s` already defined in the global scope", sym.Name),
			logging.LMKName,
			pos,
		)

		return false
	}

	w.SrcFile.Parent.GlobalTable[sym.Name] = sym
	return true
}

// defineOperator attempts to define an operator in the global scope of a file
func (w *Walker) defineOperator(opCode int, sym *sem.Symbol, pos *logging.TextPosition) bool {
	if opSyms, ok := w.SrcFile.ImportedOperators[opCode]; ok {
		for _, opSym := range opSyms {
			if typing.Equivalent(sym.Type, opSym.Type) {
				w.logError(
					fmt.Sprintf("operator's signature conflicts with that of an operator imported from package `%s`", opSym.SrcPackage.Name),
					logging.LMKDef,
					pos,
				)

				return false
			}
		}
	}

	if opSyms, ok := w.SrcFile.Parent.GlobalOperators[opCode]; ok {
		for _, opSym := range opSyms {
			if typing.Equivalent(sym.Type, opSym.Type) {
				w.logError(
					"operator's signature conflicts with that of globally defined operator",
					logging.LMKDef,
					pos,
				)

				return false
			}
		}

		w.SrcFile.Parent.GlobalOperators[opCode] = append(opSyms, sym)
	} else {
		w.SrcFile.Parent.GlobalOperators[opCode] = []*sem.Symbol{sym}
	}

	return true
}
