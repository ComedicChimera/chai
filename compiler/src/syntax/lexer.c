#include "lexer.h"

#include <stdio.h>

#define TOK_BUFF_BLOCK_SIZE 16

typedef struct {
    // fpath is the path to the Lexer's current file
    const char* fpath;

    // file the current file being read by the lexer
    FILE* file;

    // tok_buff stores the token as it is being built
    char* tok_buff;

    // tok_buff_len is the length of token buffer
    uint32_t tok_buff_len;

    // tok_buff_cap is the capacity of the token buffer
    uint32_t tok_buff_cap;

    // line indicates the current line, starting from 1
    uint32_t line;

    // col indicates the current column, starting from 1
    uint32_t col;

    // start_line indicates which line token construction began on
    uint32_t start_line;

    // start_col indicates which column token construction began on
    uint32_t start_col;
} lexer_t;

// lexer_get_pos generates a text_pos_t based on the Lexer's current position.
static text_pos_t lexer_get_pos(lexer_t* lexer) {
    return (text_pos_t){
        .start_line = lexer->start_line,
        .start_col = lexer->start_col,
        .end_line = lexer->line,
        .end_col = lexer->col
        };
}

// lexer_fail reports an error lexing the user source file
static void lexer_fail(lexer_t* lexer, const char* message) {
    report_compile_error(lexer->fpath, lexer_get_pos(lexer), message);

    // clear the token buffer as its contents are no longer useful
    free(lexer->tok_buff);
}

// lexer_mark_start marks the beginning of a token as it is being read in
static void lexer_mark_start(lexer_t* lexer) {
    lexer->start_line = lexer->line;
    lexer->start_col = lexer->col;
}

// lexer_peek peeks ahead one character in the token stream
static char lexer_peek(lexer_t* lexer) {
    if (feof(lexer->file))
        return EOF;

    char ahead = fgetc(lexer->file);
    ungetc(ahead, lexer->file);
    return ahead;
}

// lexer_update_pos updates the lexer's position based on a character
static void lexer_update_pos(lexer_t* lexer, char c) {
    switch (c) {
        case '\n':
            lexer->line++;
            lexer->col = 1;
            break;
        case '\t':
            lexer->col += 4;
            break;
        default:
            lexer->col++;
            break;
    }
}

// lexer_skip skips a character in the input stream
static void lexer_skip(lexer_t* lexer) {
    // handle EOF case
    if (feof(lexer->file))
        return;

    char c = fgetc(lexer->file);
    if (c == EOF)
        // something went wrong reading the file
        return;

    // still want to update the position
    lexer_update_pos(lexer, c);
}

// lexer_write_char writes a character to the token buffer
static void lexer_write_char(lexer_t* lexer, char c) {
    // check if the length is equal to the capacity => if so, we need to grow
    // the capacity to fit more chars
    if (lexer->tok_buff_len == lexer->tok_buff_cap) {
        lexer->tok_buff_cap += TOK_BUFF_BLOCK_SIZE;
        lexer->tok_buff = (char*)realloc(lexer->tok_buff, lexer->tok_buff_cap);
    }

    // insert at the current length, then increment the length
    lexer->tok_buff[lexer->tok_buff_len++] = c;
}

// lexer_read reads a character from the input stream into the token buffer
static char lexer_read(lexer_t* lexer) {
    // handle EOF case
    if (feof(lexer->file))
        return EOF;

    char c = fgetc(lexer->file);
    if (c == EOF)
        // something went wrong reading the file
        return;

    // update lexer state
    lexer_update_pos(lexer, c);
    lexer_write_char(lexer, c);

    return 0;
}

// lexer_make_token produces a new token from the lexer's current state
token_t lexer_make_token(lexer_t* lexer, token_kind_t kind) {
    // add null terminator to end the token value
    lexer_write_char(lexer, '\0');

    // build the token
    token_t tok = (token_t){
        .kind = kind, 
        .value = lexer->tok_buff, 
        .position = lexer_get_pos(lexer)
    };

    // mark the token buffer as "empty" -- make sure it isn't freed when the
    // lexer is disposed (don't want the last token to be destroyed)
    lexer->tok_buff = NULL;

    return tok;
}

/* -------------------------------------------------------------------------- */

lexer_t* lexer_new(const char* fpath) {
    lexer_t* lexer = (lexer_t*)malloc(sizeof(lexer_t));

    // store the fpath in the lexer for error reporting purposes
    lexer->fpath = fpath;

    // open the file for reading
    lexer->file = fopen(fpath, "r");
    if (lexer->file == NULL) {
        char buff[256];
        sprintf(buff, "failed to open file at `%s`", fpath);
        report_fatal(buff);
    }

    // initialize the Lexer's position
    lexer->line = 1;
    lexer->start_line = 1;
    lexer->col = 1;
    lexer->start_col = 1;

    // mark the token buffer as empty
    lexer->tok_buff = NULL;
}

bool lexer_next(lexer_t* lexer, token_t* tok) {
    // initialize the token buffer
    lexer->tok_buff = (char*)malloc(TOK_BUFF_BLOCK_SIZE);
    lexer->tok_buff_len = 0;
    lexer->tok_buff_cap = TOK_BUFF_BLOCK_SIZE;

    // read as long as there are more tokens to be consumed
    for (char ahead = lexer_peek(lexer); ahead != EOF; ahead = lexer_peek(lexer)) {
        switch (ahead) {
            // skip whitespace
            case ' ':
            case '\t':
            case '\v':
            case '\f':
            case '\r':
                lexer_skip(lexer);
                break;
            // handle newlines
            case '\n':
                // mark the start of the newline
                lexer_mark_start(lexer);

                // read it in
                lexer_read(lexer);

                // make and return the token
                *tok = lexer_make_token(lexer, TOK_NEWLINE);
                return true;
        }
    }

    // if the loop breaks, and we reach here, we have found an EOF
    tok->kind = TOK_EOF;
    return true;
}

void lexer_dispose(lexer_t* lexer) {
    // close the opened file
    fclose(lexer->file);

    // if there is still content remaining in the token buffer, free it
    if (lexer->tok_buff != NULL) {
        free(lexer->tok_buff);
    }

    // free the Lexer itself
    free(lexer);
}