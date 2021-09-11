#include "parser.hpp"

#include <format>

#include "report/message.hpp"

namespace chai {
    Parser::Parser(Module* m, BuildProfile& profile, Scanner& sc)
    : sc(sc)
    , parentMod(m)
    , globalProfile(profile)
    {}

    void Parser::expect(TokenKind kind) {
        auto tok = next();

        if (tok.kind != kind) {
            // end of file is substitutable for a newline
            if (kind == TokenKind::Newline && tok.kind == TokenKind::EndOfFile)
                return;

            reject(tok);
        }         
    }

    void Parser::reject(const Token& tok) {
        if (tok.kind == TokenKind::EndOfFile)
            throw CompileMessage("unexpected end of file", tok.position, parentMod->getErrorPath(sc.getFilePath()));

        throw CompileMessage(std::format("unexpected token: `{}`", tok.value), tok.position, parentMod->getErrorPath(sc.getFilePath()));
    }

    std::optional<ASTRoot*> Parser::parse() {
        ASTRoot* root = new ASTRoot;

        bool more = true;
        while (more) {
            auto curr = next();

            switch (curr.kind) {
                // metadata
                case TokenKind::Not:
                    // expecting `!!`
                    expect(TokenKind::Not);

                    // TODO
                    break;
                // import statement
                case TokenKind::Import:
                    // TODO
                    break;
                // leading newlines get skipped
                case TokenKind::Newline:
                    continue;
                // end of file breaks out of the loop
                case TokenKind::EndOfFile:
                    more = false;
                    break;
                // otherwise, it must be a top level declaration
                default:
                    // TODO

                    // top level will consume all else
                    more = false;
                    break;
            }   
        }             

        return root;
    }
}