#ifndef PARSER_H_INCLUDED
#define PARSER_H_INCLUDED

#include "token.h"

// parser_t is parser for the Chai programming language.  It is
// recursive-descent parser and performs 4 distinct functions, namely:
// 
// 1. Syntax Validation
// 2. AST Construction
// 3. Module Import Analysis (identifying/walking imports and passing them to
//    the resolver)
// 4. Symbol Declaration and Dependency Identification (declaring/resolving
//    symbols and marking their definitions as required or resolving them if
//    possible)
//
// The parser is designed to parse a single file in one pass; however, it uses
// the dependency graph and symbol table to ensure that out of order and
// recursive symbol definition is still completely legal and compileable.
typedef struct parser_t parser_t;

// parser_new creates a new parser for a given a source file 
// TODO: formalize to use constructs for `source.h` (ie. modules, sources files,
// etc).
parser_t* parser_new(const char* fpath);

// parser_dispose disposes of all resources associated with the parser including
// the parser itself.
void parser_dispose(parser_t* p);

// parser_expect asserts the presence of a specific token kind.  This function
// does not accept EOFs.  However, it will skip any newlines unless newline is
// the expected token.  It also accept a semicolon in place of a newline
// anywhere.  If the expectation is satisfied, the token is returned via the
// passed in pointer and true is returned.  If the expectation fails or a
// lexical error occurs, false is returned and an appropriate unexpected token
// error is reported as necessary.
bool parser_expect(parser_t* p, token_kind_t tok_kind, token_t* tok);

// parser_consume reads the next token from the stream that isn't a newline. The
// resulting token is returned via the pointer if there are no lexical errors as
// indicating via the boolean return being true.
bool parser_consume(parser_t* p, token_t* tok);

// parser_report_unexpected reports an unexpected token error (or unexpected end
// of file error as necessary)
void parser_report_unexpected(parser_t* p, token_t* tok);


#endif