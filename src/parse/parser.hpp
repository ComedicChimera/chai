#ifndef PARSER_H_INCLUDED
#define PARSER_H_INCLUDED

#include <optional>
#include <vector>

#include "tokenize/scanner.hpp"
#include "ast.hpp"
#include "depm/module.hpp"
#include "depm/srcfile.hpp"
#include "compile/compiler.hpp"

namespace chai {
    class Parser {
        Compiler* compiler;
        SrcFile& file;

        Scanner& sc;
        std::optional<Token> lookahead;

        // next gets the next token from the input stream
        Token next();

        // expect asserts that a given token occurs next in the token stream. It
        // throws an error if this assertion fails.  It returns the token it
        // reads in.
        Token expect(TokenKind);

        // reject throws an unexpected token error for a given token
        void reject(const Token&);

        // peek looks ahead one token without moving the parser's state forward
        Token peek();

        // Parsing functions
        bool parseMetadata();
        void parseImport();

        std::vector<Token> parseIdentList(TokenKind);

        // Symbol management functions
        bool globalCollides(const std::string&) const;

        // Error functions
        template<typename ...T>
        void throwError(const TextPosition&, const std::string&, T...);

        void throwSymbolAlreadyDefined(const TextPosition&, const std::string&);

    public:
        Parser(Compiler*, SrcFile&, Scanner&);

        // parse returns an ASTNode if the file should be used (ie. not marked
        // no-compile by metadata)
        std::optional<ASTRoot*> parse();
    };
}

#endif 