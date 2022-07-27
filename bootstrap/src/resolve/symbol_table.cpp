#include "chaic.hpp"

#include "report.hpp"

namespace chai {
    void ChaiPackage::define(Symbol* symbol) {
        if (m_symbolTable.find(symbol->name()) == m_symbolTable.end()) {
            m_symbolTable.emplace(std::make_pair(symbol->name(), std::unique_ptr<Symbol>(symbol)));
            return;
        }

        delete symbol;
        throw CompileError(
            symbol->parent->displayPath(),
            symbol->parent->absPath(),
            std::format("multiple symbols named %s defined in global scope", symbol->name()),
            symbol->defSpan()
        );
    }
}