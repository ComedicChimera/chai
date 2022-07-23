#include "parser.hpp"

#include "types.hpp"

namespace chai {
    Type* Parser::parseTypeExt() {
        want(TokenKind::COLON);

        return parseTypeLabel();
    }

    Type* Parser::parseTypeLabel() {
        switch (m_tok.kind) {
        case TokenKind::I8:
            return new IntegerType(1, false);
        case TokenKind::U8:
            return new IntegerType(1, true);
        case TokenKind::I16:
            return new IntegerType(2, false);
        case TokenKind::U16:
            return new IntegerType(2, true);
        case TokenKind::I32:
            return new IntegerType(4, false);
        case TokenKind::U32:
            return new IntegerType(4, true);
        case TokenKind::I64:
            return new IntegerType(8, false);
        case TokenKind::U64:
            return new IntegerType(8, true);
        case TokenKind::F32:
            return new FloatingType(4);
        case TokenKind::F64:
            return new FloatingType(8);
        case TokenKind::BOOL:
            return new BoolType();
        case TokenKind::LPAREN:
            next();
            
            if (has(TokenKind::RPAREN)) {
                next();

                return new UnitType();
            } else {
                // TODO: tuples
            }
            break;
        case TokenKind::STAR:
            next();
            {
                bool isConst = false;
                if (has(TokenKind::CONST)) {
                    next();
                    isConst = true;
                }

                return new PointerType(parseTypeLabel(), isConst);
            }
        }

        reject();
        return nullptr;
    }

    /* ---------------------------------------------------------------------- */

    std::vector<Token> Parser::parseIdentList() {
        std::vector<Token> tokens;

        do {
            tokens.emplace_back(want(TokenKind::IDENTIFIER));
        } while (has(TokenKind::COMMA));

        return tokens;
    }
}