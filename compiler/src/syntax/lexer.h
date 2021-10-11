#ifndef LEXER_H_INCLUDED
#define LEXER_H_INCLUDED

#include "token.h"
#include "depm/source.h"

// Lexer is responsible for turning a file into a stream of tokens
typedef struct lexer_t lexer_t;

// lexer_new creates a new lexer for the file at the given path
lexer_t* lexer_new(source_file_t* src_file, const char* file_abs_path);

// lexer_next attempts to get the next token from the input stream 
bool lexer_next(lexer_t* lexer, token_t* tok);

// lexer_dispose disposes of all resources associated with the lexer (besides
// the tokens it produced -- those must be managed separately)
void lexer_dispose(lexer_t* lexer);


#endif