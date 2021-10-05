#ifndef TOKEN_H_INCLUDED
#define TOKEN_H_INCLUDED

#include "report/report.h"

// token_kind_t enumerates the possible kinds of Chai token
typedef enum {
    // Keyword tokens
    TOK_DEF,
    TOK_END,

    TOK_TYPE,

    TOK_IF,
    TOK_ELIF,
    TOK_ELSE,
    TOK_MATCH,
    TOK_CASE,
    TOK_WHILE,

    TOK_IMPORT,
    TOK_FROM,
    TOK_AS,
    TOK_PUB,

    // Operator tokens
    TOK_PLUS,
    TOK_MINUS,
    TOK_STAR,
    TOK_DIV,
    TOK_FLOORDIV,
    TOK_POWER,
    TOK_MODULO,

    TOK_LSHIFT,
    TOK_RSHIFT,
    TOK_AMP,
    TOK_PIPE,
    TOK_CARRET,

    TOK_AND,
    TOK_OR,
    TOK_NOT,

    TOK_EQ,
    TOK_NEQ,
    TOK_LT,
    TOK_LTEQ,
    TOK_GT,
    TOK_GTEQ,

    TOK_ASSIGN,
    TOK_INCREM,
    TOK_DECREM,

    TOK_DOT,
    TOK_RANGETO,

    TOK_ARROW,

    // Punctuation tokens
    TOK_LPAREN,
    TOK_RPAREN,
    TOK_LBRACKET,
    TOK_RBRACKET,
    TOK_LBRACE,
    TOK_RBRACE,
    TOK_ANNOTAT,
    TOK_COMMA,
    TOK_COLON,
    TOK_SEMICOLON,
    TOK_ELLIPSIS,

    // Value tokens
    TOK_IDENTIFIER,
    TOK_STRINGLIT,
    TOK_RUNELIT,
    TOK_INTLIT,
    TOK_FLOATLIT,
    TOK_NUMLIT,
    TOK_BOOLLIT,
    
    // Special "whitespace" tokens
    TOK_NEWLINE,
    TOK_EOF
} token_kind_t;

// token_t represents a single lexical element of Chai' source code
typedef struct {
    token_kind_t kind;
    char* value;
    text_pos_t position;
} token_t;

#endif