#include "parser.hpp"

namespace chai {
    Parser::Parser(Compiler* c, SrcFile& file, Scanner& sc)
    : sc(sc)
    , file(file)
    , compiler(c)
    {}

    Token Parser::next() {
        if (lookahead) {
            auto l = lookahead;
            lookahead = {};
            return l.value();
        }

        return sc.scanNext();
    }

    Token Parser::expect(TokenKind kind) {
        auto tok = next();

        if (tok.kind != kind) {
            // end of file is substitutable for a newline
            if (kind == TokenKind::Newline && tok.kind == TokenKind::EndOfFile)
                return tok;

            reject(tok);
        }     

        return tok;    
    }

    void Parser::reject(const Token& tok) {
        if (tok.kind == TokenKind::EndOfFile)
            throwError(tok.position, "unexpected end of file");

        throwError(tok.position, "unexpected token: `{}`", tok.value);
    }

    Token Parser::peek() {
        if (!lookahead)
            lookahead = sc.scanNext();

        return lookahead.value();
    }

    // -------------------------------------------------------------------------- //

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

                    if (!parseMetadata())
                        return {};
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

    // -------------------------------------------------------------------------- //

    bool Parser::globalCollides(const std::string& name) const {
        return file.parent->globalTable.collidesWith(name) ||
            file.visiblePackages.contains(name) ||
            file.importedSymbols.contains(name);
    }
}