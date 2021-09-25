#include "symbol_table.hpp"

namespace chai {
    std::optional<Symbol*> SymbolTable::lookup(const std::string& name, const TextPosition& pos, DefKind kind, bool pub, Mutability m) {
        auto newSymbol = new Symbol {
            .name = name, 
            .parentID = parentID,
            .isPublic = pub,
            .mutability = m,
            .defKind = kind,     
        };
        
        if (symbols.contains(name)) {
            auto declaredSymbol = symbols[name];

            // check to make sure the symbols match: if they don't, we return
            // none and the lookup fails (suitable symbol can never be defined)
            if (symbolsMatch(declaredSymbol, newSymbol)) {
                // update the properties of the declared symbol if the lookup
                // provides more information about what the symbol should be
                if (m != Mutability::NeverMutated)
                    declaredSymbol->mutability = m;

                if (declaredSymbol->defKind == DefKind::Unknown)
                    declaredSymbol->defKind = kind;

                return declaredSymbol;
            }

            return {};
        }

        symbols.emplace(name, newSymbol);
        unresolved.emplace(name, pos);
        return newSymbol;
    }

    std::optional<Symbol*> SymbolTable::define(Symbol* sym) {
        if (symbols.contains(sym->name)) {
            if (unresolved.contains(sym->name)) {
                auto declaredSymbol = symbols[sym->name];

                // in order for a definition to succeed, the symbols must match
                if (!symbolsMatch(declaredSymbol, sym))
                    return {};

                *declaredSymbol = *sym;
                delete sym;
                return symbols[declaredSymbol->name];
            }

            // duplicate symbol
            return {};
        }

        symbols.emplace(sym->name, sym);
        return sym;
    }

    // symbolsMatch checks if a declared symbol can be replaced/defined by a new
    // symbol -- the definitions are suitable
    bool SymbolTable::symbolsMatch(Symbol* declared, Symbol* replacement) {
        // check for a matching definition kind
        if (declared->defKind != DefKind::Unknown && replacement->defKind != DefKind::Unknown && declared->defKind != replacement->defKind)
            return false;

        // check for matching mutability
        if (declared->mutability == Mutability::Immutable && replacement->mutability == Mutability::Mutable)
            return false;
        else if (declared->mutability == Mutability::Mutable && replacement->mutability == Mutability::Immutable)
            return false;

        // check to make sure the symbol is public if it being imported
        return !declared->isPublic && replacement->isPublic;
    }

    SymbolTable::~SymbolTable() {
        for (auto& pair : symbols) {
            delete pair.second;
        }
    }
}