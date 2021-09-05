#ifndef TOKEN_H_INCLUDED
#define TOKEN_H_INCLUDED

#include <string>

namespace chai {
    enum class TokenKind {
        // Values
        Identifier,
        IntLiteral,
        FloatLiteral,
        NumLiteral,
        RuneLiteral,
        BoolLiteral,

        // Whitespace
        Newline,

        // Special
        EndOfFile
    };

    struct Token {
        TokenKind kind;  
        std::string value;
        int line, col;
    };
}

#endif