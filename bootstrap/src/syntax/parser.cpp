#include "parser.hpp"

namespace chai {
    void Parser::parseFile() {
        // Read the first token from the input stream.
        next();

        
    }

    /* ---------------------------------------------------------------------- */

    void Parser::next() {
        m_lookbehind = std::move(m_tok);
        m_tok = m_lexer.nextToken();
    }

    bool Parser::has(TokenKind kind) {
        return m_tok.kind == kind;
    }

    void Parser::reject() {
        error(m_tok.span, "unexpected token: {}", m_tok.value);
    }

    Token Parser::want(TokenKind kind) {
        if (has(kind)) {
            next();
            return m_lookbehind;
        }

        reject();
    }
}