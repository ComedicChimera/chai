#ifndef PATTERNS_H_INCLUDED
#define PATTERNS_H_INCLUDED

#include <optional>

#include "token.hpp"

namespace chai {
    // matchKeyword takes in a given identifier string and checks if it matches
    // one of Chai's reserved keywords.  If it does, the matching keyword
    // TokenKind is returned.
    std::optional<TokenKind> matchKeyword(const std::string&);
    std::optional<TokenKind> matchSymbol(const std::string&);
}

#endif