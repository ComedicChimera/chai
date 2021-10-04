#include <stdio.h>
#include <string.h>

#include "syntax/lexer.h"

int main(int argc, char* argv[]) {
    if (argc == 1) {
        printf("missing required argument: `fpath`");
        return 1;
    }

    lexer_t* lexer = lexer_new(argv[1]);

    token_t tok;
    while (lexer_next(lexer, &tok)) {
        if (tok.kind == TOK_EOF) {
            printf("hit eof");
            break;
        }
        
        if (!strcmp(tok.value, "\n"))
            printf("%d: \\n\n", tok.kind);
        else
            printf("%d: %s\n", tok.kind, tok.value);
    }

    lexer_dispose(lexer);
    return 0;
}