#ifndef PARSER_H_INCLUDED
#define PARSER_H_INCLUDED

#include <optional>

#include "tokenize/scanner.hpp"
#include "ast.hpp"
#include "depm/module.hpp"

namespace chai {
    class Parser {
        Module* parentMod;
        Scanner& sc;
        BuildProfile& globalProfile;

        // next gets the next token from the input stream
        inline Token next() { return sc.scanNext(); }

        // expect asserts that a given token occurs next in the token stream.
        // It throws an error if this assertion fails.
        void expect(TokenKind);

        // reject throws an unexpected token error for a given token
        void reject(const Token&);

    public:
        Parser(Module*, BuildProfile&, Scanner&);

        // parse returns an ASTNode if the file should be used (ie. not marked
        // no-compile by metadata)
        std::optional<ASTRoot*> parse();
    };
}

#endif 