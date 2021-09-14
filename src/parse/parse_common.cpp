#include "parser.hpp"

namespace chai {
    std::vector<Token> Parser::parseIdentList(TokenKind delimiter) {
        std::vector<Token> idents;

        while (peek().kind == TokenKind::Identifier) {
            idents.push_back(next());

            if (peek().kind == delimiter)
                next();
            else
                break;
        }

        return idents;
    }
}