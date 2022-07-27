#include "parser.hpp"

#include <unordered_set>

#include "types.hpp"
#include "types/type_store.hpp"

namespace chai {
    void Parser::parsePackageDecl() {
        want(TokenKind::PACKAGE);

        auto identTok = want(TokenKind::IDENTIFIER);

        // TODO: parse `.` package path.

        // Validate the package path.
        if (identTok.value != m_chFile->parent->name()) {
            error(identTok.span, "package declaration must match the directory structure");
        }

        want(TokenKind::SEMICOLON);
    }

    /* ---------------------------------------------------------------------- */

    AnnotationMap Parser::parseAnnotations() {
        want(TokenKind::ATSIGN);

        bool canHaveMultiple = false;
        if (has(TokenKind::LBRACKET)) {
            next();
            canHaveMultiple = true;
        }

        AnnotationMap annots;
        do {
            auto nameTok = want(TokenKind::IDENTIFIER);

            if (annots.find(nameTok.value) != annots.end())
                error(nameTok.span, "annotation {} specified multiple times", nameTok.value);

            if (has(TokenKind::LPAREN)) {
                next();

                auto valueTok = want(TokenKind::STRINGLIT);

                want(TokenKind::RPAREN);

                annots[nameTok.value] = std::make_unique<Annotation>(nameTok.value, nameTok.span, valueTok.value, valueTok.span);
            } else {
                annots[nameTok.value] = std::make_unique<Annotation>(nameTok.value, nameTok.span);
            }
        } while (canHaveMultiple && has(TokenKind::COMMA));

        if (canHaveMultiple)
            want(TokenKind::RBRACKET);

        return annots;
    }

    FuncDef* Parser::parseFuncDef(AnnotationMap& annots) {
        want(TokenKind::FUNC);

        auto nameTok = want(TokenKind::IDENTIFIER);

        want(TokenKind::LPAREN);

        std::vector<std::unique_ptr<Symbol>> params;
        if (!has(TokenKind::LPAREN)) {
            parseFuncParams(params);
        }

        want(TokenKind::RPAREN);

        Type* returnType = nullptr;
        ASTNode* funcBody = nullptr;
        do {
            switch (m_tok.kind) {
            case TokenKind::ASSIGN:
                // TODO: expr
                break;
            case TokenKind::LBRACE:
                // TODO: block
                break;
            case TokenKind::SEMICOLON:
                goto loop_exit;
            default:
                returnType = parseTypeLabel();
                break;
            }
        } while (funcBody == nullptr);

    loop_exit:
        // Set the default type if none is given.
        if (returnType == nullptr)
            returnType = typeStore.unitType();

        //auto* symbol = new Symbol(
        //    m_chFile,
        //    nameTok.value,
        //    // new FunctionType()
        //);
        
        return nullptr;
    }

    void Parser::parseFuncParams(std::vector<std::unique_ptr<Symbol>>& params) {
        std::unordered_set<std::string_view> paramNames;
        
        do {
            auto idents = parseIdentList();

            auto* typ = parseTypeExt();

            for (auto& ident : idents) {
                if (paramNames.find(ident.value) == paramNames.end()) {
                    paramNames.insert(ident.value);

                    params.emplace_back(std::make_unique<Symbol>(
                        m_chFile,
                        ident.value,
                        typ,
                        ident.span,
                        SymbolKind::VALUE,
                        false
                    ));
                } else
                    error(ident.span, "multiple parameters named {}", ident.value);
            }

        } while (has(TokenKind::COMMA));
    }
}