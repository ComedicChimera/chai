#ifndef SCANNER_H_INCLUDED
#define SCANNER_H_INCLUDED

#include <string>
#include <fstream>
#include <optional>

#include "token.hpp"

namespace chai {
    // Scanner is responsible for reading in and tokenizing user files.   It
    // works as a state machine: when called it produces a single new token
    // until there are no tokens remaining.
    class Scanner {
        std::ifstream file;
        int line, col;

        std::string tokBuff;

        void skipComment();

        Token makeToken(TokenKind);

        std::optional<char> readChar();
        bool skipChar();
        std::optional<char> peekChar();
        void updatePosition(char);

    public:
        Scanner(const std::string&);

        // scanNext reads a new token from the scanner
        Token scanNext();
    };
}

#endif