#ifndef _LEXER_H_
#define _LEXER_H_

#include <fstream>

#include "chaic.hpp"
#include "token.hpp"

namespace chai {
    // Lexer is responsible for tokenizing Chai source code.
    class Lexer {
        // The file stream the lexer is reading.
        std::ifstream& m_file;

        // The Chai source file the lexer is reading.
        ChaiFile *m_chFile;

        // The buffer used to build token values.
        std::string m_tokBuff;

        // The current line and column.
        size_t m_line, m_col;

        // The current start line and column of the current token.
        size_t m_startLine, m_startCol;

    public:
        Lexer(std::ifstream& file, ChaiFile *chFile)
        : m_file(file)
        , m_chFile(chFile)
        , m_line(0)
        , m_col(0)
        {}

        // nextToken reads the next token from the input stream.
        Token nextToken();

    private:
        // lexPunctuationOrOperator lexes a punctuation or operator symbol.
        Token lexPunctuationOrOperator();

        /* ------------------------------------------------------------------ */

        // lexKeywordOrIdent lexes a keyword or identifier.
        Token lexKeywordOrIdent();

        /* ------------------------------------------------------------------ */

        // lexNumeric lexes a numeric literal.
        Token lexNumeric();

        /* ------------------------------------------------------------------ */

        // readEscapeCode reads an escape code/escape sequence.
        void readEscapeCode();

        // lexRune lexes a rune literal.
        Token lexRune();

        // lexStdString lexes a standard string literal.
        Token lexStdString();

        // lexRawString lexes a raw string literal.
        Token lexRawString();

        /* ------------------------------------------------------------------ */

        // skipComment skips a line or block comment.  `c` is the lookahead.
        void skipComment(char c);

        /* ------------------------------------------------------------------ */

        // makeToken builds a token of `kind` based on the contents of the token
        // buffer with spans between the lexer's token start position and
        // current position.
        Token makeToken(TokenKind kind);

        // mark marks the beginning of a new token.
        void mark();

        // error throws a compile error based on the lexer's current position.
        void error(std::string&& message);

        /* ------------------------------------------------------------------ */

        // read moves the lexer forward one character and reads that character
        // into the token buffer.
        char read();

        // skip moves the lexer forward one character and discards it.
        char skip();

        // peek returns the next character in the input stream without moving
        // the lexer forward.
        char peek();

        // updatePosition updates the position of the lexer based on a read in
        // character.
        void updatePosition(char c);
    };
}


#endif