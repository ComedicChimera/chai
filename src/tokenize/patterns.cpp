#include "patterns.hpp"

#include <unordered_map>

namespace chai {
    std::unordered_map<std::string, TokenKind> keywordPatterns = { 
        { "let", TokenKind::Let },
        { "def", TokenKind::Def } 
    };

    std::optional<TokenKind> matchKeyword(const std::string& identifier) {
        if (keywordPatterns.contains(identifier))
            return keywordPatterns[identifier];

        return {};
    }

    std::unordered_map<std::string, TokenKind> symbolPatterns = {
        { "+", TokenKind::Plus },
        { "-", TokenKind::Minus },
        { "*", TokenKind::Star },
        { "/", TokenKind::Div },
        { "//", TokenKind::FloorDiv },
        { "%", TokenKind::Mod },
        { "**", TokenKind::Power }
    };

    std::optional<TokenKind> matchSymbol(const std::string& symbol) {
        if (symbolPatterns.contains(symbol))
            return symbolPatterns[symbol];

        return {};
    }
}
