#ifndef SYMBOL_TABLE_H_INCLUDED
#define SYMBOL_TABLE_H_INCLUDED

#include <unordered_map>
#include <string>
#include <optional>
#include <vector>
#include <unordered_set>

#include "symbol.hpp"
#include "util.hpp"

namespace chai {
    // SymbolTable represents the global symbol table for a package
    class SymbolTable {
        // parentID is the ID of the package whose symbol table this is
        u64 parentID;

        // symbols is the map symbols organized by name
        std::unordered_map<std::string, Symbol*> symbols;

        // unresolved is the set of all unresolved symbol names
        std::unordered_set<std::string> unresolved;

    public:
        // lookup attempts to find a symbol in the list of symbols.  If it
        // succeeds, it returns the matched symbol. If not, it adds the symbol
        // to the list of unresolved, creates a new entry for the symbol in the
        // global table and returns it.  It also takes an expected definition
        // kind for the symbol and, optionally, a mutability.  If the defined
        // symbol does not match these constraints, then nothing is returned.
        std::optional<Symbol*> lookup(const std::string&, DefKind, Mutability = Mutability::NeverMutated);

        // define takes a symbol and attempts to define it in the global
        // namespace. If it the symbol is unresolved, this action will resolve
        // it.  If the symbol is already defined, this action will fail.  If it
        // succeeds, it returns a pointer to the symbol in the table that MUST
        // be used as the acting pointer to this symbol -- the old pointer may
        // be deleted if it is used to update a preexisting symbol entry.
        std::optional<Symbol*> define(Symbol*);

        // getUnresolved gets a vector of the unresolved symbols.  This should
        // be called at the end of error resolution for the purposes of error
        // reporting.
        std::vector<Symbol*> getUnresolved();

        ~SymbolTable();
    };
}

#endif