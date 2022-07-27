#include "parser.hpp"

namespace chai {
    void Parser::parseFile() {
        // Read the first token from the input stream.
        next();

        // Parse the package declaration.
        parsePackageDecl();

        // TODO: parse imports

        // Parse the actual content of the file.
        while (!has(TokenKind::END_OF_FILE)) {
            // Parse annotations.
            AnnotationMap annots;
            if (has(TokenKind::ATSIGN))
                annots = parseAnnotations();

            // TODO: public

            // Parse the definition.
            ASTNode* def;
            switch (m_tok.kind) {
            case TokenKind::FUNC:
                def = parseFuncDef(annots);
                break;
            default:
                reject();
                break;
            }

            // Add the definition to the file.
            m_chFile->definitions.emplace_back(std::unique_ptr<ASTNode>(def));
        }
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