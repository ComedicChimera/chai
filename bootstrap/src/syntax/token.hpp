#ifndef _TOKEN_H_
#define _TOKEN_H_

#include "chaic.hpp"

namespace chai {
    // Enumeration of different kinds of tokens.
    enum class TokenKind {
        PACKAGE,

        FUNC,
        OPER,

        LET,
        CONST,

        IF,
        ELIF,
        ELSE,
        WHILE,
        FOR,

        BREAK,
        CONTINUE,
        RETURN,

        AS,
        
        UNIT,
        BOOL,
        I8,
        U8,
        I16,
        U16,
        I32,
        U32,
        I64,
        U64,
        F32,
        F64,

        PLUS,
        MINUS,
        STAR,
        DIV,
        MOD,
        POWER,

        LT,
        LTEQ,
        GT,
        GTEQ,
        EQ,
        NEQ,

        LSHIFT,
        RSHIFT,
        COMPL,

        AMP,
        PIPE,
        CARRET,

        AND,
        OR,
        NOT,

        ASSIGN,
        INCREMENT,
        DECREMENT,

        LPAREN,
        RPAREN,
        LBRACE,
        RBRACE,
        LBRACKET,
        RBRACKET,
        COMMA,
        COLON,
        SEMICOLON,
        ATSIGN,
        DOT,
        
        STRINGLIT,
        RUNELIT,
        NUMLIT,
        FLOATLIT,
        INTLIT,
        BOOLLIT,
        NULLLIT,
        IDENTIFIER,

        END_OF_FILE,
    };

    // Token represents a single lexical token.
    struct Token {
        TokenKind kind { TokenKind::END_OF_FILE };
        std::string value;
        TextSpan span;
    };
}

#endif