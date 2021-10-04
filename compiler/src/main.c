#include <stdio.h>

#include "syntax/lexer.h"

int main() {
    lexer_t* lexer = lexer_new("../tests/lex_test.chai");

    token_t tok;
    while (lexer_next(lexer, &tok)) {
        if (tok.kind == TOK_EOF)
            break;

        printf("%d: %s", tok.kind, tok.value);
    }

    lexer_dispose(lexer);
    return 0;
}