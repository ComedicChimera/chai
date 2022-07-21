#ifndef _PARSER_H_
#define _PARSER_H_

#include <fstream>

#include "chaic.hpp"
#include "lexer.hpp"
#include "ast.hpp"
#include "report.hpp"

namespace chai {
    // Parser is used to parse Chai code into an AST.
    class Parser {
        // The Chai source file being parsed.
        ChaiFile* m_chFile;

        // The lexer being used by the parser.
        Lexer m_lexer;

        // The parser's current token.
        Token m_tok;

        // The parser's previous token.
        Token m_lookbehind;

    public:
        // Parser creates a new parser using the input stream `file` for the
        // Chai source file `chFile`.
        Parser(std::ifstream& file, ChaiFile* chFile)
        : m_lexer(file, chFile)
        , m_chFile { chFile }
        {}

        // parseFile parses the file and stores the resulting AST into the file.
        void parseFile();

    private:
        // next moves the parser forward one token.
        void next();

        // has returns whether the token the parser is positioned over is of `kind`.
        bool has(TokenKind kind);

        // error throws a compile error over `span` with a formatted message.
        template<typename ...T>
        inline void error(const TextSpan& span, const std::string& foramt, T ...args) {
            reportError(CompileError(
                m_chFile->displayPath(),
                m_chFile->absPath(),
                std::format(format, args...),
                span,
            ))
        }

        // reject rejects the token the parser is currently positioned over.
        void reject();

        // want asserts that the parser has a token of `kind`.  If it does, the
        // parser is moved forward one token.  Otherwise, a compile error is
        // thrown.  The token the parser was positioned over is returned.
        Token want(TokenKind kind);
    };
}

#endif