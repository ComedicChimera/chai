#include "symbol_table.hpp"

namespace chai {
    std::optional<Symbol*> SymbolTable::lookup(const std::string& name, const TextPosition& pos, DefKind kind, Mutability m) {
        if (symbols.contains(name)) {
            auto matchingSymbol = symbols[name];

            if (matchingSymbol->defKind != kind)
                return {};

            if (m == Mutability::Immutable && matchingSymbol->mutability == Mutability::Mutable)
                return {};
            else if (m == Mutability::Mutable && matchingSymbol->mutability == Mutability::Immutable)
                return {};
            else
                return matchingSymbol;

            matchingSymbol->mutability = m;
            return matchingSymbol;
        }

        auto newSymbol = new Symbol {
            .name = name, 
            .parentID = parentID,
            .mutability = m,
            .defKind = kind
        };
        symbols.emplace(name, newSymbol);
        unresolved.emplace(name, pos);
        return newSymbol;
    }

    std::optional<Symbol*> SymbolTable::define(Symbol* sym) {
        if (symbols.contains(sym->name)) {
            if (unresolved.contains(sym->name)) {
                auto matchingSymbol = symbols[sym->name];
                *matchingSymbol = *sym;
                delete sym;
                return symbols[matchingSymbol->name];
            }

            // duplicate symbol
            return {};
        }

        symbols.emplace(sym->name, sym);
        return sym;
    }

    SymbolTable::~SymbolTable() {
        for (auto& pair : symbols) {
            delete pair.second;
        }
    }
}