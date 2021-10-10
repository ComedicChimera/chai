#include "parser.h"

#include "lexer.h"

// This file only included the common/utility functions for the parser. The
// actual grammatical structure of Chai as encoded by the parser is written in
// the `grammar.c` file.

typedef struct parser_t {
    // fpath is the path to source file (TODO: replace `source_file_t`)
    const char* fpath;

    // lexer is the lexer for this parser
    lexer_t* lexer;

    // TODO: rest as necessary
} parser_t;

parser_t* parser_new(const char* fpath) {
    // parser is dynamically allocated primarily for convenience
    parser_t* p = malloc(sizeof(parser_t));

    p->fpath = fpath;

    // create the lexer for the parser;
    p->lexer = lexer_new(fpath);
}

void parser_dispose(parser_t* p) {
    lexer_dispose(p->lexer);

    free(p);
}

bool parser_expect(parser_t* p, token_kind_t tok_kind, token_t* tok) {
    // if it is a newline, we can't use consume as otherwise the newline would
    // be skipped -- so we just call the lexer directly
    if (tok_kind == TOK_NEWLINE)
        return lexer_next(p->lexer, tok) && tok->kind == tok_kind;

    // otherwise, we can just use consume
    return parser_consume(p, tok) && tok->kind == tok_kind;
}

bool parser_consume(parser_t* p, token_t* tok) {
    while (lexer_next(p->lexer, tok)) {
        // skip until we encounter a newline :)
        if (tok->kind != TOK_NEWLINE)
            return true;
    }

    // if we reach here => lexical error
    return false;
}

