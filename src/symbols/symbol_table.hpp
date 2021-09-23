#ifndef SYMBOL_TABLE_H_INCLUDED
#define SYMBOL_TABLE_H_INCLUDED

#include <unordered_map>
#include <string>
#include <optional>
#include <vector>

#include "symbol.hpp"
#include "util.hpp"

namespace chai {
    // SymbolTable represents the global symbol table for a package
    class SymbolTable {
        // parentID is the ID of the package whose symbol table this is
        u64 parentID;

        // symbols is the map symbols organized by name
        std::unordered_map<std::string, Symbol*> symbols;

        // unresolved is the set of all unresolved symbol names paired with
        // position of their first usage
        std::unordered_map<std::string, TextPosition> unresolved;

    public:
        // lookup attempts to find a symbol in the list of symbols.  If it
        // succeeds, it returns the matched symbol. If not, it adds the symbol
        // to the list of unresolved, creates a new entry for the symbol in the
        // global table and returns it.  It also takes an expected definition
        // kind for the symbol and, optionally, a mutability.  If the defined
        // symbol does not match these constraints, then nothing is returned.
        std::optional<Symbol*> lookup(const std::string&, const TextPosition&, DefKind, Mutability = Mutability::NeverMutated);

        // define takes a symbol and attempts to define it in the global
        // namespace. If it the symbol is unresolved, this action will resolve
        // it.  If the symbol is already defined, this action will fail.  If it
        // succeeds, it returns a pointer to the symbol in the table that MUST
        // be used as the acting pointer to this symbol -- the old pointer may
        // be deleted if it is used to update a preexisting symbol entry.
        std::optional<Symbol*> define(Symbol*);

        // collidesWith tests if a given symbol name collides with an already defined symbol
        inline bool collidesWith(const std::string& name) const { return symbols.contains(name); }

        // getUnresolved gets a map of the unresolved symbol names with their
        // first usage position.  This should be called at the end of error
        // resolution for the purposes of error reporting.
        inline const std::unordered_map<std::string, TextPosition>& getUnresolved() const { return unresolved; };

        ~SymbolTable();
    };
}

#endif