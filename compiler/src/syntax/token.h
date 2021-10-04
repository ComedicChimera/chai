#ifndef TOKEN_H_INCLUDED
#define TOKEN_H_INCLUDED

#include "report/report.h"

// token_kind_t enumerates the possible kinds of Chai token
typedef enum {
    TOK_IDENTIFIER,
    
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