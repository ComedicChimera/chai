#include "parser.hpp"

#include "types.hpp"
#include "types/type_store.hpp"

namespace chai {
    Type* Parser::parseTypeExt() {
        want(TokenKind::COLON);

        return parseTypeLabel();
    }

    Type* Parser::parseTypeLabel() {
        switch (m_tok.kind) {
        case TokenKind::I8:
            return typeStore.i8Type();
        case TokenKind::U8:
            return typeStore.u8Type();
        case TokenKind::I16:
            return typeStore.i16Type();
        case TokenKind::U16:
            return typeStore.u16Type();
        case TokenKind::I32:
            return typeStore.i32Type();
        case TokenKind::U32:
            return typeStore.u32Type();
        case TokenKind::I64:
            return typeStore.i64Type();
        case TokenKind::U64:
            return typeStore.u64Type();
        case TokenKind::F32:
            return typeStore.f32Type();
        case TokenKind::F64:
            return typeStore.f64Type();
        case TokenKind::BOOL:
            return typeStore.boolType();
        case TokenKind::LPAREN:
            next();
            
            if (has(TokenKind::RPAREN)) {
                next();

                return typeStore.unitType();
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

                return typeStore.pointerType(parseTypeLabel(), isConst);
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