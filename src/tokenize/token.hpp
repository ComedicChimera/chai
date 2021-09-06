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
        StringLiteral,

        // Whitespace
        Newline,

        // Special
        EndOfFile,

        // Keywords
        Let,

        If,
        Elif,
        Else,
        While,
        Do,
        End,

        Def,
        Type,
        // TODO

        // Operators
        Plus,
        Minus,
        Star,
        Div,
        FloorDiv,
        Mod,
        Power,
        
        Assign,
        // TODO
    };

    struct Token {
        TokenKind kind;  
        std::string value;
        int line, col;
    };
}

#endif