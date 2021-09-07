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
        Continue,
        Break,
        After,
        Return,
        End,

        Def,
        Type,
        Oper,

        Pub,
        Vol,

        Import,
        From,

        As,

        I8,
        I16,
        I32,
        I64,
        U8,
        U16,
        U32,
        U64,
        F32,
        F64,
        Rune,
        Bool,
        Nothing,

        // Operators
        Plus,
        Minus,
        Star,
        Div,
        FloorDiv,
        Mod,
        Power,

        Amp,
        Pipe,
        Carret,
        LShift,
        RShift,
        Complement,

        And,
        Or,
        Not,

        Lt,
        Gt,
        LtEq,
        GtEq,
        Eq,
        NEq,

        Dot,
        RangeTo,
        Ellipsis,

        TypeProp,
        
        Assign,
        Increm,
        Decrem,

        // Punctuation
        LParen,
        RParen,
        LBracket,
        RBracket,
        LBrace,
        RBrace,
        Comma,
        Colon,
        Arrow,
        Semicolon
    };

    struct Token {
        TokenKind kind;  
        std::string value;
        int line, col;
    };
}

#endif